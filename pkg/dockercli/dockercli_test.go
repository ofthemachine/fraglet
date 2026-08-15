package dockercli

import (
	"os"
	"path/filepath"
	"testing"
)

func withEnv(t *testing.T, key, value string) {
	t.Helper()
	old, had := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if had {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	})
}

func TestBinary_EnvOverrideWins(t *testing.T) {
	withEnv(t, "FRAGLET_DOCKER_BIN", "/some/explicit/docker")
	if got := Binary(); got != "/some/explicit/docker" {
		t.Fatalf("got %q, want explicit override", got)
	}
}

// TestBinary_FallsBackWhenNotOnPATH reproduces the exact failure found
// live: a real Docker Desktop install (daemon up, CLI answering fine via
// its full path) that PATH simply doesn't reach -- e.g. a launchd
// LaunchDaemon's minimal default PATH, or Apple Silicon's missing
// /usr/local/bin. With PATH empty, an ordinary exec.LookPath("docker")
// always fails, so Binary() must fall through to a well-known install
// location -- proven here against a real, freshly-created file standing in
// for one, not the real Docker Desktop path (which may not exist on the
// machine running this test).
func TestBinary_FallsBackWhenNotOnPATH(t *testing.T) {
	withEnv(t, "PATH", "")

	dir := t.TempDir()
	fake := filepath.Join(dir, "docker")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\necho fake docker\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	old := fallbackPaths
	fallbackPaths = []string{"/definitely/does/not/exist/docker", fake}
	t.Cleanup(func() { fallbackPaths = old })

	if got := Binary(); got != fake {
		t.Fatalf("got %q, want fallback %q", got, fake)
	}
}

func TestBinary_NoResolutionReturnsLiteralDocker(t *testing.T) {
	withEnv(t, "PATH", "")

	old := fallbackPaths
	fallbackPaths = []string{"/definitely/does/not/exist/docker"}
	t.Cleanup(func() { fallbackPaths = old })

	if got := Binary(); got != "docker" {
		t.Fatalf("got %q, want literal \"docker\" so the exec error stays familiar", got)
	}
}
