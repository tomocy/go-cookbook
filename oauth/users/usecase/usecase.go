package usecase

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/users"
)

func NewCreateUser(userRepo users.UserRepo) CreateUser {
	return CreateUser{
		repo: userRepo,
	}
}

type CreateUser struct {
	repo users.UserRepo
}

func (u CreateUser) Do(email, pass string) (users.User, error) {
	ctx := context.TODO()

	_, found, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return users.User{}, fmt.Errorf("failed to find user by email: %w", err)
	}
	if found {
		return users.User{}, users.ErrInvalidArg("duplicated email address")
	}

	id, err := u.repo.NextID(ctx)
	if err != nil {
		return users.User{}, fmt.Errorf("failed to generate user id: %w", err)
	}
	hashed, err := users.HashPassword(pass)
	if err != nil {
		return users.User{}, fmt.Errorf("failed to hash password: %w", err)
	}
	user, err := users.NewUser(id, email, hashed)
	if err != nil {
		return users.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	if err := u.repo.Save(ctx, user); err != nil {
		return users.User{}, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func NewAuthenticateUser(repo users.UserRepo) AuthenticateUser {
	return AuthenticateUser{
		repo: repo,
	}
}

type AuthenticateUser struct {
	repo users.UserRepo
}

func (u AuthenticateUser) Do(email, pass string) (users.User, bool, error) {
	ctx := context.TODO()

	user, found, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return users.User{}, false, err
	}
	if !found {
		return users.User{}, false, nil
	}
	if !user.Password().IsSame(pass) {
		return users.User{}, false, nil
	}

	return user, true, nil
}
