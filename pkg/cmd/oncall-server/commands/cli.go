package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/InariTheFox/oncall/pkg/api"
	"github.com/InariTheFox/oncall/pkg/server"
	"github.com/InariTheFox/oncall/pkg/services/apiserver/standalone"
	"github.com/InariTheFox/oncall/pkg/setting"
)

func ServerCommand(version, commit, buildBranch, buildStamp string) *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "run the OnCall server",
		Flags: commonFlags,
		Action: func(context *cli.Context) error {
			return RunServer(standalone.BuildInfo{
				Version:     version,
				Commit:      commit,
				BuildBranch: buildBranch,
				BuildStamp:  buildStamp,
			}, context)
		},
		Subcommands: []*cli.Command{TargetCommand(version, commit, buildBranch, buildStamp)},
	}
}

func RunServer(opts standalone.BuildInfo, cli *cli.Context) error {
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

	SetBuildInfo(opts)

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

type gserver interface {
	Shutdown(context.Context, string) error
}

func listenToSystemSignals(ctx context.Context, s gserver) {
	signalChan := make(chan os.Signal, 1)
	sighupChan := make(chan os.Signal, 1)

	signal.Notify(sighupChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-sighupChan:
		case sig := <-signalChan:
			ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			if err := s.Shutdown(ctx, fmt.Sprintf("System signal: %s", sig)); err != nil {
				fmt.Fprintf(os.Stderr, "Timed out waiting for server to shut down\n")
			}
			return
		}
	}
}
