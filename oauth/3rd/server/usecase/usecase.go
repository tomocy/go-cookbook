package usecase

import "fmt"

type DoHaveAccessToken struct{}

func (u DoHaveAccessToken) Do(id string) (bool, error) {
	return id != "test_user_id", nil
}

type GenerateAuthzCodeURI struct{}

func (u GenerateAuthzCodeURI) Do(id string) (string, error) {
	return fmt.Sprintf("http://localhost:8080?client_id=test_client_id&client_secret=test_client_secret"), nil
}
