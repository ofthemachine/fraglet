package fraglet

import (
	"sort"
	"strings"
)

// shortSentinel is the primary, language-agnostic directive marker: a line
// (after trimming leading whitespace) that starts with "#:" carries a
// fraglet directive. fragletMetaSentinel is the verbose legacy form, kept as
// a recognized (non-default) alias — it matches anywhere the literal
// substring "fraglet-meta:" appears on a header line, mirroring its
// historical behavior.
const shortSentinel = "#:"
const fragletMetaSentinel = "fraglet-meta:"

// SplitHeader splits code into its header and body. The header is the
// mandatory shebang line plus the maximal contiguous run of subsequent lines
// that are each either blank or start with "#" (after trimming leading
// whitespace); the first non-blank, non-"#"-prefixed line ends the header
// and everything from there on is body. Header-scanning never resumes, so a
// later body line that happens to start with "#" (a genuine in-language
// comment) is never mistaken for header.
//
// The header is fragletc's own domain: engine.Execute strips it entirely
// before mounting code into a container, so header lines never need to look
// like valid comments in whatever language the body is written in. This is
// what lets a single "#:" sentinel work uniformly across every target
// language instead of adapting per comment syntax.
func SplitHeader(code string) (header, body string) {
	lines := strings.Split(code, "\n")
	end := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			end = i + 1
			continue
		}
		break
	}
	return strings.Join(lines[:end], "\n"), strings.Join(lines[end:], "\n")
}

// directiveLine returns the text following a directive sentinel on line, and
// whether one was found. The short sentinel must be the line's first
// non-whitespace characters; the legacy verbose sentinel matches anywhere in
// the line (its historical behavior).
func directiveLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, shortSentinel) {
		return trimmed[len(shortSentinel):], true
	}
	if idx := strings.Index(line, fragletMetaSentinel); idx >= 0 {
		return line[idx+len(fragletMetaSentinel):], true
	}
	return "", false
}

// ParamDecl represents a declared parameter from the fraglet header.
type ParamDecl struct {
	Alias     string            // user-facing name: "city", "host"
	EnvVar    string            // resolved env var: "CITY", "HURL_VARIABLE_host"
	Shape     string            // "string" (default) or "file"
	Modifiers map[string]string // "required" → "", "default" → "metric", "envvar" → "HURL_VARIABLE_host"
}

// IsRequired returns true if the param has a "required" modifier.
func (d ParamDecl) IsRequired() bool {
	_, ok := d.Modifiers["required"]
	return ok
}

// IsOptional returns true if the param has an "optional" modifier or no required modifier.
func (d ParamDecl) IsOptional() bool {
	_, req := d.Modifiers["required"]
	return !req
}

// Default returns the default value and whether one was declared.
func (d ParamDecl) Default() (string, bool) {
	v, ok := d.Modifiers["default"]
	return v, ok
}

// Description returns the param-level description (description= or d= modifier)
// and whether one was declared.
func (d ParamDecl) Description() (string, bool) {
	if v, ok := d.Modifiers["description"]; ok && v != "" {
		return v, true
	}
	if v, ok := d.Modifiers["d"]; ok && v != "" {
		return v, true
	}
	return "", false
}

// OutputDecl represents a declared output file from the fraglet header: the
// script writes this file, relative to the output mount, when it runs.
type OutputDecl struct {
	RelPath   string
	Modifiers map[string]string // "required" → "", "optional" → ""
}

// IsRequired returns true if the output has a "required" modifier.
func (d OutputDecl) IsRequired() bool {
	_, ok := d.Modifiers["required"]
	return ok
}

// ParseParamDecls extracts param= tokens from a code string's header.
// Returns declarations sorted by alias for determinism.
//
// A param token runs from "param=" to the next " param=" or end of line, so
// description= values may contain spaces (and must be the last modifier):
//
//	#: param=settle_ms:default=2000:description=Extra wait after page load
func ParseParamDecls(code string) []ParamDecl {
	header, _ := SplitHeader(code)
	var decls []ParamDecl
	seen := make(map[string]bool)

	for _, line := range strings.Split(header, "\n") {
		rest, ok := directiveLine(line)
		if !ok {
			continue
		}
		for _, tok := range paramTokens(rest) {
			decl := parseParamToken(tok)
			if decl.Alias == "" || seen[decl.Alias] {
				continue
			}
			seen[decl.Alias] = true
			decls = append(decls, decl)
		}
	}

	sort.Slice(decls, func(i, j int) bool {
		return decls[i].Alias < decls[j].Alias
	})
	return decls
}

