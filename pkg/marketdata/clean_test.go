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

// ramp builds a smooth n-point series starting at start, each step *factor plus
// small noise-free drift, so tests can splice two scales cleanly.
func ramp(start, step float64, n int) []float64 {
	out := make([]float64, n)
	v := start
	for i := range out {
		out[i] = v
		v += step
	}
	return out
}

func concat(a, b []float64) []float64 { return append(append([]float64{}, a...), b...) }

func TestMendScaleBreak(t *testing.T) {
	// Older segment at ~100x scale, then a clean junction to the real ~120 NAV
	// (the IBGS.L shape). After mending, the older segment must sit on the newer
	// scale and the series must be continuous (no >=8x jump left).
	t.Run("single clean break rescaled", func(t *testing.T) {
		old := ramp(12000, 20, 25) // ~12000..12480
		new := ramp(120, 0.2, 25)  // ~120..124.8
		got := closes(mendScaleBreak(pts(concat(old, new)...)))
		// Junction ratio at index 25: 120/12480 ≈ 0.009615; older *= that.
		if got[24] > 200 || got[24] < 100 {
			t.Errorf("older segment not rescaled onto newer: got[24]=%.2f", got[24])
		}
		for i := 1; i < len(got); i++ {
			if r := got[i] / got[i-1]; r >= scaleBreakFactor || r <= 1/scaleBreakFactor {
				t.Errorf("scale break remains at %d: %.2f -> %.2f", i, got[i-1], got[i])
			}
		}
	})
	// A spliced share class with several breaks (CL2.PA) is ambiguous: leave it.
	t.Run("multiple breaks untouched", func(t *testing.T) {
		s := concat(concat(ramp(12000, 20, 25), ramp(120, 1, 25)), ramp(40000, 50, 25))
		in := append([]float64{}, s...)
		got := closes(mendScaleBreak(pts(s...)))
		if !eq(got, in...) {
			t.Errorf("multiple-break series was modified")
		}
	})
	// A too-short older side is a leading placeholder / stray tail, not a scale.
	t.Run("short side untouched", func(t *testing.T) {
		s := concat(ramp(12000, 20, 5), ramp(120, 1, 40))
		in := append([]float64{}, s...)
		got := closes(mendScaleBreak(pts(s...)))
		if !eq(got, in...) {
			t.Errorf("short-older-side series was modified")
		}
	})
	// A moderate 3x move (below the 8x threshold) is not a denomination break.
	t.Run("moderate move untouched", func(t *testing.T) {
		s := concat(ramp(40, 0.1, 25), ramp(120, 0.3, 25))
		in := append([]float64{}, s...)
		got := closes(mendScaleBreak(pts(s...)))
		if !eq(got, in...) {
			t.Errorf("moderate 3x move was modified")
		}
	})
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
