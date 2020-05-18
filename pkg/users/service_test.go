package users

/*
import (
	"context"
	"testing"

	"github.com/nsmith5/talaria/pkg/kv"
)

func TestCRUD(t *testing.T) {
	user := User{
		Username:     "nathan",
		PasswordHash: "boo",
		IsAdmin:      true,
		Email:        "example.com",
		Aliases: []string{
			"other@example.com",
			"other@example.net",
		},
	}

	store := kv.NewMemStore()
	us := NewService(store)

	ctx := context.Background()

	err := us.Create(ctx, user)
	if err != nil {
		t.Error("Failed to create user")
	}

	user.IsAdmin = false
	err = us.Update(ctx, user)
	if err != nil {
		t.Error("Failed to update user to non-admin")
	}

	user2, err := us.Fetch(ctx, user.Username)
	if err != nil {
		t.Error("Failed to fetch user again")
	}

	if user2.IsAdmin != false {
		t.Error("Failed to update user")
	}

	err = us.Delete(ctx, user.Username)
	if err != nil {
		t.Error("Failed to delete user")
	}

	ctx, cancel := context.WithCancel(ctx)
	cancel()

	err = us.Delete(ctx, user.Username)
	if err != context.Canceled {
		t.Error("Should failed as cancelled")
	}
}
*/
