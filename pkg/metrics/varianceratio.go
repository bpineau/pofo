package metrics

import (
	"math"
	"time"
)

const monthsPerYear = 12

// VolTermStructure compares a value series' volatility measured at its own
// sampling (daily for market closes, weekly for a weekly NAV) and at monthly
// sampling, the ingredients of the Lo-MacKinlay variance ratio.
//
// Annualizing the native and the monthly variance and taking their ratio
// reveals the autocorrelation that single-frequency statistics hide:
//
//   - Ratio ≈ 1: returns are serially uncorrelated (i.i.d.); the native
//     volatility is a faithful estimate of multi-period risk.
//   - Ratio < 1: returns mean-revert; daily noise that never compounds
//     overstates the dispersion realized over months.
//   - Ratio > 1: returns trend (positive autocorrelation, e.g. managed-futures
//     sleeves); the native volatility understates the realized risk.
//
// MonthlyN is the number of monthly returns behind MonthlyVol; with the usual
// multi-year report periods it is small (≈ 12 per year), so MonthlyVol and the
// ratio are noisier point estimates than the native figures and should be read
// with that caveat in mind.
type VolTermStructure struct {
	NativeVol  float64 // annualized stdev of the series' own returns (stdev·√PeriodsPerYear: √252 on daily closes)
	MonthlyVol float64 // annualized stdev of monthly returns (stdev·√12)
	Ratio      float64 // monthly annualized variance / native annualized variance
	MonthlyN   int     // number of monthly returns behind MonthlyVol

	// Sharpe and Sortino recomputed from the same monthly returns (risk-free
	// rate 0), i.e. annualized mean monthly return over MonthlyVol (resp. over
	// the annualized downside deviation). They are the risk-adjusted twins of
	// the native Stats.Sharpe/Stats.Sortino: where the variance ratio
	// differs from 1 they diverge from the native figures, so a mean-reverting
	// series scores a higher monthly Sharpe (its realized risk is lower than the
	// daily volatility implies) and a trending one a lower monthly Sharpe. Read
	// with the small-sample caveat above (MonthlyN points).
	MonthlySharpe  float64
	MonthlySortino float64
}

// VarianceRatio resamples values to calendar month-end closes and returns the
// volatility term structure of the series: the annualized volatility at its
// own sampling (annualized at PeriodsPerYear(dates)) and at monthly sampling,
// and their variance ratio (Lo-MacKinlay).
//
// dates must be ascending and the same length as values, sampled more often
// than monthly, with at least two monthly returns available (the series must
// span three distinct calendar months); ok is false otherwise, or when the
// native variance is zero. A series sampled monthly or less has no finer
// sampling to compare its month-end closes with.
func VarianceRatio(dates []time.Time, values []float64) (vt VolTermStructure, ok bool) {
	if len(dates) != len(values) || len(values) < 2 {
		return VolTermStructure{}, false
	}
	ppy := PeriodsPerYear(dates)
	if !(ppy > monthsPerYear) {
		return VolTermStructure{}, false
	}
	nativeStd := sampleStdev(Returns(values))
	if !(nativeStd > 0) {
		return VolTermStructure{}, false
	}

	_, monthCloses := monthEndCloses(dates, values)
	if len(monthCloses) < 3 {
		return VolTermStructure{}, false
	}
	monthReturns := Returns(monthCloses)
	monthStd := sampleStdev(monthReturns)

	vt.NativeVol = nativeStd * math.Sqrt(ppy)
	vt.MonthlyVol = monthStd * math.Sqrt(monthsPerYear)
	vt.Ratio = (monthStd * monthStd * monthsPerYear) / (nativeStd * nativeStd * ppy)
	vt.MonthlyN = len(monthReturns)

	// Monthly-sampled Sharpe/Sortino, same rf=0 convention as Stats: the
	// annualized mean over the annualized (downside) deviation. Downside
	// deviation divides by n (target semideviation), matching Compute.
	monthMean := Mean(monthReturns)
	var downSq float64
	for _, x := range monthReturns {
		if x < 0 {
			downSq += x * x
		}
	}
	if vt.MonthlyVol > 0 {
		vt.MonthlySharpe = monthMean * monthsPerYear / vt.MonthlyVol
	}
	if downDev := math.Sqrt(downSq/float64(len(monthReturns))) * math.Sqrt(monthsPerYear); downDev > 0 {
		vt.MonthlySortino = monthMean * monthsPerYear / downDev
	}
	return vt, true
}

// monthEndCloses resamples a daily series to one close per calendar month, the
// last observation of each month. dates must be ascending.
func monthEndCloses(dates []time.Time, values []float64) ([]time.Time, []float64) {
	var outDates []time.Time
	var outValues []float64
	for i := range dates {
		y, m := dates[i].Year(), dates[i].Month()
		last := i == len(dates)-1 || dates[i+1].Year() != y || dates[i+1].Month() != m
		if last {
			outDates = append(outDates, dates[i])
			outValues = append(outValues, values[i])
		}
	}
	return outDates, outValues
}

// sampleStdev is the sample (n-1) standard deviation; it returns 0 for fewer
// than two observations.
func sampleStdev(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := Mean(xs)
	var s float64
	for _, x := range xs {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(xs)-1))
}
