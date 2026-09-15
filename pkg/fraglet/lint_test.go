package fraglet

import (
	"strings"
	"testing"
)

// cleanScript is a real tool header in full — the reference shape every
// rule below is a deviation from.
const cleanScript = `#!/usr/bin/env -S fragletc --image ofthemachine/headless-browser@sha256:f79c496c6737113c0f6e2d648474a67fb1a416fa34c22220d5714d0b7c6a6036
#: d=Capture a full-page PNG screenshot of a webpage (headless Chromium via Playwright). settle_ms (default 2000) is an extra wait after page load.
#: when=Use when the user wants a picture of a rendered page, or a visual check of a URL, rather than its text.
#: network=required
#: param=url:required:d=Page URL to capture
#: param=headers:d=JSON object of extra HTTP headers (e.g. Authorization)
#: param=settle_ms:default=2000:description=Extra wait after page load for async JS content
#: output=page.png
import json
import os
import sys

url = os.environ["URL"]
headers_raw = os.environ.get("HEADERS", "")
settle_ms = int(os.environ["SETTLE_MS"])
print(url, headers_raw, settle_ms)
`

func rules(fs []Finding) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Rule)
	}
	return out
}

func findRule(t *testing.T, fs []Finding, rule string) Finding {
	t.Helper()
	for _, f := range fs {
		if f.Rule == rule {
			return f
		}
	}
	t.Fatalf("no %s finding in %v", rule, rules(fs))
	return Finding{}
}

func assertNoRule(t *testing.T, fs []Finding, rule string) {
	t.Helper()
	for _, f := range fs {
		if f.Rule == rule {
			t.Fatalf("unexpected %s finding: %s", rule, f.Message)
		}
	}
}

func TestLint_CleanScriptHasNoFindings(t *testing.T) {
	if fs := Lint(cleanScript); len(fs) != 0 {
		t.Fatalf("clean script produced findings: %v", fs)
	}
}

// withHeader builds a script from a three-line prelude (shebang, d=, when=
// on lines 1-3) followed by the given header lines, so a caller's first
// header line is line 4, and then the body.
func withHeader(headerLines string, body string) string {
	prelude := "#!/usr/bin/env -S fragletc --image x\n#: d=Test tool.\n#: when=Use when testing.\n"
	if body == "" {
		body = "import os\nprint(os.environ.get('X', ''))\n"
	}
	return prelude + headerLines + body
}

func TestLint_DescriptionNotLast(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: param=engine:d=TeX engine:default=pdflatex\n", "print(ENGINE)\n"))
	f := findRule(t, fs, "param-desc-not-last")
	if f.Severity != Error || f.Line != 5 || !strings.Contains(f.Message, "default=") {
		t.Fatalf("finding = %+v", f)
	}
}

func TestLint_UnknownModifier(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: param=x:requried:d=X\n", "print(X)\n"))
	f := findRule(t, fs, "param-unknown-modifier")
	if f.Severity != Error || !strings.Contains(f.Message, `"requried"`) {
		t.Fatalf("finding = %+v", f)
	}
}

func TestLint_ModifierValueShape(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: param=x:default:d=X\n#: param=y:file=z:d=Y\n", "print(X, Y)\n"))
	var msgs []string
	for _, f := range fs {
		if f.Rule == "param-unknown-modifier" {
			msgs = append(msgs, f.Message)
		}
	}
	if len(msgs) != 2 {
		t.Fatalf("want 2 param-unknown-modifier findings, got %v", msgs)
	}
	if !strings.Contains(msgs[0], "needs a value") || !strings.Contains(msgs[1], "does not take a value") {
		t.Fatalf("messages = %v", msgs)
	}
}

func TestLint_RequiredWithDefault(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: param=n:required:default=1:d=N\n", "print(N)\n"))
	f := findRule(t, fs, "param-required-with-default")
	if f.Severity != Error {
		t.Fatalf("finding = %+v", f)
	}
}

func TestLint_OptionalIsRedundantWarning(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: param=count:optional:d=Count\n", "print(COUNT)\n"))
	f := findRule(t, fs, "param-optional-redundant")
	if f.Severity != Warning {
		t.Fatalf("finding = %+v", f)
	}
	assertNoRule(t, fs, "param-unknown-modifier")
}

