package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

func handleFragletHelp(scriptFile, inlineCode string) {
	code := inlineCode
	if code == "" && scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", scriptFile, err)
			os.Exit(1)
		}
		code = string(data)
	}
	if code == "" {
		fmt.Fprintf(os.Stderr, "Error: --fraglet-help requires a script file or -c code\n")
		os.Exit(1)
	}

	decls := fraglet.ParseParamDecls(code)
	desc := fraglet.ParseMetaDescription(code)
	when := fraglet.ParseMetaWhen(code)
	label := fragletHelpLabel(scriptFile)

	if desc != "" {
		fmt.Fprintf(os.Stdout, "%s\n\n", desc)
	}
	if when != "" {
		fmt.Fprintf(os.Stdout, "When: %s\n\n", when)
	}

	if len(decls) == 0 {
		fmt.Printf("No parameters declared in %s.\n", label)
		fmt.Fprintf(os.Stdout, "\nAdd param= under fraglet-meta to list names here; optional description= (or d=) and when= on their own fraglet-meta lines.\n")
		return
	}

	fmt.Printf("Parameters for %s:\n", label)
	writeParamList(os.Stdout, decls)
	printFragletInvokeHint(label)
}

// writeParamList prints one line per declared param (required/optional,
// default, env var override) — the body of --fraglet-help's listing, also
// reused by the missing-required-param error so both surfaces show the
// exact same shape.
func writeParamList(w io.Writer, decls []fraglet.ParamDecl) {
	for _, d := range decls {
		var parts []string
		if d.IsRequired() {
			parts = append(parts, "required")
		} else {
			parts = append(parts, "optional")
		}
		if def, ok := d.Default(); ok {
			parts = append(parts, "default: "+def)
		}
		modStr := strings.Join(parts, ", ")
		desc := ""
		if text, ok := d.Description(); ok {
			desc = " — " + text
		}
		fmt.Fprintf(w, "  %-12s (%s)%s%s\n", d.Alias, modStr, envVarArrow(d), desc)
	}
}

// validateParams checks, host-side and before any container runs, that
// paramStrs actually resolves against code's declared params: every alias
// is known, and every param= declared "required" (and without a default= —
// a default already satisfies "the caller must supply this", see
// fraglet.MissingRequired) has a value.
//
// Without this, a missing required param silently expands to an empty
// string wherever the fraglet body references it (argv mode's ExpandArgv
// and script-mode env vars both treat an unset var as ""), and whatever
// error surfaces comes from deep inside the wrapped tool instead of from
// fragletc itself — e.g. a confusing usage dump for a CLI the caller never
// invoked directly.
//
// engine.Run repeats this same parse+resolve internally to build the
// container's transport env — that redundant pass is deliberate, not an
// oversight: engine.RunOptions.ParamStrs is the library's public shape,
// which callers outside this binary construct directly, so re-typing it
// as pre-resolved Params would ripple into every consumer for the sake of
// skipping a few microseconds of work on a handful of strings. What must
// not exist twice is the *policy* — the "required unless defaulted" rule
// lives exactly once, in fraglet.MissingRequired.
func validateParams(code string, paramStrs []string, label string) error {
	decls := fraglet.ParseParamDecls(code)

	var params fraglet.Params
	for _, s := range paramStrs {
		p, err := fraglet.ParseParam(s)
		if err != nil {
			return err
		}
		params = append(params, p)
	}
	params, err := params.ResolveAliases(decls)
	if err != nil {
		return err
	}

	missing := fraglet.MissingRequired(decls, params)
	if len(missing) == 0 {
		return nil
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "missing required parameter(s): %s\n\n", strings.Join(missing, ", "))
	fmt.Fprintf(&buf, "Parameters for %s:\n", label)
	writeParamList(&buf, decls)
	fmt.Fprintf(&buf, "\nSee --fraglet-help for the full description.")
	return errors.New(buf.String())
}

func fragletHelpLabel(scriptFile string) string {
	if scriptFile == "" {
		return "<inline>"
	}
	return filepath.Base(scriptFile)
}

func envVarArrow(d fraglet.ParamDecl) string {
	defaultEnv := strings.ToUpper(d.Alias)
	if d.EnvVar != defaultEnv {
		return fmt.Sprintf("    → %s", d.EnvVar)
	}
	return ""
}

