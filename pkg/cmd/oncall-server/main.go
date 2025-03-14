package main

import (
	"os"

	"github.com/InariTheFox/oncall/pkg/util/cmd"
)

func main() {
	os.Exit(cmd.RunOnCallCmd("server"))
}
