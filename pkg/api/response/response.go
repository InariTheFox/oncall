package response

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	contextModel "github.com/InariTheFox/oncall/pkg/services/contexthandler/model"
)

type Response interface {
	Body() []byte
	Status() int
	WriteTo(ctx *contextModel.RequestContext)
}

type NormalResponse struct {
	body       *bytes.Buffer
	err        error
	errMessage string
	header     http.Header
	status     int
}

// Body gets the response's body.
func (r *NormalResponse) Body() []byte {
	return r.body.Bytes()
}

// Body gets the response's body.
// Required to implement api.Response.
func (r *RedirectResponse) Body() []byte {
	return nil
}

// Err gets the response's err.
func (r *NormalResponse) Err() error {
	return r.err
}

// ErrMessage gets the response's errMessage.
func (r *NormalResponse) ErrMessage() string {
	return r.errMessage
}

func Error(status int, message string, err error) *NormalResponse {
	data := make(map[string]any)

	switch status {
	case 404:
		data["message"] = "Not Found"
	case 500:
		data["message"] = "Internal Server Error"
	}

	if message != "" {
		data["message"] = message
	}

	resp := JSON(status, data)

	if err != nil {
		resp.errMessage = message
		resp.err = err
	}

	return resp
}

// Header implements http.ResponseWriter
func (r *NormalResponse) Header() http.Header {
	return r.header
}

func JSON(status int, body any) *NormalResponse {
	return Respond(status, body).
		SetHeader("Content-Type", "application/json")
}

func Respond(status int, body any) *NormalResponse {
	var b []byte
	switch t := body.(type) {
	case []byte:
		b = t
	case string:
		b = []byte(t)
	case nil:
		break
	default:
		var err error
		if b, err = json.Marshal(body); err != nil {
			return Error(http.StatusInternalServerError, "body json marshal", err)
		}
	}

	return &NormalResponse{
		status: status,
		body:   bytes.NewBuffer(b),
		header: make(http.Header),
	}
}

// RedirectResponse represents a redirect response.
type RedirectResponse struct {
	location string
}

func Redirect(location string) *RedirectResponse {
	return &RedirectResponse{location: location}
}

func (r *NormalResponse) SetHeader(key, value string) *NormalResponse {
	r.header.Set(key, value)
	return r
}

// Status gets the response's status.
func (r *NormalResponse) Status() int {
	return r.status
}

// Status gets the response's status.
// Required to implement api.Response.
func (*RedirectResponse) Status() int {
	return http.StatusFound
}

func (r *NormalResponse) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

// WriteHeader implements http.ResponseWriter
func (r *NormalResponse) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *NormalResponse) WriteTo(ctx *contextModel.RequestContext) {
	if r.err != nil {
		fmt.Errorf("%s", r.err)
	}

	header := ctx.Response.Header()
	for k, v := range r.header {
		header[k] = v
	}

	ctx.Response.WriteHeader(r.status)

	if _, err := ctx.Response.Write(r.body.Bytes()); err != nil {
		fmt.Errorf("Error writing to response: %s", err)
	}
}

// WriteTo writes to a response.
func (r *RedirectResponse) WriteTo(ctx *contextModel.RequestContext) {
	ctx.Redirect(r.location)
}
