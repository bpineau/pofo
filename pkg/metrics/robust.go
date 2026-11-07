package metrics

import (
	"fmt"
	"math"
)

// ShrinkageCovariance returns sample covariance linearly shrunk toward its
// own diagonal: diagonal entries stay unchanged and off-diagonal entries are
// multiplied by 1-shrinkage. The coefficient is explicit in [0,1]; this does
// not estimate an automatic or optimal shrinkage intensity.
//
// returns is [asset][period], finite samples on one calendar with at least
// two observations. Units are squared fractions per period, as Covariance;
// multiply by periods per year to annualize. Empty, ragged, nonfinite samples
// and coefficients outside [0,1] are errors.
func ShrinkageCovariance(returns [][]float64, shrinkage float64) ([][]float64, error) {
	if isNaNOrInf(shrinkage) || shrinkage < 0 || shrinkage > 1 {
		return nil, fmt.Errorf("metrics: ShrinkageCovariance: shrinkage must be finite and in [0,1]")
	}
	if len(returns) == 0 || len(returns[0]) < 2 {
		return nil, fmt.Errorf("metrics: ShrinkageCovariance: need assets with at least two observations")
	}
	if err := checkRectangular("ShrinkageCovariance", returns); err != nil {
		return nil, err
	}
	for i, row := range returns {
		for j, v := range row {
			if isNaNOrInf(v) {
				return nil, fmt.Errorf("metrics: ShrinkageCovariance: asset %d observation %d is not finite", i, j)
			}
		}
	}
	cov, err := Covariance(returns)
	if err != nil {
		return nil, err
	}
	for i := range cov {
		for j := range cov[i] {
			if isNaNOrInf(cov[i][j]) {
				return nil, fmt.Errorf("metrics: ShrinkageCovariance: covariance overflows")
			}
			if i != j {
				cov[i][j] *= 1 - shrinkage
			}
		}
	}
	return cov, nil
}

// RegressHAC fits the same intercept and coefficients as Regress, replacing
// their classical standard errors and t-statistics with the Newey-West HAC
// sandwich estimator. Bartlett weights are 1-lag/(lags+1), with lags in
// [0,N-1]. Lags zero gives the heteroskedasticity-robust HC0 estimator.
// No degrees-of-freedom correction is applied to the sandwich covariance.
// ResidualSD remains Regress's residual deviation with N-k-1 denominator.
//
// The observations must be in time order at one regular cadence. Callers
// choose the lag count; the function does not infer one. Rank deficiency,
// nonfinite inputs and too few observations are errors, as for Regress.
// The robust SEs are asymptotic estimation uncertainty under weak dependence,
// not predictive intervals or a correction for a misspecified mean model.
// Centered Householder QR supplies the inverse triangular factors; the
// implementation never inverts the raw normal equations.
//
// Reference: Newey and West, "A Simple, Positive Semi-Definite,
// Heteroskedasticity and Autocorrelation Consistent Covariance Matrix",
// Econometrica 55(3), 1987, pp. 703-708, equation (5):
// https://users.ssc.wisc.edu/~behansen/718/NeweyWest1987.pdf.
func RegressHAC(y []float64, lags int, xs ...[]float64) (Regression, error) {
	if lags < 0 || lags >= len(y) {
		return Regression{}, fmt.Errorf("metrics: RegressHAC: lags must be in [0,N-1]")
	}
	out, err := Regress(y, xs...)
	if err != nil {
		return Regression{}, err
	}
	n, k := len(y), len(xs)
	means := make([]float64, k)
	centered := make([][]float64, k)
	for j, x := range xs {
		means[j] = Mean(x)
		centered[j] = make([]float64, n)
		for t, v := range x {
			centered[j][t] = v - means[j]
		}
	}
	r, _, err := householderQR(centered, make([]float64, n))
	if err != nil {
		return Regression{}, err
	}
	rinv := invertUpper(r)
	// With centered regressors, bread is block diagonal: 1/N for the
	// mean and R^-1 R^-T for the slopes. Influence rows below map each
	// observation's residual to an ORIGINAL (uncentered) coefficient.
	bread := squareMatrix(k)
	for i := range k {
		for j := range k {
			bread[i][j] = dot(rinv[i], rinv[j])
		}
	}
	influence := make([][]float64, k+1)
	for i := range influence {
		influence[i] = make([]float64, n)
	}
	for t := range n {
		residual := y[t] - out.Alpha.Value
		for j, x := range xs {
			residual -= out.Betas[j].Value * x[t]
		}
		if isNaNOrInf(residual) {
			return Regression{}, fmt.Errorf("metrics: RegressHAC: residual overflows")
		}
		intercept := 1 / float64(n)
		for i := range k {
			v := 0.0
			for j, x := range xs {
				v += bread[i][j] * (x[t] - means[j])
			}
			influence[i+1][t] = v * residual
			intercept -= means[i] * v
		}
		influence[0][t] = intercept * residual
	}
	// Only diagonal sandwich entries are needed for coefficient SEs.
	for i, row := range influence {
		variance := dot(row, row)
		for lag := 1; lag <= lags; lag++ {
			weight := 1 - float64(lag)/float64(lags+1)
			variance += 2 * weight * dot(row[lag:], row[:n-lag])
		}
		if isNaNOrInf(variance) || variance < -1e-12*dot(row, row) {
			return Regression{}, fmt.Errorf("metrics: RegressHAC: covariance is not numerically finite and nonnegative")
		}
		se := math.Sqrt(math.Max(variance, 0))
		if i == 0 {
			out.Alpha = estimate(out.Alpha.Value, se)
		} else {
			out.Betas[i-1] = estimate(out.Betas[i-1].Value, se)
		}
	}
	return out, nil
}
