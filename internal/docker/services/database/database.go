package database

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/train360-corp/projconf/internal/docker/types"
	"github.com/train360-corp/projconf/internal/fs"
	"io"
	"os/exec"
	"path/filepath"
	"time"
)

//go:embed embeds/realtime.sql
var RealtimeSQL []byte

//go:embed embeds/webhooks.sql
var WebhooksSQL []byte

//go:embed embeds/roles.sql
var RolesSQL []byte

//go:embed embeds/jwt.sql
var JwtSQL []byte

//go:embed embeds/_supabase.sql
var SupabaseSQL []byte

//go:embed embeds/logs.sql
var LogsSQL []byte

//go:embed embeds/pooler.sql
var PoolerSQL []byte

type Service struct {
}

const ContainerName = "projconf-internal-supabase-db"

func (srvc Service) WaitFor(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errors.New(fmt.Sprintf("timeout waiting for database"))
		default:
		}

		cmd := exec.CommandContext(ctx, "pg_isready", "-h", "127.0.0.1", "-U", "postgres")
		cmd.Stderr = io.Discard
		cmd.Stdout = io.Discard
		cmd.Stdin = nil
		if err := cmd.Run(); err == nil {
			return nil
		}

		// Retry after 5 seconds
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for Postgres: %w", ctx.Err())
		}
	}
}

func (srvc Service) GetDisplay() string {
	return "Database (Postgres)"
}

func (srvc Service) GetArgs(sharedEvn *types.SharedEnv) []string {

	userRoot, _ := fs.GetUserRoot()

	args := []string{
		"--name", ContainerName,
		"--label", "com.docker.compose.project=projconf",
		"--label", "com.docker.compose.service=db",
		"--label", "com.docker.compose.version=2.0",
		"--network", "projconf-net",
		"--network-alias", "db",
		"-e", "POSTGRES_HOST=/var/run/postgresql",
		"-e", "PGPORT=5432",
		"-e", "POSTGRES_PORT=5432",
		"-e", "PGPASSWORD=" + sharedEvn.PGPASSWORD,
		"-e", "POSTGRES_PASSWORD=" + sharedEvn.PGPASSWORD,
		"-e", "PGDATABASE=postgres",
		"-e", "POSTGRES_DB=postgres",
		"-e", "JWT_SECRET=" + sharedEvn.JWT_SECRET,
		"-e", "JWT_EXP=3600",
		"-v", "db-config:/etc/postgresql-custom",
		"-p", "127.0.0.1:5432:5432",
	}

	for _, writeable := range srvc.GetWriteables() {
		args = append(args, "-v", fmt.Sprintf("%s:%s", writeable.LocalPath, writeable.ContainerPath))
	}

	dataDir := filepath.Join(userRoot, "db", "data")
	args = append(args, "-v", fmt.Sprintf("%s:/var/lib/postgresql/data", dataDir))

	args = append(args,
		"supabase/postgres:17.4.1.055",
		"postgres",
		"-c", "config_file=/etc/postgresql/postgresql.conf",
		"-c", "log_min_messages=fatal",
		"-c", "wal_level=minimal",
		"-c", "max_wal_senders=0",
	)

	return args
}

func (srvc Service) GetWriteables() []types.Writeable {

	usrRoot, err := fs.EnsureUserRoot()
	if err != nil {
		panic(errors.New(fmt.Sprintf("failed to get user root: %s", err)))
	}

	return []types.Writeable{
		{
			LocalPath:     filepath.Join(usrRoot, "db", "realtime.sql"),
			Data:          RealtimeSQL,
			Perm:          0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-realtime.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "webhooks.sql"),
			Data:          WebhooksSQL,
			Perm:          0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "roles.sql"),
			Data:          RolesSQL,
			Perm:          0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/99-roles.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "jwt.sql"),
			Data:          JwtSQL,
			Perm:          0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "_supabase.sql"),
			Data:          SupabaseSQL,
			Perm:          0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/97-_supabase.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "logs.sql"),
			Data:          LogsSQL,
			Perm:          0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-logs.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "pooler.sql"),
			Data:          PoolerSQL,
			Perm:          0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-pooler.sql:ro",
		},
	}
}
