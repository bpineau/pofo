package optimize

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const tradingDays = 252

// Objective is the quantity the optimizer targets.
type Objective string

// Supported objectives.
const (
	MaxSharpe     Objective = "max-sharpe"
	MinVolatility Objective = "min-volatility"
	RiskParity    Objective = "risk-parity"
	// MaxSortino maximizes the portfolio's own annualized Sortino ratio
	// (return over downside deviation), rewarding assets that cut the
	// downside, i.e. non-correlation and positive skew.
	MaxSortino Objective = "max-sortino"
	// ReturnToDrawdown maximizes the portfolio's own return-to-maximum-
	// drawdown (a Calmar-style ratio), rewarding shallow drawdowns.
	ReturnToDrawdown Objective = "return-to-drawdown"
	// MinUlcer minimizes the portfolio's Ulcer Index (root-mean-square
	// drawdown), which grows with both the depth and the duration of
	// underwater periods: the smooth way to shorten and shallow the stretches
	// that are hard to sit through.
	MinUlcer Objective = "min-ulcer"
	// MaxWorst5y maximizes the worst rolling five-year return, the robust
	// worst-case medium-term outcome (measured over 5*252 trading days).
	MaxWorst5y Objective = "max-worst-5y"
	// CWARP maximizes the portfolio's Cole Wins Above Replacement Portfolio
	// score against a replacement/benchmark series; solved by SolveCWARP,
	// which needs that extra series, not Solve.
	CWARP Objective = "cwarp"
	// MaxReturn maximizes the portfolio's own CAGR (geometric, over the
	// blended daily path). Alone it is degenerate, since the whole budget
	// goes to the single best-performing asset; it earns its keep under a
	// constraint (Spec.Limits), where it is the frontier point: the most
	// return reachable without exceeding a volatility or drawdown budget.
	MaxReturn Objective = "max-return"
)

// Limits are the feasibility constraints a solve must respect, on top of the
// weight bounds. A zero field means "no limit": a 0 % volatility cap, a 0 %
// return floor and a 0 % drawdown budget are all meaningless or infeasible,
// so nothing is lost by using zero as the sentinel. Every limit is measured
// on the blended daily return path, so MaxVolatility reads exactly like the
// report's "Volatility (annualized)" row and MinReturn like its CAGR row.
type Limits struct {
	// MaxVolatility caps the annualized volatility, as a fraction
	// (0.095 = 9.5 %/yr).
	MaxVolatility float64
	// MinReturn floors the annualized GEOMETRIC return (CAGR), as a
	// fraction (0.105 = 10.5 %/yr).
	MinReturn float64
	// MaxDrawdown caps the deepest peak-to-trough loss, as a POSITIVE
	// fraction (0.20 tolerates a -20 % drawdown, not a deeper one).
	MaxDrawdown float64
}

// Any reports whether at least one limit applies.
func (l Limits) Any() bool {
	return l.MaxVolatility > 0 || l.MinReturn > 0 || l.MaxDrawdown > 0
}

// Window is the optional fitting period parsed from a "train:" constraint.
// A zero Start or End leaves that end open.
//
// Solve IGNORES IT: this package is date-free, and slicing the returns to the
// training window is the caller's job (pkg/compare does it, then measures the
// resulting weights over the full window so the report's numbers are
// out-of-sample). It travels in Spec only because it arrives in the same
// "#meta optimize:" directive and one parser is better than two.
type Window struct {
	Start, End time.Time
}

// IsZero reports whether the window constrains nothing.
func (w Window) IsZero() bool { return w.Start.IsZero() && w.End.IsZero() }

// Contains reports whether t falls in the window (bounds inclusive).
func (w Window) Contains(t time.Time) bool {
	if !w.Start.IsZero() && t.Before(w.Start) {
		return false
	}
	return w.End.IsZero() || !t.After(w.End)
}

// String renders the window as "start→end", with "…" for an open end.
func (w Window) String() string {
	f := func(t time.Time) string {
		if t.IsZero() {
			return "…"
		}
		return t.Format("2006-01-02")
	}
	return f(w.Start) + "→" + f(w.End)
}

