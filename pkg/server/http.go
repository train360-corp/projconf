/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	ginvalidator "github.com/oapi-codegen/gin-middleware"
	"github.com/train360-corp/projconf/pkg/api"
	"net/http"
)

type ProjConfServer struct {
	cfg    *Config
	router *gin.Engine
	http   *http.Server
}

func NewHTTPServer(cfg *Config) (*ProjConfServer, error) {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery()) // handle panics, etc.
	router.Use(auth(cfg))      // authentication middleware

	// use custom validation
	swagger := api.MustSpec()
	router.Use(ginvalidator.OapiRequestValidator(swagger))

	// use route handlers
	api.RegisterHandlers(router, GetServerInterface())

	return &ProjConfServer{
		cfg:    cfg,
		router: router,
		http: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler: router,
		},
	}, nil
}

func (s *ProjConfServer) Serve() error {
	return s.http.ListenAndServe()
}

func (s *ProjConfServer) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
