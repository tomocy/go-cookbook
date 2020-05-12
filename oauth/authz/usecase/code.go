package usecase

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/authz"
)

func NewGenerateRedirectURIWithAuthzCode(repo authz.ClientRepo) GenerateRedirectURIWithAuthzCode {
	return GenerateRedirectURIWithAuthzCode{
		repo: repo,
	}
}

type GenerateRedirectURIWithAuthzCode struct {
	repo authz.ClientRepo
}

func (u GenerateRedirectURIWithAuthzCode) Do(clientID authz.ClientID) (string, error) {
	ctx := context.TODO()

	client, found, err := u.repo.Find(ctx, clientID)
	if err != nil {
		return "", fmt.Errorf("failed to find client: %w", err)
	}
	if !found {
		return "", fmt.Errorf("no such client")
	}

	rawCode, err := u.repo.NewCode(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to generate authz code: %w", err)
	}
	code, err := authz.NewCode(rawCode)
	if err != nil {
		return "", fmt.Errorf("failed to generate authz code: %w", err)
	}

	if err := client.StoreCode(code); err != nil {
		return "", fmt.Errorf("failed to store authz code: %w", err)
	}
	if err := u.repo.Save(ctx, client); err != nil {
		return "", fmt.Errorf("failed to save client: %w", err)
	}

	uri, err := client.RedirectURIWithAuthzCode(code)
	if err != nil {
		return "", fmt.Errorf("failed to generate redirect uri with authz code: %w", err)
	}

	return uri, nil
}

func NewGenerateAccessToken(repo authz.ClientRepo) GenerateAccessToken {
	return GenerateAccessToken{
		repo: repo,
	}
}

type GenerateAccessToken struct {
	repo authz.ClientRepo
}

func (u GenerateAccessToken) Do(clientID authz.ClientID, rawCode string) (authz.AccessToken, error) {
	ctx := context.TODO()

	client, ok, err := u.repo.Find(ctx, clientID)
	if err != nil {
		return authz.AccessToken{}, fmt.Errorf("failed to find client: %w", err)
	}
	if !ok {
		return authz.AccessToken{}, fmt.Errorf("no such client")
	}

	code, ok := client.Code(rawCode)
	if !ok {
		return authz.AccessToken{}, fmt.Errorf("invalid code")
	}
	if code.IsExpired() {
		return authz.AccessToken{}, fmt.Errorf("expired code")
	}

	rawTok, err := u.repo.NewAccessToken(ctx)
	if err != nil {
		return authz.AccessToken{}, fmt.Errorf("failed to generate access token: %w", err)
	}
	tok, err := authz.NewAccessToken(rawTok)
	if err != nil {
		return authz.AccessToken{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	if err := client.StoreAccessToken(tok); err != nil {
		return authz.AccessToken{}, fmt.Errorf("failed to store access token: %w", err)
	}
	if err := u.repo.Save(ctx, client); err != nil {
		return authz.AccessToken{}, fmt.Errorf("failed to save client: %w", err)
	}

	return tok, nil
}
