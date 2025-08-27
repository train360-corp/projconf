/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

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

func VariablesCommand() *cli.Command {
	return &cli.Command{
		Name:        "variables",
		Description: "manage variables in a ProjConf project",
		Subcommands: []*cli.Command{
			listVariablesSubcommand(),
			createVariableSubcommand(),
		},
	}
}

func listVariablesSubcommand() *cli.Command {
	sharedCfg, flags := config.GetClientFlags()
	var projectIdStr string
	return &cli.Command{
		Name:        "ls",
		Usage:       "list accessible variables",
		Description: "list accessible variables",
		Flags: append(flags,
			&cli.StringFlag{
				Name:        "project-id",
				Usage:       "id of the project to return variables for",
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
			resp, err := client.GetVariablesV1WithResponse(c.Context, projectId)
			if err != nil {
				return cli.Exit(fmt.Sprintf("request failed: %v", err.Error()), 1)
			}

			if resp.JSON200 != nil {
				if len(*resp.JSON200) == 0 {
					fmt.Println("no variables found")
				}
				for _, secret := range *resp.JSON200 {
					fmt.Println(fmt.Sprintf("%s: %s - %s", secret.Id, secret.Key, secret.Description))
				}
			} else {
				return cli.Exit(utils.GetAPIError(resp), 1)
			}

			return nil
		},
	}
}

func createVariableSubcommand() *cli.Command {

	var name string
	var projectIdStr string
	var typ string

	// static generator
	var static string

	// random generator
	var length int
	var nums, ltrs, symbs bool

	sharedCfg, flags := config.GetClientFlags()

	return &cli.Command{
		Name:        "create",
		Usage:       "create a variable",
		Description: "create a variable",
		Flags: append(flags,
			&cli.StringFlag{
				Name:        "key",
				Aliases:     []string{"k"},
				Usage:       "variable key",
				Destination: &name,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "project-id",
				Usage:       "id of the project to return environments for",
				Destination: &projectIdStr,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "type",
				Usage:       "STATIC | RANDOM",
				Destination: &typ,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "static",
				Usage:       "the static value to use (only applied when --type=STATIC)",
				Destination: &static,
				Required:    false,
			},
			&cli.IntFlag{
				Name:        "length",
				Usage:       "length of the random value (min: 1)",
				Destination: &length,
				Required:    false,
			},
			&cli.BoolFlag{
				Name:        "numbers",
				Usage:       "whether to use numbers",
				Destination: &nums,
				Required:    false,
			},
			&cli.BoolFlag{
				Name:        "letters",
				Usage:       "whether to use letters",
				Destination: &ltrs,
				Required:    false,
			},
			&cli.BoolFlag{
				Name:        "symbols",
				Usage:       "whether to use symbols",
				Destination: &symbs,
				Required:    false,
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
			req := api.CreateVariableV1JSONRequestBody{
				Key: name,
			}

			switch typ {
			case string(api.GeneratorTypeSTATIC):
				req.Generator.FromSecretGeneratorStatic(api.SecretGeneratorStatic{
					Type: api.SecretGeneratorStaticType(typ),
					Data: static,
				})
			case string(api.GeneratorTypeRANDOM):
				req.Generator.FromSecretGeneratorRandom(api.SecretGeneratorRandom{
					Type: api.SecretGeneratorRandomType(typ),
					Data: api.RandomGeneratorData{
						Length:  float32(length),
						Letters: ltrs,
						Symbols: symbs,
						Numbers: nums,
					},
				})
			default:
				return cli.Exit(fmt.Sprintf("\"%s\" is not a handled variable type", typ), 1)
			}

			resp, err := client.CreateVariableV1WithResponse(c.Context, projectId, req)
			if err != nil {
				return cli.Exit(fmt.Sprintf("request failed: %v", err.Error()), 1)
			}

			if resp.JSON201 != nil {
				fmt.Println(fmt.Sprintf("\"%v\"", resp.JSON201.Id))
			} else {
				if strings.Index(string(resp.Body), "\"ERROR: duplicate key value violates unique constraint") != -1 {
					return cli.Exit(fmt.Sprintf("variable \"%s\" already exists", name), 1)
				}
				return cli.Exit(utils.GetAPIError(resp), 1)
			}

			return nil
		},
	}
}
