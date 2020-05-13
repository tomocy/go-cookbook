package presentation

import (
	jsonPkg "encoding/json"
	"io"

	"github.com/tomocy/go-cookbook/oauth/dto"
)

var JSON json

type json struct{}

func (json) RenderIntrospection(w io.Writer, intro dto.Introspection) error {
	return jsonPkg.NewEncoder(w).Encode(intro)
}

func (json) RenderAccessToken(w io.Writer, tok dto.AccessToken) error {
	return jsonPkg.NewEncoder(w).Encode(tok)
}

func (json) RenderErr(w io.Writer, err error) error {
	return jsonPkg.NewEncoder(w).Encode(errResp{
		Error: err.Error(),
	})
}

type errResp struct {
	Error string `json:"error"`
}
