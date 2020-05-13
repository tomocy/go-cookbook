package presentation

import (
	jsonPkg "encoding/json"
	"io"

	"github.com/tomocy/go-cookbook/oauth"
)

var JSON json

type json struct{}

func (json) RenderIntrospection(w io.Writer, intro oauth.Introspection) error {
	return jsonPkg.NewEncoder(w).Encode(intro)
}

func (json) RenderErr(w io.Writer, err error) error {
	return jsonPkg.NewEncoder(w).Encode(errResp{
		Error: err.Error(),
	})
}

type errResp struct {
	Error string `json:"error"`
}
