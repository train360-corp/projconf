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
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/urfave/cli/v2"
	"reflect"
	"strings"
)

func AuthCommand() *cli.Command {
	flags := struct {
		Host         string
		ClientId     string
		ClientSecret string
	}{}
	return &cli.Command{
		Name:  "auth",
		Usage: "authenticate to a ProjConf server",
		Subcommands: []*cli.Command{
			{
				Name:  "set",
				Usage: "update the default account settings used by the cli",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "host",
						Aliases:     []string{"H"},
						Destination: &flags.Host,
					},
					&cli.StringFlag{
						Name:        "client-id",
						Aliases:     []string{"U"},
						Destination: &flags.ClientId,
					},
					&cli.StringFlag{
						Name:        "client-secret",
						Aliases:     []string{"P"},
						Destination: &flags.ClientSecret,
					},
				},
				Action: func(ctx *cli.Context) error {

					cfg, err := config.Load()
					if err != nil {
						return err
					}
					v := reflect.ValueOf(flags)
					t := reflect.TypeOf(flags)

					noneSet := true
					for i := 0; i < t.NumField(); i++ {
						field := t.Field(i)
						value := v.Field(i)
						if value.Kind() != reflect.Ptr {
							switch value.Kind() {
							case reflect.String:
								if strings.Trim(value.String(), " ") != "" {
									noneSet = false
									fmt.Printf("set %s = %v\n", field.Name, value.Interface())
									switch field.Name {
									case "ClientId":
										if !validators.IsValidUUID(value.String()) {
											return cli.Exit("--client-id is not a valid UUID", 1)
										}
										cfg.Account.Client.Id = value.String()
									case "ClientSecret":
										cfg.Account.Client.Secret = value.String()
									case "Host":
										if err := validators.ValidateHTTPHostURL(flags.Host); err != nil {
											return cli.Exit(fmt.Sprintf("invalid host: %s (%s)", flags.Host, err.Error()), 1)
										} else {
											cfg.Account.Url = flags.Host
										}
									default:
										panic(fmt.Errorf("string field '%s' unhandled", field.Name))
									}

								}
							default:
								panic(fmt.Errorf("invalid value type: %s", value.Kind().String()))
							}
						}
					}

					if noneSet {
						return cli.Exit("no values were set", 1)
					}

					return cfg.Flush()
				},
			},
		},
	}
}