func printFragletInvokeHint(label string) {
	if label == "<inline>" {
		fmt.Fprintf(os.Stdout, "\nPass: fragletc --vein=<vein> -p name=value ... -c '<code>'\n")
		return
	}
	fmt.Fprintf(os.Stdout, "\nPass: ./%s -p name=value ...  (repeat -p per parameter; see fragletc --help)\n", label)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: fragletc [flags] [script-file] [script-args...]
       fragletc refresh [options] [vein-name]
       fragletc lint [--strict] <script-or-dir>...

Execute fraglet code in a container using either --vein or --image.

Flags:
  -v, --vein string
        Vein name with optional mode (e.g., python, python:main)
  -i, --image string
        Container image to use directly (e.g., my-registry/python:latest)
  -c, --code string
        Program passed in as string (like python -c)
  -p, --param (preprocessed, not a flag)
        Fraglet-meta parameters as KEY=value (repeatable; any position before "--"). Forms include
        -p K=V, --param K=V, -p=K=V, --param=K=V, and -pK=V when '=' appears in the suffix.
        Optional encodings: -p key=b64:...  See --fraglet-help on a script for its declarations.
  --fraglet-path string
        Path where code is mounted in container (default: /FRAGLET; long form only)
  --output relpath[=hostdest] (preprocessed, not a flag)
        Copy one declared output= file out after a successful run (repeatable; any position
        before "--"). relpath must match a fraglet-meta output= declaration. hostdest defaults
        to ./relpath. Never a directory dump — always a named, explicit file.
        If the fraglet declares exactly one output=, relpath may be omitted: a bare
        "--output <dest>" is shorthand for "the one declared output, saved as <dest>".
        /output is mounted whenever the fraglet declares any output=, whether or not --output
        is passed — a script that writes its declared output always runs. Without --output the
        file is produced then discarded; a stderr note names what you could have copied out.
  --output-dir string
        Copy everything the fraglet writes to /output into this host directory after the run,
        instead of the declared-output=/--output dance — for wrapping a real CLI tool whose
        output filename isn't known ahead of time (an explicit -o flag it's passed, or its own
        default naming). No declaration needed; same-named files are overwritten, same as a
        locally installed tool would. Mutually exclusive with --output.
  --receipt path (preprocessed, not a flag)
        Any position before "--", like -p and --output. Write a JSON receipt of the run to path.
        Without it, FRAGLETC_RECEIPT_DIR=<dir> saves every run's receipt there (named by start
        time, script, and memo key); unset, receipts are computed but not saved. A receipt is the
        invocation -- procedure hash (sha256 of the script file, shebang included), effective
        image and digest, params, inputs by content hash (stdin, :file params), argv, -e names --
        and the outcome: exit code, stdout digest, outputs (every file under /output, or stdout
        when none is declared). Its memo key -- the documented sha256 over procedure hash,
        params and inputs -- is present only when nothing was bound outside the declared
        parameters (no argv, no -e, stdin not streamed), so it identifies the computation.
  -e string
        Environment variable to forward into container (repeatable)
        Use -e FOO to forward host value, -e FOO=bar for explicit value
  --fraglet-help
        Show parameter declarations from fraglet-meta and exit (may appear before or after script-file).
        Like -p/--param, removed from argv before your program runs (any position before "--").
        After "--", --fraglet-help and -p/--param pass through unchanged.
  -m, --mode string
        Fraglet mode (sets FRAGLET_MODE=mode)
  --network string
        Container network mode (e.g. 'none', 'bridge', 'host'; defaults to 'none' when '#: network=none' is declared, otherwise Docker default bridge)

Positional:
  script-file   Path to code file (required if -c not set)
  script-args   Tail arguments for your program inside the container

First, -p/--param/--fraglet-help are removed from argv anywhere before a bare "--". Then normal
flags (-v, -c, …) are parsed and must come before script-file. Example: ./tool.py -p city=paris
--profile prod strips -p; --profile and prod are program argv. Use "--" so -p/--param/--fraglet-help
are not stripped.

Stdin:
  Stdin is an input like any other. A terminal stdin is never forwarded; a pipe or file is:
    #: stdin=buffer     (default) read to EOF, hashed into the receipt, forwarded:  cat data.csv | ./process.py
    #: stdin=stream     live passthrough for interactive programs; nothing to hash, so no memo key.
    #: stdin=none       the program does not read stdin; nothing is attached.
  --stdin=<mode> overrides the header for one run; like -p and --output it is preprocessed, so
  it works in any position before "--".

Subcommands:
  refresh       Refresh (pull) container images for veins
                Use "fragletc refresh --help" for details
  guide         Show fraglet guide (vein registry or --image; flags and vein in any order)
                Use "fragletc guide --help" for details
  essence       Show fraglet essence (vein registry or --image; flags and vein in any order)
                Use "fragletc essence --help" for details
  version       Show build version, commit, and lineage info
  lint          Check fraglet headers against fragletc's grammar and conventions
                Use "fragletc lint --help" for the rule list
`)
}
