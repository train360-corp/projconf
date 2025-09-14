/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package environments

import (
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/flags"
	"github.com/train360-corp/projconf/pkg/server"
)

var (
	authFlags    *flags.AuthFlags = flags.GetAuthFlags()
	projectIdStr string
	projectId    uuid.UUID
)

var Command = &cobra.Command{
	Use:           "environments",
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "Manage environments in a cmd server instance",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		return server.IsReady(authFlags.Url)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Command.AddCommand(listEnvironmentsCmd)
	Command.AddCommand(createEnvironmentCmd)
}
