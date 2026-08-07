package metrics

import (
	"errors"
	"math"
)

// ErrNoContributions is returned when an attribution is asked of an empty or
// too short contribution matrix.
var ErrNoContributions = errors.New("metrics: no usable contribution series")

// Attribution splits one portfolio quantity across its holdings. Shares are
// fractions of the portfolio total (0.25 = 25 %) in the input's asset order,
// and they sum to 1 by construction.
type Attribution struct {
	Risk   []float64 // share of the portfolio's variance (Euler decomposition)
	Return []float64 // share of the portfolio's realized return
}

// Attribute decomposes a portfolio's risk and realized return across its
// holdings, from the per-holding contribution series that a simulation
// produces (portfolio.SimResult.Contributions: contrib[i][k] is holding i's
// contribution to period k's portfolio return, held shares x price move over
// portfolio value).
//
// Working from contributions rather than from weights and returns separately
// is what makes this exact rather than indicative: the contribution series
// already carries the drifting weights between rebalancings, the rebalancing
// trades themselves and any embedded leverage, so no static weight vector has
// to stand in for a position that moved all period.
//
// The return share is the plain sum of a holding's contributions over the
// portfolio's, i.e. how much of what the portfolio earned came from it.
//
// The risk share is the Euler decomposition of variance: holding i's share is
// Cov(c_i, r_p) / Var(r_p), where r_p is the portfolio return (the sum of the
// contributions). Covariance is linear in its first argument, so these shares
// sum to exactly 1 with no normalization, and each one answers "how much does
// this holding move WITH the portfolio", which is the honest question: an
// individually volatile line that zigs when the book zags takes a small share,
// and can legitimately take a negative one.
//
// Both are shares, so both are scale-free: the result does not depend on the
// period length or on whether contributions are daily or monthly. It does
// depend on the window, and unequally so, which callers should surface rather
// than hide. Covariances are estimated far more reliably than means, so the
// risk split is comparatively stable across windows while the return split
// moves with every regime; and a holding bought as insurance is MEANT to show
// a near-zero return share against a real risk share, so the two columns must
// never be read as an efficiency ratio.
//
// Contributions that do not belong to any holding (envelope fees, the leverage
// cash leg) are absent from the matrix by construction, so the portfolio return
// rebuilt here is the attributable part of it.
func Attribute(contrib [][]float64) (Attribution, error) {
	if len(contrib) == 0 || len(contrib[0]) < 2 {
		return Attribution{}, ErrNoContributions
	}
	n := len(contrib[0])
	for _, c := range contrib {
		if len(c) != n {
			return Attribution{}, errors.New("metrics: contribution series of unequal length")
		}
	}

	// The portfolio return of each period is the sum of its contributions.
	port := make([]float64, n)
	for _, c := range contrib {
		for k, v := range c {
			port[k] += v
		}
	}

	att := Attribution{
		Risk:   make([]float64, len(contrib)),
		Return: make([]float64, len(contrib)),
	}

	total := 0.0
	for _, v := range port {
		total += v
	}
	for i, c := range contrib {
		sum := 0.0
		for _, v := range c {
			sum += v
		}
		if total != 0 {
			att.Return[i] = sum / total
		}
	}

	meanPort := total / float64(n)
	varPort := 0.0
	scale := 0.0
	for _, v := range port {
		varPort += (v - meanPort) * (v - meanPort)
		scale = math.Max(scale, math.Abs(v))
	}
	// A portfolio that does not move has no variance to share out, and the
	// test has to be RELATIVE: subtracting a mean from identical values leaves
	// cancellation noise around 1e-33 rather than an exact zero, which would
	// sail past "varPort == 0" and turn that noise into confident-looking
	// shares. Comparing the deviation to the size of the contributions
	// themselves rejects the degenerate case at any scale.
	if scale == 0 || math.Sqrt(varPort/float64(n)) <= 1e-9*scale {
		return Attribution{}, ErrNoContributions
	}
	for i, c := range contrib {
		meanI := 0.0
		for _, v := range c {
			meanI += v
		}
		meanI /= float64(n)
		cov := 0.0
		for k, v := range c {
			cov += (v - meanI) * (port[k] - meanPort)
		}
		att.Risk[i] = cov / varPort
	}
	return att, nil
}
