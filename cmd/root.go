/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cmd

import (
	"github.com/train360-corp/projconf/internal/app"
	"os"
)

func Run() error {
	return app.Get().Run(os.Args)
}
