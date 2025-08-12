package supabase

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/dependencies/files/types"
	"github.com/train360-corp/projconf/internal/fs"
	"path/filepath"
)

const (
	REALTIME  = "/docker-entrypoint-initdb.d/migrations/99-realtime.sql"
	WEBHOOKS  = "/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql"
	ROLES     = "/docker-entrypoint-initdb.d/init-scripts/99-roles.sql"
	JWT       = "/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql"
	_SUPABASE = "/docker-entrypoint-initdb.d/migrations/97-_supabase.sql"
	LOGS      = "/docker-entrypoint-initdb.d/migrations/99-logs.sql"
	POOLER    = "/docker-entrypoint-initdb.d/migrations/99-pooler.sql"
)

//go:embed database/realtime.sql
var RealtimeSQL []byte

//go:embed database/webhooks.sql
var WebhooksSQL []byte

//go:embed database/roles.sql
var RolesSQL []byte

//go:embed database/jwt.sql
var JwtSQL []byte

//go:embed database/_supabase.sql
var SupabaseSQL []byte

//go:embed database/logs.sql
var LogsSQL []byte

//go:embed database/pooler.sql
var PoolerSQL []byte

type DatabaseEmbed struct{}

func (e DatabaseEmbed) GetDisplay() string {
	return "Database"
}

func (e DatabaseEmbed) GetWriteables() []types.Writeable {

	tmpRoot, err := fs.GetTempRoot()
	if err != nil {
		panic(errors.New(fmt.Sprintf("failed to get temp root: %s", err)))
	}

	return []types.Writeable{
		{
			Path: filepath.Join(tmpRoot, "db", "realtime.sql"),
			Data: RealtimeSQL,
			Perm: 0o444,
		},
		{
			Path: filepath.Join(tmpRoot, "db", "webhooks.sql"),
			Data: WebhooksSQL,
			Perm: 0o444,
		},
		{
			Path: filepath.Join(tmpRoot, "db", "roles.sql"),
			Data: RolesSQL,
			Perm: 0o444,
		},
		{
			Path: filepath.Join(tmpRoot, "db", "jwt.sql"),
			Data: JwtSQL,
			Perm: 0o444,
		},
		{
			Path: filepath.Join(tmpRoot, "db", "_supabase.sql"),
			Data: SupabaseSQL,
			Perm: 0o444,
		},
		{
			Path: filepath.Join(tmpRoot, "db", "logs.sql"),
			Data: LogsSQL,
			Perm: 0o444,
		},
		{
			Path: filepath.Join(tmpRoot, "db", "pooler.sql"),
			Data: PoolerSQL,
			Perm: 0o444,
		},
	}
}
