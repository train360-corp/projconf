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
	"github.com/train360-corp/projconf/internal/commands/server/serve"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/train360-corp/projconf/pkg/api"
	"net/http"
	"sync"
	"time"
)

var (
	AdminApiKey string
	Host        string = "127.0.0.1"
	Port        uint16 = 8080

	server *ProjConfServer
	once   sync.Once
	mu     sync.Mutex
)

type ProjConfServer struct {
	router *gin.Engine
	http   *http.Server
}

func initHTTPServer() (err error) {

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

		server = &ProjConfServer{
			router: router,
			http: &http.Server{
				Addr:    fmt.Sprintf("%s:%d", Host, Port),
				Handler: router,
			},
		}
	})

	return
}

func startHTTPServer(ctx context.Context) <-chan error {

	if server == nil {
		serve.Logger.Fatal("server not initialized")
	}

	errCh := make(chan error, 1)
	go func() {
		serve.Logger.Debug(fmt.Sprintf("preparing to start http serve on %s:%d", Host, Port))
		if err := server.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serve.Logger.Warn(fmt.Sprintf("http server shutdown: %v", err))
			errCh <- err
			return
		}
		errCh <- nil
	}()

	// tie shutdown to ctx
	go func() {
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.http.Shutdown(shCtx); err != nil {
			serve.Logger.Error(fmt.Sprintf("server shutdown error: %v", err))
		}
	}()
	return errCh
}
