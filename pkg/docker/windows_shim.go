//go:build windows

/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package docker

import "os/exec"

func applyPlatformProcAttrs(cmd *exec.Cmd) {
	// No-op on Windows here (you could use Job Objects if you need tree control).
}

func terminateProcessTree(cmd *exec.Cmd) error {
	// We can only kill the docker client process; it will tear down the attached container.
	// If you need stricter control, run the container detached and `docker stop` it instead.
	if cmd != nil && cmd.Process != nil {
		return cmd.Process.Kill() // hard kill (no SIGTERM semantics on Windows)
	}
	return nil
}
