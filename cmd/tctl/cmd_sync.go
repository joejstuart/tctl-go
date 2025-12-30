package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/scanner"
)

func syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Rescan all sources and validate tools",
		Long: `Scans all registered source directories and validates tools.
Run this after adding or modifying tool files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			paths := cfg.SourcePaths()
			if len(paths) == 0 {
				fmt.Println("No sources registered.")
				fmt.Println("Register a directory with: tctl add <path>")
				return nil
			}

			fmt.Printf("[sync] Scanning %d sources...\n", len(paths))

			registry, err := scanner.ScanDirectories(paths)
			if err != nil {
				return err
			}

			tools := registry.All()
			fmt.Printf("[sync] Found %d tools\n", len(tools))

			// Validate
			fmt.Println("[sync] Validating...")
			hasErrors := false
			for _, t := range tools {
				if t.Name == "" {
					fmt.Printf("  ⚠ %s: missing @tool tag\n", t.File)
					hasErrors = true
				}
				if len(t.Provides) == 0 {
					fmt.Printf("  ⚠ %s: missing @provides tag\n", t.Name)
				}
			}

			if hasErrors {
				fmt.Println()
				fmt.Println("[sync] ⚠ Some tools have issues. Run 'tctl doctor' for details.")
			} else {
				fmt.Println("[sync] ✓ All tools valid")
			}

			fmt.Println()
			return nil
		},
	}
}
