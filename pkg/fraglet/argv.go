package fraglet

import (
	"fmt"
	"strconv"
	"strings"
)

// ExpandArgv turns a fraglet body — one shell-*looking* command line, no
// more — into a literal argv for direct exec.Command, with zero shell
// involved. It exists for containers with no interpreter at all (FROM
// scratch: a single static binary plus fraglet-entrypoint): the body still
// reads like a normal invocation of the target binary, but every word is
// produced by this tokenizer/expander, never handed to /bin/sh.
//
// Supported, deliberately minimal:
//   - whitespace-separated words
//   - 'single quotes'   — literal, no expansion
//   - "double quotes"   — $VAR/${VAR}/$@/$1../$9 expanded inside
//   - bare $VAR/${VAR}/$@/$1..$9 outside quotes, each expanding to its own
//     word(s) (unquoted $@ and "$@" behave the same here: each positional
//     arg is always its own token, since there is no field-splitting step
//     to collapse them back together)
//
// Deliberately unsupported: pipes, redirects, `;`/`&&`/`||`, command
// substitution, globbing — anything that would need a real shell to mean
// what it looks like. ExpandArgv errors on the first such character rather
// than silently producing a plausible-looking but wrong argv.
func ExpandArgv(body string, positional []string, lookupEnv func(string) (string, bool)) ([]string, error) {
	line, err := singleLogicalLine(body)
	if err != nil {
		return nil, err
	}

	var argv []string
	var cur strings.Builder
	haveCur := false
	flush := func() {
		if haveCur {
			argv = append(argv, cur.String())
			cur.Reset()
			haveCur = false
		}
	}

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch {
		case c == ' ' || c == '\t':
			flush()

		case c == '\'':
			haveCur = true
			j := i + 1
			for j < len(runes) && runes[j] != '\'' {
				cur.WriteRune(runes[j])
				j++
			}
			if j >= len(runes) {
				return nil, fmt.Errorf("argv: unterminated '")
			}
			i = j

		case c == '"':
			// "$@" is special-cased as a standalone quoted span — the one
			// real-world shape ("$@" wrapping the whole quoted string) that
			// must still expand to one token per positional arg, exactly
			// like a real shell's "$@". Anything else quoted is joined into
			// one literal token, including a $@ embedded alongside other
			// text — full POSIX field-splicing semantics aren't worth the
			// complexity for what this mode is for.
			if i+3 < len(runes) && runes[i+1] == '$' && runes[i+2] == '@' && runes[i+3] == '"' {
				flush()
				argv = append(argv, positional...)
				i += 3
				continue
			}

			haveCur = true
			j := i + 1
			for j < len(runes) && runes[j] != '"' {
				if runes[j] == '$' {
					words, consumed, err := expandDollar(runes, j, positional, lookupEnv)
					if err != nil {
						return nil, err
					}
					cur.WriteString(strings.Join(words, " "))
					j += consumed
					continue
				}
				cur.WriteRune(runes[j])
				j++
			}
			if j >= len(runes) {
				return nil, fmt.Errorf("argv: unterminated \"")
			}
			i = j

		case c == '$':
			words, consumed, err := expandDollar(runes, i, positional, lookupEnv)
			if err != nil {
				return nil, err
			}
			for k, w := range words {
				if k > 0 {
					flush()
				}
				haveCur = true
				cur.WriteString(w)
			}
			i += consumed - 1

		case strings.ContainsRune("|;&`<>*?()", c):
			return nil, fmt.Errorf("argv: unsupported shell character %q — this mode execs the target binary directly with no shell, so pipes/redirects/substitution/globbing aren't available; use the script/injection mode if you need them", c)

		default:
			haveCur = true
			cur.WriteRune(c)
		}
	}
	flush()

	if len(argv) == 0 {
		return nil, fmt.Errorf("argv: empty command line")
	}
	return argv, nil
}

// singleLogicalLine returns body with its fraglet-meta header stripped and
// verifies what's left is exactly one non-blank line (argv mode has no
// concept of a multi-statement body — that's what script/injection mode is
// for).
func singleLogicalLine(body string) (string, error) {
	_, rest := SplitHeader(body)
	var line string
	found := false
	for _, l := range strings.Split(rest, "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if found {
			return "", fmt.Errorf("argv: body must be a single command line, found a second non-blank line: %q", l)
		}
		line = l
		found = true
	}
	if !found {
		return "", fmt.Errorf("argv: empty command line")
	}
	return line, nil
}

// expandDollar expands the $-expression starting at runes[i] (runes[i] ==
// '$'). Returns the resulting word(s) — usually one, but $@ expands to one
// word per positional arg — and how many runes were consumed.
func expandDollar(runes []rune, i int, positional []string, lookupEnv func(string) (string, bool)) ([]string, int, error) {
	if i+1 >= len(runes) {
		return []string{"$"}, 1, nil
	}

	switch next := runes[i+1]; {
	case next == '@':
		return append([]string{}, positional...), 2, nil

	case next >= '1' && next <= '9':
		idx, _ := strconv.Atoi(string(next))
		if idx <= len(positional) {
			return []string{positional[idx-1]}, 2, nil
		}
		return []string{""}, 2, nil

	case next == '{':
		j := i + 2
		for j < len(runes) && runes[j] != '}' {
			j++
		}
		if j >= len(runes) {
			return nil, 0, fmt.Errorf("argv: unterminated ${")
		}
		name := string(runes[i+2 : j])
		val, _ := lookupEnv(name)
		return []string{val}, j - i + 1, nil

	case isIdentStart(next):
		j := i + 1
		for j < len(runes) && isIdentPart(runes[j]) {
			j++
		}
		name := string(runes[i+1 : j])
		val, _ := lookupEnv(name)
		return []string{val}, j - i, nil

	default:
		return []string{"$"}, 1, nil
	}
}

func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}
