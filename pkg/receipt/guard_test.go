package receipt

import (
	"bytes"
	"slices"
	"testing"
)

// A guard masks every secret value a stream carries, however the writes
// split it, notes each by name, and passes everything else on unchanged;
// keep notes without masking.
func TestGuard(t *testing.T) {
	secrets := []Secret{{Name: "API_KEY", Value: []byte("sk-0123456789")}, {Name: "PIN", Value: []byte("4242")}, {Name: "EMPTY"}}
	text := "key sk-0123456789 and sk-0123456789, pin 4242.\nplain line\n"
	for _, split := range []int{1, 3, 5, len(text)} {
		for _, keep := range []bool{false, true} {
			var out bytes.Buffer
			g := NewGuard(&out, "stdout", secrets, keep)
			for i := 0; i < len(text); i += split {
				if _, err := g.Write([]byte(text[i:min(i+split, len(text))])); err != nil {
					t.Fatal(err)
				}
			}
			if err := g.Close(); err != nil {
				t.Fatal(err)
			}
			want := "key *** and ***, pin ***.\nplain line\n"
			if keep {
				want = text
			}
			if out.String() != want {
				t.Errorf("split %d keep %v: wrote %q, want %q", split, keep, out.String(), want)
			}
			seen := g.Seen()
			SortSeen(seen)
			if !slices.Equal(seen, []SecretSeen{{"API_KEY", "stdout"}, {"PIN", "stdout"}}) {
				t.Errorf("split %d keep %v: seen %v, want API_KEY and PIN in stdout", split, keep, seen)
			}
		}
	}
	var out bytes.Buffer
	g := NewGuard(&out, "stderr", secrets, false)
	g.Write([]byte("nothing to see"))
	g.Close()
	if out.String() != "nothing to see" || len(g.Seen()) != 0 {
		t.Errorf("a clean stream changed: %q, seen %v", out.String(), g.Seen())
	}
}

func TestGuardBytes(t *testing.T) {
	secrets := []Secret{{Name: "API_KEY", Value: []byte("sk-0123456789")}}
	got, seen := GuardBytes([]byte("x sk-0123456789 y"), "output:a.txt", secrets, false)
	if string(got) != "x *** y" || !slices.Equal(seen, []SecretSeen{{"API_KEY", "output:a.txt"}}) {
		t.Errorf("got %q, seen %v", got, seen)
	}
	got, seen = GuardBytes([]byte("x sk-0123456789 y"), "output:a.txt", secrets, true)
	if string(got) != "x sk-0123456789 y" || len(seen) != 1 {
		t.Errorf("keep: got %q, seen %v", got, seen)
	}
}
