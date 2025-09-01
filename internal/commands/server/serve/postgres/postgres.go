/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package postgres

import (
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/train360-corp/projconf/internal/commands/server/serve/types"
	"os"
	"path/filepath"
)

type ServiceRequest struct {
	DbDataRoot          string
	TmpFileRoot         string
	ProjConfAdminApiKey string
	PostgresPassword    string
	JwtSecret           string
}

const ContainerName = "projconf-internal-supabase-db"

func Service(req ServiceRequest) (*types.Service, error) {

	files := getTempFiles()
	var mounts []mount.Mount
	for _, file := range files {
		localPath := filepath.Join(req.TmpFileRoot, file.Name)
		if err := os.WriteFile(localPath, file.Data, 0o444); err != nil {
			return nil, fmt.Errorf("failed to write temp-file %s: %v", file.Name, err)
		}
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   localPath,
			Target:   file.ContainerPath,
			ReadOnly: true,
		})
	}

	mounts = append(mounts,
		mount.Mount{
			Type:   mount.TypeVolume,
			Source: "db-config",
			Target: "/etc/postgresql-custom",
		},
		mount.Mount{
			Type:   mount.TypeBind,
			Source: filepath.Join(req.DbDataRoot, "data"),
			Target: "/var/lib/postgresql/data",
		},
	)

	return &types.Service{
		Image:  "supabase/postgres:17.4.1.055",
		Name:   ContainerName,
		Mounts: mounts,
		Ports:  make([]uint16, 0),
		Labels: map[string]string{
			"projconf.service": "postgres",
		},
		Cmd: []string{
			"postgres",
			"-c", "config_file=/etc/postgresql/postgresql.conf",
			"-c", "log_min_messages=error",
			"-c", "wal_level=minimal",
			"-c", "max_wal_senders=0",
			"-c", fmt.Sprintf("projconf.x_admin_api_key=%s", req.ProjConfAdminApiKey),
		},
		Env: []string{
			"POSTGRES_HOST=/var/run/postgresql",
			"PGPORT=5432",
			"POSTGRES_PORT=5432",
			fmt.Sprintf("PGPASSWORD=%s", req.PostgresPassword),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", req.PostgresPassword),
			"PGDATABASE=postgres",
			"POSTGRES_DB=postgres",
			fmt.Sprintf("JWT_SECRET=%s", req.JwtSecret),
			"JWT_EXP=3600",
		},
	}, nil
}
