package testutil

import (
	"strings"
	"testing"
)

func TestUnindent(t *testing.T) {
	got := Unindent(`
		#!/usr/bin/env -S fragletc
		#: network=none

		print(1)
	`)
	want := "#!/usr/bin/env -S fragletc\n#: network=none\n\nprint(1)"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestUnindent_PreservesRelativeIndent(t *testing.T) {
	got := Unindent(`
		def f():
			return 1
	`)
	want := "def f():\n\treturn 1"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestUnindent_Empty(t *testing.T) {
	if got := Unindent(""); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
	if got := Unindent("\n\n\t\n"); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestUnindent_JaggedLeftPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on jagged-left")
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("panic value %v (%T), want error", r, r)
		}
		if !strings.Contains(err.Error(), "jagged-left") {
			t.Fatalf("panic %q, want it to mention jagged-left", err.Error())
		}
	}()
	Unindent(`
		first line
	jagged
	`)
}

func TestUnindent_Spaces(t *testing.T) {
	got, err := unindent("\n    a\n    b\n")
	if err != nil {
		t.Fatal(err)
	}
	if got != "a\nb" {
		t.Fatalf("got %q, want a\\nb", got)
	}
}
