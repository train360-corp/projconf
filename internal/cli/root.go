/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cli

import (
	"fmt"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/fs"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
)

func Get() *cli.App {

	// ensure root directory
	if _, err := fs.EnsureUserRoot(); err != nil {
		log.Fatal(fmt.Sprintf("failed to ensure system root: %s", err))
	}
	cli.VersionPrinter = func(ctx *cli.Context) {
		_, _ = fmt.Fprintf(ctx.App.Writer, "%v\n", ctx.App.Version)
	}

	return &cli.App{
		Name:        "projconf",
		Usage:       "projconf [args...]",
		Description: "A CLI utility for ProjConf",
		Version:     config.Version,
		Commands: []*cli.Command{
			ServerCommand(),
			AuthCommand(),
			ProjectsCommand(),
			EnvironmentsCommand(),
			VariablesCommand(),
			ClientsCommand(),
		},
		Action: func(c *cli.Context) error {

			if c.Args().Len() == 0 {
				return cli.ShowAppHelp(c)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, _ := getAPIClient(cfg, nil)
			if resp, err := client.GetClientSecretsV1WithResponse(c.Context); err != nil {
				return cli.Exit(fmt.Sprintf("unable to get secrets: %s", err.Error()), 1)
			} else if resp.JSON200 == nil {
				return cli.Exit(utils.GetAPIError(resp), 1)
			} else if len(*resp.JSON200) == 0 {
				return cli.Exit("no secrets found", 1)
			} else {

				env := os.Environ()
				for _, secret := range *resp.JSON200 {
					env = append(env, fmt.Sprintf("%s=%s", secret.Variable.Key, secret.Value))
				}

				cmd := exec.Command(c.Args().Slice()[0], c.Args().Slice()[1:]...)
				cmd.Env = env

			}

			return nil
		},
	}
}
