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
	"github.com/train360-corp/projconf/pkg"
	"github.com/train360-corp/projconf/pkg/api"
	URL "net/url"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// shared by multiple commands, but with different default implementations
var (
	url            string
	adminApiKey    string
	clientSecretId string
	clientSecret   string

	environmentIdStr string
	environmentId    uuid.UUID

	projectIdStr string
	projectId    uuid.UUID
)

const (
	UrlFlag            string = "url"
	AdminApiKeyFlag    string = "admin-api-key"
	ClientSecretIdFlag string = "client-secret-id"
	ClientSecretFlag   string = "client-secret"
	EnvironmentIdFlag  string = "environment-id"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "projconf",
	Version:       pkg.Version,
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.MinimumNArgs(1),
	Short:         "A Supabase-powered project configuration utility.",
	Long: `ProjConf (short for Project Configuration) is a utility for creating, managing, and using secret and configuration parameters.

The CLI supports both hosting a ProjConf server instance (using the Supabase framework) and connecting to a remote instance from a local client.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		env := os.Environ()

		client, _ := api.GetAPIClient(url, adminApiKey, clientSecretId, clientSecret)
		if adminApiKey == "" {
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
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

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
		if url == "" {
			return errors.New("url not set (use --url flag or \"PROJCONF_URL\" environment variable)")
		} else {
			u, err := URL.Parse(url)
			if err != nil {
				return fmt.Errorf("could not parse URL: %v", err)
			} else if u.Scheme == "" || u.Host == "" { // require scheme and host at minimum
				return fmt.Errorf("invalid URL: %q", u)
			}
		}

		return nil
	},
}

func init() {

	addUrlFlag(rootCmd)
	rootCmd.Flags().StringVarP(&environmentIdStr, EnvironmentIdFlag, "e", "", "environment to run with")
	rootCmd.Flags().StringVar(&adminApiKey, AdminApiKeyFlag, "", "authenticate using admin api key")
	rootCmd.Flags().StringVar(&clientSecretId, ClientSecretIdFlag, "", "authenticate using a client")
	rootCmd.Flags().StringVar(&clientSecret, ClientSecretFlag, "", "secret for the client to authenticate with")

	rootCmd.MarkFlagsMutuallyExclusive(AdminApiKeyFlag, ClientSecretIdFlag)
	rootCmd.MarkFlagsOneRequired(AdminApiKeyFlag, ClientSecretFlag)
	rootCmd.MarkFlagsRequiredTogether(ClientSecretIdFlag, ClientSecretFlag)
	rootCmd.MarkFlagsRequiredTogether(AdminApiKeyFlag, EnvironmentIdFlag)

	viper.BindPFlags(rootCmd.Flags())
}

func CLI() *cobra.Command {
	return rootCmd
}

func addAuthFlags(cmd *cobra.Command) {
	addUrlFlag(cmd)
	addAdminApiKeyFlag(cmd)
	cmd.Flags().StringVar(&clientSecretId, ClientSecretIdFlag, "", "authenticate using a client")
	cmd.Flags().StringVar(&clientSecret, ClientSecretFlag, "", "secret for the client to authenticate with")

	cmd.MarkFlagsMutuallyExclusive(AdminApiKeyFlag, ClientSecretIdFlag)
	cmd.MarkFlagsOneRequired(AdminApiKeyFlag, ClientSecretFlag)
	cmd.MarkFlagsRequiredTogether(ClientSecretIdFlag, ClientSecretFlag)
}

func addAdminApiKeyFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&adminApiKey, AdminApiKeyFlag, "", "authenticate using admin api key")
}

func addUrlFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&url, UrlFlag, fmt.Sprintf("http://%s:%d", defaultServerHost, defaultServerPort), "ProjConf API URL")
}
