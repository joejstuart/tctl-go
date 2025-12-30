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

func whereCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "where <feature>",
		Short: "Suggest where a feature should go",
		Long: `Analyzes existing tools to suggest where a new feature belongs.
Searches tool names, descriptions, capabilities, and keywords.

Examples:
  tctl where "jira summary"    # Where should jira summaries go?
  tctl where "parse logs"      # Which tool handles log parsing?`,
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

			feature := strings.Join(args, " ")
			tools := registry.All()

			matches, excluded := analyzeFeaturePlacement(tools, feature)

			fmt.Println()
			fmt.Printf("# Where should '%s' go?\n", feature)
			fmt.Println()

			if len(matches) > 0 {
				// Sort by score
				sort.Slice(matches, func(i, j int) bool {
					return matches[i].score > matches[j].score
				})

				fmt.Println("## Best matches\n")
				for i, m := range matches {
					if i >= 5 {
						break
					}
					printWhereMatch(m)
				}
			}

			if len(excluded) > 0 {
				fmt.Println("## Explicitly excluded\n")
				fmt.Println("These tools have @boundary tags that exclude this feature:\n")
				for i, e := range excluded {
					if i >= 3 {
						break
					}
					fmt.Printf("- **%s**: %s\n", e.tool.Name, e.reason)
				}
				fmt.Println()
			}

			if len(matches) == 0 {
				featureWords := strings.Fields(strings.ToLower(feature))
				suggestedName := strings.Join(featureWords[:min(3, len(featureWords))], "-")
				fmt.Println("No existing tool matches this feature.\n")
				fmt.Println("Create a new tool:")
				fmt.Printf("```bash\ntctl new %s\n```\n", suggestedName)
			}

			return nil
		},
	}
}

type featureMatch struct {
	tool    *tool.Tool
	reasons []string
	reason  string
	score   int
}

func analyzeFeaturePlacement(tools []*tool.Tool, feature string) (matches, excluded []featureMatch) {
	terms := strings.Fields(strings.ToLower(feature))

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
				reasons = append(reasons, fmt.Sprintf("description mentions '%s'", term))
			}
		}

		// Check provides
		for _, p := range t.Provides {
			pLower := strings.ToLower(p)
			for _, term := range terms {
				if strings.Contains(pLower, term) {
					score += 6
					reasons = append(reasons, fmt.Sprintf("provides '%s'", p))
				}
			}
		}

		// Check capabilities (good indicator for feature placement)
		for _, cap := range t.Capabilities {
			capLower := strings.ToLower(cap)
			for _, term := range terms {
				if strings.Contains(capLower, term) {
					score += 4
					reasons = append(reasons, fmt.Sprintf("capability: %s", cap))
				}
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

		if score > 0 {
			matches = append(matches, featureMatch{
				tool:    t,
				reasons: reasons,
				score:   score,
			})
		}

		// Check boundaries (negative match)
		for _, boundary := range t.Boundaries {
			boundaryLower := strings.ToLower(boundary)
			for _, term := range terms {
				if strings.Contains(boundaryLower, term) {
					excluded = append(excluded, featureMatch{
						tool:   t,
						reason: boundary,
					})
					break
				}
			}
		}
	}

	return matches, excluded
}

func printWhereMatch(m featureMatch) {
	t := m.tool

	fmt.Printf("### %s\n", t.Name)
	if t.Description != "" {
		fmt.Printf("%s\n", t.Description)
	}
	fmt.Println()

	fmt.Printf("**File:** `%s`\n", t.File)

	if len(t.Provides) > 0 {
		fmt.Printf("**Provides:** %s\n", strings.Join(t.Provides, ", "))
	}

	if len(m.reasons) > 0 {
		fmt.Println()
		fmt.Println("**Why this tool:**")
		// Deduplicate reasons
		seen := make(map[string]bool)
		for _, r := range m.reasons {
			if !seen[r] {
				seen[r] = true
				fmt.Printf("- %s\n", r)
			}
		}
	}

	if len(t.Capabilities) > 0 {
		fmt.Println()
		fmt.Println("**Existing capabilities:**")
		for _, cap := range t.Capabilities {
			fmt.Printf("- %s\n", cap)
		}
	}

	fmt.Println()
	fmt.Println("---")
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
