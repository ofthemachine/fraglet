package entrypoint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

func TestManager_Process_ArgvModeSkipsInjection(t *testing.T) {
	dir := t.TempDir()
	fragletPath := filepath.Join(dir, "FRAGLET")
	if err := os.WriteFile(fragletPath, []byte("meme-cli render drake\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &fraglet.EntrypointConfig{
		FragletTempPath: fragletPath,
		ModeConfig: fraglet.ModeConfig{
			// A codePath that doesn't exist: if Process() ever tried to
			// inject in argv mode, this would fail loudly, catching a
			// regression where the Argv skip is bypassed.
			Injection: fraglet.InjectionConfig{CodePath: filepath.Join(dir, "does-not-exist.sh"), Match: "FRAGLET"},
			Execution: &fraglet.EntrypointExecutionConfig{Path: "/meme-cli", Argv: true},
		},
	}

	if err := NewManager(cfg).Process(); err != nil {
		t.Fatalf("Process() in argv mode: error = %v, want nil (injection must be skipped)", err)
	}

	// The mounted body must be left untouched for the executor to read.
	data, err := os.ReadFile(fragletPath)
	if err != nil {
		t.Fatalf("fraglet body was removed/unreadable after Process(): %v", err)
	}
	if string(data) != "meme-cli render drake\n" {
		t.Fatalf("fraglet body modified by Process() in argv mode: %q", data)
	}
}

func TestManager_Process_NonArgvModeStillInjects(t *testing.T) {
	dir := t.TempDir()
	fragletPath := filepath.Join(dir, "FRAGLET")
	if err := os.WriteFile(fragletPath, []byte("echo REPLACED"), 0o644); err != nil {
		t.Fatal(err)
	}
	codePath := filepath.Join(dir, "hello-world.sh")
	if err := os.WriteFile(codePath, []byte("#!/bin/sh\necho FRAGLET\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &fraglet.EntrypointConfig{
		FragletTempPath: fragletPath,
		ModeConfig: fraglet.ModeConfig{
			Injection: fraglet.InjectionConfig{CodePath: codePath, Match: "FRAGLET"},
			Execution: &fraglet.EntrypointExecutionConfig{Path: codePath},
		},
	}

	if err := NewManager(cfg).Process(); err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	data, err := os.ReadFile(codePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "#!/bin/sh\necho REPLACED\n" {
		t.Fatalf("codePath after injection = %q, want the injected line to have replaced the marker", data)
	}
}
