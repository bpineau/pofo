package optimize

import (
	"fmt"
	"math"
)

// BlackLitterman anchors the expected returns on the portfolio itself before
// optimizing: it reverse-optimizes the prior weights into the returns they
// implicitly expect, blends the owner's views into them, and maximizes the
// mean-variance utility of the result over the same box simplex every other
// objective uses. See the package documentation for the model and its traps.
const BlackLitterman Objective = "black-litterman"

// blTau is the scalar of the Black-Litterman model, the size of the
// equilibrium distribution's own uncertainty (the prior is N(pi, tau*Sigma)).
// It is fixed at He and Litterman's 0.05 and exposed nowhere because it
// CANCELS: the view uncertainty Omega below is proportional to tau, so tau
// divides out of the posterior mean exactly (TestBlackLittermanTauCancels).
const blTau = 0.05

// DefaultPriorSharpe is the Sharpe ratio the prior allocation is assumed to
// earn when the caller states no Spec.PriorReturn, which fixes the risk
// aversion at DefaultPriorSharpe divided by the prior's volatility. It is
// exported so a report can say where the risk aversion came from.
//
// He and Litterman's delta = 2.5 is deliberately NOT the default: it was
// calibrated on an equity market at 15 to 20 % volatility, and on a defensive
// 8 %-volatility book it implies a 1.6 %/yr expected return, against which
// every ordinary view reads as a large surprise and the posterior swings to
// the corners.
const DefaultPriorSharpe = 0.4

// View is one belief about expected returns: absolute when Versus is empty
// ("Asset earns Return a year"), relative otherwise ("Asset beats Versus by
// Return a year").
//
// The identifiers are the ones written in the portfolio file (or the resolved
// symbols); Spec.Resolve turns them into the positions the model works on.
// Posterior and Solve read those positions, so a View built by hand and
// passed in without a Resolve points at position 0: build views through a
// Spec and resolve them, exactly as bounds are.
type View struct {
	// Asset is the line the view is about, and Versus the line it is
	// measured against (empty for an absolute view).
	Asset, Versus string
	// Return is the view's figure as a fraction per year: the expected
	// return of Asset for an absolute view, the margin over Versus for a
	// relative one. It may be negative or zero, which are statements too.
	Return float64
	// Confidence is how sure the owner is, in (0,1). It sets the view's
	// uncertainty in closed form (see Posterior): 0.5 is He and Litterman's
	// own choice, near 0 switches the view off and near 1 makes it certain.
	Confidence float64

	// asset and versus are the resolved positions, filled by Spec.Resolve;
	// versus is -1 for an absolute view.
	asset, versus int
}

// String renders a view the way a portfolio file writes it.
func (v View) String() string {
	if v.Versus != "" {
		return fmt.Sprintf("%s beats %s by %.1f points/yr, confidence %.0f %%",
			v.Asset, v.Versus, v.Return*100, v.Confidence*100)
	}
	return fmt.Sprintf("%s earns %.1f %%/yr, confidence %.0f %%", v.Asset, v.Return*100, v.Confidence*100)
}

// ImpliedReturns is reverse optimization: the annualized expected returns
// under which prior is the optimal portfolio at risk aversion lambda, i.e.
// pi = lambda * cov * prior. It answers the question a portfolio's owner
// rarely asks, "what return does my allocation implicitly expect from each
// line", and it is the anchor the views are blended into.
//
// prior is a weight vector in the same order as cov; it need not sum to 1
// here (Solve requires that of Spec.Prior, since only an allocation can be
// the reference), and nil is returned when the two lengths disagree.
func ImpliedReturns(prior []float64, cov [][]float64, lambda float64) []float64 {
	if len(prior) != len(cov) {
		return nil
	}
	return scale(matVec(cov, prior), lambda)
}

// Posterior blends the implied (equilibrium) returns with the views and
// returns the Black-Litterman expected returns
//
//	mu = pi + tau*Sigma*P' * (P*tau*Sigma*P' + Omega)^-1 * (Q - P*pi)
//
// where P and Q are the view rows and figures, and the view uncertainty is
// the closed form
//
//	Omega = diag((1/c_k - 1) * (P*tau*Sigma*P')_kk)
//
// which is He and Litterman's own choice at a confidence c of 0.5, is
// monotone in c, and needs no inner solve. Idzorek's iterative confidence
// mapping is deliberately not used.
//
// The views must carry the positions Spec.Resolve fills; an unresolved or
// out-of-range view, a confidence outside (0,1) and a line compared with
// itself are all errors. With no view at all the implied returns are
// returned unchanged, which is the identity the objective is built on.
func Posterior(implied []float64, cov [][]float64, views []View) ([]float64, error) {
	return posteriorAt(implied, cov, views, blTau)
}

