package fraglet

import (
	"reflect"
	"testing"
)

func TestParseSecretDecls(t *testing.T) {
	code := `#!/usr/bin/env -S fragletc --image x
#: d=Tool. It mentions secret=NOT_A_DECL in prose.
#: secret=HA_TOKEN:d=Home Assistant long-lived token: with a colon
#: secret=API_KEY
#: secret=HA_TOKEN:d=duplicate is dropped
#: param=secret:d=a param named secret is not a secret
print(open(os.environ["HA_TOKEN_FILE"]).read())
#: secret=BODY_LINE_IS_NOT_HEADER
`
	got := ParseSecretDecls(code)
	want := []SecretDecl{
		{Name: "HA_TOKEN", Description: "Home Assistant long-lived token: with a colon"},
		{Name: "API_KEY"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if p := got[0].ContainerPath(); p != "/run/fraglet/secrets/HA_TOKEN" {
		t.Errorf("ContainerPath = %q", p)
	}
	if e := got[0].FileEnv(); e != "HA_TOKEN_FILE" {
		t.Errorf("FileEnv = %q", e)
	}
}

func TestValidSecretName(t *testing.T) {
	for name, want := range map[string]bool{
		"HA_TOKEN": true, "K": true, "K2": true,
		"": false, "ha_token": false, "_X": false, "2X": false,
		"FRAGLET_PARAM_X": false, "HA_TOKEN:env": false, "A-B": false,
	} {
		if got := ValidSecretName(name); got != want {
			t.Errorf("ValidSecretName(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestLintSecrets(t *testing.T) {
	code := `#!/usr/bin/env -S fragletc --image x
#: d=Tool
#: when=Use when testing
#: network=required
#: secret=HA_TOKEN:d=token
#: secret=HA_TOKEN:d=again
#: secret=lower:d=bad name
#: secret=VIA_ENV:env:d=modifiers are not grammar
#: secret=UNDESCRIBED
#: secret=UNUSED:d=never read
#: secret=CITY:d=clashes with the param
#: param=city:required:d=City
print(os.environ["HA_TOKEN_FILE"], os.environ["UNDESCRIBED_FILE"], os.environ["CITY"])
`
	rules := map[string]int{}
	for _, f := range Lint(code) {
		rules[f.Rule]++
	}
	want := map[string]int{
		"secret-duplicate":       1,
		"secret-name":            2,
		"secret-no-description":  1,
		"secret-unused":          1,
		"secret-param-collision": 1,
	}
	for rule, n := range want {
		if rules[rule] != n {
			t.Errorf("rule %s: got %d findings, want %d (all: %v)", rule, rules[rule], n, rules)
		}
	}
}

func TestLintSecretClean(t *testing.T) {
	code := `#!/usr/bin/env -S fragletc --image x
#: d=Tool
#: when=Use when testing
#: network=required
#: secret=HA_TOKEN:d=Home Assistant token
token = open(os.environ["HA_TOKEN_FILE"]).read()
`
	if f := Lint(code); len(f) != 0 {
		t.Fatalf("expected clean, got %+v", f)
	}
}
