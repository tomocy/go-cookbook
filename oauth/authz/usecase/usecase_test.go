package usecase

import (
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/authz"
)

func assertClient(actual, expected authz.Client) error {
	if actual.ID() != expected.ID() {
		return reportUnexpected("id", actual.ID(), expected.ID())
	}

	return nil
}

func assertCode(actual, expected authz.Code) error {
	if actual.Code() != expected.Code() {
		return reportUnexpected("code", actual.Code(), expected.Code())
	}

	return nil
}

func assertAccessToken(actual, expected authz.AccessToken) error {
	if actual.Token() != expected.Token() {
		return reportUnexpected("token", actual.Token(), expected.Token())
	}

	return nil
}

func assertUser(actual, expected authz.User) error {
	if actual.ID() != expected.ID() {
		return reportUnexpected("id", actual.ID(), expected.ID())
	}

	return nil
}

func reportUnexpected(name string, actual, expected interface{}) error {
	return fmt.Errorf("unexpected %s: got %v, but expected %v", name, actual, expected)
}
