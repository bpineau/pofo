package metrics

import (
	"math"
	"testing"
)

// TrackingError is Volatility over the active returns: here they alternate
// +1 % and -1 %, sample deviation sqrt(4 * 0.0001 / 3) per period.
func TestTrackingError(t *testing.T) {
	a := []float64{0.02, 0.00, 0.03, 0.01}
	b := []float64{0.01, 0.01, 0.02, 0.02}
	want := math.Sqrt(4*0.0001/3) * math.Sqrt(12)
	if got := TrackingError(a, b, 12); math.Abs(got-want) > 1e-15 {
		t.Errorf("TrackingError = %v, want %v", got, want)
	}
	if got := TrackingError(a, a, 252); got != 0 {
		t.Errorf("identity = %v", got)
	}
	for name, got := range map[string]float64{
		"mismatch": TrackingError(a, b[:3], 12),
		"one":      TrackingError(a[:1], b[:1], 12),
	} {
		if !math.IsNaN(got) {
			t.Errorf("%s = %v, want NaN", name, got)
		}
	}
}

func TestLeadLagGaps(t *testing.T) {
	nan := math.NaN()
	for _, tc := range []struct {
		name string
		a, b []float64
		want []float64
	}{
		{"in step", []float64{0.01, -0.02, 0.03}, []float64{0.01, -0.02, 0.03}, []float64{0, 0, 0}},
		// b one session late: each a[t] meets itself in b[t+1]; the last one
		// has no b[t+1] and keeps its distance to b[t-1], the nearer reading.
		{"b lags", []float64{0.01, -0.02, 0.03}, []float64{0, 0.01, -0.02}, []float64{0, 0, 0.02}},
		// A stale print (a repeated close) then the catch-up carrying two of
		// b's sessions: forgiven, since a[0] is exactly 0; the stale session
		// itself still reads its distance to b.
		{"catch-up", []float64{0, 0.05}, []float64{0.02, 0.03}, []float64{0.02, 0}},
		// Without the repeated close, two sessions are never pooled.
		{"no pooling", []float64{0.001, 0.05}, []float64{0.02, 0.03}, []float64{0.019, 0.02}},
		// A NaN in b convicts nothing and serves no neighbour.
		{"hole", []float64{0.01, 0.02, 0.04}, []float64{nan, 0.01, 0.01}, []float64{0, 0.01, 0.03}},
		{"mismatch", []float64{0.01}, nil, nil},
	} {
		got := LeadLagGaps(tc.a, tc.b)
		if len(got) != len(tc.want) {
			t.Fatalf("%s: %v, want %v", tc.name, got, tc.want)
		}
		for i := range got {
			if math.Abs(got[i]-tc.want[i]) > 1e-15 {
				t.Errorf("%s: gap %d = %v, want %v", tc.name, i, got[i], tc.want[i])
			}
		}
	}
}

// Track is the package's single-purpose figures gathered, never a second
// formula: each field must equal the function it names.
func TestTrack(t *testing.T) {
	a := []float64{0.021, -0.012, 0.034, 0.008, -0.019, 0.015, 0.002}
	b := []float64{0.018, -0.010, 0.030, 0.011, -0.020, 0.012, 0.004}
	tr, err := Track(a, b, 12)
	if err != nil {
		t.Fatal(err)
	}
	growth := func(r []float64) float64 {
		g := 1.0
		for _, x := range r {
			g *= 1 + x
		}
		return math.Pow(g, 12.0/float64(len(r))) - 1
	}
	beta := slope(b, a)
	for name, c := range map[string][2]float64{
		"Corr":          {tr.Corr, Corr(a, b)},
		"VolA":          {tr.VolA, Volatility(a, 12)},
		"VolB":          {tr.VolB, Volatility(b, 12)},
		"VolRatio":      {tr.VolRatio, Volatility(a, 12) / Volatility(b, 12)},
		"TrackingError": {tr.TrackingError, TrackingError(a, b, 12)},
		"Difference":    {tr.Difference, growth(a) - growth(b)},
		"Beta":          {tr.Beta, beta},
		"Alpha":         {tr.Alpha, (Mean(a) - beta*Mean(b)) * 12},
		"DifferenceSE":  {tr.DifferenceSE(), TrackingError(a, b, 12) / math.Sqrt(7.0/12)},
	} {
		if math.Abs(c[0]-c[1]) > 1e-15 {
			t.Errorf("%s = %v, want %v", name, c[0], c[1])
		}
	}
	if tr.Periods != 7 || tr.PeriodsPerYear != 12 {
		t.Errorf("calendar = %d periods at %v, want 7 at 12", tr.Periods, tr.PeriodsPerYear)
	}
	// The slope agrees with the least-squares fit Regress computes.
	if reg, err := Regress(a, b); err != nil || math.Abs(reg.Betas[0].Value-tr.Beta) > 1e-12 {
		t.Errorf("Beta %v, Regress %v (%v)", tr.Beta, reg.Betas, err)
	}

	// A constant reference: co-movement figures only, no NaN.
	flat, err := Track(a, []float64{0.001, 0.001, 0.001, 0.001, 0.001, 0.001, 0.001}, 12)
	if err != nil {
		t.Fatal(err)
	}
	if flat.Beta != 0 || flat.Alpha != 0 || flat.VolRatio != 0 || flat.Corr != 0 || !(flat.TrackingError > 0) {
		t.Errorf("constant b = %+v, want zero beta, alpha, ratio and correlation, a tracking error", flat)
	}
	if (Tracking{}).DifferenceSE() != 0 {
		t.Error("an empty Tracking has a standard error")
	}

	for name, c := range map[string]struct {
		a, b []float64
		ppy  float64
	}{
		"mismatch":  {a, b[:3], 12},
		"one":       {a[:1], b[:1], 12},
		"cadence":   {a, b, 0},
		"NaN":       {[]float64{0.01, math.NaN()}, []float64{0.01, 0.02}, 12},
		"wiped out": {[]float64{0.01, -1}, []float64{0.01, 0.02}, 12},
	} {
		if _, err := Track(c.a, c.b, c.ppy); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}
