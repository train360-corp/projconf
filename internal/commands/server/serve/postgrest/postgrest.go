/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package postgrest

import (
	"fmt"
	"github.com/train360-corp/projconf/internal/commands/server/serve/postgres"
	"github.com/train360-corp/projconf/internal/commands/server/serve/types"
)

type ServiceRequest struct {
	JwtSecret        string
	PostgresPassword string
}

func Service(req ServiceRequest) *types.Service {
	return &types.Service{
		Image: "postgrest/postgrest:v12.2.12",
		Name:  "projconf-internal-supabase-postgrest",
		Cmd: []string{
			"postgrest",
		},
		Labels: map[string]string{
			"projconf.service": "postgrest",
		},
		Ports: []uint16{3000, 3001},
		Env: []string{
			fmt.Sprintf("PGRST_DB_URI=postgres://authenticator:%s@%s:5432/postgres", req.PostgresPassword, postgres.ContainerName),
			"PGRST_DB_SCHEMAS=public",
			"PGRST_DB_ANON_ROLE=anon",
			fmt.Sprintf("PGRST_JWT_SECRET=%s", req.JwtSecret),
			"PGRST_DB_USE_LEGACY_GUCS=false",
			fmt.Sprintf("PGRST_APP_SETTINGS_JWT_SECRET=%s", req.JwtSecret),
			"PGRST_APP_SETTINGS_JWT_EXP=3600",
			"PGRST_ADMIN_SERVER_PORT=3001",
		},
	}
}
