package background_svcs

import (
	"github.com/InariTheFox/oncall/pkg/api"
	"github.com/InariTheFox/oncall/pkg/registry"
	"github.com/InariTheFox/oncall/pkg/services/apiserver"
)

func ProvideBackgroundServiceRegistry(
	httpServer *api.HTTPServer,
	apiServer apiserver.Service,
) *BackgroundServiceRegistry {
	return NewBackgroundServiceRegistry(
		httpServer,
		apiServer,
	)
}

type BackgroundServiceRegistry struct {
	Services []registry.BackgroundService
}

func NewBackgroundServiceRegistry(services ...registry.BackgroundService) *BackgroundServiceRegistry {
	return &BackgroundServiceRegistry{services}
}

func (r *BackgroundServiceRegistry) GetServices() []registry.BackgroundService {
	return r.Services
}
