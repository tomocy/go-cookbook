package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tomocy/go-cookbook/oauth/dto"
)

type AuthzCodeClient struct {
	AuthzServerEndpoint Endpoint
	Creds               Creds
}

func (c AuthzCodeClient) ExchangeAccessToken(ctx context.Context, code string) (AccessToken, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := c.buildExchangeAccessTokenRequest(ctx, code)
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to build exhange access token request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return AccessToken{}, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AccessToken{}, fmt.Errorf(resp.Status)
	}

	var tok dto.AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return AccessToken{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return AccessToken{
		Token: tok.AccessToken,
	}, nil
}

func (c AuthzCodeClient) buildExchangeAccessTokenRequest(ctx context.Context, code string) (*http.Request, error) {
	vals := url.Values{
		"grant_type": []string{"authorization_code"},
		"code":       []string{code},
	}
	body := strings.NewReader(vals.Encode())
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.AuthzServerEndpoint.URI(PathToken), body)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return r, nil
}

type ClientCredentialsService interface {
	Introspector
}

type Introspector interface {
	Introspect(context.Context, string) (AccessToken, bool, error)
}

type Client struct {
	AuthzServerEndpoint Endpoint
	Creds               Creds
	tok                 string
}

func (c *Client) fetchAccessTokenIfNone(ctx context.Context) error {
	if c.tok != "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := c.buildAccessTokenRequest(ctx)
	if err != nil {
		return fmt.Errorf("failed to build access token request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(resp.Status)
	}

	var tok dto.AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	c.tok = tok.AccessToken

	return nil
}

func (c Client) buildAccessTokenRequest(ctx context.Context) (*http.Request, error) {
	vals := url.Values{
		"grant_type": []string{"client_credentials"},
	}
	body := strings.NewReader(vals.Encode())
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.AuthzServerEndpoint.URI(PathToken), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	queries := url.Values{
		"client_id":     []string{c.Creds.ID},
		"client_secret": []string{c.Creds.Secret},
	}
	r.URL.RawQuery = queries.Encode()

	return r, nil
}

func (c *Client) Introspect(ctx context.Context, tok string) (AccessToken, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.fetchAccessTokenIfNone(ctx); err != nil {
		return AccessToken{}, false, fmt.Errorf("failed to fetch access token: %w", err)
	}

	req, err := c.buildIntrospectAccessTokenRequest(ctx, tok)
	if err != nil {
		return AccessToken{}, false, fmt.Errorf("failed to build introspect access token request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return AccessToken{}, false, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AccessToken{}, false, fmt.Errorf(resp.Status)
	}

	var intro dto.Introspection
	if err := json.NewDecoder(resp.Body).Decode(&intro); err != nil {
		return AccessToken{}, false, err
	}

	if !intro.Active {
		return AccessToken{}, false, err
	}

	return AccessToken{
		Token:  tok,
		UserID: intro.Username,
	}, true, nil
}

func (c Client) buildIntrospectAccessTokenRequest(ctx context.Context, tok string) (*http.Request, error) {
	vals := url.Values{
		"token": []string{tok},
	}
	body := strings.NewReader(vals.Encode())
	uri := c.composeIntrospectionsURI()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.tok))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func (c Client) composeIntrospectionsURI() string {
	addr := compensateHTTPAddr(c.AuthzServerEndpoint.Addr)
	return fmt.Sprintf("%s/%s", addr, strings.TrimLeft(c.AuthzServerEndpoint.Paths[PathIntrospection], "/"))
}

const (
	PathAuthz         = "authz"
	PathIntrospection = "introspection"
	PathToken         = "token"
)

type Endpoint struct {
	Addr  string
	Paths map[string]string
}

func (e Endpoint) URI(name string) string {
	addr := compensateHTTPAddr(e.Addr)
	return fmt.Sprintf("%s/%s", addr, strings.TrimLeft(e.Paths[name], "/"))
}

func compensateHTTPAddr(addr string) string {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = fmt.Sprintf("http://%s", addr)
	}

	return addr
}

type Creds struct {
	ID     string
	Secret string
}

type Log struct {
	Out io.Writer
}

func (l Log) Introspect(_ context.Context, tok string) (AccessToken, bool, error) {
	l.logfln("target access token: %s", tok)
	return AccessToken{}, false, nil
}

func (l Log) logfln(format string, as ...interface{}) {
	l.logf(format+"\n", as...)
}

func (l Log) logf(format string, as ...interface{}) {
	fmt.Fprintf(l.Out, format, as...)
}

type AccessToken struct {
	UserID string
	Token  string
}

func PushBearerAccessToken(r *http.Request, tok string) {
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok))
}

func ExtractBearerAccessToken(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	var tok string
	fmt.Sscanf(authz, "Bearer %s", &tok)

	return tok
}
