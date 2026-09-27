package metrics

import (
	"errors"
	"fmt"
	"math"
)

// Estimate is one fitted coefficient of a Regression: its Value, the
// standard error of that value, and T = Value / SE, the t-statistic of the
// hypothesis that the true coefficient is zero (|T| above about 2 is the
// usual threshold of a coefficient distinguishable from zero). On an exact
// fit SE is zero and T is infinite, or NaN for a zero Value.
type Estimate struct {
	Value, SE, T float64 // the coefficient, its standard error, Value / SE
}

// Regression is an ordinary least squares fit with an intercept,
//
//	y[t] = Alpha + Betas[0]*xs[0][t] + Betas[1]*xs[1][t] + ... + e[t],
//
// as Regress returns it. Every figure is PER PERIOD, in the unit of the
// samples: on monthly returns (fractions), Alpha is a monthly return and
// ResidualSD a monthly standard deviation. AnnualAlpha and AnnualResidualVol
// annualize them at the samples' cadence; the betas and R2 are unit-free and
// need nothing.
type Regression struct {
	N          int        // observations fitted
	Alpha      Estimate   // the intercept, per period
	Betas      []Estimate // one per regressor, in the order Regress received them
	R2         float64    // share of y's variance the regressors explain, in [0, 1]; NaN when y is constant
	ResidualSD float64    // standard deviation of the residuals, sqrt(SSR / (N - k - 1)) for k regressors, per period
}

// AnnualAlpha is the intercept annualized arithmetically, Alpha.Value times
// periodsPerYear (12 for monthly returns, TradingDaysPerYear for daily ones):
// the yearly return y earns beyond what its exposures to the regressors
// explain. Its standard error scales the same way, so Alpha.T is unchanged.
func (r Regression) AnnualAlpha(periodsPerYear float64) float64 {
	return r.Alpha.Value * periodsPerYear
}

// AnnualResidualVol is ResidualSD annualized like a volatility, times the
// square root of periodsPerYear: the yearly tracking error of y against the
// fitted combination of its regressors, the risk the regression leaves
// unexplained.
func (r Regression) AnnualResidualVol(periodsPerYear float64) float64 {
	return r.ResidualSD * math.Sqrt(periodsPerYear)
}

// collinearTolerance is how small, relative to its own centered norm, the
// part of a regressor that the previous ones do not already span may be
// before the regressor is declared collinear with them: a coefficient
// resting on less would amplify rounding noise by more than 1e10.
const collinearTolerance = 1e-10

// Regress fits y on the regressors xs by ordinary least squares with an
// intercept, the regression behind a fund's alpha and betas on its index and
// a currency, or a strategy's exposure to two factors. y and every
// regressor are samples of one length, paired by position (the columns of a
// marketdata.Panel are).
//
// The fit is numerically sound: the data are centered, which removes the
// intercept from the linear system, and the centered regressors are
// factorized by Householder QR, so the solution never forms the normal
// equations X'X, whose condition number is the square of the data's. It
// reproduces the certified values of the NIST StRD Longley benchmark (six
// nearly collinear regressors) to eleven significant digits and more (see
// pkg/datasets/golden).
//
// The standard errors are the classical homoskedastic ones, from the
// residual variance SSR / (N - k - 1): they assume residuals that are
// independent and of constant variance, which overlapping or autocorrelated
// returns are not, so read a T-statistic on such data as optimistic.
//
// It returns an error when xs is empty, when the lengths differ, when a
// value is not finite, when there are fewer than k + 2 observations for k
// regressors (no residual degree of freedom is left), and when a regressor
// is constant or collinear with the others (its coefficient is then not
// identified). A constant y is not an error: its betas are zero, its fit
// exact and its R2 NaN.
func Regress(y []float64, xs ...[]float64) (Regression, error) {
	n, k := len(y), len(xs)
	if k == 0 {
		return Regression{}, errors.New("metrics: Regress: no regressor")
	}
	for j, x := range xs {
		if len(x) != n {
			return Regression{}, fmt.Errorf("metrics: Regress: regressor %d has %d values, y has %d", j, len(x), n)
		}
	}
	if n < k+2 {
		return Regression{}, fmt.Errorf("metrics: Regress: %d observations for %d regressors, at least %d needed", n, k, k+2)
	}
	for t := range n {
		if isNaNOrInf(y[t]) {
			return Regression{}, fmt.Errorf("metrics: Regress: y[%d] is not finite", t)
		}
		for j, x := range xs {
			if isNaNOrInf(x[t]) {
				return Regression{}, fmt.Errorf("metrics: Regress: regressor %d, value %d is not finite", j, t)
			}
		}
	}

	// Centering: the slopes of the centered data are the slopes of the
	// model with an intercept, and the intercept follows from the means.
	my := Mean(y)
	yc := make([]float64, n)
	for t, v := range y {
		yc[t] = v - my
	}
	means := make([]float64, k)
	cols := make([][]float64, k) // centered regressors, factorized in place
	for j, x := range xs {
		means[j] = Mean(x)
		cols[j] = make([]float64, n)
		for t, v := range x {
			cols[j][t] = v - means[j]
		}
	}

	r, qty, err := householderQR(cols, yc)
	if err != nil {
		return Regression{}, err
	}
	rinv := invertUpper(r)

	// The slopes solve R b = (Q'y)[:k].
	b := make([]float64, k)
	for i := range k {
		for m := i; m < k; m++ {
			b[i] += rinv[i][m] * qty[m]
		}
	}

	// Residuals on the original centered data rather than off Q'y's tail:
	// the same number in exact arithmetic, one rounding step closer to the
	// data in floating point.
	var ssr, sst float64
	for t := range n {
		fit := 0.0
		for j := range k {
			fit += b[j] * (xs[j][t] - means[j])
		}
		e := yc[t] - fit
		ssr += e * e
		sst += yc[t] * yc[t]
	}
	s := math.Sqrt(ssr / float64(n-k-1))

	out := Regression{N: n, ResidualSD: s, Betas: make([]Estimate, k), R2: math.NaN()}
	if sst > 0 {
		out.R2 = 1 - ssr/sst
	}
	// Var(b) = s² (X'X)⁻¹ = s² R⁻¹R⁻ᵀ, so SE(b_i) is s times the norm of
	// row i of R⁻¹; Var(alpha) = s² (1/n + x̄'(X'X)⁻¹x̄) = s² (1/n + |R⁻ᵀx̄|²).
	alpha := my
	for i := range k {
		alpha -= b[i] * means[i]
		var sq float64
		for m := i; m < k; m++ {
			sq += rinv[i][m] * rinv[i][m]
		}
		out.Betas[i] = estimate(b[i], s*math.Sqrt(sq))
	}
	var lever float64
	for m := range k {
		var z float64 // (R⁻ᵀx̄)_m
		for i := 0; i <= m; i++ {
			z += rinv[i][m] * means[i]
		}
		lever += z * z
	}
	out.Alpha = estimate(alpha, s*math.Sqrt(1/float64(n)+lever))
	return out, nil
}

