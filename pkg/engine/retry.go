package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"
)

// ExitTempFail is the exit code a fraglet uses to say its failure is
// temporary (sysexits' EX_TEMPFAIL): a rate limit, a server down for a
// moment. Only it is retried, so a fraglet that is not safe to run twice
// never exits with it: the exit code is the opt-in.
const ExitTempFail = 75

// ExitTimeout is the exit code of a run its deadline ended, as timeout(1)
// has it.
const ExitTimeout = 124

// Retry says how often to run a fraglet again while it exits ExitTempFail.
type Retry struct {
	// Times is how many more runs after the first; 0 runs once.
	Times int
	// Wait is the wait before the first retry, doubling before each next
	// one, each jittered to between half and all of it; 0 means a second.
	Wait time.Duration
	// Waiting, when set, hears each wait before it starts (retry counts
	// from 1).
	Waiting func(retry int, wait time.Duration)
}

// RunWithRetry is RunWithReport, run again while the fraglet exits
// ExitTempFail and retries remain. A retry is just another run: the report
// is the last run's, and of the earlier ones nothing is kept but what they
// wrote to stderr. ctx's deadline bounds every run and every wait: a run it
// ends reports ExitTimeout (and no error), and a wait that would pass it is
// not started. With retries, each run reads the same stdin, starts with an
// empty OutputHostDir, and only the last run's stdout reaches opts.Stdout.
func RunWithRetry(ctx context.Context, opts RunOptions, retry Retry) (RunReport, error) {
	if retry.Times <= 0 {
		return timedOut(ctx)(runOnce(ctx, opts))
	}
	var stdin []byte
	if opts.Stdin != nil {
		b, err := io.ReadAll(opts.Stdin)
		if err != nil {
			return RunReport{}, fmt.Errorf("stdin: %w", err)
		}
		stdin = b
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	wait := retry.Wait
	if wait <= 0 {
		wait = time.Second
	}
	for n := 0; ; n++ {
		var buf bytes.Buffer
		run := opts
		run.Stdout = &buf
		if stdin != nil {
			run.Stdin = bytes.NewReader(stdin)
		}
		if n > 0 && opts.OutputHostDir != "" {
			if err := emptyDir(opts.OutputHostDir); err != nil {
				return RunReport{}, fmt.Errorf("retry: %w", err)
			}
		}
		report, err := timedOut(ctx)(runOnce(ctx, run))
		last := err != nil || report.ExitCode != ExitTempFail || n == retry.Times
		pause := wait/2 + rand.N(wait/2+1)
		if dl, ok := ctx.Deadline(); ok && time.Until(dl) < pause {
			last = true
		}
		if last {
			_, werr := stdout.Write(buf.Bytes())
			return report, errors.Join(err, werr)
		}
		if retry.Waiting != nil {
			retry.Waiting(n+1, pause)
		}
		select {
		case <-ctx.Done():
			_, werr := stdout.Write(buf.Bytes())
			return report, werr
		case <-time.After(pause):
		}
		wait *= 2
	}
}

// runOnce is one run; a test stands in for it.
var runOnce = RunWithReport

// timedOut makes a run ctx's deadline ended report ExitTimeout and no
// error, so it still has a receipt saying what happened.
func timedOut(ctx context.Context) func(RunReport, error) (RunReport, error) {
	return func(report RunReport, err error) (RunReport, error) {
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return report, err
		}
		report.ExitCode = ExitTimeout
		if report.StderrTail != "" {
			report.StderrTail += "\n"
		}
		report.StderrTail += "fragletc: the run's deadline passed"
		return report, nil
	}
}

// emptyDir removes everything in dir, keeping dir.
func emptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
