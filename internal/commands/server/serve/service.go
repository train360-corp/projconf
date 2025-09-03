/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package serve

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/train360-corp/projconf/internal/commands/server/serve/types"
	"io"
	"strconv"
	"time"
)

func RunService(ctx context.Context, svc *types.Service, onExit func()) (string, func() error, error) {

	MustLogger()
	MustCli()
	MustNetwork()

	Logger.Debug(fmt.Sprintf("pulling image for %s (%s)", svc.Name, svc.Image))
	pull, err := Cli.ImagePull(ctx, svc.Image, image.PullOptions{})
	if err != nil {
		return "", nil, fmt.Errorf("failed to pull image for %s (%s): %v", svc.Name, svc.Image, err)
	}
	if pull != nil {
		_, _ = io.Copy(io.Discard, pull)
		_ = pull.Close()
	}
	Logger.Debug(fmt.Sprintf("pulled image for %s (%s)", svc.Name, svc.Image))

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, p := range svc.Ports {
		port := nat.Port(fmt.Sprintf("%d/tcp", p))
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: strconv.Itoa(int(p)),
			},
		}
	}

	// add projconf label
	svc.Labels["com.docker.compose.project"] = "projconf"

	// create container
	resp, err := Cli.ContainerCreate(ctx,
		&container.Config{
			Image:        svc.Image,
			Cmd:          svc.Cmd,
			Env:          svc.Env,
			OpenStdin:    true,  // keep a stdin pipe open from our process
			StdinOnce:    true,  // when our stdin attach disconnects, close container's STDIN
			Tty:          false, // keep streams multiplexed for stdcopy
			ExposedPorts: exposedPorts,
			Labels:       svc.Labels,
		},
		&container.HostConfig{
			AutoRemove:    true, // like --rm
			RestartPolicy: container.RestartPolicy{Name: "no"},
			NetworkMode:   container.NetworkMode(networkName),
			Mounts:        svc.Mounts,
			PortBindings:  portBindings,
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				networkName: {
					NetworkID: NetworkID,
					Aliases:   []string{}, // e.g., "db"
				},
			},
		},
		nil,
		svc.Name,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create docker container: %v", err)
	}
	Logger.Debug(fmt.Sprintf("created docker container (%s)", PreviewString(resp.ID)))

	// start container
	if err := Cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", nil, fmt.Errorf("failed to start docker container: %v", err)
	}
	Logger.Debug(fmt.Sprintf("started docker container (%s)", PreviewString(resp.ID)))

	// listen to status
	statusCh, errCh := Cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	go func() {
		select {
		case err := <-errCh:
			if err != nil {
				// the wait itself failed (e.g., context canceled)
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					Logger.Debug("wait canceled (shutdown)")
				} else {
					Logger.Error(fmt.Sprintf("docker container wait error: %v", err))
				}
			}
		case st := <-statusCh:
			// container exited (possibly immediately)
			Logger.Warn(fmt.Sprintf("docker container (%s) exited with status %d", PreviewString(resp.ID), st.StatusCode))
			onExit()
		}
	}()

	// attach (stdin); keep this connection open while our app is alive.
	att, err := Cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stdin:  true,
		Stream: true,
		Stdout: false,
		Stderr: false,
	})
	if err != nil {
		Logger.Error(fmt.Sprintf("attach failed, attempting to stop docker container"))
		timeoutSeconds := 5
		// use background context in case ctx is already stopped
		if err := Cli.ContainerStop(context.Background(), resp.ID, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
			Logger.Error(fmt.Sprintf("failed to stop docker container: %v", err))
		} else {
			Logger.Info(fmt.Sprintf("successfully stopped docker container"))
		}
		return "", nil, fmt.Errorf("failed to attach docker container: %v", err)
	}

	stop := func() error {

		// close stdin (causes EOF in the container)
		err := att.CloseWrite()
		if err != nil {
			return fmt.Errorf("failed to close container stdin: %v", err)
		}
		att.Close()

		// attempt to manually stop, too (use background context in case ctx is already stopped)
		timeoutSeconds := 5
		if err := Cli.ContainerStop(context.Background(), resp.ID, container.StopOptions{Timeout: &timeoutSeconds}); err != nil {
			return fmt.Errorf("failed to stop docker container: %v", err)
		}

		// give it a brief moment (AutoRemove will delete it)
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	return resp.ID, stop, nil
}

// ExecInContainer runs "cmd" inside container cid and streams output to stdout/stderr
func ExecInContainer(ctx context.Context, cid string, cmd []string) (string, error) {
	// create ExecInContainer instance
	execResp, err := Cli.ContainerExecCreate(ctx, cid, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	})
	if err != nil {
		return "", fmt.Errorf("ExecInContainer create failed: %w", err)
	}

	// attach
	att, err := Cli.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{Tty: false})
	if err != nil {
		return "", fmt.Errorf("ExecInContainer attach failed: %w", err)
	}
	defer att.Close()

	// copy output to local stdout/stderr
	var buf bytes.Buffer
	_, _ = stdcopy.StdCopy(&buf, &buf, att.Reader)
	output := buf.String()

	// check exit code
	inspect, err := Cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return output, fmt.Errorf("ExecInContainer inspect failed: %w", err)
	}
	if inspect.ExitCode != 0 {
		return output, fmt.Errorf("ExecInContainer command exited with code %d", inspect.ExitCode)
	}

	return output, nil
}
