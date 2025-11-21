package fraglet

import "testing"

func TestInjectWithMatch(t *testing.T) {
	template := "line1\n    # FRAGLET\nline3"
	code := "print('hi')\nprint('bye')"

	rendered, err := InjectString(template, code, &InjectionConfig{Match: "# FRAGLET"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := rendered
	want := "line1\n    print('hi')\n    print('bye')\nline3"
	if got != want {
		t.Fatalf("render mismatch\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestInjectWithMatch_NotFound(t *testing.T) {
	_, err := InjectString("line", "code", &InjectionConfig{Match: "missing"})
	if err == nil {
		t.Fatal("expected error for missing match")
	}
}

func TestInjectWithMatchRegion(t *testing.T) {
	template := "start\n/* START */\nold\n/* END */\nend"
	code := "new-line"

	rendered, err := InjectString(template, code, &InjectionConfig{
		MatchStart: "/* START */",
		MatchEnd:   "/* END */",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "start\nnew-line\nend"
	if rendered != want {
		t.Fatalf("render mismatch\nwant:\n%s\n\ngot:\n%s", want, rendered)
	}
}
