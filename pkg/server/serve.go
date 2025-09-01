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
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/commands/server/serve"
	"github.com/train360-corp/projconf/internal/commands/server/serve/postgres"
	"github.com/train360-corp/projconf/internal/commands/server/serve/postgrest"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/train360-corp/projconf/pkg/server/state"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Command(c *cobra.Command, _ []string) {

	sigCtx, stopSig := signal.NotifyContext(c.Context(), os.Interrupt, syscall.SIGTERM)
	defer stopSig()
	ctx, cancel := context.WithCancel(sigCtx)
	defer cancel()

	// shared vars
	pgPass := utils.RandomString(32)

	// setup Logger
	serve.InitLogger()
	serve.MustLogger()
	defer serve.Logger.Sync()
	serve.Logger.Debug("successfully initialized Logger")

	serve.Logger.Debug("starting server")
	if err := initHTTPServer(); err != nil {
		serve.Logger.Fatal(fmt.Sprintf("unable to initialize server: %v", err))
	}
	httpErrCh := startHTTPServer(ctx)
	serve.Logger.Info(fmt.Sprintf("http server listening at %s:%d", Host, Port))
	serve.Logger.Info(fmt.Sprintf("x-admin-api-key: %s", AdminApiKey))

	// create scratch/temp dir
	tmpDir, err := os.MkdirTemp("", "projconf-*")
	if err != nil {
		serve.Logger.Fatal(fmt.Sprintf("failed to create temporary/working directory: %v", err))
	}
	defer os.RemoveAll(tmpDir) // deletes whole dir recursively
	serve.Logger.Debug(fmt.Sprintf("temporary/working directory: %s", tmpDir))

	// create data dir
	dataDir, err := ensureSystemProjConfDir()
	if err != nil {
		serve.Logger.Fatal(fmt.Sprintf("failed to ensure system projconf dir: %v", err))
	}
	serve.Logger.Debug(fmt.Sprintf("system projconf dir: %s", dataDir))

	// setup client
	serve.Logger.Debug("initializing docker client")
	serve.InitCli(ctx)
	serve.MustCli()
	defer serve.Cli.Close()
	serve.Logger.Info("successfully initialized docker client")

	// setup network
	serve.Logger.Debug("initializing docker network")
	if err := serve.InitNetwork(ctx); err != nil {
		serve.Logger.Fatal(fmt.Sprintf("failed to initalize network: %v", err))
	}
	serve.MustNetwork()
	serve.Logger.Info(fmt.Sprintf("successfully initialized docker network (%v)", serve.PreviewString(serve.NetworkID)))

	var stops []func() error
	addStop := func(f func() error) {
		if f != nil {
			stops = append(stops, f)
		}
	}
	stopAll := func() {
		// stop in reverse creation order
		for i := len(stops) - 1; i >= 0; i-- {
			if err := stops[i](); err != nil {
				serve.Logger.Warn(fmt.Sprintf("stop failed: %v", err))
			}
		}
	}

	// start database
	if pg, err := postgres.Service(postgres.ServiceRequest{
		DbDataRoot:          dataDir,
		TmpFileRoot:         tmpDir,
		ProjConfAdminApiKey: "", // TODO: set
		PostgresPassword:    pgPass,
		JwtSecret:           state.Get().JwtSecret(),
	}); err != nil {
		serve.Logger.Fatal(fmt.Sprintf("failed to initialize postgres: %v", err))
	} else {
		id, stop, err := serve.RunService(ctx, pg, func() {
			state.Get().SetPostgresAlive(false)
		})
		if err != nil {
			serve.Logger.Fatal(fmt.Sprintf("failed to run postgres: %v", err))
		} else {
			addStop(stop)
			serve.Logger.Debug("waiting for postgres ready status")

			// wait for postgres to come alive
			if err := serve.WaitPostgresReady(ctx, id); err != nil {
				stopAll()
				serve.Logger.Fatal(fmt.Sprintf("failed to wait for postgres ready: %v", err))
			}

			serve.Logger.Debug("patching postgres password")
			output, err := serve.ExecInContainer(ctx, id, []string{
				"psql",
				"-h", "127.0.0.1",
				"-U", "supabase_admin",
				"-d", "postgres",
				"-v", "ON_ERROR_STOP=1",
				"-c",
				fmt.Sprintf(`
ALTER USER anon                    WITH PASSWORD '%s';
ALTER USER authenticated           WITH PASSWORD '%s';
ALTER USER authenticator           WITH PASSWORD '%s';
ALTER USER dashboard_user          WITH PASSWORD '%s';
ALTER USER pgbouncer               WITH PASSWORD '%s';
ALTER USER postgres                WITH PASSWORD '%s';
ALTER USER service_role            WITH PASSWORD '%s';
ALTER USER supabase_admin          WITH PASSWORD '%s';
ALTER USER supabase_auth_admin     WITH PASSWORD '%s';
ALTER USER supabase_read_only_user WITH PASSWORD '%s';
ALTER USER supabase_replication_admin WITH PASSWORD '%s';
ALTER USER supabase_storage_admin  WITH PASSWORD '%s';
`, pgPass, pgPass, pgPass, pgPass, pgPass, pgPass,
					pgPass, pgPass, pgPass, pgPass, pgPass, pgPass),
			})
			if err != nil {
				stopAll()
				serve.Logger.Fatal(fmt.Sprintf("failed to patch postgres password: %v (%s)", err, strings.ReplaceAll(strings.TrimSpace(output), "\n", "\\n")))
			} else {
				serve.Logger.Debug(fmt.Sprintf("postgres password patched: %s", strings.ReplaceAll(strings.TrimSpace(output), "\n", "\\n")))
			}

			serve.Logger.Info(fmt.Sprintf("successfully started postgres (%v)", serve.PreviewString(id)))

			serve.Logger.Debug("migrating postgres")
			if err := serve.Migrate(ctx, id); err != nil {
				serve.Logger.Fatal(fmt.Sprintf("failed to migrate postgres: %v", err))
			}
			serve.Logger.Info(fmt.Sprintf("successfully migrated postgres (%v)", serve.PreviewString(id)))

			state.Get().SetPostgresAlive(true)
		}
	}

	if id, stop, err := serve.RunService(ctx, postgrest.Service(postgrest.ServiceRequest{
		JwtSecret:        state.Get().JwtSecret(),
		PostgresPassword: pgPass,
	}), func() {
		state.Get().SetPostgrestAlive(false)
	}); err != nil {
		serve.Logger.Fatal(fmt.Sprintf("failed to initialize postgrest: %v", err))
	} else {
		addStop(stop)
		serve.Logger.Debug("waiting for postgrest ready status")
		if err := serve.WaitHTTPReady(ctx, "http://127.0.0.1:3001/ready"); err != nil {
			stopAll()
			serve.Logger.Fatal(fmt.Sprintf("failed to wait for postgrest ready: %v", err))
		}

		serve.Logger.Info(fmt.Sprintf("successfully started postgrest (%v)", serve.PreviewString(id)))
		state.Get().SetPostgrestAlive(true)
	}

	// ---- block until ctx cancels, then shut everything down ----
	select {
	case err := <-httpErrCh:
		// http server died unexpectedly (nil means clean/intentional shutdown)
		if err != nil {
			serve.Logger.Error("http server error", zap.Error(err))
		}
	case <-ctx.Done():
		serve.Logger.Debug("runtime context cancelled, shutting down")
		// context canceled (signal); start coordinated shutdown
	}

	// stop containers (uses ContainerStop inside)
	stopAll()

}
