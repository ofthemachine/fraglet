//go:build integration

package main_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ofthemachine/fraglet/pkg/clitest"
)

func TestEnvelopes(t *testing.T) {
	// Find the test file's directory to locate test cases relative to it
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)

	opts := clitest.Options{
		BaseDirs:          []string{testDir},
		EnvOverrideVar:    "ENVELOPES_TEST_SUITE_DIR",
		BinaryName:        "fragletc",
		BuildCommand:      []string{"sh", "-c", "cd " + filepath.Dir(testDir) + " && make build-cli && cp fragletc " + testDir + "/"},
		ProjectRootMarker: "go.mod",
	}

	clitest.RunSuite(t, opts)
}

