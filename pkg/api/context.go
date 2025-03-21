package api

import (
	"context"
	"net/http"

	"github.com/InariTheFox/oncall/pkg/web"
)

func (s *HTTPServer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		requestContext := &web.Context{
			Response: w,
		}

		ctx = context.WithValue(ctx, web.Key{}, requestContext)
		requestContext.Request = r.WithContext(context.WithValue(r.Context(), web.Key{}, requestContext))
		*requestContext.Request = *requestContext.Request.WithContext(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
