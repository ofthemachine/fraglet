package clitest

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Options controls how the CLI harness runs.
type Options struct {
	BaseDirs          []string
	EnvOverrideVar    string
	BinaryName        string
	BuildCommand      []string // e.g. {"go","build","-o","smax"}
	ProjectRootMarker string   // e.g. "go.mod"
	DefaultPatterns   map[string]string
	Environment       map[string]string // Additional environment variables to set
}

type Result struct {
	Total, Passed, Failed int32
	FailedDetails         []string
}

type CLITestCase struct {
	Name       string
	Path       string
	ActScript  string
	AssertFile string
}

// Build-once state
var (
	projectRoot string
	builtOnce   bool
	buildErr    error
)

// RunSuite discovers and runs CLI tests using act.sh + assert.txt.
func RunSuite(t *testing.T, opts Options) Result {
	// defaults
	if len(opts.BaseDirs) == 0 {
		opts.BaseDirs = []string{"cmd_samples", "integration"}
	}
	if opts.EnvOverrideVar == "" {
		opts.EnvOverrideVar = "CLI_TEST_SUITE_DIR"
	}
	if opts.BinaryName == "" {
		opts.BinaryName = "app"
	}
	if len(opts.BuildCommand) == 0 {
		opts.BuildCommand = []string{"go", "build", "-o", opts.BinaryName}
	}
	if opts.ProjectRootMarker == "" {
		opts.ProjectRootMarker = "go.mod"
	}
	if opts.DefaultPatterns == nil {
		opts.DefaultPatterns = defaultPatterns
	}

	binPath := buildOnce(t, opts)

	// Allow single dir override
	if override := os.Getenv(opts.EnvOverrideVar); override != "" {
		opts.BaseDirs = []string{override}
	}

	testCases, err := discoverTestCases(opts.BaseDirs)
	if err != nil {
		t.Fatalf("discover test cases: %v", err)
	}
	if len(testCases) == 0 {
		t.Logf("No test cases found under: %s", strings.Join(opts.BaseDirs, ", "))
		return Result{}
	}
	t.Logf("Discovered %d test cases", len(testCases))

	var total, passed, failed int32
	var failedDetails []string

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			atomic.AddInt32(&total, 1)
			start := time.Now()
			defer func() {
				status := "PASS"
				if t.Failed() {
					status = "FAIL"
					atomic.AddInt32(&failed, 1)
					failedDetails = append(failedDetails, fmt.Sprintf("%s: %s", tc.Name, tc.AssertFile))
				} else {
					atomic.AddInt32(&passed, 1)
				}
				fmt.Printf("TEST_SUMMARY_MARKER: %s - %s (%.2fs)\n", tc.Name, status, time.Since(start).Seconds())
			}()

			tempDir := t.TempDir()

			// Copy all files from test directory to tempDir (except act.sh and assert.txt)
			if err := copyTestDirectoryContents(tc.Path, tempDir); err != nil {
				t.Fatalf("copy test directory contents: %v", err)
			}

			// place binary
			localBin := filepath.Join(tempDir, opts.BinaryName)
			if err := copyFile(binPath, localBin); err != nil {
				t.Fatalf("copy binary: %v", err)
			}
			_ = os.Chmod(localBin, 0755)

			// act
			stdout, stderr, exitCode, actErr := runActScript(t, tempDir, tc.ActScript, opts.Environment)
			if actErr != nil && exitCode == -1 {
				t.Fatalf("act.sh execution failed: %v", actErr)
			}

			// assert
			if err := assertResultsText(t, tc.AssertFile, stdout+stderr, opts.DefaultPatterns); err != nil {
				t.Errorf("assert failed: %v", err)
			}
		})
	}

	return Result{Total: total, Passed: passed, Failed: failed, FailedDetails: failedDetails}
}

func exists(path string) bool { _, err := os.Stat(path); return err == nil }

func findProjectRoot(startPath string, marker string) (string, error) {
	current, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("marker %s not found", marker)
		}
		current = parent
	}
}

func buildOnce(t *testing.T, opts Options) string {
	if projectRoot == "" {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatalf("runtime.Caller failed")
		}
		root, err := findProjectRoot(filepath.Dir(file), opts.ProjectRootMarker)
		if err != nil {
			t.Fatalf("find project root: %v", err)
		}
		projectRoot = root
		t.Logf("Project root: %s", projectRoot)
	}
	if builtOnce {
		return filepath.Join(projectRoot, opts.BinaryName)
	}
	cmd := exec.Command(opts.BuildCommand[0], opts.BuildCommand[1:]...)
	cmd.Dir = projectRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		buildErr = fmt.Errorf("failed to build (cwd: %s): %v\nOutput:\n%s", projectRoot, err, string(out))
		t.Fatalf("%s", buildErr.Error())
	}
	builtOnce = true
	return filepath.Join(projectRoot, opts.BinaryName)
}

