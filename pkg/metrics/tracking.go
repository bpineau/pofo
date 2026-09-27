package metrics

import (
	"fmt"
	"math"
)

// Tracking is how returns a follow returns b over ONE calendar of periods,
// as Track measures them: the figures a replica, a share class or a
// reconstruction is judged by against what it is meant to follow. Returns
// are fractions and every annualized figure is per year at PeriodsPerYear.
type Tracking struct {
	Periods        int     // returns compared
	PeriodsPerYear float64 // the calendar's cadence, what annualizes every figure below
	Corr           float64 // Pearson correlation (Corr); zero when either side is constant
	VolA, VolB     float64 // annualized sample volatilities (Volatility)
	VolRatio       float64 // VolA / VolB; zero when b does not move
	TrackingError  float64 // annualized sample volatility of a - b (TrackingError)
	// Difference is the tracking difference: a's compound annual return over
	// the periods minus b's, each (product of (1 + r))^(PeriodsPerYear /
	// Periods) - 1. On a calendar with gaps (periods dropped at a junction)
	// it compounds the periods that remain, so a step that is not a move
	// never reaches it.
	Difference float64
	Beta       float64 // least-squares slope of a on b, cov(a, b) / var(b), as Beta; zero when b does not move
	Alpha      float64 // Jensen's alpha, (mean a - Beta mean b) per year, arithmetic, as VsBenchmark
}

// DifferenceSE is the standard error of a yearly return gap between the two
// series, TrackingError over the square root of the years the periods span:
// a Difference (or any CAGR gap over the same window) inside two of it is
// not a measurement. It is zero when no period was compared.
func (t Tracking) DifferenceSE() float64 {
	if t.Periods == 0 || !(t.PeriodsPerYear > 0) {
		return 0
	}
	return t.TrackingError / math.Sqrt(float64(t.Periods)/t.PeriodsPerYear)
}

// Track measures how returns a follow returns b: a and b are parallel
// per-period returns on one calendar (two columns of a marketdata.Panel,
// whose Track method pairs them by name), and periodsPerYear its cadence
// (Panel.PeriodsPerYear, or PeriodsPerYear of the dates). Every figure is
// the one the package's single-purpose function computes (Corr, Volatility,
// TrackingError, Beta's slope, VsBenchmark's alpha), gathered so that no
// caller pairs them by hand. A constant b leaves Beta, Alpha and VolRatio at
// zero rather than an error: its correlation and tracking error still mean
// something.
//
// It is an error when the lengths differ, when fewer than two periods are
// given, when a return is not finite or not above -1, and when
// periodsPerYear is not a positive number.
func Track(a, b []float64, periodsPerYear float64) (Tracking, error) {
	switch {
	case len(a) != len(b):
		return Tracking{}, fmt.Errorf("metrics: Track: %d returns against %d (not one calendar)", len(a), len(b))
	case len(a) < 2:
		return Tracking{}, fmt.Errorf("metrics: Track: %d period(s), at least 2 needed", len(a))
	case !(periodsPerYear > 0) || math.IsInf(periodsPerYear, 1):
		return Tracking{}, fmt.Errorf("metrics: Track: %v periods per year", periodsPerYear)
	}
	growthA, growthB := 1.0, 1.0
	for t := range a {
		for _, r := range []float64{a[t], b[t]} {
			if !(r > -1) || math.IsInf(r, 1) {
				return Tracking{}, fmt.Errorf("metrics: Track: period %d holds the return %v", t, r)
			}
		}
		growthA *= 1 + a[t]
		growthB *= 1 + b[t]
	}
	n := float64(len(a))
	tr := Tracking{
		Periods:        len(a),
		PeriodsPerYear: periodsPerYear,
		Corr:           Corr(a, b),
		VolA:           Volatility(a, periodsPerYear),
		VolB:           Volatility(b, periodsPerYear),
		TrackingError:  TrackingError(a, b, periodsPerYear),
		Difference:     math.Pow(growthA, periodsPerYear/n) - math.Pow(growthB, periodsPerYear/n),
	}
	if tr.VolB > 0 {
		tr.VolRatio = tr.VolA / tr.VolB
		tr.Beta = slope(b, a)
		tr.Alpha = (Mean(a) - tr.Beta*Mean(b)) * periodsPerYear
	}
	return tr, nil
}

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
