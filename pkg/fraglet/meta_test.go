package fraglet

import (
	"testing"

	"github.com/ofthemachine/fraglet/internal/testutil"
)

func TestSplitHeader_ShebangAndMeta(t *testing.T) {
	code := "#!/usr/bin/env -S fragletc --image x\n#: d=desc\n#: param=city\n\nprint(1)\nprint(2)"
	header, body := SplitHeader(code)
	wantHeader := "#!/usr/bin/env -S fragletc --image x\n#: d=desc\n#: param=city\n"
	if header != wantHeader {
		t.Fatalf("header = %q, want %q", header, wantHeader)
	}
	if body != "print(1)\nprint(2)" {
		t.Fatalf("body = %q", body)
	}
}

func TestSplitHeader_NoHeader(t *testing.T) {
	code := "print(\"hello\")"
	header, body := SplitHeader(code)
	if header != "" {
		t.Fatalf("header = %q, want empty", header)
	}
	if body != code {
		t.Fatalf("body = %q, want unchanged", body)
	}
}

func TestSplitHeader_AllHeader(t *testing.T) {
	code := "#!/usr/bin/env -S fragletc\n#: d=desc\n"
	header, body := SplitHeader(code)
	if header != code {
		t.Fatalf("header = %q, want whole file", header)
	}
	if body != "" {
		t.Fatalf("body = %q, want empty", body)
	}
}

func TestSplitHeader_BodyHashNeverReparsedAsHeader(t *testing.T) {
	// A body-level "#" comment (past the header boundary) must never be
	// mistaken for a header line, even if it looks exactly like a directive.
	code := "#!/usr/bin/env -S fragletc\n#: param=city\n\nreal_code = 1\n#: param=evil\n"
	if decls := ParseParamDecls(code); len(decls) != 1 || decls[0].Alias != "city" {
		t.Fatalf("decls = %+v, want only city (the body's #: param=evil must be ignored)", decls)
	}
}

func TestParseParamDecls_ShortSentinel(t *testing.T) {
	code := "#: param=city param=units"
	decls := ParseParamDecls(code)
	if len(decls) != 2 {
		t.Fatalf("len = %d, want 2", len(decls))
	}
	if decls[0].Alias != "city" || decls[0].EnvVar != "CITY" {
		t.Fatalf("decls[0] = %+v, want alias=city envvar=CITY", decls[0])
	}
	if decls[1].Alias != "units" || decls[1].EnvVar != "UNITS" {
		t.Fatalf("decls[1] = %+v, want alias=units envvar=UNITS", decls[1])
	}
}

func TestParseParamDecls_LegacySentinelStillWorks(t *testing.T) {
	code := `# fraglet-meta: param=city param=units`
	decls := ParseParamDecls(code)
	if len(decls) != 2 {
		t.Fatalf("len = %d, want 2", len(decls))
	}
}

func TestParseParamDecls_MultiLine(t *testing.T) {
	code := `# fraglet-meta: determinism:deterministic
# fraglet-meta: param=city:required
# fraglet-meta: param=date:required
# fraglet-meta: param=units:optional:default=metric`
	decls := ParseParamDecls(code)
	if len(decls) != 3 {
		t.Fatalf("len = %d, want 3", len(decls))
	}
	// sorted: city, date, units
	if decls[0].Alias != "city" {
		t.Fatalf("decls[0].Alias = %q, want city", decls[0].Alias)
	}
	if !decls[0].IsRequired() {
		t.Fatal("city should be required")
	}
	if decls[2].Alias != "units" {
		t.Fatalf("decls[2].Alias = %q, want units", decls[2].Alias)
	}
	if def, ok := decls[2].Default(); !ok || def != "metric" {
		t.Fatalf("units default = %q/%v, want metric/true", def, ok)
	}
}

func TestParseParamDecls_MixedAnnotationsAndParams(t *testing.T) {
	code := `# fraglet-meta: determinism:deterministic param=city math:algebra`
	decls := ParseParamDecls(code)
	if len(decls) != 1 {
		t.Fatalf("len = %d, want 1 (only param, not annotations)", len(decls))
	}
	if decls[0].Alias != "city" {
		t.Fatalf("alias = %q, want city", decls[0].Alias)
	}
}

func TestParseParamDecls_NonHashLeaderIsBodyNotHeader(t *testing.T) {
	// The header is universally "#"-led (fragletc strips it before the body
	// ever reaches the target language), so a directive-shaped line under a
	// different comment leader is body content, not a recognized directive.
	code := `// fraglet-meta: param=name`
	decls := ParseParamDecls(code)
	if len(decls) != 0 {
		t.Fatalf("decls = %+v, want none (line is body, not header)", decls)
	}
}

func TestParseParamDecls_EnvvarMapping(t *testing.T) {
	code := `# fraglet-meta: param=host:envvar=HURL_VARIABLE_host:required`
	decls := ParseParamDecls(code)
	if len(decls) != 1 {
		t.Fatalf("len = %d, want 1", len(decls))
	}
	d := decls[0]
	if d.Alias != "host" {
		t.Fatalf("Alias = %q, want host", d.Alias)
	}
	if d.EnvVar != "HURL_VARIABLE_host" {
		t.Fatalf("EnvVar = %q, want HURL_VARIABLE_host", d.EnvVar)
	}
	if !d.IsRequired() {
		t.Fatal("should be required")
	}
}

