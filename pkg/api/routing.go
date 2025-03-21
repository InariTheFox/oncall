package api

import (
	"net/http"
	"path/filepath"

	"github.com/InariTheFox/oncall/pkg/web"
	"github.com/go-chi/chi/v5"
)

func (s *HTTPServer) Get(pattern string, h web.Handler) {
	s.route(pattern, http.MethodGet, h)
}

func (s *HTTPServer) Post(pattern string, h web.Handler) {
	s.route(pattern, http.MethodPost, h)
}

func (s *HTTPServer) route(pattern string, method string, h web.Handler) {
	s.router.Route(pattern, func(r chi.Router) {

		r.Use(s.Middleware)
		r.Use(web.Renderer(filepath.Join(s.Cfg.StaticRootPath, "views"), "[[", "]]"))

		switch method {
		case http.MethodGet:
			r.Get(pattern, h)
		}
	})
}
