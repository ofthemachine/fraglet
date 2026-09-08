package fraglet

import (
	"reflect"
	"testing"
)

func lookup(env map[string]string) func(string) (string, bool) {
	return func(name string) (string, bool) {
		v, ok := env[name]
		return v, ok
	}
}

func TestExpandArgv_Basic(t *testing.T) {
	got, err := ExpandArgv(`meme-cli render "$TEMPLATE" "$TOP" "$BOTTOM" -o /output/meme.png`,
		nil, lookup(map[string]string{"TEMPLATE": "drake", "TOP": "writing tests", "BOTTOM": "hoping it works"}))
	if err != nil {
		t.Fatalf("ExpandArgv() error = %v", err)
	}
	want := []string{"meme-cli", "render", "drake", "writing tests", "hoping it works", "-o", "/output/meme.png"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandArgv_At(t *testing.T) {
	got, err := ExpandArgv(`meme-cli $@`, []string{"render", "drake", "a b", "c"}, lookup(nil))
	if err != nil {
		t.Fatalf("ExpandArgv() error = %v", err)
	}
	want := []string{"meme-cli", "render", "drake", "a b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandArgv_QuotedAtStaysSeparateWords(t *testing.T) {
	// "$@" must still yield one token per positional arg, same as real
	// shell "$@" — never collapsed into a single space-joined string.
	got, err := ExpandArgv(`meme-cli "$@"`, []string{"a b", "c"}, lookup(nil))
	if err != nil {
		t.Fatalf("ExpandArgv() error = %v", err)
	}
	want := []string{"meme-cli", "a b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandArgv_PositionalIndex(t *testing.T) {
	got, err := ExpandArgv(`meme-cli render $1 $2`, []string{"drake", "hi"}, lookup(nil))
	if err != nil {
		t.Fatalf("ExpandArgv() error = %v", err)
	}
	want := []string{"meme-cli", "render", "drake", "hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandArgv_SingleQuoteLiteral(t *testing.T) {
	got, err := ExpandArgv(`meme-cli render 'literal $TEMPLATE text'`, nil, lookup(map[string]string{"TEMPLATE": "drake"}))
	if err != nil {
		t.Fatalf("ExpandArgv() error = %v", err)
	}
	want := []string{"meme-cli", "render", "literal $TEMPLATE text"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandArgv_HeaderStripped(t *testing.T) {
	body := "#!/usr/bin/env -S fragletc --image=x\n#: param=template:required\n\nmeme-cli render \"$TEMPLATE\"\n"
	got, err := ExpandArgv(body, nil, lookup(map[string]string{"TEMPLATE": "drake"}))
	if err != nil {
		t.Fatalf("ExpandArgv() error = %v", err)
	}
	want := []string{"meme-cli", "render", "drake"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandArgv_RejectsShellOperators(t *testing.T) {
	for _, body := range []string{
		"meme-cli render drake | cat",
		"meme-cli render drake; echo done",
		"meme-cli render drake && echo ok",
		"meme-cli render $(whoami)",
		"meme-cli render *.png",
	} {
		if _, err := ExpandArgv(body, nil, lookup(nil)); err == nil {
			t.Errorf("ExpandArgv(%q): want error, got nil", body)
		}
	}
}

func TestExpandArgv_RejectsMultiLineBody(t *testing.T) {
	_, err := ExpandArgv("meme-cli render drake\nmeme-cli render doge", nil, lookup(nil))
	if err == nil {
		t.Fatal("ExpandArgv() with two command lines: want error, got nil")
	}
}

func TestExpandArgv_UnknownVarExpandsEmpty(t *testing.T) {
	got, err := ExpandArgv(`meme-cli render "$MISSING"`, nil, lookup(nil))
	if err != nil {
		t.Fatalf("ExpandArgv() error = %v", err)
	}
	want := []string{"meme-cli", "render", ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandArgv_EmptyBody(t *testing.T) {
	if _, err := ExpandArgv("", nil, lookup(nil)); err == nil {
		t.Fatal("ExpandArgv(\"\"): want error, got nil")
	}
}
