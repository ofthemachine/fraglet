//go:build integration

package main_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ofthemachine/clitest"
)

func TestVeins(t *testing.T) {
	// Find the test file's directory to locate test cases relative to it
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)
	rootDir := filepath.Clean(filepath.Join(testDir, ".."))

	opts := clitest.Options{
		RootDir:           rootDir,
		BaseDirs:          []string{testDir},
		EnvOverrideVar:    "VEINS_TEST_SUITE_DIR",
		BinaryName:        "fragletc",
		BuildCommand:      []string{"sh", "-c", "cd " + filepath.Dir(testDir) + " && make build && cp fragletc " + testDir + "/"},
		ProjectRootMarker: "go.mod",
		Environment: map[string]string{
			"FRAGLET_VEIN_TAG_DISCOVERY_ORDER": "local,latest",
		},
	}

	clitest.RunSuite(t, opts)
}
