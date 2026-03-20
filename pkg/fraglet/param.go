package fraglet

import (
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"sort"
	"strings"
)

const transportPrefix = "FRAGLET_PARAM_"

// reserved env var names that cannot be used as param targets.
var reserved = map[string]bool{"CONFIG": true}

// Param represents a typed parameter for fraglet execution.
// EnvVar is the resolved env var name (e.g. "CITY", "HURL_VARIABLE_host").
// Encoding is "raw", "b64", or "cb64". Value is the encoded form.
type Param struct {
	EnvVar   string // resolved env var name
	Encoding string // "raw", "b64", "cb64"
	Value    string // encoded value as provided
}

// ParseParam parses "KEY=value" or "KEY=type:value" into a Param.
// The key is treated as the env var name (uppercased).
// Use ResolveAliases to apply fraglet-meta envvar= mappings afterward.
func ParseParam(s string) (Param, error) {
	eqIdx := strings.Index(s, "=")
	if eqIdx < 0 {
		return Param{}, fmt.Errorf("param %q: missing '=' separator", s)
	}
	name := s[:eqIdx]
	raw := s[eqIdx+1:]
	if name == "" {
		return Param{}, fmt.Errorf("param %q: empty name", s)
	}

	envVar := strings.ToUpper(name)
	if reserved[envVar] {
		return Param{}, fmt.Errorf("param %q: reserved env var name %q", s, envVar)
	}

	encoding, value := parseEncodedValue(raw)
	return Param{EnvVar: envVar, Encoding: encoding, Value: value}, nil
}

// parseEncodedValue splits "type:value" or returns ("raw", value).
func parseEncodedValue(s string) (encoding, value string) {
	// Check for known encoding prefixes
	for _, enc := range []string{"raw:", "b64:", "cb64:"} {
		if strings.HasPrefix(s, enc) {
			return enc[:len(enc)-1], s[len(enc):]
		}
	}
	return "raw", s
}

// Decode returns the decoded value (applies b64/cb64 decoding).
func (p Param) Decode() (string, error) {
	switch p.Encoding {
	case "raw", "":
		return p.Value, nil
	case "b64":
		data, err := base64.StdEncoding.DecodeString(p.Value)
		if err != nil {
			return "", fmt.Errorf("b64 decode %q: %w", p.EnvVar, err)
		}
		return string(data), nil
	case "cb64":
		compressed, err := base64.StdEncoding.DecodeString(p.Value)
		if err != nil {
			return "", fmt.Errorf("cb64 base64 decode %q: %w", p.EnvVar, err)
		}
		r, err := zlib.NewReader(strings.NewReader(string(compressed)))
		if err != nil {
			return "", fmt.Errorf("cb64 zlib open %q: %w", p.EnvVar, err)
		}
		defer r.Close()
		data, err := io.ReadAll(r)
		if err != nil {
			return "", fmt.Errorf("cb64 zlib read %q: %w", p.EnvVar, err)
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unknown encoding %q for %q", p.Encoding, p.EnvVar)
	}
}

// TransportEnvName returns "FRAGLET_PARAM_" + EnvVar.
func (p Param) TransportEnvName() string {
	return transportPrefix + p.EnvVar
}

// TransportEnvValue returns the encoded form: "raw:london" or "b64:SGVsbG8=".
func (p Param) TransportEnvValue() string {
	enc := p.Encoding
	if enc == "" {
		enc = "raw"
	}
	return enc + ":" + p.Value
}

// Canonical returns the deterministic form for hashing/ledger: "CITY=raw:london".
func (p Param) Canonical() string {
	return p.EnvVar + "=" + p.TransportEnvValue()
}

// Params is a slice of Param with collection-level operations.
type Params []Param

// ToTransportEnv returns sorted "FRAGLET_PARAM_X=decoded_value" pairs for docker -e.
// Values are decoded caller-side so the entrypoint can coerce (strip prefix, set bare names).
// Images built as generic language containers (e.g. upstream python) may not run that entrypoint;
// coercion behavior is covered in entrypoint/tests. Returns an error if any value fails to decode.
func (ps Params) ToTransportEnv() ([]string, error) {
	if len(ps) == 0 {
		return nil, nil
	}
	out := make([]string, len(ps))
	for i, p := range ps {
		decoded, err := p.Decode()
		if err != nil {
			return nil, err
		}
		out[i] = p.TransportEnvName() + "=" + decoded
	}
	sort.Strings(out)
	return out, nil
}

// ToCanonical returns sorted "key=type:value" pairs for hashing/ledger.
func (ps Params) ToCanonical() []string {
	if len(ps) == 0 {
		return nil
	}
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Canonical()
	}
	sort.Strings(out)
	return out
}

// ResolveAliases applies fraglet-meta param declarations to map aliases to env var names.
// Called host-side by the runner before building transport env vars.
// If decls is empty, params pass through with default uppercased env var names.
func (ps Params) ResolveAliases(decls []ParamDecl) (Params, error) {
	if len(decls) == 0 {
		return ps, nil
	}

	// Build alias → decl lookup
	byAlias := make(map[string]ParamDecl, len(decls))
	for _, d := range decls {
		byAlias[d.Alias] = d
	}

	resolved := make(Params, len(ps))
	for i, p := range ps {
		// The param's EnvVar at this point is the uppercased alias from ParseParam.
		// Try to find a matching declaration by lowercase alias.
		alias := strings.ToLower(p.EnvVar)
		if d, ok := byAlias[alias]; ok {
			resolved[i] = Param{EnvVar: d.EnvVar, Encoding: p.Encoding, Value: p.Value}
		} else {
			// Also try the original casing
			if d, ok := byAlias[p.EnvVar]; ok {
				resolved[i] = Param{EnvVar: d.EnvVar, Encoding: p.Encoding, Value: p.Value}
			} else {
				return nil, fmt.Errorf("param %q: not declared in fraglet-meta", alias)
			}
		}
	}
	return resolved, nil
}
