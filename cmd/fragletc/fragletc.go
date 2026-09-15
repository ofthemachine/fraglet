package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/engine"
	"github.com/ofthemachine/fraglet/pkg/receipt"
)

const defaultFragletPath = "/FRAGLET"

// envListFlag implements flag.Value for repeatable -e flags.
type envListFlag []string

func (e *envListFlag) String() string { return strings.Join(*e, ",") }

func (e *envListFlag) Set(val string) error {
	*e = append(*e, val)
	return nil
}

func main() {
	// Subcommands are checked before flag parsing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "refresh":
			handleRefresh()
			return
		case "guide":
			handleGuide()
			return
		case "essence":
			handleEssence()
			return
		case "version":
			handleVersion()
			return
		case "lint":
			handleLint()
			return
		}
	}

	flag.Usage = usage

	// Flags
	veinSpec := flag.String("vein", "", "Vein name with optional mode (e.g., python, python:main)")
	image := flag.String("image", "", "Container image to use directly")
	fragletPath := flag.String("fraglet-path", defaultFragletPath, "Path where code is mounted in container")
	mode := flag.String("mode", "", "Fraglet mode (sets FRAGLET_MODE=mode)")
	network := flag.String("network", "", "Container network mode (e.g. 'none', 'bridge', 'host'; defaults to 'none' when '#: network=none' is declared, otherwise Docker default bridge)")
	inlineCode := flag.String("c", "", "Program passed in as string (like python -c)")
	outputDir := flag.String("output-dir", "", "Copy everything the fraglet writes to /output into this host directory after the run — for a fraglet whose output filename isn't known ahead of time (e.g. wrapping a real CLI tool's own -o/default naming). No declared-output= needed: whatever lands in /output is copied out, overwriting same-named files, same as a locally installed tool would. Mutually exclusive with --output.")
	var envFlags envListFlag
	flag.Var(&envFlags, "e", "Environment variable to forward (repeatable, e.g. -e FOO -e BAR=val)")

	// Short forms
	flag.StringVar(veinSpec, "v", "", "Vein name with optional mode (short form)")
	flag.StringVar(image, "i", "", "Container image (short form)")
	flag.StringVar(mode, "m", "", "Fraglet mode (short form)")
	flag.StringVar(inlineCode, "code", "", "Program passed in as string (like python -c)")

	// Preprocess argv for params, output requests, and help
	pre, err := preprocessFragletArgv(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	if err := flag.CommandLine.Parse(pre.filtered); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	args := flag.Args()
	var scriptFile string
	var scriptArgs []string
	if len(args) > 0 {
		scriptFile = args[0]
		scriptArgs = args[1:]
	}

	if pre.wantHelp {
		handleFragletHelp(scriptFile, *inlineCode)
		return
	}

	var stdinReader io.Reader
	if fi, err := os.Stdin.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		stdinReader = os.Stdin
	}

	outputRequests, err := parseOutputRequests(pre.outputs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	if *outputDir != "" && len(outputRequests) > 0 {
		fmt.Fprintln(os.Stderr, "--output-dir and --output are mutually exclusive: --output-dir already copies everything out, so there's nothing left to declare or copy individually")
		os.Exit(2)
	}

	declaredOutputs, outputRequests := preflightValidate(*inlineCode, scriptFile, pre.params, *outputDir, outputRequests)

	// finalOutputDir is set for --output-dir: the caller's own real
	// directory, which fragletc never mounts or chmods directly (it isn't
	// fragletc's to manage). Instead the container always writes into
	// fragletc's own disposable scratch mount below, and that gets copied
	// into finalOutputDir after the run — the same "our own scratch dir,
	// caller's directory touched only by writing named files into it" shape
	// declared-output/--output already uses (copyRequestedOutputs), just
	// copying everything instead of specific declared relpaths.
	var finalOutputDir string
	var outputHostDir string
	if *outputDir != "" || len(declaredOutputs) > 0 || len(outputRequests) > 0 {
		dir, err := os.MkdirTemp("", "fragletc-output-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(dir)
		// Containers run with --cap-drop=all (pkg/runner/docker.go), which
		// strips CAP_DAC_OVERRIDE: root inside the container no longer
		// bypasses host file permission checks the way an unconstrained
		// root would, so a freshly created 0700 temp dir is unwritable by
		// whatever UID the container's root maps to on the host. Since
		// this is fragletc's own throwaway scratch directory (removed
		// above), opening it up is free — unlike a caller-supplied
		// --output-dir, which fragletc must never chmod.
		if err := os.Chmod(dir, 0o777); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		outputHostDir = dir

		if *outputDir != "" {
			abs, err := filepath.Abs(*outputDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "--output-dir %q: %v\n", *outputDir, err)
				os.Exit(2)
			}
			if err := os.MkdirAll(abs, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "--output-dir %q: %v\n", *outputDir, err)
				os.Exit(1)
			}
			finalOutputDir = abs
		}
	}

	opts := engine.RunOptions{
		VeinSpec:      *veinSpec,
		Image:         *image,
		FragletPath:   *fragletPath,
		Mode:          *mode,
		InlineCode:    *inlineCode,
		EnvFlags:      envFlags,
		ScriptFile:    scriptFile,
		ScriptArgs:    scriptArgs,
		Stdin:         stdinReader,
		ParamStrs:     pre.params,
		NetworkMode:   *network,
		OutputHostDir: outputHostDir,
		StdinMode:     pre.stdin,
	}

	report, err := engine.RunWithReport(context.Background(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	exitCode := report.ExitCode

	if finalOutputDir != "" {
		// Copy regardless of exitCode, mirroring what a live mount would
		// have shown: whatever the container wrote before succeeding or
		// failing is what the caller gets, including partial output from a
		// failed run useful for debugging.
		if err := copyOutputTreeContents(outputHostDir, finalOutputDir); err != nil {
			fmt.Fprintf(os.Stderr, "--output-dir: %v\n", err)
			os.Exit(1)
		}
	}

	if exitCode == 0 && finalOutputDir == "" {
		if len(outputRequests) > 0 {
			if err := copyRequestedOutputs(outputHostDir, outputRequests); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			// Confirm exactly what landed and under what host name. This
			// is the fraglet-meta relpath's ONLY host-visible destination
			// record: the script's own stdout (if it prints anything at
			// all) necessarily uses the declared relpath, never the
			// caller's --output=hostdest rename, so without this line a
			// renamed copy is invisible — a reader sees the script claim
			// "wrote diagram.png" and has no way to learn
			// the file that actually exists on disk is called foo.png.
			for _, r := range outputRequests {
				fmt.Fprintf(os.Stderr, "fragletc: copied %s -> %s\n", r.RelPath, r.Dest)
			}
		}
		// A declared output not among outputRequests was produced (per
		// declaredOutputs) but never copied out — that's true whether zero
		// or only some outputs were requested. Report every such gap, not
		// just the all-zero case, or a partial --output silently drops the
		// rest with no explanation.
		if undelivered := undeliveredOutputs(declaredOutputs, outputRequests); len(undelivered) > 0 {
			printUndeliveredOutputsHint(outputHostDir, undelivered)
		}
	}

	if dest := receiptDestination(pre.receipt, os.Getenv("FRAGLETC_RECEIPT_DIR"), scriptFile, report); dest != "" {
		r := receipt.Build(report.Invocation, report.Outcome, loadBuildInfo().Version, scriptFile)
		if err := r.Write(dest); err != nil {
			fmt.Fprintf(os.Stderr, "--receipt: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "fragletc: receipt -> %s\n", dest)
	}

	os.Exit(exitCode)
}
