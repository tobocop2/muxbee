package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop the running beetrix server",
	Long:  "Signal a running beetrix instance to shut down gracefully via Unix socket.",
	RunE:  runDown,
}

func init() {
	rootCmd.AddCommand(downCmd)
}

func runDown(_ *cobra.Command, _ []string) error {
	// TODO: Implement Unix socket IPC to signal running instance.
	return fmt.Errorf("not yet implemented")
}
