package cmd

import (
	"github.com/train360-corp/projconf/internal/version"
	"github.com/urfave/cli/v2"
	"os"
)

func Run() error {
	app := &cli.App{
		Name:    "projconf",
		Usage:   "A CLI utility for ProjConf",
		Version: version.Version,
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

	return app.Run(os.Args)
}
