package fraglet

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"testing"
)

func TestParseParam_Raw(t *testing.T) {
	p, err := ParseParam("city=london")
	if err != nil {
		t.Fatal(err)
	}
	if p.EnvVar != "CITY" {
		t.Fatalf("EnvVar = %q, want CITY", p.EnvVar)
	}
	if p.Encoding != "raw" {
		t.Fatalf("Encoding = %q, want raw", p.Encoding)
	}
	if p.Value != "london" {
		t.Fatalf("Value = %q, want london", p.Value)
	}
}

func TestParseParam_ExplicitRaw(t *testing.T) {
	p, err := ParseParam("city=raw:london")
	if err != nil {
		t.Fatal(err)
	}
	if p.Encoding != "raw" || p.Value != "london" {
		t.Fatalf("got encoding=%q value=%q, want raw/london", p.Encoding, p.Value)
	}
}

func TestParseParam_B64(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("Hello"))
	p, err := ParseParam("msg=b64:" + encoded)
	if err != nil {
		t.Fatal(err)
	}
	if p.Encoding != "b64" {
		t.Fatalf("Encoding = %q, want b64", p.Encoding)
	}
	decoded, err := p.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "Hello" {
		t.Fatalf("Decode() = %q, want Hello", decoded)
	}
}

func TestParseParam_Cb64(t *testing.T) {
	// Compress then base64 encode
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	w.Write([]byte("compressed data"))
	w.Close()
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	p, err := ParseParam("data=cb64:" + encoded)
	if err != nil {
		t.Fatal(err)
	}
	if p.Encoding != "cb64" {
		t.Fatalf("Encoding = %q, want cb64", p.Encoding)
	}
	decoded, err := p.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "compressed data" {
		t.Fatalf("Decode() = %q, want 'compressed data'", decoded)
	}
}

func TestParseParam_DefaultUppercase(t *testing.T) {
	p, err := ParseParam("city=london")
	if err != nil {
		t.Fatal(err)
	}
	if p.EnvVar != "CITY" {
		t.Fatalf("EnvVar = %q, want CITY", p.EnvVar)
	}
}

func TestParseParam_ReservedName(t *testing.T) {
	_, err := ParseParam("config=value")
	if err == nil {
		t.Fatal("expected error for reserved name CONFIG")
	}
}

func TestParseParam_MissingSeparator(t *testing.T) {
	_, err := ParseParam("noequals")
	if err == nil {
		t.Fatal("expected error for missing =")
	}
}

func TestParseParam_EmptyName(t *testing.T) {
	_, err := ParseParam("=value")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestParam_TransportEnvName(t *testing.T) {
	p := Param{EnvVar: "CITY", Encoding: "raw", Value: "london"}
	if got := p.TransportEnvName(); got != "FRAGLET_PARAM_CITY" {
		t.Fatalf("TransportEnvName() = %q, want FRAGLET_PARAM_CITY", got)
	}
}

func TestParam_TransportEnvValue(t *testing.T) {
	p := Param{EnvVar: "CITY", Encoding: "raw", Value: "london"}
	if got := p.TransportEnvValue(); got != "raw:london" {
		t.Fatalf("TransportEnvValue() = %q, want raw:london", got)
	}
}

func TestParam_Canonical(t *testing.T) {
	p := Param{EnvVar: "CITY", Encoding: "raw", Value: "london"}
	if got := p.Canonical(); got != "CITY=raw:london" {
		t.Fatalf("Canonical() = %q, want CITY=raw:london", got)
	}
}

func TestParams_ToTransportEnv_Sorted(t *testing.T) {
	ps := Params{
		{EnvVar: "UNITS", Encoding: "raw", Value: "metric"},
		{EnvVar: "CITY", Encoding: "raw", Value: "london"},
	}
	got, err := ps.ToTransportEnv()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	// Should be sorted: CITY before UNITS; values are decoded (raw = pass-through)
	if got[0] != "FRAGLET_PARAM_CITY=london" {
		t.Fatalf("got[0] = %q, want FRAGLET_PARAM_CITY=london", got[0])
	}
	if got[1] != "FRAGLET_PARAM_UNITS=metric" {
		t.Fatalf("got[1] = %q, want FRAGLET_PARAM_UNITS=metric", got[1])
	}
}

func TestParams_ToTransportEnv_DecodesB64(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("Hello"))
	ps := Params{{EnvVar: "MSG", Encoding: "b64", Value: encoded}}
	got, err := ps.ToTransportEnv()
	if err != nil {
		t.Fatal(err)
	}
	// Transport carries decoded value — entrypoint is dumb
	if got[0] != "FRAGLET_PARAM_MSG=Hello" {
		t.Fatalf("got[0] = %q, want FRAGLET_PARAM_MSG=Hello", got[0])
	}
}

