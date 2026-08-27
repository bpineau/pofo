package optimize

import (
	"math"
	"strings"
	"testing"
)

func approx(t *testing.T, got, want, tol float64, what string) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s = %.5f, want %.5f (±%.0e)", what, got, want, tol)
	}
}

// riskContribs returns each asset's share of total risk, RCᵢ = wᵢ·(Σw)ᵢ
// normalized to sum to 1.
func riskContribs(w []float64, cov [][]float64) []float64 {
	sx := matVec(cov, w)
	total := dot(w, sx)
	rc := make([]float64, len(w))
	for i := range w {
		rc[i] = w[i] * sx[i] / total
	}
	return rc
}

// For a diagonal covariance the optima have closed forms.
func TestSolveDiagonalClosedForms(t *testing.T) {
	mu := []float64{0.10, 0.05}
	cov := [][]float64{{0.04, 0}, {0, 0.01}} // σ = 20 %, 10 %

	// Min-variance: wᵢ ∝ 1/σᵢ²  → 25 : 100 → 0.2 : 0.8.
	minv, err := solve(mu, cov, Spec{Objective: MinVolatility})
	if err != nil {
		t.Fatal(err)
	}
	approx(t, minv.Weights[0], 0.2, 2e-3, "min-vol w0")
	approx(t, minv.Weights[1], 0.8, 2e-3, "min-vol w1")

	// Max-Sharpe (tangency, rf=0): wᵢ ∝ μᵢ/σᵢ²  → 2.5 : 5 → 1/3 : 2/3.
	ms, err := solve(mu, cov, Spec{Objective: MaxSharpe})
	if err != nil {
		t.Fatal(err)
	}
	approx(t, ms.Weights[0], 1.0/3, 3e-3, "max-sharpe w0")
	approx(t, ms.Weights[1], 2.0/3, 3e-3, "max-sharpe w1")

	// Risk parity: wᵢ ∝ 1/σᵢ → 5 : 10 → 1/3 : 2/3, equal risk contributions.
	rp, err := solve(mu, cov, Spec{Objective: RiskParity})
	if err != nil {
		t.Fatal(err)
	}
	approx(t, rp.Weights[0], 1.0/3, 3e-3, "risk-parity w0")
	rc := riskContribs(rp.Weights, cov)
	approx(t, rc[0], 0.5, 2e-3, "risk-parity RC0")
	approx(t, rc[1], 0.5, 2e-3, "risk-parity RC1")
}

// Min-variance with correlation, against the two-asset analytic weight.
func TestSolveCorrelatedMinVol(t *testing.T) {
	s1, s2, rho := 0.2, 0.1, 0.3
	cov := [][]float64{{s1 * s1, rho * s1 * s2}, {rho * s1 * s2, s2 * s2}}
	mu := []float64{0.08, 0.06}
	w1 := (s2*s2 - rho*s1*s2) / (s1*s1 + s2*s2 - 2*rho*s1*s2) // 0.10526…
	r, err := solve(mu, cov, Spec{Objective: MinVolatility})
	if err != nil {
		t.Fatal(err)
	}
	approx(t, r.Weights[0], w1, 2e-3, "correlated min-vol w0")
	approx(t, r.Weights[0]+r.Weights[1], 1, 1e-9, "weights sum")
}

// Risk parity equalizes risk contributions even with correlation.
func TestRiskParityCorrelated(t *testing.T) {
	cov := [][]float64{
		{0.040, 0.012, 0.000},
		{0.012, 0.025, 0.006},
		{0.000, 0.006, 0.010},
	}
	mu := []float64{0.07, 0.06, 0.04}
	r, err := solve(mu, cov, Spec{Objective: RiskParity})
	if err != nil {
		t.Fatal(err)
	}
	rc := riskContribs(r.Weights, cov)
	for i := range rc {
		approx(t, rc[i], 1.0/3, 3e-3, "risk contribution")
	}
	sum := 0.0
	for _, w := range r.Weights {
		if w <= 0 {
			t.Fatalf("risk parity weight not positive: %v", r.Weights)
		}
		sum += w
	}
	approx(t, sum, 1, 1e-9, "weights sum")
}

