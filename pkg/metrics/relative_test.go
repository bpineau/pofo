package metrics

import (
	"math"
	"testing"
)

func TestDrawdowns(t *testing.T) {
	dd := Drawdowns([]float64{100, 120, 90, 120, 130})
	want := []float64{0, 0, 90.0/120 - 1, 0, 0}
	for i := range dd {
		if math.Abs(dd[i]-want[i]) > 1e-12 {
			t.Errorf("dd[%d] = %v, want %v", i, dd[i], want[i])
		}
	}
}

func TestRollingCAGRConstantGrowth(t *testing.T) {
	// 10 % per 365.25-day year, exactly: every rolling window agrees.
	n := 8 * 253
	dates := days(n) // helper from metrics_test.go (daily)
	values := make([]float64, n)
	for i := range values {
		yearsElapsed := dates[i].Sub(dates[0]).Hours() / 24 / 365.25
		values[i] = 100 * math.Pow(1.10, yearsElapsed)
	}
	worst, median, best, windows, ok := RollingCAGR(dates, values, 5)
	if !ok || windows < 100 {
		t.Fatalf("ok=%v windows=%d", ok, windows)
	}
	for name, v := range map[string]float64{"worst": worst, "median": median, "best": best} {
		if math.Abs(v-0.10) > 1e-6 {
			t.Errorf("%s = %v, want 0.10", name, v)
		}
	}
	// Series shorter than the window: not ok.
	if _, _, _, _, ok := RollingCAGR(dates[:300], values[:300], 5); ok {
		t.Error("expected ok=false for a window longer than the series")
	}
}

func TestVsBenchmarkLeveredClone(t *testing.T) {
	// Portfolio = 2× benchmark daily returns: beta 2, alpha 0, IR sign of
	// the mean active return, captures 2 on both sides (approximately:
	// geometric compounding makes 2× slightly path-dependent).
	n := 300
	dates := days(n)
	bench := make([]float64, n)
	port := make([]float64, n)
	bench[0], port[0] = 100, 100
	for i := 1; i < n; i++ {
		r := 0.01 * math.Sin(float64(i))
		bench[i] = bench[i-1] * (1 + r)
		port[i] = port[i-1] * (1 + 2*r)
	}
	rel, ok := VsBenchmark(dates, port, dates, bench)
	if !ok {
		t.Fatal("expected ok")
	}
	if math.Abs(rel.Beta-2) > 1e-9 {
		t.Errorf("beta = %v, want 2", rel.Beta)
	}
	if math.Abs(rel.Alpha) > 1e-9 {
		t.Errorf("alpha = %v, want 0", rel.Alpha)
	}
	if rel.UpCapture < 1.8 || rel.UpCapture > 2.2 {
		t.Errorf("up capture = %v, want ≈2", rel.UpCapture)
	}
	if rel.DownCapture < 1.8 || rel.DownCapture > 2.2 {
		t.Errorf("down capture = %v, want ≈2", rel.DownCapture)
	}
}
