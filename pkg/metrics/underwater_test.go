package metrics

import (
	"math"
	"testing"
)

func TestUlcer(t *testing.T) {
	up := make([]float64, 100)
	for i := range up {
		up[i] = 0.001 // monotonically rising: never underwater
	}
	if u := Ulcer(up); u > 1e-9 {
		t.Fatalf("Ulcer of a rising series = %g, want 0", u)
	}
	dip := append(append([]float64{}, up...), -0.05, -0.05, 0.01)
	if Ulcer(dip) <= 0 {
		t.Fatalf("Ulcer with a drawdown should be positive")
	}
}

func TestWorstRollingReturn(t *testing.T) {
	if _, ok := WorstRollingReturn([]float64{0.01, 0.01}, 10, TradingDaysPerYear); ok {
		t.Fatalf("window longer than the series should be not ok")
	}
	// Steady +0.03%/day ~ +7.8%/yr; the worst window equals that rate.
	steady := make([]float64, 600)
	for i := range steady {
		steady[i] = 0.0003
	}
	w, ok := WorstRollingReturn(steady, 252, TradingDaysPerYear)
	if !ok {
		t.Fatalf("not ok")
	}
	want := math.Pow(1.0003, 252) - 1
	if math.Abs(w-want) > 1e-6 {
		t.Fatalf("worst = %.4f, want %.4f", w, want)
	}
	// A crash in the middle drags the worst window well below the steady rate.
	crashed := append([]float64{}, steady...)
	for i := 300; i < 320; i++ {
		crashed[i] = -0.03
	}
	wc, _ := WorstRollingReturn(crashed, 252, TradingDaysPerYear)
	if wc >= w {
		t.Fatalf("worst with a crash %.4f should be below the steady %.4f", wc, w)
	}
	// The same 252 returns read as twelve monthly steps a year span 21 years:
	// the window's growth is spread over those years, not over one.
	wm, _ := WorstRollingReturn(steady, 252, 12)
	if want := math.Pow(1.0003, 12) - 1; math.Abs(wm-want) > 1e-9 {
		t.Fatalf("worst at 12 a year = %.6f, want %.6f", wm, want)
	}
}