// A cap forces diversification away from the unconstrained tangency.
func TestMaxWeightCap(t *testing.T) {
	mu := []float64{0.10, 0.05}
	cov := [][]float64{{0.04, 0}, {0, 0.01}} // unconstrained max-sharpe → 1/3 : 2/3
	r, err := solve(mu, cov, Spec{Objective: MaxSharpe, MaxWeight: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	approx(t, r.Weights[1], 0.5, 2e-3, "capped w1")
	approx(t, r.Weights[0], 0.5, 2e-3, "capped w0")

	if _, err := solve(mu, cov, Spec{Objective: MinVolatility, MaxWeight: 0.4}); err == nil {
		t.Fatal("a 40% cap on 2 assets cannot reach 100%: expected an error")
	}
}

// The uncapped max-sharpe answer must be the GLOBAL optimum, not a good local
// one: Schaible's transformation is what buys that, so check it against a
// brute-force sweep of the whole simplex on a correlated three-asset problem.
func TestMaxSharpeIsGlobalOptimum(t *testing.T) {
	mu := []float64{0.09, 0.06, 0.04}
	cov := [][]float64{
		{0.0400, 0.0120, -0.0020},
		{0.0120, 0.0225, 0.0030},
		{-0.0020, 0.0030, 0.0100},
	}
	r, err := solve(mu, cov, Spec{Objective: MaxSharpe})
	if err != nil {
		t.Fatal(err)
	}
	sharpe := func(w []float64) float64 {
		v := quad(cov, w)
		if v <= 0 {
			return 0
		}
		return dot(mu, w) / math.Sqrt(v)
	}
	const steps = 200
	bestGrid, bestW := 0.0, []float64{}
	for i := 0; i <= steps; i++ {
		for j := 0; i+j <= steps; j++ {
			w := []float64{float64(i) / steps, float64(j) / steps, float64(steps-i-j) / steps}
			if s := sharpe(w); s > bestGrid {
				bestGrid, bestW = s, w
			}
		}
	}
	if r.Sharpe < bestGrid-1e-6 {
		t.Fatalf("max-sharpe %.6f (%v) is below the grid best %.6f (%v)",
			r.Sharpe, r.Weights, bestGrid, bestW)
	}
	approx(t, r.Sharpe, sharpe(r.Weights), 1e-12, "reported Sharpe")
}

// tangencyQP has no feasible point when every mean is negative (muᵀy = 1 is
// unreachable with y ≥ 0); the solver must then fall back to its other
// starting points instead of returning garbage.
func TestMaxSharpeAllNegativeMeans(t *testing.T) {
	mu := []float64{-0.02, -0.05}
	cov := [][]float64{{0.04, 0}, {0, 0.01}}
	if _, ok := tangencyQP(mu, cov); ok {
		t.Fatal("tangencyQP should report no feasible point when every mean is negative")
	}
	r, err := solve(mu, cov, Spec{Objective: MaxSharpe})
	if err != nil {
		t.Fatal(err)
	}
	sum := 0.0
	for _, w := range r.Weights {
		if w < -1e-12 {
			t.Fatalf("negative weight: %v", r.Weights)
		}
		sum += w
	}
	approx(t, sum, 1, 1e-9, "weights sum")
	// The best available Sharpe is the least negative one: asset 0 loses
	// 2%/yr at 20% vol (-0.10) against asset 1's 5% at 10% vol (-0.50).
	if r.Weights[0] <= r.Weights[1] {
		t.Errorf("expected the least-negative Sharpe to dominate, got %v", r.Weights)
	}
}

func TestStatsConsistency(t *testing.T) {
	mu := []float64{0.10, 0.05}
	cov := [][]float64{{0.04, 0}, {0, 0.01}}
	r, err := solve(mu, cov, Spec{Objective: MaxSharpe})
	if err != nil {
		t.Fatal(err)
	}
	approx(t, r.Return, dot(mu, r.Weights), 1e-12, "Return")
	approx(t, r.Volatility, math.Sqrt(quad(cov, r.Weights)), 1e-12, "Volatility")
	approx(t, r.Sharpe, r.Return/r.Volatility, 1e-12, "Sharpe")
}

func TestParseSpec(t *testing.T) {
	s, err := ParseSpec("max-sharpe,max-weight:40")
	if err != nil {
		t.Fatal(err)
	}
	if s.Objective != MaxSharpe || math.Abs(s.MaxWeight-0.4) > 1e-9 {
		t.Fatalf("parsed %+v", s)
	}
	if s, err := ParseSpec("RISK-PARITY"); err != nil || s.Objective != RiskParity {
		t.Fatalf("case-insensitive objective: %+v %v", s, err)
	}
	for _, bad := range []string{"", "sharpe", "max-sharpe,max-weight:0", "max-sharpe,max-weight:150", "min-volatility,foo:1", "min-volatility,bar"} {
		if _, err := ParseSpec(bad); err == nil {
			t.Fatalf("ParseSpec(%q) should fail", bad)
		}
	}
}

// The Black-Litterman grammar: views absolute and relative, their default
// confidence, and the prior return that fixes the risk aversion.
func TestParseSpecBlackLitterman(t *testing.T) {
	s, err := ParseSpec("black-litterman,view:IGLN:2,view:DBMFE>DTLA:-3.5@70,prior-return:4.6,max-vol:9")
	if err != nil {
		t.Fatal(err)
	}
	if s.Objective != BlackLitterman {
		t.Fatalf("objective = %q", s.Objective)
	}
	approx(t, s.PriorReturn, 0.046, 1e-12, "prior-return")
	approx(t, s.Limits.MaxVolatility, 0.09, 1e-12, "max-vol still composes")
	if len(s.Views) != 2 {
		t.Fatalf("parsed %d views, want 2", len(s.Views))
	}
	if s.Views[0].Asset != "IGLN" || s.Views[0].Versus != "" {
		t.Fatalf("absolute view = %+v", s.Views[0])
	}
	approx(t, s.Views[0].Return, 0.02, 1e-12, "absolute view return")
	approx(t, s.Views[0].Confidence, 0.5, 1e-12, "default confidence")
	if s.Views[1].Asset != "DBMFE" || s.Views[1].Versus != "DTLA" {
		t.Fatalf("relative view = %+v", s.Views[1])
	}
	approx(t, s.Views[1].Return, -0.035, 1e-12, "a negative view is a statement too")
	approx(t, s.Views[1].Confidence, 0.7, 1e-12, "stated confidence")

	// A zero view is a statement as well (a zero-real-return belief).
	zero, err := ParseSpec("black-litterman,view:GDE:0")
	if err != nil {
		t.Fatal(err)
	}
	approx(t, zero.Views[0].Return, 0, 1e-12, "zero view")

	for _, bad := range []string{
		"black-litterman,view:X",           // no return
		"black-litterman,view:X:abc",       // not a number
		"black-litterman,view:X:2@0",       // a switched-off view
		"black-litterman,view:X:2@100",     // a certainty
		"black-litterman,view:X:2@110",     // out of range
		"black-litterman,view:>B:2",        // missing side
		"black-litterman,view:A>:2",        // missing side
		"black-litterman,prior-return:0",   // no scale at all
		"black-litterman,prior-return:abc", // not a number
		"max-sharpe,view:X:2",              // views belong to black-litterman
		"risk-parity,view:X:2",             //
		"max-sortino,prior-return:4",       //
		"cwarp,view:X:2",                   //
	} {
		if _, err := ParseSpec(bad); err == nil {
			t.Fatalf("ParseSpec(%q) should fail", bad)
		}
	}
	// The refusals name the offending token, not just the objective.
	if _, err := ParseSpec("max-sharpe,view:X:2"); err == nil || !strings.Contains(err.Error(), "view") {
		t.Fatalf("error = %v, want one naming view", err)
	}
	if _, err := ParseSpec("max-sharpe,prior-return:4"); err == nil || !strings.Contains(err.Error(), "prior-return") {
		t.Fatalf("error = %v, want one naming prior-return", err)
	}
}

func TestSolveValidation(t *testing.T) {
	if _, err := Solve(nil, Spec{Objective: MaxSharpe}); err == nil {
		t.Fatal("no assets should fail")
	}
	if _, err := Solve([][]float64{{0.01}}, Spec{Objective: MaxSharpe}); err == nil {
		t.Fatal("single observation should fail")
	}
	if _, err := Solve([][]float64{{0.01, 0.02}, {0.01}}, Spec{Objective: MaxSharpe}); err == nil {
		t.Fatal("ragged returns should fail")
	}
	// Single asset is trivially fully weighted.
	r, err := Solve([][]float64{{0.01, -0.02, 0.03}}, Spec{Objective: MaxSharpe})
	if err != nil || len(r.Weights) != 1 || math.Abs(r.Weights[0]-1) > 1e-12 {
		t.Fatalf("single asset: %+v %v", r, err)
	}
}
