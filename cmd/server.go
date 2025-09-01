/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/commands/server/serve"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/train360-corp/projconf/pkg/server"
	"go.uber.org/zap/zapcore"
	"strings"
)

var (
	logLevelStr string = "info"
)

var serverCmd = &cobra.Command{
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

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:           "serve",
	Aliases:       []string{"start", "run"},
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.NoArgs,
	Short:         "host a ProjConf server instance",
	Long: `Create an initialize a ProjConf server.

Default values will be created and stored in a local 
file accessible only by the current user.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {

		if strings.TrimSpace(server.AdminApiKey) == "" {
			server.AdminApiKey = utils.RandomString(32)
		}

		if level, err := zapcore.ParseLevel(logLevelStr); err != nil {
			return fmt.Errorf("invalid log level (\"%s\"): %v", logLevelStr, err)
		} else {
			serve.LogLevel = level
		}
		return nil
	},
	Run: server.Command,
}

func init() {
	serverCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&server.AdminApiKey, AdminApiKeyFlag, server.AdminApiKey, "authentication token for an admin api client")
	serveCmd.Flags().StringVarP(&server.Host, "host", "H", server.Host, fmt.Sprintf("host to serve on (default: %s)", server.Host))
	serveCmd.Flags().Uint16VarP(&server.Port, "port", "P", server.Port, fmt.Sprintf("port to serve on (default: %d)", server.Port))
	serveCmd.Flags().StringVarP(&logLevelStr, "log-level", "l", logLevelStr, "log level (default: warn; available: debug | info | warn | error | panic | fatal)")
	serveCmd.Flags().BoolVar(&serve.LogJsonFmt, "log-json", serve.LogJsonFmt, "log json-formatted output (default: human-readable)")

	rootCmd.AddCommand(serverCmd)
}
