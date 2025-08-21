/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package commands

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/urfave/cli/v2"
	"strings"
)

func EnvironmentsCommand() *cli.Command {
	return &cli.Command{
		Name:        "environments",
		Description: "manage environments in a ProjConf project",
		Subcommands: []*cli.Command{
			listEnvironmentsSubcommand(),
			createEnvironmentSubcommand(),
		},
	}
}

func listEnvironmentsSubcommand() *cli.Command {
	sharedCfg, flags := config.GetClientFlags()
	var projectIdStr string
	return &cli.Command{
		Name:        "ls",
		Usage:       "list accessible environments",
		Description: "list accessible environments",
		Flags: append(flags,
			&cli.StringFlag{
				Name:        "project-id",
				Usage:       "id of the project to return environments for",
				Destination: &projectIdStr,
				Required:    true,
			},
		),
		Action: func(c *cli.Context) error {

			projectId, err := uuid.Parse(projectIdStr)
			if err != nil {
				return cli.Exit(fmt.Sprintf("\"%s\" is not a valid project id (UUIDv4)", projectIdStr), 1)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, _ := getAPIClient(cfg, sharedCfg)
			resp, err := client.GetV1ProjectsProjectIdEnvironmentsWithResponse(c.Context, projectId)
			if err != nil {
				return cli.Exit(fmt.Sprintf("request failed: %v", err.Error()), 1)
			}

			if resp.JSON200 != nil {
				if len(*resp.JSON200) == 0 {
					fmt.Println("no environments found")
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

func createEnvironmentSubcommand() *cli.Command {

	var name string
	var projectIdStr string
	sharedCfg, flags := config.GetClientFlags()

	return &cli.Command{
		Name:        "create",
		Usage:       "create an environment",
		Description: "create an environment",
		Flags: append(flags,
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Usage:       "environment name",
				Destination: &name,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "project-id",
				Usage:       "id of the project to return environments for",
				Destination: &projectIdStr,
				Required:    true,
			},
		),
		Action: func(c *cli.Context) error {

			projectId, err := uuid.Parse(projectIdStr)
			if err != nil {
				return cli.Exit(fmt.Sprintf("\"%s\" is not a valid project id (UUIDv4)", projectIdStr), 1)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, _ := getAPIClient(cfg, sharedCfg)
			resp, err := client.PostV1ProjectsProjectIdEnvironmentsWithResponse(c.Context, projectId, api.PostV1ProjectsProjectIdEnvironmentsJSONRequestBody{
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
