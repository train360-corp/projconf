/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package postgrest

import (
	"context"
	"fmt"
	"github.com/train360-corp/projconf/pkg/docker"
	"github.com/train360-corp/projconf/pkg/docker/services/database"
)

type Service struct{}

const ContainerName = "projconf-internal-supabase-postgrest"

func (s Service) ContainerName() string {
	return ContainerName
}

func (s Service) Display() string {
	return "Postgrest"
}

func (s Service) Args(env docker.Env) []string {
	return []string{
		"--name", ContainerName,
		"--label", "com.docker.compose.project=projconf",
		"--label", "com.docker.compose.service=postgrest",
		"--label", "com.docker.compose.version=2.0",
		"--network", "projconf-net",
		"--network-alias", "rest",
		"-e", fmt.Sprintf("PGRST_DB_URI=postgres://authenticator:%s@%s:5432/postgres", env.PGPASSWORD, database.ContainerName),
		"-e", "PGRST_DB_SCHEMAS=public",
		"-e", "PGRST_DB_ANON_ROLE=anon",
		"-e", fmt.Sprintf("PGRST_JWT_SECRET=%s", env.JWT_SECRET),
		"-e", "PGRST_DB_USE_LEGACY_GUCS=false",
		"-e", fmt.Sprintf("PGRST_APP_SETTINGS_JWT_SECRET=%s", env.JWT_SECRET),
		"-e", "PGRST_APP_SETTINGS_JWT_EXP=3600",
		"-e", "PGRST_ADMIN_SERVER_PORT=3001",
		"postgrest/postgrest:v12.2.12",
		"postgrest",
	}
}

func (s Service) HealthCheck(ctx context.Context) (bool, int) {
	return true, 0
}

func (s Service) TempFiles() []docker.ServiceTempFile {
	return []docker.ServiceTempFile{}
}
