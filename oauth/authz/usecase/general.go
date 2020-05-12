package usecase

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/authz"
)

func NewFindClient(repo authz.ClientRepo) FindClient {
	return FindClient{
		repo: repo,
	}
}

type FindClient struct {
	repo authz.ClientRepo
}

func (u FindClient) Do(id authz.ClientID) (authz.Client, bool, error) {
	ctx := context.TODO()

	client, found, err := u.repo.Find(ctx, id)
	if err != nil {
		return authz.Client{}, false, fmt.Errorf("failed to find client: %w", err)
	}

	return client, found, nil
}

func NewCreateClient(repo authz.ClientRepo) CreateClient {
	return CreateClient{
		repo: repo,
	}
}

type CreateClient struct {
	repo authz.ClientRepo
}

func (u CreateClient) Do(redirectURI string) (authz.Client, error) {
	ctx := context.TODO()

	id, secret, err := u.repo.NewCreds(ctx)
	if err != nil {
		return authz.Client{}, fmt.Errorf("failed to create client credentials: %w", err)
	}

	client, err := authz.NewClient(id, secret, redirectURI)
	if err != nil {
		return authz.Client{}, fmt.Errorf("failed to create client: %w", err)
	}

	if err := u.repo.Save(ctx, client); err != nil {
		return authz.Client{}, fmt.Errorf("failed to save client: %w", err)
	}

	return client, nil
}

func NewIntrospectAccessToken(repo authz.ClientRepo) IntrospectAccessToken {
	return IntrospectAccessToken{
		repo: repo,
	}
}

type IntrospectAccessToken struct {
	repo authz.ClientRepo
}

func (u IntrospectAccessToken) Do(clientID authz.ClientID, rawTok string) (authz.AccessToken, bool, error) {
	ctx := context.TODO()

	client, ok, err := u.repo.Find(ctx, clientID)
	if err != nil {
		return authz.AccessToken{}, false, fmt.Errorf("failed to find client: %w", err)
	}
	if !ok {
		return authz.AccessToken{}, false, nil
	}

	tok, ok := client.AccessToken(rawTok)
	if !ok {
		return authz.AccessToken{}, false, nil
	}
	if tok.IsExpired() {
		return authz.AccessToken{}, false, nil
	}

	return tok, true, nil
}
