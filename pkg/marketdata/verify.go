package marketdata

import (
	"fmt"
	"math"
	"time"
)

// Issue is a data-quality problem found by Verify.
type Issue struct {
	// Severity is "warn" for findings worth a look (suspicious moves,
	// stale feeds) and "error" for data that breaks computations
	// (non-positive prices).
	Severity string
	Date     time.Time // first date concerned (zero when global)
	Message  string
}

// String renders the issue as a one-line "[severity] message", prefixed with
// the concerned date when the issue is tied to a specific quote.
func (i Issue) String() string {
	if i.Date.IsZero() {
		return fmt.Sprintf("[%s] %s", i.Severity, i.Message)
	}
	return fmt.Sprintf("[%s] %s: %s", i.Severity, i.Date.Format("2006-01-02"), i.Message)
}

// Verify inspects a series for common data-quality problems: non-positive
// prices, suspiciously large daily moves (missed split or bad point),
// calendar gaps, long flat stretches (stale feed) and a stale last quote.
// now anchors the staleness check (pass time.Now() outside tests).
//
// Two families are judged by their own rules. A RATE series (^IRX, ^ESTR,
// ^ECB-DFR, …) is an annualized percent level, not a price: zero and negative
// values are legitimate readings of the world, a ratio between two levels near
// zero means nothing, and a policy rate stays flat for months by design, so the
// non-positive, relative-move and flat-run checks would fire on correct data.
// And a series is measured against ITS OWN cadence: a monthly rate or a weekly
// fund NAV is not stale because it has no quote from last Tuesday.
//
// Findings are heuristics on real, sometimes wild, market data; review
// them rather than treating every warning as corruption.
func Verify(s *Series, now time.Time) []Issue {
	const (
		maxDailyMove = 0.25 // |daily return| beyond this is suspicious
		maxRateJump  = 3.0  // percentage points in a day, for a rate level
		maxGapDays   = 14   // calendar days without a quote
		maxFlatRun   = 20   // consecutive identical closes
		maxStaleDays = 10   // calendar days since the last quote
	)
	var issues []Issue
	warn := func(d time.Time, format string, args ...any) {
		issues = append(issues, Issue{Severity: "warn", Date: d, Message: fmt.Sprintf(format, args...)})
	}
	if len(s.Points) == 0 {
		return []Issue{{Severity: "error", Message: "no quotes at all"}}
	}
	rate := isRateSymbol(s.Symbol) || isPolicyRate(s.Symbol)
	// A series reports at its own pace; the calendar thresholds follow it.
	cadence := medianSpacingDays(s.Points)
	gapLimit, staleLimit := float64(maxGapDays), float64(maxStaleDays)
	if lim := 3 * cadence; lim > gapLimit {
		gapLimit, staleLimit = lim, lim
	}

	flatRun := 1
	for k, pt := range s.Points {
		if math.IsNaN(pt.Close) || math.IsInf(pt.Close, 0) || (pt.Close <= 0 && !rate) {
			issues = append(issues, Issue{Severity: "error", Date: pt.Date,
				Message: fmt.Sprintf("non-positive price (%g)", pt.Close)})
			continue
		}
		if k == 0 {
			continue
		}
		prev := s.Points[k-1]
		if !pt.Date.After(prev.Date) {
			issues = append(issues, Issue{Severity: "error", Date: pt.Date,
				Message: "dates not strictly increasing"})
		}
		switch {
		case rate:
			if d := pt.Close - prev.Close; math.Abs(d) > maxRateJump {
				warn(pt.Date, "rate jumped %+.2f points in a day, bad point?", d)
			}
		case prev.Close > 0:
			if r := pt.Close/prev.Close - 1; math.Abs(r) > maxDailyMove {
				warn(pt.Date, "daily move of %+.1f %%, missed split or bad point?", r*100)
			}
		}
		if gap := pt.Date.Sub(prev.Date).Hours() / 24; gap > gapLimit {
			warn(pt.Date, "no quotes for %.0f days (since %s)", gap, prev.Date.Format("2006-01-02"))
		}
		if pt.Close == prev.Close {
			flatRun++
			if flatRun == maxFlatRun && !rate {
				warn(pt.Date, "price unchanged for %d sessions, stale feed?", maxFlatRun)
			}
		} else {
			flatRun = 1
		}
	}
	if age := now.Sub(s.Last().Date).Hours() / 24; age > staleLimit {
		warn(s.Last().Date, "last quote is %.0f days old", age)
	}
	return issues
}

// medianSpacingDays is the typical number of calendar days between two
// consecutive quotes: 1 for a daily series (weekends included, the median
// absorbs them), 7 for a weekly NAV, ~30 for a monthly publication.
func medianSpacingDays(pts []Point) float64 {
	if len(pts) < 3 {
		return 1
	}
	gaps := make([]float64, 0, len(pts)-1)
	for k := 1; k < len(pts); k++ {
		gaps = append(gaps, pts[k].Date.Sub(pts[k-1].Date).Hours()/24)
	}
	sortFloats(gaps)
	return gaps[len(gaps)/2]
}

// sortFloats is an insertion sort, adequate for the sizes here and cheaper
// than importing sort for one call.
func sortFloats(v []float64) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && v[j] < v[j-1]; j-- {
			v[j], v[j-1] = v[j-1], v[j]
		}
	}
}
