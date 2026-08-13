package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildFragletc compiles the fragletc binary once per test run into a temp
// dir and returns its path.
func buildFragletc(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available, skipping")
	}
	if exec.Command("docker", "version").Run() != nil {
		t.Skip("docker not available, skipping")
	}
	bin := filepath.Join(t.TempDir(), "fragletc")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build fragletc: %v\n%s", err, out)
	}
	return bin
}

// TestMain_RunsWithoutOutputFlag_WhenOutputDeclared is the regression test
// for the POLA bug: a fraglet that declares "output=" and unconditionally
// writes it must run successfully with no --output flag at all (today's
// declaration alone is enough to get /output mounted), and must say what it
// discarded rather than silently dropping the file with no explanation.
func TestMain_RunsWithoutOutputFlag_WhenOutputDeclared(t *testing.T) {
	bin := buildFragletc(t)

	dir := t.TempDir()
	// The script body never actually runs here — trailing args below
	// override the container command (alpine has no fraglet-aware
	// entrypoint to execute /FRAGLET as Python). Only the header's
	// output= declaration matters for this test.
	script := filepath.Join(dir, "script.py")
	if err := os.WriteFile(script, []byte(
		"#!/usr/bin/env -S fragletc --image alpine:latest\n#: output=result.txt\n\necho unused\n",
	), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "--image", "alpine:latest", script, "sh", "-c", "echo -n hi > /output/result.txt")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		t.Fatalf("fragletc failed (this is the POLA bug: declared output= must not require --output to run) — err=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "result.txt") {
		t.Errorf("expected a hint naming the discarded declared output, got stderr=%q", stderr.String())
	}
}

// TestMain_OutputFlag_CopiesRequestedFile confirms the existing --output
// path still works after the fix: the file is copied to the requested host
// destination.
func TestMain_OutputFlag_CopiesRequestedFile(t *testing.T) {
	bin := buildFragletc(t)

	dir := t.TempDir()
	script := filepath.Join(dir, "script.py")
	if err := os.WriteFile(script, []byte(
		"#!/usr/bin/env -S fragletc --image alpine:latest\n#: output=result.txt\n\necho unused\n",
	), 0o755); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "copied.txt")

	cmd := exec.Command(bin, "--image", "alpine:latest", "--output", "result.txt="+dest, script, "sh", "-c", "echo -n hi > /output/result.txt")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("fragletc failed: err=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read copied output: %v", err)
	}
	if string(data) != "hi" {
		t.Fatalf("copied content = %q, want hi", data)
	}
	// Regression: fragletc must confirm the ACTUAL delivered name (the
	// --output=hostdest rename), not just the declared relpath — otherwise
	// a rename is invisible in fragletc's own output and a reader has no
	// way to learn the delivered file isn't called "result.txt".
	wantConfirm := "fragletc: copied result.txt -> " + dest
	if !strings.Contains(stderr.String(), wantConfirm) {
		t.Errorf("expected delivery confirmation %q, got stderr=%q", wantConfirm, stderr.String())
	}
	if strings.Contains(stderr.String(), "discarded") {
		t.Errorf("requested output must not print the discard hint, got stderr=%q", stderr.String())
	}
}
