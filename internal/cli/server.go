/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/train360-corp/projconf/internal/config"
	"github.com/train360-corp/projconf/internal/docker"
	"github.com/train360-corp/projconf/internal/docker/services/database"
	"github.com/train360-corp/projconf/internal/docker/types"
	"github.com/train360-corp/projconf/internal/fs"
	"github.com/train360-corp/projconf/internal/server"
	"github.com/train360-corp/projconf/internal/supabase/migrations"
	"github.com/train360-corp/projconf/internal/utils/postgres"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

func ServerCommand() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "commands to serve a ProjConf server",
		Subcommands: []*cli.Command{
			serveCommand(),
			updateCommand(),
			resetCommand(),
		},
	}
}

const (
	projectLabelKey   = "com.docker.compose.project"
	projectLabelValue = "projconf"
	networkName       = "projconf-net"
)

func resetCommand() *cli.Command {
	return &cli.Command{
		Name:  "reset",
		Usage: "destructively removes a ProjConf server and resets the config",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "confirm",
				Aliases:  []string{"y"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {

			root, err := fs.GetUserRoot()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// remove local db folder
			if err := os.RemoveAll(filepath.Join(root, "db")); err != nil {
				return errors.New(fmt.Sprintf("failed to remove database: %s", err))
			} else {
				log.Printf("removed database directory")
			}

			// reset db config
			cfg.Supabase = config.GenDefaultConfig().Supabase
			if err := cfg.Flush(); err != nil {
				return errors.New(fmt.Sprintf("failed to reset database config: %s", err))
			} else {
				log.Printf("reset database config")
			}

			return nil
		},
	}
}

func updateCommand() *cli.Command {
	return &cli.Command{
		Name:    "update",
		Usage:   "update a ProjConf server",
		Aliases: []string{"migrate"},
		Action: func(c *cli.Context) error {

			tmp, err := fs.GetTempRoot()
			if err != nil {
				return err
			}

			defer os.RemoveAll(tmp)

			// honor existing context and add SIGINT/SIGTERM
			ctx, stop := signal.NotifyContext(c.Context, syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			// --- Clean start (blocking, cancel-aware)
			if err := mustCleanStart(ctx); err != nil {
				return err
			}

			// load config once, handle error
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}
			env := types.SharedEnv{ // (or SharedEnv if that's the real name)
				PGPASSWORD:  cfg.Supabase.Db.Password,
				JWT_SECRET:  cfg.Supabase.JwtSecret,
				ANON_KEY:    cfg.Supabase.Keys.Public,
				SERVICE_KEY: cfg.Supabase.Keys.Private,
			}

			db := &database.Service{}
			runErrCh := make(chan error, 1)
			go func() {
				runErrCh <- docker.RunService(ctx, db, env)
			}()
			if err := db.WaitFor(ctx); err != nil {
				return fmt.Errorf("waiting for %q: %w", db.GetDisplay(), err)
			}

			// Non-blocking check: if RunService already errored, surface it.
			select {
			case err := <-runErrCh:
				if err != nil {
					return fmt.Errorf("docker service %q error: %w", db.GetDisplay(), err)
				}
			default:
			}

			if err := postgres.ExecuteOnEmbeddedDatabase(c.Context, migrations.MigrationsSchemaStatements); err != nil {
				return errors.New(fmt.Sprintf("failed to apply migrations schema migrations: %s", err))
			} else {
				log.Printf("migrations schema migrated successfully")
			}

			existingMigrations, err := migrations.LoadSchemaMigrations(c.Context)
			if err != nil {
				return errors.New(fmt.Sprintf("failed to load existing schema migrations: %s", err.Error()))
			}

			m, _ := migrations.Get()
			commands := make([]string, 0)
			applied := 0
			passed := 0
			for _, migration := range m {

				parts := strings.Split(migration.Name, "_")
				version := parts[0]
				name := strings.TrimSuffix(parts[1], ".sql")

				alreadyApplied := false
				for _, existingMigration := range existingMigrations {
					if existingMigration.Version == version {
						alreadyApplied = true
						passed++
					}
				}

				if !alreadyApplied {
					applied++
					commands = append(commands, string(migration.Data))
					commands = append(commands, fmt.Sprintf("INSERT INTO supabase_migrations.schema_migrations (version, name) VALUES ('%s', '%s')", version, name))
				}
			}
			if err := postgres.ExecuteOnEmbeddedDatabase(c.Context, commands); err != nil {
				return errors.New(fmt.Sprintf("failed to apply migrations: %s", err))
			} else {
				log.Printf("%v migrations applied successfully (%v already applied)", applied, passed)
			}

			return nil
		},
	}
}

func serveCommand() *cli.Command {
	srvCfg := server.Config{}
	sharedCfg, flags := config.GetServerFlags()
	return &cli.Command{
		Name:    "serve",
		Aliases: []string{"start"},
		Usage:   "start the HTTP server",
		Flags: append(flags,
			&cli.StringFlag{
				Name:        "host",
				Usage:       "host interface to bind",
				Value:       "0.0.0.0",
				Destination: &srvCfg.Host,
				EnvVars:     []string{"PROJCONF_HOST"},
			},
			&cli.IntFlag{
				Name:        "port",
				Usage:       "port to listen on",
				Value:       8080,
				Destination: &srvCfg.Port,
				EnvVars:     []string{"PROJCONF_PORT"},
			}),
		Action: func(c *cli.Context) error {

			srvCfg.AdminAPIKey = sharedCfg.AdminAPIKey

			// Create HTTP server
			srv, err := server.NewHTTPServer(&srvCfg)
			if err != nil {
				return err
			}

			// Handle SIGINT/SIGTERM
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			cfg, _ := config.Load()
			env := types.SharedEnv{
				PGPASSWORD:             cfg.Supabase.Db.Password,
				JWT_SECRET:             cfg.Supabase.JwtSecret,
				ANON_KEY:               cfg.Supabase.Keys.Public,
				SERVICE_KEY:            cfg.Supabase.Keys.Private,
				PROJCONF_ADMIN_API_KEY: sharedCfg.AdminAPIKey,
			}

			services := docker.GetServices()
			errCh := make(chan error, len(services)+2)

			// --- Clean start (blocking, cancel-aware)
			if err := mustCleanStart(ctx); err != nil {
				return err
			}

			// --- Start Docker services
			log.Println("starting Docker services...")
			for _, service := range services {
				svc := service // capture

				go func(s types.Service) {
					if err := docker.RunService(ctx, s, env); err != nil {
						errCh <- fmt.Errorf("docker service %q error: %w", s.GetDisplay(), err)
						return
					}
				}(svc)

				// Gate each service on readiness; cancels immediately on SIGINT/SIGTERM.
				if err := svc.WaitFor(ctx); err != nil {
					return fmt.Errorf("error while waiting for %q: %w", svc.GetDisplay(), err)
				}

				log.Printf("started %q service\n", svc.GetDisplay())
			}

			// --- Start HTTP server
			go func() {
				addr := fmt.Sprintf("%s:%d", srvCfg.Host, srvCfg.Port)
				log.Printf("HTTP server listening on http://%s\n", addr)
				log.Println(fmt.Sprintf("admin access token: %s", sharedCfg.AdminAPIKey))
				if err := srv.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					errCh <- fmt.Errorf("http server error: %w", err)
				}
			}()

			// --- Dedicated shutdown goroutine
			doneCh := make(chan struct{})
			go func() {
				<-ctx.Done()
				log.Println("Shutdown signal received; shutting down HTTP server...")
				_ = srv.Shutdown(context.Background())
				close(doneCh)
			}()

			// --- Wait for graceful shutdown or a real error
			select {
			case <-doneCh:
				return nil
			case err := <-errCh:
				_ = srv.Shutdown(context.Background())
				return err
			}
		},
	}
}

