package runner

import (
	"context"
	"testing"
)

func TestLocalRunner_Run_WithShebang(t *testing.T) {
	r := &localRunner{}

	// Use a simple bash shebang that should work on any Unix-like system
	spec := RunSpec{
		Command: `#!/bin/sh
echo "hello from shebang"
`,
	}

	ctx := context.Background()
	result, err := r.Run(ctx, spec)
	if err != nil {
		t.Fatalf("Unexpected error: %v, stderr: %q", err, result.Stderr)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d, stderr: %q", result.ExitCode, result.Stderr)
	}

	if result.Stdout != "hello from shebang\n" {
		t.Errorf("Expected 'hello from shebang\\n', got %q", result.Stdout)
	}
}

func TestLocalRunner_Run_WithEntrypoint(t *testing.T) {
	r := &localRunner{}

	// Test entrypoint handling (same as docker, just locally)
	spec := RunSpec{
		Command:    "print('hello from python')",
		Entrypoint: "python3",
	}

	ctx := context.Background()
	result, err := r.Run(ctx, spec)
	if err != nil {
		t.Fatalf("Unexpected error: %v, stderr: %q", err, result.Stderr)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d, stderr: %q", result.ExitCode, result.Stderr)
	}

	if result.Stdout != "hello from python\n" {
		t.Errorf("Expected 'hello from python\\n', got %q", result.Stdout)
	}
}

func TestLocalRunner_Run_NoShebang(t *testing.T) {
	r := &localRunner{}

	spec := RunSpec{
		Command: "echo 'hello world'",
	}

	ctx := context.Background()
	result, err := r.Run(ctx, spec)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Stdout != "hello world\n" {
		t.Errorf("Expected 'hello world\\n', got %q", result.Stdout)
	}
}
