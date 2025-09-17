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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/go/internal/flags"
	"github.com/train360-corp/projconf/go/internal/utils/validators"
	api2 "github.com/train360-corp/projconf/go/pkg/api"
	"strings"
)

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
	RunE: func(c *cobra.Command, args []string) error {

		client, _ := api2.FromFlags(authFlags)
		resp, err := client.CreateEnvironmentV1WithResponse(c.Context(), projectId, api2.CreateEnvironmentV1JSONRequestBody{
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
			return errors.New(api2.GetAPIError(resp))
		}

		return nil
	},
}

func init() {
	createEnvironmentCmd.Flags().StringVar(&projectIdStr, flags.ProjectIdFlag, "", "the id of the project to list environments for")
	createEnvironmentCmd.MarkFlagRequired(flags.ProjectIdFlag)
	flags.SetupAuthFlags(createEnvironmentCmd, authFlags)
	err := viper.BindPFlags(createEnvironmentCmd.Flags())
	if err != nil {
		panic(err)
	}
}
