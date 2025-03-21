package main

import "github.com/urfave/cli/v2"

var (
	ConfigFile      string
	ConfigOverrides string
	HomePath        string
)

var commonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "config",
		Usage:       "Path to config file",
		Destination: &ConfigFile,
	},
	&cli.StringFlag{
		Name:        "configOverrides",
		Usage:       "Configuration options to override defaults as a string. e.g. cfg:default.paths.log=/dev/null",
		Destination: &ConfigOverrides,
	},
	&cli.StringFlag{
		Name:        "homepath",
		Usage:       "Path to Grafana install/home path, defaults to working directory",
		Destination: &HomePath,
	},
}
