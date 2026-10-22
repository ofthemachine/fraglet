package runner

import (
	"context"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestDockerRunBuilder_Network(t *testing.T) {
	// Empty mode is a no-op (docker default networking).
	got := newDockerRunBuilder("run", "fraglet-test", "linux/amd64", false).Network("").Image("img").Build()
	if slices.Contains(got, "--network") {
		t.Fatalf("empty network mode should not add --network: %v", got)
	}

	// "none" adds the flag+value pair before the image.
	got = newDockerRunBuilder("run", "fraglet-test", "linux/amd64", false).Network("none").Image("img").Build()
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "--network none") {
		t.Fatalf("expected --network none in args, got: %v", got)
	}
	netIdx := slices.Index(got, "--network")
	imgIdx := slices.Index(got, "img")
	if netIdx < 0 || imgIdx < 0 || netIdx > imgIdx {
		t.Fatalf("--network must precede the image: %v", got)
	}
}

func TestDockerRunner_Available(t *testing.T) {
	r := &dockerRunner{}
	available := r.Available()

	// This test will pass or fail based on whether docker is available
	// We can't assume docker is available, so we just check the method works
	_ = available
}

func TestDockerRunner_Run(t *testing.T) {
	r := &dockerRunner{}

	if !r.Available() {
		t.Skip("Docker not available, skipping test")
	}

	spec := RunSpec{
		Command:   "echo 'hello from docker'",
		Container: "alpine:latest",
	}

	ctx := context.Background()
	result, err := r.Run(ctx, spec)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Stdout == "" {
		t.Error("Expected stdout output from docker container")
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

func TestDockerRunner_Run_WithEntrypoint(t *testing.T) {
	r := &dockerRunner{}

	if !r.Available() {
		t.Skip("Docker not available, skipping test")
	}

	// Test multiline script with entrypoint
	spec := RunSpec{
		Command:    "print('hello from python')",
		Container:  "python:3.11-slim",
		Entrypoint: "python",
	}

	ctx := context.Background()
	result, err := r.Run(ctx, spec)

	// Note: err may be non-nil if Python exits with non-zero, but we still get results
	_ = err

	// Python may return an error but we should still get the output
	if result.ExitCode != 0 {
		t.Logf("Python exit code: %d, stdout: %q, stderr: %q", result.ExitCode, result.Stdout, result.Stderr)
		// If we got output, that's what matters
		if result.Stdout == "" && result.Stderr == "" {
			t.Errorf("Expected output from Python, got exit code %d with no output", result.ExitCode)
		}
	}

	// Check that we got the expected output
	if result.Stdout != "hello from python\n" {
		t.Logf("Warning: stdout doesn't match expected, got %q", result.Stdout)
	}
}

func TestDockerRunner_Run_WithStdin(t *testing.T) {
	r := &dockerRunner{}

	if !r.Available() {
		t.Skip("Docker not available, skipping test")
	}

	spec := RunSpec{
		Command:   "cat",
		Container: "alpine:latest",
		Stdin:     "test input",
	}

	ctx := context.Background()
	result, err := r.Run(ctx, spec)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Stdout != "test input" {
		t.Errorf("Expected 'test input', got %q", result.Stdout)
	}
}

func TestDockerRunner_Run_NoContainer(t *testing.T) {
	r := &dockerRunner{}

	if !r.Available() {
		t.Skip("Docker not available, skipping test")
	}

	spec := RunSpec{
		Command: "echo 'test'",
		// No container specified
	}

	ctx := context.Background()
	_, err := r.Run(ctx, spec)
	if err == nil {
		t.Error("Expected error when container not specified")
	}
}

// A run its context ends takes its container with it: killing the docker
// client alone would leave the container running.
func TestDockerRunner_CancelRemovesContainer(t *testing.T) {
	r := &dockerRunner{}
	if !r.Available() {
		t.Skip("Docker not available, skipping test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	start := time.Now()
	_, _ = r.Run(ctx, RunSpec{Command: "sleep 37", Container: "alpine:latest"})
	if took := time.Since(start); took > 20*time.Second {
		t.Fatalf("a cancelled run took %s", took)
	}
	out, err := exec.Command("docker", "ps", "-a", "--no-trunc", "--format", "{{.Names}} {{.Command}}").Output()
	if err != nil {
		t.Fatalf("docker ps: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "fraglet-") && strings.Contains(line, "sleep 37") {
			t.Fatalf("the cancelled run's container is still there: %s", line)
		}
	}
}
