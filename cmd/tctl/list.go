package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/scanner"
)

func listCmd() *cobra.Command {
	var sourceName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all tools",
		Long: `List all tools from all registered sources.

Examples:
  tctl list                    # All tools
  tctl list --source scripts   # Only from 'scripts' source`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			// Get source paths to scan
			var paths []string
			if sourceName != "" {
				src := cfg.FindSourceByName(sourceName)
				if src == nil {
					return fmt.Errorf("unknown source: %s", sourceName)
				}
				paths = []string{src.Path}
			} else {
				paths = cfg.SourcePaths()
			}

			if len(paths) == 0 {
				fmt.Println("No sources registered.")
				fmt.Println("Register a directory with: tctl add <path>")
				return nil
			}

			registry, err := scanner.ScanDirectories(paths)
			if err != nil {
				return err
			}

			tools := registry.All()
			if len(tools) == 0 {
				fmt.Println("No tools found.")
				return nil
			}

			// Sort by name
			sort.Slice(tools, func(i, j int) bool {
				return tools[i].Name < tools[j].Name
			})

			// Build source name lookup
			sourceNames := make(map[string]string)
			for _, src := range cfg.Sources.Sources {
				sourceNames[src.Path] = src.Name
			}

			fmt.Println()
			fmt.Println("Tools:")

			for _, t := range tools {
				provides := strings.Join(t.Provides, ", ")
				srcName := sourceNames[filepath.Dir(t.File)]
				if srcName == "" {
					srcName = filepath.Base(filepath.Dir(t.File))
				}

				if provides != "" {
					fmt.Printf("  %-24s [%s] → %s\n", t.Name, srcName, provides)
				} else {
					fmt.Printf("  %-24s [%s]\n", t.Name, srcName)
				}

				if t.Output != "" {
					fmt.Printf("  %-24s       %s\n", "", t.Output)
				}
			}

			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVarP(&sourceName, "source", "s", "", "Filter by source name")
	return cmd
}