// posteriorAt is Posterior at an explicit tau, so a test can show that the
// scalar cancels.
func posteriorAt(implied []float64, cov [][]float64, views []View, tau float64) ([]float64, error) {
	n := len(implied)
	mu := append([]float64(nil), implied...)
	if len(views) == 0 {
		return mu, nil
	}
	if len(cov) != n {
		return nil, fmt.Errorf("black-litterman: %d implied returns for a %dx%d covariance", n, len(cov), len(cov))
	}
	p, q, err := viewRows(views, n)
	if err != nil {
		return nil, err
	}
	k := len(views)

	// tsp = tau*Sigma*P' (n x k), m = P*tau*Sigma*P' (k x k).
	tsp := make([][]float64, n)
	for i := range tsp {
		tsp[i] = make([]float64, k)
		for j := 0; j < k; j++ {
			s := 0.0
			for l := 0; l < n; l++ {
				s += cov[i][l] * p[j][l]
			}
			tsp[i][j] = tau * s
		}
	}
	m := make([][]float64, k)
	for a := 0; a < k; a++ {
		m[a] = make([]float64, k)
		for b := 0; b < k; b++ {
			s := 0.0
			for i := 0; i < n; i++ {
				s += p[a][i] * tsp[i][b]
			}
			m[a][b] = s
		}
	}
	// The system to invert, m + Omega, and the surprise the views carry.
	rhs := make([]float64, k)
	for a := 0; a < k; a++ {
		s := 0.0
		for i := 0; i < n; i++ {
			s += p[a][i] * implied[i]
		}
		rhs[a] = q[a] - s
		m[a][a] += (1/views[a].Confidence - 1) * m[a][a]
	}
	x, ok := solveLinear(m, rhs)
	if !ok {
		return nil, fmt.Errorf("black-litterman: the views are not independent enough to blend (the view covariance is singular)")
	}
	for i := 0; i < n; i++ {
		for a := 0; a < k; a++ {
			mu[i] += tsp[i][a] * x[a]
		}
	}
	return mu, nil
}

// viewRows turns the resolved views into the P rows and the Q vector of the
// model, refusing anything that would silently mean something else.
func viewRows(views []View, n int) ([][]float64, []float64, error) {
	p := make([][]float64, len(views))
	q := make([]float64, len(views))
	for j, v := range views {
		if v.asset < 0 || v.asset >= n {
			return nil, nil, fmt.Errorf("view %d (%s): unresolved holding, call Spec.Resolve first", j+1, v.Asset)
		}
		if v.Confidence <= 0 || v.Confidence >= 1 {
			return nil, nil, fmt.Errorf("view %d (%s): confidence %.4f is not in (0,1), where 0 switches the view off and 1 makes it a certainty", j+1, v.Asset, v.Confidence)
		}
		row := make([]float64, n)
		row[v.asset] = 1
		if v.Versus != "" {
			if v.versus < 0 || v.versus >= n {
				return nil, nil, fmt.Errorf("view %d (%s): unresolved holding %q, call Spec.Resolve first", j+1, v.Asset, v.Versus)
			}
			if v.versus == v.asset {
				return nil, nil, fmt.Errorf("view %d: %q is compared with itself", j+1, v.Asset)
			}
			row[v.versus] = -1
		}
		p[j], q[j] = row, v.Return
	}
	return p, q, nil
}

// solveLinear solves a*x = b by Gaussian elimination with partial pivoting,
// reporting ok=false when a is singular. The system is one row per view, so
// it is tiny and a direct method is the simplest thing that works.
func solveLinear(a [][]float64, b []float64) ([]float64, bool) {
	n := len(b)
	m := make([][]float64, n)
	for i := range m {
		m[i] = make([]float64, n+1)
		copy(m[i], a[i])
		m[i][n] = b[i]
	}
	for col := 0; col < n; col++ {
		pivot := col
		for r := col + 1; r < n; r++ {
			if math.Abs(m[r][col]) > math.Abs(m[pivot][col]) {
				pivot = r
			}
		}
		if math.Abs(m[pivot][col]) < 1e-14 {
			return nil, false
		}
		m[col], m[pivot] = m[pivot], m[col]
		for r := col + 1; r < n; r++ {
			f := m[r][col] / m[col][col]
			for c := col; c <= n; c++ {
				m[r][c] -= f * m[col][c]
			}
		}
	}
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		s := m[i][n]
		for c := i + 1; c < n; c++ {
			s -= m[i][c] * x[c]
		}
		x[i] = s / m[i][i]
	}
	return x, true
}

