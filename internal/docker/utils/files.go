/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package utils

import (
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/docker/types"
	"github.com/train360-corp/projconf/internal/fs"
	"os"
)

func WriteTempFiles(writeables []types.Writeable) error {
	for _, writeable := range writeables {
		//log.Printf("wrote temp file: %s\n", writeable.Path)
		if _, err := os.Stat(writeable.LocalPath); errors.Is(err, os.ErrNotExist) {
			if err := fs.WriteDependencies(writeable.LocalPath, writeable.Data, writeable.Perm); err != nil {
				return errors.New(fmt.Sprintf("write temp file failed (path=%s): %s", writeable.LocalPath, err))
			}
		}
	}
	return nil
}
