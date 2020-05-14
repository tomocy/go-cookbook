package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomocy/go-cookbook/oauth"
	"github.com/tomocy/go-cookbook/oauth/app"
	"github.com/tomocy/go-cookbook/oauth/app/usecase"
)

func NewHTTPServer(
	w io.Writer, addr string,
	oauthClient oauth.AuthzCodeClient, ren httpServerRenderer,
	userRepo app.UserRepo,
) HTTPServer {
	return HTTPServer{
		w:           w,
		addr:        addr,
		oauthClient: oauthClient,
		renderer:    ren,
		userRepo:    userRepo,
	}
}

type HTTPServer struct {
	w           io.Writer
	addr        string
	oauthClient oauth.AuthzCodeClient
	renderer    httpServerRenderer
	userRepo    app.UserRepo
}

func (s HTTPServer) Run() error {
	r := chi.NewRouter()

	r.Get("/", handlerFunc(s.showFetchOwnerPage()))
	r.Get("/authzs/{provider}", handlerFunc(s.exchangeAccessToken()))

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

func handlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func (s HTTPServer) showFetchOwnerPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.renderer.RenderFetchOwnerPage(w); err != nil {
			s.renderErr(w, "render fetch user page", err)
			return
		}
	})
}

func (s HTTPServer) exchangeAccessToken() http.Handler {
	return s.keepRedirectURI(
		s.requireAuthenticated(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.TODO()

				userID := s.popAuthenticatedUserID(r)
				providerName := chi.URLParam(r, "provider")
				tok, err := s.oauthClient.ExchangeAccessToken(ctx, "123")
				if err != nil {
					s.renderErr(w, "exchange access token", err)
					return
				}

				add := usecase.NewAddProvider(s.userRepo)
				if err := add.Do(userID, providerName, tok.Token); err != nil {
					s.renderErr(w, "add provider", err)
					return
				}

				s.redirectToKeptLocation(w, r)
			}),
		),
	)
}

func (s HTTPServer) keepRedirectURI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next == nil {
			return
		}

		uri := r.URL.Query().Get("redirect_uri")
		r = s.pushRedirectURI(r, uri)

		next.ServeHTTP(w, r)
	})
}

func (s HTTPServer) redirectToKeptLocation(w http.ResponseWriter, r *http.Request) {
	uri := s.popRedirectURI(r)
	if uri == "" {
		uri = "/"
	}

	w.Header().Set("Location", uri)
	w.WriteHeader(http.StatusFound)
}

func (s HTTPServer) requireAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next == nil {
			return
		}

		r = s.pushAuthenticatedUserID(r, "aiueo_user_id")

		next.ServeHTTP(w, r)
	})
}

type ctxKey string

const (
	ctxRedirectURI         = ctxKey("redirect_uri")
	ctxAuthenticatedUserID = ctxKey("authenticated_user_id")
)

func (s HTTPServer) pushRedirectURI(r *http.Request, uri string) *http.Request {
	ctx := context.WithValue(r.Context(), ctxRedirectURI, uri)
	return r.WithContext(ctx)
}

func (s HTTPServer) popRedirectURI(r *http.Request) string {
	uri, _ := r.Context().Value(ctxRedirectURI).(string)
	return uri
}

func (s HTTPServer) pushAuthenticatedUserID(r *http.Request, id app.UserID) *http.Request {
	ctx := context.WithValue(r.Context(), ctxAuthenticatedUserID, id)
	return r.WithContext(ctx)
}

func (s HTTPServer) popAuthenticatedUserID(r *http.Request) app.UserID {
	id, _ := r.Context().Value(ctxAuthenticatedUserID).(app.UserID)
	return id
}

func (s HTTPServer) renderErr(w http.ResponseWriter, did string, err error) {
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

func (s HTTPServer) logfln(format string, as ...interface{}) {
	s.logf(format+"\n", as...)
}

func (s HTTPServer) logf(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format, as...)
}

type httpServerRenderer interface {
	RenderFetchOwnerPage(io.Writer) error
	RenderErr(io.Writer, error) error
}
