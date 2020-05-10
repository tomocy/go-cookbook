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
		addr          = flags.String("addr", ":80", "the address to listen and serve")
		authzAddr     = flags.String("authz-addr", "localhost:8081", "the addr of authorization server")
		authzPath     = flags.String("authz-path", "/authz", "the path for authorization")
		tokenPath     = flags.String("authz-token-path", "/tokens", "the path for access token")
		resourcesAddr = flags.String("resources-addr", "localhost:8082", "the addr of resource server")
		resourcesPath = flags.String("resources-path", "/", "the path for resources")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse args: %w", err)
	}

	serv := server{
		w:    w,
		addr: *addr,
		authzHandler: authzHandler{
			clientID:     "app1",
			clientSecret: "aiueoaiueoaiueo",
			authzEndpoint: endpoint{
				scheme: "http",
				addr:   *authzAddr,
				path:   *authzPath,
			},
			tokenEndpoint: endpoint{
				scheme: "http",
				addr:   *authzAddr,
				path:   *tokenPath,
			},
		},
		resourcesEndpoint: endpoint{
			scheme: "http",
			addr:   *resourcesAddr,
			path:   *resourcesPath,
		},
	}
	if err := serv.run(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}

type server struct {
	w                 io.Writer
	addr              string
	resourcesEndpoint endpoint
	authzHandler      authzHandler
}

func (s server) run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.resources)
	mux.HandleFunc("/a", s.resourcesOfA)
	mux.HandleFunc("/authzs/a", s.fetchAccessTokenFromA)

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s server) logfln(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format+"\n", as...)
}

func (s server) resources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprint(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<html>
	<body>
		<style>
		.window {
			height: 100vh;
			width: 100%;
		}
		.flex-center {
			height: inherit;
			width: inherit;
			display: flex;
			align-items: center;
			justify-content: center;
		}
		.btn {
			display: inline;
			text-decoration: none;
			padding: 10px 20px;
		}
		.btn-green {
			border: 1px solid green;
			color: green;
		}
		</style>
		<div class="window">
			<div class="flex-center">
				<a href="/a" class="btn btn-green">Get resources of A</a>
			</div>
		</div>
	</body>
</html>`)
}

func (s server) resourcesOfA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, fmt.Sprint(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	uri, err := s.authzHandler.authzCodeURI()
	if err != nil {
		s.logfln("failed to generate uri for authz code: %w", err)
		reportErr(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", uri)
	w.WriteHeader(http.StatusFound)
}

func (s server) fetchAccessTokenFromA(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	s.authzHandler.fetchAccessTokenFromA(code)
	fmt.Fprintln(w, s.authzHandler.tok)
}

func reportErr(w http.ResponseWriter, code int) {
	w.WriteHeader(code)

	fmt.Fprintln(w, fmt.Sprint(code))
}

func renderResourcesOfA(w io.Writer, rs []resourceOfA) {
	var b strings.Builder
	for _, r := range rs {
		fmt.Fprintf(&b, `<li class="resource"><p class="name">%s</p></li>`, r.Name)
	}

	fmt.Fprintf(w, `
	<html>
		<body>
			<style>
			.window {
				width: 100%%;
			}		
			.container {
				padding: 30px 0;
			}	
			.ul {
				list-style: none;
			}
			.resource {
				width: 90%%;
				border: 1px solid green;
				margin: 0 auto;
				padding: 10px 20px;
				text-align: left;
			}
			.resource .name {
				color: green;
			}
			</style>
			<div class="window">
				<div class="container">
					<ul class="ul">
						%s
					</ul>
				</div>
			</div>
		</body>
	</html>
`, b.String())
}

func composeErrFromResponse(r io.Reader) error {
	var decoded errRespOfA
	if err := json.NewDecoder(r).Decode(&decoded); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return fmt.Errorf(decoded.Message)
}

type authzHandler struct {
	clientID, clientSecret string
	authzEndpoint          endpoint
	tokenEndpoint          endpoint
	tok                    string
}

func (h authzHandler) authzCodeURI() (string, error) {
	uri, err := url.Parse(h.authzEndpoint.uri())
	if err != nil {
		return "", fmt.Errorf("failed to parse uri of authorization server")
	}

	queries := url.Values{
		"response_type": []string{"code"},
		"client_id":     []string{h.clientID},
		"redirect_uri":  []string{"http://localhost/authzs/a"},
	}
	uri.RawQuery = queries.Encode()

	return uri.String(), nil
}

func (h *authzHandler) fetchAccessTokenFromA(code string) error {
	vals := url.Values{
		"client_id":    []string{h.clientID},
		"redirect_uri": []string{"http://localhost/authzs/a"},
		"grant_type":   []string{"authorization_code"},
		"code":         []string{code},
	}
	req, err := http.NewRequest(http.MethodPost, h.tokenEndpoint.uri(), strings.NewReader(vals.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request for access token: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request for access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}

	var decoded accessTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return fmt.Errorf("failed to decode response of access token: %w", err)
	}

	h.tok = decoded.AccessToken

	return nil
}

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
}

type resourceOfA struct {
	Name string `json:"name"`
}

type errRespOfA struct {
	Message string `json:"message"`
}

type endpoint struct {
	scheme string
	addr   string
	path   string
}

func (e endpoint) uri() string {
	return fmt.Sprintf("%s://%s/%s", e.scheme, strings.TrimRight(e.addr, "/"), strings.TrimLeft(e.path, "/"))
}
