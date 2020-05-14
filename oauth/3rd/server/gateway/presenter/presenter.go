package presenter

import "text/template"

const (
	htmlTemplateFetchOwner = "user.owner.fetch"
)

var HTML = html{
	htmlTemplateFetchOwner: template.Must(template.ParseFiles("views/html/template/user/owner/fetch.html")),
}

type html map[string]*template.Template
