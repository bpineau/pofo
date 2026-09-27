package metrics

import (
	"fmt"
	"math"
)

// Corr is the Pearson correlation of two equal-length samples, a number in
// [-1, 1]. The samples are whatever the caller pairs, usually the per-period
// returns of two assets on one calendar (fractions, though correlation does
// not depend on the unit). It is 0 when the lengths differ, when there are
// fewer than two points, or when either sample is constant: with no variance
// there is no co-movement to measure.
func Corr(a, b []float64) float64 {
	if len(a) != len(b) || len(a) < 2 {
		return 0
	}
	ma, mb := Mean(a), Mean(b)
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

// CorrelationMatrix is the Pearson correlation of every pair of assets:
// out[i][j] = Corr(returns[i], returns[j]). returns is [asset][period], every
// asset's per-period returns (fractions) on ONE calendar, which is what
// marketdata.Aligned.Returns produces. The matrix is symmetric, its diagonal
// is 1, except for an asset whose returns are constant, whose whole row and
// column read 0 as Corr says. It is an error when the rows differ in length:
// periods that are not the same periods pair into no number at all (build
// the rows on one calendar, a marketdata.Panel or Aligned, first).
func CorrelationMatrix(returns [][]float64) ([][]float64, error) {
	if err := checkRectangular("CorrelationMatrix", returns); err != nil {
		return nil, err
	}
	out := squareMatrix(len(returns))
	for i := range returns {
		if sampleStdev(returns[i]) > 0 {
			out[i][i] = 1
		}
		for j := i + 1; j < len(returns); j++ {
			c := Corr(returns[i], returns[j])
			out[i][j], out[j][i] = c, c
		}
	}
	return out, nil
}

// Covariance is the per-period SAMPLE covariance (n-1 denominator) of every
// pair of assets. returns is [asset][period], per-period returns as fractions
// on ONE calendar (marketdata.Aligned.Returns), and the result is in squared
// fractions per period: multiply by the calendar's PeriodsPerYear to
// annualize it (252 for daily returns, 12 for monthly ones). The matrix is
// symmetric and its diagonal is each asset's sample variance. Fewer than two
// periods leave every entry NaN. It is an error when the rows differ in
// length, as for CorrelationMatrix.
func Covariance(returns [][]float64) ([][]float64, error) {
	if err := checkRectangular("Covariance", returns); err != nil {
		return nil, err
	}
	out := squareMatrix(len(returns))
	means := make([]float64, len(returns))
	for i, r := range returns {
		means[i] = Mean(r)
	}
	for i := range returns {
		for j := i; j < len(returns); j++ {
			c := math.NaN()
			if n := len(returns[i]); n >= 2 {
				var s float64
				for k := range returns[i] {
					s += (returns[i][k] - means[i]) * (returns[j][k] - means[j])
				}
				c = s / float64(n-1)
			}
			out[i][j], out[j][i] = c, c
		}
	}
	return out, nil
}

// checkRectangular is an error naming the first row of returns whose length
// differs from the first one's.
func checkRectangular(fn string, returns [][]float64) error {
	for i, r := range returns {
		if len(r) != len(returns[0]) {
			return fmt.Errorf("metrics: %s: asset %d has %d returns, asset 0 has %d (not one calendar)", fn, i, len(r), len(returns[0]))
		}
	}
	return nil
}

// squareMatrix is an n by n matrix of zeros.
func squareMatrix(n int) [][]float64 {
	out := make([][]float64, n)
	for i := range out {
		out[i] = make([]float64, n)
	}
	return out
}
