package engine

import (
	"bytes"
	"cmp"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ofthemachine/fraglet/pkg/fraglet"
	"github.com/ofthemachine/fraglet/pkg/receipt"
	"github.com/ofthemachine/fraglet/pkg/runner"
)

// secretGuard watches what a run writes to one stream for the values of
// its secrets: each one seen is noted and, unless the mode keeps them,
// replaced by receipt.Mask before it passes on. It reads left to right,
// the longest value first at each position, and decides a position only
// once every value that could start there has arrived (it holds back one
// fewer bytes than the longest value), so a value split across writes is
// seen, and one inside another is masked whole, never leaving a rest.
type secretGuard struct {
	w       io.Writer
	in      string          // "stdout", "stderr" or "output:<relpath>"
	secrets []runner.Secret // longest first
	keep    bool
	first   [256]bool // a byte some value starts with
	hold    int
	buf     []byte
	seen    []receipt.SecretSeen
}

// newSecretGuard guards writes to w, the place named in, as mode says.
// Empty values are never watched: they match everywhere.
func newSecretGuard(w io.Writer, in string, secrets []runner.Secret, mode string) *secretGuard {
	g := &secretGuard{w: w, in: in, keep: mode == fraglet.SecretOutputUncensored}
	for _, s := range secrets {
		if len(s.Value) > 0 {
			g.secrets = append(g.secrets, s)
			g.first[s.Value[0]] = true
			g.hold = max(g.hold, len(s.Value)-1)
		}
	}
	slices.SortStableFunc(g.secrets, func(a, b runner.Secret) int { return cmp.Compare(len(b.Value), len(a.Value)) })
	return g
}

func (g *secretGuard) Write(p []byte) (int, error) {
	if len(g.secrets) == 0 {
		return g.w.Write(p)
	}
	g.buf = append(g.buf, p...)
	out, n := g.guard(g.buf, len(g.buf)-g.hold)
	if len(out) > 0 {
		if _, err := g.w.Write(out); err != nil {
			return 0, err
		}
	}
	g.buf = append(g.buf[:0], g.buf[n:]...)
	return len(p), nil
}

// close passes on what Write held back.
func (g *secretGuard) close() error {
	out, _ := g.guard(g.buf, len(g.buf))
	g.buf = g.buf[:0]
	if len(out) == 0 {
		return nil
	}
	_, err := g.w.Write(out)
	return err
}

// guard reads b from its start through position limit (exclusive), noting
// each secret value that starts there and masking it unless the guard
// keeps them. It returns what passes on and how much of b it read (past
// limit when a value starting before it ends after).
func (g *secretGuard) guard(b []byte, limit int) ([]byte, int) {
	var out []byte
	i := 0
	for i < limit {
		if !g.first[b[i]] {
			out = append(out, b[i])
			i++
			continue
		}
		j := slices.IndexFunc(g.secrets, func(s runner.Secret) bool { return bytes.HasPrefix(b[i:], s.Value) })
		if j < 0 {
			out = append(out, b[i])
			i++
			continue
		}
		s := g.secrets[j]
		if seen := (receipt.SecretSeen{Secret: s.Name, In: g.in}); !slices.Contains(g.seen, seen) {
			g.seen = append(g.seen, seen)
		}
		if g.keep {
			out = append(out, s.Value...)
		} else {
			out = append(out, receipt.Mask...)
		}
		i += len(s.Value)
	}
	return out, i
}

// guardOutputs reads every file a run left under dir for the values of its
// secrets, masking each in place unless mode keeps them, and says where
// each appeared ("output:<relpath>"), before anything hashes the files.
func guardOutputs(dir string, secrets []runner.Secret, mode string) ([]receipt.SecretSeen, error) {
	if dir == "" || len(secrets) == 0 {
		return nil, nil
	}
	var seen []receipt.SecretSeen
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		g := newSecretGuard(io.Discard, "output:"+filepath.ToSlash(rel), secrets, mode)
		guarded, _ := g.guard(data, len(data))
		seen = append(seen, g.seen...)
		if len(g.seen) == 0 || g.keep {
			return nil
		}
		return os.WriteFile(p, guarded, 0o600) // an existing file keeps its mode
	})
	return seen, err
}

// secretOutput is the receipt's record of a run given secrets: its mode
// and every place a value appeared, in order (a receipt is deterministic);
// nil for a run given none.
func secretOutput(mode string, secrets []runner.Secret, seen []receipt.SecretSeen) *receipt.SecretOutput {
	if len(secrets) == 0 {
		return nil
	}
	slices.SortFunc(seen, func(a, b receipt.SecretSeen) int {
		return cmp.Or(strings.Compare(a.In, b.In), strings.Compare(a.Secret, b.Secret))
	})
	return &receipt.SecretOutput{Mode: mode, Seen: seen}
}

// secretOutputMode is code's secret-output= mode, refused when it is not
// one fragletc has: an unknown mode never runs as if it were another.
func secretOutputMode(code string) (string, error) {
	mode := fraglet.ParseSecretOutput(code)
	if !slices.Contains(fraglet.SecretOutputModes, mode) {
		return "", fmt.Errorf("#: secret-output=%s is not a mode fragletc has; use one of %s", mode, strings.Join(fraglet.SecretOutputModes, ", "))
	}
	return mode, nil
}

// secretOutputErr is the error a run ends with when, in mode fail, a
// secret's value appeared in what it wrote; nil otherwise.
func secretOutputErr(so *receipt.SecretOutput) error {
	if so == nil || so.Mode != fraglet.SecretOutputFail || len(so.Seen) == 0 {
		return nil
	}
	where := make([]string, len(so.Seen))
	for i, s := range so.Seen {
		where[i] = s.Secret + " in " + s.In
	}
	return fmt.Errorf("the script printed a secret it was given (%s, masked); stop printing it, or declare #: secret-output=redact (masked, the run succeeds) or secret-output=uncensored", strings.Join(where, ", "))
}
