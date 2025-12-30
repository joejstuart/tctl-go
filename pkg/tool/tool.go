// Package tool defines the core Tool type and metadata structures.
// This is language-agnostic - scanners for each language populate these structs.
package tool

// Tool represents a single tool with its metadata extracted from source.
type Tool struct {
	Name         string            `yaml:"name" json:"name"`
	Version      string            `yaml:"version,omitempty" json:"version,omitempty"`
	File         string            `yaml:"file" json:"file"`
	Language     string            `yaml:"language" json:"language"`
	Description  string            `yaml:"description,omitempty" json:"description,omitempty"`
	Provides     []string          `yaml:"provides,omitempty" json:"provides,omitempty"`
	Requires     []string          `yaml:"requires,omitempty" json:"requires,omitempty"`
	Output       string            `yaml:"output,omitempty" json:"output,omitempty"`
	Freshness    string            `yaml:"freshness,omitempty" json:"freshness,omitempty"`
	Capabilities []string          `yaml:"capabilities,omitempty" json:"capabilities,omitempty"`
	Boundaries   []string          `yaml:"boundaries,omitempty" json:"boundaries,omitempty"`
	Keywords     []string          `yaml:"keywords,omitempty" json:"keywords,omitempty"`
	Interface    map[string]Arg    `yaml:"interface,omitempty" json:"interface,omitempty"`
	Examples     []string          `yaml:"examples,omitempty" json:"examples,omitempty"`
}

// Arg represents a command-line argument in the tool's interface.
type Arg struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Required    bool   `yaml:"required" json:"required"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// Registry holds all discovered tools, indexed by name.
type Registry struct {
	Tools map[string]*Tool `yaml:"tools" json:"tools"`
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		Tools: make(map[string]*Tool),
	}
}

// Add adds a tool to the registry.
func (r *Registry) Add(t *Tool) {
	if t != nil && t.Name != "" {
		r.Tools[t.Name] = t
	}
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) *Tool {
	return r.Tools[name]
}

// FindByProvides finds the tool that provides the given data.
func (r *Registry) FindByProvides(data string) *Tool {
	for _, t := range r.Tools {
		for _, p := range t.Provides {
			if p == data {
				return t
			}
		}
	}
	return nil
}

// All returns all tools as a slice.
func (r *Registry) All() []*Tool {
	tools := make([]*Tool, 0, len(r.Tools))
	for _, t := range r.Tools {
		tools = append(tools, t)
	}
	return tools
}

