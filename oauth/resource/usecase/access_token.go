package usecase

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/resource"
)

func NewIntrospectAccessToken(repo resource.AccessTokenRepo) IntrospectAccessToken {
	return IntrospectAccessToken{
		repo: repo,
	}
}

type IntrospectAccessToken struct {
	repo resource.AccessTokenRepo
}

func (u IntrospectAccessToken) Do(clientID, clientSecret, rawTok string) (resource.AccessToken, bool, error) {
	ctx := context.TODO()

	tok, valid, err := u.repo.Introspect(ctx, clientID, clientSecret, rawTok)
	if err != nil {
		return resource.AccessToken{}, false, fmt.Errorf("failed to introspect access token: %w", err)
	}

	return tok, valid, nil
}
