package commands

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/train360-corp/projconf/internal/docker"
	"github.com/train360-corp/projconf/internal/docker/services/database"
	"github.com/train360-corp/projconf/internal/docker/types"
	"github.com/train360-corp/projconf/internal/server"
	"github.com/urfave/cli/v2"
	"log"
	"os/signal"
	"syscall"
)

func ServerCommand() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "commands to serve a ProjConf server",
		Subcommands: []*cli.Command{
			serveCommand(),
		},
	}
}

// randomString returns a secure random string of length n.
func randomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes)
}

func serveCommand() *cli.Command {

	cfg := server.Config{}

	return &cli.Command{
		Name:  "serve",
		Usage: "start the HTTP server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "host",
				Usage:       "host interface to bind",
				Value:       "0.0.0.0",
				Destination: &cfg.Host,
				EnvVars:     []string{"PROJCONF_HOST"},
			},
			&cli.IntFlag{
				Name:        "port",
				Usage:       "port to listen on",
				Value:       8080,
				Destination: &cfg.Port,
				EnvVars:     []string{"PROJCONF_PORT"},
			},
		},
		Action: func(c *cli.Context) error {

			// create http server
			srv, err := server.NewHTTPServer(cfg)
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			// channel to capture goroutine errors
			errCh := make(chan error, 2)

			env := types.SharedEvn{
				PGPASSWORD: randomString(32),
				JWT_SECRET: randomString(32),
			}

			// start Docker services in the foreground (blocking) in its own goroutine.
			go func() {
				log.Println("starting Docker services...")
				if err := docker.RunDockerServices(env); err != nil {
					errCh <- fmt.Errorf("docker services error: %w", err)
					return
				}
				// if StartServices ever returns nil, that means Docker foreground exited cleanly.
				errCh <- fmt.Errorf("docker services exited")
			}()

			if err := (database.Service{}).WaitFor(ctx, env); err != nil {
				return fmt.Errorf("an error occurred while waiting for database: %w", err)
			}

			// start HTTP server (foreground) in its own goroutine.
			go func() {
				addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
				log.Printf("HTTP server listening on http://%s", addr)
				errCh <- srv.Serve()
			}()

			// wait for either a signal or an error from either goroutine.
			select {
			case <-ctx.Done():
				log.Println("Shutdown signal received; shutting down HTTP server...")
				// graceful HTTP shutdown; Docker foreground will be terminated by process exit.
				if err := srv.Shutdown(context.Background()); err != nil {
					return err
				}
				return nil

			case err := <-errCh:
				// if Docker dies or HTTP Serve returns an error, stop the server and exit.
				log.Println("Service error:", err)
				_ = srv.Shutdown(context.Background())
				return err
			}
		},
	}
}
