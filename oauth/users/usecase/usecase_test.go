package usecase

import (
	"context"
	"fmt"
	"testing"

	"github.com/tomocy/go-cookbook/oauth/authz"
	"github.com/tomocy/go-cookbook/oauth/authz/infra/memory"
)

func TestCreateUser(t *testing.T) {
	userRepo := memory.NewUserRepo()
	sessRepo := memory.NewSessionRepo()

	email, pass := "email", "pass"
	u := NewCreateUser(userRepo, sessRepo)
	user, err := u.Do(email, pass)
	if err != nil {
		t.Errorf("should have created user: %s", err)
		return
	}

	saved, ok, _ := userRepo.Find(context.Background(), user.ID())
	if !ok {
		t.Errorf("should have saved the created user")
		return
	}
	if err := assertUser(user, saved); err != nil {
		t.Errorf("should have returned the create user: %s", err)
		return
	}
}

func TestAuthenticateUser(t *testing.T) {
	userRepo := memory.NewUserRepo()
	sessRepo := memory.NewSessionRepo()

	email, pass := "email", "pass"
	create := NewCreateUser(userRepo, sessRepo)
	user, _ := create.Do(email, pass)

	u := NewAuthenticateUser(userRepo, sessRepo)
	returned, authenticated, err := u.Do(email, pass)
	if err != nil {
		t.Errorf("should have authenticated user: %s", err)
		return
	}
	if !authenticated {
		t.Errorf("should have authenticated user")
		return
	}
	if err := assertUser(returned, user); err != nil {
		t.Errorf("should have returned the authenticated user: %s", err)
		return
	}
}

func assertUser(actual, expected authz.User) error {
	if actual.ID() != expected.ID() {
		return reportUnexpected("id", actual.ID(), expected.ID())
	}

	return nil
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, but expected %v", name, actual, expected)
}
