package http

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/tomocy/go-cookbook/oauth/authz"
	"github.com/tomocy/go-cookbook/oauth/authz/usecase"
)

func NewServer(
	w io.Writer, addr string,
	clientRepo authz.ClientRepo, userRepo authz.UserRepo, sessRepo authz.SessionRepo,
) (Server, error) {
	s := Server{
		w:          w,
		addr:       addr,
		clientRepo: clientRepo,
		userRepo:   userRepo,
		sessRepo:   sessRepo,
	}
	if err := s.prepareHTMLRenderer(); err != nil {
		return Server{}, fmt.Errorf("failed to prepare html renderer: %w", err)
	}

	return s, nil
}

type Server struct {
	w          io.Writer
	addr       string
	clientRepo authz.ClientRepo
	userRepo   authz.UserRepo
	sessRepo   authz.SessionRepo
	renderer   htmlRenderer
}

func (s *Server) prepareHTMLRenderer() error {
	ren, err := newHTMLRenderer(map[string][]string{
		"client.new":     []string{"views/html/templates/client/new.html"},
		"client.single":  []string{"views/html/templates/client/single.html"},
		"user.new":       []string{"views/html/templates/user/new.html"},
		"user.authn.new": []string{"views/html/templates/user/authn/new.html"},
	})
	if err != nil {
		return fmt.Errorf("failed to create html renderer: %w", err)
	}

	s.renderer = ren

	return nil
}

func (s Server) Run() error {
	r := chi.NewRouter()

	r.Get("/css/", handlerFunc(http.StripPrefix("/css/", http.FileServer(http.Dir("views/css")))))

	r.Get("/authzs", handlerFunc(s.authorize()))
	r.Post("/tokens", handlerFunc(s.createAccessToken()))
	r.Post("/introspections", handlerFunc(s.introspectAccessToken()))

	r.Route("/developers/tester/clients", func(r chi.Router) {
		r.Get("/new", handlerFunc(s.showCreateClientPage()))
		r.Get("/{id}", handlerFunc(s.showClient()))
		r.Post("/", handlerFunc(s.createClient()))
	})

	r.Route("/users", func(r chi.Router) {
		r.Get("/new", handlerFunc(s.showCreateUserPage()))
		r.Post("/", handlerFunc(s.createUser()))

		r.Route("/authns", func(r chi.Router) {
			r.Get("/new", handlerFunc(s.showAuthenticateUserPage()))
			r.Post("/", handlerFunc(s.authenticateUser()))
		})
	})

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

const (
	respTypeCode = "code"
)

func (s Server) authorize() http.Handler {
	return s.allowMethods(
		[]string{http.MethodGet},
		s.requireUserAuthenticated(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				respType := r.URL.Query().Get("response_type")
				switch respType {
				case respTypeCode:
					s.redirectWithAuthzCode(w, r)
					return
				default:
					s.responseStatus(w, http.StatusBadRequest)
				}
			}),
		),
	)
}

func (s Server) redirectWithAuthzCode(w http.ResponseWriter, r *http.Request) {
	clientID := authz.ClientID(r.URL.Query().Get("client_id"))
	u := usecase.NewGenerateRedirectURIWithAuthzCode(s.clientRepo)
	uri, err := u.Do(clientID)
	if err != nil {
		s.responseErr(w, err, "generate redirect uri with authz code")
		return
	}

	w.Header().Set("Location", uri)
	w.WriteHeader(http.StatusFound)
}

const (
	grantTypeAuthzCode = "authorization_code"
)

func (s Server) createAccessToken() http.Handler {
	return s.allowMethods(
		[]string{http.MethodPost},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				s.responseErr(w, err, "parse form")
				return
			}

			grantType := r.PostFormValue("grant_type")
			switch grantType {
			case grantTypeAuthzCode:
				s.createAccessTokenInAuthzCodeFlow(w, r)
				return
			default:
				s.responseStatus(w, http.StatusBadRequest)
			}
		}),
	)
}

func (s Server) createAccessTokenInAuthzCodeFlow(w http.ResponseWriter, r *http.Request) {
	var (
		clientID = authz.ClientID(r.PostFormValue("client_id"))
		code     = r.PostFormValue("code")
	)
	u := usecase.NewGenerateAccessToken(s.clientRepo)
	tok, err := u.Do(clientID, code)
	if err != nil {
		s.responseErr(w, err, "generate access token")
		return
	}

	if err := json.NewEncoder(w).Encode(accessTokenResp{
		AccessToken: tok.Token(),
	}); err != nil {
		s.responseErr(w, err, "encode response of access token")
		return
	}
}

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
}

