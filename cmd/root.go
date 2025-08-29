/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cmd

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// shared by multiple commands, but with different default implementations
var adminApiKey string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "projconf",
	Short: "A Supabase-powered project configuration utility.",
	Long: `ProjConf (short for Project Configuration) is a utility for creating, managing, and using secret and configuration parameters.

The CLI supports both hosting a ProjConf server instance (using the Supabase framework) and connecting to a remote instance from a local client.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	rootCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	err := rootCmd.ExecuteContext(rootCtx)
	if err != nil {
		if rootCtx.Err() != nil {
			// if context was canceled due to signal, exit code 130 is conventional for SIGINT
			os.Exit(130)
		}
		fmt.Fprintln(os.Stderr, color.RedString("error: %v", err))
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		// config with environment variables
		viper.SetEnvPrefix("PROJCONF")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*")) // Allow for nested keys in environment variables (e.g. `PROJCONF_DATABASE_HOST`)

		// config with file
		//if cfgFile != "" { // use config file from the flag.
		//	viper.SetConfigFile(cfgFile)
		//} else {
		//	home, err := fs.EnsureUserRoot()
		//	cobra.CheckErr(err)
		//	viper.AddConfigPath(home)
		//	viper.SetConfigType("yaml")
		//	viper.SetConfigName(".projconf")
		//}

		viper.AutomaticEnv()
		//if err := viper.ReadInConfig(); err == nil {
		//	fmt.Println("using config file:", viper.ConfigFileUsed())
		//}
	})
}
