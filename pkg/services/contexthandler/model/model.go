package contextModel

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/InariTheFox/oncall/pkg/setting"
	"github.com/InariTheFox/oncall/pkg/web"
)

type RequestContext struct {
	*web.Context

	Error error
}

func (ctx *RequestContext) Handle(cfg *setting.Cfg, status int, title string, err error) {
	data := struct {
		Title     string
		AppTitle  string
		AppSubUrl string
		ThemeType string
		ErrorMsg  error
	}{title, "OnCall", cfg.AppSubURL, "dark", nil}

	if err != nil {
		fmt.Errorf(title, "error", err)
	}

	ctx.HTML(status, cfg.ErrTemplateName, data)
}

func (ctx *RequestContext) IsApiRequest() bool {
	return strings.HasPrefix(ctx.Request.URL.Path, "/api")
}

func (ctx *RequestContext) JsonApiErr(status int, message string, err error) {
	resp := make(map[string]interface{})

	if err != nil {
		if status == http.StatusInternalServerError {
			fmt.Errorf(message, err)
		} else {
			fmt.Errorf(message, err)
		}
	}

	switch status {
	case http.StatusNotFound:
		resp["message"] = "Not Found"
	case http.StatusInternalServerError:
		resp["message"] = "Internal Server Error"
	}

	if message != "" {
		resp["message"] = message
	}

	ctx.JSON(status, resp)
}
