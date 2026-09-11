package fraglet

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"os"
	"path/filepath"
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

func TestResolveFileParams_RelativePath(t *testing.T) {
	hostFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(hostFile, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relPath, err := filepath.Rel(wd, hostFile)
	if err != nil {
		t.Fatal(err)
	}

	decls := []ParamDecl{{Alias: "doc", EnvVar: "DOC", Shape: "file"}}
	params := Params{{EnvVar: "DOC", Encoding: "raw", Value: relPath}}
	rewritten, mounts, err := ResolveFileParams(decls, params, "/input")
	if err != nil {
		t.Fatalf("ResolveFileParams: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("mounts len = %d, want 1", len(mounts))
	}
	if mounts[0].HostPath != hostFile {
		t.Fatalf("HostPath = %q, want absolute %q", mounts[0].HostPath, hostFile)
	}
	if mounts[0].ContainerPath != "/input/doc" {
		t.Fatalf("ContainerPath = %q, want /input/doc", mounts[0].ContainerPath)
	}
	if rewritten[0].Value != "/input/doc" {
		t.Fatalf("rewritten value = %q, want /input/doc", rewritten[0].Value)
	}
}

func TestResolveFileParams_DecodedB64Path(t *testing.T) {
	hostFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(hostFile, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(hostFile))

	decls := []ParamDecl{{Alias: "doc", EnvVar: "DOC", Shape: "file"}}
	params := Params{{EnvVar: "DOC", Encoding: "b64", Value: encoded}}
	_, mounts, err := ResolveFileParams(decls, params, "/input")
	if err != nil {
		t.Fatalf("ResolveFileParams: %v", err)
	}
	if len(mounts) != 1 || mounts[0].HostPath != hostFile {
		t.Fatalf("mounts = %+v, want host %q", mounts, hostFile)
	}
}

func TestResolveFileParams_StringParamUnchanged(t *testing.T) {
	decls := []ParamDecl{
		{Alias: "city", EnvVar: "CITY", Shape: "string"},
		{Alias: "doc", EnvVar: "DOC", Shape: "file"},
	}
	hostFile := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(hostFile, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	params := Params{
		{EnvVar: "CITY", Encoding: "raw", Value: "london"},
		{EnvVar: "DOC", Encoding: "raw", Value: hostFile},
	}
	rewritten, mounts, err := ResolveFileParams(decls, params, "/input")
	if err != nil {
		t.Fatalf("ResolveFileParams: %v", err)
	}
	if len(mounts) != 1 {
		t.Fatalf("mounts len = %d, want 1", len(mounts))
	}
	if rewritten[0].Value != "london" {
		t.Fatalf("string param rewritten to %q, want london", rewritten[0].Value)
	}
}

func TestResolveFileParams_MissingFile(t *testing.T) {
	decls := []ParamDecl{{Alias: "doc", EnvVar: "DOC", Shape: "file"}}
	params := Params{{EnvVar: "DOC", Encoding: "raw", Value: filepath.Join(t.TempDir(), "missing.txt")}}
	if _, _, err := ResolveFileParams(decls, params, "/input"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestResolveFileParams_DirectoryRejected(t *testing.T) {
	dir := t.TempDir()
	decls := []ParamDecl{{Alias: "doc", EnvVar: "DOC", Shape: "file"}}
	params := Params{{EnvVar: "DOC", Encoding: "raw", Value: dir}}
	if _, _, err := ResolveFileParams(decls, params, "/input"); err == nil {
		t.Fatal("expected error for directory path")
	}
}

func TestApplyDefaults(t *testing.T) {
	decls := []ParamDecl{
		{Alias: "encoding", EnvVar: "ENCODING", Modifiers: map[string]string{"default": "cl100k_base"}},
		{Alias: "text", EnvVar: "TEXT", Modifiers: map[string]string{"required": ""}},
	}
	got := ApplyDefaults(decls, nil)
	if len(got) != 1 || got[0].EnvVar != "ENCODING" || got[0].Value != "cl100k_base" {
		t.Fatalf("ApplyDefaults(nil) = %+v, want ENCODING=cl100k_base", got)
	}

	provided := Params{{EnvVar: "ENCODING", Encoding: "raw", Value: "o200k_base"}}
	got = ApplyDefaults(decls, provided)
	if len(got) != 1 || got[0].Value != "o200k_base" {
		t.Fatalf("explicit -p must win: %+v", got)
	}

	empty := Params{{EnvVar: "ENCODING", Encoding: "raw", Value: ""}}
	got = ApplyDefaults(decls, empty)
	if len(got) != 1 || got[0].Value != "" {
		t.Fatalf("explicit empty -p must not be overwritten: %+v", got)
	}
}
