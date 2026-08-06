package optimize

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const tradingDays = 252

// Objective is the quantity the optimizer targets.
type Objective string

// Supported objectives.
const (
	MaxSharpe     Objective = "max-sharpe"
	MinVolatility Objective = "min-volatility"
	RiskParity    Objective = "risk-parity"
)

// Spec describes an optimization: an objective and its constraints.
type Spec struct {
	Objective Objective
	// MaxWeight caps each asset's weight, as a fraction in (0,1]; 0 means
	// no cap. Ignored for RiskParity.
	MaxWeight float64
}

// Result is an optimized allocation and its in-sample statistics.
type Result struct {
	Weights    []float64 // one per asset, in input order, summing to 1
	Return     float64   // expected annualized return
	Volatility float64   // expected annualized volatility
	Sharpe     float64   // expected annualized Sharpe ratio (risk-free 0)
}

// ParseSpec reads a "#meta optimize:" value: an objective optionally
// followed by comma-separated constraints, e.g. "max-sharpe,max-weight:40"
// (the cap is a percentage). Whitespace is not allowed between tokens.
func ParseSpec(s string) (Spec, error) {
	tokens := strings.Split(s, ",")
	obj := Objective(strings.ToLower(strings.TrimSpace(tokens[0])))
	switch obj {
	case MaxSharpe, MinVolatility, RiskParity:
	default:
		return Spec{}, fmt.Errorf("unknown objective %q (max-sharpe, min-volatility or risk-parity)", tokens[0])
	}
	spec := Spec{Objective: obj}
	for _, tok := range tokens[1:] {
		key, val, ok := strings.Cut(strings.TrimSpace(tok), ":")
		if !ok {
			return Spec{}, fmt.Errorf("invalid constraint %q (expected key:value)", tok)
		}
		switch strings.ToLower(key) {
		case "max-weight":
			pct, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(val), "%"), 64)
			if err != nil || pct <= 0 || pct > 100 {
				return Spec{}, fmt.Errorf("max-weight: %q is not a percentage in (0,100]", val)
			}
			spec.MaxWeight = pct / 100
		default:
			return Spec{}, fmt.Errorf("unknown constraint %q", key)
		}
	}
	return spec, nil
}

// Solve returns the weights optimizing spec.Objective over returns, the
// aligned daily simple returns of each asset (returns[i] is asset i's
// series; every slice must have the same length, at least 2).
func Solve(returns [][]float64, spec Spec) (Result, error) {
	n := len(returns)
	if n == 0 {
		return Result{}, fmt.Errorf("no assets to optimize")
	}
	t := len(returns[0])
	if t < 2 {
		return Result{}, fmt.Errorf("need at least 2 observations, got %d", t)
	}
	for i, r := range returns {
		if len(r) != t {
			return Result{}, fmt.Errorf("asset %d has %d observations, expected %d", i, len(r), t)
		}
	}
	mu, cov := meanCov(returns)
	return solve(mu, cov, spec)
}

// meanCov returns the annualized mean vector and covariance matrix of the
// per-asset daily returns. Covariance uses the sample estimator (T−1).
func meanCov(returns [][]float64) (mu []float64, cov [][]float64) {
	n := len(returns)
	t := len(returns[0])
	mean := make([]float64, n)
	for i, r := range returns {
		s := 0.0
		for _, x := range r {
			s += x
		}
		mean[i] = s / float64(t)
	}
	cov = make([][]float64, n)
	for i := range cov {
		cov[i] = make([]float64, n)
	}
	denom := float64(t - 1)
	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			s := 0.0
			for k := 0; k < t; k++ {
				s += (returns[i][k] - mean[i]) * (returns[j][k] - mean[j])
			}
			c := s / denom * tradingDays
			cov[i][j], cov[j][i] = c, c
		}
	}
	mu = make([]float64, n)
	for i := range mean {
		mu[i] = mean[i] * tradingDays
	}
	return mu, cov
}

