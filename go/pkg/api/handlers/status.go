/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/train360-corp/projconf/go/pkg"
	"github.com/train360-corp/projconf/go/pkg/api"
	"github.com/train360-corp/projconf/go/pkg/server/state"
	"net/http"
)

func (r RouteHandlers) GetStatusV1(c *gin.Context) {
	c.JSON(http.StatusOK, api.Status{
		Server: struct {
			IsReady bool   `json:"is_ready"`
			Version string `json:"version"`
		}{
			IsReady: state.Get().IsAlive(),
			Version: pkg.Version,
		},
		Services: struct {
			Postgres  bool `json:"postgres"`
			Postgrest bool `json:"postgrest"`
		}{
			Postgres:  state.Get().IsPostgresAlive(),
			Postgrest: state.Get().IsPostgrestAlive(),
		},
	})
}

func (r RouteHandlers) GetStatusReadyV1(c *gin.Context) {
	if state.Get().IsAlive() {
		c.JSON(http.StatusOK, api.Ready{Msg: "ready"})
	} else {
		c.JSON(http.StatusServiceUnavailable, api.Error{
			Error:       "not available",
			Description: "one or more services are not ready",
		})
	}
}
