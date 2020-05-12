package memory

import (
	"context"

	"github.com/tomocy/go-cookbook/oauth/users"
	"github.com/tomocy/go-cookbook/oauth/users/infra/rand"
)

func NewUserRepo() *UserRepo {
	return &UserRepo{
		users: make(map[users.UserID]users.User),
	}
}

type UserRepo struct {
	users map[users.UserID]users.User
}

func (UserRepo) NextID(context.Context) (users.UserID, error) {
	return users.UserID(rand.GenerateString(20)), nil
}

func (r UserRepo) Find(_ context.Context, id users.UserID) (users.User, bool, error) {
	for _, stored := range r.users {
		if stored.ID() == id {
			return stored, true, nil
		}
	}

	return users.User{}, false, nil
}

func (r UserRepo) FindByEmail(_ context.Context, email string) (users.User, bool, error) {
	for _, stored := range r.users {
		if stored.Email() == email {
			return stored, true, nil
		}
	}

	return users.User{}, false, nil
}

func (r *UserRepo) Save(_ context.Context, user users.User) error {
	r.users[user.ID()] = user
	return nil
}

func (r *UserRepo) Delete(_ context.Context, user users.User) error {
	delete(r.users, user.ID())
	return nil
}