// paramTokens splits a directive rest into param declaration bodies (the
// text after each "param="). Tokens are delimited by " param=".
//
// Without description=/d=, a param body is a single Fields token so
// trailing annotations on the same line (e.g. "param=city math:algebra")
// stay out of the alias. With description= or d= (must be last), the body
// runs to the next " param=" or end of line so the prose may contain spaces.
func paramTokens(rest string) []string {
	const prefix = "param="
	var toks []string
	for {
		i := indexFieldPrefix(rest, prefix)
		if i < 0 {
			break
		}
		rest = rest[i+len(prefix):]
		end := len(rest)
		if j := strings.Index(rest, " "+prefix); j >= 0 {
			end = j
		}
		if tok := takeParamBody(rest[:end]); tok != "" {
			toks = append(toks, tok)
		}
		rest = rest[end:]
	}
	return toks
}

// indexFieldPrefix finds prefix at the start of a whitespace-delimited field.
func indexFieldPrefix(s, prefix string) int {
	if strings.HasPrefix(s, prefix) {
		return 0
	}
	for i := 0; i < len(s); i++ {
		if (s[i] == ' ' || s[i] == '\t') && strings.HasPrefix(s[i+1:], prefix) {
			return i + 1
		}
	}
	return -1
}

// takeParamBody returns one param declaration body from a chunk that may
// still contain trailing same-line annotations when description is absent.
func takeParamBody(chunk string) string {
	chunk = strings.TrimSpace(chunk)
	if chunk == "" {
		return ""
	}
	if strings.Contains(chunk, ":description=") || strings.Contains(chunk, ":d=") {
		return chunk
	}
	return strings.Fields(chunk)[0]
}

// ParseOutputDecls extracts output= tokens from a code string's header.
// Returns declarations sorted by relpath for determinism.
func ParseOutputDecls(code string) []OutputDecl {
	header, _ := SplitHeader(code)
	var decls []OutputDecl
	seen := make(map[string]bool)

	for _, line := range strings.Split(header, "\n") {
		rest, ok := directiveLine(line)
		if !ok {
			continue
		}
		for _, tok := range strings.Fields(rest) {
			if !strings.HasPrefix(tok, "output=") {
				continue
			}
			decl := parseOutputToken(tok[len("output="):])
			if decl.RelPath == "" || seen[decl.RelPath] {
				continue
			}
			seen[decl.RelPath] = true
			decls = append(decls, decl)
		}
	}

	sort.Slice(decls, func(i, j int) bool {
		return decls[i].RelPath < decls[j].RelPath
	})
	return decls
}

// ParseTags extracts tags= tokens from a code string's header: a
// comma-separated free-text label list, deduped and sorted.
func ParseTags(code string) []string {
	header, _ := SplitHeader(code)
	var tags []string
	seen := make(map[string]bool)

	for _, line := range strings.Split(header, "\n") {
		rest, ok := directiveLine(line)
		if !ok {
			continue
		}
		for _, tok := range strings.Fields(rest) {
			if !strings.HasPrefix(tok, "tags=") {
				continue
			}
			for _, tag := range strings.Split(tok[len("tags="):], ",") {
				tag = strings.TrimSpace(tag)
				if tag == "" || seen[tag] {
					continue
				}
				seen[tag] = true
				tags = append(tags, tag)
			}
		}
	}

	sort.Strings(tags)
	return tags
}

// ParseNetwork extracts the network declaration from a code string's header
// (e.g. "#: network=none" or "#: network=required").
// Returns the declared network mode (e.g. "none", "required") or "" if not declared.
// When multiple declarations appear, the last non-empty one wins.
func ParseNetwork(code string) string { return lastDirectiveValue(code, "network=") }

// lastDirectiveValue returns the lowercased value of the last "key=value"
// token on any fraglet-meta line, or "" when no line carries the key.
func lastDirectiveValue(code, key string) string {
	header, _ := SplitHeader(code)
	var value string
	for _, line := range strings.Split(header, "\n") {
		rest, ok := directiveLine(line)
		if !ok {
			continue
		}
		for _, tok := range strings.Fields(rest) {
			if strings.HasPrefix(tok, key) {
				if v := strings.ToLower(strings.TrimSpace(tok[len(key):])); v != "" {
					value = v
				}
			}
		}
	}
	return value
}

