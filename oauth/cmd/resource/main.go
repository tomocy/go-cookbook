package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

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
	var (
		addr              = flags.String("addr", ":80", "the address to listen and serve")
		authzAddr         = flags.String("authz-addr", "localhost:8081", "the addr of authorization server")
		authzClientID     = flags.String("authz-client-id", "resource", "the client id for authorization server")
		authzClientSecret = flags.String("authz-client-secret", "12345", "the client secret for authorization server")
		tokenPath         = flags.String("authz-token-path", "/tokens", "the path for access token")
		introspectionPath = flags.String("authz-introspection-path", "/introspections", "the path for introspections of access token")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	serv := server{
		w:    w,
		addr: *addr,
		authzHandler: authzHandler{
			clientID:     *authzClientID,
			clientSecret: *authzClientSecret,
			tokenEndpoint: endpoint{
				sheme: "http",
				addr:  *authzAddr,
				path:  *tokenPath,
			},
			introspectionEndpoint: endpoint{
				sheme: "http",
				addr:  *authzAddr,
				path:  *introspectionPath,
			},
		},
	}
	if err := serv.run(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}

type server struct {
	w             io.Writer
	addr          string
	authzHandler  authzHandler
	tokenEndpoint endpoint
}

func (s server) run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.resources)

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s server) logfln(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format+"\n", as...)
}

var resources = []resource{
	{Name: "a"}, {Name: "b"}, {Name: "c"},
	{Name: "d"}, {Name: "e"}, {Name: "f"},
	{Name: "g"}, {Name: "h"}, {Name: "i"},
}

func (s server) resources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.reportErr(w, http.StatusMethodNotAllowed, fmt.Sprint(http.StatusMethodNotAllowed))
		return
	}

	clientID := r.URL.Query().Get("client_id")
	tok := retrieveAccessToken(r)
	isValid, err := s.authzHandler.validateToken(clientID, tok)
	if err != nil {
		s.reportErr(w, http.StatusInternalServerError, fmt.Sprintf("failed to validate access token: %s", err))
		return
	}
	if !isValid {
		s.reportErr(w, http.StatusUnauthorized, "invalid access token")
		return
	}

	if err := json.NewEncoder(w).Encode(resources); err != nil {
		s.reportErr(w, http.StatusInternalServerError, fmt.Sprintf("failed to encode response: %s", err))
		return
	}
}

func retrieveAccessToken(r *http.Request) string {
	raw := r.Header.Get("Authorization")
	raw = strings.Trim(raw, " ")

	var tok string
	fmt.Sscanf(raw, "Bearer %s", &tok)

	return tok
}

func (s server) reportErr(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)

	resp := response{
		Code:    code,
		Message: msg,
	}
	json.NewEncoder(w).Encode(resp)
}

type response struct {
	Code      int        `json:"code"`
	Message   string     `json:"message"`
	Resources []resource `json:"resources"`
}

type resource struct {
	Name string `json:"name"`
}

type authzHandler struct {
	clientID, clientSecret string
	tokenEndpoint          endpoint
	introspectionEndpoint  endpoint
	tok                    string
}

func (h *authzHandler) validateToken(clientID, tok string) (bool, error) {
	if err := h.fetchAccessToken(); err != nil {
		return false, fmt.Errorf("failed to fetch access token: %w", err)
	}

	vals := url.Values{
		"client_id": []string{clientID},
		"token":     []string{tok},
	}
	req, err := http.NewRequest(http.MethodPost, h.introspectionEndpoint.uri(), strings.NewReader(vals.Encode()))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	h.setCredsToRequest(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, composeErrFromResponse(resp.Body)
	}

	var decoded validateTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return decoded.Active, nil
}

type validateTokenResp struct {
	Active bool `json:"active"`
}

func (h authzHandler) setCredsToRequest(target *http.Request) {
	queries := url.Values{
		"client_id": []string{h.clientID},
	}
	target.URL.RawQuery = queries.Encode()

	target.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.tok))
}

func (h *authzHandler) fetchAccessToken() error {
	if h.tok != "" {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, h.tokenEndpoint.uri(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request to get token: %w", err)
	}
	queries := url.Values{
		"client_id":     []string{h.clientID},
		"client_secret": []string{h.clientSecret},
	}
	req.URL.RawQuery = queries.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return composeErrFromResponse(resp.Body)
	}

	var decoded tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	h.tok = decoded.AccessToken

	return nil
}

func composeErrFromResponse(r io.Reader) error {
	var decoded errResp
	if err := json.NewDecoder(r).Decode(&decoded); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return fmt.Errorf(decoded.Message)
}

type tokenResp struct {
	AccessToken string `json:"access_token"`
}

type errResp struct {
	Message string `json:"message"`
}

type endpoint struct {
	sheme string
	addr  string
	path  string
}

func (e endpoint) uri() string {
	return fmt.Sprintf("%s://%s/%s", e.sheme, strings.TrimRight(e.addr, "/"), strings.TrimLeft(e.path, "/"))
}
