package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

// undeliveredOutputs returns the declared outputs not named by any request,
// in declaration order.
func undeliveredOutputs(declared []fraglet.OutputDecl, requests []outputRequest) []fraglet.OutputDecl {
	if len(declared) == 0 {
		return nil
	}
	requested := make(map[string]bool, len(requests))
	for _, r := range requests {
		requested[r.RelPath] = true
	}
	var undelivered []fraglet.OutputDecl
	for _, d := range declared {
		if !requested[d.RelPath] {
			undelivered = append(undelivered, d)
		}
	}
	return undelivered
}

// printUndeliveredOutputsHint runs after a successful execution when the
// fraglet declared output= but the caller didn't pass --output for any of
// them: the container-side files are about to be discarded along with the
// throwaway outputHostDir. Silently discarding a declared output would be
// its own POLA violation, so name what was produced and how to keep it.
func printUndeliveredOutputsHint(outputHostDir string, declared []fraglet.OutputDecl) {
	absRoot, err := filepath.Abs(outputHostDir)
	if err != nil {
		return
	}
	var written []string
	for _, d := range declared {
		if _, err := resolveOutputFile(absRoot, d.RelPath); err == nil {
			written = append(written, d.RelPath)
		}
	}
	if len(written) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "fragletc: discarded %d declared output(s) not requested: %s\n",
		len(written), strings.Join(written, ", "))
	fmt.Fprintf(os.Stderr, "fragletc: pass --output %s to copy one out\n", written[0])
}

// outputRequest is one --output <relpath>[=<hostdest>] token: copy the
// fraglet's declared output file at relpath (relative to its output mount)
// to hostdest (default "./<relpath>", relative to the caller's cwd) after a
// successful run. Never a directory dump — always a named, explicit file.
type outputRequest struct {
	RelPath string
	Dest    string
	// hadExplicitDest is true for "relpath=dest", false for a bare
	// "relpath" (dest defaulted). Distinguishes the two for
	// resolveSingleOutputShorthand below: a bare token that isn't a known
	// relpath might be shorthand for "the sole output's destination";
	// an explicit relpath=dest that doesn't match a known relpath is just
	// a mistake and should surface the normal "not declared" error.
	hadExplicitDest bool
}

func parseOutputRequests(raw []string) ([]outputRequest, error) {
	var reqs []outputRequest
	for _, s := range raw {
		relPath, dest, hadEq := strings.Cut(s, "=")
		relPath = strings.TrimSpace(relPath)
		if relPath == "" {
			return nil, fmt.Errorf("--output: empty relpath in %q", s)
		}
		if err := validateOutputRelPath(relPath); err != nil {
			return nil, fmt.Errorf("--output %q: %w", relPath, err)
		}
		if dest == "" {
			dest = "./" + relPath
		}
		reqs = append(reqs, outputRequest{RelPath: relPath, Dest: dest, hadExplicitDest: hadEq})
	}
	return reqs, nil
}

// resolveSingleOutputShorthand lets a bare "--output <dest>" (no "=", and
// not itself a declared relpath) mean "the fraglet's one declared output,
// saved as <dest>", when it declares exactly one. There is nothing to
// disambiguate in that case, so making the caller repeat the fraglet's own
// internal filename back to it is pure friction, not a safety check —
// left alone whenever more than one output is declared (genuinely
// ambiguous) or the request already names a real declared relpath, or the
// caller wrote an explicit "relpath=dest" (a mismatched explicit relpath is
// a real mistake and should surface the normal "not declared" error, not
// be silently reinterpreted).
func resolveSingleOutputShorthand(reqs []outputRequest, declared []fraglet.OutputDecl) []outputRequest {
	if len(declared) != 1 {
		return reqs
	}
	sole := declared[0].RelPath
	out := make([]outputRequest, len(reqs))
	for i, r := range reqs {
		if r.hadExplicitDest || r.RelPath == sole {
			out[i] = r
			continue
		}
		out[i] = outputRequest{RelPath: sole, Dest: r.RelPath, hadExplicitDest: true}
	}
	return out
}

// validateOutputRelPath rejects absolute paths and any relpath that could
// escape the output mount root.
func validateOutputRelPath(relPath string) error {
	if filepath.IsAbs(relPath) {
		return errors.New("relpath must be relative")
	}
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if clean == "." {
		return errors.New("relpath must name a file")
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return errors.New("relpath must not escape the output directory")
	}
	for _, part := range strings.Split(clean, string(filepath.Separator)) {
		if part == ".." {
			return errors.New("relpath must not contain ..")
		}
	}
	return nil
}

