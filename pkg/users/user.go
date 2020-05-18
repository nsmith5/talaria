package users

//go:generate flatc -o .. --go user.fbs

import (
	"encoding"

	flatbuffers "github.com/google/flatbuffers/go"
)

type User interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	Username() string
	PasswordHash() string
	Admin() bool
	Email() string
	Aliases() []string
}

// user is a flatbuffer implementation of the User interface
type user struct {
	fb *fbuser
}

func NewUser(username, hash string, admin bool, email string, aliases []string) User {
	builder := flatbuffers.NewBuilder(0)

	uOffset := builder.CreateString(username)
	hashOffset := builder.CreateString(hash)
	eOffset := builder.CreateString(email)

	var offsets []flatbuffers.UOffsetT
	for i := 0; i < len(aliases); i++ {
		offsets = append(offsets, builder.CreateString(aliases[i]))
	}

	fbuserStartAliasesVector(builder, len(aliases))
	for i := 0; i < len(aliases); i++ {
		builder.PrependUOffsetT(offsets[i])
	}
	aliasesOffset := builder.EndVector(len(aliases))

	fbuserStart(builder)
	fbuserAddUsername(builder, uOffset)
	fbuserAddPasswordHash(builder, hashOffset)
	fbuserAddAdmin(builder, admin)
	fbuserAddEmail(builder, eOffset)
	fbuserAddAliases(builder, aliasesOffset)

	offset := fbuserEnd(builder)
	builder.Finish(offset)

	buf := builder.FinishedBytes()
	fb := GetRootAsfbuser(buf, 0)
	return user{fb}
}

func (u user) Username() string {
	return string(u.fb.Username())
}

func (u user) PasswordHash() string {
	return string(u.fb.PasswordHash())
}

func (u user) Admin() bool {
	return u.fb.Admin()
}

func (u user) Email() string {
	return string(u.fb.Email())
}

func (u user) Aliases() []string {
	var aliases []string
	for i := 0; i < u.fb.AliasesLength(); i++ {
		aliases = append(aliases, string(u.fb.Aliases(i)))
	}
	return aliases
}

func (u user) UnmarshalBinary(data []byte) error {
	u.fb = GetRootAsfbuser(data, 0)
	return nil
}

func (u user) MarshalBinary() ([]byte, error) {
	data := u.fb.Table().Bytes[u.fb.Table().Pos:]
	output := make([]byte, len(data))
	copy(output, data)
	return output, nil
}