func (s Server) introspectAccessToken() http.Handler {
	return s.allowMethods(
		[]string{http.MethodPost},
		s.denyUnAuthorizedUser(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					s.responseErr(w, err, "parse form")
					return
				}

				var (
					clientID = authz.ClientID(r.PostFormValue("client_id"))
					tok      = r.PostFormValue("token")
				)
				u := usecase.NewIntrospectAccessToken(s.clientRepo)
				_, valid, err := u.Do(clientID, tok)
				if err != nil {
					s.responseErr(w, err, "intropsct access token")
					return
				}

				if err := json.NewEncoder(w).Encode(introspectAccessTokenResp{
					Active: valid,
				}); err != nil {
					s.responseErr(w, err, "encode response of introspection of access token")
					return
				}
			}),
		),
	)
}

func (s Server) denyUnAuthorizedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			clientID = authz.ClientID(r.URL.Query().Get("client_id"))
			tok      = s.retreiveAcessToken(r)
		)
		u := usecase.NewIntrospectAccessToken(s.clientRepo)
		_, valid, err := u.Do(clientID, tok)
		if err != nil {
			s.responseErr(w, err, "introspect access token")
			return
		}
		if !valid {
			s.responseStatus(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s Server) retreiveAcessToken(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	var tok string
	fmt.Sscanf(authz, "Bearer %s", &tok)

	return tok
}

type introspectAccessTokenResp struct {
	Active bool `json:"active"`
}

func (s Server) showCreateClientPage() http.Handler {
	return s.allowMethods(
		[]string{http.MethodGet},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := s.renderer.Render(w, "client.new", nil); err != nil {
				s.logfln("failed to render create client page: %s", err)
				s.responseStatus(w, http.StatusInternalServerError)
				return
			}
		}),
	)
}

func (s Server) showClient() http.Handler {
	return s.allowMethods(
		[]string{http.MethodGet},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := decodeClientID(chi.URLParam(r, "id"))
			if err != nil {
				s.responseStatus(w, http.StatusBadRequest)
				return
			}

			u := usecase.NewFindClient(s.clientRepo)
			client, found, err := u.Do(id)
			if err != nil {
				s.logfln("failed to find client: %s", err)
				s.responseStatus(w, http.StatusInternalServerError)
				return
			}
			if !found {
				s.responseStatus(w, http.StatusNotFound)
				return
			}

			if err := s.renderer.Render(w, "client.single", map[string]interface{}{
				"Client": map[string]interface{}{
					"ID":     client.ID(),
					"Secret": client.Secret(),
				},
			}); err != nil {
				s.logfln("failed to render client page: %s", err)
				s.responseStatus(w, http.StatusInternalServerError)
				return
			}
		}),
	)
}

func (s Server) createClient() http.Handler {
	return s.allowMethods(
		[]string{http.MethodPost},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseForm(); err != nil {
				s.logfln("failed to parse form: %s", err)
				s.responseStatus(w, http.StatusInternalServerError)
				return
			}

			reidrectURI := r.PostFormValue("redirect_uri")
			u := usecase.NewCreateClient(s.clientRepo)
			client, err := u.Do(reidrectURI)
			if err != nil {
				s.responseErr(w, err, "create client")
				return
			}

			encoded := encodeClientID(client.ID())
			w.Header().Set("Location", fmt.Sprintf("/developers/tester/clients/%s", encoded))
			w.WriteHeader(http.StatusFound)
		}),
	)
}

func encodeClientID(id authz.ClientID) string {
	return url.PathEscape(string(id))
}

func decodeClientID(raw string) (authz.ClientID, error) {
	decoded, err := url.PathUnescape(raw)
	return authz.ClientID(decoded), err
}

const (
	cookieSessionID = "session"
)

func (s Server) showCreateUserPage() http.Handler {
	return s.allowMethods(
		[]string{http.MethodGet},
		s.denyAuthenticatedUser(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := s.renderer.Render(w, "user.new", nil); err != nil {
					s.logfln("failed to render create user page: %s", err)
					s.responseStatus(w, http.StatusInternalServerError)
					return
				}
			}),
		),
	)
}

func (s Server) createUser() http.Handler {
	return s.allowMethods(
		[]string{http.MethodPost},
		s.denyAuthenticatedUser(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					s.logfln("failed to parse request form: %s", err)
					s.responseStatus(w, http.StatusInternalServerError)
					return
				}

				email, pass := r.PostFormValue("email"), r.PostFormValue("password")
				u := usecase.NewCreateUser(s.userRepo, s.sessRepo)
				_, sess, err := u.Do(email, pass)
				if err != nil {
					s.responseErr(w, err, "create user")
					return
				}

				s.startSession(w, sess)
			}),
		),
	)
}

func (s Server) showAuthenticateUserPage() http.Handler {
	return s.allowMethods(
		[]string{http.MethodGet},
		s.denyAuthenticatedUser(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := s.renderer.Render(w, "user.authn.new", nil); err != nil {
					s.responseErr(w, err, "render authenticate user page")
				}
			}),
		),
	)
}

