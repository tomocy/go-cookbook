package presenter

import (
	"io"
	"text/template"
)

const (
	htmlTemplateFetchOwner = "user.owner.fetch"
)

var HTML = html{
	htmlTemplateFetchOwner: template.Must(template.ParseFiles("views/html/templates/user/owner/fetch.html")),
}

type html map[string]*template.Template

func (h html) ShowFetchOwnerPage(w io.Writer) error {
	return h[htmlTemplateFetchOwner].Execute(w, nil)
}
