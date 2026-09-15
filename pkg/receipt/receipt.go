// Package receipt records what one fraglet run was: everything bound to the
// container (Invocation) and what came out (Outcome), with every input and
// output referenced by its content hash. The receipt itself is a plain JSON
// file wherever the caller put it; it is not stored by its own hash.
//
// A fraglet is an attested computation in the sense of the Open Knowledge
// Format (SPEC.md §10): a sanctioned script with declared parameter holes an
// agent may fill but not edit. The receipt is the executor's evidence for one
// run, and Invocation is the expanded artifact an attester compares against:
// what actually reached the container, including anything bound outside the
// declared surface (Argv, Env). A run is conformant when those are empty;
// only a conformant run has a memo key, because only then is the key an
// honest identity for the computation.
//
// Identity is derived from content by two formulas that are part of the
// receipt schema, pinned by known-answer vectors in this package's tests:
//
//   - ProcedureHash: "sha256:" + hex(sha256(script file bytes)) — the whole
//     file, shebang included, so a pinned image digest is part of the identity.
//   - MemoKey: params and inputs merged into one map, keys sorted, each
//     rendered "lower(key):value", then
//     sha256(procedureHash + "\n" + "\n" + join(pairs, "\n") + "\n").
//     The empty line between is a reserved runtime-context slot.
//
// Anyone holding a receipt can recompute its key from its own fields;
// nothing here depends on where receipts are kept.
package receipt

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Schema names the receipt shape and is the receipt's only version field.
// Bump it when a field changes meaning or the key formula changes; an older
// receipt is then simply older (a key miss), never a verification problem.
const Schema = "fraglet-receipt/2"

// AnonKey is the anonymous slot: under Inputs it carries the stdin hash,
// under Outputs the stdout result of a script that declares no output=.
const AnonKey = ""

// Execution classes. See Invocation.Class.
const (
	Hermetic      = "hermetic"
	Environmental = "environmental"
)

// Invocation is everything bound to the container for one run: the
// expanded artifact whose identity is the memo key.
type Invocation struct {
	// ProcedureHash is the script's identity (whole file, shebang included).
	ProcedureHash string `json:"procedure_hash"`
	// Image is the container reference the run used and, when resolvable,
	// its registry digest. These are the effective values: a CLI override
	// of the shebang is visible here, and a reader compares them to the
	// sanctioned image. The key does not encode them; a shebang pin is
	// inside ProcedureHash.
	Image       string `json:"image"`
	ImageDigest string `json:"image_digest,omitempty"`
	// Network is the effective network stance ("none", "required", "" when
	// undeclared and not overridden); Mode is FRAGLET_MODE when set.
	Network string `json:"network"`
	Mode    string `json:"mode,omitempty"`
	// StdinMode is how stdin was bound: "none" (not attached), "buffer"
	// (read to EOF, hashed under Inputs[AnonKey], forwarded) or "stream"
	// (live passthrough; nothing hashed, so no key).
	StdinMode string `json:"stdin_mode"`
	// Params are the declared scalar params by alias and value, defaults
	// folded. Inputs are the inputs by content hash: AnonKey -> stdin
	// hash (absent for stream), <alias> -> a :file param's content hash.
	Params map[string]string `json:"params"`
	Inputs map[string]string `json:"inputs"`
	// Argv are undeclared positional script arguments and Env the names of
	// -e variables forwarded into the container (names only: -e is how
	// secrets reach a tool, and a receipt never carries or fingerprints a
	// value). Both are bindings outside the declared parameter surface;
	// a conformant run has neither.
	Argv []string `json:"argv"`
	Env  []string `json:"env"`
}

// Unbound lists why the invocation is not conformant -- what was bound to
// the container outside the declared surface, or left unrecorded -- and is
// nil when it is. A conformant invocation is fully described by its
// declared parameters, so its memo key identifies the computation.
func (i Invocation) Unbound() []string {
	var why []string
	if len(i.Argv) > 0 {
		why = append(why, "positional script arguments are not declared parameters")
	}
	if len(i.Env) > 0 {
		why = append(why, "-e environment is bound outside the declared parameters")
	}
	if _, ok := i.Inputs[AnonKey]; !ok {
		why = append(why, "stdin was streamed, not recorded")
	}
	return why
}

// MemoKey is the memo key of the invocation, and false when Unbound() says
// no honest key exists.
func (i Invocation) MemoKey() (string, bool) {
	if i.Unbound() != nil {
		return "", false
	}
	return memoKey(i.ProcedureHash, i.Params, i.Inputs), true
}

