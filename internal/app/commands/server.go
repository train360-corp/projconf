package commands

import (
	"context"
	"fmt"
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
			srv, err := server.NewHTTPServer(cfg)
			if err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			errCh := make(chan error, 1)
			go func() {
				addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
				log.Printf("HTTP listening on http://%s", addr)
				errCh <- srv.Serve()
			}()

			select {
			case <-ctx.Done():
				log.Println("Shutdown signal received; shutting down...")
				return srv.Shutdown(context.Background())
			case err := <-errCh:
				return err
			}
		},
	}
}
