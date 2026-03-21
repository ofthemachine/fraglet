package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

// Executor handles executing code files
type Executor struct {
	cfg *fraglet.EntrypointConfig
}

// NewExecutor creates a new executor
func NewExecutor(cfg *fraglet.EntrypointConfig) *Executor {
	return &Executor{
		cfg: cfg,
	}
}

// Execute finds and executes the code file
// Returns the exit code from the executed command, or 1 if there was an error
func (e *Executor) Execute() (int, error) {
	return e.executeWithArgs(nil)
}

// ExecuteWithArgs executes the code file and passes remaining args
// Returns the exit code from the executed command, or 1 if there was an error
func (e *Executor) ExecuteWithArgs(args []string) (int, error) {
	return e.executeWithArgs(args)
}

// executeWithArgs is the shared implementation
func (e *Executor) executeWithArgs(args []string) (int, error) {
	// Determine the command and arguments to execute
	var cmdPath string
	var cmdArgs []string

	if e.cfg.Execution != nil && e.cfg.Execution.Path != "" {
		// Parse execution path - split by spaces to handle "python3 /path/to/script.py"
		pathParts := strings.Fields(e.cfg.Execution.Path)
		if len(pathParts) == 0 {
			return 1, fmt.Errorf("execution path is empty")
		}
		cmdPath = pathParts[0]
		cmdArgs = pathParts[1:]
		// Append any additional args from command line
		cmdArgs = append(cmdArgs, args...)

		// Implicitly make executable
		if len(pathParts) > 1 {
			// If there are multiple parts, the second part is likely the file
			// e.g. "python /path/to/script.py"
			_ = e.makeExecutable(pathParts[1])
		} else {
			// Single part, treat as file path
			_ = e.makeExecutable(cmdPath)
		}
	} else if len(args) > 0 {
		// No execution path configured, use args as command (shift: args[0] is command, args[1:] are arguments)
		cmdPath = args[0]
		cmdArgs = args[1:]
	} else {
		return 1, fmt.Errorf("no execution path configured and no args provided")
	}

	// Execute the command with arguments
	cmd := exec.Command(cmdPath, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		// Extract exit code from exec.ExitError if available
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}
		return 1, err
	}
	return 0, nil
}

// makeExecutable makes a file executable. Errors are ignored as the path might
// not be a local file (e.g. a command on PATH).
func (e *Executor) makeExecutable(file string) error {
	return os.Chmod(file, 0755)
}
