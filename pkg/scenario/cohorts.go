package scenario

import "math/rand/v2"

// HistoricalCohorts yields every actual historical window of Periods
// consecutive returns from the Panel, with no resampling: the deterministic
// "every retirement start date" backtest. Count is the number of windows;
// Draw picks one at random so it also satisfies Source.
type HistoricalCohorts struct {
	Panel   Panel
	Weights []float64
	Periods int

	hist Sequence // Panel combined at Weights, filled by prepare
}

// history is the weighted history to window: the one Prepare computed, or a
// fresh combination.
func (h HistoricalCohorts) history() Sequence {
	if h.hist != nil {
		return h.hist
	}
	return h.Panel.Combine(h.Weights)
}

// prepare implements preparer: it combines the panel once.
func (h HistoricalCohorts) prepare() Source {
	h.hist = h.history()
	return h
}

// Len reports the path length.
func (h HistoricalCohorts) Len() int { return h.Periods }

// Count is the number of distinct start windows, 0 when history is shorter
// than Periods.
func (h HistoricalCohorts) Count() int {
	if n := h.Panel.Periods() - h.Periods + 1; n > 0 {
		return n
	}
	return 0
}

// Cohort returns the i-th historical window (start index i).
func (h HistoricalCohorts) Cohort(i int) Sequence {
	hist := h.history()
	return Sequence(append([]float64(nil), hist[i:i+h.Periods]...))
}

// Draw returns a uniformly random cohort.
func (h HistoricalCohorts) Draw(rng *rand.Rand) Sequence {
	if h.Count() == 0 {
		return make(Sequence, h.Periods)
	}
	return h.Cohort(rng.IntN(h.Count()))
}