// estimate pairs a coefficient with its standard error and t-statistic.
func estimate(value, se float64) Estimate {
	return Estimate{Value: value, SE: se, T: value / se}
}

// householderQR factorizes the n by k matrix whose columns are cols (n >= k)
// as Q R by Householder reflections, overwriting cols, and applies the same
// reflections to y. It returns R (k by k, upper triangular, row-major) and
// Q'y. A column whose part orthogonal to the previous ones is below
// collinearTolerance of its own norm is reported as collinear, naming it.
func householderQR(cols [][]float64, y []float64) ([][]float64, []float64, error) {
	k, n := len(cols), len(y)
	norms := make([]float64, k)
	for j, c := range cols {
		norms[j] = norm(c)
	}
	qty := append([]float64(nil), y...)
	r := squareMatrix(k)
	for j := range k {
		if norms[j] == 0 {
			return nil, nil, fmt.Errorf("metrics: Regress: regressor %d is constant", j)
		}
		v := cols[j][j:]
		sigma := norm(v)
		if sigma <= collinearTolerance*norms[j] {
			return nil, nil, fmt.Errorf("metrics: Regress: regressor %d is collinear with the ones before it", j)
		}
		// Reflect v onto -sign(v0) sigma e1: the sign that avoids
		// cancellation in v0 - alpha.
		alpha := -math.Copysign(sigma, v[0])
		v[0] -= alpha
		vv := dot(v, v)
		reflect := func(w []float64) {
			f := 2 * dot(v, w) / vv
			for t := range w {
				w[t] -= f * v[t]
			}
		}
		for m := j + 1; m < k; m++ {
			reflect(cols[m][j:])
			r[j][m] = cols[m][j]
		}
		reflect(qty[j:n])
		r[j][j] = alpha
	}
	return r, qty, nil
}

// invertUpper inverts an upper-triangular matrix with a non-zero diagonal by
// back substitution, column by column.
func invertUpper(r [][]float64) [][]float64 {
	k := len(r)
	inv := squareMatrix(k)
	for c := range k {
		for i := c; i >= 0; i-- {
			sum := 0.0
			if i == c {
				sum = 1
			}
			for m := i + 1; m <= c; m++ {
				sum -= r[i][m] * inv[m][c]
			}
			inv[i][c] = sum / r[i][i]
		}
	}
	return inv
}

// norm is the Euclidean norm of v, scaled against overflow and underflow.
func norm(v []float64) float64 {
	var scale, ssq float64 = 0, 1
	for _, x := range v {
		if x == 0 {
			continue
		}
		a := math.Abs(x)
		if scale < a {
			ssq = 1 + ssq*(scale/a)*(scale/a)
			scale = a
		} else {
			ssq += (a / scale) * (a / scale)
		}
	}
	return scale * math.Sqrt(ssq)
}

// dot is the inner product of two equal-length vectors.
func dot(a, b []float64) float64 {
	var s float64
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

// isNaNOrInf reports whether x is not a finite number.
func isNaNOrInf(x float64) bool { return math.IsNaN(x) || math.IsInf(x, 0) }
