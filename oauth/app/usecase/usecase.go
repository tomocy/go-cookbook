package usecase

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/app"
)

func NewFindUser(repo app.UserRepo) FindUser {
	return FindUser{
		repo: repo,
	}
}

type FindUser struct {
	repo app.UserRepo
}

func (u FindUser) Do(id app.UserID) (app.User, bool, error) {
	ctx := context.TODO()

	user, found, err := u.repo.Find(ctx, id)
	if err != nil {
		return app.User{}, false, err
	}

	return user, found, nil
}

func NewAddProvider(repo app.UserRepo) AddProvider {
	return AddProvider{
		repo: repo,
	}
}

type AddProvider struct {
	repo app.UserRepo
}

func (u AddProvider) Do(id app.UserID, name, tok string) error {
	ctx := context.TODO()

	user, found, err := u.repo.Find(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if !found {
		return app.ErrInvalidArg("no such user")
	}

	prov, err := app.NewProvider(name, tok)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}
	if err := user.AddProvider(prov); err != nil {
		return err
	}

	if err := u.repo.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func NewFetchOwner(repo app.UserRepo, serv app.UserService) FetchOwner {
	return FetchOwner{
		repo: repo,
		serv: serv,
	}
}

type FetchOwner struct {
	repo app.UserRepo
	serv app.UserService
}

func (u FetchOwner) Do(id app.UserID, name string) (app.User, error) {
	ctx := context.TODO()

	user, found, err := u.repo.Find(ctx, id)
	if err != nil {
		return app.User{}, fmt.Errorf("failed to find user: %w", err)
	}
	if !found {
		return app.User{}, app.ErrInvalidArg("no such user")
	}

	prov, found := user.Provider(name)
	if !found {
		return app.User{}, app.ErrInvalidArg("no such provider")
	}

	owner, err := u.serv.FetchWithAccessToken(ctx, prov.Token())
	if err != nil {
		return app.User{}, fmt.Errorf("failed to fetch user with access token: %w", err)
	}

	return owner, nil
}
