package tracing

import (
	"context"

	"github.com/grafana/dskit/services"
)

var _ services.NamedService = &ServiceTracer{}

type ServiceTracer struct {
	services.NamedService
}

func NewServiceTracer(service services.NamedService) *ServiceTracer {
	return &ServiceTracer{NamedService: service}
}

func (s *ServiceTracer) StartAsync(ctx context.Context) error {
	go func() {
		if err := s.AwaitRunning(ctx); err != nil {

		}
	}()

	return s.NamedService.StartAsync(ctx)
}
