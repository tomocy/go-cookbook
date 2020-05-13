package presentation

import (
	"html/template"
	"io"
)

const (
	htmlTemplFetchUser = "user.fetch"
	htmlTemplErr       = "error"
)

var HTML = html{
	htmlTemplFetchUser: template.Must(template.ParseFiles("views/html/templates/user/fetch.html")),
	htmlTemplErr:       template.Must(template.ParseFiles("views/html/templates/error.html")),
}

type html map[string]*template.Template

func (h html) RenderFetchUserPage(w io.Writer) error {
	return h[htmlTemplFetchUser].Execute(w, nil)
}

func (h html) RenderErr(w io.Writer, err error) error {
	return h[htmlTemplErr].Execute(w, map[string]interface{}{
		"Error": err.Error(),
	})
}
