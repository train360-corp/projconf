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
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/train360-corp/projconf/cmd/clients"
	"github.com/train360-corp/projconf/cmd/environments"
	"github.com/train360-corp/projconf/cmd/projects"
	srv "github.com/train360-corp/projconf/cmd/server"
	"github.com/train360-corp/projconf/cmd/variables"
	"github.com/train360-corp/projconf/internal/flags"
	"github.com/train360-corp/projconf/pkg"
	"github.com/train360-corp/projconf/pkg/api"
	"github.com/train360-corp/projconf/pkg/server"
	URL "net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	authFlags        *flags.AuthFlags = flags.GetAuthFlags()
	environmentIdStr string
)

var preRun = func(cmd *cobra.Command, args []string) error {

	// see https://github.com/carolynvs/stingoftheviper
	v := viper.New()

	// When we bind flags to environment variables expect that the
	// environment variables are prefixed, e.g. a flag like --number
	// binds to an environment variable PROJCONF_NUMBER. This helps
	// avoid conflicts.
	v.SetEnvPrefix("PROJCONF")

	// Environment variables can't have dashes in them, so bind them to their equivalent
	// keys with underscores, e.g. --favorite-color to PROJCONF_FAVORITE_COLOR
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Bind to environment variables
	// Works great for simple config names, but needs help for names
	// like --favorite-color which we fix in the bindFlags function
	v.AutomaticEnv()

	// Bind the current command's flags to viper
	// Bind each cobra flag to its associated viper configuration (environment variable)
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})

	// handle validation
	if authFlags.Url == "" {
		return errors.New("url not set (use --url flag or \"PROJCONF_URL\" environment variable)")
	} else {
		u, err := URL.Parse(authFlags.Url)
		if err != nil {
			return fmt.Errorf("could not parse URL: %v", err)
		} else if u.Scheme == "" || u.Host == "" { // require scheme and host at minimum
			return fmt.Errorf("invalid URL: %q", u)
		}
	}

	return nil
}

var run = func(cmd *cobra.Command, args []string) error {

	env := os.Environ()

	// server must have ready state
	if err := server.IsReady(authFlags.Url); err != nil {
		return err
	}

	client, _ := api.FromFlags(authFlags)
	if authFlags.AdminApiKey == "" {
		if resp, err := client.GetClientSecretsV1WithResponse(cmd.Context()); err != nil {
			return fmt.Errorf("could not get client secrets: %v", err)
		} else if resp.JSON200 == nil {
			return fmt.Errorf("could not get client secrets: %v", api.GetAPIError(resp))
		} else if len(*resp.JSON200) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), color.YellowString("WARN: no secrets found"))
		} else {
			for _, secret := range *resp.JSON200 {
				env = append(env, fmt.Sprintf("%s=%s", secret.Variable.Key, secret.Value))
			}
		}
	} else {
		envId, err := uuid.Parse(environmentIdStr)
		if err != nil {
			return fmt.Errorf("\"%v\" is not a valid environment id (%v)", environmentIdStr, err)
		}
		if resp, err := client.GetEnvironmentSecretsV1WithResponse(cmd.Context(), envId); err != nil {
			return fmt.Errorf("could not get client secrets: %v", err)
		} else if resp.JSON200 == nil {
			return fmt.Errorf("could not get client secrets: %v", api.GetAPIError(resp))
		} else if len(*resp.JSON200) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), color.YellowString("WARN: no secrets found"))
		} else {
			for _, secret := range *resp.JSON200 {
				env = append(env, fmt.Sprintf("%s=%s", secret.Variable.Key, secret.Value))
			}
		}
	}

	c := exec.CommandContext(cmd.Context(), args[0], args[1:]...)
	c.Env = env
	c.Stdin = nil
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if ws, ok := exitError.Sys().(syscall.WaitStatus); ok {
				return fmt.Errorf("command exited with status: %d", ws.ExitStatus())
			} else {
				return fmt.Errorf("command exited with undeterminable status")
			}
		}
	}
	return nil
}

// cmd represents the base command when called without any subcommands
var cmd = &cobra.Command{
	Use:           "projconf",
	Version:       pkg.Version,
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.MinimumNArgs(1),
	Short:         "A Supabase-powered project configuration utility.",
	Long: `cmd (short for Project Configuration) is a utility for creating, managing, and using secret and configuration parameters.

The CLI supports both hosting a cmd server instance (using the Supabase framework) and connecting to a remote instance from a local client.`,
	PersistentPreRunE: preRun,
	RunE:              run,
}

func init() {
	flags.SetupAuthFlags(cmd, authFlags)
	cmd.Flags().StringVarP(&environmentIdStr, flags.EnvironmentIdFlag, "e", "", "environment to run with")
	cmd.MarkFlagsRequiredTogether(flags.AdminApiKeyFlag, flags.EnvironmentIdFlag)
	viper.BindPFlags(cmd.Flags())

	cmd.AddCommand(variables.Command)
	cmd.AddCommand(srv.Command)
	cmd.AddCommand(projects.Command)
	cmd.AddCommand(environments.Command)
	cmd.AddCommand(clients.Command)
}

func ProjConf() *cobra.Command {
	return cmd
}
