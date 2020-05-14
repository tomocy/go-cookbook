package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tomocy/go-cookbook/oauth/3rd/server"
	"github.com/tomocy/go-cookbook/oauth/3rd/server/usecase"
)

func NewHTTPServer(w io.Writer, addr string, ren renderer, ownerServ server.OwnerService) HTTPServer {
	return HTTPServer{
		w:         w,
		addr:      addr,
		renderer:  ren,
		ownerServ: ownerServ,
	}
}

type HTTPServer struct {
	w         io.Writer
	addr      string
	renderer  renderer
	ownerServ server.OwnerService
}

func (s HTTPServer) Run() error {
	r := chi.NewRouter()

	r.Route("/owners", func(r chi.Router) {
		r.Get("/", handlerFunc(s.showFetchOwnerPage()))
		r.Get("/{provider}", handlerFunc(s.fetchOwner()))
	})

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
		if err := s.renderer.ShowFetchOwnerPage(w); err != nil {
			s.renderErr(w, err)
			return
		}
	})
}

func (s HTTPServer) fetchOwner() http.Handler {
	return s.requireUserAuthenticated(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := s.popAuthenticatedUserID(r)
			var haveToken usecase.DoHaveAccessToken
			had, err := haveToken.Do(id)
			if err != nil {
				s.renderErr(w, err)
				return
			}

			if had {
				s.pushIntoSession(w, sessOriginalURI, r.URL.String())
				var generate usecase.GenerateAuthzCodeURI
				uri, err := generate.Do(id)
				if err != nil {
					s.renderErr(w, err)
					return
				}

				w.Header().Set("Location", uri)
				w.WriteHeader(http.StatusSeeOther)
				return
			}

			fetch := usecase.FetchOwner{
				Service: s.ownerServ,
			}
			owner, err := fetch.Do(id)
			if err != nil {
				s.renderErr(w, err)
				return
			}

			if err := s.renderer.ShowOwner(w, owner); err != nil {
				s.renderErr(w, err)
				return
			}
		}),
	)
}

func (s HTTPServer) requireUserAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if next == nil {
			return
		}

		r = s.pushIntoContext(r, ctxAuthenticatedUserID, "test_user_id")

		next.ServeHTTP(w, r)
	})
}

type ctxKey string

const (
	ctxAuthenticatedUserID = "authenticatedUserID"
)

func (s HTTPServer) popAuthenticatedUserID(r *http.Request) string {
	id, _ := s.popFromContext(r, ctxAuthenticatedUserID).(string)
	return id
}

func (s HTTPServer) pushIntoContext(r *http.Request, key, val interface{}) *http.Request {
	ctx := context.WithValue(r.Context(), key, val)
	return r.WithContext(ctx)
}

func (s HTTPServer) popFromContext(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

const (
	sessOriginalURI = "original_location"
)

func (s HTTPServer) pushIntoSession(w http.ResponseWriter, key, val string) {
	http.SetCookie(w, &http.Cookie{
		Name:  key,
		Value: val,
	})
}

func (s HTTPServer) popFromSession(r *http.Request, key string) string {
	c, err := r.Cookie(key)
	if err != nil {
		return ""
	}

	return c.Value
}

func (s HTTPServer) renderErr(w io.Writer, err error) {
	if err := s.renderer.ShowErr(w, err); err != nil {
		s.logfln("failed to show err: %w", err)
		return
	}
}

func (s HTTPServer) logfln(format string, as ...interface{}) {
	s.logf(format+"\n", as...)
}

func (s HTTPServer) logf(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format, as...)
}

type renderer interface {
	ShowFetchOwnerPage(io.Writer) error
	ShowOwner(io.Writer, server.Owner) error
	ShowErr(io.Writer, error) error
}
