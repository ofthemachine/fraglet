package save

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const shebangPrefix = "#!/usr/bin/env -S fragletc"
const fragletMetaPrefix = "# fraglet-meta:"

// ArtifactSaver persists a rendered fraglet artifact content-addressed by lang and hash.
// Implementations may be synchronous (e.g. local filesystem) or asynchronous (e.g. HTTP);
// callers must not assume durability. Used when MCP run is started with --save.
type ArtifactSaver interface {
	Save(ctx context.Context, lang string, imageWithDigest string, mode string, annotations []string, body string) error
}

// BuildArtifactContent produces the exact bytes to be stored: shebang (--image=...@sha256, --mode only when mode != ""),
// optional # fraglet-meta line with annotations in lexically sorted order, then body.
// Callers hash this content and use it as the content-address.
func BuildArtifactContent(imageWithDigest string, mode string, annotations []string, body string) []byte {
	var b strings.Builder
	// Shebang: --image=... always; --mode=X only when explicit
	b.WriteString(shebangPrefix)
	b.WriteString(" --image=")
	b.WriteString(imageWithDigest)
	if mode != "" {
		b.WriteString(" --mode=")
		b.WriteString(mode)
	}
	b.WriteString("\n")
	if len(annotations) > 0 {
		sorted := make([]string, len(annotations))
		copy(sorted, annotations)
		sort.Strings(sorted)
		b.WriteString(fragletMetaPrefix)
		b.WriteString(" ")
		b.WriteString(strings.Join(sorted, " "))
		b.WriteString("\n")
	}
	if body != "" {
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteString("\n")
		}
	}
	return []byte(b.String())
}

// HashContent returns SHA256 hex (64 chars) of content, matching smax/pkg/hash style.
func HashContent(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("%x", h[:])
}

// LocalSave is a synchronous filesystem implementation of ArtifactSaver.
// Path layout: root/lang/[tagPrefix/]h[:2]/h[2:]. If annotations are present, the first (lexically)
// annotation's tag-prefix (the part before the first ':') is used as a subfolder under lang;
// otherwise no subfolder is used.
type LocalSave struct {
	Root string
}

// tagPrefixFromAnnotations returns the tag-prefix (segment before ':') of the lexically first
// annotation, or "" if there are no annotations or the first has no colon.
func tagPrefixFromAnnotations(annotations []string) string {
	if len(annotations) == 0 {
		return ""
	}
	sorted := make([]string, len(annotations))
	copy(sorted, annotations)
	sort.Strings(sorted)
	first := sorted[0]
	prefix, _, _ := strings.Cut(first, ":")
	prefix = strings.TrimSpace(prefix)
	// Sanitize for use as path segment: no path separators
	if prefix == "" || strings.ContainsAny(prefix, "/\\") {
		return ""
	}
	return prefix
}

// Save builds the artifact, hashes it, and writes atomically under Root/lang/[tagPrefix/]h[:2]/h[2:].
func (s *LocalSave) Save(ctx context.Context, lang string, imageWithDigest string, mode string, annotations []string, body string) error {
	content := BuildArtifactContent(imageWithDigest, mode, annotations, body)
	h := HashContent(content)
	if len(h) < 4 {
		return fmt.Errorf("save: hash too short")
	}
	subDir := filepath.Join(s.Root, lang)
	if prefix := tagPrefixFromAnnotations(annotations); prefix != "" {
		subDir = filepath.Join(subDir, prefix)
	}
	dir := filepath.Join(subDir, h[:2])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("save: mkdir: %w", err)
	}
	path := filepath.Join(dir, h[2:])
	// Atomic write: temp in same dir then rename
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("save: create temp: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		return fmt.Errorf("save: write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("save: sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("save: close: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("save: rename: %w", err)
	}
	return nil
}
