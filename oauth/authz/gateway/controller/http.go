package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomocy/go-cookbook/oauth"
	"github.com/tomocy/go-cookbook/oauth/authz"
	"github.com/tomocy/go-cookbook/oauth/dto"
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

	r.Post("/tokens", handlerFunc(s.createAccessToken()))
	r.Post("/introspections", handlerFunc(s.introspect()))

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

func handlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

const (
	grantTypeAuthzCode   = "authorization_code"
	grantTypeClientCreds = "client_credentials"
)

func (s HTTPServer) createAccessToken() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			s.renderErr(w, "parse form", err)
			return
		}

		grantType := r.PostFormValue("grant_type")
		switch grantType {
		case grantTypeAuthzCode:
			s.createAccessTokenInAuthzCodeFlow(w, r)
			return
		case grantTypeClientCreds:
			s.createAccessTokenInClientCredsFlow(w, r)
			return
		default:
			s.renderErrStatus(w, http.StatusBadRequest)
			return
		}
	})
}

func (s HTTPServer) createAccessTokenInAuthzCodeFlow(w http.ResponseWriter, r *http.Request) {
	code := r.PostFormValue("code")
	if code != "123" {
		s.renderErrStatus(w, http.StatusBadRequest)
		return
	}

	if err := s.renderer.RenderAccessToken(w, dto.AccessToken{
		AccessToken: "aiueo_app_access_token",
	}); err != nil {
		s.renderErr(w, "render access token", err)
		return
	}
}

func (s HTTPServer) createAccessTokenInClientCredsFlow(w http.ResponseWriter, r *http.Request) {
	clientID, clientSec := r.URL.Query().Get("client_id"), r.URL.Query().Get("client_secret")
	if clientID != "aiueo_id" || clientSec != "aiueo_secret" {
		s.renderErrStatus(w, http.StatusNotFound)
		return
	}

	if err := s.renderer.RenderAccessToken(w, dto.AccessToken{
		AccessToken: "aiueo_client_access_token",
	}); err != nil {
		s.renderErr(w, "render access token", err)
		return
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
			if tok != "aiueo_app_access_token" {
				if err := s.renderer.RenderIntrospection(w, dto.Introspection{
					Active: false,
				}); err != nil {
					s.renderErr(w, "render introspection", err)
					return
				}
			}

			w.WriteHeader(http.StatusOK)
			if err := s.renderer.RenderIntrospection(w, dto.Introspection{
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
		if tok != "aiueo_client_access_token" {
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
	RenderIntrospection(io.Writer, dto.Introspection) error
	RenderAccessToken(io.Writer, dto.AccessToken) error
	RenderErr(io.Writer, error) error
}
