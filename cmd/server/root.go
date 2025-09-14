/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
)

var (
	logJsonFmt  bool          = false
	logLevelStr string        = "info"
	logLevel    zapcore.Level = zapcore.InfoLevel
)

var Command = &cobra.Command{
	Use:           "server",
	Aliases:       []string{"start", "run"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "Manage a ProjConf server instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	Command.AddCommand(serveCommand)
}
