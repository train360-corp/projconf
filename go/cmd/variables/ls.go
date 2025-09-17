/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package variables

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

var listVariablesCmd = &cobra.Command{
	Use:           "list",
	Aliases:       []string{"ls"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "List variables in a ProjConf server instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(projectIdStr)
		if err != nil {
			return fmt.Errorf("\"%v\" is not a valid project id (%v)", projectIdStr, err)
		}
		projectId = id
		return nil
	},
	RunE: func(c *cobra.Command, args []string) error {
		client, _ := api2.FromFlags(authFlags)
		resp, err := client.GetVariablesV1WithResponse(c.Context(), projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("request failed: %v", err.Error()))
		}

		if resp.JSON200 != nil {
			if len(*resp.JSON200) == 0 {
				fmt.Fprintln(c.OutOrStdout(), "no variables found")
			} else {
				fmt.Fprintln(c.OutOrStdout(), tables.Build[api2.Variable](
					*resp.JSON200,
					tables.ColumnsByFieldNames[api2.Variable]("Id", "Key", "GeneratorType", "GeneratorData"),
					tables.WithTitle("Variables"),
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
	listVariablesCmd.Flags().StringVar(&projectIdStr, flags.ProjectIdFlag, "", "the id of the project to list variables for")
	listVariablesCmd.MarkFlagRequired(flags.ProjectIdFlag)
	flags.SetupAuthFlags(listVariablesCmd, authFlags)
	viper.BindPFlags(listVariablesCmd.Flags())
}
