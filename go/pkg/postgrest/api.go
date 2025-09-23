/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

// convert the schema from OpenAPI2/Swagger => OpenAPI3
//go:generate go run ./tools -in http://127.0.0.1:54323/api/platform/projects/default/api/rest -out ./openapi.yaml -format yaml

// generate the types and client
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config=./config.yaml ./openapi.yaml

package postgrest

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/train360-corp/projconf/go/pkg/consts"
	"net/http"
	"strings"
)

func GetUnauthenticatedClient(baseURL string, apiKey string) (*ClientWithResponses, error) {
	return GetAuthenticatedClient(baseURL, apiKey, nil)
}

func GetAuthenticatedClient(baseURL string, apiKey string, c *gin.Context) (*ClientWithResponses, error) {
	return NewClientWithResponses(strings.TrimSuffix(baseURL, "/"), WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		req.Header.Add("apikey", apiKey)
		if c != nil {
			if c.Request.Header.Get(consts.X_ADMIN_API_KEY) != "" {
				req.Header.Add(consts.X_ADMIN_API_KEY, c.Request.Header.Get(consts.X_ADMIN_API_KEY))
			} else {
				req.Header.Add(consts.X_CLIENT_SECRET_ID, c.Request.Header.Get(consts.X_CLIENT_SECRET_ID))
				req.Header.Add(consts.X_CLIENT_SECRET, c.Request.Header.Get(consts.X_CLIENT_SECRET))
			}
		}
		return nil
	}))
}
