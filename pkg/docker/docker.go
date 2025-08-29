/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package docker

import (
	"context"
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/fs"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

func RunService(ctx context.Context, service Service, env Env) error {
	log.Printf("starting Docker service %q...\n", service.Display())

	// ensure any temp files exist
	for _, file := range service.TempFiles() {
		if _, e := os.Stat(file.LocalPath); errors.Is(e, os.ErrNotExist) {
			if err := fs.WriteDependencies(file.LocalPath, file.Data, file.Permissions); err != nil {
				return fmt.Errorf("write temp file failed (path=%s): %w", file.LocalPath, err)
			}
		}
	}

	// build docker args
	args := append([]string{"run", "--rm"}, service.Args(env)...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	// platform-specific: process group / job object, etc.
	applyPlatformProcAttrs(cmd)

	// start and wait
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// try graceful shutdown (platform-specific)
		_ = terminateProcessTree(cmd)
		_ = exec.Command("docker", "stop", "--time", "10", service.ContainerName()).Run()

		// give it a moment, then force-kill
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()
		select {
		case <-done:
			return ctx.Err() // propagate cancellation
		case <-timer.C:
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			<-done
			return ctx.Err()
		}
	}
}
