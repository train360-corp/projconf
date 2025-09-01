/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package api

import (
	"context"
	"github.com/train360-corp/projconf/pkg/consts"
	"net/http"
	"strings"
)

func GetAPIClient(hostname string, adminApiKey string, clientSecretId string, clientSecret string) (*ClientWithResponses, error) {
	return NewClientWithResponses(strings.TrimSuffix(hostname, "/"), WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		if adminApiKey != "" {
			req.Header.Add(consts.X_ADMIN_API_KEY, adminApiKey)
		} else {
			req.Header.Add(consts.X_CLIENT_SECRET_ID, clientSecretId)
			req.Header.Add(consts.X_CLIENT_SECRET, clientSecret)
		}
		return nil
	}))
}
