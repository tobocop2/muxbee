package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/tobocop2/beetrix/internal/config"
	"github.com/tobocop2/beetrix/internal/homeserver"
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

	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("creating directories: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start Dendrite homeserver
	hs, err := homeserver.New(cfg)
	if err != nil {
		return fmt.Errorf("creating homeserver: %w", err)
	}

	if err := hs.Start(ctx); err != nil {
		return fmt.Errorf("starting homeserver: %w", err)
	}

	// Wait for homeserver to be ready
	if err := waitForHealth(hs); err != nil {
		hs.Stop()
		return fmt.Errorf("homeserver health check: %w", err)
	}

	// Create admin user on first run
	if err := hs.CreateUser(ctx, cfg.Admin.Username, cfg.Admin.Password, true); err != nil {
		fmt.Printf("Warning: could not create admin user: %v\n", err)
	}

	fmt.Printf("beetrix is running on http://%s:%d\n", cfg.Address(), cfg.Port())
	fmt.Printf("Admin user: %s\n", cfg.Admin.Username)
	fmt.Println("Press Ctrl-C to stop.")

	// Wait for shutdown signal
	<-ctx.Done()
	fmt.Println("\nShutting down...")
	hs.Stop()
	return nil
}

func waitForHealth(hs *homeserver.DendriteServer) error {
	for i := 0; i < 30; i++ {
		if err := hs.Health(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("homeserver did not become healthy within 15 seconds")
}
