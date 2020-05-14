package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
)

func NewHTTPServer(w io.Writer, addr string, ren renderer) HTTPServer {
	return HTTPServer{
		w:        w,
		addr:     addr,
		renderer: ren,
	}
}

type HTTPServer struct {
	w        io.Writer
	addr     string
	renderer renderer
}

func (s HTTPServer) Run() error {
	r := chi.NewRouter()

	r.Route("/owners", func(r chi.Router) {
		r.Get("/", handlerFunc(s.showFetchOwnerPage()))
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
	ShowFetchOwnerPage(w io.Writer) error
	ShowErr(io.Writer, error) error
}
