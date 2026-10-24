package fraglet

import (
	"testing"
	"time"
)

func TestResolveDefault(t *testing.T) {
	now := time.Date(2026, 6, 16, 14, 32, 5, 0, time.UTC)
	for raw, want := range map[string]string{
		"@year":      "2026",
		"@month":     "2026-06",
		"@day":       "2026-06-16",
		"@hour":      "2026-06-16T14",
		"@minute":    "2026-06-16T14:32",
		"@iso":       "2026-06-16T14:32:05Z",
		"@epoch_ms":  "1781620325000",
		"1970-01-01": "1970-01-01",
		"metric":     "metric",
		"":           "",
	} {
		got, err := ResolveDefault(raw, now)
		if err != nil || got != want {
			t.Errorf("ResolveDefault(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}
	if _, err := ResolveDefault("@day-3", now); err == nil {
		t.Error("the vocabulary is closed: @day-3 must be refused")
	}
}
