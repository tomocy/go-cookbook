package presenter

import (
	"io"
	"text/template"

	"github.com/tomocy/go-cookbook/oauth/3rd/server"
)

const (
	htmlTemplateFetchOwner = "owner.fetch"
	htmlTemplateOwner      = "owner.single"
	htmlTemplateErr        = "error"
)

var HTML = html{
	htmlTemplateFetchOwner: template.Must(template.ParseFiles("views/html/templates/owner/fetch.html")),
	htmlTemplateOwner:      template.Must(template.ParseFiles("views/html/templates/owner/single.html")),
	htmlTemplateErr:        template.Must(template.ParseFiles("views/html/templates/error.html")),
}

type html map[string]*template.Template

func (h html) ShowFetchOwnerPage(w io.Writer) error {
	return h[htmlTemplateFetchOwner].Execute(w, nil)
}

func (h html) ShowOwner(w io.Writer, o server.Owner) error {
	return h[htmlTemplateOwner].Execute(w, map[string]interface{}{
		"Owner": o,
	})
}

func (h html) ShowErr(w io.Writer, err error) error {
	return h[htmlTemplateErr].Execute(w, map[string]interface{}{
		"Error": err.Error(),
	})
}
