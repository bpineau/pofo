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
