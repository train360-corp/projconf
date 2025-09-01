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
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

const (
	networkName string = "projconf-net"
)

var (
	NetworkID string
)

func InitNetwork(ctx context.Context) error {

	MustLogger()
	MustCli()

	if NetworkID != "" { // network already initialized
		Logger.Warn("InitNetwork called with existing network (returning)")
		return nil
	}

	nets, err := Cli.NetworkList(ctx, network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	if err != nil { // cannot check networks
		return err
	} else if len(nets) == 1 { // network exists
		NetworkID = nets[0].ID
		Logger.Debug("found existing network (returning)")
		return nil
	} else if len(nets) > 1 { // catch-all
		return fmt.Errorf("too many networks: %v", len(nets))
	}

	trueP := true
	falseP := false
	net, err := Cli.NetworkCreate(ctx, networkName, network.CreateOptions{
		Driver:     "bridge",
		Scope:      "local",
		EnableIPv4: &trueP,
		EnableIPv6: &falseP,
		Attachable: true,
	})
	if err != nil {
		return err
	}
	NetworkID = net.ID
	Logger.Debug("created network")

	return nil
}

func MustNetwork() {
	if NetworkID == "" {
		panic("network not initialized")
	}
}
