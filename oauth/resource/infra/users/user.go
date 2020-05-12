package users

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tomocy/go-cookbook/oauth/resource"
)

func NewHTTPService(createEndpoint string) HTTPService {
	return HTTPService{
		createEndpoint: createEndpoint,
	}
}

type HTTPService struct {
	createEndpoint string
}

func (s HTTPService) Create(ctx context.Context, email, pass string) (resource.UserID, error) {
	ctx, cancel := context.WithTimeout(ctx, 0)
	defer cancel()

	req, err := s.buildCreateRequest(ctx, email, pass)
	if err != nil {
		return "", fmt.Errorf("failed to build create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to do create request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errResp, err := s.decodeErrResp(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to decode error response: %w", err)
		}

		return "", errResp
	}

	userResp, err := s.decodeUserResp(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to decode user response: %w", err)
	}

	return userResp.ID, nil
}

func (s HTTPService) buildCreateRequest(ctx context.Context, email, pass string) (*http.Request, error) {
	vals := url.Values{
		"email":    []string{email},
		"password": []string{pass},
	}
	body := strings.NewReader(vals.Encode())
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, s.createEndpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	r.Header.Set("Content-Type", "x-www-form-urlencode")

	return r, nil
}

func (s HTTPService) decodeUserResp(body io.Reader) (userResp, error) {
	var decoded userResp
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return userResp{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return decoded, nil
}

func (s HTTPService) decodeErrResp(body io.Reader) (errResp, error) {
	var decoded errResp
	if err := json.NewDecoder(body).Decode(&decoded); err != nil {
		return errResp{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return decoded, nil
}

type userResp struct {
	ID resource.UserID `json:"id"`
}

type errResp struct {
	Message string `json:"error"`
}

func (r errResp) Error() string {
	return r.Message
}
