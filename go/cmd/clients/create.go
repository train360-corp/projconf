/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package clients

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/go/internal/flags"
	"github.com/train360-corp/projconf/go/internal/utils/tables"
	api2 "github.com/train360-corp/projconf/go/pkg/api"
)

var listClientsCmd = &cobra.Command{
	Use:           "list",
	Aliases:       []string{"ls"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "List clients in a ProjConf server instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(environmentIdStr)
		if err != nil {
			return fmt.Errorf("\"%v\" is not a valid environment id (%v)", environmentIdStr, err)
		}
		environmentId = id
		return nil
	},
	RunE: func(c *cobra.Command, args []string) error {

		client, _ := api2.FromFlags(authFlags)
		resp, err := client.GetClientsV1WithResponse(c.Context(), environmentId)
		if err != nil {
			return errors.New(fmt.Sprintf("request failed: %v", err.Error()))
		}

		if resp.JSON200 != nil {
			if len(*resp.JSON200) == 0 {
				fmt.Fprintln(c.OutOrStdout(), "no clients found")
			} else {
				fmt.Fprintln(c.OutOrStdout(), tables.Build(
					*resp.JSON200,
					tables.ColumnsByFieldNames[api2.ClientRepresentation]("Id", "Display", "CreatedAt"),
					tables.WithTitle("Clients"),
					tables.WithStyle(table.StyleLight),
				))
			}
		} else {
			return errors.New(api2.GetAPIError(resp))
		}

		return nil
	},
}

func init() {
	createClientCmd.Flags().StringVar(&environmentIdStr, flags.EnvironmentIdFlag, "", "the id of the environment to list clients for")
	createClientCmd.MarkFlagRequired(flags.EnvironmentIdFlag)
	flags.SetupAuthFlags(createClientCmd, authFlags)
	err := viper.BindPFlags(createClientCmd.Flags())
	if err != nil {
		panic(err)
	}
}
