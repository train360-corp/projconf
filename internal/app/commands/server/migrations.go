package server

import (
	"fmt"
	"github.com/train360-corp/projconf/internal/supabase/migrations"
	"github.com/urfave/cli/v2"
)

func MigrationsCommand() *cli.Command {
	return &cli.Command{
		Name:  "migrations",
		Usage: "commands to migrate a ProjConf server",
		Subcommands: []*cli.Command{
			{
				Name:  "ls",
				Usage: "list all available migrations",
				Action: func(c *cli.Context) error {
					return migrations.ProcessMigrations(func(migration migrations.Migration) error {
						fmt.Println(fmt.Sprintf("%v: %s", migration.Number, migration.Name))
						return nil
					})
				},
			},
		},
	}
}
