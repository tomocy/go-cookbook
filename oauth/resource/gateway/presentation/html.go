package presentation

import (
	"errors"
	"fmt"
	"html/template"
	"io"

	"github.com/tomocy/go-cookbook/oauth/resource"
)

const (
	htmlTemplNewUser    = "user.new"
	htmlTemplSingleUser = "user.single"
	htmlTemplErr        = "error"
)

var HTML = html{
	htmlTemplNewUser:    template.Must(template.ParseFiles("views/html/templates/user/new.html")),
	htmlTemplSingleUser: template.Must(template.ParseFiles("views/html/templates/user/single.html")),
	htmlTemplErr:        template.Must(template.ParseFiles("views/html/templates/error.html")),
}

type html map[string]*template.Template

func (h html) RenderCreateUserPage(w io.Writer) error {
	return h[htmlTemplNewUser].Execute(w, nil)
}

func (h html) RenderUser(w io.Writer, user resource.User) error {
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

func (h html) RenderErr(w io.Writer, err error) error {
	return h.renderErr(w, htmlTemplErr, err)
}

func (h html) renderErr(w io.Writer, name string, err error) error {
	templ, ok := h[name]
	if !ok {
		return fmt.Errorf("no such template")
	}

	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		err = unwrapped
	}
	data := map[string]interface{}{
		"Error": err.Error(),
	}

	return templ.Execute(w, data)
}
