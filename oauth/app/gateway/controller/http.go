package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
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

	r.Get("/", handlerFunc(s.showFetchUserPage()))

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

func handlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func (s HTTPServer) showFetchUserPage() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := s.renderer.RenderFetchUserPage(w); err != nil {
			s.renderErr(w, "render fetch user page", err)
			return
		}
	})
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
	RenderFetchUserPage(io.Writer) error
	RenderErr(io.Writer, error) error
}
