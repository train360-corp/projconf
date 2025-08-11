package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	ginvalidator "github.com/oapi-codegen/gin-middleware"
	"github.com/train360-corp/projconf/internal/server/api"
	"net/http"
)

type HTTPServer struct {
	cfg    Config
	router *gin.Engine
	http   *http.Server
}

func NewHTTPServer(cfg Config) (*HTTPServer, error) {
	switch cfg.Mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
		gin.SetMode(cfg.Mode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	swagger := api.MustSpec()
	r.Use(ginvalidator.OapiRequestValidator(swagger))

	routes := &RouteHandlers{}
	api.RegisterHandlers(r, routes)

	return &HTTPServer{
		cfg:    cfg,
		router: r,
		http: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler: r,
		},
	}, nil
}

func (s *HTTPServer) Serve() error {
	return s.http.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
