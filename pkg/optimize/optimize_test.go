package optimize

import (
	"math"
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
