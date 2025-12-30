package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/scanner"
	"github.com/yourname/tctl/internal/util"
	"github.com/yourname/tctl/pkg/tool"
)

func whatCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "what",
		Short: "Show what's available (dynamic from tool metadata)",
		Long: `Scans all registered sources and displays:
  - Available data (what you can 'tctl get')
  - Common keywords for searching`,
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

			registry, err := scanner.ScanDirectories(paths)
			if err != nil {
				return err
			}

			tools := registry.All()
			if len(tools) == 0 {
				fmt.Println("No tools found.")
				return nil
			}

			// Print available data
			fmt.Println()
			fmt.Println("📊 DATA AVAILABLE:")
			fmt.Println()

			for _, t := range tools {
				for _, p := range t.Provides {
					fmt.Printf("  %-24s → tctl get %s\n", p, p)
				}
			}

			// Build and print keywords
			keywordMap := buildKeywordMap(tools)
			printKeywords(keywordMap)

			fmt.Println()
			fmt.Println("Run 'tctl find <keyword>' for specific matching")
			fmt.Println()

			return nil
		},
	}
}

// buildKeywordMap builds a map of keywords to tool names.
func buildKeywordMap(tools []*tool.Tool) map[string][]string {
	keywordMap := make(map[string]map[string]bool)

	for _, t := range tools {
		for _, kw := range t.Keywords {
			kw = strings.ToLower(kw)
			if keywordMap[kw] == nil {
				keywordMap[kw] = make(map[string]bool)
			}
			keywordMap[kw][t.Name] = true
		}
		for _, cap := range t.Capabilities {
			for _, word := range util.ExtractKeywords(cap) {
				if keywordMap[word] == nil {
					keywordMap[word] = make(map[string]bool)
				}
				keywordMap[word][t.Name] = true
			}
		}
	}

	// Convert to string slices
	result := make(map[string][]string)
	for kw, toolSet := range keywordMap {
		var toolNames []string
		for name := range toolSet {
			toolNames = append(toolNames, name)
		}
		result[kw] = toolNames
	}
	return result
}

// printKeywords prints the top keywords sorted by frequency.
func printKeywords(keywordMap map[string][]string) {
	type kwCount struct {
		kw    string
		tools []string
	}
	var sorted []kwCount
	for kw, tools := range keywordMap {
		sorted = append(sorted, kwCount{kw, tools})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i].tools) > len(sorted[j].tools)
	})

	fmt.Println()
	fmt.Println("🔍 KEYWORDS:")
	fmt.Println()

	for i, item := range sorted {
		if i >= 15 {
			break
		}
		toolsStr := strings.Join(item.tools[:util.Min(2, len(item.tools))], ", ")
		fmt.Printf("  '%s' → %s\n", item.kw, toolsStr)
	}
}
