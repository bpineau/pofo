package optimize

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/metrics"
)

// penaltyScale multiplies the relative constraint violation subtracted from
// the objective. It only has to dominate any plausible objective value so the
// search always prefers a feasible point, while staying finite so the
// gradient keeps pointing back toward feasibility from outside.
const penaltyScale = 1e3

// solveConstrained is the one path for every solve the closed forms cannot
// serve: per-asset weight bounds, feasibility limits, an objective that
// depends on the whole return path (MaxReturn), or a bounded BlackLitterman
// utility. It maximizes
//
//	objective(w) − penaltyScale × Σ relative violations
//
// over the box simplex {w : Σwᵢ = 1, loᵢ ≤ wᵢ ≤ hiᵢ} by the same multi-start
// projected descent the unconstrained path-dependent objectives use. The
// violations are relative to each limit, so a volatility cap, a return floor
// and a drawdown budget stay commensurable.
//
// The weights are a good allocation, not a certified optimum: these
// objectives are non-convex and non-smooth.
// bl carries the Black-Litterman posterior and risk aversion when that is the
// objective, and is nil otherwise.
func solveConstrained(returns [][]float64, spec Spec, bl *blProblem) (Result, error) {
	n := len(returns)
	t := len(returns[0])
	lo, hi := spec.box(n)
	if err := feasible(lo, hi); err != nil {
		return Result{}, err
	}
	if spec.Objective == MaxWorst5y && t < fiveYearWindow {
		return Result{}, fmt.Errorf("max-worst-5y needs at least 5 years of common history, got %d trading days", t)
	}

	mu, cov := meanCov(returns)
	buf := make([]float64, t)
	objective, err := objectiveFn(returns, mu, cov, buf, spec, bl)
	if err != nil {
		return Result{}, err
	}
	score := func(w []float64) (float64, bool) {
		v, ok := objective(w)
		if !ok {
			return 0, false
		}
		if p := violation(returns, buf, w, spec.Limits); p > 0 {
			return v - penaltyScale*p, true
		}
		return v, true
	}

	w := maximizeBox(lo, hi, score, spec.startingPoints(n, lo, hi))
	res := seriesResult(w, returns)
	res.Feasible = violation(returns, buf, w, spec.Limits) == 0
	res.CAGR = pathCAGR(returns, buf, w)
	return res, nil
}

// objectiveFn returns the quantity to maximize for the spec's objective,
// evaluated on the blended path (or on mean/covariance where that is what the
// objective means), and whether it is defined at that point.
func objectiveFn(returns [][]float64, mu []float64, cov [][]float64, buf []float64, spec Spec, bl *blProblem) (func([]float64) (float64, bool), error) {
	switch spec.Objective {
	case BlackLitterman:
		// The mean-variance utility at the POSTERIOR means, which is what
		// makes the no-view answer the prior itself.
		return func(w []float64) (float64, bool) {
			return dot(bl.posterior, w) - 0.5*bl.lambda*quad(cov, w), true
		}, nil
	case MaxReturn:
		return func(w []float64) (float64, bool) {
			c := pathCAGR(returns, buf, w)
			return c, !math.IsNaN(c)
		}, nil
	case MaxSharpe:
		return func(w []float64) (float64, bool) {
			v := quad(cov, w)
			if v <= 0 {
				return 0, false
			}
			return dot(mu, w) / math.Sqrt(v), true
		}, nil
	case MinVolatility:
		return func(w []float64) (float64, bool) {
			v := quad(cov, w)
			if v < 0 {
				return 0, false
			}
			return -math.Sqrt(v), true // minimize: maximize the negative
		}, nil
	case MaxSortino:
		return func(w []float64) (float64, bool) {
			blend(returns, w, buf)
			s := metrics.Sortino(buf, 0)
			return s, !math.IsNaN(s)
		}, nil
	case ReturnToDrawdown:
		return func(w []float64) (float64, bool) {
			blend(returns, w, buf)
			return metrics.ReturnToMaxDrawdown(buf, 0)
		}, nil
	case MinUlcer:
		return func(w []float64) (float64, bool) {
			blend(returns, w, buf)
			u := metrics.Ulcer(buf)
			return -u, !math.IsNaN(u)
		}, nil
	case MaxWorst5y:
		return func(w []float64) (float64, bool) {
			blend(returns, w, buf)
			return metrics.WorstRollingReturn(buf, fiveYearWindow)
		}, nil
	}
	return nil, fmt.Errorf("objective %q does not accept weight bounds or limits", spec.Objective)
}

// violation measures how far a blend misses the limits, as the sum of the
// relative overshoots (0 when every limit holds). Relative, so the three
// limits can be added: 10 % over a volatility cap weighs like 10 % under a
// return floor.
func violation(returns [][]float64, buf []float64, w []float64, lim Limits) float64 {
	if !lim.Any() {
		return 0
	}
	blend(returns, w, buf)
	total := 0.0
	if lim.MaxVolatility > 0 {
		if vol := metrics.Volatility(buf); vol > lim.MaxVolatility {
			total += vol/lim.MaxVolatility - 1
		}
	}
	if lim.MinReturn > 0 {
		if c := compound(buf); c < lim.MinReturn {
			total += (lim.MinReturn - c) / math.Abs(lim.MinReturn)
		}
	}
	if lim.MaxDrawdown > 0 {
		if dd := math.Abs(pathMaxDrawdown(buf)); dd > lim.MaxDrawdown {
			total += dd/lim.MaxDrawdown - 1
		}
	}
	return total
}

// pathCAGR is the blend's annualized geometric return.
func pathCAGR(returns [][]float64, buf, w []float64) float64 {
	blend(returns, w, buf)
	return compound(buf)
}

