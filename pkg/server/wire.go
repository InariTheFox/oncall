//go:build wireinject
// +build wireinject

package server

import (
	"github.com/google/wire"

	"github.com/InariTheFox/oncall/pkg/api"
	"github.com/InariTheFox/oncall/pkg/api/routing"
	"github.com/InariTheFox/oncall/pkg/services/apiserver"
	"github.com/InariTheFox/oncall/pkg/services/apiserver/standalone"
	"github.com/InariTheFox/oncall/pkg/setting"
)

var wireBasicSet = wire.NewSet(
	New,
	api.ProvideHTTPServer,
	routing.ProvideRegister,
	wire.Bind(new(routing.RouteRegister), new(*routing.RouteRegisterImpl)),
	apiserver.WireSet,
)

var wireSet = wire.NewSet(
	wireBasicSet,
)

var wireCLISet = wire.NewSet(
	NewRunner,
	wireBasicSet,
)

func Initialize(cfg *setting.Cfg, opts Options, apiOpts api.ServerOptions) (*Server, error) {
	wire.Build(wireExtsSet)
	return &Server{}, nil
}

func InitializeForCLI(cfg *setting.Cfg) (Runner, error) {
	wire.Build(wireExtsCLISet)
	return Runner{}, nil
}

func InitializeAPIServerFactory() (standalone.APIServerFactory, error) {
	wire.Build(wireExtsStandaloneAPIServerSet)
	return &standalone.NoOpAPIServerFactory{}, nil // Wire will replace this with a real interface
}