// solve runs the chosen objective over an already-annualized mean vector
// and covariance matrix.
func solve(mu []float64, cov [][]float64, spec Spec) (Result, error) {
	n := len(mu)
	if n == 1 {
		return stats([]float64{1}, mu, cov), nil
	}
	maxW := spec.MaxWeight
	if maxW <= 0 || maxW > 1 {
		maxW = 1
	}

	var w []float64
	switch spec.Objective {
	case MinVolatility:
		if float64(n)*maxW < 1-1e-9 {
			return Result{}, fmt.Errorf("max-weight too low: %d assets cannot sum to 100%% under a %.0f%% cap", n, maxW*100)
		}
		w = minimizeSimplex(
			func(x []float64) float64 { return quad(cov, x) },
			func(x []float64) []float64 { return scale(matVec(cov, x), 2) },
			maxW, equalStart(n, maxW))
	case MaxSharpe:
		if float64(n)*maxW < 1-1e-9 {
			return Result{}, fmt.Errorf("max-weight too low: %d assets cannot sum to 100%% under a %.0f%% cap", n, maxW*100)
		}
		w = maxSharpe(mu, cov, maxW)
	case RiskParity:
		w = riskParity(cov)
	default:
		return Result{}, fmt.Errorf("unknown objective %q", spec.Objective)
	}
	return stats(w, mu, cov), nil
}

// maxSharpe maximizes the Sharpe ratio (risk-free 0) over the capped
// simplex. The objective is not convex, so several deterministic starting
// points are tried and the best feasible Sharpe is kept.
func maxSharpe(mu []float64, cov [][]float64, maxW float64) []float64 {
	n := len(mu)
	neg := func(x []float64) float64 {
		v := quad(cov, x)
		if v <= 0 {
			return 0
		}
		return -dot(mu, x) / math.Sqrt(v)
	}
	grad := func(x []float64) []float64 {
		q := quad(cov, x)
		if q <= 0 {
			return make([]float64, n)
		}
		sx := matVec(cov, x)
		num := dot(mu, x)
		g := make([]float64, n)
		inv := 1 / math.Sqrt(q)
		for i := range g {
			g[i] = -(inv * (mu[i] - num/q*sx[i]))
		}
		return g
	}

	starts := [][]float64{equalStart(n, maxW)}
	// Each single asset (clamped to the cap, remainder spread evenly).
	for i := 0; i < n; i++ {
		s := make([]float64, n)
		s[i] = 1
		starts = append(starts, projectCappedSimplex(s, maxW))
	}
	// Inverse-variance start, a decent guess for the tangency portfolio.
	iv := make([]float64, n)
	for i := range iv {
		if cov[i][i] > 0 {
			iv[i] = 1 / cov[i][i]
		}
	}
	starts = append(starts, projectCappedSimplex(iv, maxW))

	var best []float64
	bestVal := math.Inf(1)
	for _, s := range starts {
		w := minimizeSimplex(neg, grad, maxW, s)
		if v := neg(w); v < bestVal {
			bestVal, best = v, w
		}
	}
	return best
}

// riskParity returns the long-only portfolio whose assets contribute equal
// shares of risk. It minimizes the convex Spinu objective
// ½·wᵀΣw − (1/n)·Σ ln(wᵢ) over w>0, then normalizes the weights to sum to 1;
// at the minimum every wᵢ·(Σw)ᵢ is equal, i.e. equal risk contributions.
func riskParity(cov [][]float64) []float64 {
	n := len(cov)
	c := 1.0 / float64(n)
	f := func(x []float64) float64 {
		v := 0.5 * quad(cov, x)
		for _, xi := range x {
			v -= c * math.Log(xi)
		}
		return v
	}
	grad := func(x []float64) []float64 {
		sx := matVec(cov, x)
		g := make([]float64, n)
		for i := range g {
			g[i] = sx[i] - c/x[i]
		}
		return g
	}
	// Inverse-volatility start, projected onto the positive orthant.
	x := make([]float64, n)
	for i := range x {
		if cov[i][i] > 0 {
			x[i] = 1 / math.Sqrt(cov[i][i])
		} else {
			x[i] = 1
		}
	}
	x = minimizeOrthant(f, grad, x)
	return normalize(x)
}

// minimizeSimplex minimizes f (gradient grad) over the capped simplex by
// projected gradient descent with backtracking line search.
func minimizeSimplex(f func([]float64) float64, grad func([]float64) []float64, maxW float64, x0 []float64) []float64 {
	project := func(x []float64) []float64 { return projectCappedSimplex(x, maxW) }
	return projectedGradient(f, grad, project, project(x0))
}

