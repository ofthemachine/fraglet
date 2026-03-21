package main

import (
	"reflect"
	"testing"
)

func TestPreprocessFragletArgv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		args      []string
		wantTail  []string
		wantHelp  bool
		wantParam []string
		wantErr   bool
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotTail, gotHelp, gotParam, err := preprocessFragletArgv(tt.args)
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
			if !reflect.DeepEqual(gotTail, tt.wantTail) {
				t.Errorf("tail: got %#v want %#v", gotTail, tt.wantTail)
			}
		})
	}
}
