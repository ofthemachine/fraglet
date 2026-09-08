package executor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

// writeRecorder writes an executable shell script that appends each argv
// word (one per line) to dir/argv.out and exits with exitCode. Used to
// observe exactly what argv executeArgvMode hands to exec.Command, since
// it wires the real process's stdout/stderr/stdin, not a pipe we can read.
func writeRecorder(t *testing.T, dir, name string, exitCode int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	script := fmt.Sprintf("#!/bin/sh\nfor a in \"$@\"; do printf '%%s\\n' \"$a\"; done > %s\nexit %d\n",
		filepath.Join(dir, "argv.out"), exitCode)
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func readArgvOut(t *testing.T, dir string) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "argv.out"))
	if err != nil {
		t.Fatalf("reading argv.out: %v", err)
	}
	trimmed := strings.TrimRight(string(data), "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

func TestExecuteArgvMode_NoBodyMounted_PassesArgsThrough(t *testing.T) {
	dir := t.TempDir()
	bin := writeRecorder(t, dir, "recorder", 3)

	cfg := &fraglet.EntrypointConfig{
		FragletTempPath: filepath.Join(dir, "FRAGLET"), // never written: no fraglet mounted
		ModeConfig: fraglet.ModeConfig{
			Execution: &fraglet.EntrypointExecutionConfig{Path: bin, Argv: true},
		},
	}

	code, err := NewExecutor(cfg).ExecuteWithArgs([]string{"foo", "bar baz"})
	if err != nil {
		t.Fatalf("ExecuteWithArgs: %v", err)
	}
	if code != 3 {
		t.Fatalf("exit code = %d, want 3 (must propagate from the executed process)", code)
	}
	if got := readArgvOut(t, dir); len(got) != 2 || got[0] != "foo" || got[1] != "bar baz" {
		t.Fatalf("argv = %q, want [foo, \"bar baz\"] passed straight through", got)
	}
}

func TestExecuteArgvMode_BodyMounted_ExpandsAndExecs(t *testing.T) {
	dir := t.TempDir()
	bin := writeRecorder(t, dir, "recorder", 0)
	fragletPath := filepath.Join(dir, "FRAGLET")
	body := bin + " --flag $1\n"
	if err := os.WriteFile(fragletPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &fraglet.EntrypointConfig{
		FragletTempPath: fragletPath,
		ModeConfig: fraglet.ModeConfig{
			Execution: &fraglet.EntrypointExecutionConfig{Path: bin, Argv: true},
		},
	}

	code, err := NewExecutor(cfg).ExecuteWithArgs([]string{"value"})
	if err != nil {
		t.Fatalf("ExecuteWithArgs: %v", err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if got := readArgvOut(t, dir); len(got) != 2 || got[0] != "--flag" || got[1] != "value" {
		t.Fatalf("argv = %q, want [--flag, value] from the expanded body", got)
	}
}

func TestExecuteArgvMode_BodyInvokesWrongBinary_Errors(t *testing.T) {
	dir := t.TempDir()
	bin := writeRecorder(t, dir, "recorder", 0)
	fragletPath := filepath.Join(dir, "FRAGLET")
	if err := os.WriteFile(fragletPath, []byte("some-other-tool --flag\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &fraglet.EntrypointConfig{
		FragletTempPath: fragletPath,
		ModeConfig: fraglet.ModeConfig{
			Execution: &fraglet.EntrypointExecutionConfig{Path: bin, Argv: true},
		},
	}

	_, err := NewExecutor(cfg).ExecuteWithArgs(nil)
	if err == nil || !strings.Contains(err.Error(), "body invokes") {
		t.Fatalf("err = %v, want an error naming the mismatched binary", err)
	}
}

func TestExecuteArgvMode_NoExecutionPath_Errors(t *testing.T) {
	cfg := &fraglet.EntrypointConfig{
		FragletTempPath: filepath.Join(t.TempDir(), "FRAGLET"),
		ModeConfig: fraglet.ModeConfig{
			Execution: &fraglet.EntrypointExecutionConfig{Argv: true},
		},
	}

	_, err := NewExecutor(cfg).ExecuteWithArgs(nil)
	if err == nil || !strings.Contains(err.Error(), "execution.path is required") {
		t.Fatalf("err = %v, want the missing execution.path error", err)
	}
}
