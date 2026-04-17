package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/tobocop2/beetrix/internal/config"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the beetrix server",
	Long:  "Start Dendrite homeserver and all enabled bridges in-process.",
	RunE:  runUp,
}

func init() {
	rootCmd.AddCommand(upCmd)
}

func runUp(_ *cobra.Command, _ []string) error {
	if !config.Exists() {
		return fmt.Errorf("no configuration found; run 'beetrix init' first")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("Starting beetrix on %s:%d...\n", cfg.Address(), cfg.Port())

	// TODO: Replace with server.New(cfg).Start(ctx) once homeserver package is implemented.
	fmt.Println("Dendrite homeserver embedding not yet implemented.")
	fmt.Println("Waiting for signal (Ctrl-C to stop)...")

	<-ctx.Done()
	fmt.Println("\nShutting down...")
	return nil
}
