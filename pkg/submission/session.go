package submission

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/textproto"
	"runtime/debug"
	"strings"

	"code.nfsmith.ca/nsmith/talaria/pkg/email"
)

const (
	// 512B limit to the size of and command
	// https://tools.ietf.org/html/rfc5321#section-4.5.3.1.4
	LINELIMIT = 1 << 9

	// 10 MiB limit to size of message body. This was a personal
	// preference and will be imposed using RCF 1870.
	// https://tools.ietf.org/html/rfc1870
	BODYLIMIT = 1 << 20
)

// State of a SMTP submission session
const (
	stateStart int = 0
	stateEhlo      = iota
	stateAuthenticated
	stateMail
	stateRcpt
	stateData
)

// Session is a single TCP connection containing mail submission dialog.
//
// NOTE: There are /no/ network related entities here. The session / server
// boundary is an abstraction boundry for network related things. Network
// specific details (e.g. Read and write timeouts) are all operated at the
// server level and all protocol specific things happen at the session
// level.
type Session struct {
	// Raw connection
	conn io.ReadWriteCloser

	// Limit reader wrapped on top of raw connection
	// to migirate DOS and message sizing limits
	lmt io.LimitedReader

	// Buffered reader on top of limit reader to
	// peak at commands
	buf *bufio.Reader

	hostname string
	state    int

	// Message state
	from       string
	recipients []string
	msg        []byte

	// Sender
	sender email.Sender
}

// Run initiates a mail submission session.
func (s *Session) Run() {
	// Any panics during a session should be recovered we don't bring down the
	// whole server because of one naughty connection.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Session panicked:", string(debug.Stack()))
		}
	}()
	defer s.conn.Close()

	s.Greet()

	for {
		s.lmt.N = LINELIMIT
		cmd, err := s.buf.Peek(4)
		if err != nil {
			// Whelp, not sure we're recovering from this one
			return
		}

		// Must accept commands in a case insensitive fashion
		// https://tools.ietf.org/html/rfc5321#section-2.4
		cmdstr := strings.ToUpper(string(cmd))

		switch cmdstr {

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.1
		case "HELO":
			// This is probably a violation of some RFC, but there is no
			// supported way for this HELO to lead to authentication. No way
			// I'm going to accept an HELO. Step up your game client.
			s.NotImplemented()

			// The scorn for HELO gets worse: we're going to drop the
			// connection. RFC be damned.
			return

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.1
		case "EHLO":
			s.ExtendedHello()

		// https://tools.ietf.org/html/rfc4954#section-4
		case "AUTH":
			s.Authenticate()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.2
		case "MAIL":
			s.Mail()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.3
		case "RCPT":
			s.Recipient()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.4
		case "DATA":
			s.Data()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.5
		case "RSET":
			s.Reset()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.6
		case "VRFY":
			s.NotImplemented()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.7
		case "EXPN":
			s.NotImplemented()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.8
		case "HELP":
			s.NotImplemented()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.9
		case "NOOP":
			s.Noop()

		// https://tools.ietf.org/html/rfc5321#section-4.1.1.10
		case "QUIT":
			s.Quit()

			// NOTE: We're quiting whether or not the QUIT command was
			// well formed. Probably a violation of RFC, but we're not
			// going to hang around leaving this connection open because
			// of a typo in a QUIT command.
			return

		default:
			s.BadRequest()
		}
	}
}

func (s *Session) Greet() {
	fmt.Fprintf(s.conn, "220 %s ESMTP talaria\r\n", s.hostname)
}

func (s *Session) NotImplemented() {
	fmt.Fprintf(s.conn, "502 Not implemented\r\n")
}

func (s *Session) ExtendedHello() {
	if s.state != stateStart {
		fmt.Fprint(s.conn, "503 Talaria only accepts EHLO at the start of a session\r\n")
		return
	}

	cmd, err := s.buf.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Fprintf(s.conn, "500 Who there, line is too long\r\n")

	case err != nil:
		fmt.Fprintf(s.conn, "500 I don't understand\r\n")

	case !strings.HasSuffix(cmd, "\r\n"):
		fmt.Fprintf(s.conn, "501 EHLO must end with CRLF\r\n")

	default:
		fmt.Fprintf(s.conn, "250-%s Arrrr, gimme yer mail\r\n", s.hostname)
		fmt.Fprintf(s.conn, "250-8BITMIME\r\n")
		fmt.Fprintf(s.conn, "250-SIZE %d\r\n", BODYLIMIT)
		fmt.Fprintf(s.conn, "250 AUTH PLAIN\r\n")
		s.state = stateEhlo
	}
}

func (s *Session) Authenticate() {
	// TODO: Implement me
	s.buf.ReadString('\n')
	fmt.Fprint(s.conn, "235 I didn't do anything\r\n")
	s.state = stateAuthenticated
}

