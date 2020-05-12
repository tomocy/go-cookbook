package json

import (
	"encoding/json"
	"io"

	"github.com/tomocy/go-cookbook/oauth/users"
)

var Renderer = renderer{}

type renderer struct{}

func (renderer) RenderUser(w io.Writer, user users.User) error {
	return json.NewEncoder(w).Encode(userResp{
		ID: string(user.ID()),
	})
}

type userResp struct {
	ID string `json:"id"`
}

func (renderer) RenderErr(w io.Writer, err error) error {
	return json.NewEncoder(w).Encode(errResp{
		Error: err.Error(),
	})
}

type errResp struct {
	Error string `json:"error"`
}
