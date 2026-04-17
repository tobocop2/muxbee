package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tobocop2/beetrix/internal/config"
)

var nukeYes bool

var nukeCmd = &cobra.Command{
	Use:   "nuke",
	Short: "Remove all beetrix data and configuration",
	Long:  "Permanently delete all configuration, databases, and bridge data.",
	RunE:  runNuke,
}

func init() {
	nukeCmd.Flags().BoolVar(&nukeYes, "yes", false, "Skip confirmation prompt")
	rootCmd.AddCommand(nukeCmd)
}

func runNuke(_ *cobra.Command, _ []string) error {
	configDir := config.ConfigDir()
	dataDir := config.DataDir()

	if !config.Exists() {
		fmt.Println("Nothing to remove.")
		return nil
	}

	if !nukeYes {
		fmt.Println("This will permanently delete:")
		fmt.Printf("  %s\n", configDir)
		fmt.Printf("  %s\n", dataDir)
		fmt.Print("\nAre you sure? [y/N] ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if err := os.RemoveAll(configDir); err != nil {
		return fmt.Errorf("removing config: %w", err)
	}
	if err := os.RemoveAll(dataDir); err != nil {
		return fmt.Errorf("removing data: %w", err)
	}

	fmt.Println("All beetrix data removed.")
	return nil
}