// Class is the execution class a reader derives: Hermetic when the
// invocation fully determines the outputs -- conformant, network=none, and
// an image pinned by digest so the procedure hash covers the image bytes --
// and Environmental otherwise. It is computed, never stored.
func (i Invocation) Class() string {
	if i.Unbound() == nil && i.Network == "none" && strings.Contains(i.Image, "@sha256:") {
		return Hermetic
	}
	return Environmental
}

// Outcome is what the run produced.
type Outcome struct {
	Started  time.Time `json:"started"`
	Finished time.Time `json:"finished"`
	ExitCode int       `json:"exit_code"`
	// Stdout is always recorded, whether or not it is the anonymous output.
	Stdout Digest `json:"stdout"`
	// Outputs: <relpath> -> hash for every file written under /output, or
	// AnonKey -> stdout hash when the script declares no output=.
	Outputs map[string]string `json:"outputs"`
}

// Receipt is the JSON document --receipt writes: Invocation and Outcome,
// flattened, plus the one derived value a consumer cites -- the memo key.
type Receipt struct {
	Schema   string `json:"schema"`
	Fragletc string `json:"fragletc"`
	// Procedure is the script as it was named on the command line, for a
	// reader; ProcedureHash is its identity.
	Procedure string `json:"procedure,omitempty"`

	Invocation

	// MemoKey is Invocation.MemoKey(); absent when the run has no honest key.
	MemoKey string `json:"memo_key,omitempty"`

	Outcome
}

// Build assembles a receipt. The memo key is the only derived field.
func Build(inv Invocation, out Outcome, fragletc, procedure string) Receipt {
	if inv.Params == nil {
		inv.Params = map[string]string{}
	}
	if inv.Inputs == nil {
		inv.Inputs = map[string]string{}
	}
	if inv.Argv == nil {
		inv.Argv = []string{}
	}
	if inv.Env == nil {
		inv.Env = []string{}
	}
	if out.Outputs == nil {
		out.Outputs = map[string]string{}
	}
	r := Receipt{Schema: Schema, Fragletc: fragletc, Procedure: procedure, Invocation: inv, Outcome: out}
	r.MemoKey, _ = inv.MemoKey()
	return r
}

// Write serializes the receipt as indented JSON to path, creating the
// parent directory if needed.
func (r Receipt) Write(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// Digest is a content hash in "sha256:<hex>" form, with length.
type Digest struct {
	Hash  string `json:"hash"`
	Bytes int64  `json:"bytes"`
}

// Hash formats a sha256 sum as "sha256:<hex>".
func Hash(sum [32]byte) string { return "sha256:" + hex.EncodeToString(sum[:]) }

// SumBytes digests a byte slice.
func SumBytes(b []byte) Digest {
	return Digest{Hash: Hash(sha256.Sum256(b)), Bytes: int64(len(b))}
}

// SumFile digests a file's bytes.
func SumFile(path string) (Digest, error) {
	f, err := os.Open(path)
	if err != nil {
		return Digest{}, err
	}
	defer f.Close()
	h := NewHasher()
	if _, err := io.Copy(h, f); err != nil {
		return Digest{}, err
	}
	return h.Digest(), nil
}

// ProcedureHash is a fraglet's identity: the hash of its whole file, shebang
// and header included.
func ProcedureHash(code []byte) string { return Hash(sha256.Sum256(code)) }

// memoKey is the formula in the package doc: params and inputs merge into
// one sorted "lower(key):value" space; the key is
// sha256(procedureHash + "\n" + rctx + "\n" + join(pairs, "\n") + "\n")
// with rctx, a reserved runtime-context slot, always empty here.
func memoKey(procedureHash string, params, inputs map[string]string) string {
	merged := make(map[string]string, len(params)+len(inputs))
	for k, v := range params {
		merged[k] = v
	}
	for k, v := range inputs {
		merged[k] = v
	}
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, strings.ToLower(k)+":"+merged[k])
	}
	combined := procedureHash + "\n" + "" + "\n" + strings.Join(pairs, "\n") + "\n"
	return Hash(sha256.Sum256([]byte(combined)))
}

// Hasher is a streaming sha256 writer, for hashing stdout as it is forwarded.
type Hasher struct {
	h     hash.Hash
	bytes int64
}

// NewHasher returns an empty Hasher.
func NewHasher() *Hasher { return &Hasher{h: sha256.New()} }

func (w *Hasher) Write(p []byte) (int, error) {
	n, err := w.h.Write(p)
	w.bytes += int64(n)
	return n, err
}

// Digest returns what has been written so far.
func (w *Hasher) Digest() Digest {
	var sum [32]byte
	copy(sum[:], w.h.Sum(nil))
	return Digest{Hash: Hash(sum), Bytes: w.bytes}
}
