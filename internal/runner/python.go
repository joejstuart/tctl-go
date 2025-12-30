package runner

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/yourname/tctl/pkg/tool"
)

func init() {
	Register(&PythonRunner{})
}

// PythonRunner executes Python tools.
type PythonRunner struct {
	// PythonPath is the path to the Python interpreter.
	// If empty, uses "python3" or "python" from PATH.
	PythonPath string
}

func (r *PythonRunner) Language() string {
	return "python"
}

func (r *PythonRunner) CanRun(t *tool.Tool) bool {
	return t.Language == "python" || filepath.Ext(t.File) == ".py"
}

func (r *PythonRunner) Run(t *tool.Tool, args []string) (int, error) {
	pythonPath := r.findPython()
	if pythonPath == "" {
		return 1, &PythonNotFoundError{}
	}

	// Build command: python /path/to/tool.py args...
	cmdArgs := append([]string{t.File}, args...)
	return execCommand(pythonPath, cmdArgs...)
}

// findPython locates the Python interpreter.
func (r *PythonRunner) findPython() string {
	if r.PythonPath != "" {
		return r.PythonPath
	}

	// Check for uv (fast Python runner)
	if uvPath, err := exec.LookPath("uv"); err == nil {
		// Check if we're in a project with pyproject.toml
		if _, err := os.Stat("pyproject.toml"); err == nil {
			// Use "uv run python" for better venv handling
			return uvPath
		}
	}

	// Try python3 first
	if path, err := exec.LookPath("python3"); err == nil {
		return path
	}

	// Fall back to python
	if path, err := exec.LookPath("python"); err == nil {
		return path
	}

	return ""
}

// RunWithUV runs a Python tool using uv if available.
func (r *PythonRunner) RunWithUV(t *tool.Tool, args []string) (int, error) {
	uvPath, err := exec.LookPath("uv")
	if err != nil {
		// Fall back to regular Python
		return r.Run(t, args)
	}

	// uv run python /path/to/tool.py args...
	cmdArgs := append([]string{"run", "python", t.File}, args...)
	return execCommand(uvPath, cmdArgs...)
}

// PythonNotFoundError is returned when Python is not found.
type PythonNotFoundError struct{}

func (e *PythonNotFoundError) Error() string {
	return "python interpreter not found"
}

