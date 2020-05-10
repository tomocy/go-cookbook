package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	if err := run(os.Stdout, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(w io.Writer, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("too few arguments")
	}

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	addr := flags.String("addr", ":80", "the address to listen and serve")
	if err := flags.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	serv := server{
		w:    w,
		addr: *addr,
		clients: map[string]client{
			"resource": {
				id:     "resource",
				secret: "12345",
				codes:  make(map[string]struct{}),
			},
			"app1": {
				id:     "app1",
				secret: "12345",
				redirectURIs: []string{
					"http://localhost/authzs/a",
				},
				toks: []string{
					"aiueoaiueoaiueo",
				},
				codes: make(map[string]struct{}),
			},
		},
	}
	if err := serv.run(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}

type server struct {
	w       io.Writer
	addr    string
	clients map[string]client
}

func (s server) run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/authz", s.authorize)
	mux.HandleFunc("/tokens", s.tokens)
	mux.HandleFunc("/introspections", s.introspect)

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s server) logfln(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format+"\n", as...)
}

const (
	respTypeCode = "code"
)

func (s server) authorize(w http.ResponseWriter, r *http.Request) {
	s.acceptRequestOfMethods(w, r, http.MethodGet)

	respType := r.URL.Query().Get("response_type")
	switch respType {
	case respTypeCode:
		s.generateAuthzCode(w, r)
	default:
		s.reportErr(w, http.StatusBadRequest, fmt.Sprint(http.StatusBadRequest))
	}
}

func (s *server) generateAuthzCode(w http.ResponseWriter, r *http.Request) {
	s.acceptRequestOfMethods(w, r, http.MethodGet)

	clientID, rawURI := r.URL.Query().Get("client_id"), r.URL.Query().Get("redirect_uri")
	if err := s.validateRedirectURI(clientID, rawURI); err != nil {
		s.reportErr(w, http.StatusBadRequest, fmt.Sprint(http.StatusBadRequest))
		return
	}

	uri, err := url.Parse(rawURI)
	if err != nil {
		s.reportErr(w, http.StatusBadRequest, fmt.Sprint(http.StatusBadRequest))
		return
	}
	code := fmt.Sprintf("%03d", rand.Intn(1000))
	queries := url.Values{
		"code": []string{code},
	}
	uri.RawQuery = queries.Encode()

	c := s.clients[clientID]
	c.codes[code] = struct{}{}
	s.clients[clientID] = c

	w.Header().Set("Location", uri.String())
	w.WriteHeader(http.StatusFound)
}

func (s server) validateRedirectURI(clientID, uri string) error {
	stored := s.clients[clientID].redirectURIs
	for _, stored := range stored {
		if stored == uri {
			return nil
		}
	}

	return fmt.Errorf("invalid redirect uri")
}

func (s server) tokens(w http.ResponseWriter, r *http.Request) {
	s.acceptRequestOfMethods(w, r, http.MethodPost)

	if isClientCredentialRequest(r) {
		s.tokenForClient(w, r)
		return
	}

	s.tokenForApp(w, r)
}

func isClientCredentialRequest(r *http.Request) bool {
	id, secret := r.URL.Query().Get("client_id"), r.URL.Query().Get("client_secret")
	return id != "" && secret != ""
}

func (s *server) tokenForClient(w http.ResponseWriter, r *http.Request) {
	client, err := s.authenticateClient(r)
	if err != nil {
		s.reportErr(w, http.StatusUnauthorized, err.Error())
		return
	}

	tok := generateAccessToken()
	client.toks = append(client.toks, tok)
	s.clients[client.id] = client

	resp := accessTokenResp{
		AccessToken: tok,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logfln("failed to encode access token response: %s", err)
		s.reportErr(w, http.StatusInternalServerError, fmt.Sprint(http.StatusInternalServerError))
		return
	}
}

func (s server) authenticateClient(r *http.Request) (client, error) {
	id, secret := retrieveClientCred(r)
	c, ok := s.clients[id]
	if !ok || c.secret != secret {
		return client{}, fmt.Errorf("invalid credentials")
	}

	return c, nil
}

