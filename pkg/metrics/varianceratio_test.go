package metrics

import (
	"math"
	"testing"
	"time"
)

// sampleStd is an independent reference for the sample standard deviation,
// kept tiny so the variance-ratio expectations below are checkable by hand.
func sampleStd(xs []float64) float64 {
	m := Mean(xs)
	var s float64
	for _, x := range xs {
		s += (x - m) * (x - m)
	}
	return math.Sqrt(s / float64(len(xs)-1))
}

func TestVarianceRatioMeanRevertingDailyFlatMonthly(t *testing.T) {
	// Daily prices oscillate 100/101 within each month but every month closes
	// at 100, so the month-end series is flat: monthly variance is zero while
	// daily variance is not. The ratio collapses to 0 (extreme mean reversion),
	// and resampling must pick the month-end close (100), not the oscillation.
	var dates []time.Time
	var values []float64
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for m := range 3 { // Jan, Feb, Mar 2020
		monthStart := start.AddDate(0, m, 0)
		for d := 0; d <= 10; d++ { // 11 obs, last (d=10) closes the month at 100
			dates = append(dates, monthStart.AddDate(0, 0, d))
			if d%2 == 0 {
				values = append(values, 100)
			} else {
				values = append(values, 101)
			}
		}
	}

	vr, ok := VarianceRatio(dates, values)
	if !ok {
		t.Fatal("VarianceRatio returned ok=false")
	}
	if vr.MonthlyN != 2 {
		t.Errorf("MonthlyN = %d, want 2", vr.MonthlyN)
	}
	if vr.MonthlyVol != 0 {
		t.Errorf("MonthlyVol = %v, want 0 (flat month-end closes)", vr.MonthlyVol)
	}
	if !(vr.DailyVol > 0) {
		t.Errorf("DailyVol = %v, want > 0", vr.DailyVol)
	}
	if vr.Ratio != 0 {
		t.Errorf("Ratio = %v, want 0", vr.Ratio)
	}
}

func TestVarianceRatioKnownSeries(t *testing.T) {
	// Four months with explicit month-end closes [100, 110, 105, 115]; the
	// in-between daily values give a non-zero daily variance.
	months := []struct {
		mon  int
		days []float64 // values on days 1..N; the last is the month-end close
	}{
		{1, []float64{100, 102, 98, 100}},  // Jan closes at 100
		{2, []float64{101, 104, 108, 110}}, // Feb closes at 110
		{3, []float64{109, 103, 107, 105}}, // Mar closes at 105
		{4, []float64{106, 112, 109, 115}}, // Apr closes at 115
	}
	var dates []time.Time
	var values []float64
	for _, mm := range months {
		for i, v := range mm.days {
			dates = append(dates, time.Date(2021, time.Month(mm.mon), i+1, 0, 0, 0, 0, time.UTC))
			values = append(values, v)
		}
	}

	// Independent expectations.
	monthCloses := []float64{100, 110, 105, 115}
	monthRets := Returns(monthCloses)
	expMonthlyVol := sampleStd(monthRets) * math.Sqrt(12)
	expDailyVol := sampleStd(Returns(values)) * math.Sqrt(tradingDaysPerYear)
	expRatio := (expMonthlyVol * expMonthlyVol) / (expDailyVol * expDailyVol)
	expMonthlySharpe := Mean(monthRets) * 12 / expMonthlyVol
	var downSq float64
	for _, x := range monthRets {
		if x < 0 {
			downSq += x * x
		}
	}
	expMonthlySortino := Mean(monthRets) * 12 / (math.Sqrt(downSq/float64(len(monthRets))) * math.Sqrt(12))

	vr, ok := VarianceRatio(dates, values)
	if !ok {
		t.Fatal("VarianceRatio returned ok=false")
	}
	if vr.MonthlyN != 3 {
		t.Errorf("MonthlyN = %d, want 3", vr.MonthlyN)
	}
	const eps = 1e-12
	if math.Abs(vr.MonthlyVol-expMonthlyVol) > eps {
		t.Errorf("MonthlyVol = %v, want %v", vr.MonthlyVol, expMonthlyVol)
	}
	if math.Abs(vr.DailyVol-expDailyVol) > eps {
		t.Errorf("DailyVol = %v, want %v", vr.DailyVol, expDailyVol)
	}
	if math.Abs(vr.Ratio-expRatio) > eps {
		t.Errorf("Ratio = %v, want %v", vr.Ratio, expRatio)
	}
	if math.Abs(vr.MonthlySharpe-expMonthlySharpe) > eps {
		t.Errorf("MonthlySharpe = %v, want %v", vr.MonthlySharpe, expMonthlySharpe)
	}
	if math.Abs(vr.MonthlySortino-expMonthlySortino) > eps {
		t.Errorf("MonthlySortino = %v, want %v", vr.MonthlySortino, expMonthlySortino)
	}
}

func TestVarianceRatioTooShort(t *testing.T) {
	// Fewer than three month-end closes cannot yield two monthly returns.
	dates := []time.Time{
		time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 1, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 7, 0, 0, 0, 0, time.UTC),
	}
	values := []float64{100, 101, 102}
	if _, ok := VarianceRatio(dates, values); ok {
		t.Error("VarianceRatio ok=true on a series spanning under three months")
	}

	if _, ok := VarianceRatio(nil, nil); ok {
		t.Error("VarianceRatio ok=true on an empty series")
	}
	if _, ok := VarianceRatio(dates[:2], values[:1]); ok {
		t.Error("VarianceRatio ok=true on mismatched lengths")
	}
}
