// Package config handles global tctl configuration.
// Config is stored in ~/.config/tctl/ (or $XDG_CONFIG_HOME/tctl/).
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	ConfigDirName  = "tctl"
	SourcesFile    = "sources.yaml"
	CacheFile      = "cache.yaml"
	SettingsFile   = "settings.yaml"
)

// Source represents a registered tool directory.
type Source struct {
	Path    string    `yaml:"path"`
	Name    string    `yaml:"name,omitempty"`
	Added   time.Time `yaml:"added"`
}

// Sources holds all registered tool directories.
type Sources struct {
	Sources []Source `yaml:"sources"`
}

// Settings holds global tctl settings.
type Settings struct {
	DefaultLanguage string `yaml:"default_language,omitempty"`
}

// Intent represents a named workflow.
type Intent struct {
	Description string   `yaml:"description,omitempty"`
	Includes    []string `yaml:"includes,omitempty"`
}

// Intents holds all defined intents (loaded from any source).
type Intents struct {
	Intents map[string]Intent `yaml:"intents,omitempty"`
}

// Global represents the global tctl configuration.
type Global struct {
	ConfigDir string
	Sources   *Sources
	Settings  *Settings
	Intents   *Intents
}

// ConfigDir returns the tctl config directory path.
// Uses $XDG_CONFIG_HOME/tctl or ~/.config/tctl.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, ConfigDirName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".config", ConfigDirName)
	}
	return filepath.Join(home, ".config", ConfigDirName)
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	dir := ConfigDir()
	return os.MkdirAll(dir, 0755)
}

// Load loads the global configuration.
func Load() (*Global, error) {
	dir := ConfigDir()

	g := &Global{
		ConfigDir: dir,
		Sources:   &Sources{Sources: []Source{}},
		Settings:  &Settings{DefaultLanguage: "python"},
		Intents:   &Intents{Intents: make(map[string]Intent)},
	}

	// Load sources
	sourcesPath := filepath.Join(dir, SourcesFile)
	if data, err := os.ReadFile(sourcesPath); err == nil {
		yaml.Unmarshal(data, g.Sources)
	}

	// Load settings
	settingsPath := filepath.Join(dir, SettingsFile)
	if data, err := os.ReadFile(settingsPath); err == nil {
		yaml.Unmarshal(data, g.Settings)
	}

	// Load intents from all sources that have state.yaml
	for _, src := range g.Sources.Sources {
		statePath := filepath.Join(filepath.Dir(src.Path), "state.yaml")
		if data, err := os.ReadFile(statePath); err == nil {
			var srcIntents Intents
			if yaml.Unmarshal(data, &srcIntents) == nil {
				for name, intent := range srcIntents.Intents {
					g.Intents.Intents[name] = intent
				}
			}
		}
	}

	return g, nil
}

// Save saves the sources configuration.
func (g *Global) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return err
	}

	// Save sources
	sourcesPath := filepath.Join(g.ConfigDir, SourcesFile)
	data, err := yaml.Marshal(g.Sources)
	if err != nil {
		return err
	}
	return os.WriteFile(sourcesPath, data, 0644)
}

// AddSource adds a new source directory.
func (g *Global) AddSource(path, name string) error {
	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Check it exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Check if already registered
	for _, src := range g.Sources.Sources {
		if src.Path == absPath {
			return fmt.Errorf("already registered: %s", absPath)
		}
	}

	// Auto-generate name if not provided
	if name == "" {
		name = filepath.Base(absPath)
	}

	g.Sources.Sources = append(g.Sources.Sources, Source{
		Path:  absPath,
		Name:  name,
		Added: time.Now(),
	})

	return g.Save()
}

// RemoveSource removes a source directory.
func (g *Global) RemoveSource(pathOrName string) error {
	absPath, _ := filepath.Abs(pathOrName)

	var newSources []Source
	found := false

	for _, src := range g.Sources.Sources {
		if src.Path == absPath || src.Name == pathOrName {
			found = true
			continue
		}
		newSources = append(newSources, src)
	}

	if !found {
		return fmt.Errorf("not registered: %s", pathOrName)
	}

	g.Sources.Sources = newSources
	return g.Save()
}

// SourcePaths returns all registered source paths.
func (g *Global) SourcePaths() []string {
	paths := make([]string, len(g.Sources.Sources))
	for i, src := range g.Sources.Sources {
		paths[i] = src.Path
	}
	return paths
}

// FindSourceByName finds a source by its name.
func (g *Global) FindSourceByName(name string) *Source {
	for i := range g.Sources.Sources {
		if g.Sources.Sources[i].Name == name {
			return &g.Sources.Sources[i]
		}
	}
	return nil
}

// GetIntent returns an intent by name.
func (g *Global) GetIntent(name string) (Intent, bool) {
	intent, ok := g.Intents.Intents[name]
	return intent, ok
}
