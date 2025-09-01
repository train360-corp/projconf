/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package postgres

import (
	_ "embed"
	"github.com/train360-corp/projconf/internal/commands/server/serve/types"
)

//go:embed realtime.sql
var RealtimeSQL []byte

//go:embed webhooks.sql
var WebhooksSQL []byte

//go:embed roles.sql
var RolesSQL []byte

//go:embed jwt.sql
var JwtSQL []byte

//go:embed _supabase.sql
var SupabaseSQL []byte

//go:embed logs.sql
var LogsSQL []byte

//go:embed pooler.sql
var PoolerSQL []byte

func getTempFiles() []types.TempFile {
	return []types.TempFile{
		{
			Name:          "realtime.sql",
			Data:          RealtimeSQL,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-realtime.sql:ro",
		},
		{
			Name:          "webhooks.sql",
			Data:          WebhooksSQL,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql:ro",
		},
		{
			Name:          "roles.sql",
			Data:          RolesSQL,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/99-roles.sql:ro",
		},
		{
			Name:          "jwt.sql",
			Data:          JwtSQL,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql:ro",
		},
		{
			Name:          "_supabase.sql",
			Data:          SupabaseSQL,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/97-_supabase.sql:ro",
		},
		{
			Name:          "logs.sql",
			Data:          LogsSQL,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-logs.sql:ro",
		},
		{
			Name:          "pooler.sql",
			Data:          PoolerSQL,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-pooler.sql:ro",
		},
	}
}
