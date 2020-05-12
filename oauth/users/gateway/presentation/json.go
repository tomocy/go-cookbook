package presentation

import (
	"io"

	jsonPkg "encoding/json"

	"github.com/tomocy/go-cookbook/oauth/users"
)

var JSON = json{}

type json struct{}

func (json) RenderUser(w io.Writer, user users.User) error {
	return jsonPkg.NewEncoder(w).Encode(userResp{
		ID: string(user.ID()),
	})
}

type userResp struct {
	ID string `json:"id"`
}

func (json) RenderErr(w io.Writer, err error) error {
	return jsonPkg.NewEncoder(w).Encode(errResp{
		Error: err.Error(),
	})
}

type errResp struct {
	Error string `json:"error"`
}
