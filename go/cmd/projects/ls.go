/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package projects

import (
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/go/internal/flags"
	"github.com/train360-corp/projconf/go/internal/utils/tables"
	api2 "github.com/train360-corp/projconf/go/pkg/api"
)

var listProjectsCmd = &cobra.Command{
	Use:           "list",
	Aliases:       []string{"ls"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "List projects in a ProjConf server instance",
	RunE: func(c *cobra.Command, args []string) error {
		client, _ := api2.FromFlags(authFlags)
		resp, err := client.GetProjectsV1WithResponse(c.Context())
		if err != nil {
			return errors.New(fmt.Sprintf("request failed: %v", err.Error()))
		}

		if resp.JSON200 != nil {
			if len(*resp.JSON200) == 0 {
				fmt.Fprintln(c.OutOrStdout(), "no projects found")
			} else {
				fmt.Fprintln(c.OutOrStdout(), tables.Build(
					*resp.JSON200,
					tables.ColumnsByFieldNames[api2.Project]("Id", "Display"),
					tables.WithTitle("Projects"),
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
	flags.SetupAuthFlags(listProjectsCmd, authFlags)
	err := viper.BindPFlags(listProjectsCmd.Flags())
	if err != nil {
		panic(err)
	}
}
