package app

import (
	"fmt"
	"github.com/train360-corp/projconf/internal/app/commands"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/urfave/cli/v2"
)

func Get() *cli.App {

	config.MustLoad()

	cli.VersionPrinter = func(ctx *cli.Context) {
		_, _ = fmt.Fprintf(ctx.App.Writer, "%v\n", ctx.App.Version)
	}

	return &cli.App{
		Name:    "projconf",
		Usage:   "A CLI utility for ProjConf",
		Version: Version,
		Commands: []*cli.Command{
			commands.ServerCommand(),
		},
	}
}
