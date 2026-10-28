package receipt

import (
	"bytes"
	"io"
	"slices"
	"strings"
)

// Mask is what a masked secret value becomes, as in Outcome.StderrTail.
const Mask = "***"

// Secret is a value a run was given, by its name.
type Secret struct {
	Name  string
	Value []byte
}

// Guard watches what a run writes to one stream for the values of its
// secrets: each one seen is noted (Seen) and, unless keep, replaced by Mask
// before it passes on to w. It holds back the last bytes a value could
// still be completing (one fewer than the longest value) until more arrive
// or Close, so a value split across writes is still seen. Close flushes.
type Guard struct {
	w       io.Writer
	in      string
	secrets []Secret
	keep    bool
	hold    int
	buf     []byte
	seen    []SecretSeen
}

// NewGuard guards writes to w, the stream named in ("stdout", "stderr").
// Empty values are never watched: they match everywhere.
func NewGuard(w io.Writer, in string, secrets []Secret, keep bool) *Guard {
	g := &Guard{w: w, in: in, keep: keep}
	for _, s := range secrets {
		if len(s.Value) > 0 {
			g.secrets = append(g.secrets, s)
			g.hold = max(g.hold, len(s.Value)-1)
		}
	}
	return g
}

func (g *Guard) Write(p []byte) (int, error) {
	g.buf = append(g.buf, p...)
	g.buf = g.scan(g.buf)
	if n := len(g.buf) - g.hold; n > 0 {
		if _, err := g.w.Write(g.buf[:n]); err != nil {
			return 0, err
		}
		g.buf = append(g.buf[:0], g.buf[n:]...)
	}
	return len(p), nil
}

// Close passes on what Write held back.
func (g *Guard) Close() error {
	if len(g.buf) == 0 {
		return nil
	}
	_, err := g.w.Write(g.buf)
	g.buf = g.buf[:0]
	return err
}

// Seen is where each secret appeared, by name, in this stream.
func (g *Guard) Seen() []SecretSeen { return g.seen }

// scan notes each secret in b and, unless keep, masks it.
func (g *Guard) scan(b []byte) []byte {
	for _, s := range g.secrets {
		if !bytes.Contains(b, s.Value) {
			continue
		}
		g.note(s.Name)
		if !g.keep {
			b = bytes.ReplaceAll(b, s.Value, []byte(Mask))
		}
	}
	return b
}

func (g *Guard) note(name string) {
	seen := SecretSeen{Secret: name, In: g.in}
	if !slices.Contains(g.seen, seen) {
		g.seen = append(g.seen, seen)
	}
}

// GuardBytes is a Guard over bytes already whole (a file the run left):
// where each secret appeared in data, and data masked unless keep.
func GuardBytes(data []byte, in string, secrets []Secret, keep bool) ([]byte, []SecretSeen) {
	g := NewGuard(io.Discard, in, secrets, keep)
	g.hold = 0
	return g.scan(data), g.seen
}

// SortSeen orders seen by place, then secret: a receipt is deterministic.
func SortSeen(seen []SecretSeen) {
	slices.SortFunc(seen, func(a, b SecretSeen) int {
		if c := strings.Compare(a.In, b.In); c != 0 {
			return c
		}
		return strings.Compare(a.Secret, b.Secret)
	})
}
