package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tobocop2/muxbee/internal/tui"
)

// Version is set at build time via ldflags
var Version = "0.1.1"

var rootCmd = &cobra.Command{
	Use:   "muxbee",
	Short: "Self-hosted Matrix server with messaging bridges",
	Long: `muxbee is a self-hosted Matrix server with messaging bridges.

Run without arguments to launch the TUI, or use subcommands for CLI access.`,
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.Run()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate("muxbee version {{.Version}}\n")
}
