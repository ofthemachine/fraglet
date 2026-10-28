package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/receipt"
	"github.com/ofthemachine/fraglet/pkg/runner"
)

// A guard masks every secret value a stream carries, however the writes
// split it, notes each by name, and passes everything else on unchanged;
// uncensored notes without masking. A value inside a longer one is masked
// whole, never leaving the longer one's rest.
func TestSecretGuard(t *testing.T) {
	secrets := []runner.Secret{{Name: "PIN", Value: []byte("4242")}, {Name: "API_KEY", Value: []byte("sk-4242-0123")}, {Name: "EMPTY"}}
	text := "key sk-4242-0123 and sk-4242-0123, pin 4242.\nplain line\n"
	for _, split := range []int{1, 3, 5, len(text)} {
		for _, mode := range fraglet.SecretOutputModes {
			var out bytes.Buffer
			g := newSecretGuard(&out, "stdout", secrets, mode)
			for i := 0; i < len(text); i += split {
				if _, err := g.Write([]byte(text[i:min(i+split, len(text))])); err != nil {
					t.Fatal(err)
				}
			}
			if err := g.close(); err != nil {
				t.Fatal(err)
			}
			want := "key *** and ***, pin ***.\nplain line\n"
			if mode == fraglet.SecretOutputUncensored {
				want = text
			}
			if out.String() != want {
				t.Errorf("split %d, %s: wrote %q, want %q", split, mode, out.String(), want)
			}
			so := secretOutput(mode, secrets, g.seen)
			if !slices.Equal(so.Seen, []receipt.SecretSeen{{Secret: "API_KEY", In: "stdout"}, {Secret: "PIN", In: "stdout"}}) {
				t.Errorf("split %d, %s: seen %v, want API_KEY and PIN in stdout", split, mode, so.Seen)
			}
			if err := secretOutputErr(so); (err != nil) != (mode == fraglet.SecretOutputFail) {
				t.Errorf("%s: secretOutputErr = %v; want an error only in mode fail", mode, err)
			}
		}
	}
	var out bytes.Buffer
	g := newSecretGuard(&out, "stderr", secrets, fraglet.SecretOutputFail)
	if _, err := g.Write([]byte("nothing to see")); err != nil {
		t.Fatal(err)
	}
	if err := g.close(); err != nil {
		t.Fatal(err)
	}
	if out.String() != "nothing to see" || len(g.seen) != 0 || secretOutputErr(secretOutput(fraglet.SecretOutputFail, secrets, g.seen)) != nil {
		t.Errorf("a clean stream: wrote %q, seen %v; want it unchanged and no failure", out.String(), g.seen)
	}
	if secretOutput(fraglet.SecretOutputFail, nil, nil) != nil {
		t.Error("a run given no secrets has no secret_output")
	}
}

// Files a run left are read before they are hashed: masked in place unless
// uncensored, and each place a secret appeared named by its path.
func TestGuardOutputs(t *testing.T) {
	secrets := []runner.Secret{{Name: "API_KEY", Value: []byte("sk-0123456789")}}
	for _, mode := range fraglet.SecretOutputModes {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
			t.Fatal(err)
		}
		leak, clean := filepath.Join(dir, "sub", "a.txt"), filepath.Join(dir, "b.txt")
		for p, body := range map[string]string{leak: "x sk-0123456789 y", clean: "nothing"} {
			if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		seen, err := guardOutputs(dir, secrets, mode)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(seen, []receipt.SecretSeen{{Secret: "API_KEY", In: "output:sub/a.txt"}}) {
			t.Errorf("%s: seen %v; want API_KEY in output:sub/a.txt", mode, seen)
		}
		got, err := os.ReadFile(leak)
		if err != nil {
			t.Fatal(err)
		}
		want := "x *** y"
		if mode == fraglet.SecretOutputUncensored {
			want = "x sk-0123456789 y"
		}
		if string(got) != want {
			t.Errorf("%s: the file holds %q; want %q", mode, got, want)
		}
	}
}

// A mode fragletc does not have refuses the run: it never runs as another.
func TestSecretOutputMode(t *testing.T) {
	if mode, err := secretOutputMode("#: d=T\n"); err != nil || mode != fraglet.SecretOutputFail {
		t.Errorf("undeclared = %q, %v; want fail", mode, err)
	}
	if mode, err := secretOutputMode("#: secret-output=Redact\n"); err != nil || mode != fraglet.SecretOutputRedact {
		t.Errorf("Redact = %q, %v; want redact", mode, err)
	}
	if _, err := secretOutputMode("#: secret-output=hide\n"); err == nil || !strings.Contains(err.Error(), "hide") {
		t.Errorf("hide = %v; want refused, naming it", err)
	}
}
