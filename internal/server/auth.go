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
	"github.com/train360-corp/projconf/internal/consts"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/supabase"
	"net/http"
	"strings"
)

func auth(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		appCfg, _ := config.Load()

		// ----- ADMIN FLOW -----
		if c.GetHeader(consts.X_ADMIN_API_KEY) != "" {
			raw := strings.TrimSpace(c.GetHeader(consts.X_ADMIN_API_KEY))
			if raw == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
					Error:       "Unauthorized",
					Description: "missing 'x-admin-api-key' header",
				})
				return
			}

			// accept case-insensitive Bearer and trim
			token := raw
			if len(raw) >= 7 && strings.EqualFold(raw[:7], "Bearer ") {
				token = strings.TrimSpace(raw[7:])
			}

			// constant-time compare (same length)

			if len(token) != len(cfg.AdminAPIKey) ||
				subtle.ConstantTimeCompare([]byte(token), []byte(cfg.AdminAPIKey)) != 1 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, api.Unauthorized{
					Error:       "Unauthorized",
					Description: "invalid 'x-admin-api-key' header",
				})
				return
			}
			c.Next()
			return
		}

		// ----- CLIENT FLOW -----
		id := c.GetHeader(consts.X_CLIENT_SECRET_ID)
		sec := c.GetHeader(consts.X_CLIENT_SECRET)
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
