package app

import (
	"fmt"
	"github.com/train360-corp/projconf/internal/app/commands"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/fs"
	"github.com/urfave/cli/v2"
	"log"
)

func Get() *cli.App {

	// ensure config
	config.MustLoad()

	// ensure root directory
	if root, err := fs.EnsureUserRoot(); err != nil {
		log.Fatal(fmt.Sprintf("failed to ensure system root: %s", err))
	} else {
		log.Printf("loaded from \"%s\"\n", root)
	}

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
