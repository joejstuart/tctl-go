package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"

	// Import runners and scanners to register them
	_ "github.com/yourname/tctl/internal/runner"
	_ "github.com/yourname/tctl/internal/scanner"
)

const version = "0.2.0"

func main() {
	// Ensure config directory exists
	config.EnsureConfigDir()

	rootCmd := &cobra.Command{
		Use:   "tctl",
		Short: "Tool management CLI",
		Long: `tctl manages a personal library of tools.

Register tool directories from anywhere, then discover and run tools globally.

Quick start:
  tctl add ~/my-tools      # Register a directory
  tctl list                # See all tools
  tctl run my-tool         # Run a tool`,
		Version: version,
	}

	// Source management
	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(removeCmd())
	rootCmd.AddCommand(sourcesCmd())

	// Tool discovery
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(whatCmd())
	rootCmd.AddCommand(findCmd())
	rootCmd.AddCommand(whereCmd())
	rootCmd.AddCommand(showCmd())

	// Tool execution
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(getCmd())

	// Maintenance
	rootCmd.AddCommand(newCmd())
	rootCmd.AddCommand(syncCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(lintCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
