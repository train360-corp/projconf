/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package config

import (
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/urfave/cli/v2"
)

type Flags struct {
	AdminAPIKey string
}

func getAdminAPIKeyFlag(def string, f *Flags) cli.Flag {
	return &cli.StringFlag{
		Name:        "admin-api-key",
		Aliases:     []string{"A"},
		Usage:       "authenticate with an admin api key",
		EnvVars:     []string{"PROJCONF_ADMIN_ACCESS_KEY"},
		Value:       def,
		Destination: &f.AdminAPIKey,
	}
}

func GetServerFlags() (*Flags, []cli.Flag) {
	f := &Flags{}
	return f, []cli.Flag{
		getAdminAPIKeyFlag(utils.RandomString(24), f),
	}
}

func GetClientFlags() (*Flags, []cli.Flag) {
	f := &Flags{}
	return f, []cli.Flag{
		getAdminAPIKeyFlag("", f),
	}
}
