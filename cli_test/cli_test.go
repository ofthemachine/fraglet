//go:build integration

package main_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ofthemachine/clitest"
)

func TestFragletcCLI(t *testing.T) {
	// Find the test file's directory to locate test cases relative to it
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	rootDir := filepath.Clean(filepath.Join(testDir, ".."))

	opts := clitest.Options{
		RootDir:           rootDir,
		BaseDirs:          []string{testDir},
		EnvOverrideVar:    "CLI_TEST_SUITE_DIR",
		BinaryName:        "fragletc",
		BuildCommand:      []string{"sh", "-c", "cd " + filepath.Dir(testDir) + " && make build && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o fraglet-entrypoint ./cmd/entrypoint"},
		ProjectRootMarker: "go.mod",
		CopyGlobs:         []string{"fraglet-entrypoint"},
	}

	clitest.RunSuite(t, opts)
}
