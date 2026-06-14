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
// Findings are heuristics on real, sometimes wild, market data — review
// them rather than treating every warning as corruption.
func Verify(s *Series, now time.Time) []Issue {
	const (
		maxDailyMove = 0.25 // |daily return| beyond this is suspicious
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

	flatRun := 1
	for k, pt := range s.Points {
		if pt.Close <= 0 || math.IsNaN(pt.Close) || math.IsInf(pt.Close, 0) {
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
		if prev.Close > 0 {
			if r := pt.Close/prev.Close - 1; math.Abs(r) > maxDailyMove {
				warn(pt.Date, "daily move of %+.1f %% — missed split or bad point?", r*100)
			}
		}
		if gap := pt.Date.Sub(prev.Date).Hours() / 24; gap > maxGapDays {
			warn(pt.Date, "no quotes for %.0f days (since %s)", gap, prev.Date.Format("2006-01-02"))
		}
		if pt.Close == prev.Close {
			flatRun++
			if flatRun == maxFlatRun {
				warn(pt.Date, "price unchanged for %d sessions — stale feed?", maxFlatRun)
			}
		} else {
			flatRun = 1
		}
	}
	if age := now.Sub(s.Last().Date).Hours() / 24; age > maxStaleDays {
		warn(s.Last().Date, "last quote is %.0f days old", age)
	}
	return issues
}
