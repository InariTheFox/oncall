package server

import (
	"github.com/google/wire"

	"github.com/InariTheFox/oncall/pkg/registry"
	"github.com/InariTheFox/oncall/pkg/registry/background_svcs"
	"github.com/InariTheFox/oncall/pkg/services/apiserver/standalone"
)

var wireExtsBasicSet = wire.NewSet(
	background_svcs.ProvideBackgroundServiceRegistry,
	wire.Bind(new(registry.BackgroundServiceRegistry), new(*background_svcs.BackgroundServiceRegistry)),
)

var wireExtsSet = wire.NewSet(
	wireSet,
	wireExtsBasicSet,
)

var wireExtsCLISet = wire.NewSet(
	wireCLISet,
	wireExtsBasicSet,
)

var wireExtsStandaloneAPIServerSet = wire.NewSet(
	standalone.ProvideAPIServerFactory,
)
