package routing

import (
	"net/http"

	"github.com/InariTheFox/oncall/pkg/api/response"
	contextModel "github.com/InariTheFox/oncall/pkg/services/contexthandler/model"
	"github.com/InariTheFox/oncall/pkg/web"
)

var (
	ServerError = func(err error) response.Response {
		return response.Error(http.StatusInternalServerError, "Server error", err)
	}
)

func Wrap(handler func(c *contextModel.RequestContext) response.Response) web.Handler {
	return func(c *contextModel.RequestContext) {
		if res := handler(c); res != nil {
			res.WriteTo(c)
		}
	}
}
