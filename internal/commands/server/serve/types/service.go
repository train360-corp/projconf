/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package types

import "github.com/docker/docker/api/types/mount"

type Service struct {
	Image  string
	Name   string
	Cmd    []string
	Env    []string
	Labels map[string]string
	Mounts []mount.Mount
	Ports  []uint16
}
