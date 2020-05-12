package presentation

import (
	"fmt"
	"html/template"
	"io"

	"github.com/tomocy/go-cookbook/oauth/resource"
)

const (
	htmlTemplSingleUser = "user.single"
)

var HTML = html{
	htmlTemplSingleUser: template.Must(template.ParseFiles("views/html/templates/user/single.html")),
}

type html map[string]*template.Template

func (h html) RenderCreatedUser(w io.Writer, user resource.User) error {
	return h.renderUser(w, user)
}

func (h html) renderUser(w io.Writer, user resource.User) error {
	data := map[string]interface{}{
		"User": htmlSingleUser{
			ID:    string(user.ID()),
			Name:  user.Name(),
			Email: user.Email(),
		},
	}

	return h[htmlTemplSingleUser].Execute(w, data)
}

type htmlSingleUser struct {
	ID    string
	Name  string
	Email string
}

func (h html) RenderCreateUserErr(w io.Writer, err error) error {
	return h.renderErr(w, htmlTemplSingleUser, err)
}

func (h html) renderErr(w io.Writer, name string, err error) error {
	templ, ok := h[name]
	if !ok {
		return fmt.Errorf("no such template")
	}

	data := map[string]interface{}{
		"Error": err.Error(),
	}

	return templ.Execute(w, data)
}