func TestParams_ToCanonical_Deterministic(t *testing.T) {
	// Same params in different order should produce same canonical
	ps1 := Params{
		{EnvVar: "CITY", Encoding: "raw", Value: "london"},
		{EnvVar: "DATE", Encoding: "raw", Value: "2024-01-15"},
	}
	ps2 := Params{
		{EnvVar: "DATE", Encoding: "raw", Value: "2024-01-15"},
		{EnvVar: "CITY", Encoding: "raw", Value: "london"},
	}
	c1 := ps1.ToCanonical()
	c2 := ps2.ToCanonical()
	if len(c1) != len(c2) {
		t.Fatalf("different lengths: %d vs %d", len(c1), len(c2))
	}
	for i := range c1 {
		if c1[i] != c2[i] {
			t.Fatalf("canonical[%d] differs: %q vs %q", i, c1[i], c2[i])
		}
	}
}

func TestParams_Empty(t *testing.T) {
	var ps Params
	got, err := ps.ToTransportEnv()
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("ToTransportEnv() = %v, want nil", got)
	}
	if got := ps.ToCanonical(); got != nil {
		t.Fatalf("ToCanonical() = %v, want nil", got)
	}
}

func TestParams_ResolveAliases_Default(t *testing.T) {
	ps := Params{{EnvVar: "CITY", Encoding: "raw", Value: "london"}}
	decls := []ParamDecl{{Alias: "city", EnvVar: "CITY"}}
	resolved, err := ps.ResolveAliases(decls)
	if err != nil {
		t.Fatal(err)
	}
	if resolved[0].EnvVar != "CITY" {
		t.Fatalf("EnvVar = %q, want CITY", resolved[0].EnvVar)
	}
}

func TestParams_ResolveAliases_EnvvarMapping(t *testing.T) {
	ps := Params{{EnvVar: "HOST", Encoding: "raw", Value: "localhost"}}
	decls := []ParamDecl{{Alias: "host", EnvVar: "HURL_VARIABLE_host", Modifiers: map[string]string{"envvar": "HURL_VARIABLE_host"}}}
	resolved, err := ps.ResolveAliases(decls)
	if err != nil {
		t.Fatal(err)
	}
	if resolved[0].EnvVar != "HURL_VARIABLE_host" {
		t.Fatalf("EnvVar = %q, want HURL_VARIABLE_host", resolved[0].EnvVar)
	}
	// Canonical should use the resolved env var name
	if got := resolved[0].Canonical(); got != "HURL_VARIABLE_host=raw:localhost" {
		t.Fatalf("Canonical() = %q, want HURL_VARIABLE_host=raw:localhost", got)
	}
}

func TestParams_ResolveAliases_UnknownAlias(t *testing.T) {
	ps := Params{{EnvVar: "UNKNOWN", Encoding: "raw", Value: "val"}}
	decls := []ParamDecl{{Alias: "city", EnvVar: "CITY"}}
	_, err := ps.ResolveAliases(decls)
	if err == nil {
		t.Fatal("expected error for unknown alias")
	}
}

func TestParams_ResolveAliases_EmptyDecls(t *testing.T) {
	ps := Params{{EnvVar: "CITY", Encoding: "raw", Value: "london"}}
	resolved, err := ps.ResolveAliases(nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolved[0].EnvVar != "CITY" {
		t.Fatalf("EnvVar = %q, want CITY", resolved[0].EnvVar)
	}
}
