package aiueo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tomocy/go-cookbook/oauth/3rd/server"
)

var OwnerService ownerService

type ownerService struct{}

func (s ownerService) Fetch(ctx context.Context, tok string) (server.Owner, error) {
	req, err := s.buildFetchRequest(ctx, tok)
	if err != nil {
		return server.Owner{}, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return server.Owner{}, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return server.Owner{}, fmt.Errorf(resp.Status)
	}

	var decoded ownerResp
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return server.Owner{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return server.Owner{
		Name:     decoded.Name,
		Provider: "aiueo",
	}, nil
}

func (ownerService) buildFetchRequest(ctx context.Context, tok string) (*http.Request, error) {
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8081/owners", nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok))

	return r, nil
}

type ownerResp struct {
	Name string `json:"name"`
}
