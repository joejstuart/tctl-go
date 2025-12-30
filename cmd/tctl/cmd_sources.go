package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/scanner"
)

func sourcesCmd() *cobra.Command {
	var showTools bool

	cmd := &cobra.Command{
		Use:   "sources",
		Short: "List registered tool directories",
		Long: `Show all directories registered with tctl.

Examples:
  tctl sources           # List all sources
  tctl sources --tools   # Include tool counts`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if len(cfg.Sources.Sources) == 0 {
				fmt.Println("No sources registered.")
				fmt.Println()
				fmt.Println("Register a directory with:")
				fmt.Println("  tctl add <path>")
				return nil
			}

			fmt.Println()
			fmt.Println("Registered sources:")
			fmt.Println()

			for _, src := range cfg.Sources.Sources {
				// Check if path exists
				exists := "✓"
				if _, err := os.Stat(src.Path); os.IsNotExist(err) {
					exists = "✗"
				}

				name := src.Name
				if name == "" {
					name = "(unnamed)"
				}

				fmt.Printf("  %s %-16s %s\n", exists, name, src.Path)

				if showTools {
					registry, err := scanner.ScanDirectory(src.Path)
					if err == nil {
						tools := registry.All()
						for _, t := range tools {
							provides := ""
							if len(t.Provides) > 0 {
								provides = " → " + t.Provides[0]
							}
							fmt.Printf("      • %s%s\n", t.Name, provides)
						}
					}
				}
			}

			fmt.Println()
			fmt.Printf("Config: %s\n", config.ConfigDir())
			return nil
		},
	}

	cmd.Flags().BoolVarP(&showTools, "tools", "t", false, "Show tools in each source")
	return cmd
}

