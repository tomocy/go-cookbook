package memory

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/authz"
	"github.com/tomocy/go-cookbook/oauth/authz/infra/rand"
)

func NewClientRepo() *ClientRepo {
	return &ClientRepo{
		clients: make(map[authz.ClientID]authz.Client),
	}
}

type ClientRepo struct {
	clients map[authz.ClientID]authz.Client
}

func (r ClientRepo) NewCreds(context.Context) (authz.ClientID, string, error) {
	id, secret := rand.GenerateString(30), rand.GenerateString(50)
	return authz.ClientID(id), secret, nil
}

func (r ClientRepo) NewAccessToken(context.Context) (string, error) {
	return rand.GenerateString(40), nil
}

func (r ClientRepo) NewCode(context.Context) (string, error) {
	code := rand.GenerateInt(900) + 100
	return fmt.Sprint(code), nil
}

func (r ClientRepo) Find(_ context.Context, id authz.ClientID) (authz.Client, bool, error) {
	client, ok := r.clients[id]
	return client, ok, nil
}

func (r *ClientRepo) Save(_ context.Context, client authz.Client) error {
	r.clients[client.ID()] = client

	return nil
}

func (r *ClientRepo) Delete(_ context.Context, client authz.Client) error {
	delete(r.clients, client.ID())

	return nil
}
