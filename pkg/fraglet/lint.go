package fraglet

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Severity classifies a lint finding. Errors are grammar problems that make
// fragletc misread or silently drop part of a declaration; warnings are
// conventions that a header can violate without fragletc noticing.
type Severity int

const (
	Warning Severity = iota
	Error
)

func (s Severity) String() string {
	if s == Error {
		return "error"
	}
	return "warning"
}

// Finding is one lint result. Line is 1-based within the script; 0 means the
// finding is about the file as a whole rather than a specific line.
type Finding struct {
	Line     int
	Rule     string
	Severity Severity
	Message  string
}

// paramAliasRe is the alias shape every declared param must have: it becomes
// an env var by uppercasing, so it must be a valid identifier; the single
// optional dotted suffix is the `source.tex` idiom for file params whose
// container path needs an extension.
var paramAliasRe = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z0-9]+)?$`)

// knownParamModifiers is the closed modifier vocabulary parseParamToken gives
// meaning to. The bool says whether the modifier takes a value.
var knownParamModifiers = map[string]bool{
	"required": false,
	"optional": false,
	"file":     false,
	"default":  true,
	"envvar":   true,
}

// descTrailingModifiers are modifier spellings that, found inside a
// description's prose, mean the author put description= before another
// modifier — parseParamToken has swallowed that modifier into the prose.
var descTrailingModifiers = []string{":required", ":optional", ":file", ":default=", ":envvar="}

// Lint checks a fraglet script's header against fragletc's own grammar and
// the conventions that keep a header honest about its body. It is a pure
// function of the script text; callers attach the path.
func Lint(code string) []Finding {
	var out []Finding
	header, body := SplitHeader(code)
	lines := strings.Split(header, "\n")

	seenParams := make(map[string]int) // alias → first line
	seenOutputs := make(map[string]int)
	sawNetwork := false

	for i, line := range lines {
		lineNo := i + 1
		rest, ok := directiveLine(line)
		if !ok {
			continue
		}

		for _, tok := range paramTokens(rest) {
			out = append(out, lintParamToken(tok, lineNo, seenParams, body)...)
		}

		for _, field := range strings.Fields(rest) {
			switch {
			case strings.HasPrefix(field, "output="):
				out = append(out, lintOutputToken(field[len("output="):], lineNo, seenOutputs)...)
			case strings.HasPrefix(field, "network="):
				sawNetwork = true
				v := strings.ToLower(strings.TrimSpace(field[len("network="):]))
				if v != "none" && v != "required" {
					out = append(out, Finding{lineNo, "network-value", Error,
						fmt.Sprintf("network=%q is not a declaration fragletc understands; use network=none or network=required", v)})
				}
			}
		}
	}

	if !sawNetwork {
		out = append(out, Finding{0, "network-missing", Warning,
			"no #: network= line; declare network=none (enforced hermetic) or network=required"})
	}
	if ParseMetaDescription(code) == "" {
		out = append(out, Finding{0, "desc-missing", Warning,
			"no tool-level #: d= (or description=) line"})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		return out[i].Rule < out[j].Rule
	})
	return out
}

// lintParamToken checks one "alias[:modifier...]" body exactly as
// parseParamToken will read it, then checks that the script body actually
// reads the declared env var. Nothing more is inferred from the body: with
// hundreds of target languages, per-language idiom checks are not sustainable.
func lintParamToken(tok string, lineNo int, seen map[string]int, body string) []Finding {
	var out []Finding

	// Split off the description the same way parseParamToken does so the
	// structural part and the prose part are judged separately.
	structural := tok
	desc, hasDesc := "", false
	if i := strings.Index(tok, ":description="); i >= 0 {
		desc, hasDesc = tok[i+len(":description="):], true
		structural = tok[:i]
	} else if i := strings.Index(tok, ":d="); i >= 0 {
		desc, hasDesc = tok[i+len(":d="):], true
		structural = tok[:i]
	}

	parts := strings.Split(structural, ":")
	alias := parts[0]
	if alias == "" {
		return []Finding{{lineNo, "param-alias", Error, fmt.Sprintf("param=%q has no alias", tok)}}
	}

	mods := make(map[string]string)
	for _, part := range parts[1:] {
		name, value, hasValue := part, "", false
		if eq := strings.Index(part, "="); eq >= 0 {
			name, value, hasValue = part[:eq], part[eq+1:], true
		}
		wantsValue, known := knownParamModifiers[name]
		switch {
		case !known:
			out = append(out, Finding{lineNo, "param-unknown-modifier", Error,
				fmt.Sprintf("param=%s: modifier %q is not one fragletc understands (required, optional, file, default=, envvar=, description=/d=)", alias, part)})
		case wantsValue && !hasValue:
			out = append(out, Finding{lineNo, "param-unknown-modifier", Error,
				fmt.Sprintf("param=%s: modifier %q needs a value (%s=...)", alias, part, name)})
		case !wantsValue && hasValue:
			out = append(out, Finding{lineNo, "param-unknown-modifier", Error,
				fmt.Sprintf("param=%s: modifier %q does not take a value", alias, part)})
		}
		mods[name] = value
	}

	if !paramAliasRe.MatchString(alias) {
		out = append(out, Finding{lineNo, "param-alias", Error,
			fmt.Sprintf("param=%s: alias must match %s (it becomes the env var %s)", alias, paramAliasRe, strings.ToUpper(alias))})
	} else if _, isFile := mods["file"]; strings.Contains(alias, ".") && !isFile {
		out = append(out, Finding{lineNo, "param-alias", Error,
			fmt.Sprintf("param=%s: a dotted alias is only meaningful with :file (it names the mount path /input/%s)", alias, alias)})
	}

	if first, dup := seen[alias]; dup {
		out = append(out, Finding{lineNo, "param-duplicate", Error,
			fmt.Sprintf("param=%s already declared on line %d; fragletc keeps the first and drops this one", alias, first)})
		return out
	}
	seen[alias] = lineNo

	_, required := mods["required"]
	_, optional := mods["optional"]
	_, hasDefault := mods["default"]
	if required && hasDefault {
		out = append(out, Finding{lineNo, "param-required-with-default", Error,
			fmt.Sprintf("param=%s: default= silently exempts required; declare one or the other", alias)})
	}
	if optional {
		out = append(out, Finding{lineNo, "param-optional-redundant", Warning,
			fmt.Sprintf("param=%s: optional is the default and does nothing; omit it", alias)})
	}

	if hasDesc {
		for _, m := range descTrailingModifiers {
			if strings.Contains(desc, m) {
				out = append(out, Finding{lineNo, "param-desc-not-last", Error,
					fmt.Sprintf("param=%s: description must be the last modifier; %q was swallowed into the prose", alias, strings.TrimPrefix(m, ":"))})
				break
			}
		}
	}
	if !hasDesc || strings.TrimSpace(desc) == "" {
		out = append(out, Finding{lineNo, "param-no-description", Warning,
			fmt.Sprintf("param=%s has no description; add :d=<what the caller should pass>", alias)})
	}

	envVar := strings.ToUpper(alias)
	if ev, ok := mods["envvar"]; ok && ev != "" {
		envVar = ev
	}
	if body != "" && !strings.Contains(body, envVar) {
		out = append(out, Finding{lineNo, "param-unused", Warning,
			fmt.Sprintf("param=%s is declared but the body never references %s", alias, envVar)})
	}
	return out
}

// lintOutputToken checks one "relpath[:modifier...]" output declaration.
func lintOutputToken(tok string, lineNo int, seen map[string]int) []Finding {
	decl := parseOutputToken(tok)
	rel := decl.RelPath
	if rel == "" {
		return []Finding{{lineNo, "output-relpath", Error, fmt.Sprintf("output=%q has no path", tok)}}
	}
	var out []Finding
	if strings.HasPrefix(rel, "/") {
		out = append(out, Finding{lineNo, "output-relpath", Error,
			fmt.Sprintf("output=%s must be relative to /output, not absolute", rel)})
	}
	for _, seg := range strings.Split(rel, "/") {
		if seg == ".." {
			out = append(out, Finding{lineNo, "output-relpath", Error,
				fmt.Sprintf("output=%s must not escape /output with ..", rel)})
			break
		}
	}
	if first, dup := seen[rel]; dup {
		out = append(out, Finding{lineNo, "output-relpath", Error,
			fmt.Sprintf("output=%s already declared on line %d", rel, first)})
	} else {
		seen[rel] = lineNo
	}
	return out
}
