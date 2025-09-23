/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	"context"
	"crypto/subtle"
	"github.com/gin-gonic/gin"
	"github.com/train360-corp/projconf/go/pkg/api"
	"github.com/train360-corp/projconf/go/pkg/consts"
	"github.com/train360-corp/projconf/go/pkg/postgrest"
	"github.com/train360-corp/supago"
	"net/http"
	"strings"
)

func authHandler(config *supago.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		// status endpoints are public
		if strings.HasPrefix(c.Request.URL.Path, "/v1/status") {
			c.Next()
		}

		// ----- ADMIN FLOW -----
		if c.GetHeader(consts.X_ADMIN_API_KEY) != "" {
			raw := strings.TrimSpace(c.GetHeader(consts.X_ADMIN_API_KEY))
			if raw == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
					Error:       "Unauthorized",
					Description: "missing 'x-admin-api-key' header",
				})
			} else {
				token := raw

				// accept case-insensitive Bearer and trim
				if len(raw) >= 7 && strings.EqualFold(raw[:7], "Bearer ") {
					token = strings.TrimSpace(raw[7:])
				}

				// constant-time compare (same length)
				if len(token) != len(AdminApiKey) ||
					subtle.ConstantTimeCompare([]byte(token), []byte(AdminApiKey)) != 1 {
					c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
						Error:       "Unauthorized",
						Description: "invalid 'x-admin-api-key' header",
					})
				} else {
					c.Next()
				}
			}
		} else {

			// ----- CLIENT FLOW -----
			id := c.GetHeader(consts.X_CLIENT_SECRET_ID)
			sec := c.GetHeader(consts.X_CLIENT_SECRET)
			if id == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
					Error:       "Unauthorized",
					Description: "missing 'x-client-secret-id' header",
				})
			} else if sec == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
					Error:       "Unauthorized",
					Description: "missing 'x-client-secret' header",
				})
			} else if supabase, err := postgrest.GetAuthenticatedClient("http://127.0.0.1:8000/rest/v1/", config.Keys.PublicJwt, c); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, api.Error{
					Error:       "client error",
					Description: "unable to create a client",
				})
			} else if clients, err := supabase.GetClientsWithResponse(context.Background(), &postgrest.GetClientsParams{}); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, api.Error{
					Error:       "query error",
					Description: "unable to query clients",
				})
			} else if len(*clients.JSON200) == 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.Error{
					Error:       "unauthorized",
					Description: "client credentials rejected",
				})
			} else {
				c.Next()
			}
		}
	}
}
