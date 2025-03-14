package commands

import "github.com/urfave/cli/v2"

var pluginCommands = []*cli.Command{}

var adminCommands = []*cli.Command{}

var Commands = []*cli.Command{
	{
		Name:        "plugins",
		Usage:       "Manage plugins for OnCall",
		Subcommands: pluginCommands,
	},
	{
		Name:        "admin",
		Usage:       "OnCall admin commands",
		Subcommands: adminCommands,
	},
}
