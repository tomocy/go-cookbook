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
		r.Get("/new", handlerFunc(s.showCreateUserPage()))
		r.Get("/{id}", handlerFunc(s.user()))
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

func (s HTTPServer) showCreateUserPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.renderer.RenderCreateUserPage(w); err != nil {
			s.renderErr(w, "render create user page", err)
			return
		}
	})
}

func (s HTTPServer) user() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := resource.UserID(chi.URLParam(r, "id"))
		find := usecase.NewFindUser(s.userRepo)
		user, found, err := find.Do(id)
		if err != nil {
			s.renderErr(w, "find user", err)
			return
		}
		if !found {
			s.renderErrMessage(w, http.StatusNotFound, "no such user")
			return
		}

		if err := s.renderer.RenderUser(w, user); err != nil {
			s.renderErr(w, "render user", err)
			return
		}
	})
}

func (s HTTPServer) createUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			s.renderErr(w, "parse form", err)
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
			s.renderErr(w, "create user", err)
			return
		}

		s.redirectToUserPage(w, user)
	})
}

func (s HTTPServer) renderErr(w http.ResponseWriter, did string, err error) {
	if resource.IsErrInput(err) {
		s.renderErrMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	s.logfln("failed to %s: %s", did, err)
	s.renderErrStatus(w, http.StatusInternalServerError)
}

func (s HTTPServer) renderErrStatus(w http.ResponseWriter, code int) {
	s.renderErrMessage(w, code, http.StatusText(code))
}

func (s HTTPServer) renderErrMessage(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	if err := s.renderer.RenderErr(w, fmt.Errorf(msg)); err != nil {
		s.logfln("failed to render error with message: %s", err)
		return
	}
}

func (s HTTPServer) redirectToUserPage(w http.ResponseWriter, user resource.User) {
	s.redirect(w, fmt.Sprintf("/users/%s", user.ID()))
}

func (s HTTPServer) redirect(w http.ResponseWriter, loc string) {
	w.Header().Set("Location", loc)
	w.WriteHeader(http.StatusFound)
}

func (s HTTPServer) logfln(format string, as ...interface{}) {
	s.logf(format+"\n", as...)
}

func (s HTTPServer) logf(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format, as...)
}

type httpServerRenderer interface {
	RenderCreateUserPage(io.Writer) error
	RenderUser(io.Writer, resource.User) error
	RenderErr(io.Writer, error) error
}
