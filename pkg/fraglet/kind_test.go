package fraglet

import (
	"reflect"
	"testing"
)

func TestParamKind(t *testing.T) {
	code := `#!/usr/bin/env -S fragletc --image x@sha256:abc
#: param=hex:kind=color:d=Light color
#: param=on:kind=enum(on|off|toggle)
#: param=bri:kind=range(0..100,5):default=50
#: param=photo:file
#: param=city
`
	got := map[string]Kind{}
	for _, d := range ParseParamDecls(code) {
		got[d.Alias] = d.Kind()
	}
	want := map[string]Kind{
		"hex":   {Name: "color"},
		"on":    {Name: "enum", Options: []string{"on", "off", "toggle"}},
		"bri":   {Name: "range", Min: 0, Max: 100, Step: 5},
		"photo": {Name: "file"},
		"city":  {Name: "string"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("kinds = %+v, want %+v", got, want)
	}
	if s := want["bri"].String(); s != "range(0..100,5)" {
		t.Errorf("String = %q", s)
	}
}

func TestKindNormalize(t *testing.T) {
	cases := []struct {
		kind, in, want string
		bad            bool
	}{
		{"color", "#F80", "#ff8800", false},
		{"color", "ff8000", "#ff8000", false},
		{"color", "Warm White", "#ffd6aa", false},
		{"color", "nope", "", true},
		{"enum(on|off)", "ON", "on", false},
		{"enum(on|off)", "maybe", "", true},
		{"range(0..100)", "101", "", true},
		{"range(0..100)", "42.5", "42.5", false},
		{"bool", "yes", "true", false},
		{"integer", "4.2", "", true},
		{"date", "2026-10-06", "2026-10-06", false},
		{"date", "10/06/2026", "", true},
		{"string", " any ", "any", false},
	}
	for _, c := range cases {
		k, err := ParseKind(c.kind)
		if err != nil {
			t.Fatalf("ParseKind(%q): %v", c.kind, err)
		}
		got, err := k.Normalize(c.in)
		if (err != nil) != c.bad || got != c.want {
			t.Errorf("%s.Normalize(%q) = %q, %v; want %q (error %v)", c.kind, c.in, got, err, c.want, c.bad)
		}
	}
	for _, bad := range []string{"enum(one)", "range(5..1)", "colour", "color(x)", "range(a..b)"} {
		if _, err := ParseKind(bad); err == nil {
			t.Errorf("ParseKind(%q) accepted", bad)
		}
	}
}

func TestLintKind(t *testing.T) {
	code := "#!/usr/bin/env -S fragletc --image x@sha256:abc\n#: d=x\n#: param=a:kind=colour\n#: param=b:kind=range(0..10):default=11\n#: param=c:kind=enum(x|y)\necho $A $B $C\n"
	var kinds int
	for _, f := range Lint(code) {
		if f.Rule == "param-kind" {
			kinds++
		}
	}
	if kinds != 2 {
		t.Errorf("param-kind findings = %d, want 2 (unknown kind, default out of range): %+v", kinds, Lint(code))
	}
}
