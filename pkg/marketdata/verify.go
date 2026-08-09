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
// And a series is measured against ITS OWN cadence, LOCALLY: a monthly rate or
// a weekly fund NAV is not stale because it has no quote from last Tuesday, and
// a series that reported monthly for eighty years before turning daily is
// judged by the pace it kept at the time. Without that, every one of the S&P
// 500's month-ends before 1928 reads as a three-week outage and every one of
// its month-over-month moves as a bad print.
//
// Findings are heuristics on real, sometimes wild, market data; review
// them rather than treating every warning as corruption.
//
// VerifyAsset wraps Verify with the checks that need the catalog record, and
// judges the same moves against the asset class's own limit instead of the
// blanket one below.
func Verify(s *Series, now time.Time) []Issue {
	return verify(s, now, 0)
}

// maxMove is the |return| beyond which a single observation is suspicious, for
// a series about which nothing is known but its cadence. A quarter in one
// session is beyond any unlevered asset; VerifyAsset replaces it with the
// asset class's own Move, which is both tighter (a money market has no business
// moving 3.5 %) and looser where it must be (a daily 3x fund is allowed three
// times the equity limit).
const maxMove = 0.25

// verify is Verify with an optional per-observation move limit: 0 asks for the
// blanket maxMove, anything positive replaces it. Either way the limit is
// scaled by the square root of the LOCAL cadence, because a monthly observation
// of an equity index is a month of moves and legitimately dwarfs a daily one:
// the S&P 500's month-end series really did gain 40 % in August 1932.
func verify(s *Series, now time.Time, moveLimit float64) []Issue {
	const (
		maxRateJump  = 3.0 // percentage points in a day, for a rate level
		maxGapDays   = 14  // calendar days without a quote
		maxFlatRun   = 20  // consecutive identical closes
		maxStaleDays = 10  // calendar days since the last quote
	)
	var issues []Issue
	warn := func(d time.Time, format string, args ...any) {
		issues = append(issues, Issue{Severity: "warn", Date: d, Message: fmt.Sprintf(format, args...)})
	}
	if len(s.Points) == 0 {
		return []Issue{{Severity: "error", Message: "no quotes at all"}}
	}
	rate := isRateSymbol(s.Symbol) || isPolicyRate(s.Symbol)
	// A series reports at its own pace, and may change it; every calendar
	// threshold follows the pace in force around each observation.
	cadence := localSpacingDays(s.Points)
	if moveLimit <= 0 {
		moveLimit = maxMove
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
			limit := moveLimit * math.Sqrt(math.Max(1, cadence[k]))
			if r := pt.Close/prev.Close - 1; math.Abs(r) > limit {
				warn(pt.Date, "move of %+.1f %% in one observation, beyond the %.1f %% this series can make",
					r*100, limit*100)
			}
		}
		if gap, limit := pt.Date.Sub(prev.Date).Hours()/24, math.Max(maxGapDays, 3*cadence[k]); gap > limit {
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
	staleLimit := math.Max(maxStaleDays, 3*cadence[len(cadence)-1])
	if age := now.Sub(s.Last().Date).Hours() / 24; age > staleLimit {
		warn(s.Last().Date, "last quote is %.0f days old", age)
	}
	return issues
}

// spacingWindow is how many neighbouring intervals define the pace in force
// around one observation. Twenty-one is a trading month either way: long enough
// for the median to shrug off a holiday week, short enough to follow a series
// that changes pace once, at a date nobody recorded.
const spacingWindow = 21

// localSpacingDays returns, for every index of pts, the typical number of
// calendar days between two consecutive quotes AROUND it: 1 for a daily stretch
// (weekends included, the median absorbs them), 7 for a weekly NAV, ~30 for a
// monthly one. Index 0 carries the same value as index 1, so callers can index
// it by the later point of any pair.
func localSpacingDays(pts []Point) []float64 {
	out := make([]float64, len(pts))
	if len(pts) < 3 {
		for i := range out {
			out[i] = 1
		}
		return out
	}
	gaps := make([]float64, len(pts)) // gaps[k] = days between pts[k-1] and pts[k]
	for k := 1; k < len(pts); k++ {
		gaps[k] = pts[k].Date.Sub(pts[k-1].Date).Hours() / 24
	}
	window := make([]float64, 0, spacingWindow)
	for k := 1; k < len(pts); k++ {
		lo, hi := max(1, k-spacingWindow/2), min(len(pts)-1, k+spacingWindow/2)
		window = append(window[:0], gaps[lo:hi+1]...)
		sortFloats(window)
		out[k] = window[len(window)/2]
	}
	out[0] = out[1]
	return out
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
