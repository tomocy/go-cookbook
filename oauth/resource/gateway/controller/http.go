package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomocy/go-cookbook/oauth/resource"
	"github.com/tomocy/go-cookbook/oauth/resource/usecase"
)

func NewHTTPServer(
	w io.Writer, addr string,
	ren httpServerRenderer,
	userServ resource.UserService, userRepo resource.UserRepo,
) HTTPServer {
	return HTTPServer{
		w:        w,
		addr:     addr,
		renderer: ren,
		userServ: userServ,
		userRepo: userRepo,
	}
}

type HTTPServer struct {
	w        io.Writer
	addr     string
	renderer httpServerRenderer
	userServ resource.UserService
	userRepo resource.UserRepo
}

func (s HTTPServer) Run() error {
	r := chi.NewRouter()

	r.Route("/users", func(r chi.Router) {
		r.Post("/", handlerFunc(s.createUser()))
	})

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

func handlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func (s HTTPServer) createUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			s.renderErr(s.renderer.RenderCreateUserErr, w, "parse form", err)
			return
		}

		var (
			name  = r.PostFormValue("name")
			email = r.PostFormValue("email")
			pass  = r.PostFormValue("password")
		)
		create := usecase.NewCreateUser(s.userServ, s.userRepo)
		user, err := create.Do(name, email, pass)
		if err != nil {
			s.renderErr(s.renderer.RenderCreateUserErr, w, "create user", err)
			return
		}

		if err := s.renderer.RenderCreatedUser(w, user); err != nil {
			s.renderErr(s.renderer.RenderCreateUserErr, w, "render user", err)
		}
	})
}

func (s HTTPServer) renderErr(render errRenderer, w http.ResponseWriter, did string, err error) {
	if resource.IsErrInput(err) {
		s.renderErrMessage(render, w, http.StatusBadRequest, err.Error())
		return
	}

	s.logfln("failed to %s: %s", did, err)
	s.renderErrStatus(render, w, http.StatusInternalServerError)
}

func (s HTTPServer) renderErrStatus(render errRenderer, w http.ResponseWriter, code int) {
	s.renderErrMessage(render, w, code, http.StatusText(code))
}

func (s HTTPServer) renderErrMessage(render errRenderer, w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	if err := render(w, fmt.Errorf(msg)); err != nil {
		s.logfln("failed to render error with message: %s", err)
		return
	}
}

func (s HTTPServer) logfln(format string, as ...interface{}) {
	s.logf(format+"\n", as...)
}

func (s HTTPServer) logf(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format, as...)
}

type httpServerRenderer interface {
	RenderCreatedUser(io.Writer, resource.User) error
	RenderCreateUserErr(io.Writer, error) error
}

type errRenderer func(io.Writer, error) error
