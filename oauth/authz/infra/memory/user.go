package memory

import (
	"context"

	"github.com/tomocy/go-cookbook/oauth/authz"
	"github.com/tomocy/go-cookbook/oauth/authz/infra/rand"
)

func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[authz.UserID]authz.User),
	}
}

type UserRepo struct {
	users map[authz.UserID]authz.User
}

func (UserRepo) NextID(context.Context) (authz.UserID, error) {
	return authz.UserID(rand.GenerateString(20)), nil
}

func (r UserRepo) Find(_ context.Context, id authz.UserID) (authz.User, bool, error) {
	for _, stored := range r.users {
		if stored.ID() == id {
			return stored, true, nil
		}
	}

	return authz.User{}, false, nil
}

func (r UserRepo) FindByEmail(_ context.Context, email string) (authz.User, bool, error) {
	for _, stored := range r.users {
		if stored.Email() == email {
			return stored, true, nil
		}
	}

	return authz.User{}, false, nil
}

func (r *UserRepo) Save(_ context.Context, user authz.User) error {
	r.users[user.ID()] = user
	return nil
}

func (r *UserRepo) Delete(_ context.Context, user authz.User) error {
	delete(r.users, user.ID())
	return nil
}
