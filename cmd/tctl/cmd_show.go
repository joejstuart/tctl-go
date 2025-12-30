package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/scanner"
	"github.com/yourname/tctl/pkg/tool"
)

func showCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <tool-name>",
		Short: "Show detailed information about a tool",
		Long: `Displays all metadata extracted from a tool's docstring:
  - Capabilities and boundaries
  - Input/output specifications
  - Interface arguments
  - Usage examples`,
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

			registry, err := scanner.ScanDirectories(paths)
			if err != nil {
				return err
			}

			toolName := args[0]
			t := registry.Get(toolName)
			if t == nil {
				fmt.Printf("Unknown tool: %s\n", toolName)
				fmt.Println("Run 'tctl list' to see available tools.")
				return nil
			}

			printToolDetails(t)
			return nil
		},
	}
}

func printToolDetails(t *tool.Tool) {
	fmt.Println()
	fmt.Printf("# %s\n", t.Name)
	fmt.Println()

	if t.Description != "" {
		fmt.Printf("  %s\n", t.Description)
		fmt.Println()
	}

	fmt.Printf("  File: %s\n", t.File)
	fmt.Printf("  Language: %s\n", t.Language)

	if t.Version != "" {
		fmt.Printf("  Version: %s\n", t.Version)
	}

	fmt.Printf("  Provides: %s\n", strings.Join(t.Provides, ", "))
	if len(t.Requires) > 0 {
		fmt.Printf("  Requires: %s\n", strings.Join(t.Requires, ", "))
	}
	fmt.Printf("  Output: %s\n", t.Output)
	fmt.Printf("  Freshness: %s\n", t.Freshness)

	if len(t.Capabilities) > 0 {
		fmt.Println()
		fmt.Println("  Capabilities:")
		for _, cap := range t.Capabilities {
			fmt.Printf("    • %s\n", cap)
		}
	}

	if len(t.Boundaries) > 0 {
		fmt.Println()
		fmt.Println("  Boundaries:")
		for _, b := range t.Boundaries {
			fmt.Printf("    ✗ %s\n", b)
		}
	}

	if len(t.Keywords) > 0 {
		fmt.Println()
		fmt.Printf("  Keywords: %s\n", strings.Join(t.Keywords, ", "))
	}

	if len(t.Interface) > 0 {
		fmt.Println()
		fmt.Println("  Interface:")
		for name, arg := range t.Interface {
			req := ""
			if arg.Required {
				req = " (required)"
			}
			fmt.Printf("    %s: %s%s\n", name, arg.Type, req)
			if arg.Description != "" {
				fmt.Printf("      %s\n", arg.Description)
			}
		}
	}

	if len(t.Examples) > 0 {
		fmt.Println()
		fmt.Println("  Examples:")
		for _, ex := range t.Examples {
			fmt.Printf("    $ %s\n", ex)
		}
	}

	fmt.Println()
}
