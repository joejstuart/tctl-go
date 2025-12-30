// Package runner provides interfaces and implementations for executing
// tools written in various languages.
package runner

import (
	"os"
	"os/exec"

	"github.com/yourname/tctl/pkg/tool"
)

// Runner executes tools in a specific language.
type Runner interface {
	// Language returns the name of the language this runner handles.
	Language() string

	// CanRun returns true if this runner can execute the given tool.
	CanRun(t *tool.Tool) bool

	// Run executes a tool with the given arguments.
	// Returns the exit code.
	Run(t *tool.Tool, args []string) (int, error)
}

// RunResult contains the result of running a tool.
type RunResult struct {
	ExitCode int
	Error    error
}

// registry of all available runners
var runners []Runner

// Register adds a runner to the registry.
func Register(r Runner) {
	runners = append(runners, r)
}

// GetRunner returns a runner that can handle the given tool, or nil.
func GetRunner(t *tool.Tool) Runner {
	for _, r := range runners {
		if r.CanRun(t) {
			return r
		}
	}
	return nil
}

// GetRunnerByLanguage returns a runner for the given language.
func GetRunnerByLanguage(lang string) Runner {
	for _, r := range runners {
		if r.Language() == lang {
			return r
		}
	}
	return nil
}

// Run executes a tool with the given arguments using the appropriate runner.
func Run(t *tool.Tool, args []string) (int, error) {
	runner := GetRunner(t)
	if runner == nil {
		return 1, &UnsupportedLanguageError{Language: t.Language}
	}
	return runner.Run(t, args)
}

// UnsupportedLanguageError is returned when no runner exists for a language.
type UnsupportedLanguageError struct {
	Language string
}

func (e *UnsupportedLanguageError) Error() string {
	return "unsupported language: " + e.Language
}

// execCommand is a helper for running external commands.
// It connects stdin/stdout/stderr to the current terminal.
func execCommand(name string, args ...string) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

