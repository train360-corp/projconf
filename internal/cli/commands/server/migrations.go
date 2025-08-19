/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

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
