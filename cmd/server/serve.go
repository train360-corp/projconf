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
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/flags"
	"github.com/train360-corp/projconf/internal/utils/random"
	"github.com/train360-corp/projconf/pkg"
	"github.com/train360-corp/projconf/pkg/server"
	"github.com/train360-corp/projconf/pkg/server/state"
	"github.com/train360-corp/projconf/pkg/supabase/migrations"
	"github.com/train360-corp/supago"
	"go.uber.org/zap/zapcore"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// serveCommand represents the serveCommand command
var serveCommand = &cobra.Command{
	Use:           "serve",
	Aliases:       []string{"start", "run"},
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.NoArgs,
	Short:         "host a ProjConf server instance",
	Long: `Create an initialize a ProjConf server.

Default values will be created and stored in a local 
file accessible only by the current user.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {

		if strings.TrimSpace(server.AdminApiKey) == "" {
			server.AdminApiKey = random.String(32)
		}

		if level, err := zapcore.ParseLevel(logLevelStr); err != nil {
			return fmt.Errorf("invalid log level (\"%s\"): %v", logLevelStr, err)
		} else {
			logLevel = level
		}
		return nil
	},
	Run: func(cmd *cobra.Command, _ []string) {

		var sg *supago.SupaGo

		// Create a root ctx that is canceled on SIGINT/SIGTERM.
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		// setup custom logger
		logger := supago.NewOpinionatedLogger(logLevel, logJsonFmt)
		defer logger.Sync()

		// helper func
		shutdown := func() {
			logger.Debugf("shutdown triggering...")
			stop()
			if sg != nil {
				sg.Stop()
			}
		}

		// catch panic's handling
		defer func() {
			if r := recover(); r != nil {
				shutdown()
				logger.Fatalf("%v", r)
			}
		}()

		// get data directory
		dir, err := server.EnsureSystemProjConfDir()
		if err != nil {
			logger.Panicf("unable to get system-wide data directory: %v", err)
		} else {
			logger.Debugf("system-wide data directory: %s", dir)
		}

		// create config
		cfg, err := supago.NewRandomConfigE("projconf")
		if err != nil {
			logger.Panicf("unable to create config: %v", err)
		} else {
			sg = supago.New(*cfg).SetLogger(logger)
			cfg.Database.DataDirectory = filepath.Join(dir, "data", "postgres")
			logger.Debugf("database data directory: %s", cfg.Database.DataDirectory)
			cfg.Global.DebugMode = logLevel == zapcore.DebugLevel
		}

		// init http server (early)
		state.Get().SetAnonymousKey(cfg.Keys.PublicJwt)
		if err := server.Init(logger); err != nil {
			logger.Panicf("failed to initialize server: %v", err)
		}
		srv := server.Start(ctx)

		// patch postgres after-start to handle migrations
		postgres := supago.Services.Postgres(*cfg)
		postgres.Cmd = append(postgres.Cmd, "-c", "projconf.x_admin_api_key="+server.AdminApiKey)
		patchPgPass := postgres.AfterStart
		postgres.AfterStart = func(ctx context.Context, docker *client.Client, containerID string) error {

			// run hard-wired after-start
			err := patchPgPass(ctx, docker, containerID)
			if err != nil {
				return err
			}

			// get existing migrations
			existingMigrations, err := migrations.LoadExistingSchemaMigrations(ctx, docker, containerID)
			if err != nil {
				return err
			}

			// apply migrations
			if err, applied, passed := migrations.ApplyMigrations(ctx, docker, containerID, existingMigrations); err != nil {
				return err
			} else {
				logger.Debugf("%v new migrations applied (%v already applied, %d total)", applied, passed, applied+passed)
				return nil
			}
		}

		version := pkg.Version
		if strings.HasPrefix(version, "0.0.0-SNAPSHOT-") {
			version = "latest"
		}
		var projconfStudioPort uint16 = 3000
		projconf := &supago.Service{
			Image: fmt.Sprintf("%s:%s", "ghcr.io/train360-corp/projconf", version),
			Name:  "projconf-studio",
			Ports: []uint16{projconfStudioPort},
			Healthcheck: &container.HealthConfig{
				Test: []string{
					"CMD",
					"node",
					"-e",
					fmt.Sprintf("fetch('http://localhost:%d/api/v1/status/ready').then((r) => {if (r.status !== 200) throw new Error(r.status)})", projconfStudioPort),
				},
				Interval: 5 * time.Second,
				Timeout:  10 * time.Second,
				Retries:  3,
			},
			Env: []string{
				fmt.Sprintf("%s=%d", "PORT", projconfStudioPort),
				fmt.Sprintf("%s=%s", "SUPABASE_URL", "http://docker.host.internal:8000"),
				fmt.Sprintf("%s=%s", "SUPABASE_PUBLISHABLE_OR_ANON_KEY", cfg.Keys.PublicJwt),
				fmt.Sprintf("%s=%s", "X_ADMIN_API_KEY", server.AdminApiKey),
			},
		}

		// run services
		// - postgres/db
		// - postgrest
		sg.AddService(
			postgres,
			supago.Services.Kong(*cfg),
			supago.Services.Postgrest(*cfg),
			projconf,
		)
		if err := sg.RunForcefully(ctx); err != nil {
			logger.Panicf("failed to initialize runner: %v", err)
		}

		// TODO: dynamically tie to status of containers
		state.Get().SetPostgresAlive(true)
		state.Get().SetPostgrestAlive(true)

		<-srv.Done() // wait for stop signal
		logger.Warn("shutdown signal received")
		logger.Debugf("exit cause: %v", context.Cause(srv))
		shutdown()
		logger.Debugf("main context done")
	},
}

func init() {

	// connection flags
	flags.SetupAdminApiKeyFlag(serveCommand, &server.AdminApiKey)
	serveCommand.Flags().StringVarP(&server.Host, "host", "H", server.Host, fmt.Sprintf("host to serveCommand on (default: %s)", server.Host))
	serveCommand.Flags().Uint16VarP(&server.Port, "port", "P", server.Port, fmt.Sprintf("port to serveCommand on (default: %d)", server.Port))

	// logging flags
	serveCommand.Flags().StringVarP(&logLevelStr, "log-level", "l", logLevelStr, "log level (default: warn; available: debug | info | warn | error | panic | fatal)")
	serveCommand.Flags().BoolVar(&logJsonFmt, "log-json", logJsonFmt, "log json-formatted output (default: false/human-readable)")
}
