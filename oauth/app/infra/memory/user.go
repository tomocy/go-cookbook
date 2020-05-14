package mock

import (
	"context"

	"github.com/tomocy/go-cookbook/oauth/app"
)

var UserRepo = userRepo{
	users: make(map[app.UserID]app.User),
}

type userRepo struct {
	users map[app.UserID]app.User
}

func (r *userRepo) Save(_ context.Context, user app.User) error {
	r.users[user.ID()] = user
	return nil
}
