package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourname/tctl/internal/config"
	"github.com/yourname/tctl/internal/scanner"
	"github.com/yourname/tctl/pkg/tool"
)

func findCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find <keywords...>",
		Short: "Find tools by keyword",
		Long: `Search for tools matching the given keywords.
Searches tool name, description, keywords, and capabilities.

Examples:
  tctl find logs           # Find log-related tools
  tctl find "error parse"  # Find error parsing tools`,
		Args: cobra.MinimumNArgs(1),
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

			searchTerms := strings.ToLower(strings.Join(args, " "))
			tools := registry.All()

			matches := findToolMatches(tools, searchTerms)

			if len(matches) == 0 {
				fmt.Printf("No tools found matching: %s\n", strings.Join(args, " "))
				fmt.Println()
				fmt.Println("Try:")
				fmt.Println("  tctl what     - See all keywords")
				fmt.Println("  tctl list     - See all tools")
				return nil
			}

			// Sort by score (best matches first)
			sort.Slice(matches, func(i, j int) bool {
				return matches[i].score > matches[j].score
			})

			fmt.Println()
			fmt.Printf("# Tools matching '%s'\n", strings.Join(args, " "))
			fmt.Println()

			for i, m := range matches {
				if i >= 10 {
					fmt.Printf("... and %d more matches\n", len(matches)-10)
					break
				}
				printToolMatch(m)
			}

			return nil
		},
	}
}

type toolMatch struct {
	tool    *tool.Tool
	score   int
	reasons []string
}

func findToolMatches(tools []*tool.Tool, searchTerms string) []toolMatch {
	var matches []toolMatch
	terms := strings.Fields(searchTerms)

	for _, t := range tools {
		var reasons []string
		score := 0

		// Check tool name (highest weight)
		nameLower := strings.ToLower(t.Name)
		for _, term := range terms {
			if strings.Contains(nameLower, term) {
				score += 10
				reasons = append(reasons, fmt.Sprintf("name contains '%s'", term))
			}
		}

		// Check description
		descLower := strings.ToLower(t.Description)
		for _, term := range terms {
			if strings.Contains(descLower, term) {
				score += 5
				reasons = append(reasons, fmt.Sprintf("description contains '%s'", term))
			}
		}

		// Check keywords
		for _, kw := range t.Keywords {
			kwLower := strings.ToLower(kw)
			for _, term := range terms {
				if strings.Contains(kwLower, term) || strings.Contains(term, kwLower) {
					score += 3
					reasons = append(reasons, fmt.Sprintf("keyword '%s'", kw))
				}
			}
		}

		// Check capabilities
		for _, cap := range t.Capabilities {
			capLower := strings.ToLower(cap)
			for _, term := range terms {
				if strings.Contains(capLower, term) {
					score += 4
					reasons = append(reasons, fmt.Sprintf("capability matches '%s'", term))
				}
			}
		}

		// Check provides
		for _, p := range t.Provides {
			pLower := strings.ToLower(p)
			for _, term := range terms {
				if strings.Contains(pLower, term) {
					score += 3
					reasons = append(reasons, fmt.Sprintf("provides '%s'", p))
				}
			}
		}

		if score > 0 {
			matches = append(matches, toolMatch{t, score, reasons})
		}
	}

	return matches
}

func printToolMatch(m toolMatch) {
	t := m.tool

	fmt.Printf("## %s\n", t.Name)
	if t.Description != "" {
		fmt.Printf("%s\n", t.Description)
	}
	fmt.Println()

	fmt.Printf("**File:** `%s`\n", t.File)

	if len(t.Provides) > 0 {
		fmt.Printf("**Provides:** %s\n", strings.Join(t.Provides, ", "))
	}
	if len(t.Requires) > 0 {
		fmt.Printf("**Requires:** %s\n", strings.Join(t.Requires, ", "))
	}
	if t.Output != "" {
		fmt.Printf("**Output:** %s\n", t.Output)
	}

	if len(t.Capabilities) > 0 {
		fmt.Println()
		fmt.Println("**Capabilities:**")
		for _, cap := range t.Capabilities {
			fmt.Printf("- %s\n", cap)
		}
	}

	if len(t.Boundaries) > 0 {
		fmt.Println()
		fmt.Println("**Boundaries:**")
		for _, b := range t.Boundaries {
			fmt.Printf("- %s\n", b)
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()
}
