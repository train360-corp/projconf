/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package fs

import (
	"os"
	"path/filepath"
)

func WriteDependencies(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	} else if err := os.WriteFile(path, data, perm); err != nil {
		return err
	} else {
		return nil
	}
}
