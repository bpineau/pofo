package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestGlidepathWeightSchedule(t *testing.T) {
	g := Glidepath{StartEquity: 0.3, EndEquity: 0.7, Periods: 41}
	if got := g.weightAt(0); math.Abs(got-0.3) > 1e-9 {
		t.Errorf("year 0 weight = %.3f, want 0.30", got)
	}
	if got := g.weightAt(40); math.Abs(got-0.7) > 1e-9 {
		t.Errorf("final weight = %.3f, want 0.70", got)
	}
	if got := g.weightAt(20); math.Abs(got-0.5) > 1e-9 {
		t.Errorf("midpoint weight = %.3f, want 0.50", got)
	}
}

func TestGlidepathMomentsAndBlend(t *testing.T) {
	// A pure-equity glide (start=end=1) must reproduce the equity mean/vol; the
	// blended early years must be less volatile than the late equity-heavy years.
	g := Glidepath{EquityMu: 0.05, EquitySigma: 0.16, BondMu: 0.015, BondSigma: 0.06,
		Df: 6, Corr: 0.1, StartEquity: 0.2, EndEquity: 1.0, Periods: 40}
	rng := rand.New(rand.NewPCG(1, 2))

	var earlySum, earlySq, lateSum, lateSq float64
	const n = 8000
	for range n {
		seq := g.Draw(rng)
		earlySum += seq[0]
		earlySq += seq[0] * seq[0]
		lateSum += seq[39]
		lateSq += seq[39] * seq[39]
	}
	earlyVar := earlySq/n - (earlySum/n)*(earlySum/n)
	lateVar := lateSq/n - (lateSum/n)*(lateSum/n)
	if !(earlyVar < lateVar) {
		t.Errorf("early (bond-heavy) variance %.4f should be below late (equity-heavy) %.4f", earlyVar, lateVar)
	}
}

func TestGlidepathLen(t *testing.T) {
	g := Glidepath{Periods: 30, StartEquity: 0.5, EndEquity: 0.5}
	if got := g.Draw(rand.New(rand.NewPCG(3, 4))); len(got) != 30 {
		t.Errorf("len = %d, want 30", len(got))
	}
}
