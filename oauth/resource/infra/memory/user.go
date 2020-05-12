package memory

import (
	"context"

	"github.com/tomocy/go-cookbook/oauth/resource"
)

func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[resource.UserID]resource.User),
	}
}

type UserRepo struct {
	users map[resource.UserID]resource.User
}

func (r *UserRepo) Save(_ context.Context, user resource.User) error {
	r.users[user.ID()] = user
	return nil
}