// compound annualizes a daily return series geometrically (252 trading days),
// the CAGR of the path it describes. NaN if the path wipes out.
func compound(r []float64) float64 {
	if len(r) == 0 {
		return math.NaN()
	}
	sum := 0.0
	for _, x := range r {
		if 1+x <= 0 {
			return math.NaN()
		}
		sum += math.Log(1 + x)
	}
	return math.Expm1(sum * tradingDays / float64(len(r)))
}

// pathMaxDrawdown is the deepest peak-to-trough loss of the path, as a
// negative fraction (0 when the path never falls).
func pathMaxDrawdown(r []float64) float64 {
	v, peak, worst := 1.0, 1.0, 0.0
	for _, x := range r {
		v *= 1 + x
		if !(v > 0) {
			return -1
		}
		if v > peak {
			peak = v
		}
		if dd := v/peak - 1; dd < worst {
			worst = dd
		}
	}
	return worst
}

// feasible checks that the box can hold a portfolio at all.
func feasible(lo, hi []float64) error {
	sumLo, sumHi := 0.0, 0.0
	for i := range lo {
		if lo[i] > hi[i] {
			return fmt.Errorf("asset %d: weight floor %.1f %% above its cap %.1f %%", i+1, lo[i]*100, hi[i]*100)
		}
		sumLo += lo[i]
		sumHi += hi[i]
	}
	if sumLo > 1+1e-9 {
		return fmt.Errorf("weight floors sum to %.0f %%, above 100 %%", sumLo*100)
	}
	if sumHi < 1-1e-9 {
		return fmt.Errorf("weight caps sum to %.0f %%, below 100 %%", sumHi*100)
	}
	return nil
}

// startingPoints returns the multi-start set: the equal-weight point, each
// vertex (all the free budget on one asset), and for Black-Litterman the
// prior allocation, all projected onto the box simplex.
func (s Spec) startingPoints(n int, lo, hi []float64) [][]float64 {
	equal := make([]float64, n)
	for i := range equal {
		equal[i] = 1 / float64(n)
	}
	starts := [][]float64{projectBoxSimplex(equal, lo, hi)}
	for i := 0; i < n; i++ {
		v := make([]float64, n)
		copy(v, lo)
		v[i] = 1
		starts = append(starts, projectBoxSimplex(v, lo, hi))
	}
	// Black-Litterman starts from its own reference too: with no view the
	// prior IS the optimum, and with one it is the right neighbourhood.
	if s.Objective == BlackLitterman && len(s.Prior) == n {
		starts = append(starts, projectBoxSimplex(s.Prior, lo, hi))
	}
	return starts
}

// maximizeBox maximizes score over the box simplex by multi-start projected
// descent with a numerical gradient, the bounded twin of maximizeSimplex.
func maximizeBox(lo, hi []float64, score func([]float64) (float64, bool), starts [][]float64) []float64 {
	n := len(lo)
	if n == 1 {
		return []float64{1}
	}
	neg := func(w []float64) float64 {
		if v, ok := score(w); ok {
			return -v
		}
		return 1e6
	}
	grad := func(w []float64) []float64 {
		const h = 1e-5
		base := neg(w)
		g := make([]float64, n)
		for i := range w {
			wp := append([]float64(nil), w...)
			wp[i] += h
			g[i] = (neg(wp) - base) / h
		}
		return g
	}
	project := func(x []float64) []float64 { return projectBoxSimplex(x, lo, hi) }
	var best []float64
	bestVal := math.Inf(1)
	for _, s := range starts {
		w := projectedGradient(neg, grad, project, project(s))
		if v := neg(w); v < bestVal {
			bestVal, best = v, w
		}
	}
	return best
}

// projectBoxSimplex returns the Euclidean projection of v onto
// {w : Σwᵢ = 1, loᵢ ≤ wᵢ ≤ hiᵢ}. Like projectCappedSimplex it solves
// Σ clamp(vᵢ−λ, loᵢ, hiᵢ) = 1 for λ by bisection, the sum being
// monotonically non-increasing in λ; per-asset bounds change nothing but the
// clamp. The caller must have checked feasibility (see feasible).
func projectBoxSimplex(v, lo, hi []float64) []float64 {
	sumClamp := func(lam float64) float64 {
		s := 0.0
		for i, x := range v {
			s += clamp(x-lam, lo[i], hi[i])
		}
		return s
	}
	loLam, hiLam := -1.0, 1.0
	for sumClamp(loLam) < 1 {
		loLam -= 1
	}
	for sumClamp(hiLam) > 1 {
		hiLam += 1
	}
	for i := 0; i < 100; i++ {
		mid := (loLam + hiLam) / 2
		if sumClamp(mid) > 1 {
			loLam = mid
		} else {
			hiLam = mid
		}
	}
	lam := (loLam + hiLam) / 2
	w := make([]float64, len(v))
	for i, x := range v {
		w[i] = clamp(x-lam, lo[i], hi[i])
	}
	return w
}

// PathStats returns the annualized geometric return (CAGR), the annualized
// volatility and the deepest drawdown (a negative fraction) of the weighted
// blend of returns. It is what the solver's limits are measured on, exposed
// so a caller can report the same quantities on another window (a training
// slice and the stretch that follows it, say).
func PathStats(returns [][]float64, w []float64) (cagr, vol, maxDrawdown float64) {
	if len(returns) == 0 || len(returns[0]) == 0 {
		return math.NaN(), math.NaN(), math.NaN()
	}
	buf := make([]float64, len(returns[0]))
	blend(returns, w, buf)
	return compound(buf), metrics.Volatility(buf), pathMaxDrawdown(buf)
}