func TestParseParamDecls_Dedup(t *testing.T) {
	code := `# fraglet-meta: param=city
# fraglet-meta: param=city`
	decls := ParseParamDecls(code)
	if len(decls) != 1 {
		t.Fatalf("len = %d, want 1 (dedup)", len(decls))
	}
}

func TestParseParamDecls_NoFragletMeta(t *testing.T) {
	code := `print("hello")`
	decls := ParseParamDecls(code)
	if len(decls) != 0 {
		t.Fatalf("len = %d, want 0", len(decls))
	}
}

func TestParseParamDecls_MultipleModifiers(t *testing.T) {
	code := `# fraglet-meta: param=port:envvar=HURL_VARIABLE_port:default=8080`
	decls := ParseParamDecls(code)
	if len(decls) != 1 {
		t.Fatalf("len = %d, want 1", len(decls))
	}
	d := decls[0]
	if d.EnvVar != "HURL_VARIABLE_port" {
		t.Fatalf("EnvVar = %q, want HURL_VARIABLE_port", d.EnvVar)
	}
	if def, ok := d.Default(); !ok || def != "8080" {
		t.Fatalf("default = %q/%v, want 8080/true", def, ok)
	}
}

func TestParseParamDecls_FileShape(t *testing.T) {
	code := `#: param=doc:file`
	decls := ParseParamDecls(code)
	if len(decls) != 1 {
		t.Fatalf("len = %d, want 1", len(decls))
	}
	if decls[0].Shape != "file" {
		t.Fatalf("Shape = %q, want file", decls[0].Shape)
	}
}

func TestParseParamDecls_DefaultShapeIsString(t *testing.T) {
	code := `#: param=city`
	decls := ParseParamDecls(code)
	if len(decls) != 1 || decls[0].Shape != "string" {
		t.Fatalf("decls = %+v, want Shape=string", decls)
	}
}

func TestParamDecl_IsOptional(t *testing.T) {
	d := ParamDecl{Alias: "x", EnvVar: "X", Modifiers: map[string]string{"optional": ""}}
	if !d.IsOptional() {
		t.Fatal("should be optional")
	}
	if d.IsRequired() {
		t.Fatal("should not be required")
	}
}

func TestParseOutputDecls(t *testing.T) {
	code := "#: output=series.csv\n#: output=plot.png:optional"
	decls := ParseOutputDecls(code)
	if len(decls) != 2 {
		t.Fatalf("len = %d, want 2", len(decls))
	}
	if decls[0].RelPath != "plot.png" || decls[0].IsRequired() {
		t.Fatalf("decls[0] = %+v, want plot.png optional", decls[0])
	}
	if decls[1].RelPath != "series.csv" {
		t.Fatalf("decls[1] = %+v, want series.csv", decls[1])
	}
}

func TestParseOutputDecls_Dedup(t *testing.T) {
	code := "#: output=series.csv\n#: output=series.csv"
	if decls := ParseOutputDecls(code); len(decls) != 1 {
		t.Fatalf("len = %d, want 1 (dedup)", len(decls))
	}
}

func TestParseTags(t *testing.T) {
	code := "#: tags=stats,plotting,r-lang"
	tags := ParseTags(code)
	want := []string{"plotting", "r-lang", "stats"}
	if len(tags) != len(want) {
		t.Fatalf("tags = %v, want %v", tags, want)
	}
	for i, w := range want {
		if tags[i] != w {
			t.Fatalf("tags[%d] = %q, want %q", i, tags[i], w)
		}
	}
}

func TestParseTags_None(t *testing.T) {
	if tags := ParseTags("#: d=desc"); len(tags) != 0 {
		t.Fatalf("tags = %v, want none", tags)
	}
}

func TestParseMetaDescription(t *testing.T) {
	code := `# fraglet-metaparam=mistake
# fraglet-meta: description=Line one.
# fraglet-meta: param=city
# fraglet-meta: d=Line two (d= short form).`
	if got := ParseMetaDescription(code); got != "Line one.\n\nLine two (d= short form)." {
		t.Fatalf("got %q", got)
	}
	if ParseMetaDescription("# fraglet-meta: param=city") != "" {
		t.Fatal("want empty when no description or d=")
	}
}

func TestParseMetaDescription_ShortSentinel(t *testing.T) {
	code := "#: d=Convert a series to CSV."
	if got := ParseMetaDescription(code); got != "Convert a series to CSV." {
		t.Fatalf("got %q", got)
	}
}

func TestParseNetwork(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "empty code",
			code: "",
			want: "",
		},
		{
			name: "no network declaration",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc
				#: param=city

				print(1)
			`),
			want: "",
		},
		{
			name: "network none",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc
				#: network=none

				print(1)
			`),
			want: "none",
		},
		{
			name: "network required",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc
				#: network=required

				print(1)
			`),
			want: "required",
		},
		{
			name: "case insensitive",
			code: testutil.Unindent(`
				#: network=NONE
			`),
			want: "none",
		},
		{
			name: "legacy sentinel",
			code: testutil.Unindent(`
				# fraglet-meta: network=none
			`),
			want: "none",
		},
		{
			name: "body directive ignored",
			code: testutil.Unindent(`
				#!/usr/bin/env -S fragletc

				print(1)
				#: network=none
			`),
			want: "",
		},
		{
			name: "multiple declarations last wins",
			code: testutil.Unindent(`
				#: network=required
				#: network=none
			`),
			want: "none",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseNetwork(tc.code); got != tc.want {
				t.Fatalf("ParseNetwork() = %q, want %q", got, tc.want)
			}
		})
	}
}
