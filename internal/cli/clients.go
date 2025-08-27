/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cli

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/urfave/cli/v2"
	"strings"
)

func ClientsCommand() *cli.Command {
	return &cli.Command{
		Name:        "clients",
		Description: "manage clients in a ProjConf environment",
		Subcommands: []*cli.Command{
			listClientsSubcommand(),
			createClientSubcommand(),
		},
	}
}

func listClientsSubcommand() *cli.Command {
	sharedCfg, flags := config.GetClientFlags()
	var environmentIdStr string
	return &cli.Command{
		Name:        "ls",
		Usage:       "list accessible clients",
		Description: "list accessible clients",
		Flags: append(flags,
			&cli.StringFlag{
				Name:        "environment-id",
				Usage:       "id of the environment to create clients for",
				Destination: &environmentIdStr,
				Required:    true,
			},
		),
		Action: func(c *cli.Context) error {

			environmentId, err := uuid.Parse(environmentIdStr)
			if err != nil {
				return cli.Exit(fmt.Sprintf("\"%s\" is not a valid environment id (UUIDv4)", environmentIdStr), 1)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, _ := getAPIClient(cfg, sharedCfg)
			resp, err := client.GetClientsV1WithResponse(c.Context, environmentId)
			if err != nil {
				return cli.Exit(fmt.Sprintf("request failed: %v", err.Error()), 1)
			}

			if resp.JSON200 != nil {
				if len(*resp.JSON200) == 0 {
					fmt.Println("no clients found")
				}
				for _, client := range *resp.JSON200 {
					fmt.Println(fmt.Sprintf("%s: %s", client.Id, client.Display))
				}
			} else {
				return cli.Exit(utils.GetAPIError(resp), 1)
			}

			return nil
		},
	}
}

func createClientSubcommand() *cli.Command {

	var name string
	var environmentIdStr string
	sharedCfg, flags := config.GetClientFlags()

	return &cli.Command{
		Name:        "create",
		Usage:       "create a client",
		Description: "create a client",
		Flags: append(flags,
			&cli.StringFlag{
				Name:        "name",
				Aliases:     []string{"n"},
				Usage:       "client name",
				Destination: &name,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "environment-id",
				Usage:       "id of the environment to create clients for",
				Destination: &environmentIdStr,
				Required:    true,
			},
		),
		Action: func(c *cli.Context) error {

			environmentId, err := uuid.Parse(environmentIdStr)
			if err != nil {
				return cli.Exit(fmt.Sprintf("\"%s\" is not a valid environment id (UUIDv4)", environmentIdStr), 1)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			client, _ := getAPIClient(cfg, sharedCfg)
			resp, err := client.CreateClientV1WithResponse(c.Context, environmentId, api.CreateClientV1JSONRequestBody{
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
