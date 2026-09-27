package marketdata

import (
	"strings"
	"testing"
	"time"
)

// TestSnapshotsCurrentWhenGenerated holds every bundled snapshot to the day it
// was written: its last row must be as recent as its source's release cadence
// allows on its "# generated:" date. It is what catches a source that stopped
// while still answering, which a refresh otherwise records faithfully: on
// 2026-09-27 hicp-fr.csv was regenerated and still ended at 2025-12, because
// Eurostat had moved the HICP to a new dataset and frozen the one read here.
//
// The allowances are cmd/gen-snapshots' (dailyMaxLag, monthlyMaxLag), which
// refuses to write such a file in the first place: three weeks after a daily
// row, ten weeks after the end of a monthly row's month. The test compares two
// dates inside each file, never the clock, so it cannot rot as time passes.
func TestSnapshotsCurrentWhenGenerated(t *testing.T) {
	const (
		dailyMaxLag   = 21 * 24 * time.Hour
		monthlyMaxLag = 70 * 24 * time.Hour
	)
	for name, csv := range map[string]string{
		"vix.csv":         vixSnapshot,
		"cpi-us.csv":      cpiUSSnapshot,
		"hicp-fr.csv":     hicpFRSnapshot,
		"eurusd-long.csv": eurusdLongCSV,
		"gbpusd-long.csv": gbpusdLongCSV,
		"jpyusd-long.csv": jpyusdLongCSV,
		"chfusd-long.csv": chfusdLongCSV,
	} {
		var generated time.Time
		var last string
		for line := range strings.SplitSeq(csv, "\n") {
			line = strings.TrimSpace(line)
			if stamp, ok := strings.CutPrefix(line, "# generated:"); ok {
				generated, _ = time.Parse("2006-01-02", strings.TrimSpace(stamp))
				continue
			}
			if line != "" && !strings.HasPrefix(line, "#") {
				last, _, _ = strings.Cut(line, ",")
			}
		}
		if generated.IsZero() {
			t.Errorf("%s: no \"# generated: YYYY-MM-DD\" stamp", name)
			continue
		}
		var end time.Time
		var limit time.Duration
		if d, err := time.Parse("2006-01-02", last); err == nil {
			end, limit = d, dailyMaxLag
		} else if m, err := time.Parse("2006-01", last); err == nil {
			end, limit = m.AddDate(0, 1, 0), monthlyMaxLag
		} else {
			t.Errorf("%s: last row %q is not a date", name, last)
			continue
		}
		if lag := generated.Sub(end); lag > limit {
			t.Errorf("%s: generated %s but ends at %s, %.0f days earlier: its source has stopped updating",
				name, generated.Format("2006-01-02"), last, lag.Hours()/24)
		}
	}
}