// Spec describes an optimization: an objective and its constraints.
type Spec struct {
	Objective Objective
	// MaxWeight caps each asset's weight, as a fraction in (0,1]; 0 means
	// no cap. Ignored for RiskParity.
	MaxWeight float64
	// MinWeight floors each asset's weight, as a fraction in [0,1). It
	// keeps every line in the book instead of letting the search drop the
	// ones that only pay off out of sample. Ignored for RiskParity.
	MinWeight float64
	// Bounds holds per-asset weight ranges as [low, high] fractions, keyed
	// by the identifier written in the portfolio file. It is what ParseSpec
	// can produce from text, and this package never reads it: the caller
	// resolves it against its own assets with Resolve, which fills the
	// Lower/Upper the solver uses. Identifier resolution belongs to
	// pkg/marketdata, not here.
	Bounds map[string][2]float64
	// Lower and Upper are the resolved per-asset bounds, in the same order
	// as the returns passed to Solve, as fractions. A nil slice (or a NaN
	// entry) falls back to MinWeight/MaxWeight for that asset.
	Lower, Upper []float64
	// Limits are the feasibility constraints (volatility cap, return floor,
	// drawdown budget). Any limit routes the solve through the penalized
	// path search, whatever the objective.
	Limits Limits
	// Train is the fitting window parsed from "train:", ignored here and
	// applied by the caller: see Window.
	Train Window
}

// Resolve turns Spec.Bounds (keyed by identifier) into the Lower/Upper slices
// the solver reads, for assets whose identifiers are ids[i] (several spellings
// per asset are accepted: pass the written id and the resolved symbol). An id
// in Bounds that matches no asset is an error, since a typo would otherwise
// read as "no bound at all".
func (s *Spec) Resolve(ids [][]string) error {
	n := len(ids)
	s.Lower, s.Upper = make([]float64, n), make([]float64, n)
	for i := range s.Lower {
		s.Lower[i], s.Upper[i] = math.NaN(), math.NaN()
	}
	for key, b := range s.Bounds {
		found := -1
		for i, names := range ids {
			for _, name := range names {
				if strings.EqualFold(strings.TrimSpace(name), key) {
					found = i
					break
				}
			}
			if found >= 0 {
				break
			}
		}
		if found < 0 {
			return fmt.Errorf("bounds: %q matches no holding", key)
		}
		s.Lower[found], s.Upper[found] = b[0], b[1]
	}
	return nil
}

// box returns the effective per-asset bounds for n assets: the resolved
// Lower/Upper where set, the scalar MinWeight/MaxWeight elsewhere.
func (s Spec) box(n int) (lo, hi []float64) {
	lo, hi = make([]float64, n), make([]float64, n)
	defLo, defHi := s.MinWeight, s.MaxWeight
	if defHi <= 0 || defHi > 1 {
		defHi = 1
	}
	for i := range lo {
		lo[i], hi[i] = defLo, defHi
		if i < len(s.Lower) && !math.IsNaN(s.Lower[i]) {
			lo[i] = s.Lower[i]
		}
		if i < len(s.Upper) && !math.IsNaN(s.Upper[i]) {
			hi[i] = s.Upper[i]
		}
	}
	return lo, hi
}

// Bounded reports whether the spec constrains weights per asset (a floor or
// a resolved per-asset bound), beyond the plain MaxWeight cap the closed-form
// solvers already handle.
func (s Spec) Bounded() bool {
	if s.MinWeight > 0 {
		return true
	}
	for _, v := range s.Lower {
		if !math.IsNaN(v) && v > 0 {
			return true
		}
	}
	for _, v := range s.Upper {
		if !math.IsNaN(v) {
			return true
		}
	}
	return false
}

