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

	// Default platform to linux/amd64 unless explicitly set
	platform := spec.Platform
	if platform == "" {
		platform = "linux/amd64"
	}

	// Ensure image exists locally; if not, pull it for the requested platform
	if err := ensureDockerImage(ctx, spec.Container, platform); err != nil {
		return nil, err
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
	cmdArgs = append(cmdArgs, "docker", "run", "--rm", "-i", "--platform", platform)

	// Build docker command based on what's provided
	if len(spec.Volumes) > 0 && spec.Command == "" && spec.Entrypoint == "" {
		// Volumes only: mount and let container's default entrypoint handle execution
		// This is for fraglet-entrypoint containers
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
		// Append script args after container name
		cmdArgs = append(cmdArgs, spec.Args...)
		dockerCmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	} else if spec.Entrypoint != "" {
		// Entrypoint specified: use it to execute command
		if spec.Command != "" {
			// Write command to temp file and mount it
			var err error
			tempFile, cleanup, err = writeTempScript(spec.Command)
			if err != nil {
				return nil, fmt.Errorf("failed to create temp script: %w", err)
			}
			cmdArgs := []string{"docker", "run", "--rm", "-i", "--platform", platform,
				"--entrypoint", spec.Entrypoint,
				"-v", fmt.Sprintf("%s:/tmp/script:ro", tempFile),
				spec.Container,
				"/tmp/script"}
			cmdArgs = append(cmdArgs, spec.Args...)
			dockerCmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		} else {
			// Entrypoint with no command - just use entrypoint
			cmdArgs = append(cmdArgs, "--entrypoint", spec.Entrypoint, spec.Container)
			cmdArgs = append(cmdArgs, spec.Args...)
			dockerCmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		}
	} else if spec.Command != "" {
		// Command specified: execute via sh -c
		dockerCmd = exec.CommandContext(ctx, "docker", "run", "--rm", "-i", "--platform", platform,
			spec.Container,
			"sh", "-c", spec.Command)
		// Args don't make sense with sh -c (they'd be part of the command string)
	} else {
		// Nothing specified - just run container
		cmdArgs = append(cmdArgs, spec.Container)
		cmdArgs = append(cmdArgs, spec.Args...)
		dockerCmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	}

	if spec.StdinReader != nil {
		dockerCmd.Stdin = spec.StdinReader
	} else if spec.Stdin != "" {
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

	// Add volume mounts (only if not already added in volumes-only case above)
	if !(len(spec.Volumes) > 0 && spec.Command == "" && spec.Entrypoint == "") {
		for _, vol := range spec.Volumes {
			mountSpec := fmt.Sprintf("%s:%s", vol.HostPath, vol.ContainerPath)
			if vol.ReadOnly {
				mountSpec += ":ro"
			}
			dockerCmd.Args = append(dockerCmd.Args, "-v", mountSpec)
		}
	}

	if spec.Stdout != nil {
		dockerCmd.Stdout = spec.Stdout
	}
	if spec.Stderr != nil {
		dockerCmd.Stderr = spec.Stderr
	}

	var stdoutPipe, stderrPipe io.ReadCloser
	if spec.Stdout == nil {
		var err error
		stdoutPipe, err = dockerCmd.StdoutPipe()
		if err != nil {
			if cleanup != nil {
				cleanup()
			}
			return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
		}
	}
	if spec.Stderr == nil {
		var err error
		stderrPipe, err = dockerCmd.StderrPipe()
		if err != nil {
			if cleanup != nil {
				cleanup()
			}
			return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
		}
	}

	if err := dockerCmd.Start(); err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, fmt.Errorf("failed to start docker command: %w", err)
	}

	if spec.Stdout == nil {
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
	}
	if spec.Stderr == nil {
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
	}

	go func() {
		err := dockerCmd.Wait()
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
		if spec.Stdout != nil {
			close(stdoutChan)
		}
		if spec.Stderr != nil {
			close(stderrChan)
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

// ensureDockerImage checks if the image exists locally; if not, it pulls it for the given platform.
func ensureDockerImage(ctx context.Context, image, platform string) error {
	inspect := exec.CommandContext(ctx, "docker", "image", "inspect", image)
	if err := inspect.Run(); err == nil {
		return nil // already present
	}
	// Pull with platform
	pull := exec.CommandContext(ctx, "docker", "pull", "--platform", platform, image)
	if out, err := pull.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull image %s: %v\n%s", image, err, string(out))
	}
	return nil
}