// ParseMetaDescription returns human-oriented text from header lines that are
// only description=... or the short form d=... (one line per block; multiple
// lines are joined with a blank line). Use a dedicated meta line per
// paragraph. Multiline values inside a single description are a future
// format extension.
func ParseMetaDescription(code string) string {
	header, _ := SplitHeader(code)
	var parts []string
	for _, line := range strings.Split(header, "\n") {
		rest, ok := directiveLine(line)
		if !ok {
			continue
		}
		rest = strings.TrimSpace(rest)
		var v string
		switch {
		case strings.HasPrefix(rest, "description="):
			v = strings.TrimSpace(rest[len("description="):])
		case strings.HasPrefix(rest, "d="):
			v = strings.TrimSpace(rest[len("d="):])
		default:
			continue
		}
		if v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, "\n\n")
}

// Stdin modes a header may declare with "#: stdin=". Stdin is an input like
// any other: StdinBuffer (the default when undeclared) reads a non-terminal
// stdin to EOF, hashes it into the receipt, and forwards it; StdinStream
// passes it through live for interactive programs, which leaves nothing to
// hash, so such a run has no memo key; StdinNone asserts the program does
// not read stdin, so nothing is attached.
const (
	StdinNone   = "none"
	StdinBuffer = "buffer"
	StdinStream = "stream"
)

// ParseStdin returns the declared stdin mode, or "" when the header has no
// stdin= line (callers treat that as StdinBuffer). The last non-empty
// declaration wins, like ParseNetwork.
func ParseStdin(code string) string { return lastDirectiveValue(code, "stdin=") }

// ParseMetaWhen returns the tool-level trigger clause: the situations in which
// an agent should reach for this tool, from "#: when=" lines. d= says what the
// tool does; when= says when to use it. Catalogs that publish a tool as an
// agent skill append this to the skill description verbatim, so it should be
// written as a sentence an agent can match a request against ("Use when the
// user wants ..."). Multiple lines join with a space.
func ParseMetaWhen(code string) string {
	header, _ := SplitHeader(code)
	var parts []string
	for _, line := range strings.Split(header, "\n") {
		rest, ok := directiveLine(line)
		if !ok {
			continue
		}
		rest = strings.TrimSpace(rest)
		if !strings.HasPrefix(rest, "when=") {
			continue
		}
		if v := strings.TrimSpace(rest[len("when="):]); v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, " ")
}

// parseParamToken parses "alias[:modifier[:modifier...]]" into a ParamDecl.
// description= (or short d=) must be last: everything after that marker is
// the description, including spaces and colons.
func parseParamToken(s string) ParamDecl {
	mods := make(map[string]string)
	if i := strings.Index(s, ":description="); i >= 0 {
		mods["description"] = s[i+len(":description="):]
		s = s[:i]
	} else if i := strings.Index(s, ":d="); i >= 0 {
		mods["d"] = s[i+len(":d="):]
		s = s[:i]
	}

	parts := strings.Split(s, ":")
	alias := parts[0]
	if alias == "" {
		return ParamDecl{}
	}

	for _, part := range parts[1:] {
		if eqIdx := strings.Index(part, "="); eqIdx >= 0 {
			mods[part[:eqIdx]] = part[eqIdx+1:]
		} else {
			mods[part] = ""
		}
	}

	// Resolve env var: explicit envvar= modifier, or alias uppercased
	envVar := strings.ToUpper(alias)
	if ev, ok := mods["envvar"]; ok {
		envVar = ev
	}

	shape := "string"
	if _, ok := mods["file"]; ok {
		shape = "file"
	}

	return ParamDecl{
		Alias:     alias,
		EnvVar:    envVar,
		Shape:     shape,
		Modifiers: mods,
	}
}

// parseOutputToken parses "relpath[:modifier[:modifier...]]" into an OutputDecl.
func parseOutputToken(s string) OutputDecl {
	parts := strings.Split(s, ":")
	relPath := parts[0]
	if relPath == "" {
		return OutputDecl{}
	}

	mods := make(map[string]string)
	for _, part := range parts[1:] {
		if eqIdx := strings.Index(part, "="); eqIdx >= 0 {
			mods[part[:eqIdx]] = part[eqIdx+1:]
		} else {
			mods[part] = ""
		}
	}

	return OutputDecl{RelPath: relPath, Modifiers: mods}
}
