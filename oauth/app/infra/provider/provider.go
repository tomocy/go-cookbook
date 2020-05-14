package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/tomocy/go-cookbook/oauth"
	"github.com/tomocy/go-cookbook/oauth/app"
)

type DefaultUserService struct {
	ResourceServerAddr oauth.Endpoint
}

func (s DefaultUserService) FetchWithAccessToken(ctx context.Context, tok string) (app.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := s.buildFetchWithAccessTokenRequest(ctx, tok)
	if err != nil {
		return app.User{}, fmt.Errorf("failed to build fetch with access token request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return app.User{}, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return app.User{}, fmt.Errorf(resp.Status)
	}

}

func (s DefaultUserService) buildFetchWithAccessTokenRequest(ctx context.Context, tok string) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, s.ResourceServerAddr.URI(PathDefaultOwner), nil)
	if err != nil {
		return nil, err
	}

	oauth.PushBearerAccessToken(r, tok)

	return r, nil
}

const (
	PathDefaultOwner = "default.owner"
)
