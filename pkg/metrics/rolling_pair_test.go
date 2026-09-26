package metrics

import (
	"math"
	"testing"
	"time"
)

// leveredPair is a benchmark with wobbling daily returns and a series that
// returns exactly k times as much every day, both on days(n).
func leveredPair(n int, k float64) (dates []time.Time, series, bench []float64) {
	dates = days(n)
	series, bench = make([]float64, n), make([]float64, n)
	series[0], bench[0] = 100, 100
	for i := 1; i < n; i++ {
		r := 0.0003 + 0.008*math.Sin(float64(i)*0.9)
		bench[i] = bench[i-1] * (1 + r)
		series[i] = series[i-1] * (1 + k*r)
	}
	return dates, series, bench
}

func TestRollingBetaCorrLevered(t *testing.T) {
	dates, series, bench := leveredPair(3*366, 2)
	bd, betas, ok := RollingBeta(dates, series, dates, bench, 1)
	if !ok || len(betas) == 0 {
		t.Fatalf("RollingBeta ok=%v len=%d", ok, len(betas))
	}
	cd, corrs, ok := RollingCorr(dates, series, dates, bench, 1)
	if !ok {
		t.Fatal("RollingCorr not ok")
	}
	for i := range betas {
		near(t, "beta", betas[i], 2, 1e-9)
		near(t, "corr", corrs[i], 1, 1e-12)
	}
	// On a shared calendar the windows are RollingVol's, date for date.
	vd, _, _ := RollingVol(dates, series, 1)
	if len(bd) != len(vd) || len(cd) != len(vd) {
		t.Fatalf("%d beta and %d corr windows, RollingVol has %d", len(bd), len(cd), len(vd))
	}
	for i := range vd {
		if !bd[i].Equal(vd[i]) || !cd[i].Equal(vd[i]) {
			t.Fatalf("window %d ends %s / %s, RollingVol's %s", i, bd[i], cd[i], vd[i])
		}
	}
	// The whole-sample Beta agrees with every window here.
	b, _ := Beta(dates, series, dates, bench)
	near(t, "Beta", b, 2, 1e-9)
}

func TestRollingCorrInverse(t *testing.T) {
	dates, series, bench := leveredPair(2*366, -1)
	_, corrs, ok := RollingCorr(dates, series, dates, bench, 1)
	if !ok {
		t.Fatal("not ok")
	}
	for _, c := range corrs {
		near(t, "corr", c, -1, 1e-12)
	}
}

func TestRollingBetaDifferentCalendars(t *testing.T) {
	// The benchmark skips every seventh day; the pairs are the days both
	// quote, each return from its own series' previous point.
	dates, series, bench := leveredPair(3*366, 2)
	var bd []time.Time
	var bv []float64
	for i := range dates {
		if i%7 != 3 {
			bd, bv = append(bd, dates[i]), append(bv, bench[i])
		}
	}
	pts, betas, ok := RollingBeta(dates, series, bd, bv, 1)
	if !ok || len(betas) == 0 {
		t.Fatalf("ok=%v len=%d", ok, len(betas))
	}
	for i, d := range pts {
		if math.IsNaN(betas[i]) {
			t.Fatalf("window ending %s is NaN", d)
		}
	}
	// The last window is the slope over exactly the pairs Beta would use on
	// that span.
	near(t, "last window", betas[len(betas)-1], rollingReference(t, dates, series, bd, bv, pts[len(pts)-1]), 1e-12)
}

// rollingReference recomputes the last window's beta by hand: every pair
// whose return is measured from a point after end-1y.
func rollingReference(t *testing.T, dates []time.Time, values []float64, bd []time.Time, bv []float64, end time.Time) float64 {
	t.Helper()
	lo := end.Add(-time.Duration(365.25 * 24 * float64(time.Hour)))
	p := pairReturns(dates, values, bd, bv)
	var own, bench []float64
	for k := range p.own {
		if p.start[k].After(lo) && !p.end[k].After(end) {
			own, bench = append(own, p.own[k]), append(bench, p.bench[k])
		}
	}
	return slope(bench, own)
}

func TestRollingBetaCorrEdgeCases(t *testing.T) {
	dates, series, bench := leveredPair(100, 2)
	if _, _, ok := RollingBeta(dates, series, dates, bench, 1); ok {
		t.Error("RollingBeta: ok on a series shorter than the window")
	}
	if _, _, ok := RollingCorr(dates, series, dates, bench, 0); ok {
		t.Error("RollingCorr: ok with years=0")
	}
	if _, _, ok := RollingBeta(dates, series[:50], dates, bench, 0.1); ok {
		t.Error("RollingBeta: ok on mismatched slices")
	}
	if _, _, ok := RollingCorr(dates[:1], series[:1], dates, bench, 0.1); ok {
		t.Error("RollingCorr: ok on a one-point series")
	}
	// A window too short to hold a pair (one calendar day on daily data:
	// every pair starts on the window's lower bound) reads NaN, not a number.
	_, betas, ok := RollingBeta(dates, series, dates, bench, 1/365.25)
	if !ok || len(betas) == 0 {
		t.Fatalf("one-day windows: ok=%v len=%d", ok, len(betas))
	}
	_, corrs, _ := RollingCorr(dates, series, dates, bench, 1/365.25)
	for i := range betas {
		if !math.IsNaN(betas[i]) || !math.IsNaN(corrs[i]) {
			t.Fatalf("one-day window %d: beta %v corr %v, want NaN", i, betas[i], corrs[i])
		}
	}
}
