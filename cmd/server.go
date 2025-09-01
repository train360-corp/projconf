/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/commands/server/serve"
	"github.com/train360-corp/projconf/internal/fs"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/zalando/go-keyring"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	projectLabelKey          = "com.docker.compose.project"
	projectLabelValue        = "projconf"
	networkName              = "projconf-net"
	defaultServerHost        = "127.0.0.1"
	defaultServerPort uint16 = 8080
)

var (
	host        string
	port        uint16
	logLevelStr string = "warn"
)

var serverCmd = &cobra.Command{
	Use:           "server",
	Aliases:       []string{"start", "run"},
	SilenceUsage:  false,
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	Short:         "Manage a ProjConf server instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var resetCmd = &cobra.Command{
	Use:           "reset",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.NoArgs,
	Short:         "reset a ProjConf server instance",
	Long: `Delete all data associated with a ProjConf server.
This action cannot be undone.`,
	RunE: func(c *cobra.Command, args []string) error {

		// remove local db folder
		if root, err := fs.GetUserRoot(); err != nil {
			return err
		} else if err := os.RemoveAll(filepath.Join(root, "db")); err != nil {
			return errors.New(fmt.Sprintf("failed to remove database: %s", err))
		} else {
			log.Printf("removed database directory")
		}

		// remove keyring
		if err := keyring.DeleteAll("projconf"); err != nil {
			return fmt.Errorf("failed to remove keyring(s): %v", err)
		} else {
			log.Printf("removed keyring(s)")
		}

		return nil
	},
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
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

		if strings.TrimSpace(adminApiKey) == "" {
			adminApiKey = utils.RandomString(32)
		}

		if level, err := zapcore.ParseLevel(logLevelStr); err != nil {
			return fmt.Errorf("invalid log level (\"%s\"): %v", logLevelStr, err)
		} else {
			serve.LogLevel = level
		}

		if !validators.IsValidHost(host) {
			return errors.New(fmt.Sprintf("invalid host: %s", host))
		}

		if !validators.IsValidPort(port) {
			return errors.New(fmt.Sprintf("invalid port: %d", port))
		}

		return nil
	},
	Run: serve.Command,
	//RunE: func(c *cobra.Command, args []string) error {
	//
	//	log.Printf("starting ProjConf Server %s", pkg.Version)
	//
	//	ctx, cancel := context.WithCancel(c.Context())
	//	defer cancel()
	//	g, ctx := errgroup.WithContext(ctx)
	//
	//	if err := mustCleanStart(ctx); err != nil {
	//		return fmt.Errorf("failed to clean-up dangling docker services: %v", err)
	//	}
	//
	//	pgPass, err := keyring.Get("projconf", "postgres")
	//	if err != nil {
	//		if errors.Is(err, keyring.ErrNotFound) {
	//			pgPass = utils.RandomString(32)
	//			if err := keyring.Set("projconf", "postgres", pgPass); err != nil {
	//				return fmt.Errorf("could not store postgres password: %v", err)
	//			}
	//		} else {
	//			return fmt.Errorf("could not get postgres password from keyring: %v", err)
	//		}
	//	}
	//
	//	srvCfg := server.Config{
	//		Host:        host,
	//		Port:        port,
	//		AdminAPIKey: adminApiKey,
	//		AnonKey:     state.Get().AnonKey(),
	//	}
	//	srv, err := server.NewHTTPServer(&srvCfg)
	//	if err != nil {
	//		return fmt.Errorf("unable to initialize server: %v", err.Error())
	//	}
	//
	//	// run the server
	//	g.Go(func() error {
	//		addr := fmt.Sprintf("%s:%d", srvCfg.Host, srvCfg.Port)
	//		log.Printf("HTTP server listening on http://%s\t(admin access token: %s)\n", addr, adminApiKey)
	//		if err := srv.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	//			return fmt.Errorf("http server error: %w", err)
	//		}
	//		return nil
	//	})
	//
	//	// run required services
	//	for _, service := range services.GetServices() {
	//		g.Go(func() error {
	//			err := docker.RunService(ctx, service, docker.Env{
	//				PGPASSWORD:             pgPass,
	//				JWT_SECRET:             state.Get().JwtSecret(),
	//				PROJCONF_ADMIN_API_KEY: adminApiKey,
	//				SUPABASE_PUBLIC_KEY:    state.Get().AnonKey(),
	//			})
	//			if errors.Is(err, context.Canceled) { // treat graceful cancellation as success.
	//				return nil
	//			}
	//			return err
	//		})
	//
	//		retries := 0
	//		alive := false
	//		for !alive {
	//			a, code := service.HealthCheck(ctx)
	//			if a {
	//				alive = true
	//				log.Printf("started %v", service.Display())
	//				service.AfterStart()
	//			} else {
	//				if retries > 0 {
	//					log.Printf("health check failed with code %d (retry=%d)", code, retries)
	//				}
	//				retries++
	//				time.Sleep(time.Duration(math.Pow(2, float64(retries))) * time.Second)
	//				if retries > 3 {
	//					return fmt.Errorf("timed out waiting for health check")
	//				}
	//			}
	//		}
	//
	//		if service.ContainerName() == consts.PostgresContainerName {
	//
	//			log.Println("applying migrations...")
	//
	//			// do migrations
	//			conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
	//				"supabase_admin",
	//				pgPass,
	//				"127.0.0.1",
	//				5432,
	//				"postgres",
	//			))
	//			if err != nil {
	//				return fmt.Errorf("could not connect to postgres: %v", err)
	//			}
	//
	//			// apply migrations to build migrations_schema
	//			if err := postgres.Execute(ctx, conn, migrations.MigrationsSchemaStatements); err != nil {
	//				return errors.New(fmt.Sprintf("failed to apply migrations schema migrations: %s", err))
	//			} else {
	//				log.Printf("migrations schema migrated successfully")
	//			}
	//
	//			existingMigrations, err := migrations.LoadSchemaMigrations(ctx, conn)
	//			if err != nil {
	//				return errors.New(fmt.Sprintf("failed to load existing schema migrations: %s", err.Error()))
	//			}
	//
	//			m, _ := migrations.Get()
	//			commands := make([]string, 0)
	//			applied := 0
	//			passed := 0
	//			for _, migration := range m {
	//
	//				parts := strings.Split(migration.Name, "_")
	//				version := parts[0]
	//				name := strings.TrimSuffix(parts[1], ".sql")
	//
	//				alreadyApplied := false
	//				for _, existingMigration := range existingMigrations {
	//					if existingMigration.Version == version {
	//						alreadyApplied = true
	//						passed++
	//					}
	//				}
	//
	//				if !alreadyApplied {
	//					applied++
	//					commands = append(commands, string(migration.Data))
	//					commands = append(commands, fmt.Sprintf("INSERT INTO supabase_migrations.schema_migrations (version, name) VALUES ('%s', '%s')", version, name))
	//				}
	//			}
	//			if err := postgres.Execute(ctx, conn, commands); err != nil {
	//				return errors.New(fmt.Sprintf("failed to apply migrations: %s", err))
	//			} else {
	//				log.Printf("%v new migrations applied (%v already applied)", applied, passed)
	//			}
	//
	//			conn.Close(ctx)
	//		}
	//	}
	//
	//	// gracefully shut the HTTP server down.
	//	g.Go(func() error {
	//		<-ctx.Done()
	//		shutdownCtx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
	//		defer cancel()
	//		_ = srv.Shutdown(shutdownCtx) // ignore error; the error path is returned by the server goroutine
	//		return nil
	//	})
	//
	//	// Wait for the first error; this also returns nil if everything ended cleanly.
	//	return g.Wait()
	//},
}

func init() {

	serverCmd.AddCommand(resetCmd)
	resetCmd.Flags().Bool("confirm", false, "confirm deletion of the server")
	resetCmd.MarkFlagRequired("confirm")

	serverCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&adminApiKey, AdminApiKeyFlag, "", "authentication token for an admin api client")
	serveCmd.Flags().StringVarP(&host, "host", "H", defaultServerHost, "host to serve on")
	serveCmd.Flags().Uint16VarP(&port, "port", "P", defaultServerPort, "port to serve on")
	serveCmd.Flags().StringVarP(&logLevelStr, "log-level", "l", logLevelStr, "log level (default: warn; available: debug | info | warn | error | panic | fatal)")
	serveCmd.Flags().BoolVar(&serve.LogJsonFmt, "log-json", serve.LogJsonFmt, "log json-formatted output (default: human-readable)")

	rootCmd.AddCommand(serverCmd)
}

// mustCleanStart clean start (blocking, cancel-aware)
func mustCleanStart(ctx context.Context) error {
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
	cmd := exec.CommandContext(ctx, "docker", rmArgs...)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker %v failed: %w", rmArgs, err)
	}
	return nil
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
