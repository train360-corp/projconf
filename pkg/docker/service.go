/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package docker

import (
	"context"
	"os"
)

type ServiceTempFile struct {
	LocalPath     string
	ContainerPath string
	Data          []byte
	Permissions   os.FileMode
}

type Service interface {
	Display() string
	ContainerName() string
	TempFiles() []ServiceTempFile
	Args(Env) []string
	HealthCheck(context.Context) (bool, int)
}
