package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"

	gcli "github.com/InariTheFox/oncall/pkg/cmd/oncall-cli/commands"
	"github.com/InariTheFox/oncall/pkg/cmd/oncall-server/commands"
	"github.com/InariTheFox/oncall/pkg/server"
	"github.com/InariTheFox/oncall/pkg/services/apiserver/standalone"
)

var version = "7.0.0-beta-1"
var commit = gcli.DefaultCommitValue
var buildBranch = "master"
var buildStamp string

func main() {
	app := MainApp()

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("%s: %s %s\n", color.RedString("Error"), color.RedString("✗"), err)
		os.Exit(1)
	}

	os.Exit(0)
}

func MainApp() *cli.App {
	app := &cli.App{
		Name:  "OnCall",
		Usage: "On-call management system designed that integrates with the Grafana ecosystem",
		Authors: []*cli.Author{
			{
				Name:  "InariTheFox",
				Email: "inarithefox@gmail.com",
			},
		},
		Version: version,
		Commands: []*cli.Command{
			gcli.CLICommand(version),
			commands.ServerCommand(version, commit, buildBranch, buildStamp),
		},
		CommandNotFound:      cmdNotFound,
		EnableBashCompletion: true,
	}

	buildInfo := standalone.BuildInfo{
		Version:     version,
		Commit:      commit,
		BuildBranch: buildBranch,
		BuildStamp:  buildStamp,
	}

	commands.SetBuildInfo(buildInfo)

	f, err := server.InitializeAPIServerFactory()
	if err == nil {
		cmd := f.GetCLICommand(buildInfo)
		if cmd != nil {
			app.Commands = append(app.Commands, cmd)
		}
	}

	return app
}

func cmdNotFound(c *cli.Context, command string) {
	fmt.Printf(
		"%s: '%s' is not a %s command. See '%s --help'.\n",
		c.App.Name,
		command,
		c.App.Name,
		os.Args[0])

	os.Exit(1)
}
