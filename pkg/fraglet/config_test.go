package fraglet

import "testing"

func TestInjectionTypeHelpers(t *testing.T) {
	tests := []struct {
		name      string
		inj       InjectionConfig
		wantLine  bool
		wantRange bool
		wantFile  bool
		wantEmpty bool
	}{
		{
			name:      "default empty injection",
			inj:       InjectionConfig{},
			wantEmpty: true,
		},
		{
			name:     "line injection",
			inj:      InjectionConfig{Match: "FRAGLET"},
			wantLine: true,
		},
		{
			name:      "range injection",
			inj:       InjectionConfig{MatchStart: "START", MatchEnd: "END"},
			wantRange: true,
		},
		{
			name:     "file replacement only codePath",
			inj:      InjectionConfig{CodePath: "/code/file.sh"},
			wantFile: true,
		},
		{
			name:     "line injection with explicit codePath",
			inj:      InjectionConfig{CodePath: "/code/file.sh", Match: "FRAGLET"},
			wantLine: true,
		},
		{
			name:      "range injection with explicit codePath",
			inj:       InjectionConfig{CodePath: "/code/file.sh", MatchStart: "START", MatchEnd: "END"},
			wantRange: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := isLineInjection(tt.inj); got != tt.wantLine {
				t.Fatalf("isLineInjection() = %v, want %v", got, tt.wantLine)
			}
			if got := isRangeInjection(tt.inj); got != tt.wantRange {
				t.Fatalf("isRangeInjection() = %v, want %v", got, tt.wantRange)
			}
			if got := isFileInjection(tt.inj); got != tt.wantFile {
				t.Fatalf("isFileInjection() = %v, want %v", got, tt.wantFile)
			}
			if got := isEmptyInjection(tt.inj); got != tt.wantEmpty {
				t.Fatalf("isEmptyInjection() = %v, want %v", got, tt.wantEmpty)
			}
		})
	}
}
