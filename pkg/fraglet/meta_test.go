package fraglet

import "testing"

func TestParseParamDecls_SingleLine(t *testing.T) {
	code := `# fraglet-meta: param=city param=units`
	decls := ParseParamDecls(code)
	if len(decls) != 2 {
		t.Fatalf("len = %d, want 2", len(decls))
	}
	// sorted by alias
	if decls[0].Alias != "city" || decls[0].EnvVar != "CITY" {
		t.Fatalf("decls[0] = %+v, want alias=city envvar=CITY", decls[0])
	}
	if decls[1].Alias != "units" || decls[1].EnvVar != "UNITS" {
		t.Fatalf("decls[1] = %+v, want alias=units envvar=UNITS", decls[1])
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

func TestParseParamDecls_DoubleSlashComment(t *testing.T) {
	code := `// fraglet-meta: param=name`
	decls := ParseParamDecls(code)
	if len(decls) != 1 || decls[0].Alias != "name" {
		t.Fatalf("decls = %+v, want [{name NAME ...}]", decls)
	}
}

func TestParseParamDecls_DashDashComment(t *testing.T) {
	code := `-- fraglet-meta: param=val`
	decls := ParseParamDecls(code)
	if len(decls) != 1 || decls[0].Alias != "val" {
		t.Fatalf("decls = %+v, want [{val VAL ...}]", decls)
	}
}

func TestParseParamDecls_PercentComment(t *testing.T) {
	code := `% fraglet-meta: param=x`
	decls := ParseParamDecls(code)
	if len(decls) != 1 || decls[0].Alias != "x" {
		t.Fatalf("decls = %+v, want [{x X ...}]", decls)
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

func TestParamDecl_IsOptional(t *testing.T) {
	d := ParamDecl{Alias: "x", EnvVar: "X", Modifiers: map[string]string{"optional": ""}}
	if !d.IsOptional() {
		t.Fatal("should be optional")
	}
	if d.IsRequired() {
		t.Fatal("should not be required")
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
