package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"moon-prism-power/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cli.RunMPPAuth(ctx, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
