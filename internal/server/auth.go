/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	"crypto/subtle"
	"github.com/gin-gonic/gin"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/supabase"
	"net/http"
	"strings"
)

// Auth enforces admin/client auth and writes JSON errors directly.
// Admin:   Authorization (exact or "Bearer <key>") for paths prefixed with /v1/admin
// Client:  x-client-secret-id + x-client-secret on everything else
func auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		appCfg, _ := config.Load()
		path := c.Request.URL.Path

		// ----- ADMIN FLOW -----
		if strings.HasPrefix(path, "/v1/admin") {
			raw := strings.TrimSpace(c.GetHeader("Authorization"))
			if raw == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
					Error:       "Unauthorized",
					Description: "missing 'Authorization' header",
				})
				return
			}

			// accept case-insensitive Bearer and trim
			token := raw
			if len(raw) >= 7 && strings.EqualFold(raw[:7], "Bearer ") {
				token = strings.TrimSpace(raw[7:])
			}

			// constant-time compare (same length)
			if len(token) != len(config.GetGlobal().AdminAccessKey) ||
				subtle.ConstantTimeCompare([]byte(token), []byte(config.GetGlobal().AdminAccessKey)) != 1 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
					Error:       "Unauthorized",
					Description: "invalid 'Authorization' header",
				})
				return
			}
			c.Next()
			return
		}

		// ----- CLIENT FLOW -----
		id := c.GetHeader("x-client-secret-id")
		sec := c.GetHeader("x-client-secret")
		if id == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
				Error:       "Unauthorized",
				Description: "missing 'x-client-secret-id' header",
			})
			return
		}
		if sec == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
				Error:       "Unauthorized",
				Description: "missing 'x-client-secret' header",
			})
			return
		}

		sb := supabase.GetWithAuth(&supabase.Config{
			Url:     appCfg.Supabase.Url,
			AnonKey: appCfg.Supabase.Keys.Public,
		}, &supabase.AuthConfig{
			Id:     id,
			Secret: sec,
		})

		if _, err := sb.GetSelf(); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
				Error:       "Unauthorized",
				Description: "client credentials rejected",
			})
			return
		}

		c.Next()
	}
}
