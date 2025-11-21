package executor

import (
	"fmt"
	"os"
	"os/exec"

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
		// Use configured execution path with all args
		cmdPath = e.cfg.Execution.Path
		cmdArgs = args

		// Make executable if needed
		if e.cfg.Execution.ShouldMakeExecutable() {
			if err := e.makeExecutable(cmdPath); err != nil {
				return 1, err
			}
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

// makeExecutable makes a file executable
func (e *Executor) makeExecutable(file string) error {
	if err := os.Chmod(file, 0755); err != nil {
		return fmt.Errorf("failed to make %s executable: %w", file, err)
	}
	return nil
}
