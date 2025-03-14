package middleware

import (
	"context"
	"net/http"
	"regexp"

	"github.com/InariTheFox/oncall/pkg/web"
)

type contextKey struct{}

var routeOperationNameKey = contextKey{}

var unnamedHandlers = []struct {
	pathPattern *regexp.Regexp
	handler     string
}{
	{handler: "public-assets", pathPattern: regexp.MustCompile("^/favicon.ico")},
	{handler: "public-assets", pathPattern: regexp.MustCompile("^/public/")},
	{handler: "/metrics", pathPattern: regexp.MustCompile("^/metrics")},
	{handler: "/healthz", pathPattern: regexp.MustCompile("^/healthz")},
	{handler: "/api/health", pathPattern: regexp.MustCompile("^/api/health")},
	{handler: "/robots.txt", pathPattern: regexp.MustCompile("^/robots.txt$")},
	// bundle all pprof endpoints under the same handler name
	{handler: "/debug/pprof-handlers", pathPattern: regexp.MustCompile("^/debug/pprof")},
}

// ProvideRouteOperationName creates a named middleware responsible for populating
// the context with the route operation name that can be used later in the request pipeline.
// Implements routing.RegisterNamedMiddleware.
func ProvideRouteOperationName(name string) web.Handler {
	return func(res http.ResponseWriter, req *http.Request, c *web.Context) {
		c.Request = addRouteNameToContext(c.Request, name)
	}
}

// RouteOperationName receives the route operation name from context, if set.
func RouteOperationName(req *http.Request) (string, bool) {
	if val := req.Context().Value(routeOperationNameKey); val != nil {
		op, ok := val.(string)
		return op, ok
	}

	for _, hp := range unnamedHandlers {
		if hp.pathPattern.Match([]byte(req.URL.Path)) {
			return hp.handler, true
		}
	}

	return "", false
}

func addRouteNameToContext(req *http.Request, operationName string) *http.Request {
	// don't set route name if it's set
	if _, exists := RouteOperationName(req); exists {
		return req
	}

	ctx := context.WithValue(req.Context(), routeOperationNameKey, operationName)
	return req.WithContext(ctx)
}
