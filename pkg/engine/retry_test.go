package engine

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeRuns stands in for runOnce: each run reads its stdin, writes a line
// to stdout and a file to OutputHostDir, and exits with the next of codes.
func fakeRuns(t *testing.T, codes ...int) *[]string {
	t.Helper()
	var seen []string
	prev := runOnce
	t.Cleanup(func() { runOnce = prev })
	runOnce = func(_ context.Context, o RunOptions) (RunReport, error) {
		var in []byte
		if o.Stdin != nil {
			in, _ = io.ReadAll(o.Stdin)
		}
		left, _ := os.ReadDir(o.OutputHostDir)
		n := len(seen)
		seen = append(seen, string(in))
		_, _ = io.WriteString(o.Stdout, "run "+string(rune('1'+n))+"\n")
		_ = os.WriteFile(filepath.Join(o.OutputHostDir, "out"+string(rune('1'+n))), nil, 0o644)
		if len(left) > 0 {
			t.Errorf("run %d found the last run's files: %d", n+1, len(left))
		}
		var r RunReport
		r.ExitCode = codes[n]
		return r, nil
	}
	return &seen
}

func TestRunWithRetry_RetriesOnlyTempFail(t *testing.T) {
	seen := fakeRuns(t, ExitTempFail, ExitTempFail, 0)
	var stdout bytes.Buffer
	var waits []time.Duration
	report, err := RunWithRetry(context.Background(), RunOptions{Stdin: strings.NewReader("hi"), Stdout: &stdout, OutputHostDir: t.TempDir()},
		Retry{Times: 3, Wait: 10 * time.Millisecond, Waiting: func(_ int, w time.Duration) { waits = append(waits, w) }})
	if err != nil || report.ExitCode != 0 {
		t.Fatalf("exit %d, err %v", report.ExitCode, err)
	}
	if strings.Join(*seen, ",") != "hi,hi,hi" {
		t.Fatalf("every run reads the same stdin: %q", *seen)
	}
	if stdout.String() != "run 3\n" {
		t.Fatalf("only the last run's stdout reaches the caller: %q", stdout.String())
	}
	if len(waits) != 2 || waits[0] < 5*time.Millisecond || waits[0] > 10*time.Millisecond || waits[1] < 10*time.Millisecond || waits[1] > 20*time.Millisecond {
		t.Fatalf("waits are jittered and double: %v", waits)
	}
}

func TestRunWithRetry_OtherFailuresAreFinal(t *testing.T) {
	seen := fakeRuns(t, 1, 0)
	report, _ := RunWithRetry(context.Background(), RunOptions{Stdout: io.Discard, OutputHostDir: t.TempDir()}, Retry{Times: 3, Wait: time.Millisecond})
	if report.ExitCode != 1 || len(*seen) != 1 {
		t.Fatalf("exit %d after %d runs", report.ExitCode, len(*seen))
	}
}

func TestRunWithRetry_GivesUpAfterTimes(t *testing.T) {
	seen := fakeRuns(t, ExitTempFail, ExitTempFail, ExitTempFail)
	report, _ := RunWithRetry(context.Background(), RunOptions{Stdout: io.Discard, OutputHostDir: t.TempDir()}, Retry{Times: 2, Wait: time.Millisecond})
	if report.ExitCode != ExitTempFail || len(*seen) != 3 {
		t.Fatalf("exit %d after %d runs", report.ExitCode, len(*seen))
	}
}

func TestRunWithRetry_NoWaitPastTheDeadline(t *testing.T) {
	seen := fakeRuns(t, ExitTempFail, 0)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	report, _ := RunWithRetry(ctx, RunOptions{Stdout: io.Discard, OutputHostDir: t.TempDir()}, Retry{Times: 3, Wait: time.Hour})
	if report.ExitCode != ExitTempFail || len(*seen) != 1 {
		t.Fatalf("a wait the deadline cannot hold is not started: exit %d after %d runs", report.ExitCode, len(*seen))
	}
}

// A real run: a fraglet that exits 75 is run again; one its deadline ends
// reports ExitTimeout and why, and no error.
func TestRunWithRetry_Docker(t *testing.T) {
	requireLocalImage(t, pythonImage)
	tempfail := writeScript(t, "t.py", "#!/usr/bin/env -S fragletc --image ofthemachine/python3\n#: d=t\nimport sys\nsys.stderr.write('HTTP 429: slow down\\n')\nsys.exit(75)\n")
	var stderr bytes.Buffer
	retries := 0
	report, err := RunWithRetry(context.Background(), RunOptions{Image: pythonImage, ScriptFile: tempfail, Stdout: io.Discard, Stderr: &stderr},
		Retry{Times: 1, Wait: 10 * time.Millisecond, Waiting: func(int, time.Duration) { retries++ }})
	if err != nil || report.ExitCode != ExitTempFail || retries != 1 {
		t.Fatalf("exit %d, err %v, retries %d", report.ExitCode, err, retries)
	}
	if report.StderrTail != "HTTP 429: slow down" || report.Stderr.Bytes == 0 {
		t.Fatalf("a failed run's receipt says why: tail %q, stderr %+v", report.StderrTail, report.Stderr)
	}

	slow := writeScript(t, "s.py", "#!/usr/bin/env -S fragletc --image ofthemachine/python3\n#: d=s\nimport time\ntime.sleep(60)\n")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	start := time.Now()
	report, err = RunWithRetry(ctx, RunOptions{Image: pythonImage, ScriptFile: slow, Stdout: io.Discard, Stderr: io.Discard}, Retry{})
	if err != nil || report.ExitCode != ExitTimeout || !strings.Contains(report.StderrTail, "deadline") {
		t.Fatalf("exit %d, err %v, tail %q", report.ExitCode, err, report.StderrTail)
	}
	if time.Since(start) > 20*time.Second {
		t.Fatalf("the deadline did not end the run: %s", time.Since(start))
	}
}
