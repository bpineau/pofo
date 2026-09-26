package marketdata

import (
	"sort"
	"time"
)

// Align merges the trading calendars of several series: it returns the
// sorted union of their dates clamped to [start, end] (zero end = no upper
// bound) and, for each input series, its level forward-filled at every
// returned date. Callers must pick a start at or after every series' first
// quote so that forward-filling is always defined.
//
// Two consequences of that contract bite exactly where a series does not
// belong to the calendar the frame is about, and SampleAt exists for them:
// the union takes the outsider's OWN quoting days too, and a start before its
// first quote forward-fills ZEROS rather than failing.
//
// AlignSeries is the strict sibling to reach for first: same calendar, but an
// error naming the series instead of those zeros.
func Align(list []*Series, start, end time.Time) ([]time.Time, [][]float64) {
	dateSet := map[time.Time]struct{}{}
	for _, s := range list {
		for _, p := range s.Points {
			if p.Date.Before(start) || (!end.IsZero() && p.Date.After(end)) {
				continue
			}
			dateSet[p.Date] = struct{}{}
		}
	}
	dates := make([]time.Time, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	levels := make([][]float64, len(list))
	for i, s := range list {
		levels[i] = make([]float64, len(dates))
		j, last := 0, 0.0
		for k, d := range dates {
			for j < len(s.Points) && !s.Points[j].Date.After(d) {
				last = s.Points[j].Close
				j++
			}
			levels[i][k] = last
		}
	}
	return dates, levels
}

// SampleAt reads s at each of the given (ascending) dates: the close of the
// last point at or before the date, and s's FIRST close for the dates that
// precede its history. before is that frontier, the first date sampled from a
// held-flat first close, and is zero when every date fell inside s's own
// history.
//
// It is Align's companion for a series that must NOT shape the calendar: a
// financing rate, a deflator, any exogenous level joined to a frame whose
// dates the assets alone decide. Align is the wrong tool twice over there.
// It takes the UNION of its inputs' dates, so a rate published for every
// calendar day injects weekends into the frame and hands a per-session
// accrual convention 365 steps a year (see pkg/simgen's financed, which
// rebuilds its overnight rate on the bill calendar for that very reason). And
// it forward-fills zeros before a series' first quote, so a window that opens
// earlier than the rate feed reads as 0 %/yr financing rather than as the
// oldest rate on record.
//
// Holding the first level flat backwards is the same graceful degradation
// Client.ConvertCurrency applies to an FX cross that starts late: it keeps the
// order of magnitude honest where a zero would not, and before says how far
// back that extrapolation reaches so the caller can warn.
//
// A nil or empty series yields nil and a zero frontier.
func SampleAt(s *Series, dates []time.Time) (levels []float64, before time.Time) {
	if s == nil || len(s.Points) == 0 || len(dates) == 0 {
		return nil, time.Time{}
	}
	levels = make([]float64, len(dates))
	j, last := 0, s.Points[0].Close
	for k, d := range dates {
		for j < len(s.Points) && !s.Points[j].Date.After(d) {
			last = s.Points[j].Close
			j++
		}
		if d.Before(s.Points[0].Date) && before.IsZero() {
			before = s.Points[0].Date
		}
		levels[k] = last
	}
	return levels, before
}
