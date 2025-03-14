package api

import (
	"net/http"

	"github.com/InariTheFox/oncall/pkg/api/dtos"
	contextModel "github.com/InariTheFox/oncall/pkg/services/contexthandler/model"
)

func (s *HTTPServer) Index(c *contextModel.RequestContext) {
	data, err := s.setIndexViewData(c)
	if err != nil {
		c.Handle(s.Cfg, http.StatusInternalServerError, "Failed to get settings", err)
		return
	}

	c.HTML(http.StatusOK, "index", data)
}

func (s *HTTPServer) NotFoundHandler(c *contextModel.RequestContext) {
	if c.IsApiRequest() {
		c.JsonApiErr(http.StatusNotFound, "Not found", nil)
		return
	}

	data, err := s.setIndexViewData(c)
	if err != nil {
		c.Handle(s.Cfg, http.StatusInternalServerError, "Failed to get settings", err)
		return
	}

	c.HTML(http.StatusNotFound, "index", data)
}

func (s *HTTPServer) setIndexViewData(c *contextModel.RequestContext) (*dtos.IndexViewData, error) {
	data := dtos.IndexViewData{
		AppUrl: s.Cfg.AppURL,
	}

	return &data, nil
}
