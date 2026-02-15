package runner

import (
	"context"
	"io"
	"os/exec"
	"time"
)

// Runner executes commands in various environments (local, containers, etc.)
type Runner interface {
	// Run executes a command synchronously and returns complete results
	// Blocks until command completes - no incremental feedback
	Run(ctx context.Context, spec RunSpec) (RunResult, error)

	// RunStreaming executes a command with streaming output via channels
	// Returns immediately with channels for stdout/stderr, and a done channel
	// Useful for long-running commands where incremental feedback is needed
	RunStreaming(ctx context.Context, spec RunSpec) (*StreamingResult, error)

	// Name returns the runner type (e.g., "local", "docker")
	Name() string

	// Available returns true if this runner is available on the system
	Available() bool
}

// StreamingResult provides channels for incremental output
type StreamingResult struct {
	Stdout   <-chan string // Channel for stdout chunks
	Stderr   <-chan string // Channel for stderr chunks
	Done     <-chan error  // Channel that closes when command completes, error if non-nil
	ExitCode <-chan int    // Channel that receives exit code when available
}

// VolumeMount defines a volume mount for container execution.
// Writable defaults to false (read-only mount); set true only when the container must write.
type VolumeMount struct {
	HostPath      string // Path on the host
	ContainerPath string // Path inside the container
	Writable      bool   // If true, mount is read-write; default false = read-only (secure by default)
}

// RunSpec defines what to execute
type RunSpec struct {
	Command      string        // The command to execute (rendered template)
	Stdin        string        // Optional stdin input (buffered string, for programmatic use)
	StdinReader  io.Reader     // Optional stdin stream (takes precedence over Stdin)
	Container    string        // Optional container image (e.g., "python:3.11-slim")
	Entrypoint   string        // Optional entrypoint (e.g., "python" for multiline scripts)
	Platform     string        // Optional platform (e.g., linux/amd64). Defaults to linux/amd64.
	Env          []string      // Optional environment variables (for ENVVAR input)
	WorkDir      string        // Optional working directory
	Volumes      []VolumeMount // Optional volume mounts
	Args         []string      // Arguments passed to the command
	Stdout       io.Writer     // If non-nil, command stdout is written here; otherwise captured
	Stderr       io.Writer     // If non-nil, command stderr is written here; otherwise captured
	// Note: Executor field removed - Phase 2 feature when executor registry is designed
}

// RunResult captures execution output
type RunResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// NewRunner creates an appropriate runner based on the spec
// If container is specified, returns a docker runner (if available) or local
// If no container, returns local runner
func NewRunner(container, entrypoint string) Runner {
	if container != "" {
		docker := &dockerRunner{}
		if docker.Available() {
			return docker
		}
		// Docker not available, fall back to local (ignores container)
	}
	return &localRunner{}
}

// collectStreamingResults collects all output from a streaming execution and returns a RunResult
// This is used by Run() implementations to convert RunStreaming() results to RunResult
func collectStreamingResults(ctx context.Context, streaming *StreamingResult) (RunResult, error) {
	start := time.Now()

	var stdout, stderr string
	var exitCode int
	var execErr error

	// Collect stdout chunks
	stdoutDone := make(chan struct{})
	go func() {
		defer close(stdoutDone)
		for chunk := range streaming.Stdout {
			stdout += chunk
		}
	}()

	// Collect stderr chunks
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		for chunk := range streaming.Stderr {
			stderr += chunk
		}
	}()

	// Wait for command to complete and get exit code
	select {
	case err := <-streaming.Done:
		execErr = err
	case <-ctx.Done():
		execErr = ctx.Err()
	}

	// Read exit code
	select {
	case code := <-streaming.ExitCode:
		exitCode = code
	case <-time.After(100 * time.Millisecond):
		// Exit code not available, might be 0 or error case
		if execErr != nil {
			exitCode = -1
		}
	}

	// Wait for output collection to finish
	<-stdoutDone
	<-stderrDone

	result := RunResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
		Duration: time.Since(start),
	}

	// Only return error for actual execution failures, not for non-zero exit codes
	// ExitError means the command ran but exited with non-zero code - that's a valid result
	if execErr != nil {
		// Check if it's an ExitError (non-zero exit code) - in that case, return result with no error
		if _, ok := execErr.(*exec.ExitError); ok {
			return result, nil
		}
		// For other errors (context cancelled, command not found, etc.), return the error
		return result, execErr
	}

	return result, nil
}
