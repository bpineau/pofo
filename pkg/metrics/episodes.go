package metrics

import (
	"math"
	"time"
)

// Episode is one peak-to-trough-to-recovery drawdown of a value series.
// Depth is the deepest fractional loss (-0.25 = -25%). DrawdownDays counts
// the calendar days from peak to trough, RecoveryDays from trough back to the
// prior peak. Ongoing marks an episode that had not recovered by the series
// end (RecoverDate is then zero and RecoveryDays is 0).
type Episode struct {
	PeakDate     time.Time
	TroughDate   time.Time
	RecoverDate  time.Time
	Depth        float64
	DrawdownDays int
	RecoveryDays int
	Ongoing      bool
}

// MaxDrawdown returns the deepest drawdown episode of a value series, or
// the zero Episode when the series never declines. Equal depths keep the
// earlier episode.
func MaxDrawdown(dates []time.Time, values []float64) Episode {
	var worst Episode
	for _, ep := range DrawdownEpisodes(dates, values) {
		if ep.Depth < worst.Depth {
			worst = ep
		}
	}
	return worst
}

// DrawdownEpisodes returns every drawdown of a value series, in chronological
// order: an episode opens when the series falls below its running peak and
// closes when it regains that peak. The last episode is marked Ongoing when
// the series ends underwater. dates must be ascending and the same length as
// values. Returns nil when the series never draws down.
//
// Compute reports only the single longest underwater stretch; this exposes
// the full list, for drawdown-depth and recovery-time distributions.
func DrawdownEpisodes(dates []time.Time, values []float64) []Episode {
	if len(dates) != len(values) || len(values) < 2 {
		return nil
	}
	var eps []Episode
	peak, peakDate := values[0], dates[0]
	inEpisode := false
	var trough float64
	var troughDate time.Time
	calDays := func(a, b time.Time) int {
		return int(math.Round(b.Sub(a).Hours() / 24))
	}
	for i, v := range values {
		switch {
		case v >= peak:
			if inEpisode {
				eps = append(eps, Episode{
					PeakDate:     peakDate,
					TroughDate:   troughDate,
					RecoverDate:  dates[i],
					Depth:        trough/peak - 1,
					DrawdownDays: calDays(peakDate, troughDate),
					RecoveryDays: calDays(troughDate, dates[i]),
				})
				inEpisode = false
			}
			peak, peakDate = v, dates[i]
		case !inEpisode:
			inEpisode = true
			trough, troughDate = v, dates[i]
		case v < trough:
			trough, troughDate = v, dates[i]
		}
	}
	if inEpisode {
		eps = append(eps, Episode{
			PeakDate:     peakDate,
			TroughDate:   troughDate,
			Depth:        trough/peak - 1,
			DrawdownDays: calDays(peakDate, troughDate),
			Ongoing:      true,
		})
	}
	return eps
}
