/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package supabase

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/consts"
	"github.com/train360-corp/projconf/internal/supabase/database"
)

type Client struct {
	config *Config
	auth   *AuthConfig
	self   *database.PublicClientsSelect
}

func GetWithAuth(config *Config, authConfig *AuthConfig) *Client {
	client := &Client{config: config, auth: authConfig}
	return client
}

func GetFromContext(ctx *gin.Context) *Client {

	appCfg, err := config.Load()
	if err != nil {
		panic(errors.New(fmt.Sprintf("read config error: %v", err)))
	}

	return GetWithAuth(&Config{
		Url:     appCfg.Supabase.Url,
		AnonKey: appCfg.Supabase.Keys.Public,
	}, &AuthConfig{
		Id:          ctx.GetHeader(consts.X_CLIENT_SECRET_ID),
		Secret:      ctx.GetHeader(consts.X_CLIENT_SECRET),
		AdminAPIKey: ctx.GetHeader(consts.X_ADMIN_API_KEY),
	})
}

type request struct {
	endpoint string
	single   bool
}
