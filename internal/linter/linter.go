// Package linter provides validation for tools and project structure.
package linter

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourname/tctl/internal/scanner"
)

// Level represents the severity of a lint finding.
type Level string

const (
	LevelError   Level = "error"
	LevelWarning Level = "warning"
	LevelInfo    Level = "info"
)

// Message represents a single lint finding.
type Message struct {
	Level   Level
	File    string
	Line    int
	Code    string
	Message string
}

func (m Message) String() string {
	loc := m.File
	if m.Line > 0 {
		loc = fmt.Sprintf("%s:%d", m.File, m.Line)
	}
	return fmt.Sprintf("[%s] %s: %s", m.Code, loc, m.Message)
}

// Result contains all lint findings.
type Result struct {
	Errors   []Message
	Warnings []Message
	Info     []Message
}

// OK returns true if there are no errors.
func (r *Result) OK() bool {
	return len(r.Errors) == 0
}

// Add adds a finding to the result.
func (r *Result) Add(level Level, file string, line int, code, message string) {
	msg := Message{Level: level, File: file, Line: line, Code: code, Message: message}
	switch level {
	case LevelError:
		r.Errors = append(r.Errors, msg)
	case LevelWarning:
		r.Warnings = append(r.Warnings, msg)
	case LevelInfo:
		r.Info = append(r.Info, msg)
	}
}

// LintProject lints the entire project.
func LintProject(root string) *Result {
	result := &Result{}

	toolsDir := filepath.Join(root, "tools")
	stateFile := filepath.Join(root, "state.yaml")

	// Lint all tool files
	if info, err := os.Stat(toolsDir); err == nil && info.IsDir() {
		filepath.Walk(toolsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if strings.HasPrefix(info.Name(), "_") {
				return nil
			}
			if filepath.Ext(path) == ".py" {
				lintToolFile(path, root, result)
			}
			return nil
		})
	} else {
		result.Add(LevelWarning, "project", 0, "P000", "No tools/ directory found")
	}

	// Lint state.yaml
	if _, err := os.Stat(stateFile); err == nil {
		lintStateFile(stateFile, root, result)
	}

	return result
}

func lintToolFile(path, root string, result *Result) {
	relPath, _ := filepath.Rel(root, path)
	if relPath == "" {
		relPath = filepath.Base(path)
	}

	s := scanner.GetScanner(path)
	if s == nil {
		return
	}

	tool, err := s.Scan(path)
	if err != nil {
		result.Add(LevelError, relPath, 0, "P000", fmt.Sprintf("Could not parse: %v", err))
		return
	}

	if tool == nil {
		result.Add(LevelError, relPath, 1, "D001", "Module missing docstring")
		return
	}

	// T001: Missing @tool tag
	if tool.Name == "" {
		result.Add(LevelError, relPath, 1, "T001", "Missing @tool tag in docstring")
		return
	}

	// T002: Missing @provides
	if len(tool.Provides) == 0 {
		result.Add(LevelWarning, relPath, 0, "T002",
			fmt.Sprintf("%s: Missing @provides tag", tool.Name))
	}

	// T003: Missing @capability
	if len(tool.Capabilities) == 0 {
		result.Add(LevelWarning, relPath, 0, "T003",
			fmt.Sprintf("%s: Missing @capability tags (at least one recommended)", tool.Name))
	}

	// T004: Missing @keywords
	if len(tool.Keywords) == 0 {
		result.Add(LevelWarning, relPath, 0, "T004",
			fmt.Sprintf("%s: Missing @keywords tag (reduces discoverability)", tool.Name))
	}

	// T005: Missing @output when @provides exists
	if len(tool.Provides) > 0 && tool.Output == "" {
		result.Add(LevelWarning, relPath, 0, "T005",
			fmt.Sprintf("%s: Has @provides but no @output path", tool.Name))
	}

	// T006: Missing @boundary (info only)
	if len(tool.Boundaries) == 0 {
		result.Add(LevelInfo, relPath, 0, "T006",
			fmt.Sprintf("%s: No @boundary tags (helps LLM know what tool doesn't do)", tool.Name))
	}

	// T007: Invalid @freshness
	validFreshness := map[string]bool{"daily": true, "weekly": true, "monthly": true, "manual": true}
	if !validFreshness[tool.Freshness] {
		result.Add(LevelError, relPath, 0, "T007",
			fmt.Sprintf("%s: Invalid @freshness '%s'. Must be one of: daily, weekly, monthly, manual",
				tool.Name, tool.Freshness))
	}

	// T010: Missing @example
	if len(tool.Examples) == 0 {
		result.Add(LevelInfo, relPath, 0, "T010",
			fmt.Sprintf("%s: No @example provided", tool.Name))
	}
}

