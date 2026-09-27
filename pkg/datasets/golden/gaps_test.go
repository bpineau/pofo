package golden

import (
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// A HOLE is a stretch of calendar a bundled series simply does not cover, and
// it is the most expensive defect a reference file can carry, because nothing
// downstream can see it. A consumer reads one row, then the next, and computes
// the return between them: across a hole that is one enormous "period" followed
// by no volatility at all for however long the hole lasts. Statistics computed
// on it are not noisy, they are wrong, and they look plausible.
//
// TREASURY-LONG-USD shipped one for years. The Federal Reserve discontinued the
// 20-year constant-maturity yield between 1987-01 and 1993-09, the file was
// built from that yield alone, and so it jumped from 1986-12 straight to
// 1993-10: a single +60 % step, seven years of zero variance, and a FIRE book
// plate that had to rebuild its own long Treasury leg rather than read the
// bundled one. It was invisible to every other test here, all of which measure
// calendar-year returns or long-window CAGRs, both of which a hole leaves
// almost untouched. Hence this guard: it measures COVERAGE, which nothing else
// does, over every bundled file at once.
//
// The rule is marketdata.FindGaps, the one the data doctor reports: a step
// longer than one and a half of the series' LOCAL steps, and never under
// fourteen days. A monthly file that skips a month (a 59-day step against a
// ~46-day limit) fails, and so does a daily file silent for three weeks,
// whatever its cadence and wherever it changes: the rule reads each stretch at
// the pace it kept, so a new reference file is covered the day it lands.

// knownGaps are the bundled series allowed a step past the rule, with the
// longest step they are allowed and why. An entry here is a defect that has
// been measured and deliberately left, never a tolerance widened to make a
// test pass; it must name the step it covers, so that a SECOND, larger hole in
// the same file still fails.
//
// No monthly file has an entry. The last was EUROGOV-LONG-EUR, whose deep
// segment (the OECD 10-year yield mapped to a 25-year one) was dated the
// first of the month while the ECB-curve segment taking over in 2004-09 was
// dated month-end, so the one step across the junction spanned 60 days and
// carried two months of return. The sweep of 2026-09 gave every OECD-driven
// monthly reference the month-end label the rest of the bundle uses
// (cmd/gen-euro-refdata's atMonthEnd), which closed that junction to 30 days,
// and both generators now refuse to write a file whose longest monthly step
// exceeds 45 days.
//
// The entries below arrived when the guard stopped reading monthly files
// only (2026-09-27): each is a REAL provider calendar inside a real-quote or
// real-donor era, a fund that did not price for a while, never a hole the
// pipeline made. They are measured, dated and pinned to their longest step.
var knownGaps = map[string]struct {
	maxDays int
	why     string
}{
	"simdata/DBXG": {35, "the fund's own Xetra closes over its first two years (2007-08 to 2009-08) " +
		"are up to 35 days apart, a thinly traded new listing; real quotes, spliced as served. Measured 2026-09-27."},
	"simdata/IE00B3Q8M574": {15, "2025-05-12 to 2025-05-27, one weekly NAV of the fund itself not " +
		"published. Measured 2026-09-27."},
	"simdata/LI0049587301": {28, "the fund's own weekly-to-fortnightly NAV skips a date around Easter " +
		"and year-end (seven steps, 2019 to 2025). Measured 2026-09-27."},
	"simdata/LI0115208543": {16, "the fund's own weekly NAV skips the Christmas week (2016, 2022, " +
		"2023). Measured 2026-09-27."},
	"simdata/LU1832174962": {32, "the donor's own NAV (LU0131510165): monthly steps among " +
		"semi-monthly ones in 1996, semi-monthly among weekly ones in late 2009. Measured 2026-09-27."},
}

// TestBundledSeriesHaveNoHoles walks every embedded refdata and simdata file
// and refuses a step its own pace does not allow.
func TestBundledSeriesHaveNoHoles(t *testing.T) {
	checked := 0
	for _, dir := range []struct {
		name string
		fsys fs.FS
	}{{"refdata", datasets.Refdata()}, {"simdata", datasets.Simdata()}} {
		names, err := fs.Glob(dir.fsys, "*.csv")
		if err != nil {
			t.Fatalf("%s: %v", dir.name, err)
		}
		if len(names) == 0 {
			t.Fatalf("%s: no bundled series at all", dir.name)
		}
		for _, name := range names {
			id := strings.TrimSuffix(name, ".csv")
			key := dir.name + "/" + id
			s, ok, err := marketdata.ReadSimdataFS(dir.fsys, id)
			if err != nil || !ok {
				t.Errorf("%s: ok=%v err=%v", key, ok, err)
				continue
			}
			checked++
			for _, g := range marketdata.FindGaps(s) {
				if known, ok := knownGaps[key]; ok && g.Days <= float64(known.maxDays) {
					continue
				}
				t.Errorf("%s: %.0f days between %s and %s, beyond the %.0f its pace allows",
					key, g.Days, g.From.Format(time.DateOnly), g.To.Format(time.DateOnly), g.Limit)
			}
		}
	}
	if checked < 80 {
		t.Errorf("only %d bundled series inspected, the guard is not seeing the bundle", checked)
	}
}

// TestKnownGapsAreStillThere keeps the exception list honest: an entry whose
// file no longer steps that far is a note about a file that has moved on,
// and its allowance would silently cover a new hole.
func TestKnownGapsAreStillThere(t *testing.T) {
	for key, known := range knownGaps {
		dir, id, _ := cutKey(key)
		s, ok, err := marketdata.ReadSimdataFS(bundleDir(dir), id)
		if err != nil || !ok {
			t.Errorf("knownGaps names %s, which is not bundled (ok=%v err=%v)", key, ok, err)
			continue
		}
		longest := 0.0
		for _, g := range marketdata.FindGaps(s) {
			longest = max(longest, g.Days)
		}
		if int(longest+0.5) != known.maxDays {
			t.Errorf("knownGaps[%s] allows %d days, the file's longest gap is now %.0f: update or drop the entry", key, known.maxDays, longest)
		}
	}
}
