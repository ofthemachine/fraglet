package fraglet

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// runKey starts the header line that says how a script's body runs.
const runKey = "run="

// codeOpen and codeClose delimit the one placeholder a run line carries:
// {code:<path>} is where the body is written in the container.
const (
	codeOpen  = "{code:"
	codeClose = "}"
)

// RunRecipe is how a script's body runs in an image that has no fraglet
// entrypoint, from its header's "#: run=" line:
//
//	#: run=python3 {code:/work/main.py}
//
// The body is written to CodePath and Argv runs, replacing the image's own
// entrypoint. Words are split on white space and never quoted: a command
// that needs a shell names one (sh {code:/work/run.sh}). A script without
// the line runs as its image's fraglet entrypoint runs it.
type RunRecipe struct {
	// Argv is the command, its placeholder already replaced by CodePath.
	Argv []string
	// CodePath is where the body is written: absolute, and named by the
	// script, so no run depends on a path fragletc chose.
	CodePath string
}

// ParseRun returns the recipe code's header declares, nil when it declares
// none. The line is the whole of its directive: everything after "run=" is
// the command, so no other directive shares it.
func ParseRun(code string) (*RunRecipe, error) {
	header, _ := SplitHeader(code)
	var lines []string
	for _, line := range strings.Split(header, "\n") {
		if v, ok := runLine(line); ok {
			lines = append(lines, v)
		}
	}
	switch len(lines) {
	case 0:
		return nil, nil
	case 1:
		return parseRun(lines[0])
	}
	return nil, errors.New("run=: a script says how it runs once (one #: run= line)")
}

// runLine returns what follows "run=" on a directive line that is a run line.
func runLine(line string) (string, bool) {
	rest, ok := rawDirective(line)
	if !ok {
		return "", false
	}
	return strings.CutPrefix(strings.TrimSpace(rest), runKey)
}

func parseRun(value string) (*RunRecipe, error) {
	argv := strings.Fields(value)
	if len(argv) == 0 {
		return nil, errors.New("run=: no command (run=<program> … {code:<path>})")
	}
	if strings.Contains(argv[0], codeOpen) {
		return nil, errors.New("run=: the first word is the program, not the body ({code:<path>} comes after it)")
	}
	r := &RunRecipe{Argv: argv}
	for i, word := range argv {
		before, rest, found := strings.Cut(word, codeOpen)
		if !found {
			continue
		}
		p, after, closed := strings.Cut(rest, codeClose)
		if !closed {
			return nil, fmt.Errorf("run=: %q: an unclosed {code:<path>}", word)
		}
		if r.CodePath != "" || strings.Contains(after, codeOpen) {
			return nil, errors.New("run=: the body is written to one place (one {code:<path>})")
		}
		if !path.IsAbs(p) || path.Clean(p) != p || p == "/" {
			return nil, fmt.Errorf("run=: {code:%s}: an absolute, clean file path in the container (/work/main.py)", p)
		}
		r.CodePath, argv[i] = p, before+p+after
	}
	if r.CodePath == "" {
		return nil, errors.New("run=: the command names where the body is written ({code:<path>})")
	}
	return r, nil
}
