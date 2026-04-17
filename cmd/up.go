package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/tobocop2/beetrix/internal/config"
	"github.com/tobocop2/beetrix/internal/server"
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

	srv, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	if err := srv.Start(ctx); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}

	fmt.Printf("beetrix is running on http://%s:%d\n", cfg.Address(), cfg.Port())
	fmt.Printf("Admin user: %s\n", cfg.Admin.Username)

	if len(cfg.EnabledBridges) > 0 {
		health := srv.Health()
		for name, running := range health.Bridges {
			status := "running"
			if !running {
				status = "stopped"
			}
			fmt.Printf("Bridge %s: %s\n", name, status)
		}
	}

	fmt.Println("Press Ctrl-C to stop.")

	<-ctx.Done()
	fmt.Println("\nShutting down...")
	srv.Stop()
	return nil
}
