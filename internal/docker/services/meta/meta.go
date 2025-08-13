package meta

import (
	"context"
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/docker/services/database"
	"github.com/train360-corp/projconf/internal/docker/types"
	"io"
	"os/exec"
)

type Service struct{}

const ContainerName = "projconf-internal-supabase-meta"

func (s Service) GetDisplay() string {
	return "Meta"
}

func (s Service) Run(evn *types.SharedEvn) error {
	args := []string{
		"run",
		"--rm",
		"--name", ContainerName,
		"--label", "com.docker.compose.project=projconf",
		"--label", "com.docker.compose.service=meta",
		"--label", "com.docker.compose.version=2.0",
		"--network", "projconf-net",
		"--network-alias", "meta",
		"-e", "PG_META_PORT=8080",
		"-e", fmt.Sprintf("PG_META_DB_HOST=%s", database.ContainerName),
		"-e", "PG_META_DB_PORT=5432",
		"-e", "PG_META_DB_NAME=postgres",
		"-e", "PG_META_DB_USER=supabase_admin",
		"-e", fmt.Sprintf("PG_META_DB_PASSWORD=%s", evn.PGPASSWORD),
		"supabase/postgres-meta:v0.91.0",
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf("failed to run docker command: %s", err))
	}

	return nil
}

func (s Service) GetWriteables() []types.Writeable {
	return []types.Writeable{}
}

func (s Service) WaitFor(ctx context.Context) error {
	return nil
}
