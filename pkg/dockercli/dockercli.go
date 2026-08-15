// Package dockercli resolves the real path to the docker CLI binary,
// robust to installations where it never landed on PATH. Docker Desktop on
// macOS is the common case this exists for: it symlinks its CLI into
// /usr/local/bin during first-run setup, but that step can be skipped,
// deferred, or simply never reached on a headless/automated install --
// and on Apple Silicon, /usr/local/bin may not even exist (Homebrew uses
// /opt/homebrew there instead). Found live: a real, running Docker Desktop
// installation -- daemon up, `docker info` answering fine via its full
// path -- was invisible to every exec.Command("docker", ...) call in this
// codebase, which all treated "not on PATH" as "not installed."
package dockercli

import (
	"os"
	"os/exec"
)

// fallbackPaths are well-known install locations, checked in order, only
// after an ordinary PATH lookup has already failed -- this never changes
// resolution for the common case where `docker` is already on PATH.
var fallbackPaths = []string{
	"/Applications/Docker.app/Contents/Resources/bin/docker", // Docker Desktop, macOS
	"/usr/local/bin/docker",
	"/opt/homebrew/bin/docker",
}

// Binary returns the docker CLI's resolved path. FRAGLET_DOCKER_BIN
// overrides everything (test/CI escape hatch, matching the FRAGLETC_BIN
// convention callers of this binary already use); otherwise an ordinary
// PATH lookup; otherwise the first fallback location that actually exists
// on disk. Callers pass the result straight to exec.Command -- when
// nothing resolves, "docker" is returned unchanged so the eventual exec
// error is the same familiar "executable file not found in $PATH" it
// always was, not a new failure mode to handle.
func Binary() string {
	if v := os.Getenv("FRAGLET_DOCKER_BIN"); v != "" {
		return v
	}
	if p, err := exec.LookPath("docker"); err == nil {
		return p
	}
	for _, candidate := range fallbackPaths {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return "docker"
}
