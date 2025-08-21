/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package commands

import (
	"fmt"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/urfave/cli/v2"
	"strings"
)

func ProjectsCommand() *cli.Command {
	return &cli.Command{
		Name:        "projects",
		Description: "manage projects in a ProjConf instance",
		Subcommands: []*cli.Command{
			listProjectsSubcommand(),
			createProjectSubcommand(),
		},
	}
}

func createProjectSubcommand() *cli.Command {

	var name string
	sharedCfg, flags := config.GetClientFlags()

	return &cli.Command{
		Name:        "create",
		Usage:       "create a project",
		Description: "create a project",
		Flags: append(flags, &cli.StringFlag{
			Name:        "name",
			Aliases:     []string{"n"},
			Usage:       "project name",
			Destination: &name,
			Required:    true,
		}),
		Action: func(c *cli.Context) error {

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, _ := getAPIClient(cfg, sharedCfg)
			resp, err := client.PostV1ProjectsWithResponse(c.Context, api.PostV1ProjectsJSONRequestBody{
				Name: name,
			})
			if err != nil {
				return cli.Exit(fmt.Sprintf("request failed: %v", err.Error()), 1)
			}

			if resp.JSON201 != nil {
				fmt.Println(fmt.Sprintf("\"%v\"", resp.JSON201.Id))
			} else {

				if strings.Index(string(resp.Body), "\"ERROR: duplicate key value violates unique constraint") != -1 {
					return cli.Exit(fmt.Sprintf("project \"%s\" already exists", name), 1)
				}

				return cli.Exit(utils.GetAPIError(resp), 1)
			}

			return nil
		},
	}
}

func listProjectsSubcommand() *cli.Command {

	sharedCfg, flags := config.GetClientFlags()

	return &cli.Command{
		Name:        "ls",
		Usage:       "list accessible projects",
		Description: "list accessible projects",
		Flags:       flags,
		Action: func(c *cli.Context) error {

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, _ := getAPIClient(cfg, sharedCfg)
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
