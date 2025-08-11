package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	ginvalidator "github.com/oapi-codegen/gin-middleware"
	"github.com/train360-corp/projconf/internal/server/api"
	"github.com/train360-corp/projconf/internal/supabase"
	"net/http"
	"os"
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

	SUPABASE_URL := os.Getenv("PROJCONF_SUPABASE_URL")
	if SUPABASE_URL == "" {
		return nil, errors.New("environment variable 'PROJCONF_SUPABASE_URL' is not set")
	}

	SUPABASE_ANON_KEY := os.Getenv("PROJCONF_SUPABASE_ANON_KEY")
	if SUPABASE_ANON_KEY == "" {
		return nil, errors.New("environment variable 'PROJCONF_SUPABASE_ANON_KEY' is not set")
	}

	r := gin.New()
	r.Use(gin.Recovery())

	swagger := api.MustSpec()
	opts := &ginvalidator.Options{}
	opts.Options.AuthenticationFunc = func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {

		req := input.RequestValidationInput.Request
		id := req.Header.Get("x-client-secret-id")
		sec := req.Header.Get("x-client-secret")

		if id == "" {
			return errors.New("missing 'x-client-secret-id' header")
		} else if sec == "" {
			return errors.New("missing 'x-client-secret' header")
		}

		sb := supabase.GetWithAuth(&supabase.Config{
			Url:     SUPABASE_URL,
			AnonKey: SUPABASE_ANON_KEY,
		}, &supabase.AuthConfig{
			Id:     id,
			Secret: sec,
		})

		if client, clientErr := sb.GetSelf(); clientErr != nil {

		}

		// TODO: attempt to get the client to verify if the authentication succeeds

		return errors.New("authentication not implemented")
	}
	r.Use(ginvalidator.OapiRequestValidatorWithOptions(swagger, opts))

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
