/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package serve

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/commands/server"
	"github.com/train360-corp/projconf/internal/commands/server/serve/postgres"
	"github.com/train360-corp/projconf/internal/commands/server/serve/postgrest"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/train360-corp/projconf/pkg/server/state"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func Command(c *cobra.Command, args []string) {

	sigCtx, stopSig := signal.NotifyContext(c.Context(), os.Interrupt, syscall.SIGTERM)
	defer stopSig()
	ctx, cancel := context.WithCancel(sigCtx)
	defer cancel()

	// shared vars
	pgPass := utils.RandomString(32)

	// setup Logger
	InitLogger()
	mustLogger()
	defer Logger.Sync()
	Logger.Debug("successfully initialized Logger")

	Logger.Info("starting server")

	// create scratch/temp dir
	tmpDir, err := os.MkdirTemp("", "projconf-*")
	if err != nil {
		Logger.Fatal(fmt.Sprintf("failed to create temporary/working directory: %v", err))
	}
	defer os.RemoveAll(tmpDir) // deletes whole dir recursively
	Logger.Debug(fmt.Sprintf("temporary/working directory: %s", tmpDir))

	// create data dir
	dataDir, err := server.EnsureSystemProjConfDir()
	if err != nil {
		Logger.Fatal(fmt.Sprintf("failed to ensure system projconf dir: %v", err))
	}
	Logger.Debug(fmt.Sprintf("system projconf dir: %s", dataDir))

	// setup client
	Logger.Debug("initializing docker client")
	initCli(ctx)
	mustCli()
	defer cli.Close()
	Logger.Info("successfully initialized docker client")

	// setup network
	Logger.Debug("initializing docker network")
	if err := initNetwork(ctx); err != nil {
		Logger.Fatal(fmt.Sprintf("failed to initalize network: %v", err))
	}
	Logger.Info(fmt.Sprintf("successfully initialized docker network (%v)", previewString(networkID)))

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
				Logger.Warn(fmt.Sprintf("stop failed: %v", err))
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
		Logger.Fatal(fmt.Sprintf("failed to initialize postgres: %v", err))
	} else {
		id, stop, err := runService(ctx, pg, func() {
			state.Get().SetPostgresAlive(false)
		})
		if err != nil {
			Logger.Fatal(fmt.Sprintf("failed to run postgres: %v", err))
		} else {
			addStop(stop)
			Logger.Debug("waiting for postgres ready status")

			// wait for postgres to come alive
			if err := waitPostgresReady(ctx, id); err != nil {
				stopAll()
				Logger.Fatal(fmt.Sprintf("failed to wait for postgres ready: %v", err))
			}

			Logger.Debug("patching postgres password")
			output, err := exec(ctx, id, []string{
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
				Logger.Fatal(fmt.Sprintf("failed to patch postgres password: %v (%s)", err, strings.ReplaceAll(strings.TrimSpace(output), "\n", "\\n")))
			} else {
				Logger.Debug(fmt.Sprintf("postgres password patched: %s", strings.ReplaceAll(strings.TrimSpace(output), "\n", "\\n")))
			}

			Logger.Info(fmt.Sprintf("successfully started postgres (%v)", previewString(id)))

			Logger.Debug("migrating postgres")
			if err := migrate(ctx, id); err != nil {
				Logger.Fatal(fmt.Sprintf("failed to migrate postgres: %v", err))
			}
			Logger.Info(fmt.Sprintf("successfully migrated postgres (%v)", previewString(id)))

			state.Get().SetPostgresAlive(true)
		}
	}

	if id, stop, err := runService(ctx, postgrest.Service(postgrest.ServiceRequest{
		JwtSecret:        state.Get().JwtSecret(),
		PostgresPassword: pgPass,
	}), func() {
		state.Get().SetPostgrestAlive(false)
	}); err != nil {
		Logger.Fatal(fmt.Sprintf("failed to initialize postgrest: %v", err))
	} else {
		addStop(stop)
		Logger.Debug("waiting for postgrest ready status")
		if err := waitHTTPReady(ctx, "http://127.0.0.1:3001/ready"); err != nil {
			stopAll()
			Logger.Fatal(fmt.Sprintf("failed to wait for postgrest ready: %v", err))
		}

		Logger.Info(fmt.Sprintf("successfully started postgrest (%v)", previewString(id)))
		state.Get().SetPostgrestAlive(true)
	}

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "OK")
		})

		// Start server on port 8080
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()

	// ---- block until ctx cancels, then shut everything down ----
	<-ctx.Done()

	// stop HTTP first
	//shutdownCtx, cancelSD := context.WithTimeout(context.Background(), 5*time.Second)
	//_ = srv.Shutdown(shutdownCtx)
	//cancelSD()

	// stop containers (uses ContainerStop inside)
	stopAll()

}
