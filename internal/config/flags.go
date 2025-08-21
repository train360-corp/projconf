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

func GetSharedFlags() (*Flags, []cli.Flag) {
	f := &Flags{}
	return f, []cli.Flag{
		&cli.StringFlag{
			Name:        "admin-api-key",
			Aliases:     []string{"A"},
			Usage:       "authenticate with an admin api key",
			EnvVars:     []string{"PROJCONF_ADMIN_ACCESS_KEY"},
			Value:       utils.RandomString(24),
			Destination: &f.AdminAPIKey,
		},
	}
}
