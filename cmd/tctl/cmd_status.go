package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/freshness"
	"github.com/yourname/tctl/internal/scanner"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show data freshness status",
		Long: `Displays the freshness status of all data outputs.
Shows which data is fresh, stale, or missing.`,
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

			registry, err := scanner.ScanDirectories(paths)
			if err != nil {
				return err
			}

			tools := registry.All()

			fmt.Println()
			fmt.Println("📊 Data Status")
			fmt.Println()

			hasData := false
			for _, t := range tools {
				if t.Output == "" {
					continue
				}

				hasData = true

				// Output path is relative to the tool's directory
				outputPath := t.Output
				if !filepath.IsAbs(outputPath) {
					outputPath = filepath.Join(filepath.Dir(t.File), "..", t.Output)
				}

				fresh, msg := freshness.Check(outputPath, t.Freshness)

				icon := "✓"
				if !fresh {
					if strings.Contains(msg, "missing") {
						icon = "✗"
					} else {
						icon = "⚠"
					}
				}

				dataName := t.Name
				if len(t.Provides) > 0 {
					dataName = t.Provides[0]
				}

				fmt.Printf("  %s %-24s %s\n", icon, dataName, msg)
			}

			if !hasData {
				fmt.Println("  No tools with @output defined.")
			}

			fmt.Println()
			return nil
		},
	}
}
