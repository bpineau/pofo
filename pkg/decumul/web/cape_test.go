package web

import (
	"math"
	"testing"
)

func TestCapeSnapshot(t *testing.T) {
	s := capeSnapshot()
	if s.Value <= 5 || s.Value > 60 {
		t.Errorf("CAPE value %.2f out of a sane range", s.Value)
	}
	if s.Percentile < 0 || s.Percentile > 100 {
		t.Errorf("percentile %.1f out of range", s.Percentile)
	}
	if math.Abs(s.ImpliedReal-1/s.Value) > 1e-9 {
		t.Errorf("impliedReal %.4f != 1/CAPE %.4f", s.ImpliedReal, 1/s.Value)
	}
	if s.Median <= 5 || s.Median > s.Value+1e-9 && s.Percentile < 50 {
		t.Errorf("median %.2f inconsistent with percentile %.1f", s.Median, s.Percentile)
	}
	if s.AsOf == "" {
		t.Error("missing AsOf date")
	}
}

func TestCapeAdjustedMu(t *testing.T) {
	sigma := 0.11
	got := capeAdjustedMu(sigma)
	want := capeSnapshot().ImpliedReal + sigma*sigma/2
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("capeAdjustedMu = %.4f, want %.4f", got, want)
	}
}

func TestCapeAdjustLowersCentralMean(t *testing.T) {
	// At a rich valuation the CAPE-implied mean is below a 5% slider mean, so
	// enabling the adjustment must not raise the central mean.
	pr := Params{Mu: 0.05, Sigma: 0.11, Df: 5, Years: 40, CapeAdjust: true}
	mu, _, _ := centralParams(pr, nil)
	if mu >= 0.05 {
		t.Errorf("cape-adjusted central mean %.4f should be below the 5%% slider at a rich CAPE", mu)
	}
}
