package mock

import (
	"context"

	"github.com/tomocy/go-cookbook/oauth/app"
)

var UserRepo userRepo

type userRepo struct{}

func (userRepo) FindUser(_ context.Context, id app.UserID) (app.User, bool, error) {
	u, err := app.NewUser(id)
	if err != nil {
		return app.User{}, false, err
	}

	return u, true, nil
}
