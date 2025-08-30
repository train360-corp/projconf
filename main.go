/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/train360-corp/projconf/cmd"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, shutdown := withGracefulSignals(context.Background(), 10*time.Second)
	defer shutdown()

	if err := cmd.CLI().ExecuteContext(ctx); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			os.Exit(130) // conventional for SIGINT
		}
		fmt.Fprintln(os.Stderr, color.RedString("error: %v", err))
		os.Exit(1)
	}
}

func withGracefulSignals(parent context.Context, grace time.Duration) (context.Context, func()) {
	ctx, cancel := context.WithCancel(parent)

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh // first signal
		log.Printf("received interrupt — beginning graceful shutdown (%s grace)...", grace)
		cancel() // tell workers to stop

		timer := time.NewTimer(grace)
		defer timer.Stop()

		select {
		case <-timer.C:
			log.Println("grace period elapsed — forcing exit.")
			os.Exit(1)
		case <-sigCh: // second real signal
			log.Println("second interrupt — forcing exit.")
			os.Exit(2)
		}
	}()

	// IMPORTANT: do NOT close(sigCh)
	stop := func() {
		signal.Stop(sigCh) // stop delivering to sigCh
		cancel()
	}
	return ctx, stop
}