func (s *Session) Mail() {
	if s.state != stateAuthenticated {
		s.buf.ReadString('\n')
		fmt.Fprint(s.conn, "503 Not authenticated")
		return
	}

	cmd, err := s.buf.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Fprintf(s.conn, "500 Line too long. Cut down those MAIL params\r\n")

	case err != nil:
		fmt.Fprintf(s.conn, "500 I don't even know how you failed here\r\n")

	case !(strings.HasSuffix(cmd, "\r\n") && strings.HasPrefix(cmd, "MAIL FROM:")):
		fmt.Fprintf(s.conn, "501 Bad format in your mail command\r\n")

	default:
		cmd = strings.TrimPrefix(cmd, "MAIL FROM:")
		cmd = strings.TrimSuffix(cmd, "\r\n")
		parts := strings.Split(cmd, " ")

		// TODO: check if this sender is acceptable
		s.from = parts[0]

		fmt.Fprint(s.conn, "250 Ok\r\n")
		s.state = stateMail
	}
}

func (s *Session) Recipient() {
	if !(s.state == stateRcpt || s.state == stateMail) {
		s.buf.ReadString('\n')
		fmt.Fprint(s.conn, "503 Need do MAIL before adding recipients\r\n")
		return
	}

	cmd, err := s.buf.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Fprintf(s.conn, "500 Line too long. Cut down those MAIL params\r\n")

	case err != nil:
		fmt.Fprintf(s.conn, "500 I don't even know how you failed here\r\n")

	case !(strings.HasSuffix(cmd, "\r\n") && strings.HasPrefix(cmd, "RCPT TO:")):
		fmt.Fprintf(s.conn, "501 Bad format in your mail command\r\n")

	default:
		cmd = strings.TrimPrefix(cmd, "RCPT TO:")
		cmd = strings.TrimSuffix(cmd, "\r\n")
		parts := strings.Split(cmd, " ")

		// TODO: Bring this up to spec. Strip weird routing information out etc...
		s.recipients = append(s.recipients, parts[0])

		fmt.Fprint(s.conn, "250 Ok\r\n")

		s.state = stateRcpt
	}

}

func (s *Session) Data() {
	if s.state != stateRcpt {
		s.buf.ReadString('\n')
		fmt.Fprint(s.conn, "503 No recipients recieved yet\r\n")
		return
	}

	cmd, err := s.buf.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Fprint(s.conn, "500 Gees long line for a reset huh?\r\n")

	case err != nil:
		fmt.Fprint(s.conn, "500 I don't even know how you failed here\r\n")

	case strings.ToUpper(cmd) != "DATA\r\n":
		fmt.Fprint(s.conn, "501 No args in a data! What did you do?\r\n")

	default:
		fmt.Fprint(s.conn, "354 Do your thing\r\n")

		// Set the limit reader to the /larger/ body limit
		s.lmt.N = BODYLIMIT
		msg, err := ioutil.ReadAll(textproto.NewReader(s.buf).DotReader())
		if err != nil {
			panic(err)
		}

		s.msg = msg

		// TODO: Pass mail off to MTA
		fmt.Println("From:", s.from)
		fmt.Printf("To:%v\n", s.recipients)
		fmt.Println("Message:", string(s.msg))

		fmt.Fprint(s.conn, "250 Boom baby. Email accepted\r\n")

		// Reset state
		s.from = ""
		s.recipients = nil
		s.msg = nil
		s.state = stateAuthenticated
	}
}

func (s *Session) Reset() {
	cmd, err := s.buf.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Fprintf(s.conn, "500 Gees long line for a reset huh?\r\n")

	case err != nil:
		fmt.Fprintf(s.conn, "500 I don't even know how you failed here\r\n")

	case strings.ToUpper(cmd) != "RSET\r\n":
		fmt.Fprintf(s.conn, "501 No args in a reset! What did you do?\r\n")

	default:
		for _, state := range [3]int{stateStart, stateEhlo, stateAuthenticated} {
			if s.state == state {
				return
			}
		}

		s.from = ""
		s.recipients = nil
		s.msg = nil

		s.state = stateAuthenticated
	}
}

func (s *Session) Noop() {
	cmd, err := s.buf.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Fprintf(s.conn, "500 Whoa there, line is too long\r\n")

	case err != nil:
		fmt.Fprintf(s.conn, "500 I don't even know how you failed here\r\n")

	case !strings.HasSuffix(cmd, "\r\n"):
		fmt.Fprintf(s.conn, "501 Bad Noop format....\r\n")

	default:
		fmt.Fprintf(s.conn, "250 Ok\r\n")
	}
}

func (s *Session) Quit() {
	cmd, err := s.buf.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Fprintf(s.conn, "500 Whoa there, line is too long\r\n")

	case err != nil:
		fmt.Fprintf(s.conn, "500 I don't even know how you failed here\r\n")

	case strings.ToUpper(cmd) != "QUIT\r\n":
		fmt.Fprintf(s.conn, "501 You messed up a quit? (shakes head)\r\n")

	default:
		fmt.Fprintf(s.conn, "221 Bye\r\n")
	}
}

func (s *Session) BadRequest() {
	_, err := s.buf.ReadString('\n')
	switch {
	case err == io.EOF:
		fmt.Fprint(s.conn, "500 Whoa there, the line is too long /and/ unrecognized command\r\n")

	default:
		fmt.Fprint(s.conn, "500 Command not recognized\r\n")
	}
}
