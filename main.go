/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package main

import (
	"context"
	"github.com/train360-corp/projconf/cmd"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := WithGracefulSignals(context.Background(), 10*time.Second)
	defer stop()
	cmd.Execute(ctx)
}

// WithGracefulSignals returns a context that cancels on first SIGINT/SIGTERM,
// then waits 'grace' for cleanup. A second signal (or the grace expiring)
// forces process exit.
func WithGracefulSignals(parent context.Context, grace time.Duration) (context.Context, func()) {
	ctx, cancel := context.WithCancel(parent)

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh // first signal
		log.Printf("received interrupt — beginning graceful shutdown (%s grace)...", grace)
		cancel() // tell your workers to stop (RunService will see ctx.Done())

		timer := time.NewTimer(grace)
		defer timer.Stop()

		select {
		case <-timer.C:
			log.Println("grace period elapsed — forcing exit.")
			os.Exit(1)
		case <-sigCh: // second signal
			log.Println("second interrupt — forcing exit.")
			os.Exit(2)
		}
	}()

	// stop function: stop listening and cancel context
	stop := func() {
		signal.Stop(sigCh)
		close(sigCh)
		cancel()
	}
	return ctx, stop
}
