package engine

import (
	"os"
	"strings"
	"testing"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

func TestResolveSecrets(t *testing.T) {
	decls := []fraglet.SecretDecl{{Name: "FT_SECRET_A"}}

	t.Setenv("FT_SECRET_A", "value-a")
	secrets, env, err := resolveSecrets(decls, nil, os.LookupEnv)
	if err != nil {
		t.Fatal(err)
	}
	if len(secrets) != 1 || secrets[0].Path != "/run/fraglet/secrets/FT_SECRET_A" || string(secrets[0].Value) != "value-a" {
		t.Fatalf("secrets = %+v", secrets)
	}
	if len(env) != 1 || env[0] != "FT_SECRET_A_FILE=/run/fraglet/secrets/FT_SECRET_A" {
		t.Fatalf("env = %v", env)
	}
	for _, e := range env {
		if strings.Contains(e, "value-a") {
			t.Fatalf("secret value leaked into env: %q", e)
		}
	}

	if _, _, err := resolveSecrets(decls, []string{"FT_SECRET_A"}, os.LookupEnv); err == nil {
		t.Error("forwarding a declared secret with -e must be refused")
	}
	if _, _, err := resolveSecrets(decls, []string{"FT_SECRET_A_FILE=/etc/passwd"}, os.LookupEnv); err == nil {
		t.Error("overriding NAME_FILE with -e must be refused")
	}

	t.Setenv("FT_SECRET_A", "")
	if _, _, err := resolveSecrets(decls, nil, os.LookupEnv); err == nil || !strings.Contains(err.Error(), "FT_SECRET_A") {
		t.Errorf("empty secret must be reported missing, got %v", err)
	}
	if _, _, err := resolveSecrets([]fraglet.SecretDecl{{Name: "bad"}}, nil, os.LookupEnv); err == nil {
		t.Error("invalid name must be refused")
	}
}

// A library caller supplies the values: Execute's lookup reads only its
// map, never the process environment.
func TestResolveSecretsFromCaller(t *testing.T) {
	decls := []fraglet.SecretDecl{{Name: "FT_SECRET_B"}}
	t.Setenv("FT_SECRET_B", "from-env")
	given := map[string]string{"FT_SECRET_B": "from-caller", "UNDECLARED": "x"}
	lookup := func(n string) (string, bool) { v, ok := given[n]; return v, ok }
	secrets, env, err := resolveSecrets(decls, []string{"OTHER=1"}, lookup)
	if err != nil || len(secrets) != 1 || string(secrets[0].Value) != "from-caller" || len(env) != 1 {
		t.Fatalf("secrets = %+v, env = %v, err = %v; want the caller's value only", secrets, env, err)
	}
	if _, _, err := resolveSecrets(decls, nil, func(string) (string, bool) { return "", false }); err == nil {
		t.Error("a declared secret the caller lacks must be refused, not read from the environment")
	}
	if _, _, err := resolveSecrets(decls, []string{"FT_SECRET_B=leak"}, lookup); err == nil {
		t.Error("Env setting a declared secret must be refused")
	}
}
