package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
)

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <path-or-name>",
		Short: "Unregister a tool directory",
		Long: `Remove a directory from the tctl registry.
You can specify either the path or the source name.

Examples:
  tctl remove ~/scripts
  tctl remove scripts       # By name
  tctl remove .             # Current directory`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathOrName := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if err := cfg.RemoveSource(pathOrName); err != nil {
				return err
			}

			fmt.Printf("✓ Removed: %s\n", pathOrName)
			fmt.Println()
			fmt.Println("Run 'tctl sync' to rebuild the tool cache.")
			return nil
		},
	}
}

