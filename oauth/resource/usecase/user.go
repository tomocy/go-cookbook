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
