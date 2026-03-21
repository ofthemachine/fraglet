package params

import (
	"os"
	"sort"
	"strings"
)

const transportPrefix = "FRAGLET_PARAM_"

// EnvVar is a name=value pair extracted from FRAGLET_PARAM_* transport vars.
type EnvVar struct {
	Name  string // bare env var name (prefix stripped, case-preserved)
	Value string // value as-is (caller already decoded before transport)
}

// Coerce scans the process environment for FRAGLET_PARAM_* vars,
// strips the prefix, applies the no-shadow rule, and returns bare env var pairs.
// Transport vars are unset after processing.
// The entrypoint is dumb: no decoding, no schema, no fraglet-meta parsing.
// Decoding is the caller's responsibility (fragletc decodes before setting transport).
func Coerce() ([]EnvVar, error) {
	// Collect all FRAGLET_PARAM_* entries (snapshot before mutation)
	type entry struct {
		transport string // FRAGLET_PARAM_CITY
		bare      string // CITY
		value     string // london
	}
	var entries []entry
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, transportPrefix) {
			continue
		}
		eqIdx := strings.Index(env, "=")
		if eqIdx < 0 {
			continue
		}
		key := env[:eqIdx]
		val := env[eqIdx+1:]
		bare := key[len(transportPrefix):]
		if bare == "" {
			continue
		}
		entries = append(entries, entry{transport: key, bare: bare, value: val})
	}

	// Sort for deterministic ordering
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].bare < entries[j].bare
	})

	var result []EnvVar
	for _, e := range entries {
		// Always clean up transport var
		os.Unsetenv(e.transport)

		// No-shadow: if bare env var already exists, skip
		if _, exists := os.LookupEnv(e.bare); exists {
			continue
		}

		result = append(result, EnvVar{Name: e.bare, Value: e.value})
	}
	return result, nil
}
