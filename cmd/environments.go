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
	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/internal/utils/tables"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/train360-corp/projconf/pkg/api"
	"strings"
)

var environmentsCmd = &cobra.Command{
	Use:           "environments",
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "Manage environments in a ProjConf server instance",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return isReady()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)
		resp, err := client.GetEnvironmentsV1WithResponse(cmd.Context(), projectId)
		if err != nil {
			return errors.New(fmt.Sprintf("request failed: %v", err.Error()))
		}

		if resp.JSON200 != nil {
			if len(*resp.JSON200) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no environments found")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), tables.Build(
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

var createEnvironmentCmd = &cobra.Command{
	Use:           "create",
	Aliases:       []string{"new"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.ExactArgs(1),
	Short:         "Create an environment in a ProjConf server instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {

		id, err := uuid.Parse(projectIdStr)
		if err != nil {
			return fmt.Errorf("\"%v\" is not a valid project id (%v)", projectIdStr, err)
		}
		projectId = id

		if !validators.IsValidDisplay(args[0]) {
			return fmt.Errorf("\"%v\" is not a valid display name", args[0])
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)
		resp, err := client.CreateEnvironmentV1WithResponse(cmd.Context(), projectId, api.CreateEnvironmentV1JSONRequestBody{
			Name: args[0],
		})
		if err != nil {
			return fmt.Errorf("request failed: %v", err.Error())
		}

		if resp.JSON201 != nil {
			fmt.Println(fmt.Sprintf("\"%v\"", resp.JSON201.Id))
		} else {
			if strings.Index(string(resp.Body), "\"ERROR: duplicate key value violates unique constraint") != -1 {
				return fmt.Errorf("environment \"%s\" already exists", args[0])
			}
			return errors.New(api.GetAPIError(resp))
		}

		return nil
	},
}

func init() {

	projectIdFlag := "project-id"
	listEnvironmentsCmd.Flags().StringVar(&projectIdStr, projectIdFlag, "", "the id of the project to list environments for")
	listEnvironmentsCmd.MarkFlagRequired(projectIdFlag)

	createEnvironmentCmd.Flags().StringVar(&projectIdStr, projectIdFlag, "", "the id of the project to create the environment with")
	createEnvironmentCmd.MarkFlagRequired(projectIdFlag)

	addAuthFlags(listEnvironmentsCmd)
	addAuthFlags(createEnvironmentCmd)
	environmentsCmd.AddCommand(listEnvironmentsCmd)
	environmentsCmd.AddCommand(createEnvironmentCmd)
	viper.BindPFlags(listEnvironmentsCmd.Flags())
	viper.BindPFlags(createEnvironmentCmd.Flags())
	rootCmd.AddCommand(environmentsCmd)
}