// minimizeOrthant minimizes f (gradient grad) over the positive orthant.
func minimizeOrthant(f func([]float64) float64, grad func([]float64) []float64, x0 []float64) []float64 {
	const eps = 1e-12
	project := func(x []float64) []float64 {
		y := make([]float64, len(x))
		for i, v := range x {
			y[i] = math.Max(v, eps)
		}
		return y
	}
	return projectedGradient(f, grad, project, project(x0))
}

// projectedGradient runs projected gradient descent: at each step it moves
// against the gradient, projects back onto the feasible set, and shrinks
// the step until the objective decreases. It stops when the iterate barely
// moves or the step vanishes.
func projectedGradient(f func([]float64) float64, grad func([]float64) []float64, project func([]float64) []float64, x []float64) []float64 {
	const (
		maxIters = 5000
		tol      = 1e-12
	)
	fx := f(x)
	step := 1.0
	for iter := 0; iter < maxIters; iter++ {
		g := grad(x)
		moved := false
		for bt := 0; bt < 60; bt++ {
			y := project(axpy(x, -step, g))
			if fy := f(y); fy < fx-1e-15 {
				d := dist2(x, y)
				x, fx = y, fy
				step *= 1.3
				moved = true
				if d < tol {
					return x
				}
				break
			}
			step *= 0.5
		}
		if !moved {
			return x // no feasible decrease: converged
		}
	}
	return x
}

// projectCappedSimplex returns the Euclidean projection of v onto
// {w : Σwᵢ = 1, 0 ≤ wᵢ ≤ cap}. It solves Σ clamp(vᵢ−λ, 0, cap) = 1 for the
// threshold λ by bisection (the sum is monotonically non-increasing in λ).
func projectCappedSimplex(v []float64, maxW float64) []float64 {
	sumClamp := func(lam float64) float64 {
		s := 0.0
		for _, x := range v {
			s += clamp(x-lam, 0, maxW)
		}
		return s
	}
	lo, hi := -1.0, 1.0
	for sumClamp(lo) < 1 {
		lo -= 1
	}
	for sumClamp(hi) > 1 {
		hi += 1
	}
	for i := 0; i < 100; i++ {
		mid := (lo + hi) / 2
		if sumClamp(mid) > 1 {
			lo = mid
		} else {
			hi = mid
		}
	}
	lam := (lo + hi) / 2
	w := make([]float64, len(v))
	for i, x := range v {
		w[i] = clamp(x-lam, 0, maxW)
	}
	return w
}

// stats packages weights with their in-sample return, volatility and Sharpe.
func stats(w, mu []float64, cov [][]float64) Result {
	r := Result{Weights: w, Return: dot(mu, w)}
	if v := quad(cov, w); v > 0 {
		r.Volatility = math.Sqrt(v)
		r.Sharpe = r.Return / r.Volatility
	}
	return r
}

func equalStart(n int, maxW float64) []float64 {
	x := make([]float64, n)
	for i := range x {
		x[i] = 1.0 / float64(n)
	}
	return projectCappedSimplex(x, maxW)
}

func normalize(x []float64) []float64 {
	s := 0.0
	for _, v := range x {
		s += v
	}
	out := make([]float64, len(x))
	for i, v := range x {
		out[i] = v / s
	}
	return out
}

func matVec(m [][]float64, x []float64) []float64 {
	y := make([]float64, len(m))
	for i := range m {
		s := 0.0
		for j, v := range x {
			s += m[i][j] * v
		}
		y[i] = s
	}
	return y
}

func quad(m [][]float64, x []float64) float64 { return dot(x, matVec(m, x)) }

func dot(a, b []float64) float64 {
	s := 0.0
	for i := range a {
		s += a[i] * b[i]
	}
	return s
}

func scale(x []float64, c float64) []float64 {
	y := make([]float64, len(x))
	for i, v := range x {
		y[i] = c * v
	}
	return y
}

// axpy returns x + a·g.
func axpy(x []float64, a float64, g []float64) []float64 {
	y := make([]float64, len(x))
	for i := range x {
		y[i] = x[i] + a*g[i]
	}
	return y
}

func dist2(a, b []float64) float64 {
	s := 0.0
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return s
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
