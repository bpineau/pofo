package marketdata

import (
	"testing"
	"time"
)

func pts(closes ...float64) []Point {
	out := make([]Point, len(closes))
	base := time.Date(2019, 2, 20, 0, 0, 0, 0, time.UTC)
	for i, c := range closes {
		out[i] = Point{Date: base.AddDate(0, 0, i), Close: c}
	}
	return out
}

func closes(ps []Point) []float64 {
	out := make([]float64, len(ps))
	for i, p := range ps {
		out[i] = p.Close
	}
	return out
}

func eq(a []float64, b ...float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestIsRateSymbol(t *testing.T) {
	// Rate series legitimately visit near-zero levels and must be excluded from
	// the dropout filter (^IRX hit ~0.003% in March 2020, a real value).
	for _, s := range []string{"^IRX", "^FVX", "^TNX", "^TYX"} {
		if !isRateSymbol(s) {
			t.Errorf("isRateSymbol(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"AAPL", "IB01.L", "EURUSD=X", "^GSPC", "^VIX"} {
		if isRateSymbol(s) {
			t.Errorf("isRateSymbol(%q) = true, want false", s)
		}
	}
	// The near-zero pattern the guard protects: ^IRX 0.165 -> 0.003 -> 0.013
	// (March 2020) IS an interior "dropout" by shape, so the filter would strip
	// it; the call site must skip rate series to keep it.
	got := closes(dropDropouts(pts(0.165, 0.003, 0.013)))
	if len(got) != 2 {
		t.Errorf("dropDropouts strips the real near-zero rate point; the caller must guard rate series (got %v)", got)
	}
}

func TestDropDropouts(t *testing.T) {
	tests := []struct {
		name string
		in   []float64
		want []float64
	}{
		{
			// The real IB01 (iShares $ Treasury 0-1yr) inception glitch: Yahoo
			// emits two placeholder closes of 5 before the true ~99 NAV.
			name: "leading placeholder run",
			in:   []float64{5, 5, 99.19, 99.20, 99.22, 99.25},
			want: []float64{99.19, 99.20, 99.22, 99.25},
		},
		{
			// A single interior bad print that immediately recovers.
			name: "interior dropout recovers",
			in:   []float64{100, 101, 1, 102, 103},
			want: []float64{100, 101, 102, 103},
		},
		{
			// A genuine large-but-plausible move (e.g. a leveraged/MF sleeve or
			// a distribution) must be kept: +40% then -30% is not a dropout.
			name: "moderate spike kept",
			in:   []float64{100, 140, 98, 99, 100},
			want: []float64{100, 140, 98, 99, 100},
		},
		{
			// A real permanent decline (fund winding down) must be kept: the low
			// tail never recovers, so it is not a round-trip dropout.
			name: "permanent decline kept",
			in:   []float64{100, 100, 20, 18, 17, 16},
			want: []float64{100, 100, 20, 18, 17, 16},
		},
		{
			// A legitimate high-growth ramp from a low base (split-adjusted early
			// equity) must be kept: it rises gradually, never 4x in one day.
			name: "gradual growth kept",
			in:   []float64{1, 1.3, 1.8, 2.5, 3.4, 5},
			want: []float64{1, 1.3, 1.8, 2.5, 3.4, 5},
		},
		{
			name: "clean series untouched",
			in:   []float64{50, 51, 52, 51, 53},
			want: []float64{50, 51, 52, 51, 53},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := closes(dropDropouts(pts(tc.in...)))
			if !eq(got, tc.want...) {
				t.Errorf("dropDropouts(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
