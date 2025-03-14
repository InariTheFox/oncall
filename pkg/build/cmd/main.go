package main

import (
	"log"
	"os"

	"github.com/InariTheFox/oncall/pkg/build"
	"github.com/urfave/cli/v2"
)

var additionalCommands []*cli.Command = make([]*cli.Command, 0, 5)

func main() {
	app := cli.NewApp()
	app.Commands = cli.Commands{
		{
			Name:   "build",
			Action: build.RunCmdCLI,
		},
	}

	app.Commands = append(app.Commands, additionalCommands...)

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
