package postgrest

import (
	"context"
	"fmt"
	"github.com/train360-corp/projconf/internal/docker/services/database"
	"github.com/train360-corp/projconf/internal/docker/types"
)

type Service struct{}

const ContainerName = "projconf-internal-supabase-postgrest"

func (s Service) GetDisplay() string {
	return "Postgrest"
}

func (s Service) GetArgs(evn *types.SharedEnv) []string {
	return []string{
		"--name", ContainerName,
		"--label", "com.docker.compose.project=projconf",
		"--label", "com.docker.compose.service=postgrest",
		"--label", "com.docker.compose.version=2.0",
		"--network", "projconf-net",
		"--network-alias", "rest",
		"-e", fmt.Sprintf("PGRST_DB_URI=postgres://authenticator:%s@%s:5432/postgres", evn.PGPASSWORD, database.ContainerName),
		"-e", "PGRST_DB_SCHEMAS=public",
		"-e", "PGRST_DB_ANON_ROLE=anon",
		"-e", fmt.Sprintf("PGRST_JWT_SECRET=%s", evn.JWT_SECRET),
		"-e", "PGRST_DB_USE_LEGACY_GUCS=false",
		"-e", fmt.Sprintf("PGRST_APP_SETTINGS_JWT_SECRET=%s", evn.JWT_SECRET),
		"-e", "PGRST_APP_SETTINGS_JWT_EXP=3600",
		"postgrest/postgrest:v12.2.12",
		"postgrest",
	}
}

func (s Service) GetWriteables() []types.Writeable {
	return []types.Writeable{}
}

func (s Service) WaitFor(ctx context.Context) error {
	return nil
}
