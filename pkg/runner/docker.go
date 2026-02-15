package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// dockerRunBuilder constructs "docker run ..." argv in a consistent order:
// base (run, --rm, -i, platform, hardening) → opts (volumes, env, workdir, entrypoint) → image → args.
// Add options via methods, then call Build() to get the full []string for exec.
type dockerRunBuilder struct {
	args []string
}

func newDockerRunBuilder(platform string) *dockerRunBuilder {
	return &dockerRunBuilder{
		args: []string{
			"docker", "run", "--rm", "-i", "--platform", platform,
			"--cap-drop=all", "--security-opt=no-new-privileges",
		},
	}
}

// readOnly: true = :ro mount (default for secure-by-default); false = read-write.
func (b *dockerRunBuilder) Volume(hostPath, containerPath string, readOnly bool) *dockerRunBuilder {
	spec := fmt.Sprintf("%s:%s", hostPath, containerPath)
	if readOnly {
		spec += ":ro"
	}
	b.args = append(b.args, "-v", spec)
	return b
}

func (b *dockerRunBuilder) Volumes(volumes []VolumeMount) *dockerRunBuilder {
	for _, vol := range volumes {
		b.Volume(vol.HostPath, vol.ContainerPath, !vol.Writable) // read-only by default
	}
	return b
}

func (b *dockerRunBuilder) Env(env []string) *dockerRunBuilder {
	for _, e := range env {
		b.args = append(b.args, "-e", e)
	}
	return b
}

func (b *dockerRunBuilder) WorkDir(dir string) *dockerRunBuilder {
	if dir != "" {
		b.args = append(b.args, "-w", dir)
	}
	return b
}

func (b *dockerRunBuilder) Entrypoint(ep string) *dockerRunBuilder {
	b.args = append(b.args, "--entrypoint", ep)
	return b
}

func (b *dockerRunBuilder) Image(image string) *dockerRunBuilder {
	b.args = append(b.args, image)
	return b
}

func (b *dockerRunBuilder) Args(a ...string) *dockerRunBuilder {
	b.args = append(b.args, a...)
	return b
}

func (b *dockerRunBuilder) Build() []string {
	return b.args
}

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

	var args []string
	var tempFile string
	var cleanup func()

	base := newDockerRunBuilder(platform)
	withCommon := func(b *dockerRunBuilder) *dockerRunBuilder {
		return b.Env(spec.Env).WorkDir(spec.WorkDir).Volumes(spec.Volumes)
	}

	switch {
	case len(spec.Volumes) > 0 && spec.Command == "" && spec.Entrypoint == "":
		// Volumes only: fraglet-entrypoint containers; default entrypoint runs mounted fraglet.
		args = withCommon(base).Image(spec.Container).Args(spec.Args...).Build()
	case spec.Entrypoint != "" && spec.Command != "":
		// Entrypoint + command: write command to temp file, mount it, run via entrypoint.
		var err error
		tempFile, cleanup, err = writeTempScript(spec.Command)
		if err != nil {
			return nil, fmt.Errorf("failed to create temp script: %w", err)
		}
		args = base.Entrypoint(spec.Entrypoint).
			Volume(tempFile, "/tmp/script", true).
			Env(spec.Env).WorkDir(spec.WorkDir).Volumes(spec.Volumes).
			Image(spec.Container).Args("/tmp/script").Args(spec.Args...).Build()
	case spec.Entrypoint != "":
		// Entrypoint only: no command body.
		args = withCommon(base.Entrypoint(spec.Entrypoint)).Image(spec.Container).Args(spec.Args...).Build()
	case spec.Command != "":
		// Command via sh -c; args don't apply.
		args = withCommon(base).Image(spec.Container).Args("sh", "-c", spec.Command).Build()
	default:
		// Plain run: image + optional args.
		args = withCommon(base).Image(spec.Container).Args(spec.Args...).Build()
	}

	dockerCmd := exec.CommandContext(ctx, args[0], args[1:]...)

	if spec.StdinReader != nil {
		dockerCmd.Stdin = spec.StdinReader
	} else if spec.Stdin != "" {
		dockerCmd.Stdin = bytes.NewBufferString(spec.Stdin)
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
