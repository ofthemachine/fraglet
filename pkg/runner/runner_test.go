package runner

import (
	"context"
	"testing"
	"time"
)

func TestNewRunner_Local(t *testing.T) {
	r := NewRunner("", "")
	if r.Name() != "local" {
		t.Errorf("Expected local runner, got %s", r.Name())
	}
	if !r.Available() {
		t.Error("Expected local runner to always be available")
	}
}

func TestNewRunner_Docker(t *testing.T) {
	// This will return docker if available, otherwise local
	r := NewRunner("python:3.11-slim", "")

	// Check if docker is available
	docker := &dockerRunner{}
	if docker.Available() {
		if r.Name() != "docker" {
			t.Errorf("Expected docker runner when docker is available, got %s", r.Name())
		}
	} else {
		// Docker not available, should fall back to local
		if r.Name() != "local" {
			t.Errorf("Expected local runner when docker unavailable, got %s", r.Name())
		}
	}
}

func TestLocalRunner_Run(t *testing.T) {
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

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if result.Duration == 0 {
		t.Error("Expected duration to be set")
	}
}

func TestLocalRunner_Run_WithStdin(t *testing.T) {
	r := &localRunner{}

	spec := RunSpec{
		Command: "cat",
		Stdin:   "test input",
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

func TestLocalRunner_Run_Error(t *testing.T) {
	r := &localRunner{}

	spec := RunSpec{
		Command: "false", // Command that exits with non-zero
	}

	ctx := context.Background()
	result, err := r.Run(ctx, spec)

	// We should get the exit code in the result (non-zero)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code, got %d", result.ExitCode)
	}

	// Note: exec.Run() returns an error for non-zero exit codes
	// but we still return the result with the exit code set
	// This matches the original behavior where warnings are logged but execution continues
	_ = err // Error may or may not be nil depending on exec behavior
}

func TestLocalRunner_RunStreaming(t *testing.T) {
	r := &localRunner{}

	spec := RunSpec{
		Command: "echo 'hello world'",
	}

	ctx := context.Background()
	streaming, err := r.RunStreaming(ctx, spec)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var stdout string
	var stderr string
	var exitCode int

	// Collect stdout chunks
	go func() {
		for chunk := range streaming.Stdout {
			stdout += chunk
		}
	}()

	// Collect stderr chunks
	go func() {
		for chunk := range streaming.Stderr {
			stderr += chunk
		}
	}()

	// Wait for exit code
	go func() {
		if code := <-streaming.ExitCode; code != 0 {
			exitCode = code
		}
	}()

	// Wait for command to complete
	timeout := time.After(5 * time.Second)
	select {
	case err := <-streaming.Done:
		if err != nil {
			t.Logf("Command completed with error: %v", err)
		}
		// Give goroutines a moment to finish
		time.Sleep(50 * time.Millisecond)
	case <-timeout:
		t.Fatal("Timeout waiting for command to complete")
	}

	// Read exit code if not already received
	select {
	case code := <-streaming.ExitCode:
		exitCode = code
	case <-time.After(100 * time.Millisecond):
		// Already received or won't receive
	}

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stdout == "" {
		t.Error("Expected stdout output")
	}
}
