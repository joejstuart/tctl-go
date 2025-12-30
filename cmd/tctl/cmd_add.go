package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/scanner"
)

func addCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "add [path]",
		Short: "Register a tool directory",
		Long: `Register a directory containing tools with tctl.
If no path is given, registers the current directory.

Examples:
  tctl add                      # Register current directory
  tctl add ./tools              # Register ./tools
  tctl add ~/scripts -n scripts # Register with custom name`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := cfg.AddSource(path, name); err != nil {
				return err
			}

			// The newly added source is the last one in the list
			newSource := cfg.Sources.Sources[len(cfg.Sources.Sources)-1]

			fmt.Printf("✓ Registered: %s\n", newSource.Path)
			fmt.Printf("  Name: %s\n", newSource.Name)

			// Scan to show what was found
			registry, err := scanner.ScanDirectory(newSource.Path)
			if err == nil {
				tools := registry.All()
				fmt.Printf("  Found %d tools\n", len(tools))
			}

			fmt.Println()
			fmt.Println("Run 'tctl sync' to rebuild the tool cache.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom name for this source")
	return cmd
}

func init() {
	// Ensure config dir exists on first run
	config.EnsureConfigDir()
}

