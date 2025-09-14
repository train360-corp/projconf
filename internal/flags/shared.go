/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/defaults"
)

const (
	UrlFlag            string = "url"
	AdminApiKeyFlag    string = "admin-api-key"
	ClientSecretIdFlag string = "client-secret-id"
	ClientSecretFlag   string = "client-secret"
	EnvironmentIdFlag  string = "environment-id"
	ProjectIdFlag      string = "project-id"
)

type AuthFlags struct {
	Url            string
	AdminApiKey    string
	ClientSecretId string
	ClientSecret   string
}

func GetAuthFlags() *AuthFlags {
	return &AuthFlags{
		Url:            "",
		AdminApiKey:    "",
		ClientSecretId: "",
		ClientSecret:   "",
	}
}

func SetupAuthFlags(cmd *cobra.Command, flags *AuthFlags) {
	SetupUrlFlag(cmd, &flags.Url)
	SetupAdminApiKeyFlag(cmd, &flags.AdminApiKey)
	SetupClientSecretFlags(cmd, &flags.ClientSecretId, &flags.ClientSecret)

	cmd.MarkFlagsMutuallyExclusive(AdminApiKeyFlag, ClientSecretIdFlag)
	cmd.MarkFlagsOneRequired(AdminApiKeyFlag, ClientSecretFlag)
	cmd.MarkFlagsRequiredTogether(ClientSecretIdFlag, ClientSecretFlag)
}

func SetupClientSecretFlags(cmd *cobra.Command,
	clientSecretId *string,
	clientSecret *string,
) {
	cmd.Flags().StringVar(clientSecretId, ClientSecretIdFlag, "", "authenticate using a client")
	cmd.Flags().StringVar(clientSecret, ClientSecretFlag, "", "secret for the client to authenticate with")
}

func SetupAdminApiKeyFlag(cmd *cobra.Command, adminApiKey *string) {
	cmd.Flags().StringVar(adminApiKey, AdminApiKeyFlag, "", "authenticate using admin api key")
}

func SetupUrlFlag(cmd *cobra.Command, url *string) {
	defaultServerUrl := fmt.Sprintf("http://%s:%d", defaults.ServerHost, defaults.ServerPort)
	cmd.Flags().StringVar(url, UrlFlag, defaultServerUrl, fmt.Sprintf("url of a ProjConf server (default: %s)", defaultServerUrl))
}
