package users

import (
	"fmt"
	"testing"
)

func TestNewUser(t *testing.T) {
	u := NewUser("username", "passhash", true, "email@example.com", []string{"this", "that"})
	if u.Username() != "username" {
		t.Error("Incorrect username")
	}
	if u.PasswordHash() != "passhash" {
		t.Error("Incorrect password")
	}
	if !u.Admin() {
		t.Error("Should be admin")
	}
	if u.Email() != "email@example.com" {
		t.Error("Bad email")
	}

	if !(len(u.Aliases()) == 2) {
		t.Error("Wrong length of aliases")
	}
}

func TestMarshalBinary(t *testing.T) {
	u := NewUser("username", "passhash", true, "email@example.com", []string{"this", "that"})
	data, err := u.MarshalBinary()
	if err != nil {
		t.Error("Failed to serialize")
	}
	fmt.Println(string(data))

	data2 := make([]byte, len(data))
	copy(data2, data)

	var u2 user
	err = u2.UnmarshalBinary(data2)
	if err != nil {
		t.Error("Failed to deserialize")
	}

	if u2.Username() != u.Username() {
		t.Error("Incorrect Username")
	}
}
