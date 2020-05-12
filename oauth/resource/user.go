package resource

import "context"

type UserRepo interface {
	Find(context.Context, UserID) (User, bool, error)
}

func NewUser(id UserID, name string) (User, error) {
	var u User
	if err := u.setID(id); err != nil {
		return User{}, err
	}
	if err := u.setName(name); err != nil {
		return User{}, err
	}

	return u, nil
}

type User struct {
	id   UserID
	name string
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

type ErrInvalidArg string

func (e ErrInvalidArg) Error() string {
	return string(e)
}
