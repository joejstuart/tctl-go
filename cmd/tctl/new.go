package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "new <tool-name>",
		Short: "Create a new tool from template",
		Long: `Create a new tool file with a template docstring.
By default, creates in current directory.

Examples:
  tctl new my-scraper              # Creates ./my_scraper.py
  tctl new my-scraper -o ~/tools   # Creates ~/tools/my_scraper.py`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			toolName := args[0]

			// Determine output directory
			dir := outputDir
			if dir == "" {
				var err error
				dir, err = os.Getwd()
				if err != nil {
					return err
				}
			}

			// Create file
			fileName := strings.ReplaceAll(toolName, "-", "_") + ".py"
			filePath := filepath.Join(dir, fileName)

			if _, err := os.Stat(filePath); err == nil {
				return fmt.Errorf("file already exists: %s", filePath)
			}

			content := fmt.Sprintf(pythonToolTemplate, fileName, toolName, toolName, toolName)
			if err := os.WriteFile(filePath, []byte(content), 0755); err != nil {
				return err
			}

			fmt.Printf("✓ Created: %s\n", filePath)
			fmt.Println()
			fmt.Println("Next steps:")
			fmt.Printf("  1. Edit %s - fill in @tags\n", filePath)
			fmt.Printf("  2. Register the directory: tctl add %s\n", dir)
			fmt.Println("  3. Validate: tctl doctor")
			fmt.Printf("  4. Run: tctl run %s --help\n", toolName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory")
	return cmd
}

const pythonToolTemplate = `#!/usr/bin/env python3
"""
%s

TODO: One-line description of what this tool does.

@tool %s
@version 0.1.0
@provides TODO-data-name
@requires
@output data/TODO-output.csv
@freshness daily

@capability TODO: Describe what this tool does
@capability TODO: Add more capabilities

@boundary TODO: Does NOT do X (use other-tool for that)

@keywords TODO, add, search, terms

@interface
  --out: file, required - Output file path

@example tctl run %s --out data/output.csv
"""

import argparse
import sys


def main():
    ap = argparse.ArgumentParser(description="TODO: Description")
    ap.add_argument("--out", required=True, help="Output file path")
    args = ap.parse_args()

    # TODO: Implement tool logic
    print(f"TODO: Implement %s")
    print(f"Output would go to: {args.out}")


if __name__ == "__main__":
    main()
`
