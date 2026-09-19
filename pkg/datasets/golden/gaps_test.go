package golden

import (
	"io/fs"
	"slices"
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
// The rule applies to the MONTHLY (or coarser) series, identified by their own
// median step rather than by a list, so a new reference file is covered the day
// it lands. A daily series with a three-week hole would deserve its own,
// stricter rule; this one would not catch it, and saying so is cheaper than
// pretending otherwise.
const (
	// monthlyIfStepAbove is the median step above which a series is read as
	// monthly rather than daily or weekly. Fifteen days sits clear of both.
	monthlyIfStepAbove = 15 * 24 * time.Hour
	// maxMonthlyGap is the longest step a monthly series may take. A calendar
	// month is at most 31 days, and a month-end series crossing a leap year can
	// legitimately stretch a little past that; 45 days is comfortably short of
	// the 59 a single skipped month would produce.
	maxMonthlyGap = 45 * 24 * time.Hour
)

// knownMonthlyGaps are the bundled monthly series allowed to exceed
// maxMonthlyGap, with the step they are allowed and why. An entry here is a
// defect that has been measured and deliberately left, never a tolerance
// widened to make a test pass; it must name the step it covers, so that a
// SECOND, larger hole in the same file still fails.
var knownMonthlyGaps = map[string]struct {
	maxDays int
	why     string
}{
	// EUROGOV-LONG-EUR changes dating convention at its 2004-09 junction: the
	// deep segment (the OECD 10-year yield mapped to a 25-year one) is dated the
	// first of the month, the ECB-curve segment that takes over is dated
	// month-end, so the one step across the junction spans 60 days and carries
	// two months of return. It is the same class of defect the Treasury files
	// were rebuilt for and it is worth one day's work, but it is one step in a
	// 681-point file, 22 years before the fund it stands behind starts quoting,
	// and fixing it means re-deriving the pre-2004 map. Named, measured, left.
	"EUROGOV-LONG-EUR": {maxDays: 60, why: "first-of-month to month-end dating junction at 2004-09"},
}

// TestBundledSeriesHaveNoHoles walks every embedded refdata and simdata file
// and refuses a monthly series that skips a month.
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
			s, ok, err := marketdata.ReadSimdataFS(dir.fsys, id)
			if err != nil || !ok {
				t.Errorf("%s/%s: ok=%v err=%v", dir.name, name, ok, err)
				continue
			}
			if len(s.Points) < 3 {
				continue
			}
			if medianStep(s) <= monthlyIfStepAbove {
				continue // daily or weekly: out of this guard's scope
			}
			checked++
			limit := maxMonthlyGap
			if known, ok := knownMonthlyGaps[id]; ok {
				limit = time.Duration(known.maxDays) * 24 * time.Hour
			}
			for i := 1; i < len(s.Points); i++ {
				gap := s.Points[i].Date.Sub(s.Points[i-1].Date)
				if gap <= limit {
					continue
				}
				t.Errorf("%s/%s: %.0f days between %s and %s, i.e. %d month(s) the series does not cover",
					dir.name, id, gap.Hours()/24,
					s.Points[i-1].Date.Format("2006-01-02"), s.Points[i].Date.Format("2006-01-02"),
					int(gap.Hours()/24/30))
			}
		}
	}
	t.Logf("%d monthly series inspected", checked)
	if checked < 10 {
		t.Errorf("only %d monthly series inspected, the guard is not seeing the bundle", checked)
	}
}

// medianStep is the median interval between consecutive points, which tells a
// monthly series from a daily one without trusting a name or a header. The
// median rather than the mean, precisely because a hole must not be allowed to
// promote a daily series into the monthly bucket where it would be graded
// leniently.
func medianStep(s *marketdata.Series) time.Duration {
	steps := make([]time.Duration, 0, len(s.Points)-1)
	for i := 1; i < len(s.Points); i++ {
		steps = append(steps, s.Points[i].Date.Sub(s.Points[i-1].Date))
	}
	slices.Sort(steps)
	return steps[len(steps)/2]
}
