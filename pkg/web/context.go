package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"syscall"
)

const (
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json; charset=UTF-8"
	contentTypeHTML   = "text/html; charset=UTF-8"
)

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	template template.Template
}

func (ctx *Context) HTML(status int, name string, data any) {
	ctx.Response.Header().Set(headerContentType, contentTypeHTML)
	ctx.Response.WriteHeader(status)
	if err := ctx.template.ExecuteTemplate(ctx.Response, name, data); err != nil {
		if errors.Is(err, syscall.EPIPE) { // Client has stopped listening.
			return
		}
		panic(fmt.Sprintf("Context.HTML - Error rendering template: %s. You may need to build frontend assets \n %s", name, err.Error()))
	}
}

func (ctx *Context) JSON(status int, data any) {
	ctx.Response.Header().Set(headerContentType, contentTypeJSON)
	ctx.Response.WriteHeader(status)
	enc := json.NewEncoder(ctx.Response)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		panic("Context.JSON: " + err.Error())
	}
}

// Redirect sends a redirect response
func (ctx *Context) Redirect(location string, status ...int) {
	code := http.StatusFound
	if len(status) == 1 {
		code = status[0]
	}

	http.Redirect(ctx.Response, ctx.Request, location, code)
}