// Result is an optimized allocation and its in-sample statistics.
type Result struct {
	Weights    []float64 // one per asset, in input order, summing to 1
	Return     float64   // expected annualized return
	Volatility float64   // expected annualized volatility
	Sharpe     float64   // expected annualized Sharpe ratio (risk-free 0)

	// Achieved values of the path-dependent objectives, set by the solver
	// that targets them (zero otherwise).
	Sortino       float64 // annualized Sortino ratio (MaxSortino)
	ReturnToMaxDD float64 // return-to-maximum-drawdown (ReturnToDrawdown)
	Ulcer         float64 // Ulcer Index in percent points (MinUlcer)
	Worst5y       float64 // worst rolling five-year return (MaxWorst5y)
	CWARP         float64 // CWARP vs the replacement (SolveCWARP)

	// CAGR is the annualized GEOMETRIC return of the weighted path, set by
	// the constrained solver (Return, above, is the arithmetic mean the
	// mean/covariance solvers work in, and runs higher by the volatility
	// drag). Zero when the solve did not go through that path.
	CAGR float64
	// Feasible reports whether the returned weights satisfy Spec.Limits.
	// False means the constraints could not be met and the weights are the
	// least-violating point found, not an answer.
	Feasible bool
}

// ParseSpec reads a "#meta optimize:" value: an objective optionally
// followed by comma-separated constraints, e.g. "max-sharpe,max-weight:40"
// (the cap is a percentage). Whitespace is not allowed between tokens.
func ParseSpec(s string) (Spec, error) {
	tokens := strings.Split(s, ",")
	obj := Objective(strings.ToLower(strings.TrimSpace(tokens[0])))
	switch obj {
	case MaxSharpe, MinVolatility, RiskParity, MaxSortino, ReturnToDrawdown, MinUlcer, MaxWorst5y, CWARP, MaxReturn:
	default:
		return Spec{}, fmt.Errorf("unknown objective %q (max-sharpe, min-volatility, max-return, risk-parity, max-sortino, return-to-drawdown, min-ulcer, max-worst-5y or cwarp)", tokens[0])
	}
	spec := Spec{Objective: obj}
	for _, tok := range tokens[1:] {
		key, val, ok := strings.Cut(strings.TrimSpace(tok), ":")
		if !ok {
			return Spec{}, fmt.Errorf("invalid constraint %q (expected key:value)", tok)
		}
		switch strings.ToLower(key) {
		case "max-weight":
			pct, err := percent(val, 0, 100, false)
			if err != nil {
				return Spec{}, fmt.Errorf("max-weight: %w", err)
			}
			spec.MaxWeight = pct / 100
		case "min-weight":
			pct, err := percent(val, 0, 100, true)
			if err != nil {
				return Spec{}, fmt.Errorf("min-weight: %w", err)
			}
			spec.MinWeight = pct / 100
		case "max-vol", "max-volatility":
			pct, err := percent(val, 0, 200, false)
			if err != nil {
				return Spec{}, fmt.Errorf("max-vol: %w", err)
			}
			spec.Limits.MaxVolatility = pct / 100
		case "min-return":
			pct, err := percent(val, -100, 200, false)
			if err != nil {
				return Spec{}, fmt.Errorf("min-return: %w", err)
			}
			spec.Limits.MinReturn = pct / 100
		case "max-drawdown", "max-dd":
			pct, err := percent(strings.TrimPrefix(strings.TrimSpace(val), "-"), 0, 100, false)
			if err != nil {
				return Spec{}, fmt.Errorf("max-drawdown: %w", err)
			}
			spec.Limits.MaxDrawdown = pct / 100
		case "bounds":
			id, rng, ok := strings.Cut(strings.TrimSpace(val), ":")
			if !ok {
				return Spec{}, fmt.Errorf("bounds: %q is not ID:LOW-HIGH", val)
			}
			lo, hi, err := parseRange(rng)
			if err != nil {
				return Spec{}, fmt.Errorf("bounds %s: %w", id, err)
			}
			if spec.Bounds == nil {
				spec.Bounds = map[string][2]float64{}
			}
			spec.Bounds[strings.TrimSpace(id)] = [2]float64{lo, hi}
		case "train":
			w, err := parseWindow(val)
			if err != nil {
				return Spec{}, fmt.Errorf("train: %w", err)
			}
			spec.Train = w
		default:
			return Spec{}, fmt.Errorf("unknown constraint %q", key)
		}
	}
	if spec.MinWeight > 0 && spec.MaxWeight > 0 && spec.MinWeight > spec.MaxWeight {
		return Spec{}, fmt.Errorf("min-weight (%.0f %%) above max-weight (%.0f %%)", spec.MinWeight*100, spec.MaxWeight*100)
	}
	return spec, nil
}

