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

var clientsCmd = &cobra.Command{
	Use:           "clients",
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "Manage clients in a ProjConf server instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

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
	RunE: func(cmd *cobra.Command, args []string) error {

		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)
		resp, err := client.GetClientsV1WithResponse(cmd.Context(), environmentId)
		if err != nil {
			return errors.New(fmt.Sprintf("request failed: %v", err.Error()))
		}

		if resp.JSON200 != nil {
			if len(*resp.JSON200) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no clients found")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), tables.Build(
					*resp.JSON200,
					tables.ColumnsByFieldNames[api.ClientRepresentation]("Id", "Display", "CreatedAt"),
					tables.WithTitle("Clients"),
					tables.WithStyle(table.StyleLight),
				))
			}
		} else {
			return errors.New(api.GetAPIError(resp))
		}

		return nil
	},
}

var createClientCmd = &cobra.Command{
	Use:           "create",
	Aliases:       []string{"new"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.ExactArgs(1),
	Short:         "Create a client in a ProjConf server instance",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !validators.IsValidDisplay(args[0]) {
			return fmt.Errorf("\"%v\" is not a valid display name", args[0])
		}
		id, err := uuid.Parse(environmentIdStr)
		if err != nil {
			return fmt.Errorf("\"%v\" is not a valid environment id (%v)", environmentIdStr, err)
		}
		environmentId = id
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)
		resp, err := client.CreateClientV1WithResponse(cmd.Context(), environmentId, api.CreateClientV1JSONRequestBody{
			Name: args[0],
		})
		if err != nil {
			return fmt.Errorf("request failed: %v", err.Error())
		}

		if resp.JSON201 != nil {
			fmt.Println(fmt.Sprintf("\"%v\"", resp.JSON201.Id))
		} else {
			if strings.Index(string(resp.Body), "\"ERROR: duplicate key value violates unique constraint") != -1 {
				return fmt.Errorf("client \"%s\" already exists", args[0])
			}
			return fmt.Errorf(api.GetAPIError(resp))
		}

		return nil
	},
}

func init() {

	listClientsCmd.Flags().StringVar(&environmentIdStr, "environment-id", "", "the id of the environment to list clients for")
	listClientsCmd.MarkFlagRequired("env-id")

	createClientCmd.Flags().StringVar(&environmentIdStr, "environment-id", "", "the id of the environment to create the client with")
	createClientCmd.MarkFlagRequired("env-id")

	addAuthFlags(listClientsCmd)
	addAuthFlags(createClientCmd)
	clientsCmd.AddCommand(listClientsCmd)
	clientsCmd.AddCommand(createClientCmd)
	viper.BindPFlags(listClientsCmd.Flags())
	viper.BindPFlags(createClientCmd.Flags())

	rootCmd.AddCommand(clientsCmd)
}
