/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package types

import (
	"context"
	"os"
)

type SharedEnv struct {
	PGPASSWORD             string
	JWT_SECRET             string
	ANON_KEY               string
	SERVICE_KEY            string
	PROJCONF_ADMIN_API_KEY string
}

type Writeable struct {
	LocalPath     string
	Data          []byte
	Perm          os.FileMode
	ContainerPath string
}

type Service interface {
	GetDisplay() string
	GetArgs(env *SharedEnv) []string
	GetWriteables() []Writeable
	WaitFor(ctx context.Context) error
}
