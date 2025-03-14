package commands

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/InariTheFox/oncall/pkg/api"
	"github.com/InariTheFox/oncall/pkg/server"
	"github.com/InariTheFox/oncall/pkg/services/apiserver/standalone"
	"github.com/InariTheFox/oncall/pkg/setting"
	"github.com/urfave/cli/v2"
)

func TargetCommand(version, commit, buildBranch, buildStamp string) *cli.Command {
	return &cli.Command{
		Name:  "target",
		Usage: "target specific OnCall dskit services",
		Flags: commonFlags,
		Action: func(context *cli.Context) error {
			return RunTargetServer(standalone.BuildInfo{
				Version:     version,
				Commit:      commit,
				BuildBranch: buildBranch,
				BuildStamp:  buildStamp,
			}, context)
		},
	}
}

func RunTargetServer(opts standalone.BuildInfo, cli *cli.Context) error {
	if Version {
		fmt.Printf("Version %s (commit: %s, branch: %s)")

		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			reason := fmt.Sprintf("%v", r)
			fmt.Printf("Critical error\r\n%s\r\n%s", reason, string(debug.Stack()))
			panic(r)
		}
	}()
	configOptions := strings.Split(ConfigOverrides, " ")
	cfg, err := setting.NewCfgFromArgs(setting.CommandLineArgs{
		Config:   ConfigFile,
		HomePath: HomePath,
		// tailing arguments have precedence over the options string
		Args: append(configOptions, cli.Args().Slice()...),
	})
	if err != nil {
		return err
	}

	s, err := server.Initialize(
		cfg,
		server.Options{
			Version:     opts.Version,
			Commit:      opts.Commit,
			BuildBranch: opts.BuildBranch,
		},
		api.ServerOptions{},
	)

	if err != nil {
		return err
	}

	ctx := context.Background()

	go listenToSystemSignals(ctx, s)

	return s.Run()
}
