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
	"github.com/train360-corp/projconf/go/internal/flags"
	"github.com/train360-corp/projconf/go/internal/utils/random"
	"github.com/train360-corp/projconf/go/pkg"
	server2 "github.com/train360-corp/projconf/go/pkg/server"
	"github.com/train360-corp/projconf/go/pkg/server/state"
	"github.com/train360-corp/projconf/go/pkg/supabase/migrations"
	"github.com/train360-corp/supago"
	"go.uber.org/zap/zapcore"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var withStudio bool = false

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

		if strings.TrimSpace(server2.AdminApiKey) == "" {
			server2.AdminApiKey = random.String(32)
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

		// root ctx that is canceled on SIGINT/SIGTERM.
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

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

		// catch panic
		defer func() {
			if r := recover(); r != nil {
				if logLevel != zapcore.DebugLevel {
					shutdown()
				}
				logger.Fatalf("%v", r)
			}
		}()

		// get data directory
		dir, err := server2.EnsureSystemProjConfDir()
		if err != nil {
			panic(fmt.Sprintf("unable to get system-wide data directory: %v", err))
		} else {
			logger.Debugf("system-wide data directory: %s", dir)
		}

		// create config
		cfg, err := supago.ConfigBuilder().
			Platform("projconf").
			GetEncryptionKeyUsing(supago.EncryptionKeyFromFile(filepath.Join(dir, "postgres", "encryption.key"))).
			BuildE()
		if err != nil {
			panic(fmt.Sprintf("unable to create config: %v", err))
		}

		// customize config
		cfg.Global.DebugMode = logLevel == zapcore.DebugLevel
		cfg.Database.DataDirectory = filepath.Join(dir, "postgres", "data")
		logger.Debugf("database data directory: %s", cfg.Database.DataDirectory)

		// init http server (early)
		state.Get().SetAnonymousKey(cfg.Keys.PublicJwt)
		if err := server2.Init(logger); err != nil {
			logger.Panicf("failed to initialize server: %v", err)
		}
		srv := server2.Start(ctx)

		// patch postgres after-start to handle migrations
		postgres := supago.Services.Postgres(*cfg)
		postgres.Cmd = append(postgres.Cmd, "-c", "projconf.x_admin_api_key="+server2.AdminApiKey)
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
		nodeDebug := ""
		if logLevel == zapcore.DebugLevel {
			nodeDebug = "undici"
		}

		// init supago
		sg = supago.New(cfg).
			SetLogger(logger).
			AddService(
				postgres.Build(),
				supago.Services.Kong,
				supago.Services.Postgrest,
			)

		// only start studio if enabled
		// (for security reasons, since studio auth is not implemented yet)
		if withStudio {
			sg.AddService(supago.Service{
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
					fmt.Sprintf("%s=%s", "SUPABASE_FRONTEND_URL", "http://127.0.0.1:8000"),
					fmt.Sprintf("%s=%s", "SUPABASE_BACKEND_URL", cfg.Kong.URLs.Kong),
					fmt.Sprintf("%s=%s", "SUPABASE_PUBLISHABLE_OR_ANON_KEY", cfg.Keys.PublicJwt),
					fmt.Sprintf("%s=%s", "X_ADMIN_API_KEY", server2.AdminApiKey),
					fmt.Sprintf("%s=%s", "NODE_DEBUG", nodeDebug),
				},
			}.Build())
		}

		// run services
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

	// studio flags
	serveCommand.Flags().BoolVarP(&withStudio, "with-studio", "S", withStudio, "start the ProjConf web interface")

	// connection flags
	flags.SetupAdminApiKeyFlag(serveCommand, &server2.AdminApiKey)
	serveCommand.Flags().StringVarP(&server2.Host, "host", "H", server2.Host, fmt.Sprintf("host to serveCommand on (default: %s)", server2.Host))
	serveCommand.Flags().Uint16VarP(&server2.Port, "port", "P", server2.Port, fmt.Sprintf("port to serveCommand on (default: %d)", server2.Port))

	// logging flags
	serveCommand.Flags().StringVarP(&logLevelStr, "log-level", "l", logLevelStr, "log level (default: warn; available: debug | info | warn | error | panic | fatal)")
	serveCommand.Flags().BoolVar(&logJsonFmt, "log-json", logJsonFmt, "log json-formatted output (default: false/human-readable)")
}
