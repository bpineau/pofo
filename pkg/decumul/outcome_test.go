package decumul

import (
	"math"
	"math/rand/v2"
	"testing"
)

// A single total-loss path drives the min worst-decade CAGR to -100%, but the
// robust p5 worst-decade must not be dragged all the way down by it.
func TestOutcomeWorst10yRobust(t *testing.T) {
	paths := make([]PathResult, 20)
	flat := []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100}
	for i := range paths {
		paths[i] = PathResult{Wealth: append([]float64(nil), flat...)}
	}
	// One path loses everything within the decade (ends at 0): a -100% decade.
	paths[0] = PathResult{Wealth: []float64{100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 0}, Ruined: true}
	o := Ensemble{Years: 10, Paths: paths}.Outcome()

	if math.Abs(o.Worst10yCAGR-(-1)) > 1e-9 {
		t.Errorf("Worst10yCAGR = %.4f, want -1 (the realised total loss)", o.Worst10yCAGR)
	}
	// p5 of [-1, 0, 0, …] (n=20) interpolates to -0.05: robust to the one ruin.
	if math.Abs(o.Worst10yP5-(-0.05)) > 1e-9 {
		t.Errorf("Worst10yP5 = %.4f, want -0.05", o.Worst10yP5)
	}
	if !(o.Worst10yP5 > o.Worst10yCAGR) {
		t.Errorf("p5 (%.4f) should sit above the min (%.4f)", o.Worst10yP5, o.Worst10yCAGR)
	}
}

// When every path grew through every decade, the worst decade is still one a
// path lived: a running minimum seeded at zero would report a decade of 0 %
// that no path ever produced, and would contradict the p5 sitting above it.
func TestOutcomeWorst10yWithoutALosingDecade(t *testing.T) {
	grow := func(rate float64) []float64 {
		w := make([]float64, 11)
		w[0] = 100
		for i := 1; i < len(w); i++ {
			w[i] = w[i-1] * (1 + rate)
		}
		return w
	}
	o := Ensemble{Years: 10, Paths: []PathResult{
		{Wealth: grow(0.02)}, {Wealth: grow(0.05)},
	}}.Outcome()
	if math.Abs(o.Worst10yCAGR-0.02) > 1e-9 {
		t.Errorf("Worst10yCAGR = %.4f, want 0.02, the leanest decade any path lived", o.Worst10yCAGR)
	}
	if o.Worst10yP5 < o.Worst10yCAGR {
		t.Errorf("p5 (%.4f) fell below the min (%.4f)", o.Worst10yP5, o.Worst10yCAGR)
	}
}

// Cumulative tax and the effective tax rate are medians across paths of each
// path's total tax and its tax/gross ratio.
func TestOutcomeTaxMetrics(t *testing.T) {
	o := Ensemble{Years: 1, Paths: []PathResult{
		{Wealth: []float64{100, 100}, Withdrawn: 90000, TaxPaid: 10000}, // eff 0.10
		{Wealth: []float64{100, 100}, Withdrawn: 80000, TaxPaid: 20000}, // eff 0.20
	}}.Outcome()
	if math.Abs(o.MedianCumTax-15000) > 1e-6 {
		t.Errorf("MedianCumTax = %.0f, want 15000", o.MedianCumTax)
	}
	if math.Abs(o.EffectiveTaxRate-0.15) > 1e-9 {
		t.Errorf("EffectiveTaxRate = %.4f, want 0.15", o.EffectiveTaxRate)
	}
}

func TestOutcomeBasics(t *testing.T) {
	// two paths: one survives flat at 100, one ruined.
	e := Ensemble{Years: 2, Paths: []PathResult{
		{Wealth: []float64{100, 100, 100}, Ruined: false},
		{Wealth: []float64{100, 50, 0}, Ruined: true},
	}}
	o := e.Outcome()
	if math.Abs(o.RuinProb-0.5) > 1e-9 {
		t.Errorf("RuinProb = %.3f, want 0.5", o.RuinProb)
	}
	if o.TerminalP5 > o.TerminalP50 {
		t.Errorf("p5 (%.1f) should be <= p50 (%.1f)", o.TerminalP5, o.TerminalP50)
	}
	if o.CDaR < 0 || o.CDaR > 1 {
		t.Errorf("CDaR out of range: %.3f", o.CDaR)
	}
}

// TestPathPeakStatsMatchesNaive pins the one-division-per-episode walk to the
// per-point definition it replaced, bit for bit, on random paths that ruin,
// recover, sit at zero, start at zero and carry a NaN.
func TestPathPeakStatsMatchesNaive(t *testing.T) {
	naive := func(w []float64) (under int, maxDD float64) {
		peak := w[0]
		for _, v := range w {
			if v >= peak {
				peak = v
				continue
			}
			under++
			if peak > 0 {
				if d := 1 - v/peak; d > maxDD {
					maxDD = d
				}
			}
		}
		return under, maxDD
	}
	rng := rand.New(rand.NewPCG(3, 4))
	for trial := range 5000 {
		w := make([]float64, 1+rng.IntN(60))
		w[0] = 1000 * rng.Float64()
		for k := 1; k < len(w); k++ {
			switch r := rng.Float64(); {
			case r < 0.05:
				w[k] = 0
			case r < 0.06:
				w[k] = math.NaN()
			default:
				w[k] = max(0, w[k-1]*(1+0.3*rng.NormFloat64()))
			}
		}
		if trial%50 == 0 {
			w[0] = 0
		}
		gu, gd := pathPeakStats(w)
		wu, wd := naive(w)
		if gu != wu || math.Float64bits(gd) != math.Float64bits(wd) {
			t.Fatalf("path %v: got %d, %v, want %d, %v", w, gu, gd, wu, wd)
		}
	}
}
