package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

// handleLint implements `fragletc lint [--strict] <path>...`: check fraglet
// headers against fragletc's own grammar (errors) and the conventions that
// keep a header honest (warnings). Exit 0 when clean, 1 on any error — or
// any warning under --strict — and 2 on usage.
func handleLint() {
	lintFlags := flag.NewFlagSet("lint", flag.ExitOnError)
	strict := lintFlags.Bool("strict", false, "Treat warnings as errors")
	lintFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: fragletc lint [--strict] <script-or-dir>...

Lint fraglet headers (#: lines) against fragletc's own grammar and conventions.

A file named explicitly is always linted. A directory is walked, and only
files whose shebang invokes fragletc are linted (hidden directories are
skipped), so pointing lint at a tools tree lints exactly the scripts
fragletc would run.

Errors are declarations fragletc misreads or silently drops:
  param-desc-not-last          description=/d= must be the last modifier
  param-unknown-modifier       only required, optional, file, default=, envvar=, description=/d=
  param-required-with-default  default= silently exempts required
  param-alias                  alias must be [a-z][a-z0-9_]* (dotted suffix only with :file)
  param-duplicate              same alias declared twice
  network-value                network= must be none or required
  output-relpath               output= relative to /output, no .., no duplicates

Warnings are conventions (errors under --strict):
  param-no-description         every param should say what to pass
  param-optional-redundant     optional is the default; omit it
  param-unused                 the body never references the param's env var
  network-missing              declare network=none or network=required
  desc-missing                 no tool-level d= line

Options:
  --strict    Exit 1 on warnings too

Examples:
  fragletc lint tools/weather/forecast.py
  fragletc lint --strict tools/
`)
	}

	lintFlags.Parse(os.Args[2:])
	paths := lintFlags.Args()
	if len(paths) == 0 {
		lintFlags.Usage()
		os.Exit(2)
	}

	files, err := collectLintFiles(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	var errorsN, warningsN int
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			errorsN++
			continue
		}
		for _, f := range fraglet.Lint(string(data)) {
			sev := f.Severity
			if *strict && sev == fraglet.Warning {
				sev = fraglet.Error
			}
			if sev == fraglet.Error {
				errorsN++
			} else {
				warningsN++
			}
			loc := path
			if f.Line > 0 {
				loc = fmt.Sprintf("%s:%d", path, f.Line)
			}
			fmt.Printf("%s: %s: %s: %s\n", loc, sev, f.Rule, f.Message)
		}
	}

	noun := "files"
	if len(files) == 1 {
		noun = "file"
	}
	switch {
	case errorsN == 0 && warningsN == 0:
		fmt.Printf("%d %s linted, clean\n", len(files), noun)
	case *strict:
		fmt.Printf("%d %s linted, %d errors (--strict: warnings count as errors)\n", len(files), noun, errorsN)
	default:
		fmt.Printf("%d %s linted, %d errors, %d warnings\n", len(files), noun, errorsN, warningsN)
	}
	if errorsN > 0 {
		os.Exit(1)
	}
}

// collectLintFiles expands the caller's paths: files are taken as given;
// directories are walked for fragletc-shebang files, skipping hidden
// directories. The result is sorted so output is stable.
func collectLintFiles(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != p && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if hasFragletcShebang(path) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

// hasFragletcShebang reports whether the file's first line is a shebang that
// invokes fragletc — the same test the skills catalog uses to decide which
// files under tools/ are fraglets.
func hasFragletcShebang(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	first, err := bufio.NewReader(f).ReadString('\n')
	if err != nil && first == "" {
		return false
	}
	return strings.HasPrefix(first, "#!") && strings.Contains(first, "fragletc")
}
