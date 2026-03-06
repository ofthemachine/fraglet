package save

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildArtifactContent_NoModeNoAnnotations(t *testing.T) {
	content := BuildArtifactContent("img@sha256:abc", "", nil, "print(1)")
	got := string(content)
	if got != "#!/usr/bin/env -S fragletc --image=img@sha256:abc\nprint(1)\n" {
		t.Errorf("unexpected content:\n%q", got)
	}
}

func TestBuildArtifactContent_WithMode(t *testing.T) {
	content := BuildArtifactContent("img@sha256:abc", "main", nil, "print(1)")
	got := string(content)
	if got != "#!/usr/bin/env -S fragletc --image=img@sha256:abc --mode=main\nprint(1)\n" {
		t.Errorf("unexpected content:\n%q", got)
	}
}

func TestBuildArtifactContent_AnnotationsSorted(t *testing.T) {
	content := BuildArtifactContent("img@sha256:x", "", []string{"z:last", "a:first"}, "code")
	got := string(content)
	// Must be lexically sorted: a:first before z:last
	if got != "#!/usr/bin/env -S fragletc --image=img@sha256:x\n# fraglet-meta: a:first z:last\ncode\n" {
		t.Errorf("unexpected content (annotations should be sorted):\n%q", got)
	}
}

func TestBuildArtifactContent_IdenticalInputsByteIdentical(t *testing.T) {
	a := BuildArtifactContent("i@sha256:x", "m", []string{"b:2", "a:1"}, "body")
	b := BuildArtifactContent("i@sha256:x", "m", []string{"b:2", "a:1"}, "body")
	if string(a) != string(b) {
		t.Error("identical inputs must produce byte-identical output")
	}
	// Same annotations in different order still sorted the same
	c := BuildArtifactContent("i@sha256:x", "m", []string{"a:1", "b:2"}, "body")
	if string(a) != string(c) {
		t.Error("annotation order should not affect output (sorted in code)")
	}
}

func TestHashContent_Deterministic(t *testing.T) {
	content := []byte("hello")
	h1 := HashContent(content)
	h2 := HashContent(content)
	if h1 != h2 {
		t.Errorf("hash not deterministic: %s vs %s", h1, h2)
	}
	if len(h1) != 64 {
		t.Errorf("expected 64-char hex hash, got len %d", len(h1))
	}
}

func TestLocalSave_SavesAndShards(t *testing.T) {
	dir := t.TempDir()
	s := &LocalSave{Root: dir}
	ctx := context.Background()
	err := s.Save(ctx, "python", "img@sha256:abc123", "", []string{"a:1"}, "print(1)")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	content := BuildArtifactContent("img@sha256:abc123", "", []string{"a:1"}, "print(1)")
	h := HashContent(content)
	// First annotation's tag-prefix "a" is used as subfolder
	expectPath := filepath.Join(dir, "python", "a", h[:2], h[2:])
	data, err := os.ReadFile(expectPath)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("saved content mismatch:\ngot %q\nwant %q", string(data), string(content))
	}
}

func TestLocalSave_Idempotent(t *testing.T) {
	dir := t.TempDir()
	s := &LocalSave{Root: dir}
	ctx := context.Background()
	err := s.Save(ctx, "python", "img@sha256:x", "", nil, "1+1")
	if err != nil {
		t.Fatalf("first Save: %v", err)
	}
	err = s.Save(ctx, "python", "img@sha256:x", "", nil, "1+1")
	if err != nil {
		t.Fatalf("second Save: %v", err)
	}
	content := BuildArtifactContent("img@sha256:x", "", nil, "1+1")
	h := HashContent(content)
	path := filepath.Join(dir, "python", h[:2], h[2:])
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not present after second save: %v", err)
	}
}

func TestLocalSave_NoSubfolderWhenNoAnnotations(t *testing.T) {
	dir := t.TempDir()
	s := &LocalSave{Root: dir}
	ctx := context.Background()
	err := s.Save(ctx, "ruby", "img@sha256:x", "", nil, "puts 1")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	content := BuildArtifactContent("img@sha256:x", "", nil, "puts 1")
	h := HashContent(content)
	path := filepath.Join(dir, "ruby", h[:2], h[2:])
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file under lang/h[:2]/h[2:] when no annotations: %v", err)
	}
}

func TestLocalSave_SubfolderFromLexicallyFirstAnnotation(t *testing.T) {
	dir := t.TempDir()
	s := &LocalSave{Root: dir}
	ctx := context.Background()
	// Lexically first is "math:number-theory" (m < z)
	err := s.Save(ctx, "python", "img@sha256:x", "", []string{"z:last", "math:number-theory"}, "x=1")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	content := BuildArtifactContent("img@sha256:x", "", []string{"z:last", "math:number-theory"}, "x=1")
	h := HashContent(content)
	path := filepath.Join(dir, "python", "math", h[:2], h[2:])
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file under lang/math/... (lexically first tag-prefix): %v", err)
	}
}
