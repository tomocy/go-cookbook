package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomocy/go-cookbook/oauth/users"
	"github.com/tomocy/go-cookbook/oauth/users/usecase"
)

func NewServer(w io.Writer, addr string, userRepo users.UserRepo, ren Renderer) Server {
	return Server{
		w:        w,
		addr:     addr,
		userRepo: userRepo,
		renderer: ren,
	}
}

type Server struct {
	w        io.Writer
	addr     string
	userRepo users.UserRepo
	renderer Renderer
}

func (s Server) Run() error {
	r := chi.NewRouter()

	r.Route("/users", func(r chi.Router) {
		r.Post("/", handlerFunc(s.createUser()))

		r.Route("/authns", func(r chi.Router) {
			r.Post("/", handlerFunc(s.authenticateUser()))
		})
	})

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

func handlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func (s Server) createUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			s.renderErr(w, "parse form", err)
			return
		}

		email, pass := r.PostFormValue("email"), r.PostFormValue("password")
		create := usecase.NewCreateUser(s.userRepo)
		user, err := create.Do(email, pass)
		if err != nil {
			s.renderErr(w, "create user", err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := s.renderer.RenderUser(w, user); err != nil {
			s.renderErr(w, "render user", err)
		}
	})
}

func (s Server) authenticateUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			s.renderErr(w, "parse form", err)
			return
		}

		email, pass := r.PostFormValue("email"), r.PostFormValue("password")
		authenticate := usecase.NewAuthenticateUser(s.userRepo)
		user, ok, err := authenticate.Do(email, pass)
		if err != nil {
			s.renderErr(w, "authenticate user", err)
			return
		}
		if !ok {
			s.renderErrMessage(w, http.StatusBadRequest, "invalid credentials")
			return
		}

		if err := s.renderer.RenderUser(w, user); err != nil {
			s.renderErr(w, "render user", err)
			return
		}
	})
}

func (s Server) logfln(format string, as ...interface{}) {
	s.logf(format+"\n", as...)
}

func (s Server) logf(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format, as...)
}

func (s Server) renderErr(w http.ResponseWriter, did string, err error) {
	if users.IsErrInput(err) {
		s.renderErrMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	s.logfln("failed to %s: %s", did, err)
	s.renderErrStatus(w, http.StatusInternalServerError)
}

func (s Server) renderErrStatus(w http.ResponseWriter, code int) {
	s.renderErrMessage(w, code, http.StatusText(code))
}

func (s Server) renderErrMessage(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	if err := s.renderer.RenderErr(w, fmt.Errorf("%s", msg)); err != nil {
		s.logfln("failed to render err: %s", err)
	}
}

type Renderer interface {
	RenderUser(io.Writer, users.User) error
	RenderErr(io.Writer, error) error
}