func (s Server) authenticateUser() http.Handler {
	return s.allowMethods(
		[]string{http.MethodPost},
		s.denyAuthenticatedUser(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseForm(); err != nil {
					s.responseErr(w, err, "parse form")
					return
				}

				email, pass := r.PostFormValue("email"), r.PostFormValue("password")
				u := usecase.NewAuthenticateUser(s.userRepo, s.sessRepo)
				_, sess, authenticated, err := u.Do(email, pass)
				if err != nil {
					s.responseErr(w, err, "authenticate user")
					return
				}
				if !authenticated {
					s.responseStatus(w, http.StatusBadRequest)
					return
				}

				s.startSession(w, sess)

				s.redirectWithContext(w, r)
			}),
		),
	)
}

type ctxKey string

const (
	ctxNextLocation = ctxKey("redirect_uri")
)

func (s Server) redirectWithContext(w http.ResponseWriter, r *http.Request) {
	loc, ok := r.Context().Value(ctxNextLocation).(string)
	if !ok || loc == "" {
		loc = "/"
	}

	w.Header().Set("Location", loc)
	w.WriteHeader(http.StatusSeeOther)
}

func (s Server) allowMethods(methods []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, m := range methods {
			if r.Method != m {
				continue
			}

			next.ServeHTTP(w, r)
		}
	})
}

func (s Server) catchRedirectURI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri := r.URL.Query().Get("redirect_uri")
		ctx := context.WithValue(r.Context(), ctxNextLocation, uri)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (s Server) requireUserAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := context.WithValue(r.Context(), ctxNextLocation, r.URL.String())
		r = r.WithContext(reqCtx)

		sessID, found, err := s.retrieveSessionID(r)
		if err != nil {
			s.responseErr(w, err, "retrieve session id")
			return
		}
		if !found {
			s.redirectToAuthenticateUserPage(w)
			return
		}

		findCtx := context.TODO()
		_, found, err = s.sessRepo.Find(findCtx, sessID)
		if err != nil {
			s.responseErr(w, err, "find session")
			return
		}
		if !found {
			s.redirectToAuthenticateUserPage(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s Server) redirectToAuthenticateUserPage(w http.ResponseWriter) {
	w.Header().Set("Location", "/users/authns/new")
	s.responseStatus(w, http.StatusSeeOther)
}

func (s Server) denyAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessID, found, err := s.retrieveSessionID(r)
		if err != nil {
			s.responseErr(w, err, "retrieve session id")
			return
		}
		if !found {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.TODO()
		_, found, err = s.sessRepo.Find(ctx, sessID)
		if err != nil {
			s.responseErr(w, err, "find session")
			return
		}
		if !found {
			next.ServeHTTP(w, r)
			return
		}

		s.responseStatus(w, http.StatusSeeOther)
	})
}

func (s Server) startSession(w http.ResponseWriter, sess authz.Session) {
	encoded := encodeSessionID(sess.ID())

	cookie := &http.Cookie{
		Name:  cookieSessionID,
		Value: encoded,
	}
	http.SetCookie(w, cookie)
}

func (s Server) retrieveSessionID(r *http.Request) (authz.SessionID, bool, error) {
	cookie, err := r.Cookie(cookieSessionID)
	if err != nil && err != http.ErrNoCookie {
		return "", false, err
	}
	if err != nil && err == http.ErrNoCookie {
		return "", false, nil
	}

	decoded, err := decodeSessionID(cookie.Value)
	if err != nil {
		return "", false, err
	}

	return authz.SessionID(decoded), true, nil
}

func encodeSessionID(id authz.SessionID) string {
	return url.PathEscape(string(id))
}

func decodeSessionID(raw string) (authz.SessionID, error) {
	decoded, err := url.PathUnescape(raw)
	return authz.SessionID(decoded), err
}

func handlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func (s Server) responseErr(w http.ResponseWriter, err error, did string) {
	if authz.IsErrInternal(err) {
		s.logfln("failed to %s: %s", did, err)
		s.responseStatus(w, http.StatusInternalServerError)
		return
	}

	s.response(w, http.StatusBadRequest, err.Error())
}

func (s Server) responseStatus(w http.ResponseWriter, code int) {
	s.response(w, code, http.StatusText(code))
}

func (s Server) response(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	fmt.Fprintln(w, msg)
}

func (s Server) logfln(format string, as ...interface{}) {
	s.logf(format+"\n", as...)
}

func (s Server) logf(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format, as...)
}

func newHTMLRenderer(files map[string][]string) (htmlRenderer, error) {
	ren := make(htmlRenderer)
	for name, fs := range files {
		templ, err := template.ParseFiles(fs...)
		if err != nil {
			return htmlRenderer{}, fmt.Errorf("failed to parse files: %w", err)
		}

		ren[name] = templ
	}

	return ren, nil
}

type htmlRenderer map[string]*template.Template

func (r htmlRenderer) Render(w io.Writer, name string, data interface{}) error {
	return r[name].Execute(w, data)
}
