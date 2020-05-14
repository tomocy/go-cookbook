package presentation

import (
	"html/template"
	"io"
)

const (
	htmlTemplFetchOwner = "user.owner.fetch"
	htmlTemplErr        = "error"
)

var HTML = html{
	htmlTemplFetchOwner: template.Must(template.ParseFiles("views/html/templates/user/owner/fetch.html")),
	htmlTemplErr:        template.Must(template.ParseFiles("views/html/templates/error.html")),
}

type html map[string]*template.Template

func (h html) RenderFetchOwnerPage(w io.Writer) error {
	return h[htmlTemplFetchOwner].Execute(w, nil)
}

func (h html) RenderErr(w io.Writer, err error) error {
	return h[htmlTemplErr].Execute(w, map[string]interface{}{
		"Error": err.Error(),
	})
}
