/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package fs

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// FileExists checks whether a file exists and is accessible by the runtime
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // file exists
	}
	if os.IsNotExist(err) {
		return false // definitely does not exist
	}
	// Something else went wrong (e.g., permission denied)
	return false
}

// GetUserRoot returns a per-user, no-sudo root directory for the app.
//
//	Linux/*BSD:  $XDG_CONFIG_HOME/projconf (or ~/.config/projconf)
//	macOS:       ~/Library/Application Support/projconf
//	Windows:     %APPDATA%\projconf
func GetUserRoot() (string, error) {
	switch runtime.GOOS {
	case "linux", "freebsd", "openbsd", "netbsd":
		// Prefer XDG config dir for a consistent, user-writable root.
		cfg, err := os.UserConfigDir() // usually ~/.config
		if err != nil || cfg == "" {
			home, herr := os.UserHomeDir()
			if herr != nil || home == "" {
				return "", errors.New("cannot resolve user config/home directory")
			}
			cfg = filepath.Join(home, ".config")
		}
		return filepath.Join(cfg, "projconf"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return "", errors.New("cannot resolve user home directory")
		}
		return filepath.Join(home, "Library", "Application Support", "projconf"), nil
	case "windows":
		// For config-like roots, APPDATA (Roaming) is standard.
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "projconf"), nil
		}
		return "", errors.New("APPDATA not set")
	default:
		// Fallback: hidden dir in home
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return "", errors.New("cannot resolve user home directory")
		}
		return filepath.Join(home, ".projconf"), nil
	}
}

// EnsureUserRoot creates the user root directory if it doesn't exist.
// Uses 0755 on Unix; Windows ignores mode bits but will create the directory.
func EnsureUserRoot() (string, error) {
	if root, err := GetUserRoot(); err != nil {
		return "", err
	} else if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	} else {
		return root, nil
	}
}

func GetTempRoot() (string, error) {
	tmpRoot, err := os.MkdirTemp("", "supabase-db-*")
	if err != nil {
		return "", err
	}
	return tmpRoot, nil
}
