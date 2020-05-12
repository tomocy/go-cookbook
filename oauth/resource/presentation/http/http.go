package http

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"

// 	"github.com/go-chi/chi"
// 	"github.com/tomocy/go-cookbook/oauth/resource"
// 	"github.com/tomocy/go-cookbook/oauth/resource/usecase"
// )

// func NewServer(w io.Writer, addr, clientID, clientSecret string, userRepo resource.UserRepo) Server {
// 	return Server{
// 		w:    w,
// 		addr: addr,
// 		oauthCreds: oauthCreds{
// 			clientID:     clientID,
// 			clientSecret: clientSecret,
// 		},
// 		userRepo: userRepo,
// 	}
// }

// type Server struct {
// 	w               io.Writer
// 	addr            string
// 	oauthCreds      oauthCreds
// 	userRepo        resource.UserRepo
// 	accessTokenRepo resource.AccessTokenRepo
// }

// func (s Server) Run() error {
// 	r := chi.NewRouter()

// 	r.Get("/owner", handlerFunc(s.owner()))

// 	s.logfln("liste and serve on %s", s.addr)
// 	return http.ListenAndServe(s.addr, r)
// }

// func handlerFunc(h http.Handler) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		h.ServeHTTP(w, r)
// 	}
// }

// func (s Server) owner() http.Handler {
// 	return s.allowMethods(
// 		[]string{http.MethodGet},
// 		s.introspectAccessToken(
// 			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				id, ok := ownerIDFromContext(r)
// 				if !ok {
// 					s.responseStatus(w, http.StatusUnauthorized)
// 					return
// 				}

// 				u := usecase.NewFindUser(s.userRepo)
// 				user, found, err := u.Do(id)
// 				if err != nil {
// 					s.logfln("failed to find user: %s", err)
// 					s.responseStatus(w, http.StatusInternalServerError)
// 					return
// 				}
// 				if !found {
// 					s.responseStatus(w, http.StatusNotFound)
// 					return
// 				}

// 				if err := json.NewEncoder(w).Encode(ownerResp{
// 					ID:   string(user.ID()),
// 					Name: user.Name(),
// 				}); err != nil {
// 					s.logfln("failed to encode response of authorizing user: %s", err)
// 					return
// 				}
// 			}),
// 		),
// 	)
// }

// type ownerResp struct {
// 	ID   string `json:"user_id"`
// 	Name string `json:"name"`
// }

// func (s Server) allowMethods(methods []string, next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		for _, m := range methods {
// 			if r.Method == m {
// 				next.ServeHTTP(w, r)
// 				return
// 			}
// 		}

// 		s.responseStatus(w, http.StatusMethodNotAllowed)
// 	})
// }

// func (s Server) introspectAccessToken(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		rawTok := retrieveAccessToken(r)
// 		introspect := usecase.NewIntrospectAccessToken(s.accessTokenRepo)
// 		tok, valid, err := introspect.Do(s.oauthCreds.clientID, s.oauthCreds.clientSecret, rawTok)
// 		if err != nil {
// 			s.logfln("failed to introspect access token: %s", err)
// 			s.responseStatus(w, http.StatusInternalServerError)
// 			return
// 		}
// 		if !valid {
// 			s.responseStatus(w, http.StatusUnauthorized)
// 			return
// 		}

// 		r = reqWithOwnerID(r, tok)

// 		next.ServeHTTP(w, r)
// 	})
// }

// func retrieveAccessToken(r *http.Request) string {
// 	authz := r.Header.Get("Authorization")
// 	var tok string
// 	fmt.Sscanf(authz, "Bearer %s", &tok)

// 	return tok
// }

// type ctxKey string

// const (
// 	ctxOwnerID = ctxKey("owner_id")
// )

// func ownerIDFromContext(r *http.Request) (resource.UserID, bool) {
// 	id, ok := r.Context().Value(ctxOwnerID).(resource.UserID)
// 	return id, ok
// }

// func reqWithOwnerID(r *http.Request, tok resource.AccessToken) *http.Request {
// 	ctx := context.WithValue(r.Context(), ctxOwnerID, tok.UserID())
// 	return r.WithContext(ctx)
// }

// func (s Server) responseStatus(w http.ResponseWriter, code int) {
// 	w.WriteHeader(code)
// 	fmt.Fprintln(w, http.StatusText(code))
// }

// func (s Server) logfln(format string, as ...interface{}) {
// 	s.logf(format+"\n", as...)
// }

// func (s Server) logf(format string, as ...interface{}) {
// 	fmt.Fprintf(s.w, format, as...)
// }

// type oauthCreds struct {
// 	clientID, clientSecret string
// }
