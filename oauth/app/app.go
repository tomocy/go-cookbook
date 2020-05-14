package app

import (
	"context"
	"fmt"
)

type UserRepo interface {
	Save(context.Context, User) error
}

func NewUser(id UserID) (User, error) {
	u := User{
		providers: make(map[string]Provider),
	}
	if err := u.setID(id); err != nil {
		return User{}, err
	}

	return u, nil
}

type User struct {
	id        UserID
	providers map[string]Provider
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

func (u *User) AddProvider(p Provider) error {
	if _, ok := u.providers[p.name]; ok {
		return ErrInvalidArg("duplicated provider")
	}

	u.providers[p.name] = p

	return nil
}

type UserID string

func NewProvider(name, tok string) (Provider, error) {
	var p Provider
	if err := p.setName(name); err != nil {
		return Provider{}, err
	}
	if err := p.setToken(tok); err != nil {
		return Provider{}, err
	}

	return p, nil
}

type Provider struct {
	name string
	tok  string
}

func (p *Provider) setName(name string) error {
	if name == "" {
		return ErrInvalidArg("name should not be empty")
	}

	p.name = name

	return nil
}

func (p *Provider) setToken(tok string) error {
	if tok == "" {
		return fmt.Errorf("token should not be empty")
	}

	p.tok = tok

	return nil
}

type errInput interface {
	error
	ErrInput()
}

type ErrInvalidArg string

func (e ErrInvalidArg) Error() string {
	return string(e)
}

func (ErrInvalidArg) ErrInput() {}
