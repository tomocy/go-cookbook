package usecase

import (
	"context"
	"testing"

	"github.com/tomocy/go-cookbook/oauth/authz"
	"github.com/tomocy/go-cookbook/oauth/authz/infra/memory"
)

func TestFindClient(t *testing.T) {
	repo := memory.NewClientRepo()

	create := NewCreateClient(repo)
	client, _ := create.Do("http://localhost")

	u := NewFindClient(repo)
	returned, found, err := u.Do(client.ID())
	if err != nil {
		t.Errorf("should have found client: %s", err)
		return
	}
	if !found {
		t.Errorf("should have found client")
		return
	}
	if err := assertClient(returned, client); err != nil {
		t.Errorf("should have returned the found client: %s", err)
		return
	}
}

func TestCreateClient(t *testing.T) {
	repo := memory.NewClientRepo()

	u := NewCreateClient(repo)

	returned, err := u.Do("http://localhost/1234")
	if err != nil {
		t.Errorf("should have created client: %s", err)
		return
	}

	saved, found, _ := repo.Find(context.Background(), returned.ID())
	if !found {
		t.Errorf("should have saved the created client")
		return
	}
	if err := assertClient(returned, saved); err != nil {
		t.Errorf("should have returned the created client: %s", err)
		return
	}
}

func TestIntrospectAccessToken(t *testing.T) {
	repo := memory.NewClientRepo()

	createClient := NewCreateClient(repo)
	client, _ := createClient.Do("http://localhost/1234")

	accessTok, _ := authz.NewAccessToken("test access token")
	client.StoreAccessToken(accessTok)
	repo.Save(context.Background(), client)

	u := NewIntrospectAccessToken(repo)
	introspectedTok, valid, err := u.Do(client.ID(), accessTok.Token())
	if err != nil {
		t.Errorf("should have introspected access token: %s", err)
		return
	}
	if !valid {
		t.Errorf("should have validated access token")
		return
	}

	savedClient, _, _ := repo.Find(context.Background(), client.ID())
	savedTok, ok := savedClient.AccessToken(introspectedTok.Token())
	if !ok {
		t.Errorf("should have stored the introspected access token")
		return
	}
	if err := assertAccessToken(introspectedTok, savedTok); err != nil {
		t.Errorf("should have returned the introspected access token: %s", err)
		return
	}
}
