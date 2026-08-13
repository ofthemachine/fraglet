package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseOutputRequests(t *testing.T) {
	reqs, err := parseOutputRequests([]string{"series.csv", "plot.png=/tmp/out.png"})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if len(reqs) != 2 {
		t.Fatalf("len = %d, want 2", len(reqs))
	}
	if reqs[0].RelPath != "series.csv" || reqs[0].Dest != "./series.csv" {
		t.Fatalf("reqs[0] = %+v, want default dest", reqs[0])
	}
	if reqs[1].RelPath != "plot.png" || reqs[1].Dest != "/tmp/out.png" {
		t.Fatalf("reqs[1] = %+v, want explicit dest", reqs[1])
	}
}

func TestParseOutputRequests_EmptyRelPath(t *testing.T) {
	if _, err := parseOutputRequests([]string{"=dest.txt"}); err == nil {
		t.Fatal("expected error for empty relpath")
	}
}

func TestValidateOutputRequests(t *testing.T) {
	code := "#!/usr/bin/env -S fragletc --image x\n#: output=series.csv\n#: output=plot.png\n\nbody"
	reqs := []outputRequest{{RelPath: "series.csv", Dest: "./series.csv"}}
	if err := validateOutputRequests(code, reqs); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestValidateOutputRequests_UndeclaredIsError(t *testing.T) {
	code := "#!/usr/bin/env -S fragletc --image x\n#: output=series.csv\n\nbody"
	reqs := []outputRequest{{RelPath: "missing.csv", Dest: "./missing.csv"}}
	err := validateOutputRequests(code, reqs)
	if err == nil {
		t.Fatal("expected error for undeclared output")
	}
}

func TestValidateOutputRequests_NoDeclarationsAtAll(t *testing.T) {
	code := "#!/usr/bin/env -S fragletc --image x\n\nbody"
	reqs := []outputRequest{{RelPath: "series.csv", Dest: "./series.csv"}}
	err := validateOutputRequests(code, reqs)
	if err == nil {
		t.Fatal("expected error when nothing is declared")
	}
}

func TestCopyRequestedOutputs(t *testing.T) {
	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "series.csv"), []byte("value\n1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	destDir := t.TempDir()
	dest := filepath.Join(destDir, "nested", "series.csv")
	reqs := []outputRequest{{RelPath: "series.csv", Dest: dest}}

	if err := copyRequestedOutputs(outDir, reqs); err != nil {
		t.Fatalf("err = %v", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(data) != "value\n1\n" {
		t.Fatalf("dest content = %q", data)
	}
}

func TestCopyRequestedOutputs_MissingFileIsError(t *testing.T) {
	outDir := t.TempDir()
	reqs := []outputRequest{{RelPath: "never-written.csv", Dest: filepath.Join(t.TempDir(), "x.csv")}}
	if err := copyRequestedOutputs(outDir, reqs); err == nil {
		t.Fatal("expected error for a declared-but-unwritten output")
	}
}

func TestParseOutputRequests_RejectsAbsoluteRelPath(t *testing.T) {
	if _, err := parseOutputRequests([]string{"/tmp/out.csv"}); err == nil {
		t.Fatal("expected error for absolute relpath")
	}
}

func TestParseOutputRequests_RejectsTraversalRelPath(t *testing.T) {
	if _, err := parseOutputRequests([]string{"../outside.csv"}); err == nil {
		t.Fatal("expected error for traversal relpath")
	}
}

func TestValidateOutputRequests_RejectsDeclaredTraversal(t *testing.T) {
	code := "#!/usr/bin/env -S fragletc --image x\n#: output=../escape.csv\n\nbody"
	reqs := []outputRequest{{RelPath: "../escape.csv", Dest: "./escape.csv"}}
	if err := validateOutputRequests(code, reqs); err == nil {
		t.Fatal("expected error for declared traversal output")
	}
}

func TestCopyRequestedOutputs_NestedValidOutput(t *testing.T) {
	outDir := t.TempDir()
	nested := filepath.Join(outDir, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "series.csv"), []byte("nested\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out.csv")
	reqs := []outputRequest{{RelPath: "nested/series.csv", Dest: dest}}
	if err := copyRequestedOutputs(outDir, reqs); err != nil {
		t.Fatalf("err = %v", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(data) != "nested\n" {
		t.Fatalf("dest content = %q", data)
	}
}

func TestCopyRequestedOutputs_SymlinkEscapeIsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink escape test requires Unix symlinks")
	}
	outDir := t.TempDir()
	secretDir := t.TempDir()
	secretFile := filepath.Join(secretDir, "secret.csv")
	if err := os.WriteFile(secretFile, []byte("secret\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkPath := filepath.Join(outDir, "series.csv")
	if err := os.Symlink(secretFile, linkPath); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out.csv")
	reqs := []outputRequest{{RelPath: "series.csv", Dest: dest}}
	if err := copyRequestedOutputs(outDir, reqs); err == nil {
		t.Fatal("expected error when output file resolves outside output root")
	}
}

func TestCopyRequestedOutputs_NonRegularSourceIsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory source test requires Unix semantics")
	}
	outDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(outDir, "series.csv"), 0o755); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out.csv")
	reqs := []outputRequest{{RelPath: "series.csv", Dest: dest}}
	if err := copyRequestedOutputs(outDir, reqs); err == nil {
		t.Fatal("expected error when source is not a regular file")
	}
}
