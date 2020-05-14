package server

import "context"

type User struct {
	ID        string
	Providers map[string]Provider
}

type Provider struct {
	Name  string
	Token string
}

type OwnerService interface {
	Fetch(context.Context, string) (Owner, error)
}

type Owner struct {
	Name     string
	Provider string
}
