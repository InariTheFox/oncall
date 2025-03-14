package apiserver

import (
	"context"
	"net/http"

	serviceTracing "github.com/InariTheFox/oncall/pkg/modules/tracing"
	"github.com/InariTheFox/oncall/pkg/registry"
	"github.com/InariTheFox/oncall/pkg/setting"
	"github.com/grafana/dskit/services"
)

var (
	_ Service                    = (*service)(nil)
	_ registry.BackgroundService = (*service)(nil)
	_ registry.CanBeDisabled     = (*service)(nil)
)

func init() {

}

type Service interface {
	services.NamedService
	registry.BackgroundService
	registry.CanBeDisabled
}

type service struct {
	services.NamedService

	cfg *setting.Cfg

	stopCh    chan struct{}
	stoppedCh chan error

	handler http.Handler
}

func (s *service) IsDisabled() bool {
	return false
}

func ProvideService(cfg *setting.Cfg) (*service, error) {
	s := &service{
		cfg: cfg,
	}

	service := services.NewBasicService(s.start, s.running, nil).WithName("oncall-apiserver")
	s.NamedService = serviceTracing.NewServiceTracer(service)

	return s, nil
}

func (s *service) Run(ctx context.Context) error {
	if err := s.NamedService.StartAsync(ctx); err != nil {
		return err
	}

	if err := s.NamedService.AwaitRunning(ctx); err != nil {
		return err
	}

	return s.NamedService.AwaitTerminated(ctx)
}

func (s *service) running(ctx context.Context) error {
	select {
	case err := <-s.stoppedCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (s *service) start(ctx context.Context) error {
	return nil
}
