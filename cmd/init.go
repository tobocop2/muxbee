package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/tobocop2/beetrix/internal/config"
)

var initServerName string
var initPort int

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize beetrix configuration",
	Long:  "Create configuration directory and generate settings.yaml with sensible defaults.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().StringVar(&initServerName, "server-name", "localhost", "Matrix server name")
	initCmd.Flags().IntVar(&initPort, "port", 8008, "Listen port for the Matrix API")
	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, _ []string) error {
	if config.Exists() {
		fmt.Fprintln(os.Stderr, "Configuration already exists at", config.SettingsPath())
		fmt.Fprintln(os.Stderr, "Remove it first or use 'beetrix nuke' to start fresh.")
		return fmt.Errorf("config already exists")
	}

	cfg, err := config.NewDefault()
	if err != nil {
		return fmt.Errorf("generating config: %w", err)
	}

	cfg.ServerName = initServerName
	cfg.ListenPort = initPort

	if !config.IsPortAvailable(cfg.Port()) {
		fmt.Fprintf(os.Stderr, "Warning: port %d is in use\n", cfg.Port())
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println("Configuration created at", config.SettingsPath())
	fmt.Println()
	fmt.Println("Admin credentials:")
	fmt.Printf("  Username: %s\n", cfg.Admin.Username)
	fmt.Printf("  Password: %s\n", cfg.Admin.Password)
	fmt.Println()
	fmt.Println("Next: run 'beetrix up' to start the server.")
	return nil
}
