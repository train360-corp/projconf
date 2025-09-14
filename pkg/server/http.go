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
	"github.com/gin-gonic/gin"
	ginvalidator "github.com/oapi-codegen/gin-middleware"
	"github.com/train360-corp/projconf/internal/defaults"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/train360-corp/projconf/pkg/api"
	"go.uber.org/zap"
	"net/http"
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

func Init(logger *zap.SugaredLogger) (err error) {

	mu.Lock()
	defer mu.Unlock()

	if server != nil {
		err = fmt.Errorf("server already initialized")
		return
	}

	once.Do(func() {
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
		router.Use(gin.Recovery()) // handle panics, etc.
		router.Use(auth)           // authentication middleware

		// use custom validation
		swagger := api.MustSpec()
		router.Use(ginvalidator.OapiRequestValidator(swagger))

		// use route handlers
		api.RegisterHandlers(router, GetServerInterface())

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
