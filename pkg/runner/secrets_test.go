package runner

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"
)

func TestSecretsTar(t *testing.T) {
	data, err := secretsTar([]Secret{
		{Path: "/run/fraglet/secrets/A", Value: []byte("alpha\nline2")},
		{Path: "/run/fraglet/secrets/B", Value: []byte("")},
	})
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(bytes.NewReader(data))
	want := map[string]string{"run/fraglet/secrets/A": "alpha\nline2", "run/fraglet/secrets/B": ""}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Typeflag != tar.TypeReg || h.Mode != 0o444 || h.Uid != 0 || h.Gid != 0 {
			t.Errorf("%s: type %c mode %o uid %d gid %d; want regular 0444 root", h.Name, h.Typeflag, h.Mode, h.Uid, h.Gid)
		}
		body, _ := io.ReadAll(tr)
		v, ok := want[h.Name]
		if !ok || string(body) != v {
			t.Errorf("unexpected entry %q = %q", h.Name, body)
		}
		delete(want, h.Name)
	}
	if len(want) != 0 {
		t.Errorf("missing entries: %v", want)
	}
}

func TestSecretsTarRejectsRelativePath(t *testing.T) {
	if _, err := secretsTar([]Secret{{Path: "run/x", Value: []byte("v")}}); err == nil {
		t.Fatal("expected an error for a relative path")
	}
}
