package presenter

import (
	"io"
	"text/template"
)

const (
	htmlTemplateFetchOwner = "user.owner.fetch"
	htmlTemplateErr        = "error"
)

var HTML = html{
	htmlTemplateFetchOwner: template.Must(template.ParseFiles("views/html/templates/user/owner/fetch.html")),
	htmlTemplateErr:        template.Must(template.ParseFiles("views/html/templates/error.html")),
}

type html map[string]*template.Template

func (h html) ShowFetchOwnerPage(w io.Writer) error {
	return h[htmlTemplateFetchOwner].Execute(w, nil)
}

func (h html) ShowErr(w io.Writer, err error) error {
	return h[htmlTemplateErr].Execute(w, map[string]interface{}{
		"Error": err.Error(),
	})
}
