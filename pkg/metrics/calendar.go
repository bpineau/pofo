package metrics

import (
	"fmt"
	"time"
)

// PeriodReturn is one calendar period of a value series, as CalendarReturns
// cuts it.
type PeriodReturn struct {
	// Start and End are the first and the last date of the series inside the
	// period: the trading days it actually quoted, never a calendar bound
	// nobody quoted.
	Start, End time.Time
	// Return is the period's simple return as a FRACTION (0.05 = +5 %):
	// level(End) / level(previous period's End) - 1, so consecutive periods
	// chain without a gap. The first period has no previous End and is
	// measured from the series' first point instead.
	Return float64
	// Partial marks the first period, measured from the series' first point
	// rather than from a previous period's close. The LAST period is not
	// flagged even when the series stops mid-period: read its End.
	Partial bool
}

// CalendarReturns cuts a value series into calendar periods of the given
// length in months (1 monthly, 3 quarterly, 12 yearly; marketdata.Frequency
// converts with a plain cast) and returns each period's return, in order.
// Periods are blocks of months counted from January, the rule
// marketdata.Series.Resample follows, so 3 cuts on calendar quarters and 12
// on calendar years. It is the table behind an "annual returns" row and a
// monthly heatmap: the product of every (1 + Return) is the series' total
// return over its whole span.
//
// dates must be ascending and parallel to values (levels, any positive unit:
// only their ratios are read). A period with no point in it is absent from
// the table, and the next period's return spans the gap. The result is nil
// for empty or mismatched slices. months must be positive; CalendarReturns
// panics otherwise, as for any programming error.
func CalendarReturns(dates []time.Time, values []float64, months int) []PeriodReturn {
	if months < 1 {
		panic(fmt.Sprintf("metrics: CalendarReturns with a non-positive period (%d months)", months))
	}
	if len(dates) != len(values) || len(dates) == 0 {
		return nil
	}
	period := func(d time.Time) int { return (d.Year()*12 + int(d.Month()) - 1) / months }
	var out []PeriodReturn
	base, first := values[0], 0
	for i := range dates {
		if i+1 < len(dates) && period(dates[i+1]) == period(dates[i]) {
			continue
		}
		out = append(out, PeriodReturn{
			Start:   dates[first],
			End:     dates[i],
			Return:  values[i]/base - 1,
			Partial: len(out) == 0,
		})
		base, first = values[i], i+1
	}
	return out
}
