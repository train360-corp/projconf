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

package cli

import (
	"context"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/consts"
	"github.com/train360-corp/projconf/internal/server/api"
	"net/http"
	"strings"
)

func isAdminAPIClient(flags *config.Flags) bool {
	return flags != nil && strings.Trim(flags.AdminAPIKey, " ") != ""
}

func getAPIClient(cfg *config.DiskConfig, flags *config.Flags) (*api.ClientWithResponses, error) {
	return api.NewClientWithResponses(cfg.Account.Url, api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		if isAdminAPIClient(flags) {
			req.Header.Add(consts.X_ADMIN_API_KEY, flags.AdminAPIKey)
		} else {
			req.Header.Add(consts.X_CLIENT_SECRET_ID, cfg.Account.Client.Id)
			req.Header.Add(consts.X_CLIENT_SECRET, cfg.Account.Client.Secret)
		}
		return nil
	}))
}
