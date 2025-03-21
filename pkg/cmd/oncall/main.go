package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func main() {
	app := MainApp()

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("%s: %s %s\n", color.RedString("Error"), color.RedString("✗"), err)
	}

	os.Exit(0)
}

func MainApp() *cli.App {
	app := &cli.App{
		Name:  "oncall",
		Usage: "oncall server and command line interface",
		Authors: []*cli.Author{
			{
				Name:  "Inari",
				Email: "inarithefox@gmail.com",
			},
		},
		CommandNotFound: cmdNotFound,
		Commands: []*cli.Command{
			{
				Name:   "server",
				Usage:  "run the oncall server",
				Action: Server,
			},
			{
				Name:   "worker",
				Usage:  "run the oncall worker process only",
				Action: Worker,
			},
		},
	}

	return app
}

func cmdNotFound(c *cli.Context, command string) {
	fmt.Printf(
		"%s: '%s' is not a command. See '%s --help'.\n",
		c.App.Name,
		command,
		c.App.Name,
		os.Args[0])

	os.Exit(1)
}
