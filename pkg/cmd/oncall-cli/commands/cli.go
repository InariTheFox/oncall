package commands

import (
	"github.com/urfave/cli/v2"
)

const DefaultCommitValue = "NA"

func CLICommand(version string) *cli.Command {
	return &cli.Command{
		Name:  "cli",
		Usage: "run the OnCall cli",
		Flags: []cli.Flag{},
	}
}
