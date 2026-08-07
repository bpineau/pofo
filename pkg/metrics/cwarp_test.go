package metrics

import (
	"math"
	"testing"
)

// TestCWARPScorePaperExample anchors the geometric formula against the worked
// example in Artemis's paper (p.5): a replacement Sortino of 1.50 and return-
// to-drawdown of 0.15, improved to 1.75 and 0.25, must score +39.44.
func TestCWARPScorePaperExample(t *testing.T) {
	got, ok := cwarpScore(1.50, 1.75, 0.15, 0.25)
	if !ok {
		t.Fatalf("cwarpScore not ok")
	}
	if math.Abs(got-39.44) > 0.01 {
		t.Fatalf("CWARP = %.2f, want 39.44", got)
	}
}

// replSeries builds a deterministic equity-like replacement: a gentle upward
// drift with an oscillation and an injected crash, so it has both downside
// deviation (Sortino defined) and a real drawdown (RtMDD defined).
func replSeries() []float64 {
	r := make([]float64, 300)
	for i := range r {
		r[i] = 0.0010 + 0.006*math.Sin(float64(i)*0.3)
	}
	for i := 100; i < 115; i++ {
		r[i] = -0.010 // a crash, leaving the series net positive overall
	}
	return r
}

// TestCWARPZeroOverlayNeutral: overlaying an all-zero (cash-at-zero) asset with
// no financing changes nothing, so CWARP is exactly 0.
func TestCWARPZeroOverlayNeutral(t *testing.T) {
	repl := replSeries()
	asset := make([]float64, len(repl))
	got, ok := CWARP(asset, repl, CWARPParams{})
	if !ok {
		t.Fatalf("CWARP not ok")
	}
	if math.Abs(got) > 1e-9 {
		t.Fatalf("neutral overlay CWARP = %g, want 0", got)
	}
}

// TestCWARPDiversifierPositive: an anti-correlated sleeve with a small positive
// drift dampens the replacement's swings and adds return, so CWARP > 0.
func TestCWARPDiversifierPositive(t *testing.T) {
	repl := replSeries()
	asset := make([]float64, len(repl))
	for i := range repl {
		asset[i] = -repl[i] + 0.0007 // hedge plus positive carry
	}
	got, ok := CWARP(asset, repl, CWARPParams{})
	if !ok {
		t.Fatalf("CWARP not ok")
	}
	if got <= 0 {
		t.Fatalf("diversifier CWARP = %.2f, want > 0", got)
	}
}

// TestCWARPFinancedLeverageNegative: overlaying more of the same (perfectly
// correlated) exposure, dragged by a real financing cost, cannot improve
// risk-adjusted returns, so CWARP < 0.
func TestCWARPFinancedLeverageNegative(t *testing.T) {
	repl := replSeries()
	asset := append([]float64(nil), repl...) // same exposure
	got, ok := CWARP(asset, repl, CWARPParams{Financing: 0.10})
	if !ok {
		t.Fatalf("CWARP not ok")
	}
	if got >= 0 {
		t.Fatalf("financed leverage CWARP = %.2f, want < 0", got)
	}
}

// TestCWARPUndefined covers the not-ok paths: mismatched lengths and a
// replacement that only ever rises (no drawdown, so RtMDD is undefined).
func TestCWARPUndefined(t *testing.T) {
	repl := replSeries()
	if _, ok := CWARP(repl[:10], repl, CWARPParams{}); ok {
		t.Fatalf("mismatched lengths should be not ok")
	}
	up := make([]float64, 100)
	for i := range up {
		up[i] = 0.001 // monotonically rising: no drawdown
	}
	if _, ok := CWARP(up, up, CWARPParams{}); ok {
		t.Fatalf("no-drawdown replacement should be not ok")
	}
}
