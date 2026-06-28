package scenario

import (
	"math"
	"math/rand/v2"
)

// Sequence is a periodic real-return path, e.g. 30 annual returns where
// 0.04 means +4 % over the period.
type Sequence []float64

// Source produces synthetic return paths of a fixed length.
type Source interface {
	// Draw returns one path of Len periods using rng.
	Draw(rng *rand.Rand) Sequence
	// Len is the number of periods in every path Draw returns.
	Len() int
}

// ParametricSource draws i.i.d. returns from a Student-t distribution
// scaled so the mean is Mu and the standard deviation is exactly Sigma.
// Df is the degrees of freedom (lower = fatter tails); Df <= 2 falls back
// to a Normal. Each drawn return r is clamped so that 1+r >= 0.
type ParametricSource struct {
	Mu, Sigma, Df float64
	Periods       int
}

// Len reports the path length.
func (p ParametricSource) Len() int { return p.Periods }

// Draw returns one i.i.d. path.
func (p ParametricSource) Draw(rng *rand.Rand) Sequence {
	scale := p.Sigma
	if p.Df > 2 {
		scale = p.Sigma / math.Sqrt(p.Df/(p.Df-2))
	}
	seq := make(Sequence, p.Periods)
	for i := range seq {
		r := p.Mu + scale*studentT(rng, p.Df)
		if 1+r < 0 {
			r = -1
		}
		seq[i] = r
	}
	return seq
}

// studentT returns a standard Student-t variate at df degrees of freedom
// (variance df/(df-2) for df>2). df <= 0 returns a standard normal.
func studentT(rng *rand.Rand, df float64) float64 {
	if df <= 0 {
		return rng.NormFloat64()
	}
	z := rng.NormFloat64()
	// chi-square(df) = 2 * Gamma(df/2, 1); scale cancels in the ratio.
	chi2 := 2 * gamma(rng, df/2)
	return z / math.Sqrt(chi2/df)
}

// gamma draws from Gamma(shape, 1) via Marsaglia-Tsang (shape >= 1) with
// the standard boost for shape < 1. Stdlib-only.
func gamma(rng *rand.Rand, shape float64) float64 {
	if shape < 1 {
		return gamma(rng, shape+1) * math.Pow(rng.Float64(), 1/shape)
	}
	d := shape - 1.0/3.0
	c := 1.0 / math.Sqrt(9*d)
	for {
		x := rng.NormFloat64()
		v := 1 + c*x
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := rng.Float64()
		if u < 1-0.0331*x*x*x*x {
			return d * v
		}
		if math.Log(u) < 0.5*x*x+d*(1-v+math.Log(v)) {
			return d * v
		}
	}
}
