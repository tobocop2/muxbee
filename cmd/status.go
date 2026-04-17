package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show server and bridge status",
	Long:  "Query a running beetrix instance for health information.",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(_ *cobra.Command, _ []string) error {
	// TODO: Implement Unix socket IPC to query running instance.
	return fmt.Errorf("not yet implemented")
}
