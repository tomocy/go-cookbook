package usecase

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/tomocy/go-cookbook/oauth/authz/infra/memory"
)

func TestGenerateRedirectURIWithAuthzCode(t *testing.T) {
	repo := memory.NewClientRepo()

	createClient := NewCreateClient(repo)
	redirectURI := "http://localhost/1234"
	client, _ := createClient.Do(redirectURI)

	u := NewGenerateRedirectURIWithAuthzCode(repo)
	uri, err := u.Do(client.ID())
	if err != nil {
		t.Errorf("should have generated authz code: %s", err)
		return
	}

	if !strings.HasPrefix(uri, "http://") && strings.HasPrefix(uri, "https://") {
		uri = "http://" + uri
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		t.Errorf("should have generated the valid redirect uri: %s", uri)
		return
	}

	code := parsed.Query().Get("code")
	if code == "" {
		t.Errorf("should have generate the valid redirect uri with authz code")
		return
	}

	client, _, _ = repo.Find(context.Background(), client.ID())
	if _, ok := client.Code(code); !ok {
		t.Errorf("should have saved the generated authz code")
		return
	}
}

func TestGenerateAccessToken(t *testing.T) {
	repo := memory.NewClientRepo()

	createClient := NewCreateClient(repo)
	client, _ := createClient.Do("http://localhost/1234")

	generateAuthzCode := NewGenerateRedirectURIWithAuthzCode(repo)
	uri, _ := generateAuthzCode.Do(client.ID())
	parsed, _ := url.Parse(uri)
	code := parsed.Query().Get("code")

	u := NewGenerateAccessToken(repo)
	tok, err := u.Do(client.ID(), code)
	if err != nil {
		t.Errorf("should have generate access token: %s", err)
		return
	}

	savedClient, _, _ := repo.Find(context.Background(), client.ID())
	savedTok, ok := savedClient.AccessToken(tok.Token())
	if !ok {
		t.Errorf("should have saved the generated access token")
		return
	}
	if err := assertAccessToken(tok, savedTok); err != nil {
		t.Errorf("should have returned the generated access token: %s", err)
		return
	}
}
