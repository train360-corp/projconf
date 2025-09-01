/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package serve

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
)

var cli *client.Client

func initCli(ctx context.Context) {
	c, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		Logger.Fatal(fmt.Sprintf("failed to create docker client: %v", err))
	} else {
		cli = c
	}
	Logger.Debug(fmt.Sprintf("Docker Host: %v", cli.DaemonHost()))

	// Fast, cheap liveness probe
	ping, err := cli.Ping(ctx)
	if err != nil {
		Logger.Fatal(fmt.Sprintf("docker daemon not reachable: %v", err))
	} else {
		if ping.APIVersion != "" {
			Logger.Debug(fmt.Sprintf("Docker API version: %v", ping.APIVersion))
		}
		if ping.OSType != "" {
			Logger.Debug(fmt.Sprintf("Docker OS: %v", ping.OSType))
		}
		if ping.BuilderVersion != "" {
			Logger.Debug(fmt.Sprintf("Docker Builder: %v", ping.BuilderVersion))
		}
		Logger.Debug(fmt.Sprintf("Docker Experimental: %v", ping.Experimental))
	}
}

func mustCli() {
	if cli == nil {
		panic("cli not initialized")
	}
}
