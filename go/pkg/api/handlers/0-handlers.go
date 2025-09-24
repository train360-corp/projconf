/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/train360-corp/projconf/go/pkg/api"
	"github.com/train360-corp/projconf/go/pkg/postgrest"
)

type RouteHandlers struct {
	BaseURL string
	ApiKey  string
}

func GetRouteHandlers(baseUrl string, apiKey string) api.ServerInterface {
	return &RouteHandlers{
		BaseURL: baseUrl,
		ApiKey:  apiKey,
	}
}

func (r RouteHandlers) postgrest(c *gin.Context) (*postgrest.ClientWithResponses, *api.InternalServerError) {
	client, err := postgrest.GetAuthenticatedClient(r.BaseURL, r.ApiKey, c)
	if err != nil {
		return nil, &api.InternalServerError{
			Error:       "unable to create postgrest client",
			Description: err.Error(),
		}
	}
	return client, nil
}
