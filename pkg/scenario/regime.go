package scenario

import (
	"math"
	"math/rand/v2"
)

// MarkovRegime draws returns from a two-state Markov chain, a calm state and a
// bear state, so bad years cluster into prolonged real drawdowns: the
// sequence-of-returns risk that i.i.d. parametric draws miss. The bear state
// has a lower mean and higher volatility and is sticky (a high probability of
// staying), reproducing the multi-year real bear markets of broad historical
// samples (e.g. Japan post-1990). Within a state, returns are Student-t at Df,
// standardised to that state's sigma; the path starts from the chain's
// stationary distribution so a retirement can begin in a downturn.
//
// It is a generic Source, usable wherever ParametricSource is, and is meant as
// a more honest stress alternative: at comparable calm parameters it has a
// fatter left tail and deeper multi-year drawdowns, and a slightly lower
// blended mean (time spent in the bear state).
type MarkovRegime struct {
	CalmMu, CalmSigma float64
	BearMu, BearSigma float64
	StayCalm, StayBear float64 // probability of staying in the current state, [0,1)
	Df                 float64
	Periods            int
}

// Len reports the path length.
func (m MarkovRegime) Len() int { return m.Periods }

// Draw returns one regime-switching path.
func (m MarkovRegime) Draw(rng *rand.Rand) Sequence {
	// Stationary probability of the bear state: the two off-diagonal flows
	// balance, so pi_bear = P(calm->bear) / (P(calm->bear) + P(bear->calm)).
	toBear, toCalm := 1-m.StayCalm, 1-m.StayBear
	piBear := 0.0
	if toBear+toCalm > 0 {
		piBear = toBear / (toBear + toCalm)
	}
	bear := rng.Float64() < piBear

	seq := make(Sequence, m.Periods)
	for i := range seq {
		if bear {
			bear = rng.Float64() <= m.StayBear
		} else {
			bear = rng.Float64() > m.StayCalm
		}
		mu, sigma := m.CalmMu, m.CalmSigma
		if bear {
			mu, sigma = m.BearMu, m.BearSigma
		}
		scale := sigma
		if m.Df > 2 {
			scale = sigma / math.Sqrt(m.Df/(m.Df-2))
		}
		r := mu + scale*studentT(rng, m.Df)
		if 1+r < 0 {
			r = -1 // an extreme draw cannot take capital below zero
		}
		seq[i] = r
	}
	return seq
}
