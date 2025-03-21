package api

import (
	"net/http"

	"github.com/InariTheFox/oncall/pkg/api/dto"
	"github.com/InariTheFox/oncall/pkg/web"
)

func (s *HTTPServer) Index(w http.ResponseWriter, r *http.Request) {
	ctx := web.FromContext(r.Context())

	viewData := &dto.IndexViewData{
		AppTitle:  "OnCall",
		AppSubUrl: s.Cfg.AppSubURL,
		AppUrl:    s.Cfg.AppURL,
	}

	ctx.HTML(http.StatusOK, "index", viewData)
}
