package usecase

import (
	"context"
	"testing"

	"github.com/tomocy/go-cookbook/oauth/authz/infra/memory"
)

func TestCreateUser(t *testing.T) {
	userRepo := memory.NewUserRepo()
	sessRepo := memory.NewSessionRepo()

	email, pass := "email", "pass"
	u := NewCreateUser(userRepo, sessRepo)
	user, _, err := u.Do(email, pass)
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

	CheckUserAuthenticated := NewCheckUserAuthenticated(sessRepo)
	_, authenticated, _ := CheckUserAuthenticated.Do(user.ID())
	if !authenticated {
		t.Errorf("should have authenticated user in session")
		return
	}
}

func TestAuthenticateUser(t *testing.T) {
	userRepo := memory.NewUserRepo()
	sessRepo := memory.NewSessionRepo()

	email, pass := "email", "pass"
	create := NewCreateUser(userRepo, sessRepo)
	user, _, _ := create.Do(email, pass)

	u := NewAuthenticateUser(userRepo, sessRepo)
	returned, _, authenticated, err := u.Do(email, pass)
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

	CheckUserAuthenticated := NewCheckUserAuthenticated(sessRepo)
	_, authenticated, _ = CheckUserAuthenticated.Do(returned.ID())
	if !authenticated {
		t.Errorf("should have authenticated user in session")
		return
	}
}

func TestCheckUserAuthenticated(t *testing.T) {
	userRepo := memory.NewUserRepo()
	sessRepo := memory.NewSessionRepo()

	email, pass := "email", "pass"

	createUser := NewCreateUser(userRepo, sessRepo)
	user, _, _ := createUser.Do(email, pass)

	u := NewCheckUserAuthenticated(sessRepo)
	_, authenticated, err := u.Do(user.ID())
	if err != nil {
		t.Errorf("should have checked that user is authenticated: %s", err)
		return
	}
	if !authenticated {
		t.Errorf("should have returned that user is authenticated")
		return
	}
}