func mustCleanStart(ctx context.Context) error {
	// --- Clean start (blocking, cancel-aware)
	log.Println("cleaning up old containers and network...")
	if err := removeProjectContainers(ctx, projectLabelKey, projectLabelValue); err != nil {
		return fmt.Errorf("cleanup containers: %w", err)
	}
	if err := removeNetworkIfExists(ctx, networkName); err != nil {
		return fmt.Errorf("remove network: %w", err)
	}
	if err := ensureNetwork(ctx, networkName); err != nil {
		return fmt.Errorf("create network: %w", err)
	}
	return nil
}

// removeProjectContainers stops & removes any containers with the given project label.
func removeProjectContainers(ctx context.Context, labelKey, labelValue string) error {
	ids, err := listContainersByLabel(ctx, labelKey, labelValue)
	if err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	rmArgs := append([]string{"rm", "-f"}, ids...)
	return runDocker(ctx, rmArgs...)
}

func listContainersByLabel(ctx context.Context, labelKey, labelValue string) ([]string, error) {
	out, err := exec.CommandContext(
		ctx, "docker", "ps", "-aq", "--filter", "label="+labelKey+"="+labelValue,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}
	return strings.Fields(string(out)), nil
}

func removeNetworkIfExists(ctx context.Context, name string) error {
	// Try remove; ignore "not found" errors.
	cmd := exec.CommandContext(ctx, "docker", "network", "rm", name)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	_ = cmd.Run() // ignore result
	return nil
}

func ensureNetwork(ctx context.Context, name string) error {
	// If exists, done.
	out, err := exec.CommandContext(ctx, "docker", "network", "listProjectsSubcommand", "-q", "--filter", "name=^"+name+"$").Output()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		return nil
	}
	// Create it.
	cmd := exec.CommandContext(ctx, "docker", "network", "create", name)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	if err := cmd.Run(); err != nil {
		// If a race created it meanwhile, treat as success.
		if ierr := exec.CommandContext(ctx, "docker", "network", "inspect", name).Run(); ierr == nil {
			return nil
		}
		return fmt.Errorf("unable to create docker network %q: %w", name, err)
	}
	return nil
}

func runDocker(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker %v failed: %w", args, err)
	}
	return nil
}
