package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yourname/tctl/pkg/tool"
)

func init() {
	Register(&PythonScanner{})
}

// PythonScanner extracts tool metadata from Python docstrings.
type PythonScanner struct{}

func (s *PythonScanner) Language() string {
	return "python"
}

func (s *PythonScanner) Extensions() []string {
	return []string{".py"}
}

func (s *PythonScanner) CanScan(path string) bool {
	return filepath.Ext(path) == ".py"
}

func (s *PythonScanner) Scan(path string) (*tool.Tool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Extract module docstring
	docstring, err := extractPythonDocstring(file)
	if err != nil {
		return nil, err
	}
	if docstring == "" {
		return nil, nil
	}

	// Parse @tags from docstring
	t := parseDocstringTags(docstring)
	if t == nil || t.Name == "" {
		return nil, nil
	}

	t.File = path
	t.Language = "python"

	return t, nil
}

// extractPythonDocstring extracts the module-level docstring from a Python file.
func extractPythonDocstring(file *os.File) (string, error) {
	scanner := bufio.NewScanner(file)
	var lines []string
	inDocstring := false
	docstringDelim := ""

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip shebang and encoding declarations
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Look for docstring start
		if !inDocstring {
			if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
				inDocstring = true
				docstringDelim = trimmed[:3]

				// Check for single-line docstring
				rest := trimmed[3:]
				if strings.Contains(rest, docstringDelim) {
					// Single-line docstring
					return strings.TrimSuffix(rest, docstringDelim), nil
				}
				lines = append(lines, rest)
				continue
			}
			// Not a docstring, probably code
			if trimmed != "" {
				return "", nil
			}
			continue
		}

		// Inside docstring
		if strings.Contains(line, docstringDelim) {
			// End of docstring
			idx := strings.Index(line, docstringDelim)
			lines = append(lines, line[:idx])
			break
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

// parseDocstringTags parses @tags from a docstring into a Tool struct.
func parseDocstringTags(docstring string) *tool.Tool {
	t := &tool.Tool{
		Freshness: "manual",
		Interface: make(map[string]tool.Arg),
	}

	lines := strings.Split(docstring, "\n")
	inInterface := false
	var descLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle @interface block
		if inInterface {
			if strings.HasPrefix(trimmed, "--") {
				arg := parseInterfaceLine(trimmed)
				if arg != nil {
					t.Interface[arg.Name] = *arg
				}
				continue
			} else if strings.HasPrefix(trimmed, "@") {
				inInterface = false
				// Fall through to process this tag
			} else {
				continue
			}
		}

		// Parse @tags
		switch {
		case strings.HasPrefix(trimmed, "@tool "):
			t.Name = strings.TrimSpace(trimmed[6:])

		case strings.HasPrefix(trimmed, "@version "):
			t.Version = strings.TrimSpace(trimmed[9:])

		case strings.HasPrefix(trimmed, "@provides "):
			items := strings.Fields(trimmed[10:])
			t.Provides = append(t.Provides, items...)

		case strings.HasPrefix(trimmed, "@requires "):
			items := strings.Fields(trimmed[10:])
			t.Requires = append(t.Requires, items...)

		case strings.HasPrefix(trimmed, "@output "):
			t.Output = strings.TrimSpace(trimmed[8:])

		case strings.HasPrefix(trimmed, "@freshness "):
			t.Freshness = strings.TrimSpace(trimmed[11:])

		case strings.HasPrefix(trimmed, "@capability "):
			t.Capabilities = append(t.Capabilities, strings.TrimSpace(trimmed[12:]))

		case strings.HasPrefix(trimmed, "@boundary "):
			t.Boundaries = append(t.Boundaries, strings.TrimSpace(trimmed[10:]))

		case strings.HasPrefix(trimmed, "@keywords "):
			keywordsStr := strings.TrimSpace(trimmed[10:])
			// Split by comma or whitespace
			re := regexp.MustCompile(`[,\s]+`)
			keywords := re.Split(keywordsStr, -1)
			for _, kw := range keywords {
				kw = strings.TrimSpace(kw)
				if kw != "" {
					t.Keywords = append(t.Keywords, kw)
				}
			}

		case strings.HasPrefix(trimmed, "@interface"):
			inInterface = true

		case strings.HasPrefix(trimmed, "@example "):
			t.Examples = append(t.Examples, strings.TrimSpace(trimmed[9:]))

		case !strings.HasPrefix(trimmed, "@") && trimmed != "":
			// Collect description lines (before first @tag)
			if t.Name == "" && len(t.Provides) == 0 {
				descLines = append(descLines, trimmed)
			}
		}
	}

	// Set description (skip filename line, use second line)
	if len(descLines) > 1 {
		t.Description = descLines[1]
	} else if len(descLines) == 1 {
		t.Description = descLines[0]
	}

	return t
}

// parseInterfaceLine parses a line like: --arg: type, required - Description
func parseInterfaceLine(line string) *tool.Arg {
	// Pattern: --name: type, modifiers - description
	re := regexp.MustCompile(`^(--[\w-]+):\s*(.+)$`)
	match := re.FindStringSubmatch(strings.TrimSpace(line))
	if match == nil {
		return nil
	}

	name := match[1]
	rest := match[2]

	// Split by " - " to get description
	var specPart, description string
	if idx := strings.Index(rest, " - "); idx != -1 {
		specPart = rest[:idx]
		description = rest[idx+3:]
	} else {
		specPart = rest
	}

	// Parse spec part
	parts := strings.Split(specPart, ",")
	argType := "string"
	required := false
	defaultVal := ""

	for i, part := range parts {
		part = strings.TrimSpace(part)
		if i == 0 {
			argType = part
		} else if part == "required" {
			required = true
		} else if strings.HasPrefix(part, "default=") {
			defaultVal = strings.TrimPrefix(part, "default=")
		}
	}

	return &tool.Arg{
		Name:        name,
		Type:        argType,
		Required:    required,
		Default:     defaultVal,
		Description: strings.TrimSpace(description),
	}
}

