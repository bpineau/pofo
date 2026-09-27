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