func TestLint_AliasShape(t *testing.T) {
	cases := map[string]bool{ // header → want param-alias finding
		"#: param=Input-File:file:d=X\n": true,
		"#: param=_x:d=X\n":              true,
		"#: param=source.tex:d=X\n":      true,  // dotted without :file
		"#: param=source.tex:file:d=X\n": false, // the mount-path idiom
		"#: param=ok_name2:d=X\n":        false,
	}
	for header, want := range cases {
		fs := Lint(withHeader("#: network=none\n"+header, "print(INPUT-FILE _X SOURCE.TEX OK_NAME2)\n"))
		got := false
		for _, f := range fs {
			if f.Rule == "param-alias" {
				got = true
			}
		}
		if got != want {
			t.Errorf("%q: param-alias finding = %v, want %v (%v)", header, got, want, rules(fs))
		}
	}
}

func TestLint_DuplicateAlias(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: param=x:d=first\n#: param=x:d=second\n", "print(X)\n"))
	f := findRule(t, fs, "param-duplicate")
	if f.Line != 6 || !strings.Contains(f.Message, "line 5") {
		t.Fatalf("finding = %+v", f)
	}
}

func TestLint_NoDescription(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: param=x\n#: param=y:d=\n", "print(X, Y)\n"))
	n := 0
	for _, f := range fs {
		if f.Rule == "param-no-description" {
			n++
			if f.Severity != Warning {
				t.Fatalf("finding = %+v", f)
			}
		}
	}
	if n != 2 {
		t.Fatalf("want 2 param-no-description findings, got %v", rules(fs))
	}
}

func TestLint_ParamUnused(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: param=x:d=X\n#: param=host:envvar=HURL_VARIABLE_host:d=H\n",
		"print(HURL_VARIABLE_host)\n"))
	f := findRule(t, fs, "param-unused")
	if !strings.Contains(f.Message, "param=x ") {
		t.Fatalf("finding = %+v", f)
	}
	// envvar= override is what the body must reference, and it does.
	for _, g := range fs {
		if g.Rule == "param-unused" && strings.Contains(g.Message, "param=host") {
			t.Fatalf("host was referenced via its envvar; finding = %+v", g)
		}
	}
}

func TestLint_NetworkRules(t *testing.T) {
	fs := Lint(withHeader("#: param=x:d=X\n", "print(X)\n"))
	f := findRule(t, fs, "network-missing")
	if f.Severity != Warning || f.Line != 0 {
		t.Fatalf("finding = %+v", f)
	}

	fs = Lint(withHeader("#: network=bridge\n#: param=x:d=X\n", "print(X)\n"))
	f = findRule(t, fs, "network-value")
	if f.Severity != Error || f.Line != 4 {
		t.Fatalf("finding = %+v", f)
	}
	assertNoRule(t, fs, "network-missing")
}

func TestLint_DescMissing(t *testing.T) {
	fs := Lint("#!/usr/bin/env -S fragletc --image x\n#: network=none\nprint(1)\n")
	f := findRule(t, fs, "desc-missing")
	if f.Severity != Warning {
		t.Fatalf("finding = %+v", f)
	}
}

func TestLint_OutputRelpath(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: output=/abs.png\n#: output=../up.png\n#: output=a.png\n#: output=a.png\n", "print(1)\n"))
	n := 0
	for _, f := range fs {
		if f.Rule == "output-relpath" {
			n++
			if f.Severity != Error {
				t.Fatalf("finding = %+v", f)
			}
		}
	}
	if n != 3 {
		t.Fatalf("want 3 output-relpath findings, got %v", fs)
	}
}

func TestLint_FindingsSortedByLine(t *testing.T) {
	fs := Lint(withHeader("#: param=b\n#: param=a\n", "print(A, B)\n"))
	var lines []int
	for _, f := range fs {
		lines = append(lines, f.Line)
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] < lines[i-1] {
			t.Fatalf("findings not sorted by line: %v", lines)
		}
	}
}

func TestLint_WhenMissing(t *testing.T) {
	header := "#!/usr/bin/env -S fragletc --image x\n#: d=Test tool.\n#: network=none\n"
	f := findRule(t, Lint(header+"print(1)\n"), "when-missing")
	if f.Severity != Warning || f.Line != 0 {
		t.Fatalf("finding = %+v", f)
	}
	assertNoRule(t, Lint(header+"#: when=Use when testing.\nprint(1)\n"), "when-missing")
}

func TestLint_StdinValue(t *testing.T) {
	fs := Lint(withHeader("#: network=none\n#: stdin=pipe\n", "print(1)\n"))
	if f := findRule(t, fs, "stdin-value"); f.Severity != Error || f.Line != 5 {
		t.Fatalf("finding = %+v", f)
	}
	assertNoRule(t, Lint(withHeader("#: network=none\n#: stdin=buffer\n", "print(1)\n")), "stdin-value")
	// Reading stdin without declaring it is fine: undeclared means buffer.
	assertNoRule(t, Lint(withHeader("#: network=none\n", "import sys\nprint(sys.stdin.read())\n")), "stdin-undeclared")
}
