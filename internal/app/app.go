package app

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

func Get() *cli.App {
	cli.VersionPrinter = func(ctx *cli.Context) {
		_, _ = fmt.Fprintf(ctx.App.Writer, "%v\n", ctx.App.Version)
	}

	return &cli.App{
		Name:    "projconf",
		Usage:   "A CLI utility for ProjConf",
		Version: Version,
		Commands: []*cli.Command{
			{
				Name:  "hello-world",
				Usage: "Print a greeting",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	}
}
