package usecase

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/resource"
)

func NewFindUser(repo resource.UserRepo) FindUser {
	return FindUser{
		repo: repo,
	}
}

type FindUser struct {
	repo resource.UserRepo
}

func (u FindUser) Do(id resource.UserID) (resource.User, bool, error) {
	ctx := context.TODO()

	user, found, err := u.repo.Find(ctx, id)
	if err != nil {
		return resource.User{}, false, fmt.Errorf("failed to find user: %w", err)
	}

	return user, found, nil
}

func NewCreateUser(serv resource.UserService, repo resource.UserRepo) CreateUser {
	return CreateUser{
		serv: serv,
		repo: repo,
	}
}

type CreateUser struct {
	serv resource.UserService
	repo resource.UserRepo
}

func (u CreateUser) Do(name, email, pass string) (resource.User, error) {
	ctx := context.TODO()

	id, err := u.serv.Create(ctx, email, pass)
	if err != nil {
		return resource.User{}, fmt.Errorf("failed to create user with service: %w", err)
	}

	user, err := resource.NewUser(id, name, email)
	if err != nil {
		return resource.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	if err := u.repo.Save(ctx, user); err != nil {
		return resource.User{}, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}
