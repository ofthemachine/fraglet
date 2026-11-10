package fraglet

import (
	"slices"
	"strings"
	"testing"
)

func TestParseRun(t *testing.T) {
	const shebang = "#!/usr/bin/env -S fragletc --image python:3.12-slim@sha256:abc\n"
	for _, tc := range []struct {
		name, header string
		argv         []string
		path         string
		wantErr      string
	}{
		{name: "none", header: "#: d=plain\n"},
		{name: "file", header: "#: run=python3 {code:/work/main.py}\n",
			argv: []string{"python3", "/work/main.py"}, path: "/work/main.py"},
		{name: "inside a word", header: "#:  run=tool --script={code:/tmp/s.txt} -v\n",
			argv: []string{"tool", "--script=/tmp/s.txt", "-v"}, path: "/tmp/s.txt"},
		{name: "legacy sentinel", header: "# fraglet-meta: run=sh {code:/s.sh}\n",
			argv: []string{"sh", "/s.sh"}, path: "/s.sh"},
		{name: "twice", header: "#: run=sh {code:/a}\n#: run=sh {code:/b}\n", wantErr: "once"},
		{name: "empty", header: "#: run=\n", wantErr: "no command"},
		{name: "no body", header: "#: run=sh -s\n", wantErr: "where the body is written"},
		{name: "two bodies", header: "#: run=cat {code:/a} {code:/b}\n", wantErr: "one place"},
		{name: "relative", header: "#: run=sh {code:main.sh}\n", wantErr: "absolute"},
		{name: "unclean", header: "#: run=sh {code:/a/../b}\n", wantErr: "absolute"},
		{name: "unclosed", header: "#: run=sh {code:/a\n", wantErr: "unclosed"},
		{name: "body as program", header: "#: run={code:/a}\n", wantErr: "first word"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := ParseRun(shebang + tc.header + "print('hi')\n")
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v; want one saying %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.argv == nil {
				if r != nil {
					t.Fatalf("recipe = %+v; want none", r)
				}
				return
			}
			if r == nil || !slices.Equal(r.Argv, tc.argv) || r.CodePath != tc.path {
				t.Fatalf("recipe = %+v; want %v at %s", r, tc.argv, tc.path)
			}
		})
	}
}

// Everything on a run line is its command: a word that looks like another
// directive declares nothing.
func TestRunLineIsNoOtherDirective(t *testing.T) {
	code := "#!/usr/bin/env -S fragletc --image x@sha256:abc\n" +
		"#: run=env network=none param=ghost output=ghost.txt sh {code:/s.sh}\n" +
		"#: param=real network=required\n" +
		"echo hi\n"
	if got := ParseNetwork(code); got != "required" {
		t.Errorf("network = %q; want required", got)
	}
	if ps := ParseParamDecls(code); len(ps) != 1 || ps[0].Alias != "real" {
		t.Errorf("params = %+v; want only real", ps)
	}
	if outs := ParseOutputDecls(code); len(outs) != 0 {
		t.Errorf("outputs = %+v; want none", outs)
	}
}

func TestBareEnv(t *testing.T) {
	got := BareEnv([]string{"FRAGLET_PARAM_CITY=paris", "FRAGLET_PARAM_UNITS=metric", "UNITS=kept", "OTHER=1", "FRAGLET_PARAM_=x"})
	want := []string{"FRAGLET_PARAM_CITY=paris", "FRAGLET_PARAM_UNITS=metric", "UNITS=kept", "OTHER=1", "FRAGLET_PARAM_=x", "CITY=paris"}
	if !slices.Equal(got, want) {
		t.Errorf("BareEnv = %v\nwant      %v", got, want)
	}
}

func TestLintRun(t *testing.T) {
	rules := func(code string) []string {
		var out []string
		for _, f := range Lint(code) {
			if f.Rule == "run-value" {
				out = append(out, f.Message)
			}
		}
		return out
	}
	const shebang = "#!/usr/bin/env -S fragletc --image x@sha256:abc\n"
	if got := rules(shebang + "#: run=sh {code:/s.sh}\necho\n"); len(got) != 0 {
		t.Errorf("a good run line = %v; want no finding", got)
	}
	if got := rules(shebang + "#: run=sh -s\necho\n"); len(got) != 1 {
		t.Errorf("a run line with no body = %v; want one finding", got)
	}
	if got := rules(shebang + "#: run=sh {code:/a}\n#: run=sh {code:/b}\necho\n"); len(got) != 1 || !strings.Contains(got[0], "second") {
		t.Errorf("two run lines = %v; want the second flagged", got)
	}
}
