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
