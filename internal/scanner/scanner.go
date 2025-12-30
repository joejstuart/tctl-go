// Package scanner provides interfaces and implementations for extracting
// tool metadata from source files in various languages.
package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/yourname/tctl/pkg/tool"
)

// Directories to skip when scanning for tools
var skipDirs = map[string]bool{
	".venv":         true,
	"venv":          true,
	".env":          true,
	"env":           true,
	"node_modules":  true,
	"__pycache__":   true,
	".git":          true,
	".tox":          true,
	".nox":          true,
	".mypy_cache":   true,
	".pytest_cache": true,
	"dist":          true,
	"build":         true,
	".eggs":         true,
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

// Scanner extracts tool metadata from source files.
// Each language implements this interface.
type Scanner interface {
	// Language returns the name of the language this scanner handles.
	Language() string

	// Extensions returns file extensions this scanner handles (e.g., ".py", ".go").
	Extensions() []string

	// CanScan returns true if this scanner can handle the given file.
	CanScan(path string) bool

	// Scan extracts tool metadata from a single file.
	// Returns nil if the file is not a valid tool.
	Scan(path string) (*tool.Tool, error)
}

// registry of all available scanners
var scanners []Scanner

// Register adds a scanner to the registry.
func Register(s Scanner) {
	scanners = append(scanners, s)
}

// GetScanner returns a scanner that can handle the given file, or nil.
func GetScanner(path string) Scanner {
	for _, s := range scanners {
		if s.CanScan(path) {
			return s
		}
	}
	return nil
}

// GetScannerByLanguage returns a scanner for the given language.
func GetScannerByLanguage(lang string) Scanner {
	for _, s := range scanners {
		if s.Language() == lang {
			return s
		}
	}
	return nil
}

// AllScanners returns all registered scanners.
func AllScanners() []Scanner {
	return scanners
}

// SupportedExtensions returns all file extensions that can be scanned.
func SupportedExtensions() []string {
	var exts []string
	for _, s := range scanners {
		exts = append(exts, s.Extensions()...)
	}
	return exts
}

// ScanDirectory scans a directory for tools using all registered scanners.
func ScanDirectory(dir string) (*tool.Registry, error) {
	return ScanDirectories([]string{dir})
}

// ScanDirectories scans multiple directories for tools.
func ScanDirectories(dirs []string) (*tool.Registry, error) {
	registry := tool.NewRegistry()

	exts := SupportedExtensions()
	if len(exts) == 0 {
		return registry, nil
	}

	// Build extension set for quick lookup
	extSet := make(map[string]bool)
	for _, ext := range exts {
		extSet[ext] = true
	}

	for _, dir := range dirs {
		// Check directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

			// Skip private files (starting with _ or .)
			name := info.Name()
			if len(name) > 0 && (name[0] == '_' || name[0] == '.') {
				return nil
			}

			// Check if file has a supported extension
			ext := filepath.Ext(path)
			if !extSet[ext] {
				return nil
			}

			scanner := GetScanner(path)
			if scanner == nil {
				return nil
			}

			t, err := scanner.Scan(path)
			if err != nil {
				return nil
			}
			if t != nil {
				registry.Add(t)
			}

			return nil
		})
	}

	return registry, nil
}