func retrieveClientCred(r *http.Request) (string, string) {
	return r.URL.Query().Get("client_id"), r.URL.Query().Get("client_secret")
}

func generateAccessToken() string {
	return uuid.New().String()
}

const (
	grantTypeAuthzCode = "authorization_code"
)

func (s server) tokenForApp(w http.ResponseWriter, r *http.Request) {
	s.acceptRequestOfMethods(w, r, http.MethodPost)

	if err := r.ParseForm(); err != nil {
		s.logfln("failed to parse request of access token: %s", err)
		s.reportErr(w, http.StatusInternalServerError, fmt.Sprint(http.StatusInternalServerError))
		return
	}

	grantType := r.PostFormValue("grant_type")
	if grantType != grantTypeAuthzCode {
		s.reportErr(w, http.StatusBadRequest, fmt.Sprint(http.StatusBadRequest))
		return
	}
	clientID, code := r.PostFormValue("client_id"), r.PostFormValue("code")

	client, ok := s.clients[clientID]
	if !ok {
		s.reportErr(w, http.StatusBadRequest, fmt.Sprint(http.StatusBadRequest))
		return
	}
	if _, ok := client.codes[code]; !ok {
		s.reportErr(w, http.StatusBadRequest, fmt.Sprint(http.StatusBadRequest))
		return
	}

	delete(client.codes, code)

	tok := generateAccessToken()
	client.toks = append(client.toks, tok)
	s.clients[clientID] = client

	resp := accessTokenResp{
		AccessToken: tok,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logfln("failed to encode response of access token: %s", err)
		s.reportErr(w, http.StatusInternalServerError, fmt.Sprint(http.StatusInternalServerError))
		return
	}
}

func (s server) introspect(w http.ResponseWriter, r *http.Request) {
	s.acceptRequestOfMethods(w, r, http.MethodPost)

	clientID := r.URL.Query().Get("client_id")
	clientTok := retreiveAccessToken(r)
	if err := s.validateAccessToken(clientID, clientTok); err != nil {
		s.reportErr(w, http.StatusUnauthorized, fmt.Sprint(http.StatusUnauthorized))
		return
	}

	if err := r.ParseForm(); err != nil {
		s.logfln("failed to parse request form: %w", err)
		s.reportErr(w, http.StatusInternalServerError, fmt.Sprint(http.StatusInternalServerError))
		return
	}

	targetClientID := r.PostFormValue("client_id")
	targetTok := r.PostFormValue("token")
	if err := s.validateAccessToken(targetClientID, targetTok); err != nil {
		s.reportErr(w, http.StatusBadRequest, fmt.Sprint(http.StatusBadRequest))
		return
	}

	resp := introspectResp{
		Active: true,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logfln("failed to encode response: %w", err)
		s.reportErr(w, http.StatusInternalServerError, fmt.Sprint(http.StatusInternalServerError))
		return
	}
}

func (s server) validateAccessToken(id string, tok string) error {
	stored := s.clients[id].toks
	for _, stored := range stored {
		if stored == tok {
			return nil
		}
	}

	return fmt.Errorf("invalid token")
}

type introspectResp struct {
	Active bool `json:"active"`
}

func retreiveAccessToken(r *http.Request) string {
	raw := r.Header.Get("Authorization")
	raw = strings.Trim(raw, " ")

	var tok string
	fmt.Sscanf(raw, "Bearer %s", &tok)

	return tok
}

func (s server) acceptRequestOfMethods(w http.ResponseWriter, r *http.Request, methods ...string) {
	for _, m := range methods {
		if r.Method == m {
			return
		}
	}

	s.reportErr(w, http.StatusMethodNotAllowed, fmt.Sprint(http.StatusMethodNotAllowed))
}

func (s server) reportErr(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)

	resp := errResp{
		Message: msg,
	}
	json.NewEncoder(w).Encode(resp)
}

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
}

type errResp struct {
	Message string `json:"message"`
}

type client struct {
	id           string
	secret       string
	redirectURIs []string
	toks         []string
	codes        map[string]struct{}
}
