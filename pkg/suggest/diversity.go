package suggest

import "math"

// mean returns the arithmetic mean of xs.
func mean(xs []float64) float64 {
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

// std returns the sample standard deviation (n-1) of xs.
func std(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	m := mean(xs)
	s := 0.0
	for _, x := range xs {
		d := x - m
		s += d * d
	}
	return math.Sqrt(s / float64(len(xs)-1))
}

// Correlation is the Pearson correlation of two equal-length series.
// It returns 0 when either series is constant or lengths differ.
func Correlation(a, b []float64) float64 {
	if len(a) != len(b) || len(a) < 2 {
		return 0
	}
	ma, mb := mean(a), mean(b)
	var cov, va, vb float64
	for i := range a {
		da, db := a[i]-ma, b[i]-mb
		cov += da * db
		va += da * da
		vb += db * db
	}
	if va == 0 || vb == 0 {
		return 0
	}
	return cov / math.Sqrt(va*vb)
}

// DiversificationRatio is (sum of weighted asset volatilities) / (portfolio
// volatility): 1 when every asset moves together, up to sqrt(N) when they
// are independent. The effective number of independent bets is approximately
// its square. weights are fractions; returns[i] is asset i's daily-return
// series (all equal length).
func DiversificationRatio(weights []float64, returns [][]float64) float64 {
	if len(weights) == 0 || len(returns) != len(weights) {
		return 0
	}
	n := len(returns[0])
	weightedVol := 0.0
	port := make([]float64, n)
	for i, r := range returns {
		weightedVol += weights[i] * std(r)
		for t := 0; t < n; t++ {
			port[t] += weights[i] * r[t]
		}
	}
	pv := std(port)
	if pv == 0 {
		return 0
	}
	return weightedVol / pv
}

// PortfolioReturns returns the weighted sum of the per-asset daily-return
// series (the held portfolio's aggregate return), for callers building a
// Candidate's PortReturns over an overlap window.
func PortfolioReturns(weights []float64, returns [][]float64) []float64 {
	if len(returns) == 0 {
		return nil
	}
	n := len(returns[0])
	out := make([]float64, n)
	for i, r := range returns {
		for t := 0; t < n && t < len(r); t++ {
			out[t] += weights[i] * r[t]
		}
	}
	return out
}
