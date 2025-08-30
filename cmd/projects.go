/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cmd

import (
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/internal/utils/tables"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/train360-corp/projconf/pkg/api"
	"strings"
)

var projectsCmd = &cobra.Command{
	Use:           "projects",
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "Manage projects in a ProjConf server instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var listProjectsCmd = &cobra.Command{
	Use:           "list",
	Aliases:       []string{"ls"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "List projects in a ProjConf server instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)
		resp, err := client.GetProjectsV1WithResponse(cmd.Context())
		if err != nil {
			return errors.New(fmt.Sprintf("request failed: %v", err.Error()))
		}

		if resp.JSON200 != nil {
			if len(*resp.JSON200) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no projects found")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), tables.Build(
					*resp.JSON200,
					tables.ColumnsByFieldNames[api.Project]("Id", "Display"),
					tables.WithTitle("Projects"),
					tables.WithStyle(table.StyleLight),
				))
			}
		} else {
			return errors.New(api.GetAPIError(resp))
		}

		return nil
	},
}

var createProjectCmd = &cobra.Command{
	Use:           "create",
	Aliases:       []string{"new"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.ExactArgs(1),
	Short:         "Create a project in a ProjConf server instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		if !validators.IsValidDisplay(args[0]) {
			return fmt.Errorf("\"%v\" is not a valid display name", args[0])
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)
		resp, err := client.CreateProjectV1WithResponse(cmd.Context(), api.CreateProjectV1JSONRequestBody{
			Name: args[0],
		})
		if err != nil {
			return fmt.Errorf("request failed: %v", err.Error())
		}

		if resp.JSON201 != nil {
			fmt.Println(fmt.Sprintf("\"%v\"", resp.JSON201.Id))
		} else {
			if strings.Index(string(resp.Body), "\"ERROR: duplicate key value violates unique constraint") != -1 {
				return fmt.Errorf("project \"%s\" already exists", args[0])
			}
			return errors.New(api.GetAPIError(resp))
		}

		return nil
	},
}

func init() {
	addAuthFlags(listProjectsCmd)
	addAuthFlags(createProjectCmd)
	projectsCmd.AddCommand(listProjectsCmd)
	projectsCmd.AddCommand(createProjectCmd)
	viper.BindPFlags(listProjectsCmd.Flags())
	viper.BindPFlags(createProjectCmd.Flags())
	rootCmd.AddCommand(projectsCmd)
}
