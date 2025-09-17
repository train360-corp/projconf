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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/go/internal/flags"
	"github.com/train360-corp/projconf/go/internal/utils/validators"
	api2 "github.com/train360-corp/projconf/go/pkg/api"
	"strings"
)

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
	RunE: func(c *cobra.Command, args []string) error {

		client, _ := api2.FromFlags(authFlags)
		resp, err := client.CreateProjectV1WithResponse(c.Context(), api2.CreateProjectV1JSONRequestBody{
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
			return errors.New(api2.GetAPIError(resp))
		}

		return nil
	},
}

func init() {
	flags.SetupAuthFlags(createProjectCmd, authFlags)
	err := viper.BindPFlags(createProjectCmd.Flags())
	if err != nil {
		panic(err)
	}
}
