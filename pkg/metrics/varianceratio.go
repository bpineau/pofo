package metrics

import (
	"math"
	"time"
)

const monthsPerYear = 12

// VolTermStructure compares a value series' volatility measured at daily and
// monthly sampling, the ingredients of the Lo-MacKinlay variance ratio.
//
// Annualizing the daily and the monthly variance and taking their ratio reveals
// the autocorrelation that single-frequency statistics hide:
//
//   - Ratio ≈ 1: returns are serially uncorrelated (i.i.d.); the daily-based
//     volatility is a faithful estimate of multi-period risk.
//   - Ratio < 1: returns mean-revert; daily noise that never compounds
//     overstates the dispersion realized over months.
//   - Ratio > 1: returns trend (positive autocorrelation, e.g. managed-futures
//     sleeves); the daily-based volatility understates the realized risk.
//
// MonthlyN is the number of monthly returns behind MonthlyVol; with the usual
// multi-year report periods it is small (≈ 12 per year), so MonthlyVol and the
// ratio are noisier point estimates than the daily figures and should be read
// with that caveat in mind.
type VolTermStructure struct {
	DailyVol   float64 // annualized stdev of daily returns (stdev·√252)
	MonthlyVol float64 // annualized stdev of monthly returns (stdev·√12)
	Ratio      float64 // monthly annualized variance / daily annualized variance
	MonthlyN   int     // number of monthly returns behind MonthlyVol
}

// VarianceRatio resamples values to calendar month-end closes and returns the
// volatility term structure of the series: the annualized volatility at daily
// and monthly sampling and their variance ratio (Lo-MacKinlay).
//
// dates must be ascending and the same length as values, with at least two
// monthly returns available (the series must span three distinct calendar
// months); ok is false otherwise, or when the daily variance is zero.
func VarianceRatio(dates []time.Time, values []float64) (vt VolTermStructure, ok bool) {
	if len(dates) != len(values) || len(values) < 2 {
		return VolTermStructure{}, false
	}
	dayStd := sampleStdev(Returns(values))
	if !(dayStd > 0) {
		return VolTermStructure{}, false
	}

	_, monthCloses := monthEndCloses(dates, values)
	if len(monthCloses) < 3 {
		return VolTermStructure{}, false
	}
	monthReturns := Returns(monthCloses)
	monthStd := sampleStdev(monthReturns)

	vt.DailyVol = dayStd * math.Sqrt(tradingDaysPerYear)
	vt.MonthlyVol = monthStd * math.Sqrt(monthsPerYear)
	vt.Ratio = (monthStd * monthStd * monthsPerYear) / (dayStd * dayStd * tradingDaysPerYear)
	vt.MonthlyN = len(monthReturns)
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
