package scenario

import (
	"math"
	"math/rand/v2"
)

// Glidepath draws a two-asset (equity, bond) real-return path whose equity
// weight moves linearly from StartEquity to EndEquity across the horizon. A
// rising-equity glidepath (a low equity weight at retirement climbing later, the
// Pfau-Kitces "bond tent") defends the sequence-of-returns danger zone: the
// first years, when a crash is most damaging, are the least equity-heavy.
//
// Equity and bond returns share a bivariate Student-t (same degrees of freedom,
// correlation Corr), so a joint crash is possible and the diversification is not
// overstated. Each marginal is standardised to its own Sigma. A drawn blended
// return is clamped so that 1+r >= 0.
type Glidepath struct {
	EquityMu, EquitySigma  float64
	BondMu, BondSigma      float64
	Df                     float64 // shared tail; <=2 falls back to Normal
	Corr                   float64 // equity/bond correlation in [-1,1]
	StartEquity, EndEquity float64 // equity weight at year 0 and the final year
	Periods                int
}

// Len reports the path length.
func (g Glidepath) Len() int { return g.Periods }

// Draw returns one blended real-return path.
func (g Glidepath) Draw(rng *rand.Rand) Sequence {
	eScale, bScale := g.EquitySigma, g.BondSigma
	if g.Df > 2 {
		f := math.Sqrt(g.Df / (g.Df - 2))
		eScale, bScale = g.EquitySigma/f, g.BondSigma/f
	}
	rho := math.Max(-1, math.Min(1, g.Corr))
	comp := math.Sqrt(1 - rho*rho)
	seq := make(Sequence, g.Periods)
	for t := range seq {
		z1 := rng.NormFloat64()
		z2 := rho*z1 + comp*rng.NormFloat64()
		s := 1.0
		if g.Df > 2 {
			s = math.Sqrt(g.Df / (2 * gamma(rng, g.Df/2))) // sqrt(df/chi2), the shared t mixing
		}
		e := g.EquityMu + eScale*z1*s
		b := g.BondMu + bScale*z2*s
		w := g.weightAt(t)
		r := w*e + (1-w)*b
		if 1+r < 0 {
			r = -1
		}
		seq[t] = r
	}
	return seq
}

// weightAt is the equity weight in a given year, linear from StartEquity to
// EndEquity; a single-period path uses StartEquity.
func (g Glidepath) weightAt(t int) float64 {
	if g.Periods <= 1 {
		return g.StartEquity
	}
	frac := float64(t) / float64(g.Periods-1)
	return g.StartEquity + (g.EndEquity-g.StartEquity)*frac
}
