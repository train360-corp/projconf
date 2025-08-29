/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package database

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/train360-corp/projconf/internal/fs"
	"github.com/train360-corp/projconf/pkg/docker"
	"os/exec"
	"path/filepath"
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

type Service struct {
}

const ContainerName = "projconf-internal-supabase-db"

func (s Service) ContainerName() string {
	return ContainerName
}

func (s Service) Display() string {
	return "Postgres"
}

func (s Service) Args(env docker.Env) []string {

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
		"-e", "PGPASSWORD=" + env.PGPASSWORD,
		"-e", "POSTGRES_PASSWORD=" + env.PGPASSWORD,
		"-e", "PGDATABASE=postgres",
		"-e", "POSTGRES_DB=postgres",
		"-e", "JWT_SECRET=" + env.JWT_SECRET,
		"-e", "JWT_EXP=3600",
		"-v", "db-config:/etc/postgresql-custom",
		"-p", "127.0.0.1:5432:5432",
	}

	for _, writeable := range s.TempFiles() {
		args = append(args, "-v", fmt.Sprintf("%s:%s", writeable.LocalPath, writeable.ContainerPath))
	}

	dataDir := filepath.Join(userRoot, "db", "data")
	args = append(args, "-v", fmt.Sprintf("%s:/var/lib/postgresql/data", dataDir))

	args = append(args,
		"supabase/postgres:17.4.1.055",
		"postgres",
		"-c", "config_file=/etc/postgresql/postgresql.conf",
		"-c", "log_min_messages=error",
		"-c", "wal_level=minimal",
		"-c", "max_wal_senders=0",
		"-c", fmt.Sprintf("projconf.x_admin_api_key=%s", env.PROJCONF_ADMIN_API_KEY),
	)

	return args
}

func (s Service) HealthCheck(ctx context.Context) (bool, int) {
	cmd := exec.CommandContext(ctx, "docker", "exec", ContainerName, "pg_isready")
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else { // hard failure (docker not found, ctx canceled before start, etc.)
			exitCode = -1
		}
	}
	return exitCode == 0, exitCode
}

func (s Service) TempFiles() []docker.ServiceTempFile {

	usrRoot, err := fs.EnsureUserRoot()
	if err != nil {
		panic(errors.New(fmt.Sprintf("failed to get user root: %s", err)))
	}

	return []docker.ServiceTempFile{
		{
			LocalPath:     filepath.Join(usrRoot, "db", "realtime.sql"),
			Data:          RealtimeSQL,
			Permissions:   0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-realtime.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "webhooks.sql"),
			Data:          WebhooksSQL,
			Permissions:   0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/98-webhooks.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "roles.sql"),
			Data:          RolesSQL,
			Permissions:   0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/99-roles.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "jwt.sql"),
			Data:          JwtSQL,
			Permissions:   0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/init-scripts/99-jwt.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "_supabase.sql"),
			Data:          SupabaseSQL,
			Permissions:   0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/97-_supabase.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "logs.sql"),
			Data:          LogsSQL,
			Permissions:   0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-logs.sql:ro",
		},
		{
			LocalPath:     filepath.Join(usrRoot, "db", "pooler.sql"),
			Data:          PoolerSQL,
			Permissions:   0o444,
			ContainerPath: "/docker-entrypoint-initdb.d/migrations/99-pooler.sql:ro",
		},
	}
}
