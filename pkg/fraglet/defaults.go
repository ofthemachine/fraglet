package fraglet

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// dynamicDefaultLayouts maps the closed @token vocabulary to a time layout.
// These are *named instants*, not a templating language: each token names the
// granularity at which the value, and so the memo bucket, advances. A tool
// declares its own freshness window beside the code that fetches
// (param=call_time:default=@hour); the resolved value is a param like any
// other, recorded in the receipt and part of its memo key, so the truncation
// is the cache's scope. Deliberately closed: no arithmetic (@day-3), no format
// strings. Want a different value? Pass it explicitly.
var dynamicDefaultLayouts = map[string]string{
	"@year":   "2006",
	"@month":  "2006-01",
	"@day":    "2006-01-02",
	"@hour":   "2006-01-02T15",
	"@minute": "2006-01-02T15:04",
	"@iso":    time.RFC3339,
}

// ResolveDefault expands a default= value against one observation instant. A
// leading '@' selects a dynamic value from the closed vocabulary; anything
// else is returned verbatim as a static literal. The instant is passed in, not
// read from the clock here, so every default resolved for one run (or a chain
// of runs a caller makes) shares one reading: a chain that straddles an @hour
// boundary must not bucket its steps into different hours. Resolution is in
// UTC, so a bucket is the same whatever the host's time zone.
func ResolveDefault(raw string, now time.Time) (string, error) {
	if !strings.HasPrefix(raw, "@") {
		return raw, nil
	}
	now = now.UTC()
	if raw == "@epoch_ms" {
		return strconv.FormatInt(now.UnixMilli(), 10), nil
	}
	layout, ok := dynamicDefaultLayouts[raw]
	if !ok {
		return "", fmt.Errorf("unknown dynamic default %q (valid: @year @month @day @hour @minute @iso @epoch_ms)", raw)
	}
	return now.Format(layout), nil
}
