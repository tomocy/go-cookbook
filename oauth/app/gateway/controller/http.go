package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
)

func NewHTTPServer(w io.Writer, addr string) HTTPServer {
	return HTTPServer{
		w:    w,
		addr: addr,
	}
}

type HTTPServer struct {
	w    io.Writer
	addr string
}

func (s HTTPServer) Run() error {
	r := chi.NewRouter()

	s.logfln("listen and serve on %s", s.addr)
	return http.ListenAndServe(s.addr, r)
}

func (s HTTPServer) logfln(format string, as ...interface{}) {
	s.logf(format+"\n", as...)
}

func (s HTTPServer) logf(format string, as ...interface{}) {
	fmt.Fprintf(s.w, format, as...)
}
