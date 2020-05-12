package resource

import (
	"context"
	"errors"
)

type UserRepo interface {
	Find(context.Context, UserID) (User, bool, error)
	Save(context.Context, User) error
}

type UserService interface {
	Create(context.Context, string, string) (UserID, error)
}

func NewUser(id UserID, name, email string) (User, error) {
	var u User
	if err := u.setID(id); err != nil {
		return User{}, err
	}
	if err := u.setName(name); err != nil {
		return User{}, err
	}
	if err := u.setEmail(email); err != nil {
		return User{}, err
	}

	return u, nil
}

type User struct {
	id    UserID
	name  string
	email string
}

func (u User) ID() UserID {
	return u.id
}

func (u *User) setID(id UserID) error {
	if id == "" {
		return ErrInvalidArg("id should not be empty")
	}

	u.id = id

	return nil
}

func (u User) Email() string {
	return u.email
}

func (u *User) setEmail(email string) error {
	if email == "" {
		return ErrInvalidArg("email should not be empty")
	}

	u.email = email

	return nil
}

func (u User) Name() string {
	return u.name
}

func (u *User) setName(name string) error {
	if name == "" {
		return ErrInvalidArg("name should not be empty")
	}

	u.name = name

	return nil
}

type UserID string

func IsErrInput(err error) bool {
	if _, ok := err.(errInput); ok {
		return true
	}

	err = errors.Unwrap(err)
	if err == nil {
		return false
	}
	return IsErrInput(err)
}

type errInput interface {
	ErrInput()
}

type ErrInvalidArg string

func (ErrInvalidArg) ErrInput() {}

func (e ErrInvalidArg) Error() string {
	return string(e)
}
