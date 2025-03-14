package build

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	GoOSWindows = "windows"
	GoOSLinux   = "linux"

	BackendBinary = "oncall"
	ServerBinary  = "oncall-server"
	CLIBinary     = "oncall-cli"
)

var binaries = []string{BackendBinary, ServerBinary, CLIBinary}

func logError(message string, err error) int {
	log.Println(message, err)

	return 1
}

func RunCmdCLI(c *cli.Context) error {
	os.Exit(RunCmd())

	return nil
}

func RunCmd() int {
	return 0
}
