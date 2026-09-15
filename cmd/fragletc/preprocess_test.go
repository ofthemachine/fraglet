package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ofthemachine/fraglet/pkg/engine"
	"github.com/ofthemachine/fraglet/pkg/receipt"
)

func TestPreprocessFragletArgv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		args       []string
		wantTail   []string
		wantHelp   bool
		wantParam  []string
		wantOutput []string
		wantErr    bool
	}{
		{
			name:     "help only",
			args:     []string{"--fraglet-help", "a.py"},
			wantTail: []string{"a.py"},
			wantHelp: true,
		},
		{
			name:      "p after script",
			args:      []string{"a.py", "x", "-p", "k=v"},
			wantTail:  []string{"a.py", "x"},
			wantParam: []string{"k=v"},
		},
		{
			name:      "param equals",
			args:      []string{"--param=k=v", "a.py"},
			wantTail:  []string{"a.py"},
			wantParam: []string{"k=v"},
		},
		{
			name:      "bundled p key value",
			args:      []string{"-pcity=paris"},
			wantTail:  nil,
			wantParam: []string{"city=paris"},
		},
		{
			name:     "path not bundled",
			args:     []string{"-path", "/tmp"},
			wantTail: []string{"-path", "/tmp"},
		},
		{
			name:     "passthrough after double dash",
			args:     []string{"a.py", "--", "-p", "a=b"},
			wantTail: []string{"a.py", "--", "-p", "a=b"},
		},
		{
			name:      "gobble p before double dash then pass through",
			args:      []string{"a.py", "-p", "x=1", "--", "-p", "y=2"},
			wantTail:  []string{"a.py", "--", "-p", "y=2"},
			wantParam: []string{"x=1"},
		},
		{
			name:     "missing p value",
			args:     []string{"-p"},
			wantTail: nil,
			wantErr:  true,
		},
		{
			name:       "output equals",
			args:       []string{"--output=series.csv", "a.py"},
			wantTail:   []string{"a.py"},
			wantOutput: []string{"series.csv"},
		},
		{
			name:       "output value form",
			args:       []string{"--output", "series.csv=/tmp/out.csv", "a.py"},
			wantTail:   []string{"a.py"},
			wantOutput: []string{"series.csv=/tmp/out.csv"},
		},
		{
			name:     "missing output value",
			args:     []string{"--output"},
			wantTail: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := preprocessFragletArgv(tt.args)
			gotTail, gotHelp, gotParam, gotOutput := got.filtered, got.wantHelp, got.params, got.outputs
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if gotHelp != tt.wantHelp {
				t.Errorf("help: got %v want %v", gotHelp, tt.wantHelp)
			}
			if !reflect.DeepEqual(gotParam, tt.wantParam) {
				t.Errorf("params: got %#v want %#v", gotParam, tt.wantParam)
			}
			if !reflect.DeepEqual(gotOutput, tt.wantOutput) {
				t.Errorf("outputs: got %#v want %#v", gotOutput, tt.wantOutput)
			}
			if !reflect.DeepEqual(gotTail, tt.wantTail) {
				t.Errorf("tail: got %#v want %#v", gotTail, tt.wantTail)
			}
		})
	}
}

func TestPreprocess_ReceiptAnyPosition(t *testing.T) {
	for _, args := range [][]string{
		{"--receipt", "r.json", "a.py", "-p", "k=v"},
		{"a.py", "-p", "k=v", "--receipt", "r.json"},
		{"a.py", "--receipt=r.json"},
	} {
		got, err := preprocessFragletArgv(args)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		tail, receipt := got.filtered, got.receipt
		if receipt != "r.json" {
			t.Fatalf("%v: receipt = %q", args, receipt)
		}
		for _, a := range tail {
			if a == "--receipt" || a == "r.json" || a == "--receipt=r.json" {
				t.Fatalf("%v: receipt leaked into tail %v", args, tail)
			}
		}
	}
	got, err := preprocessFragletArgv([]string{"a.py", "--", "--receipt", "r.json"})
	if err != nil || got.receipt != "" || len(got.filtered) != 4 {
		t.Fatalf("after --: %+v err=%v", got, err)
	}
	if _, err := preprocessFragletArgv([]string{"a.py", "--receipt"}); err == nil {
		t.Fatal("bare --receipt should error")
	}
}

func TestPreprocess_StdinAnyPosition(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"a.py", "-p", "k=v", "--stdin=buffer"}, "buffer"},
		{[]string{"--stdin", "stream", "a.py"}, "stream"},
		{[]string{"a.py"}, ""},
		{[]string{"a.py", "--stdin", "none"}, "none"},
	} {
		got, err := preprocessFragletArgv(tc.args)
		if err != nil || got.stdin != tc.want {
			t.Fatalf("%v: stdin=%q err=%v", tc.args, got.stdin, err)
		}
		for _, a := range got.filtered {
			if strings.HasPrefix(a, "--stdin") || a == "stream" || a == "none" {
				t.Fatalf("%v: leaked into tail %v", tc.args, got.filtered)
			}
		}
	}
	for _, args := range [][]string{{"a.py", "--stdin=pipe"}, {"--stdin", "Buffer", "a.py"}, {"a.py", "--stdin"}} {
		if _, err := preprocessFragletArgv(args); err == nil {
			t.Fatalf("%v: expected an error", args)
		}
	}
}

func TestReceiptDestination(t *testing.T) {
	inv := receipt.Invocation{ProcedureHash: "sha256:p", Inputs: map[string]string{receipt.AnonKey: "sha256:s"}}
	report := engine.RunReport{Invocation: inv, Outcome: receipt.Outcome{Started: time.Date(2026, 9, 14, 6, 6, 59, 0, time.UTC)}}
	if got := receiptDestination("r.json", "/tmp/dir", "tool.py", report); got != "r.json" {
		t.Fatalf("explicit wins: %q", got)
	}
	if got := receiptDestination("", "", "tool.py", report); got != "" {
		t.Fatalf("no dir, no receipt: %q", got)
	}
	key, _ := inv.MemoKey()
	want := filepath.Join("/tmp/dir", "20260914T060659Z-tool-"+strings.TrimPrefix(key, "sha256:")[:12]+".json")
	if got := receiptDestination("", "/tmp/dir", "some/where/tool.py", report); got != want {
		t.Fatalf("keyed name = %q, want %q", got, want)
	}
	report.Argv = []string{"x"}
	if got := receiptDestination("", "/tmp/dir", "", report); got != filepath.Join("/tmp/dir", "20260914T060659Z-inline-unkeyed.json") {
		t.Fatalf("unkeyed name = %q", got)
	}
}
