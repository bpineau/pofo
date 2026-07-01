package scenario

import (
	"math"
	"math/rand/v2"
	"sort"
	"testing"
)

// worstWindow returns the worst k-period compounded return of a path.
func worstWindow(s Sequence, k int) float64 {
	worst := math.Inf(1)
	for i := 0; i+k <= len(s); i++ {
		g := 1.0
		for j := i; j < i+k; j++ {
			g *= 1 + s[j]
		}
		if g-1 < worst {
			worst = g - 1
		}
	}
	return worst
}

// The regime source must produce deeper multi-year drawdowns than an i.i.d.
// source with the same calm parameters: bad years cluster (sequence risk), so
// the 5th-percentile worst 5-year compounded return is markedly more negative.
func TestMarkovRegimeClustersDrawdowns(t *testing.T) {
	const n, periods = 4000, 40
	regime := MarkovRegime{
		CalmMu: 0.06, CalmSigma: 0.12, BearMu: -0.20, BearSigma: 0.20,
		StayCalm: 0.85, StayBear: 0.70, Df: 5, Periods: periods,
	}
	iid := ParametricSource{Mu: 0.06, Sigma: 0.12, Df: 5, Periods: periods}

	p5 := func(s Source) float64 {
		rng := rand.New(rand.NewPCG(1, 2))
		worsts := make([]float64, n)
		for i := range worsts {
			worsts[i] = worstWindow(s.Draw(rng), 5)
		}
		sort.Float64s(worsts)
		return worsts[n/20]
	}
	rp5, ip5 := p5(regime), p5(iid)
	if !(rp5 < ip5-0.05) {
		t.Errorf("regime worst-5y p5 = %.3f, i.i.d. = %.3f; regime should be clearly worse (sequence risk)", rp5, ip5)
	}
}

// TestNewMarkovRegimeMeanPreservation verifies that NewMarkovRegime produces a
// blended long-run mean that matches the requested mu within a tight tolerance.
// It uses a fixed seed for determinism and a large sample (10,000 paths of 40
// periods each) so the empirical mean is stable to well within the 0.005 bound.
func TestNewMarkovRegimeMeanPreservation(t *testing.T) {
	cases := []struct {
		mu, sigma float64
	}{
		{0.04, 0.16},
		{0.03, 0.18},
	}
	const (
		nPaths  = 10_000
		periods = 40
		tol     = 0.005
	)
	for _, tc := range cases {
		rng := rand.New(rand.NewPCG(42, 0))
		src := NewMarkovRegime(tc.mu, tc.sigma, 5, periods)
		var sum float64
		var count int
		for range nPaths {
			for _, r := range src.Draw(rng) {
				sum += r
				count++
			}
		}
		got := sum / float64(count)
		if math.Abs(got-tc.mu) > tol {
			t.Errorf("mu=%.2f sigma=%.2f: empirical mean=%.4f, want within %.3f of %.2f",
				tc.mu, tc.sigma, got, tol, tc.mu)
		}
	}
}

// TestNewMarkovRegimeSequenceRisk verifies that NewMarkovRegime, despite
// preserving the long-run mean, still produces a worse left tail (deeper
// multi-year drawdowns) than an i.i.d. source at the same mu/sigma. Sequence
// risk comes from the clustering of bad years, not from a lower mean.
func TestNewMarkovRegimeSequenceRisk(t *testing.T) {
	const n, periods = 4000, 40
	mu, sigma := 0.04, 0.16
	regime := NewMarkovRegime(mu, sigma, 5, periods)
	iid := ParametricSource{Mu: mu, Sigma: sigma, Df: 5, Periods: periods}

	p5 := func(s Source) float64 {
		rng := rand.New(rand.NewPCG(1, 2))
		worsts := make([]float64, n)
		for i := range worsts {
			worsts[i] = worstWindow(s.Draw(rng), 5)
		}
		sort.Float64s(worsts)
		return worsts[n/20]
	}
	rp5, ip5 := p5(regime), p5(iid)
	if !(rp5 < ip5-0.03) {
		t.Errorf("NewMarkovRegime worst-5y p5 = %.3f, i.i.d. = %.3f; regime should be clearly worse (sequence risk)",
			rp5, ip5)
	}
}

// TestNewLostDecadeRegime verifies the two defining properties of the Japan-style
// stress: it is NOT mean-preserving (the blended long-run mean falls clearly
// below mu, because the deep bear is left to drag the average down), and it
// produces deeper multi-year drawdowns than the mean-preserving sequence-risk
// regime at the same mu/sigma.
func TestNewLostDecadeRegime(t *testing.T) {
	const n, periods = 4000, 40
	mu, sigma := 0.05, 0.13
	lost := NewLostDecadeRegime(mu, sigma, 5, periods)
	seq := NewMarkovRegime(mu, sigma, 5, periods)

	// Mean is materially below mu (a real haircut, not just sequence noise).
	rng := rand.New(rand.NewPCG(7, 0))
	var sum float64
	var count int
	for range n {
		for _, r := range lost.Draw(rng) {
			sum += r
			count++
		}
	}
	if got := sum / float64(count); !(got < mu-0.015) {
		t.Errorf("lost-decade mean = %.4f, want clearly below mu=%.3f (not mean-preserving)", got, mu)
	}

	// Worst 10-year window is deeper than the mean-preserving regime's.
	p5 := func(s Source) float64 {
		rng := rand.New(rand.NewPCG(1, 2))
		worsts := make([]float64, n)
		for i := range worsts {
			worsts[i] = worstWindow(s.Draw(rng), 10)
		}
		sort.Float64s(worsts)
		return worsts[n/20]
	}
	if lp5, sp5 := p5(lost), p5(seq); !(lp5 < sp5) {
		t.Errorf("lost-decade worst-10y p5 = %.3f should be below the sequence regime %.3f", lp5, sp5)
	}
}

func TestMarkovRegimeLen(t *testing.T) {
	m := MarkovRegime{Periods: 25}
	if m.Len() != 25 {
		t.Errorf("Len = %d, want 25", m.Len())
	}
	rng := rand.New(rand.NewPCG(1, 2))
	if got := len(m.Draw(rng)); got != 25 {
		t.Errorf("Draw len = %d, want 25", got)
	}
}
