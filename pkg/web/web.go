package web

import (
	"net/http"
	"strings"
)

type Handler interface{}

type Middleware = func(next http.Handler) http.Handler

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.URL.Path = strings.TrimPrefix(req.URL.Path, r.urlPrefix)
	r.f.ServeHTTP(rw, req)
}

func (r *Router) Use(h Handler) {
	r.f.Use(h)
}

func (r *Router) NotFound(h ...Handler) {
	r.f.NotFound(http.NotFound)
}
