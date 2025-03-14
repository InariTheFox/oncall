package commands

import "github.com/urfave/cli/v2"

var (
	ConfigFile      string
	HomePath        string
	ConfigOverrides string
	Version         bool
)

var commonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "config",
		Usage:       "Path to configuration file",
		Destination: &ConfigFile,
	},
	&cli.StringFlag{
		Name:        "homepath",
		Usage:       "Path to OnCall install/home path, defaults to working directory",
		Destination: &HomePath,
	},
	&cli.BoolFlag{
		Name:               "version",
		Aliases:            []string{"v"},
		Usage:              "print the version",
		DisableDefaultText: true,
		Destination:        &Version,
	},
}