func discoverTestCases(baseDirs []string) ([]CLITestCase, error) {
	var out []CLITestCase
	for _, base := range baseDirs {
		if _, err := os.Stat(base); err != nil {
			continue
		}
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				act := filepath.Join(path, "act.sh")
				if exists(act) {
					assert := filepath.Join(path, "assert.txt")
					if !exists(assert) {
						return nil // Skip if no assert.txt
					}
					name := strings.TrimPrefix(path, base+"/")
					name = strings.ReplaceAll(fmt.Sprintf("%s:%s", base, name), "/", "_")
					out = append(out, CLITestCase{Name: name, Path: path, ActScript: act, AssertFile: assert})
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func runActScript(t *testing.T, tempDir, actScriptPath string, additionalEnv map[string]string) (stdout, stderr string, exitCode int, err error) {
	if !exists(actScriptPath) {
		return "", "", -1, fmt.Errorf("act.sh not found: %s", actScriptPath)
	}
	local := filepath.Join(tempDir, "act.sh")
	if err := copyFile(actScriptPath, local); err != nil {
		return "", "", -1, err
	}
	_ = os.Chmod(local, 0755)
	allow := []string{"HOME", "PATH", "SHELL", "LANG", "TZ"}
	env := []string{"TEST_TEMP_DIR=" + tempDir, "USER=test"}
	var existingPath string
	for _, k := range allow {
		if v, ok := os.LookupEnv(k); ok {
			if k == "PATH" {
				existingPath = v
				continue
			}
			env = append(env, k+"="+v)
		}
	}

	// Add additional environment variables
	for k, v := range additionalEnv {
		env = append(env, k+"="+v)
	}
	if existingPath != "" {
		env = append(env, "PATH="+tempDir+":"+existingPath)
	} else {
		env = append(env, "PATH="+tempDir)
	}
	const maxRetries = 3
	var last error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		cmd := exec.Command("./act.sh")
		cmd.Dir = tempDir
		cmd.Env = env
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		execErr := cmd.Run()
		stdout, stderr = outBuf.String(), errBuf.String()
		if execErr != nil {
			last = execErr
			if isTransientError(execErr) && attempt < maxRetries {
				continue
			}
			if exitError, ok := execErr.(*exec.ExitError); ok {
				return stdout, stderr, exitError.ExitCode(), execErr
			}
			return stdout, stderr, -1, fmt.Errorf("running act.sh failed: %w", execErr)
		}
		return stdout, stderr, 0, nil
	}
	return "", "", -1, fmt.Errorf("max retries exceeded: %v", last)
}

func assertResultsText(t *testing.T, assertFile string, combinedOutput string, patterns map[string]string) error {
	data, err := os.ReadFile(assertFile)
	if err != nil {
		return fmt.Errorf("read assert.txt: %v", err)
	}
	lines := []string{}
	for _, l := range strings.Split(string(data), "\n") {
		s := strings.TrimRight(l, "\r")
		st := strings.TrimSpace(s)
		if st == "" {
			continue
		}
		if strings.HasPrefix(st, "#") && !strings.HasPrefix(st, "# REGEX: ") {
			continue
		}
		lines = append(lines, s)
	}
	expected := strings.Join(lines, "\n")
	return matchOutput(expected, combinedOutput, "ORDERED_LINES", patterns)
}

func matchOutput(expectedContent, actualOutput string, matchType string, patterns map[string]string) error {
	switch matchType {
	case "EXACT":
		if actualOutput != expectedContent {
			return fmt.Errorf("output mismatch\nExpected:\n%s\nActual:\n%s", expectedContent, actualOutput)
		}
		return nil
	case "SUBSTRING":
		if !strings.Contains(actualOutput, expectedContent) {
			return fmt.Errorf("missing substring:\n%q\nActual:\n%s", expectedContent, actualOutput)
		}
		return nil
	case "REGEX":
		matched, err := regexp.MatchString(expectedContent, actualOutput)
		if err != nil {
			return fmt.Errorf("invalid regex: %v", err)
		}
		if !matched {
			return fmt.Errorf("regex not matched: %q\nActual:\n%s", expectedContent, actualOutput)
		}
		return nil
	case "ORDERED_LINES":
		expLines := strings.Split(strings.TrimSpace(expectedContent), "\n")
		actLines := strings.Split(strings.TrimSpace(actualOutput), "\n")
		currentPos := 0
		const maxLookahead = 15
		for i, exp := range expLines {
			exp = strings.TrimSpace(exp)
			if exp == "" {
				continue
			}
			if strings.HasPrefix(exp, "#") && !strings.HasPrefix(exp, "# REGEX: ") {
				continue
			}
			found := false
			isRegex := false
			var regexPattern string
			if strings.HasPrefix(exp, "# REGEX: ") {
				isRegex = true
				regexPattern = strings.TrimSpace(strings.TrimPrefix(exp, "# REGEX: "))
			}
			if !isRegex && (strings.HasPrefix(exp, "re:") || strings.HasPrefix(exp, "regex:")) {
				isRegex = true
				if strings.HasPrefix(exp, "re:") {
					regexPattern = strings.TrimPrefix(exp, "re:")
				} else {
					regexPattern = strings.TrimPrefix(exp, "regex:")
				}
			}
			if !isRegex && strings.HasPrefix(exp, "/") && strings.HasSuffix(exp, "/") && len(exp) >= 2 {
				isRegex = true
				regexPattern = strings.TrimSuffix(strings.TrimPrefix(exp, "/"), "/")
			}
			searchEnd := min(currentPos+maxLookahead, len(actLines))
			for pos := currentPos; pos < searchEnd; pos++ {
				actual := strings.TrimSpace(actLines[pos])
				var matched bool
				if isRegex {
					re, err := regexp.Compile(regexPattern)
					if err != nil {
						return fmt.Errorf("invalid regex pattern %q: %v", regexPattern, err)
					}
					matched = re.MatchString(actual)
				} else {
					matched = matchLine(exp, actual, patterns)
				}
				if matched {
					found = true
					currentPos = pos + 1
					break
				}
			}
			if !found {
				var context []string
				for j := max(0, currentPos-3); j < min(len(actLines), currentPos+10); j++ {
					context = append(context, fmt.Sprintf("  %d: %s", j+1, actLines[j]))
				}
				return fmt.Errorf("ORDERED_LINES broken at expected line %d: %q\nNext few actual lines:\n%s", i+1, exp, strings.Join(context, "\n"))
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported match type: %s", matchType)
	}
}

func matchLine(expected, actual string, patterns map[string]string) bool {
	pl := regexp.MustCompile(`\{\{([^}:]+)(?::([^}]+))?\}\}`)
	if !pl.MatchString(expected) {
		return actual == expected
	}
	regexPattern := expected
	regexPattern = pl.ReplaceAllStringFunc(regexPattern, func(m string) string {
		mm := pl.FindStringSubmatch(m)
		if len(mm) < 2 {
			return m
		}
		name := mm[1]
		custom := mm[2]
		if custom != "" {
			return custom
		}
		if p, ok := patterns[name]; ok {
			return p
		}
		return m
	})
	regexPattern = regexp.QuoteMeta(regexPattern)
	for _, p := range patterns {
		regexPattern = strings.ReplaceAll(regexPattern, regexp.QuoteMeta(p), p)
	}
	matched, err := regexp.MatchString(regexPattern, actual)
	if err != nil {
		return false
	}
	return matched
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := out.ReadFrom(in); err != nil {
		return err
	}
	return nil
}

func copyDir(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}
	if err := os.MkdirAll(dst, si.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
		} else {
			if err := copyFile(s, d); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyTestDirectoryContents copies all files from the test directory to tempDir,
// excluding act.sh and assert.txt which are handled separately
func copyTestDirectoryContents(testDir, tempDir string) error {
	entries, err := os.ReadDir(testDir)
	if err != nil {
		return fmt.Errorf("read test directory: %w", err)
	}

	// Files to skip (handled separately)
	skipFiles := map[string]bool{
		"act.sh":     true,
		"assert.txt": true,
	}

	for _, entry := range entries {
		if skipFiles[entry.Name()] {
			continue
		}

		src := filepath.Join(testDir, entry.Name())
		dst := filepath.Join(tempDir, entry.Name())

		if entry.IsDir() {
			if err := copyDir(src, dst); err != nil {
				return fmt.Errorf("copy directory %s: %w", entry.Name(), err)
			}
		} else {
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("copy file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "text file busy") {
		return true
	}
	if strings.Contains(s, "not initialized") {
		return true
	}
	for _, p := range []string{"resource temporarily unavailable", "no such process", "interrupted system call", "connection reset"} {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

var defaultPatterns = map[string]string{
	"hash8":     `[a-f0-9]{8}`,
	"hash64":    `[a-f0-9]{64}`,
	"timestamp": `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}`,
	"number":    `\d+`,
	"path":      `/[^\s]+`,
	"any":       `.+`,
	"content":   `[A-Za-z0-9+/=]+`,
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
