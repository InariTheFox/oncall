package server

import (
	"context"

	"github.com/InariTheFox/oncall/pkg/api"
	"github.com/InariTheFox/oncall/pkg/setting"
	"github.com/grafana/dskit/services"
)

type coreService struct {
	*services.BasicService
	cfg     *setting.Cfg
	opts    Options
	apiOpts api.ServerOptions
	server  *Server
}

func NewService(cfg *setting.Cfg, opts Options, apiOpts api.ServerOptions) (*coreService, error) {
	s := &coreService{
		opts:    opts,
		apiOpts: apiOpts,
		cfg:     cfg,
	}

	s.BasicService = services.NewBasicService(s.start, s.running, s.stop)

	return s, nil
}

func (s *coreService) start(_ context.Context) error {
	srv, err := Initialize(s.cfg, s.opts, s.apiOpts)
	if err != nil {
		return err
	}

	s.server = srv

	return s.server.Init()
}

func (s *coreService) running(_ context.Context) error {
	return s.server.Run()
}

func (s *coreService) stop(failureReason error) error {
	return s.server.Shutdown(context.Background(), failureReason.Error())
}
