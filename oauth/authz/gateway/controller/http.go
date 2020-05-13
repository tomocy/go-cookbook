package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomocy/go-cookbook/oauth"
	"github.com/tomocy/go-cookbook/oauth/authz"
)

func NewHTTPServer(w io.Writer, addr string, ren httpServerRenderer) HTTPServer {
	return HTTPServer{
		w:        w,
		addr:     addr,
		renderer: ren,
	}
}

type HTTPServer struct {
	w        io.Writer
	addr     string
	renderer httpServerRenderer
}

func (s HTTPServer) Run() error {
	r := chi.NewRouter()

	r.Post("/introspections", handlerFunc(s.introspect()))

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

func handlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func (s HTTPServer) introspect() http.Handler {
	return s.requireValidAccessToken(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				s.renderErr(w, "parse form", err)
				return
			}

			tok := r.PostFormValue("token")
			if tok != "aiueo_app_token" {
				if err := s.renderer.RenderIntrospection(w, oauth.Introspection{
					Active: false,
				}); err != nil {
					s.renderErr(w, "render introspection", err)
					return
				}
			}

			w.WriteHeader(http.StatusOK)
			if err := s.renderer.RenderIntrospection(w, oauth.Introspection{
				Active:   true,
				Username: "aiueo",
			}); err != nil {
				s.renderErr(w, "render introspection", err)
				return
			}
		}),
	)
}

func (s HTTPServer) requireValidAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next == nil {
			return
		}

		tok := oauth.ExtractBearerAccessToken(r)
		if tok != "aiueo_client_token" {
			s.renderErrStatus(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s HTTPServer) renderErr(w http.ResponseWriter, did string, err error) {
	if authz.IsErrInput(err) {
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
	if err := s.renderer.RenderErr(w, fmt.Errorf("%s", msg)); err != nil {
		s.logfln("failed to render err: %s", err)
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
	RenderIntrospection(io.Writer, oauth.Introspection) error
	RenderErr(io.Writer, error) error
}
