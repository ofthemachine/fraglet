package runner

import (
	"os"
	"path/filepath"
	"strings"
)

// hasShebang checks if a command starts with a shebang
func hasShebang(command string) bool {
	return strings.HasPrefix(strings.TrimSpace(command), "#!")
}

// writeTempScript writes a command to a temporary file and returns the path and cleanup function
// The file is made executable and has no extension - suitable for shebang execution
func writeTempScript(command string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "fraglet-script-*")
	if err != nil {
		return "", nil, err
	}

	// Write command to temp file
	if _, err := tmpFile.WriteString(command); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	// Make script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	cleanup := func() {
		os.Remove(absPath)
	}

	return absPath, cleanup, nil
}

// writeTempFragment writes content to a temporary file for fragment injection
// Returns the absolute path and cleanup function
func writeTempFragment(content string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", "fraglet-fragment-*")
	if err != nil {
		return "", nil, err
	}

	// Write content to temp file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	absPath, err := filepath.Abs(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	cleanup := func() {
		os.Remove(absPath)
	}

	return absPath, cleanup, nil
}
