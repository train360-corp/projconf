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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/train360-corp/projconf/internal/server"
	"github.com/train360-corp/projconf/internal/utils"
	"github.com/train360-corp/projconf/internal/utils/validators"
	"github.com/train360-corp/projconf/pkg/docker"
	"github.com/train360-corp/projconf/pkg/docker/services"
	"github.com/zalando/go-keyring"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"math"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const (
	projectLabelKey   = "com.docker.compose.project"
	projectLabelValue = "projconf"
	networkName       = "projconf-net"
)

var (
	host string
	port uint16
)

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
		if !validators.IsValidHost(host) {
			return errors.New(fmt.Sprintf("invalid host: %s", host))
		}

		if !validators.IsValidPort(port) {
			return errors.New(fmt.Sprintf("invalid port: %d", port))
		}

		return nil
	},
	RunE: func(c *cobra.Command, args []string) error {

		ctx, cancel := context.WithCancel(c.Context())
		defer cancel()
		g, ctx := errgroup.WithContext(ctx)

		if err := mustCleanStart(ctx); err != nil {
			return fmt.Errorf("failed to clean-up dangling docker services: %v", err)
		}

		srvCfg := server.Config{
			Host:        host,
			Port:        port,
			AdminAPIKey: adminApiKey,
		}
		srv, err := server.NewHTTPServer(&srvCfg)
		if err != nil {
			return fmt.Errorf("unable to initialize server: %v", err.Error())
		}

		pgPass, err := keyring.Get("projconf", "postgres")
		if err != nil {
			if errors.Is(err, keyring.ErrNotFound) {
				pgPass = utils.RandomString(32)
				if err := keyring.Set("projconf", "postgres", pgPass); err != nil {
					return fmt.Errorf("could not store postgres password: %v", err)
				}
			} else {
				return fmt.Errorf("could not get postgres password from keyring: %v", err)
			}
		}

		dockerEnv, dockerEnvErr := docker.NewEnv(pgPass, adminApiKey)
		if dockerEnvErr != nil {
			return dockerEnvErr
		}

		// run required services
		for _, service := range services.GetServices() {
			g.Go(func() error {
				err := docker.RunService(ctx, service, *dockerEnv)
				if errors.Is(err, context.Canceled) { // treat graceful cancellation as success.
					return nil
				}
				return err
			})

			retries := 0
			for retries < 3 {
				alive, code := service.HealthCheck(ctx)
				if alive {
					log.Printf("started %v", service.Display())
					break
				}
				log.Printf("health check failed with code %d (retry=%d)", code, retries)
				time.Sleep(time.Duration(math.Pow(2, float64(retries))) * time.Second)
				retries++
			}
		}

		// run the server
		g.Go(func() error {
			addr := fmt.Sprintf("%s:%d", srvCfg.Host, srvCfg.Port)
			log.Printf("HTTP server listening on http://%s\n", addr)
			log.Println(fmt.Sprintf("admin access token: %s", adminApiKey))
			if err := srv.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("http server error: %w", err)
			}
			return nil
		})

		// gracefully shut the HTTP server down.
		g.Go(func() error {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(c.Context(), 15*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdownCtx) // ignore error; the error path is returned by the server goroutine
			return nil
		})

		// Wait for the first error; this also returns nil if everything ended cleanly.
		return g.Wait()
	},
}

func init() {
	serverCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVar(&adminApiKey, "admin-api-key", utils.RandomString(32), "authentication token for an admin api client")
	serveCmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "host to serve on")
	serveCmd.Flags().Uint16VarP(&port, "port", "P", 8080, "port to serve on")
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
