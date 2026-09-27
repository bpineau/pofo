package main

import (
	"strings"
	"testing"
	"time"
)

// TestStale pins the freshness guard: a monthly label is dated at the end of
// its month and a daily one at its day, each against its own allowance, so a
// source frozen for months (Eurostat's old HICP dataset, stuck at 2025-12) is
// refused while one a release behind is not.
func TestStale(t *testing.T) {
	now := time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		last  string
		stale bool
	}{
		{"2026-08", false},    // the latest release
		{"2026-07", false},    // one release behind, 58 days after its month ended
		{"2026-06", true},     // 88 days: a missed release
		{"2025-12", true},     // the frozen dataset
		{"2026-09-18", false}, // FRED's weekly H.10 lag
		{"2026-08-31", true},  // four weeks of silence from a daily source
		{"n/a", false},        // not a date: not judged
	} {
		got := stale(c.last, now)
		if (got != "") != c.stale {
			t.Errorf("stale(%q) = %q, want stale=%v", c.last, got, c.stale)
		}
		if got != "" && !strings.Contains(got, c.last) {
			t.Errorf("stale(%q) = %q, which does not name the last row", c.last, got)
		}
	}
}