// percent parses a percentage, with an optional "%" suffix, inside
// (lo, hi] (or [lo, hi] when zeroOK, i.e. when 0 is a meaningful value).
func percent(val string, lo, hi float64, zeroOK bool) (float64, error) {
	pct, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(val), "%"), 64)
	switch {
	case err != nil:
		return 0, fmt.Errorf("%q is not a number", val)
	case pct > hi, pct < lo, pct == lo && !zeroOK:
		return 0, fmt.Errorf("%q is not a percentage in %s%.0f,%.0f]", val, map[bool]string{true: "[", false: "("}[zeroOK], lo, hi)
	}
	return pct, nil
}

// parseRange reads a "LOW-HIGH" weight range in percent, either end
// omittable ("15-", "-30"), and returns the two fractions (NaN when open).
func parseRange(s string) (lo, hi float64, err error) {
	loStr, hiStr, ok := strings.Cut(strings.TrimSpace(s), "-")
	if !ok {
		return 0, 0, fmt.Errorf("%q is not LOW-HIGH", s)
	}
	lo, hi = math.NaN(), math.NaN()
	if strings.TrimSpace(loStr) != "" {
		v, err := percent(loStr, 0, 100, true)
		if err != nil {
			return 0, 0, err
		}
		lo = v / 100
	}
	if strings.TrimSpace(hiStr) != "" {
		v, err := percent(hiStr, 0, 100, false)
		if err != nil {
			return 0, 0, err
		}
		hi = v / 100
	}
	if !math.IsNaN(lo) && !math.IsNaN(hi) && lo > hi {
		return 0, 0, fmt.Errorf("%q has its low end above its high end", s)
	}
	return lo, hi, nil
}

// parseWindow reads "START..END", each end a year (2015), a full date
// (2015-12-31) or empty, and returns the calendar window it denotes (a bare
// year means the whole year: 2015 as a start is 2015-01-01, as an end
// 2015-12-31).
func parseWindow(s string) (Window, error) {
	startStr, endStr, ok := strings.Cut(strings.TrimSpace(s), "..")
	if !ok {
		return Window{}, fmt.Errorf("%q is not START..END", s)
	}
	var w Window
	var err error
	if w.Start, err = parseBound(startStr, false); err != nil {
		return Window{}, err
	}
	if w.End, err = parseBound(endStr, true); err != nil {
		return Window{}, err
	}
	if w.IsZero() {
		return Window{}, fmt.Errorf("%q leaves both ends open", s)
	}
	if !w.Start.IsZero() && !w.End.IsZero() && !w.Start.Before(w.End) {
		return Window{}, fmt.Errorf("%q starts after it ends", s)
	}
	return w, nil
}

// parseBound reads one end of a train window; end selects the last day of a
// bare year rather than its first.
func parseBound(s string, end bool) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006", s); err == nil {
		if end {
			return t.AddDate(1, 0, -1), nil
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("%q is neither a year nor a YYYY-MM-DD date", s)
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
	// Bounds and limits route every objective through the one penalized
	// path search; without them the closed forms answer exactly as before.
	if spec.Objective != RiskParity && (spec.Bounded() || spec.Limits.Any() || spec.Objective == MaxReturn) {
		return solveConstrained(returns, spec)
	}
	switch spec.Objective {
	case MaxSortino, ReturnToDrawdown, MinUlcer, MaxWorst5y:
		return solveSeries(returns, spec)
	}
	mu, cov := meanCov(returns)
	return solve(mu, cov, spec)
}

// blend writes the weighted daily returns of the portfolio into out (len equal
// to each asset's return series).
func blend(returns [][]float64, w, out []float64) {
	for k := range out {
		s := 0.0
		for i := range w {
			s += w[i] * returns[i][k]
		}
		out[k] = s
	}
}

