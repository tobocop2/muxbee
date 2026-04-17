package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/tobocop2/beetrix/internal/bridges"
	"github.com/tobocop2/beetrix/internal/config"
)

var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "Manage messaging bridges",
}

var bridgeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available bridges",
	RunE:  runBridgeList,
}

var bridgeEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a bridge",
	Args:  cobra.ExactArgs(1),
	RunE:  runBridgeEnable,
}

var bridgeDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a bridge",
	Args:  cobra.ExactArgs(1),
	RunE:  runBridgeDisable,
}

var bridgeLoginCmd = &cobra.Command{
	Use:   "login <name>",
	Short: "Start bridge login flow",
	Args:  cobra.ExactArgs(1),
	RunE:  runBridgeLogin,
}

func init() {
	bridgeCmd.AddCommand(bridgeListCmd)
	bridgeCmd.AddCommand(bridgeEnableCmd)
	bridgeCmd.AddCommand(bridgeDisableCmd)
	bridgeCmd.AddCommand(bridgeLoginCmd)
	rootCmd.AddCommand(bridgeCmd)
}

func runBridgeList(_ *cobra.Command, _ []string) error {
	var cfg *config.Config
	if config.Exists() {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION\tMODE\tSTATUS")

	for _, b := range bridges.List() {
		mode := "built-in"
		if !b.BridgeV2Ready {
			mode = "external"
		}

		status := "disabled"
		if cfg != nil && cfg.IsBridgeEnabled(b.Name) {
			status = "enabled"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", b.Name, b.Description, mode, status)
	}
	return w.Flush()
}

func runBridgeEnable(_ *cobra.Command, args []string) error {
	name := args[0]
	if !bridges.Exists(name) {
		return fmt.Errorf("unknown bridge: %s", name)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg.EnableBridge(name)

	if _, err := cfg.GetOrCreateBridgeTokens(name); err != nil {
		return fmt.Errorf("generating tokens: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Bridge '%s' enabled.\n", name)
	return nil
}

func runBridgeDisable(_ *cobra.Command, args []string) error {
	name := args[0]
	if !bridges.Exists(name) {
		return fmt.Errorf("unknown bridge: %s", name)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	cfg.DisableBridge(name)

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Bridge '%s' disabled.\n", name)
	return nil
}

func runBridgeLogin(_ *cobra.Command, args []string) error {
	name := args[0]
	if !bridges.Exists(name) {
		return fmt.Errorf("unknown bridge: %s", name)
	}

	// TODO: Implement bridge login flow (QR code rendering, user input, cookie paste).
	return fmt.Errorf("bridge login not yet implemented")
}
