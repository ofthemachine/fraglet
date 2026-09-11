package testutil

import (
	"fmt"
	"strings"
)

// Unindent strips the indent of the first non-blank line from every line.
// Blank lines are kept (as empty lines) and do not participate in the indent
// check. A non-blank line that does not start with that indent is jagged-left
// and panics: the fixture is misaligned, not a value to paper over.
func Unindent(s string) string {
	out, err := unindent(s)
	if err != nil {
		panic(err)
	}
	return out
}

func unindent(s string) (string, error) {
	lines := strings.Split(s, "\n")

	indent := ""
	found := false
	for _, line := range lines {
		if isBlank(line) {
			continue
		}
		indent = leadingWhitespace(line)
		found = true
		break
	}
	if !found {
		return "", nil
	}

	out := make([]string, 0, len(lines))
	for i, line := range lines {
		if isBlank(line) {
			out = append(out, "")
			continue
		}
		if !strings.HasPrefix(line, indent) {
			return "", fmt.Errorf("unindent: jagged-left on line %d %q (indent is %q)", i+1, line, indent)
		}
		out = append(out, line[len(indent):])
	}

	// Drop wrapping blank lines from the usual backtick-block form:
	//
	//	unindent(`
	//		contents
	//	`)
	for len(out) > 0 && out[0] == "" {
		out = out[1:]
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return strings.Join(out, "\n"), nil
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func leadingWhitespace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[:i]
}
