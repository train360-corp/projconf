/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package services

import (
	"github.com/train360-corp/projconf/pkg/docker"
	"github.com/train360-corp/projconf/pkg/docker/services/postgres"
	"github.com/train360-corp/projconf/pkg/docker/services/postgrest"
)

func GetServices() []docker.Service {
	return []docker.Service{
		postgres.Service{},
		postgrest.Service{},
	}
}
