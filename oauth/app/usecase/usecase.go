package usecase

import (
	"context"

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
