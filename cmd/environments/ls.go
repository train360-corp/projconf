/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package environments

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/internal/flags"
	"github.com/train360-corp/projconf/internal/utils/tables"
	"github.com/train360-corp/projconf/pkg/api"
)

var listEnvironmentsCmd = &cobra.Command{
	Use:           "list",
	Aliases:       []string{"ls"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "List environments in a ProjConf server instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(projectIdStr)
		if err != nil {
			return fmt.Errorf("\"%v\" is not a valid project id (%v)", projectIdStr, err)
		}
		projectId = id
		return nil
	},
	RunE: func(c *cobra.Command, args []string) error {
		client, _ := api.FromFlags(authFlags)
		resp, err := client.GetEnvironmentsV1WithResponse(c.Context(), projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("request failed: %v", err.Error()))
		}

		if resp.JSON200 != nil {
			if len(*resp.JSON200) == 0 {
				fmt.Fprintln(c.OutOrStdout(), "no environments found")
			} else {
				fmt.Fprintln(c.OutOrStdout(), tables.Build(
					*resp.JSON200,
					tables.ColumnsByFieldNames[api.Environment]("Id", "Display"),
					tables.WithTitle("Environments"),
					tables.WithStyle(table.StyleLight),
				))
			}
		} else {
			return errors.New(api.GetAPIError(resp))
		}

		return nil
	},
}

func init() {
	listEnvironmentsCmd.Flags().StringVar(&projectIdStr, flags.ProjectIdFlag, "", "the id of the project to list environments for")
	listEnvironmentsCmd.MarkFlagRequired(flags.ProjectIdFlag)
	flags.SetupAuthFlags(listEnvironmentsCmd, authFlags)
	err := viper.BindPFlags(listEnvironmentsCmd.Flags())
	if err != nil {
		panic(err)
	}
}
