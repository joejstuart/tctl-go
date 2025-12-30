package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/runner"
	"github.com/yourname/tctl/internal/scanner"
)

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <tool-name> [args...]",
		Short: "Run a tool directly with arguments",
		Long: `Execute a tool by name, passing any additional arguments.

Examples:
  tctl run fetch-prices --symbols AAPL,GOOGL
  tctl run scrape-gpu --help`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
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

			toolName := args[0]
			toolArgs := args[1:]

			registry, err := scanner.ScanDirectories(paths)
			if err != nil {
				return err
			}

			tool := registry.Get(toolName)
			if tool == nil {
				fmt.Fprintf(os.Stderr, "[tctl] ✗ Unknown tool: %s\n", toolName)
				fmt.Fprintln(os.Stderr, "Run 'tctl list' to see available tools.")
				os.Exit(1)
			}

			fmt.Printf("[tctl] running: %s\n", toolName)

			exitCode, err := runner.Run(tool, toolArgs)
			if err != nil {
				return err
			}

			os.Exit(exitCode)
			return nil
		},
	}
}
