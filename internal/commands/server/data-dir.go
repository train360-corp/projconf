/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// getSystemProjConfDir returns the system-wide config directory for ProjConf:
//   - linux:   /etc/projconf
//   - darwin:  /Library/Application Support/ProjConf
//   - windows: %ProgramData%\ProjConf  (fallback: C:\ProgramData\ProjConf)
func getSystemProjConfDir() (string, error) {
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(string(os.PathSeparator), "etc", "projconf"), nil
	case "darwin":
		// System-wide, all-users app support on macOS
		return filepath.Join(string(os.PathSeparator), "Users", "Shared", "ProjConf"), nil
	case "windows":
		programData := os.Getenv("PROGRAMDATA")
		if programData == "" {
			// Sensible fallback if the env var isn't set
			programData = `C:\ProgramData`
		}
		return filepath.Join(programData, "ProjConf"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// EnsureSystemProjConfDir creates the directory (and parents) if it doesn't exist.
// Returns the path (same as getSystemProjConfDir).
// On Unix, 0755 is a safe default for a shared system dir (adjust tighter if storing secrets).
func EnsureSystemProjConfDir() (string, error) {
	dir, err := getSystemProjConfDir()
	if err != nil {
		return "", err
	}

	perm := os.FileMode(0o755) // On Windows the mode is largely ignored; still fine to pass.
	if err := os.MkdirAll(dir, perm); err != nil {
		return "", fmt.Errorf("failed to create root directory %s: %v", dir, err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "data"), perm); err != nil {
		return "", fmt.Errorf("failed to create data directory %s: %v", dir, err)
	}

	return dir, nil
}
