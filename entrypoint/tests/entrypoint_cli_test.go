//go:build integration

package main_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ofthemachine/fraglet/pkg/clitest"
)

func TestCLI(t *testing.T) {
	// Find the test file's directory to locate test cases relative to it
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)

	opts := clitest.Options{
		BaseDirs:          []string{testDir},
		EnvOverrideVar:    "ENTRYPOINT_TEST_SUITE_DIR",
		BinaryName:        "fraglet-entrypoint",
		BuildCommand:      []string{"sh", "-c", "cd entrypoint && GOOS=linux GOARCH=amd64 go build -o ../fraglet-entrypoint ./cmd"},
		ProjectRootMarker: "go.mod",
	}

	clitest.RunSuite(t, opts)
}
