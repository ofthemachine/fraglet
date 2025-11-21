package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// localRunner executes commands directly on the host, ignoring container settings
type localRunner struct{}

func (r *localRunner) Name() string {
	return "local"
}

func (r *localRunner) Available() bool {
	// Always available - we're running on the host
	return true
}

func (r *localRunner) Run(ctx context.Context, spec RunSpec) (RunResult, error) {
	// Use RunStreaming and collect results
	streaming, err := r.RunStreaming(ctx, spec)
	if err != nil {
		return RunResult{}, err
	}
	return collectStreamingResults(ctx, streaming)
}

func (r *localRunner) RunStreaming(ctx context.Context, spec RunSpec) (*StreamingResult, error) {
	var cmd *exec.Cmd
	var cleanup func()

	// If entrypoint specified OR command has shebang, write to temp file
	// This matches docker's behavior exactly (just without container mounting)
	if spec.Entrypoint != "" || hasShebang(spec.Command) {
		tempFile, cleanupFn, err := writeTempScript(spec.Command)
		if err != nil {
			return nil, fmt.Errorf("failed to create temp script: %w", err)
		}
		cleanup = cleanupFn
		// Note: cleanup must be called in two places:
		// 1. On error paths before cmd.Start() succeeds (handled below)
		// 2. After command completes in goroutine (handled in Wait goroutine)

		if spec.Entrypoint != "" {
			// Execute via entrypoint (same as docker, just locally)
			// e.g., "python /tmp/script" or "node /tmp/script"
			cmd = exec.CommandContext(ctx, spec.Entrypoint, tempFile)
		} else {
			// Shebang present but no entrypoint - execute script directly (shebang will be honored)
			cmd = exec.CommandContext(ctx, tempFile)
		}
	} else {
		// No entrypoint and no shebang - use sh -c
		cmd = exec.CommandContext(ctx, "sh", "-c", spec.Command)
	}

	stdoutChan := make(chan string, 10)
	stderrChan := make(chan string, 10)
	doneChan := make(chan error, 1)
	exitCodeChan := make(chan int, 1)

	if spec.Stdin != "" {
		cmd.Stdin = bytes.NewBufferString(spec.Stdin)
	}

	if spec.Env != nil && len(spec.Env) > 0 {
		// Extend environment - start with current env and add spec.Env
		cmd.Env = append(os.Environ(), spec.Env...)
	}

	if spec.WorkDir != "" {
		cmd.Dir = spec.WorkDir
	}

	// Capture stdout
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Read stdout in background
	go func() {
		defer close(stdoutChan)
		buf := make([]byte, 4096)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				stdoutChan <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// Read stderr in background
	go func() {
		defer close(stderrChan)
		buf := make([]byte, 4096)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				stderrChan <- string(buf[:n])
			}
			if err != nil {
				break
			}
		}
	}()

	// Wait for command completion
	go func() {
		err := cmd.Wait()

		// Cleanup temp file after command completes (if one was created)
		if cleanup != nil {
			cleanup()
		}

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCodeChan <- exitErr.ExitCode()
			} else {
				exitCodeChan <- -1
			}
			doneChan <- err
		} else {
			exitCodeChan <- 0
			doneChan <- nil
		}
		close(exitCodeChan)
		close(doneChan)
	}()

	return &StreamingResult{
		Stdout:   stdoutChan,
		Stderr:   stderrChan,
		Done:     doneChan,
		ExitCode: exitCodeChan,
	}, nil
}
