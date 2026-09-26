package metrics

import (
	"math"
	"testing"
)

func TestVaRCVaRKnownSample(t *testing.T) {
	// Ten returns from -5 % to +5 %.
	rs := []float64{0.03, -0.05, 0.01, -0.02, 0.05, 0, -0.03, 0.02, -0.01, 0.04}
	for _, c := range []struct {
		p, varWant, cvarWant float64
	}{
		// 10th percentile: -0.05 + 0.9 * (-0.03 - -0.05) = -0.032; only -5 % is beyond it.
		{0.90, 0.032, 0.05},
		// 20th percentile: -0.03 + 0.8 * 0.01 = -0.022; -5 % and -3 % are beyond it.
		{0.80, 0.022, 0.04},
	} {
		v, ok := VaR(rs, c.p)
		if !ok {
			t.Fatalf("VaR(%v) not ok", c.p)
		}
		near(t, "VaR", v, c.varWant, 1e-12)
		near(t, "VaR vs Quantiles", v, -Quantiles(rs, 1-c.p)[0], 0)
		cv, ok := CVaR(rs, c.p)
		if !ok {
			t.Fatalf("CVaR(%v) not ok", c.p)
		}
		near(t, "CVaR", cv, c.cvarWant, 1e-12)
	}
}

func TestCVaRAtLeastVaR(t *testing.T) {
	rs := make([]float64, 500)
	for i := range rs {
		rs[i] = 0.01*math.Sin(float64(i)*1.3) - 0.004*math.Cos(float64(i)*0.2)
	}
	for _, p := range []float64{0.5, 0.9, 0.95, 0.99, 0.999} {
		v, _ := VaR(rs, p)
		cv, _ := CVaR(rs, p)
		if cv < v {
			t.Errorf("p=%v: CVaR %v below VaR %v", p, cv, v)
		}
		if v <= 0 {
			t.Errorf("p=%v: VaR %v is not a loss", p, v)
		}
	}
}

func TestVaRAllGains(t *testing.T) {
	v, ok := VaR([]float64{0.01, 0.02, 0.03}, 0.95)
	if !ok || v >= 0 {
		t.Errorf("all gains: VaR %v ok=%v, want a negative figure", v, ok)
	}
}

func TestVaRCVaREdgeCases(t *testing.T) {
	for _, c := range []struct {
		name string
		rs   []float64
		p    float64
	}{
		{"empty", nil, 0.95},
		{"p=0", []float64{0.01, -0.02}, 0},
		{"p=1", []float64{0.01, -0.02}, 1},
		{"p<0", []float64{0.01, -0.02}, -0.5},
		{"p NaN", []float64{0.01, -0.02}, math.NaN()},
	} {
		if _, ok := VaR(c.rs, c.p); ok {
			t.Errorf("VaR %s: ok", c.name)
		}
		if _, ok := CVaR(c.rs, c.p); ok {
			t.Errorf("CVaR %s: ok", c.name)
		}
	}
	// A NaN in the sample poisons the tail rather than reading a number.
	if cv, ok := CVaR([]float64{math.NaN(), -0.01, 0.02}, 0.9); ok || !math.IsNaN(cv) {
		t.Errorf("NaN sample: CVaR %v ok=%v", cv, ok)
	}
	// One point: both read that point's loss.
	v, ok1 := VaR([]float64{-0.03}, 0.95)
	cv, ok2 := CVaR([]float64{-0.03}, 0.95)
	if !ok1 || !ok2 || v != 0.03 || cv != 0.03 {
		t.Errorf("one point: VaR %v CVaR %v", v, cv)
	}
}