// maximizeSimplex maximizes score over the capped simplex by multi-start
// projected descent with a numerical gradient. score reports the value to
// maximize and whether it is defined; undefined points get a large penalty so
// the search steers away from them. It is the shared engine for the objectives
// that are non-convex and non-smooth (Sortino, return-to-drawdown, CWARP),
// mirroring maxSharpe's multi-start pattern.
func maximizeSimplex(n int, maxW float64, score func([]float64) (float64, bool)) []float64 {
	neg := func(w []float64) float64 {
		if v, ok := score(w); ok {
			return -v
		}
		return 1e6
	}
	if n == 1 {
		return []float64{1}
	}
	// Forward-difference gradient: projected descent only needs a descent
	// direction, and the projection restores the simplex after each step.
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
	starts := [][]float64{equalStart(n, maxW)}
	for i := 0; i < n; i++ {
		s := make([]float64, n)
		s[i] = 1
		starts = append(starts, projectCappedSimplex(s, maxW))
	}
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
// simplex. The ratio is not concave in the weights, so several deterministic
// starting points are tried and the best feasible Sharpe is kept. One of them
// is the EXACT uncapped tangency portfolio (tangencyQP): without a cap that
// seed is already the global optimum, and with one it starts the search in
// the right neighbourhood.
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
	// The exact uncapped tangency portfolio: the global optimum itself when
	// no cap binds, and the best available seed when one does.
	if qp, ok := tangencyQP(mu, cov); ok {
		starts = append(starts, projectCappedSimplex(qp, maxW))
	}

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

// tangencyQP returns the exact long-only tangency portfolio (risk-free 0),
// reporting ok=false when no asset has a positive mean, which leaves the
// transformation below without a feasible point.
//
// Maximizing muᵀw/√(wᵀΣw) over the simplex is not a concave program, but
// Schaible's transformation makes it one: substituting y = w/(muᵀw) turns it
// into
//
//	minimize yᵀΣy   subject to   muᵀy = 1, y ≥ 0
//
// which is convex (Σ is positive semi-definite over a polyhedron), so
// projected gradient descent on it reaches the GLOBAL minimum; w = y/Σyᵢ maps
// the solution back. See Cornuejols & Tutuncu, "Optimization Methods in
// Finance", chapter 8.
func tangencyQP(mu []float64, cov [][]float64) ([]float64, bool) {
	best := math.Inf(-1)
	for _, m := range mu {
		best = math.Max(best, m)
	}
	if best <= 0 {
		return nil, false // muᵀy = 1 unreachable with y ≥ 0
	}
	y := projectedGradient(
		func(y []float64) float64 { return quad(cov, y) },
		func(y []float64) []float64 { return scale(matVec(cov, y), 2) },
		func(y []float64) []float64 { return projectMuPlane(y, mu) },
		projectMuPlane(equalStart(len(mu), 1), mu))
	sum := 0.0
	for _, v := range y {
		sum += v
	}
	if sum <= 0 || math.IsNaN(sum) || math.IsInf(sum, 0) {
		return nil, false
	}
	w := make([]float64, len(y))
	for i, v := range y {
		w[i] = v / sum
	}
	return w, true
}

// projectMuPlane returns the Euclidean projection of v onto
// {y : muᵀy = 1, y ≥ 0}, which is yᵢ = max(vᵢ + λ·muᵢ, 0) for the λ solving
// muᵀy = 1. That sum is non-decreasing in λ (its slope is the sum of the
// muᵢ² still off the floor), so a bisection finds λ. The caller must have
// checked that some muᵢ is positive, otherwise no feasible point exists.
func projectMuPlane(v, mu []float64) []float64 {
	at := func(lam float64) []float64 {
		y := make([]float64, len(v))
		for i := range v {
			y[i] = math.Max(v[i]+lam*mu[i], 0)
		}
		return y
	}
	sum := func(lam float64) float64 { return dot(mu, at(lam)) }
	lo, hi := -1.0, 1.0
	for sum(lo) > 1 {
		lo *= 2
	}
	for sum(hi) < 1 {
		hi *= 2
	}
	for i := 0; i < 200; i++ {
		mid := (lo + hi) / 2
		if sum(mid) < 1 {
			lo = mid
		} else {
			hi = mid
		}
	}
	return at((lo + hi) / 2)
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
