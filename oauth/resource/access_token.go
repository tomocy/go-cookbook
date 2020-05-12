package resource

import "context"

type AccessTokenRepo interface {
	Introspect(context.Context, string, string, string) (AccessToken, bool, error)
}

type AccessToken struct {
	token  string
	userID UserID
}

func (t AccessToken) UserID() UserID {
	return t.userID
}