// preflightValidate loads code (if any) and enforces, before any container
// starts, everything code's header declares about its own contract:
// param=...:required (validateParams) and output=... (validateOutputRequests).
// Exits the process directly on violation, same as the rest of main()'s
// flag handling — there's no error to propagate past this point, only a
// process that either continues or has already stopped.
//
// Returns the declared outputs and outputRequests (resolveSingleOutputShorthand
// may have rewritten it); both come back unchanged when there's no code to
// load, or when code fails to load and no --output was requested (in which
// case engine.Run surfaces the real "no code source" / "file not found"
// error naturally once it tries to run).
func preflightValidate(inlineCode, scriptFile string, paramStrs []string, outputDir string, outputRequests []outputRequest) ([]fraglet.OutputDecl, []outputRequest) {
	if scriptFile == "" && inlineCode == "" {
		return nil, outputRequests
	}

	code, codeErr := loadCodeForValidation(inlineCode, scriptFile)
	if codeErr != nil {
		if len(outputRequests) > 0 {
			fmt.Fprintf(os.Stderr, "%v\n", codeErr)
			os.Exit(1)
		}
		return nil, outputRequests
	}

	// Required-param validation applies regardless of --output-dir — unlike
	// declared-output tracking below, it has nothing to do with where the
	// container's output lands.
	if err := validateParams(code, paramStrs, fragletHelpLabel(scriptFile)); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	// --output-dir mounts /output live and copies out whatever lands there:
	// there's no fixed relpath to declare or validate against when the
	// fraglet is wrapping a real CLI tool that names its own output files.
	if outputDir != "" {
		return nil, outputRequests
	}

	declaredOutputs := fraglet.ParseOutputDecls(code)
	if len(outputRequests) > 0 {
		outputRequests = resolveSingleOutputShorthand(outputRequests, declaredOutputs)
		if err := validateOutputRequests(code, outputRequests); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
	}
	return declaredOutputs, outputRequests
}

// loadCodeForValidation reads the full source (inline or file) so --output
// can be validated against declared output= names before running anything.
func loadCodeForValidation(inlineCode, scriptFile string) (string, error) {
	if inlineCode != "" {
		return inlineCode, nil
	}
	if scriptFile != "" {
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			return "", fmt.Errorf("error reading file %s: %w", scriptFile, err)
		}
		return string(data), nil
	}
	return "", errors.New("--output requires a script file or -c code")
}

func validateOutputRequests(code string, reqs []outputRequest) error {
	declared := fraglet.ParseOutputDecls(code)
	names := make([]string, 0, len(declared))
	known := make(map[string]bool, len(declared))
	for _, d := range declared {
		if err := validateOutputRelPath(d.RelPath); err != nil {
			return fmt.Errorf("declared output %q: %w", d.RelPath, err)
		}
		known[d.RelPath] = true
		names = append(names, d.RelPath)
	}
	for _, req := range reqs {
		if !known[req.RelPath] {
			if len(names) == 0 {
				return fmt.Errorf("--output %q: no output= declared in this fraglet", req.RelPath)
			}
			return fmt.Errorf("--output %q: not declared (declared: %s)", req.RelPath, strings.Join(names, ", "))
		}
	}
	return nil
}

// copyOutputTreeContents copies everything under srcDir into destDir,
// preserving relative paths and overwriting anything already there at the
// same relative path. Used by --output-dir: unlike copyRequestedOutputs
// (specific declared relpaths, known ahead of time), the whole point here is
// that filenames are only known at runtime, so it copies whatever the
// container actually produced rather than named files. Overwrite matches
// what a live mount would have done — running the same fraglet twice
// against the same --output-dir naturally replaces same-named files, the
// same as invoking the wrapped tool locally would.
func copyOutputTreeContents(srcDir, destDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		destPath := filepath.Join(destDir, rel)
		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, 0o644)
	})
}

func copyRequestedOutputs(outputHostDir string, reqs []outputRequest) error {
	absRoot, err := filepath.Abs(outputHostDir)
	if err != nil {
		return fmt.Errorf("output directory: %w", err)
	}
	for _, req := range reqs {
		src, err := resolveOutputFile(absRoot, req.RelPath)
		if err != nil {
			return fmt.Errorf("--output %q: %w (the fraglet did not write it)", req.RelPath, err)
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("--output %q: %w", req.RelPath, err)
		}
		if dir := filepath.Dir(req.Dest); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("--output %q: %w", req.RelPath, err)
			}
		}
		if err := os.WriteFile(req.Dest, data, 0o644); err != nil {
			return fmt.Errorf("--output %q: %w", req.RelPath, err)
		}
	}
	return nil
}

// resolveOutputFile returns the absolute host path of one declared output file
// and ensures it stays within outputRoot and refers to a regular file.
func resolveOutputFile(outputRoot, relPath string) (string, error) {
	if err := validateOutputRelPath(relPath); err != nil {
		return "", err
	}
	absRoot, err := filepath.Abs(outputRoot)
	if err != nil {
		return "", err
	}
	canonicalRoot, err := filepath.EvalSymlinks(absRoot)
	if err != nil {
		return "", fmt.Errorf("output directory: %w", err)
	}
	candidate := filepath.Join(canonicalRoot, filepath.FromSlash(relPath))
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", err
	}
	absResolved, err := filepath.Abs(resolved)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(canonicalRoot, absResolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes output directory")
	}
	info, err := os.Stat(absResolved)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("not a regular file")
	}
	return absResolved, nil
}
