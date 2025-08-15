package docker

import (
	"context"
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/docker/services/database"
	"github.com/train360-corp/projconf/internal/docker/services/kong"
	"github.com/train360-corp/projconf/internal/docker/services/postgrest"
	"github.com/train360-corp/projconf/internal/docker/types"
	"github.com/train360-corp/projconf/internal/docker/utils"
	"io"
	"log"
	"os/exec"
	"syscall"
)

func GetServices() []types.Service {
	return []types.Service{
		database.Service{},
		postgrest.Service{},
		//meta.Service{}, // NRB 2025.08.12: not required atm
		kong.Service{},
	}
}

func runAttachedDetachable(ctx context.Context, args []string) error {
	args = append([]string{"run", "--rm"}, args...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		_ = cmd.Process.Signal(syscall.SIGINT)
		_ = cmd.Process.Signal(syscall.SIGTERM)
		<-done
		return ctx.Err()
	}
}

func RunService(ctx context.Context, service types.Service, env types.SharedEnv) error {
	log.Printf("starting Docker service \"%s\"...\n", service.GetDisplay())
	if err := utils.WriteTempFiles(service.GetWriteables()); err != nil {
		return errors.New(fmt.Sprintf("failed to write temp files for service \"%s\": %s", service.GetDisplay(), err.Error()))
	}

	err := runAttachedDetachable(ctx, service.GetArgs(&env))

	if err != nil {
		return errors.New(fmt.Sprintf("failed to run service \"%s\": %s", service.GetDisplay(), err.Error()))
	}
	log.Printf("Docker service \"%s\" exited\n", service.GetDisplay())
	return nil
}
