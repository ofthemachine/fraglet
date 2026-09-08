package runner

import (
	"context"
	"sync"
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

	var stdout, stderr string
	var collected sync.WaitGroup

	// Collect stdout/stderr chunks. A prior version read `stdout` from the
	// main goroutine after a fixed time.Sleep, racing with these writes
	// (go test -race caught it every time) — collected.Wait() below gives
	// an actual happens-before guarantee instead of guessing at a delay.
	collected.Add(2)
	go func() {
		defer collected.Done()
		for chunk := range streaming.Stdout {
			stdout += chunk
		}
	}()
	go func() {
		defer collected.Done()
		for chunk := range streaming.Stderr {
			stderr += chunk
		}
	}()

	// Wait for command to complete
	var exitCode int
	select {
	case err := <-streaming.Done:
		if err != nil {
			t.Logf("Command completed with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for command to complete")
	}

	select {
	case exitCode = <-streaming.ExitCode:
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for exit code")
	}

	collected.Wait()

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stdout == "" {
		t.Error("Expected stdout output")
	}
}
