package optimize

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestConstantPortfolioSharpeUndefined(t *testing.T) {
	panel := [][]float64{{0, 0, 0, 0}}
	measured, err := Evaluate(panel, []float64{1}, 12)
	if err != nil {
		t.Fatal(err)
	}
	fit, err := Solve(panel, 12, Spec{Objective: MinVolatility})
	if err != nil {
		t.Fatal(err)
	}
	for name, x := range map[string]float64{"evaluated": measured.Sharpe, "solved": fit.Sharpe} {
		if !math.IsNaN(x) {
			t.Errorf("%s Sharpe = %v, want NaN", name, x)
		}
	}
	for _, v := range []any{measured, fit, FitEvaluation{Fit: fit, InSample: measured, OutOfSample: measured}, FrontierPoint{Measured: measured, Result: fit}} {
		b, err := json.Marshal(v)
		if err != nil || !strings.Contains(string(b), `"Sharpe":null`) {
			t.Errorf("undefined Sharpe JSON = %s, %v", b, err)
		}
	}
}

func TestEvaluationAssetPermutation(t *testing.T) {
	a, b := []float64{.1, -.05, .03, -.02}, []float64{.01, .02, -.01, .03}
	x, err := EvaluateWithRiskFree([][]float64{a, b}, []float64{.3, .7}, 12, .02)
	if err != nil {
		t.Fatal(err)
	}
	y, err := EvaluateWithRiskFree([][]float64{b, a}, []float64{.7, .3}, 12, .02)
	if err != nil {
		t.Fatal(err)
	}
	if x != y {
		t.Fatalf("asset order changed performance: %+v vs %+v", x, y)
	}
}

func TestNonzeroConstantReturnSharpeUndefined(t *testing.T) {
	p := [][]float64{{.1, .1, .1}}
	e, err := Evaluate(p, []float64{1}, 12)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Solve(p, 12, Spec{Objective: MinVolatility})
	if err != nil {
		t.Fatal(err)
	}
	if e.Volatility != 0 || r.Volatility != 0 || !math.IsNaN(e.Sharpe) || !math.IsNaN(r.Sharpe) {
		t.Fatalf("constant path: evaluation %+v, solve %+v", e, r)
	}
}
