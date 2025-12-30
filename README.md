# tctl

A fast CLI for managing a personal library of tools.

Register tool directories from anywhere, then discover and run tools globally.

## Quick Start

```bash
# Install
go install ./cmd/tctl

# Register a directory containing tools
tctl add ~/my-scripts
tctl add ~/repos/project/tools -n project

# Now from anywhere:
tctl list                    # See all tools
tctl find logs               # Find log-related tools
tctl run analyze-logs        # Run a tool
tctl show analyze-logs       # See tool details
```

## Why tctl?

You have scripts scattered across directories:
- `~/scripts/` - personal utilities
- `~/repos/project1/tools/` - project automation
- `~/repos/project2/scripts/` - another project

**Without tctl:** You forget what exists, grep through files, maintain multiple READMEs.

**With tctl:** Register each directory once, then discover and run tools from anywhere.

```bash
$ tctl sources
  ✓ scripts          ~/scripts
  ✓ project1         ~/repos/project1/tools
  ✓ project2         ~/repos/project2/scripts

$ tctl find logs
  analyze-logs   [project1] - Analyzes server logs for errors
  rotate-logs    [scripts]  - Rotates and compresses old logs

$ tctl run analyze-logs --input /var/log/app.log
```

## Install

```bash
# From source (requires Go 1.21+)
git clone https://github.com/yourname/tctl
cd tctl
go install ./cmd/tctl

# Verify
tctl --version
```

## Commands

### Source Management

| Command | Description |
|---------|-------------|
| `tctl add [path]` | Register a tool directory (default: current dir) |
| `tctl add path -n name` | Register with a custom name |
| `tctl remove <path-or-name>` | Unregister a directory |
| `tctl sources` | List registered directories |

### Tool Discovery

| Command | Description |
|---------|-------------|
| `tctl list` | List all tools from all sources |
| `tctl list -s name` | List tools from one source |
| `tctl what` | Show available data and keywords |
| `tctl find <keyword>` | Find tools by keyword |
| `tctl where "<feature>"` | Suggest where to add a feature |
| `tctl show <tool>` | Show detailed tool information |

### Tool Execution

| Command | Description |
|---------|-------------|
| `tctl run <tool> [args]` | Run a tool with arguments |
| `tctl get <data>` | Ensure data exists (runs dependencies) |

### Maintenance

| Command | Description |
|---------|-------------|
| `tctl new <name>` | Create a new tool from template |
| `tctl new <name> -o dir` | Create in specific directory |
| `tctl sync` | Rescan all sources |
| `tctl lint [path]` | Check tools for compatibility issues |
| `tctl status` | Show data freshness |

## How It Works

Tools are self-describing through metadata tags in their docstrings:

```python
#!/usr/bin/env python3
"""
analyze_logs.py

@tool analyze-logs
@provides log-report
@output data/log_report.json
@freshness daily

@capability Parses server logs and extracts error patterns
@capability Groups errors by frequency and severity

@boundary Does NOT send alerts (use notify-errors for that)

@keywords logs, errors, parsing, analysis

@interface
--input: file, required - Path to log file
--format: string, default=json - Output format (json, csv)

@example tctl run analyze-logs --input /var/log/app.log
"""
```

tctl scans these tags to:
- **Discover** what tools exist and what they do
- **Route** feature requests to the right tool (`tctl where`)
- **Validate** tools have proper metadata (`tctl lint`)
- **Execute** with dependency resolution (`tctl get`)

## Configuration

Config is stored in `~/.config/tctl/` (respects `$XDG_CONFIG_HOME`):

```
~/.config/tctl/
├── sources.yaml     # Registered directories
└── settings.yaml    # Global settings (optional)
```

## For LLMs

When working with an LLM on a codebase:

```bash
# "Where should I add email notification support?"
tctl where "email notifications"

# "What tools handle logs?"  
tctl find logs

# "Show me the log analyzer"
tctl show analyze-logs

# "Check if tools are properly configured"
tctl lint
```

## Adding New Languages

tctl supports tools in any language. Currently implemented:
- **Python** (`.py` files with docstring metadata)

To add a new language, implement the `Scanner` and `Runner` interfaces:

```go
// internal/scanner/golang.go
type GoScanner struct{}

func (s *GoScanner) Language() string { return "go" }
func (s *GoScanner) Extensions() []string { return []string{".go"} }
func (s *GoScanner) CanScan(path string) bool { ... }
func (s *GoScanner) Scan(path string) (*tool.Tool, error) { ... }

// internal/runner/golang.go  
type GoRunner struct{}

func (r *GoRunner) Language() string { return "go" }
func (r *GoRunner) CanRun(t *tool.Tool) bool { ... }
func (r *GoRunner) Run(t *tool.Tool, args []string) (int, error) { ... }
```

## Project Structure

```
tctl/
├── cmd/tctl/               # CLI commands (one file each)
│   ├── main.go
│   ├── cmd_add.go
│   ├── cmd_remove.go
│   ├── cmd_sources.go
│   ├── cmd_lint.go
│   ├── list.go
│   ├── run.go
│   ├── get.go
│   └── ...
├── internal/
│   ├── config/             # Global config (~/.config/tctl/)
│   ├── scanner/            # Language-specific metadata extraction
│   ├── runner/             # Language-specific execution
│   ├── linter/             # Tool validation
│   ├── freshness/          # Data freshness checking
│   └── util/               # Shared utilities
└── pkg/tool/               # Core Tool type
```

## Tag Reference

| Tag | Description | Example |
|-----|-------------|---------|
| `@tool` | Tool name (kebab-case) | `@tool analyze-logs` |
| `@version` | Semantic version | `@version 1.2.0` |
| `@provides` | Data this tool produces | `@provides log-report` |
| `@requires` | Data this tool needs | `@requires raw-logs` |
| `@output` | Output file path | `@output data/report.json` |
| `@freshness` | Refresh policy | `@freshness daily` |
| `@capability` | What this tool does | `@capability Parses server logs` |
| `@boundary` | What it does NOT do | `@boundary Does NOT send alerts` |
| `@keywords` | Search terms | `@keywords logs, parsing` |
| `@interface` | CLI arguments block | See example above |
| `@example` | Usage example | `@example tctl run analyze-logs` |

### Freshness Values

| Value | Stale After |
|-------|-------------|
| `daily` | 1 day |
| `weekly` | 7 days |
| `monthly` | 30 days |
| `manual` | Never (run explicitly) |

## License

MIT
