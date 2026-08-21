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
	// Older segment quoted in pence, then a clean junction to the same NAV in
	// pounds (the ITPS.L shape: exactly 100x, bar a day of real move). After
	// mending, the older segment must sit on the newer scale and the series must
	// be continuous (no >=8x jump left).
	t.Run("single clean break rescaled", func(t *testing.T) {
		old := ramp(12000, 20, 25) // pence, ~12000..12480
		new := ramp(124.9, 0.2, 25)
		got := closes(mendScaleBreak(pts(concat(old, new)...)))
		// Junction ratio at index 25: 124.9/12480 ≈ 0.010008; older *= that.
		if got[24] > 200 || got[24] < 100 {
			t.Errorf("older segment not rescaled onto newer: got[24]=%.2f", got[24])
		}
		for i := 1; i < len(got); i++ {
			if r := got[i] / got[i-1]; r >= scaleBreakFactor || r <= 1/scaleBreakFactor {
				t.Errorf("scale break remains at %d: %.2f -> %.2f", i, got[i-1], got[i])
			}
		}
	})
	// The real IBGS.L trap: the pre-2009 segment is the fund's EUR NAV in cents,
	// the later one its GBP line, so the junction is 104x, not 100x. That is two
	// quote lines, not a change of units: welding them buries a currency change
	// inside one series, which FX conversion then reprices as if uniform (a
	// fictitious -21 % across 2008). It must be left for the doctor to flag.
	t.Run("cross-currency weld untouched", func(t *testing.T) {
		old := ramp(13175, 12, 25) // EUR cents, ~13175..13463
		new := ramp(129.5, 0.1, 25)
		s := concat(old, new)
		in := append([]float64{}, s...)
		got := closes(mendScaleBreak(pts(s...)))
		if !eq(got, in...) {
			t.Errorf("cross-currency weld was mended: got[24]=%.4f, want %.4f", got[24], in[24])
		}
	})
	// A spliced share class with several breaks is ambiguous: leave it.
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
	// A French-franc -> euro redenomination (6.55957, below 8x) must still be
	// mended: a fund NAV spliced across the 1999 changeover (the LU0131510165
	// case). The old francs sit at 6.55957x the euro NAV.
	t.Run("euro-legacy redenomination mended", func(t *testing.T) {
		old := ramp(200, 0.75, 25)         // francs, ending at 218
		new := ramp(218/6.55957, 0.05, 25) // euros, starting at 218/6.55957
		got := closes(mendScaleBreak(pts(concat(old, new)...)))
		if got[24] > 40 || got[24] < 28 { // 218 franc point rescaled to ~33 euro
			t.Errorf("franc segment not rescaled onto euro NAV: got[24]=%.2f", got[24])
		}
		for i := 1; i < len(got); i++ {
			if r := got[i] / got[i-1]; r >= 6 || r <= 1.0/6 {
				t.Errorf("redenomination break remains at %d: %.2f -> %.2f", i, got[i-1], got[i])
			}
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
			// The mirror image, and the real CL2.PA holiday print: the session
			// the current quote line skips is filled from the pre-split one,
			// ~300x above both neighbours.
			name: "interior spike recovers",
			in:   []float64{1.7312, 522.81, 1.744, 1.7382, 1.7159},
			want: []float64{1.7312, 1.744, 1.7382, 1.7159},
		},
		{
			// A permanent step is a split or a redenomination, not a bad print:
			// the level moves and STAYS there, so no point is far from both its
			// neighbours and mendScaleBreak keeps its junction to reason about.
			name: "permanent step kept",
			in:   []float64{300, 303, 306, 1, 1.01, 1.02},
			want: []float64{300, 303, 306, 1, 1.01, 1.02},
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

// A lone spike is also a fourfold jump up, so the leading-placeholder rule can
// mistake it for the end of an inception run and condemn every real quote
// before it. That is why the isolated prints go first, and why a long prefix is
// never placeholder noise: the CL2.PA shape must come back whole.
func TestDropDropoutsKeepsHistoryBeforeASpike(t *testing.T) {
	in := make([]float64, 0, 60)
	for i := 0; i < 30; i++ {
		in = append(in, 1.70+float64(i)*0.001)
	}
	in = append(in, 522.81) // the holiday print, ~300x the level around it
	for i := 0; i < 29; i++ {
		in = append(in, 1.73+float64(i)*0.001)
	}
	got := closes(dropDropouts(pts(in...)))
	if len(got) != len(in)-1 {
		t.Fatalf("dropDropouts kept %d of %d points, want %d: the spike alone must go",
			len(got), len(in), len(in)-1)
	}
	if got[0] != in[0] {
		t.Errorf("the history before the spike was dropped: first close %.4f, want %.4f", got[0], in[0])
	}
	for i, v := range got {
		if v > 100 {
			t.Errorf("the spike survived at index %d (%.2f)", i, v)
		}
	}
}

// A leading segment at or beyond minScaleSegment is real history at another
// scale (a fund quoted in pence, an unadjusted pre-split line): dropDropouts
// must hand it to mendScaleBreak intact rather than delete it.
func TestDropDropoutsKeepsALongLeadingSegment(t *testing.T) {
	in := concat(ramp(124.9, 0.2, 25), ramp(12500, 20, 25)) // pounds, then pence
	got := closes(dropDropouts(pts(in...)))
	if !eq(got, in...) {
		t.Errorf("the 25-point leading segment was treated as placeholder noise (%d points left of %d)",
			len(got), len(in))
	}
}

func TestDropFXSpikes(t *testing.T) {
	tests := []struct {
		name string
		in   []float64
		want []float64
	}{
		{
			// The real EURUSD=X case: Yahoo printed 1.4918 on 2008-12-08
			// between 1.2717 and 1.2926 (+17% then -13%, near-cancelling).
			name: "isolated self-cancelling spike dropped",
			in:   []float64{1.2774, 1.2717, 1.4918, 1.2926, 1.3014},
			want: []float64{1.2774, 1.2717, 1.2926, 1.3014},
		},
		{
			// A genuine devaluation (2015 CHF depeg style): the move does
			// not revert, so it must be kept.
			name: "persistent shock kept",
			in:   []float64{1.20, 1.20, 0.99, 1.02, 1.04},
			want: []float64{1.20, 1.20, 0.99, 1.02, 1.04},
		},
		{
			// A whipsaw below the 8% threshold is ordinary volatility.
			name: "small whipsaw kept",
			in:   []float64{1.30, 1.35, 1.31, 1.32},
			want: []float64{1.30, 1.35, 1.31, 1.32},
		},
		{
			name: "clean series untouched",
			in:   []float64{1.10, 1.11, 1.12, 1.11},
			want: []float64{1.10, 1.11, 1.12, 1.11},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := closes(dropFXSpikes(pts(tc.in...)))
			if !eq(got, tc.want...) {
				t.Errorf("dropFXSpikes(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
