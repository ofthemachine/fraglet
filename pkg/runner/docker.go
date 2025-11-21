package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// dockerRunner executes commands inside Docker containers
type dockerRunner struct{}

func (r *dockerRunner) Name() string {
	return "docker"
}

func (r *dockerRunner) Available() bool {
	// Check if docker is available
	cmd := exec.Command("docker", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (r *dockerRunner) Run(ctx context.Context, spec RunSpec) (RunResult, error) {
	// Use RunStreaming and collect results
	streaming, err := r.RunStreaming(ctx, spec)
	if err != nil {
		return RunResult{}, err
	}
	return collectStreamingResults(ctx, streaming)
}

func (r *dockerRunner) RunStreaming(ctx context.Context, spec RunSpec) (*StreamingResult, error) {
	if spec.Container == "" {
		return nil, fmt.Errorf("docker runner requires container image")
	}

	stdoutChan := make(chan string, 10)
	stderrChan := make(chan string, 10)
	doneChan := make(chan error, 1)
	exitCodeChan := make(chan int, 1)

	var dockerCmd *exec.Cmd
	var tempFile string
	var cleanup func()
	var cmdArgs []string

	// Start building docker run command
	cmdArgs = append(cmdArgs, "docker", "run", "--rm", "-i")

	// If volumes are specified and command is empty, use container's default entrypoint
	// This is for fragment injection patterns where the container handles execution
	if len(spec.Volumes) > 0 && spec.Command == "" && spec.Entrypoint == "" {
		// Add volumes first (before container name)
		for _, vol := range spec.Volumes {
			mountSpec := fmt.Sprintf("%s:%s", vol.HostPath, vol.ContainerPath)
			if vol.ReadOnly {
				mountSpec += ":ro"
			}
			cmdArgs = append(cmdArgs, "-v", mountSpec)
		}
		// Add env vars
		for _, env := range spec.Env {
			cmdArgs = append(cmdArgs, "-e", env)
		}
		// Add workdir
		if spec.WorkDir != "" {
			cmdArgs = append(cmdArgs, "-w", spec.WorkDir)
		}
		// Add container name at the end
		cmdArgs = append(cmdArgs, spec.Container)
		dockerCmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	} else if spec.Entrypoint != "" || hasShebang(spec.Command) {
		// If entrypoint specified OR command has shebang, write to temp file
		var err error
		tempFile, cleanup, err = writeTempScript(spec.Command)
		if err != nil {
			return nil, fmt.Errorf("failed to create temp script: %w", err)
		}
		// Note: cleanup must be called in two places:
		// 1. On error paths before dockerCmd.Start() succeeds (handled below)
		// 2. After command completes in goroutine (handled in Wait goroutine)

		if spec.Entrypoint != "" {
			// Mount temp file into container and execute via entrypoint
			// Use --entrypoint flag to override container's default entrypoint
			// Use /tmp/script (generic name, no extension) - entrypoint determines how to execute
			dockerCmd = exec.CommandContext(ctx, "docker", "run", "--rm", "-i",
				"--entrypoint", spec.Entrypoint,
				"-v", fmt.Sprintf("%s:/tmp/script:ro", tempFile),
				spec.Container,
				"/tmp/script")
		} else {
			// Shebang present but no entrypoint - execute script directly (shebang will be honored)
			dockerCmd = exec.CommandContext(ctx, "docker", "run", "--rm", "-i",
				"-v", fmt.Sprintf("%s:/tmp/script:ro", tempFile),
				spec.Container,
				"/tmp/script")
		}
	} else {
		// No entrypoint and no shebang - use sh -c
		dockerCmd = exec.CommandContext(ctx, "docker", "run", "--rm", "-i",
			spec.Container,
			"sh", "-c", spec.Command)
	}

	if spec.Stdin != "" {
		dockerCmd.Stdin = bytes.NewBufferString(spec.Stdin)
	}

	if spec.Env != nil {
		for _, env := range spec.Env {
			dockerCmd.Args = append(dockerCmd.Args, "-e", env)
		}
	}

	if spec.WorkDir != "" {
		dockerCmd.Args = append(dockerCmd.Args, "-w", spec.WorkDir)
	}

	// Add volume mounts (only if not already added in fragment injection case)
	if !(len(spec.Volumes) > 0 && spec.Command == "" && spec.Entrypoint == "") {
		for _, vol := range spec.Volumes {
			mountSpec := fmt.Sprintf("%s:%s", vol.HostPath, vol.ContainerPath)
			if vol.ReadOnly {
				mountSpec += ":ro"
			}
			dockerCmd.Args = append(dockerCmd.Args, "-v", mountSpec)
		}
	}

	// Capture stdout
	stdoutPipe, err := dockerCmd.StdoutPipe()
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr
	stderrPipe, err := dockerCmd.StderrPipe()
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := dockerCmd.Start(); err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, fmt.Errorf("failed to start docker command: %w", err)
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
				if err != io.EOF {
					// Log error but continue
				}
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
				if err != io.EOF {
					// Log error but continue
				}
				break
			}
		}
	}()

	// Wait for command completion
	go func() {
		err := dockerCmd.Wait()

		// Cleanup temp file after command completes (if one was created)
		// This matches local runner behavior exactly
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
