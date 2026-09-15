package receipt

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"testing"
)

// Known-answer vectors for the memo key formula, fixed on 2026-09-13. A
// receipt's key is only useful if every implementation agrees on it, so if
// this test fails the formula moved: fix the formula, never the vector.
func TestMemoKey_KnownAnswers(t *testing.T) {
	proc := "sha256:0000000000000000000000000000000000000000000000000000000000000001"
	emptyStdin := "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	got := memoKey(proc,
		map[string]string{"Location": "Mount Fuji", "day": "2"},
		map[string]string{AnonKey: emptyStdin, "tex_source": "sha256:00000000000000000000000000000000000000000000000000000000000000ab"},
	)
	if want := "sha256:3ac95ff2169a710e42e08097b4b38f73f5223d5e947b50432d7825436e02d2e6"; got != want {
		t.Fatalf("MemoKeyV2 = %s, want %s", got, want)
	}

	got = memoKey(proc, nil, map[string]string{AnonKey: emptyStdin})
	if want := "sha256:c9d8925c0d91e46b714c48cdc278fb3a7dd7bb89b81092dda6016b8c612c4460"; got != want {
		t.Fatalf("MemoKeyV2 (no params) = %s, want %s", got, want)
	}
}

func TestMemoKey_ParamOrderIrrelevant(t *testing.T) {
	a := memoKey("sha256:p", map[string]string{"x": "1", "y": "2"}, map[string]string{AnonKey: "sha256:s"})
	b := memoKey("sha256:p", map[string]string{"y": "2", "x": "1"}, map[string]string{AnonKey: "sha256:s"})
	if a != b {
		t.Fatal("key depends on map iteration order")
	}
}

func TestProcedureHash_IsWholeFile(t *testing.T) {
	code := []byte("#!/usr/bin/env -S fragletc --image x@sha256:abc\n#: d=T\nprint(1)\n")
	if got, want := ProcedureHash(code), Hash(sha256.Sum256(code)); got != want {
		t.Fatalf("ProcedureHash = %s, want %s", got, want)
	}
	// The shebang is part of the identity: repinning the image changes it.
	other := []byte("#!/usr/bin/env -S fragletc --image x@sha256:def\n#: d=T\nprint(1)\n")
	if ProcedureHash(code) == ProcedureHash(other) {
		t.Fatal("image pin did not change the procedure hash")
	}
}

func TestSumBytes_EmptyInputHash(t *testing.T) {
	if got := SumBytes(nil).Hash; got != "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Fatalf("empty hash = %s", got)
	}
}

func TestHasher_StreamsLikeSum(t *testing.T) {
	h := NewHasher()
	h.Write([]byte("hello, "))
	h.Write([]byte("world"))
	if got, want := h.Digest(), SumBytes([]byte("hello, world")); got != want {
		t.Fatalf("hasher = %+v, want %+v", got, want)
	}
}

const emptyStdin = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

func conformant() Invocation {
	return Invocation{
		ProcedureHash: "sha256:0000000000000000000000000000000000000000000000000000000000000001",
		Image:         "ofthemachine/python3@sha256:abc",
		Network:       "none",
		StdinMode:     "buffer",
		Params:        map[string]string{"Location": "Mount Fuji", "day": "2"},
		Inputs:        map[string]string{AnonKey: emptyStdin, "tex_source": "sha256:00000000000000000000000000000000000000000000000000000000000000ab"},
	}
}

// A conformant invocation's key is the formula over its own fields: the
// first vector in TestMemoKey_KnownAnswers.
func TestInvocation_MemoKeyIsTheFormula(t *testing.T) {
	key, ok := conformant().MemoKey()
	if !ok || key != "sha256:3ac95ff2169a710e42e08097b4b38f73f5223d5e947b50432d7825436e02d2e6" {
		t.Fatalf("MemoKey = %s, %v", key, ok)
	}
}

func TestInvocation_UnboundHasNoKey(t *testing.T) {
	cases := map[string]func(*Invocation){
		"argv":   func(i *Invocation) { i.Argv = []string{"foo"} },
		"env":    func(i *Invocation) { i.Env = []string{"FOO"} },
		"stream": func(i *Invocation) { i.StdinMode = "stream"; delete(i.Inputs, AnonKey) },
	}
	for name, mutate := range cases {
		inv := conformant()
		mutate(&inv)
		if why := inv.Unbound(); len(why) != 1 {
			t.Fatalf("%s: Unbound = %v", name, why)
		}
		if _, ok := inv.MemoKey(); ok {
			t.Fatalf("%s: expected no memo key", name)
		}
		if inv.Class() != Environmental {
			t.Fatalf("%s: Class = %s", name, inv.Class())
		}
	}
	if why := conformant().Unbound(); why != nil {
		t.Fatalf("conformant Unbound = %v", why)
	}
}

func TestInvocation_Class(t *testing.T) {
	inv := conformant()
	if inv.Class() != Hermetic {
		t.Fatal("pinned + network=none + conformant should be hermetic")
	}
	inv.Image = "ofthemachine/python3:latest"
	if inv.Class() != Environmental {
		t.Fatal("unpinned image cannot be hermetic")
	}
	inv = conformant()
	inv.Network = "required"
	if inv.Class() != Environmental {
		t.Fatal("network=required cannot be hermetic")
	}
}

func TestBuild_RoundTripsFlatJSON(t *testing.T) {
	inv := conformant()
	inv.Argv, inv.Env = nil, nil
	out := Outcome{ExitCode: 0, Stdout: SumBytes([]byte("hi\n")), Outputs: map[string]string{AnonKey: SumBytes([]byte("hi\n")).Hash}}
	r := Build(inv, out, "v0.13.0", "tool.py")
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"schema", "fragletc", "procedure", "procedure_hash", "image", "network", "stdin_mode", "params", "inputs", "argv", "env", "memo_key", "started", "finished", "exit_code", "stdout", "outputs"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("missing %q in %s", k, data)
		}
	}
	for _, k := range []string{"execution_class", "memo_key_version", "files", "invocation", "outcome"} {
		if _, ok := m[k]; ok {
			t.Fatalf("unexpected %q in %s", k, data)
		}
	}
	if string(m["argv"]) != "[]" || string(m["env"]) != "[]" {
		t.Fatalf("nil argv/env should marshal as empty lists: %s", data)
	}
	var back Receipt
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.MemoKey != r.MemoKey || back.Class() != Hermetic || back.Stdout != out.Stdout {
		t.Fatalf("round trip lost data: %+v", back)
	}

	inv.StdinMode = "stream"
	delete(inv.Inputs, AnonKey)
	data, _ = json.Marshal(Build(inv, out, "v0.13.0", "tool.py"))
	if bytes.Contains(data, []byte("memo_key")) {
		t.Fatalf("stream receipt must not carry a memo key: %s", data)
	}
}
