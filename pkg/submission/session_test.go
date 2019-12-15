package submission

import (
    "io"
    "fmt"
    "time"
    "bytes"
    "testing"
)

type FakeConn struct {
    *bytes.Buffer
}

func (conn FakeConn) Close() error {
    return nil
}

func TestEhlo(t *testing.T) {
    var buf bytes.Buffer
    conn := FakeConn{Buffer: &buf}
    s := Session{
        conn: conn,
        lmt: io.LimitedReader{R: conn, N: LINELIMIT},
        hostname: "example.com:8888",
    }

    go s.Run()

    time.Sleep(1 * time.Second)
    io.WriteString(conn, "EHLO localhost\r\n")
    time.Sleep(1 * time.Second)
    fmt.Println(conn.String())
}