func lintStateFile(path, root string, result *Result) {
	// TODO: Implement state.yaml validation
	// - Check intents reference valid data
	// - Check for circular references
}

// Directories to skip when scanning for tools
var skipDirs = map[string]bool{
	".venv":        true,
	"venv":         true,
	".env":         true,
	"env":          true,
	"node_modules": true,
	"__pycache__":  true,
	".git":         true,
	".tox":         true,
	".nox":         true,
	".mypy_cache":  true,
	".pytest_cache": true,
	"dist":         true,
	"build":        true,
	".eggs":        true,
	"site-packages": true,
}

// shouldSkipDir returns true if the directory should be skipped during scanning.
func shouldSkipDir(name string) bool {
	if skipDirs[name] {
		return true
	}
	// Skip egg-info directories
	if strings.HasSuffix(name, ".egg-info") {
		return true
	}
	return false
}

// LintPath lints a file or directory for tctl compatibility.
// Unlike LintProject, this works on any path and reports what's needed
// to make files tctl-compatible.
func LintPath(path string) *Result {
	result := &Result{}

	info, err := os.Stat(path)
	if err != nil {
		result.Add(LevelError, path, 0, "F001", fmt.Sprintf("Cannot access path: %v", err))
		return result
	}

	if info.IsDir() {
		filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			// Skip excluded directories
			if info.IsDir() {
				if shouldSkipDir(info.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasPrefix(info.Name(), "_") || strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			if filepath.Ext(p) == ".py" {
				lintFileForCompatibility(p, path, result)
			}
			return nil
		})
	} else {
		lintFileForCompatibility(path, filepath.Dir(path), result)
	}

	return result
}

// lintFileForCompatibility checks a file for tctl compatibility,
// including files that have no tctl metadata at all.
func lintFileForCompatibility(path, root string, result *Result) {
	// Use absolute path so LLMs can locate the file
	displayPath, err := filepath.Abs(path)
	if err != nil {
		displayPath = path
	}

	// Check if file has a docstring at all
	hasDocstring, docstringContent := checkPythonDocstring(path)

	if !hasDocstring {
		result.Add(LevelError, displayPath, 1, "D001",
			"No module-level docstring. Add a triple-quoted docstring at the top of the file with @tool <name> tag.")
		return
	}

	// Check for @tool tag
	if !strings.Contains(docstringContent, "@tool ") {
		result.Add(LevelError, displayPath, 1, "T001",
			"Docstring exists but missing @tool tag. Add '@tool <tool-name>' line inside the docstring.")
	}

	// Now try to parse as a tool
	s := scanner.GetScanner(path)
	if s == nil {
		return
	}

	tool, err := s.Scan(path)
	if err != nil {
		result.Add(LevelError, displayPath, 0, "P001", fmt.Sprintf("Parse error: %v", err))
		return
	}

	if tool == nil {
		// Has docstring but scanner returned nil - likely missing @tool
		if strings.Contains(docstringContent, "@tool ") {
			result.Add(LevelError, displayPath, 1, "T001",
				"@tool tag found but could not parse. Check format: @tool <name>")
		}
		return
	}

	// Tool parsed successfully, check for recommended fields
	if len(tool.Provides) == 0 {
		result.Add(LevelWarning, displayPath, 0, "T002",
			fmt.Sprintf("Missing @provides tag. Add: @provides <artifact-name>"))
	}

	if tool.Output == "" {
		result.Add(LevelWarning, displayPath, 0, "T005",
			"Missing @output tag. Add: @output <path-or-description>")
	}

	if len(tool.Keywords) == 0 {
		result.Add(LevelInfo, displayPath, 0, "T004",
			"Missing @keywords tag (improves discoverability). Add: @keywords <word1>, <word2>")
	}

	if tool.Description == "" {
		result.Add(LevelInfo, displayPath, 0, "T008",
			"Missing description. Add a description line after the tool name in the docstring.")
	}

	// Only flag missing @interface if the docstring doesn't contain @interface at all
	// (tools with "no arguments" documentation are valid)
	if len(tool.Interface) == 0 && !strings.Contains(docstringContent, "@interface") {
		result.Add(LevelInfo, displayPath, 0, "T009",
			"No @interface block. Consider documenting CLI arguments if the tool accepts any.")
	}

	// Warn if tool has CLI arguments but no @example
	if len(tool.Interface) > 0 && len(tool.Examples) == 0 {
		result.Add(LevelWarning, displayPath, 0, "T010",
			"Tool has CLI arguments but no @example. Add: @example <command-line-example>")
	}

	// Info if tool has @requires - remind about dependencies
	if len(tool.Requires) > 0 && len(tool.Examples) == 0 {
		result.Add(LevelInfo, displayPath, 0, "T011",
			fmt.Sprintf("Tool requires '%s'. Consider adding @example showing the full workflow.", strings.Join(tool.Requires, ", ")))
	}
}

