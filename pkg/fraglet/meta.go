package fraglet

import (
	"sort"
	"strings"
)

const fragletMetaSentinel = "fraglet-meta:"

// ParamDecl represents a declared parameter from fraglet-meta.
type ParamDecl struct {
	Alias     string            // user-facing name: "city", "host"
	EnvVar    string            // resolved env var: "CITY", "HURL_VARIABLE_host"
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

// ParseParamDecls extracts param= tokens from a code string.
// Scans all lines for the "fraglet-meta:" sentinel, collects param= tokens.
// Returns declarations sorted by alias for determinism.
func ParseParamDecls(code string) []ParamDecl {
	var decls []ParamDecl
	seen := make(map[string]bool)

	for _, line := range strings.Split(code, "\n") {
		idx := strings.Index(line, fragletMetaSentinel)
		if idx < 0 {
			continue
		}
		// Everything after the sentinel
		rest := line[idx+len(fragletMetaSentinel):]
		tokens := strings.Fields(rest)
		for _, tok := range tokens {
			if !strings.HasPrefix(tok, "param=") {
				continue
			}
			decl := parseParamToken(tok[len("param="):])
			if decl.Alias == "" {
				continue
			}
			if seen[decl.Alias] {
				continue // dedup
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

// ParseMetaDescription returns human-oriented text from fraglet-meta lines that are only
// description=... or the short form d=... (one line per block; multiple lines are joined
// with a blank line). Use a dedicated meta line per paragraph.
// Multiline values inside a single description are a future format extension.
func ParseMetaDescription(code string) string {
	var parts []string
	for _, line := range strings.Split(code, "\n") {
		idx := strings.Index(line, fragletMetaSentinel)
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(line[idx+len(fragletMetaSentinel):])
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

// parseParamToken parses "alias[:modifier[:modifier...]]" into a ParamDecl.
func parseParamToken(s string) ParamDecl {
	parts := strings.Split(s, ":")
	alias := parts[0]
	if alias == "" {
		return ParamDecl{}
	}

	mods := make(map[string]string)
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

	return ParamDecl{
		Alias:     alias,
		EnvVar:    envVar,
		Modifiers: mods,
	}
}
