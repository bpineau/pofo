package metrics

import "math"

// TrackingError is the annualized volatility of a's return in excess of
// b's: the standard deviation of a[t] - b[t] over the periods (the sample
// one, n - 1 denominator, as Volatility), times the square root of
// periodsPerYear. a and b are returns on ONE calendar (two columns of a
// marketdata.Panel), as fractions, and so is the result, per year (0.02 = 2
// %/yr). It is the risk of holding a instead of b: a replica, a share class
// or a reconstruction measured against what it is meant to follow. It is NaN
// when the lengths differ or fewer than two periods are given.
func TrackingError(a, b []float64, periodsPerYear float64) float64 {
	if len(a) != len(b) {
		return math.NaN()
	}
	active := make([]float64, len(a))
	for t := range a {
		active[t] = a[t] - b[t]
	}
	return Volatility(active, periodsPerYear)
}

// LeadLagGaps measures, period by period, how far a's return is from b's
// once a one-session difference in closing times is forgiven: the smallest
// of |a[t] - b[t]|, |a[t] - b[t-1]| and |a[t] - b[t+1]| (b lagging or
// leading a by a session) and, beside a session where a did not move at all
// (a[t-1] or a[t+1] exactly 0, a repeated close), |a[t] - b[t] - b[t±1]|:
// the catch-up after a stale print, which carries two of b's sessions at
// once.
//
// Two series that follow one index but close at different hours (a Xetra
// line and a world index struck after New York) disagree by a session on
// every large move, and a plain |a - b| convicts each of those as a defect.
// The gap left once the clock is allowed for is what separates a bad print
// from a time zone: measured on four Xetra lines of an MSCI World fund
// against the index, the largest honest gap was 3.5 % and the smallest
// defect 14.5 %, where a one-session rule lets a stale print reach 7.4 %.
// Pooling two sessions is allowed only beside a repeated close: allowing it
// beside any session widens what a real defect can hide behind.
//
// a and b are parallel returns on one calendar, and should be LOG returns
// (math.Log1p of a simple return) so that two sessions add up exactly. A
// NaN in b marks a period b does not measure: its gap is 0 (nothing
// convicts a on evidence that does not exist) and it is no reading for its
// neighbours. The result is parallel to a and never negative; it is nil
// when the lengths differ.
func LeadLagGaps(a, b []float64) []float64 {
	if len(a) != len(b) {
		return nil
	}
	n := len(a)
	gaps := make([]float64, n)
	for t := range n {
		if math.IsNaN(b[t]) {
			continue
		}
		g := math.Abs(a[t] - b[t])
		for _, j := range []int{t - 1, t + 1} {
			if j < 0 || j >= n || math.IsNaN(b[j]) {
				continue
			}
			g = math.Min(g, math.Abs(a[t]-b[j]))
			if a[j] == 0 {
				g = math.Min(g, math.Abs(a[t]-b[t]-b[j]))
			}
		}
		gaps[t] = g
	}
	return gaps
}
