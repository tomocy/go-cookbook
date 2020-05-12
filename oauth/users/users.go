package users

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	NextID(context.Context) (UserID, error)
	FindByEmail(context.Context, string) (User, bool, error)
	Save(context.Context, User) error
	Delete(context.Context, User) error
}

func NewUser(id UserID, email string, pass Password) (User, error) {
	var u User
	if err := u.setID(id); err != nil {
		return User{}, err
	}
	if err := u.setEmail(email); err != nil {
		return User{}, err
	}
	if err := u.setPassword(pass); err != nil {
		return User{}, err
	}

	return u, nil
}

type User struct {
	id       UserID
	email    string
	password Password
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

func (u User) Password() Password {
	return u.password
}

func (u *User) setPassword(pass Password) error {
	if pass == "" {
		return ErrInvalidArg("password should not be empty")
	}

	u.password = pass

	return nil
}

type UserID string

func HashPassword(plain string) (Password, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return Password(hashed), nil
}

type Password string

func (p Password) IsSame(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(p), []byte(plain)) == nil
}

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
