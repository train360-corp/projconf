/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package commands

import (
	"context"
	"fmt"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/urfave/cli/v2"
	"net/http"
)

func ProjectsCommand() *cli.Command {
	return &cli.Command{
		Name:        "projects",
		Description: "manage projects in a ProjConf instance",
		Subcommands: []*cli.Command{
			ls(),
		},
	}
}

func ls() *cli.Command {
	return &cli.Command{
		Name:        "ls",
		Usage:       "list accessible projects",
		Description: "list accessible projects",
		Action: func(c *cli.Context) error {

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, err := api.NewClientWithResponses(cfg.Account.Url, api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
				req.Header.Add("x-client-secret-id", cfg.Account.Client.Id)
				req.Header.Add("x-client-secret", cfg.Account.Client.Secret)
				return nil
			}))

			resp, err := client.GetV1ProjectsWithResponse(c.Context)
			if err != nil {
				return cli.Exit(fmt.Sprintf("request failed: %v", err.Error()), 1)
			}

			if resp.JSON200 != nil {
				if len(*resp.JSON200) == 0 {
					fmt.Println("no projects found")
				}
				for _, secret := range *resp.JSON200 {
					fmt.Println(fmt.Sprintf("%s: %s", secret.Id, secret.Display))
				}
			} else {
				return cli.Exit(utils.GetAPIError(resp), 1)
			}

			return nil
		},
	}
}
