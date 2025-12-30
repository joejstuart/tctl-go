package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/freshness"
	"github.com/yourname/tctl/internal/runner"
	"github.com/yourname/tctl/internal/scanner"
	"github.com/yourname/tctl/pkg/tool"
)

func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <data>",
		Short: "Ensure data exists, running tools if needed",
		Long: `Ensures that the specified data is up-to-date.
Resolves dependencies, checks freshness, and runs tools if necessary.

Examples:
  tctl get prices        # Ensure prices data exists
  tctl get signals       # Runs fetch-prices first if needed`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			paths := cfg.SourcePaths()
			if len(paths) == 0 {
				fmt.Println("No sources registered.")
				return nil
			}

			target := args[0]
			fmt.Printf("[tctl] ensuring: %s\n", target)

			registry, err := scanner.ScanDirectories(paths)
			if err != nil {
				return err
			}

			visited := make(map[string]bool)
			success := ensureData(target, cfg, registry, visited)
			if success {
				fmt.Println("[tctl] ✓ done")
			} else {
				fmt.Println("[tctl] ✗ failed")
				os.Exit(1)
			}

			return nil
		},
	}
}

func ensureData(target string, cfg *config.Global, registry *tool.Registry, visited map[string]bool) bool {
	if visited[target] {
		return true // Already processed
	}
	visited[target] = true

	// Check if it's an intent
	if intent, ok := cfg.GetIntent(target); ok {
		fmt.Printf("[tctl] intent: %s\n", target)
		for _, item := range intent.Includes {
			if !ensureData(item, cfg, registry, visited) {
				return false
			}
		}
		return true
	}

	// Find tool that provides this data
	t := registry.FindByProvides(target)
	if t == nil {
		fmt.Fprintf(os.Stderr, "[tctl] ✗ Unknown data: %s\n", target)
		fmt.Fprintf(os.Stderr, "       No tool provides '%s'\n", target)
		return false
	}

	// Check freshness
	if t.Output != "" {
		outputPath := t.Output
		if !filepath.IsAbs(outputPath) {
			outputPath = filepath.Join(filepath.Dir(t.File), "..", t.Output)
		}

		fresh, msg := freshness.Check(outputPath, t.Freshness)
		if fresh {
			fmt.Printf("[tctl] ✓ %s: %s\n", target, msg)
			return true
		}
		fmt.Printf("[tctl] → %s: %s, regenerating...\n", target, msg)
	}

	// Ensure dependencies first
	for _, dep := range t.Requires {
		if !ensureData(dep, cfg, registry, visited) {
			return false
		}
	}

	// Run the tool
	exitCode, err := runner.Run(t, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[tctl] ✗ %s: %v\n", t.Name, err)
		return false
	}
	if exitCode != 0 {
		fmt.Fprintf(os.Stderr, "[tctl] ✗ %s failed with code %d\n", t.Name, exitCode)
		return false
	}

	if t.Output != "" {
		fmt.Printf("     → output: %s\n", t.Output)
	}

	return true
}
