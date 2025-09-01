/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package supabase

import (
	"github.com/gin-gonic/gin"
	"github.com/train360-corp/projconf/pkg/consts"
	"github.com/train360-corp/projconf/pkg/server/state"
	"github.com/train360-corp/projconf/pkg/supabase/database"
)

type Client struct {
	config *Config
	auth   *AuthConfig
	self   *database.PublicClientsSelect
}

func GetWithAuth(authConfig *AuthConfig) *Client {
	client := &Client{
		config: &Config{
			Url:     "http://127.0.0.1:3000",
			AnonKey: state.Get().AnonKey(),
		},
		auth: authConfig,
	}
	return client
}

func GetFromContext(ctx *gin.Context) *Client {
	return GetWithAuth(&AuthConfig{
		Id:          ctx.GetHeader(consts.X_CLIENT_SECRET_ID),
		Secret:      ctx.GetHeader(consts.X_CLIENT_SECRET),
		AdminAPIKey: ctx.GetHeader(consts.X_ADMIN_API_KEY),
	})
}

type request struct {
	endpoint string
	single   bool
}
