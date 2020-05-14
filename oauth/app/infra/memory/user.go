package memory

import (
	"context"

	"github.com/tomocy/go-cookbook/oauth/app"
)

func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[app.UserID]app.User),
	}
}

type UserRepo struct {
	users map[app.UserID]app.User
}

func (r UserRepo) Find(_ context.Context, id app.UserID) (app.User, bool, error) {
	u, ok := r.users[id]
	return u, ok, nil
}

func (r *UserRepo) Save(_ context.Context, user app.User) error {
	r.users[user.ID()] = user
	return nil
}
