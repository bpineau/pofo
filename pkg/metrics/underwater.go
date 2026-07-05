package metrics

import "math"

// Ulcer returns the Ulcer Index of a daily return series, in PERCENT POINTS
// (e.g. 12.8), the root-mean-square of the running drawdown. Because it squares
// the drawdown at every step, it grows with both the DEPTH and the DURATION of
// underwater periods, so it is the smooth, optimizable measure of "how painful
// were the drawdowns to sit through". It matches the Stats.Ulcer that Compute
// derives from a value series (to a negligible boundary term). NaN for an empty
// series.
func Ulcer(returns []float64) float64 {
	if len(returns) == 0 {
		return math.NaN()
	}
	v, peak, sumSq := 1.0, 1.0, 0.0
	for _, r := range returns {
		v *= 1 + r
		if v > peak {
			peak = v
		}
		dd := v/peak - 1
		sumSq += dd * dd
	}
	return math.Sqrt(sumSq/float64(len(returns))) * 100
}

// WorstRollingReturn returns the lowest annualized compound return over any
// window of `window` consecutive daily returns: a robust worst-case measure of
// medium-term outcomes. A window of about 5*252 approximates the worst rolling
// 5-year CAGR, the "how bad could a five-year stretch get" figure that matters
// when the drawdowns must be lived through. ok is false when the series is
// shorter than the window, or a window wiped the capital out.
func WorstRollingReturn(returns []float64, window int) (float64, bool) {
	if window < 1 || len(returns) < window {
		return 0, false
	}
	// Prefix sums of log(1+r) give each window's compound return in O(1).
	prefix := make([]float64, len(returns)+1)
	for i, r := range returns {
		if 1+r <= 0 {
			return 0, false // a wipeout inside the series
		}
		prefix[i+1] = prefix[i] + math.Log(1+r)
	}
	years := float64(window) / tradingDaysPerYear
	worst := math.Inf(1)
	for i := 0; i+window <= len(returns); i++ {
		total := math.Exp(prefix[i+window]-prefix[i]) - 1
		if cagr := math.Pow(1+total, 1/years) - 1; cagr < worst {
			worst = cagr
		}
	}
	return worst, true
}
