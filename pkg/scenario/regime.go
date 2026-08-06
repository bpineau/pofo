package scenario

import (
	"math"
	"math/rand/v2"
)

// Default Markov-regime persistence probabilities. These control how sticky
// each state is: a high stayCalm means calm years run in streaks; a high
// stayBear means downturns last several years, as in historical equity bear
// markets. Together they imply a stationary bear probability of roughly 18.6%.
const (
	stayCalm = 0.92 // probability of remaining in the calm state each period
	stayBear = 0.65 // probability of remaining in the bear state each period
)

// bearGapFactor is the multiple of sigma by which the bear-state mean falls
// below the calm-state mean. Scaling with sigma keeps the spread proportional
// to the volatility of the underlying return series. Calibrated so the sequence
// risk it adds lowers the safe withdrawal rate by a realistic ~0.3-0.5% vs the
// i.i.d. case (not the ~1% a deeper bear would imply, which overstates the
// historical sequence-risk penalty).
const bearGapFactor = 0.6

// bearSigmaFactor is the multiple applied to the calm-state sigma in the bear
// state. Volatility clustering (bear markets are more volatile than calm
// markets) increases the left-tail severity without relying solely on the lower
// mean.
const bearSigmaFactor = 1.5

// NewMarkovRegime builds a mean-preserving two-state regime: the calm mean is
// lifted just enough that the time spent in the (deeper, more volatile, sticky)
// bear state leaves the blended long-run mean equal to mu. It injects sequence
// risk (clustered, persistent real drawdowns and a fatter left tail) WITHOUT
// changing the long-run expected return.
//
// Persistence uses the package constants stayCalm and stayBear. The bear state
// is bearGapFactor*sigma worse than the calm state (mean) and bearSigmaFactor
// times as volatile.
func NewMarkovRegime(mu, sigma, df float64, periods int) MarkovRegime {
	toBear := 1 - stayCalm
	toCalm := 1 - stayBear
	piBear := toBear / (toBear + toCalm)

	bearGap := bearGapFactor * sigma // bear mean is this much below calm mean
	calmMu := mu + piBear*bearGap    // lift calm so the blend equals mu
	bearMu := calmMu - bearGap       // equivalent to mu - (1-piBear)*bearGap

	return MarkovRegime{
		CalmMu: calmMu, CalmSigma: sigma,
		BearMu: bearMu, BearSigma: sigma * bearSigmaFactor,
		StayCalm: stayCalm, StayBear: stayBear,
		Df: df, Periods: periods,
	}
}

// Lost-decade regime constants: a very persistent, deep bear that models a
// Japan-style prolonged real drawdown (Japanese equities were roughly flat to
// negative in real terms from 1990 to 2010). The calm state sits at the
// requested mu while the bear is a deep, sticky trough averaging a full decade,
// so the blended long-run mean falls BELOW mu (this regime is deliberately not
// mean-preserving, unlike NewMarkovRegime).
const (
	lostStayCalm        = 0.95 // calm-to-bear switch roughly once in 20 years
	lostStayBear        = 0.90 // mean bear run ~10 years: the "lost decade"
	lostBearGapFactor   = 0.8  // bear mean is 0.8*sigma below the calm mean
	lostBearSigmaFactor = 1.4
)

// NewLostDecadeRegime builds a Japan-style lost-decade stress: a very sticky,
// deep bear state (mean run ~10 years) layered on a calm state at mu. It is NOT
// mean-preserving: the time spent in the trough drags the blended long-run mean
// below mu, so it bakes in a genuinely lower realised return, not just an
// unlucky sequence. Read it as the tail scenario where a whole retirement lands
// inside a lost decade, the grimmest of the planning models, not a central case.
//
// Contrast NewMarkovRegime, whose calm mean is lifted to preserve mu (it stresses
// the sequence only). Here the calm mean stays at mu and the trough is left to
// pull the average down.
func NewLostDecadeRegime(mu, sigma, df float64, periods int) MarkovRegime {
	return MarkovRegime{
		CalmMu: mu, CalmSigma: sigma,
		BearMu: mu - lostBearGapFactor*sigma, BearSigma: sigma * lostBearSigmaFactor,
		StayCalm: lostStayCalm, StayBear: lostStayBear,
		Df: df, Periods: periods,
	}
}

// MarkovRegime draws returns from a two-state Markov chain, a calm state and a
// bear state, so bad years cluster into prolonged real drawdowns: the
// sequence-of-returns risk that i.i.d. parametric draws miss. The bear state
// has a lower mean and higher volatility and is sticky (a high probability of
// staying), reproducing the multi-year real bear markets of broad historical
// samples (e.g. Japan post-1990). Within a state, returns are Student-t at Df,
// standardised to that state's sigma; the path starts from the chain's
// stationary distribution so a retirement can begin in a downturn.
//
// Use NewMarkovRegime to build a mean-preserving regime from a target
// mean/sigma/df; constructing the struct directly allows custom state
// parameters, but the blended long-run mean is then set by the caller.
type MarkovRegime struct {
	CalmMu, CalmSigma  float64
	BearMu, BearSigma  float64
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
