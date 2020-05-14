package usecase

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/3rd/server"
)

type DoHaveAccessToken struct{}

func (u DoHaveAccessToken) Do(id string) (bool, error) {
	return id == "test_user_id", nil
}

type GenerateAuthzCodeURI struct{}

func (u GenerateAuthzCodeURI) Do(id string) (string, error) {
	return fmt.Sprintf("http://localhost:8080?client_id=test_client_id&client_secret=test_client_secret"), nil
}

type FetchOwner struct {
	Service server.OwnerService
}

func (u FetchOwner) Do(id string) (server.Owner, error) {
	ctx := context.TODO()

	return u.Service.Fetch(ctx, "test_access_token")
}
