package commands

import (
	"github.com/train360-corp/projconf/internal/supabase"
	"github.com/urfave/cli/v2"
)

func AuthCommand() *cli.Command {
	return &cli.Command{
		Name:  "auth",
		Usage: "authenticate to a ProjConf server",
		Subcommands: []*cli.Command{
			authCommand(),
		},
	}
}

func authCommand() *cli.Command {

	sb := supabase.Config{}

	return &cli.Command{
		Name:  "login",
		Usage: "authenticate to a ProjConf server",
		Flags: supabase.GetConfigFlags(&sb),
		Action: func(c *cli.Context) error {

		},
	}
}