// checkPythonDocstring checks if a Python file has a module-level docstring.
func checkPythonDocstring(path string) (bool, string) {
	file, err := os.Open(path)
	if err != nil {
		return false, ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	inDocstring := false
	docstringDelim := ""

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip shebang and comments at top
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		if !inDocstring {
			if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
				inDocstring = true
				docstringDelim = trimmed[:3]
				rest := trimmed[3:]
				if strings.Contains(rest, docstringDelim) {
					// Single-line docstring
					return true, strings.TrimSuffix(rest, docstringDelim)
				}
				lines = append(lines, rest)
				continue
			}
			// Non-empty, non-comment line before docstring = no docstring
			if trimmed != "" {
				return false, ""
			}
			continue
		}

		// Inside docstring
		if strings.Contains(line, docstringDelim) {
			idx := strings.Index(line, docstringDelim)
			lines = append(lines, line[:idx])
			return true, strings.Join(lines, "\n")
		}
		lines = append(lines, line)
	}

	return inDocstring, strings.Join(lines, "\n")
}

// FormatResultsForLLM formats lint results in a structured way for LLM consumption.
func FormatResultsForLLM(result *Result, path string) string {
	var sb strings.Builder

	sb.WriteString("# tctl Compatibility Report\n\n")
	sb.WriteString(fmt.Sprintf("Path analyzed: %s\n\n", path))

	totalIssues := len(result.Errors) + len(result.Warnings) + len(result.Info)
	if totalIssues == 0 {
		sb.WriteString("✓ All files are tctl-compatible. No changes needed.\n")
		return sb.String()
	}

	if len(result.Errors) > 0 {
		sb.WriteString("## Required Fixes (Errors)\n\n")
		sb.WriteString("These must be fixed for the tool to work with tctl:\n\n")
		for _, msg := range result.Errors {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`): %s\n", msg.File, msg.Code, msg.Message))
		}
		sb.WriteString("\n")
	}

	if len(result.Warnings) > 0 {
		sb.WriteString("## Recommended Fixes (Warnings)\n\n")
		sb.WriteString("These improve tool functionality:\n\n")
		for _, msg := range result.Warnings {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`): %s\n", msg.File, msg.Code, msg.Message))
		}
		sb.WriteString("\n")
	}

	if len(result.Info) > 0 {
		sb.WriteString("## Suggestions (Info)\n\n")
		sb.WriteString("Optional improvements for better discoverability:\n\n")
		for _, msg := range result.Info {
			sb.WriteString(fmt.Sprintf("- **%s** (`%s`): %s\n", msg.File, msg.Code, msg.Message))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Required Docstring Format\n\n")
	sb.WriteString("Each Python tool file must have a triple-quoted docstring at the very top of the file (after shebang/imports).\n")
	sb.WriteString("The docstring must contain an `@tool <name>` line. Other tags are optional but recommended.\n\n")
	sb.WriteString("**Example:**\n\n")
	sb.WriteString("```python\n")
	sb.WriteString(`#!/usr/bin/env python3
"""
tool-name
Brief description of what this tool does.

@tool tool-name
@provides artifact-name
@output data/output.csv
@keywords keyword1, keyword2

@interface
--arg1: string, required - Description of arg1
--arg2: int, default=10 - Description of arg2
"""

import sys
# ... rest of code
`)
	sb.WriteString("```\n")

	return sb.String()
}
