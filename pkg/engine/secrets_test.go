package engine

import (
	"strings"
	"testing"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
)

func TestResolveSecrets(t *testing.T) {
	decls := []fraglet.SecretDecl{{Name: "FT_SECRET_A"}}

	t.Setenv("FT_SECRET_A", "value-a")
	secrets, env, err := resolveSecrets(decls, nil)
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

	if _, _, err := resolveSecrets(decls, []string{"FT_SECRET_A"}); err == nil {
		t.Error("forwarding a declared secret with -e must be refused")
	}
	if _, _, err := resolveSecrets(decls, []string{"FT_SECRET_A_FILE=/etc/passwd"}); err == nil {
		t.Error("overriding NAME_FILE with -e must be refused")
	}

	t.Setenv("FT_SECRET_A", "")
	if _, _, err := resolveSecrets(decls, nil); err == nil || !strings.Contains(err.Error(), "FT_SECRET_A") {
		t.Errorf("empty secret must be reported missing, got %v", err)
	}
	if _, _, err := resolveSecrets([]fraglet.SecretDecl{{Name: "bad"}}, nil); err == nil {
		t.Error("invalid name must be refused")
	}
}
