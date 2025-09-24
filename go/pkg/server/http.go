/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	ginvalidator "github.com/oapi-codegen/gin-middleware"
	"github.com/train360-corp/projconf/go/internal/defaults"
	"github.com/train360-corp/projconf/go/internal/utils/validators"
	"github.com/train360-corp/projconf/go/pkg/api"
	"github.com/train360-corp/projconf/go/pkg/api/handlers"
	"github.com/train360-corp/projconf/go/pkg/server/state"
	"github.com/train360-corp/supago"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

var (
	AdminApiKey string
	Host        = defaults.ServerHost
	Port        = defaults.ServerPort

	server *ProjConfServer
	once   sync.Once
	mu     sync.Mutex
)

type ProjConfServer struct {
	router *gin.Engine
	http   *http.Server
	logger *zap.SugaredLogger
}

func Init(logger *zap.SugaredLogger, config *supago.Config) (err error) {

	mu.Lock()
	defer mu.Unlock()

	if server != nil {
		err = fmt.Errorf("server already initialized")
		return
	}

	once.Do(func() {

		state.Get().SetLogger(logger)

		if AdminApiKey == "" {
			err = fmt.Errorf("admin api key is empty")
			return
		} else if !validators.IsValidHost(Host) {
			err = fmt.Errorf("invalid host: %s", Host)
			return
		} else if !validators.IsValidPort(Port) {
			err = fmt.Errorf("invalid port: %d", Port)
			return
		}
		gin.SetMode(gin.ReleaseMode)

		router := gin.New()
		router.Use(gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, recovered any) {
			logger.Errorf("panic recovered: %v\n%s", recovered, debug.Stack())
			c.AbortWithStatusJSON(http.StatusInternalServerError, api.Error{
				Description: "a panic occurred and was recovered (see server logs for more details)",
				Error:       "panic recovered",
			})
		})) // handle panics, etc.
		router.Use(authHandler(config)) // authentication middleware

		// use custom validation
		swagger := api.MustSpec()

		// request validation
		router.Use(ginvalidator.OapiRequestValidatorWithOptions(swagger, &ginvalidator.Options{
			ErrorHandler: func(c *gin.Context, message string, statusCode int) {
				logger.Errorf("%s: %s", c.Request.URL.Path, message)
				c.AbortWithStatusJSON(statusCode, api.Error{
					Error:       "request validation failed",
					Description: "the request to the server failed validation (check server logs for more details)",
				})
			},
			Options: openapi3filter.Options{
				AuthenticationFunc:    openapi3filter.NoopAuthenticationFunc,
				IncludeResponseStatus: true,
			},
		}))

		// response validation
		router.Use(ginvalidator.OapiResponseValidatorWithOptions(swagger, &ginvalidator.Options{
			ErrorHandler: func(c *gin.Context, message string, statusCode int) {
				logger.Errorf("%s: %s", c.Request.URL.Path, message)
				c.AbortWithStatusJSON(statusCode, api.Error{
					Error:       "response validation failed",
					Description: "the response from the server failed validation (check server logs for more details)",
				})
			},
			Options: openapi3filter.Options{
				AuthenticationFunc:    openapi3filter.NoopAuthenticationFunc,
				IncludeResponseStatus: true,
			},
		}))

		// use route handlers
		api.RegisterHandlers(router, handlers.GetRouteHandlers("http://127.0.0.1:8000/rest/v1/", config.Keys.PublicJwt))

		// configure logger
		log := logger
		if log == nil {
			log = zap.NewNop().Sugar()
		}

		server = &ProjConfServer{
			logger: log,
			router: router,
			http: &http.Server{
				Addr:    fmt.Sprintf("%s:%d", Host, Port),
				Handler: router,
			},
		}
	})

	return
}

func Start(parentCtx context.Context) context.Context {

	ctx, done := context.WithCancelCause(parentCtx)

	if server == nil || server.http == nil || server.logger == nil {
		panic("server not initialized")
	}

	// Run the HTTP server
	go func() {
		server.logger.Infof("http server starting on %s:%d", Host, Port)
		if err := server.http.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) { // normal path when Shutdown() is invoked
				server.logger.Warnf("http server closed")
				done(nil)
			} else { // unexpected error
				server.logger.Warnf("http server exited with error: %v", err)
				done(err)
			}
		} else { // ListenAndServe returned nil (rare), treat as normal stop
			done(nil)
		}
	}()

	// graceful shutdown only when the *parent* cancels
	go func() {
		<-parentCtx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.http.Shutdown(shCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			server.logger.Errorf("server shutdown error: %v", err)
		}
	}()

	return ctx
}
