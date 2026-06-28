package metrics

import (
	"math"
	"testing"
)

func TestRollingConstantSeries(t *testing.T) {
	// A flat series: every trailing one-year window has mean 7.
	n := 3 * 366
	dates := days(n)
	values := make([]float64, n)
	for i := range values {
		values[i] = 7
	}
	pts, out, ok := Rolling(dates, values, 1, Mean)
	if !ok {
		t.Fatal("expected ok")
	}
	if len(pts) != len(out) || len(out) == 0 {
		t.Fatalf("pts=%d out=%d", len(pts), len(out))
	}
	for i, v := range out {
		near(t, "window mean", v, 7, 1e-12)
		_ = i
	}
	// The first emitted point is at least one year after the series start.
	if pts[0].Sub(dates[0]).Hours()/24 < 365 {
		t.Errorf("first window ends too early: %v", pts[0])
	}
}

func TestRollingTooShort(t *testing.T) {
	if _, _, ok := Rolling(days(100), make([]float64, 100), 5, Mean); ok {
		t.Error("expected ok=false when the series is shorter than the window")
	}
}

func TestRollingVolConstantGrowth(t *testing.T) {
	// Constant daily growth => zero daily-return variance => zero rolling vol.
	n := 3 * 366
	dates := days(n)
	values := make([]float64, n)
	values[0] = 100
	for i := 1; i < n; i++ {
		values[i] = values[i-1] * 1.0003
	}
	_, out, ok := RollingVol(dates, values, 1)
	if !ok {
		t.Fatal("expected ok")
	}
	for _, v := range out {
		if math.Abs(v) > 1e-9 {
			t.Fatalf("rolling vol = %v, want ~0", v)
		}
	}
}

func TestRollingSharpePositiveTrend(t *testing.T) {
	// A noisy upward drift should yield a positive rolling Sharpe.
	n := 3 * 366
	dates := days(n)
	values := make([]float64, n)
	values[0] = 100
	for i := 1; i < n; i++ {
		values[i] = values[i-1] * (1 + 0.0005 + 0.002*math.Sin(float64(i)))
	}
	_, out, ok := RollingSharpe(dates, values, 1)
	if !ok || len(out) == 0 {
		t.Fatalf("ok=%v len=%d", ok, len(out))
	}
	if out[len(out)-1] <= 0 {
		t.Errorf("rolling Sharpe = %v, want positive", out[len(out)-1])
	}
}
