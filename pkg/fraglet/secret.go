package fraglet

import (
	"regexp"
	"strings"
)

// SecretMountDir is the container directory where each declared secret is
// delivered as a read-only file: SecretMountDir + "/" + NAME.
const SecretMountDir = "/run/fraglet/secrets"

// secretNameRe is the shape a secret name must have: it is both the caller's
// host env var the value is read from and the file name inside the
// container. FRAGLET_ names are fragletc's own transport namespace.
var secretNameRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// SecretDecl is a credential the tool needs, declared on its own header line:
//
//	#: secret=HA_TOKEN:d=Home Assistant long-lived access token
//
// The value comes from the caller's env var of the same name and reaches the
// container only as the file SecretMountDir/NAME, whose path is in the env
// var NAME_FILE. The value is never a container env var, a param, or part of
// a receipt; the declaration itself is covered by procedure_hash.
type SecretDecl struct {
	Name        string
	Description string
}

// ContainerPath is where the secret's value is readable inside the container.
func (d SecretDecl) ContainerPath() string { return SecretMountDir + "/" + d.Name }

// FileEnv is the env var holding ContainerPath, after the Docker _FILE convention.
func (d SecretDecl) FileEnv() string { return d.Name + "_FILE" }

// ValidSecretName reports whether name can be a secret: an upper-case env var
// name outside fragletc's FRAGLET_ namespace.
func ValidSecretName(name string) bool {
	return secretNameRe.MatchString(name) && !strings.HasPrefix(name, "FRAGLET_")
}

// ParseSecretDecls returns the header's secret declarations in header order.
// A declaration is a whole directive line starting with "secret=", so a
// secret= inside another directive's prose is never one. Duplicate names keep
// the first; malformed names are returned as-is for callers to reject (lint
// reports them, the engine refuses to run).
func ParseSecretDecls(code string) []SecretDecl {
	header, _ := SplitHeader(code)
	var decls []SecretDecl
	seen := make(map[string]bool)
	for _, line := range strings.Split(header, "\n") {
		tok, ok := secretToken(line)
		if !ok {
			continue
		}
		decl := parseSecretToken(tok)
		if decl.Name == "" || seen[decl.Name] {
			continue
		}
		seen[decl.Name] = true
		decls = append(decls, decl)
	}
	return decls
}

// secretToken returns the text after "secret=" when line is a secret directive.
func secretToken(line string) (string, bool) {
	rest, ok := directiveLine(line)
	if !ok {
		return "", false
	}
	rest = strings.TrimSpace(rest)
	if !strings.HasPrefix(rest, "secret=") {
		return "", false
	}
	return rest[len("secret="):], true
}

// parseSecretToken parses "NAME[:d=prose]" (description= is accepted too).
// Anything else after NAME is kept in the name, so lint and the engine both
// see it as malformed rather than silently dropping it.
func parseSecretToken(s string) SecretDecl {
	var desc string
	if i := strings.Index(s, ":description="); i >= 0 {
		desc, s = s[i+len(":description="):], s[:i]
	} else if i := strings.Index(s, ":d="); i >= 0 {
		desc, s = s[i+len(":d="):], s[:i]
	}
	return SecretDecl{Name: strings.TrimSpace(s), Description: strings.TrimSpace(desc)}
}
