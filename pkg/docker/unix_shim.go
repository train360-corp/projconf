//go:build unix

/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package docker

import (
	"os/exec"
	"syscall"
)

func applyPlatformProcAttrs(cmd *exec.Cmd) {
	// Put child in its own process group so we can signal the whole group.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminateProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	// Negative PID targets the process group.
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}
