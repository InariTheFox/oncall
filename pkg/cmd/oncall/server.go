package main

import (
	"strings"
	"time"

	"github.com/InariTheFox/oncall/pkg/api"
	"github.com/InariTheFox/oncall/pkg/server"
	"github.com/InariTheFox/oncall/pkg/setting"
	"github.com/InariTheFox/oncall/pkg/worker"
	"github.com/InariTheFox/oncall/pkg/worker/handlers"
	"github.com/urfave/cli/v2"
)

func Server(ctx *cli.Context) error {
	configOptions := strings.Split(ConfigOverrides, " ")
	cfg, err := setting.NewCfgFromArgs(setting.CommandLineArgs{
		Config:   ConfigFile,
		HomePath: HomePath,
		Args:     append(configOptions, ctx.Args().Slice()...),
	})
	if err != nil {
		return err
	}

	api, err := api.New(cfg)
	if err != nil {
		return err
	}

	worker, err := worker.NewRabbitWorker(cfg.RabbitMqHost, cfg.RabbitMqUsername, cfg.RabbitMqPassword, cfg.RabbitMqVhost, cfg.RabbitMqPort, cfg.RabbitMqQueueName, cfg.RabbitMqExchangeName, 5*time.Second)
	if err != nil {
		return err
	}

	worker.RegisterHandler("test", handlers.Handle, nil)

	s, err := server.New(cfg, api, worker)
	if err != nil {
		return err
	}

	return s.Run()
}