// riskAversion returns the lambda that scales the implied returns: the one
// the caller stated through prior-return (lambda = R / variance of the prior,
// so the prior expects exactly R), or the default that assumes the prior
// earns a Sharpe of DefaultPriorSharpe.
func riskAversion(prior []float64, cov [][]float64, priorReturn float64) (float64, error) {
	variance := quad(cov, prior)
	if !(variance > 0) {
		return 0, fmt.Errorf("black-litterman: the prior allocation has no variance, so no risk aversion can be read off it")
	}
	if priorReturn > 0 {
		return priorReturn / variance, nil
	}
	return DefaultPriorSharpe / math.Sqrt(variance), nil
}

// checkPrior refuses anything but an allocation: the prior is the weights
// written in the portfolio file, and reverse optimization is meaningless
// against a vector that is not one.
func checkPrior(prior []float64, n int) error {
	if len(prior) != n {
		return fmt.Errorf("black-litterman: the prior holds %d weights for %d assets; Spec.Prior is the portfolio's own weights, as fractions in asset order", len(prior), n)
	}
	sum := 0.0
	for _, w := range prior {
		sum += w
	}
	if math.Abs(sum-1) > 1e-6 {
		return fmt.Errorf("black-litterman: the prior weights sum to %.6f, not to 1", sum)
	}
	return nil
}

// blProblem carries the derived Black-Litterman inputs into the shared
// constrained search; it is nil for every other objective.
type blProblem struct {
	lambda    float64
	posterior []float64
}

// solveBlackLitterman derives the risk aversion, the implied returns and the
// posterior means, then maximizes the mean-variance utility
//
//	mu'w - (lambda/2) w'Sigma w
//
// over the box simplex. Without bounds or limits that is a concave program
// solved exactly from the prior as the starting point, which is what makes
// the no-view answer the prior itself; with them it goes through the shared
// penalized search like every other constrained objective.
func solveBlackLitterman(returns [][]float64, spec Spec) (Result, error) {
	n := len(returns)
	if err := checkPrior(spec.Prior, n); err != nil {
		return Result{}, err
	}
	_, cov := meanCov(returns)
	lambda, err := riskAversion(spec.Prior, cov, spec.PriorReturn)
	if err != nil {
		return Result{}, err
	}
	implied := ImpliedReturns(spec.Prior, cov, lambda)
	posterior, err := Posterior(implied, cov, spec.Views)
	if err != nil {
		return Result{}, err
	}

	var res Result
	if spec.Bounded() || spec.Limits.Any() {
		bounded, err := solveConstrained(returns, spec, &blProblem{lambda: lambda, posterior: posterior})
		if err != nil {
			return Result{}, err
		}
		// The path statistics stay as measured; the headline return,
		// volatility and Sharpe are the EXPECTED ones, at the posterior.
		res = stats(bounded.Weights, posterior, cov)
		res.CAGR, res.Feasible = bounded.CAGR, bounded.Feasible
	} else {
		maxW := spec.MaxWeight
		if maxW <= 0 || maxW > 1 {
			maxW = 1
		}
		if float64(n)*maxW < 1-1e-9 {
			return Result{}, fmt.Errorf("max-weight too low: %d assets cannot sum to 100%% under a %.0f%% cap", n, maxW*100)
		}
		res = stats(maxUtility(posterior, cov, lambda, maxW, spec.Prior), posterior, cov)
	}
	res.Implied, res.Posterior, res.Lambda = implied, posterior, lambda
	return res, nil
}

// maxUtility maximizes mu'w - (lambda/2) w'Sigma w over the capped simplex.
// The objective is concave (Sigma is positive semi-definite) so projected
// descent from a single starting point reaches the global optimum; the prior
// is that starting point, and with no view its gradient already vanishes
// there, which is how the no-view answer comes back exact.
func maxUtility(mu []float64, cov [][]float64, lambda, maxW float64, prior []float64) []float64 {
	if len(mu) == 1 {
		return []float64{1}
	}
	neg := func(w []float64) float64 { return 0.5*lambda*quad(cov, w) - dot(mu, w) }
	grad := func(w []float64) []float64 {
		g := scale(matVec(cov, w), lambda)
		for i := range g {
			g[i] -= mu[i]
		}
		return g
	}
	return minimizeSimplex(neg, grad, maxW, prior)
}
