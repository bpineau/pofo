package optimize

import (
	"math"
	"reflect"
	"testing"
)

func TestRiskBudgetsDiagonal(t *testing.T) {
	// Orthogonal zero-mean rows: variances are in ratio 1:4:9.
	// With budgets 1/6, 1/3, 1/2, w_i is proportional to sqrt(b_i)/sigma_i.
	panel := [][]float64{{.01, .01, -.01, -.01}, {.02, -.02, .02, -.02}, {.03, -.03, -.03, .03}}
	budgets := []float64{1.0 / 6, 1.0 / 3, .5}
	r, err := Solve(panel, 12, Spec{Objective: RiskParity, RiskBudgets: budgets})
	if err != nil {
		t.Fatal(err)
	}
	_, cov := meanCov(panel, 12)
	sw := matVec(cov, r.Weights)
	variance := dot(r.Weights, sw)
	total := math.Sqrt(budgets[0])/.01 + math.Sqrt(budgets[1])/.02 + math.Sqrt(budgets[2])/.03
	for i, w := range r.Weights {
		want := math.Sqrt(budgets[i]) / []float64{.01, .02, .03}[i] / total
		if math.Abs(w-want) > 2e-5 {
			t.Errorf("weight %d = %.8f, want %.8f", i, w, want)
		}
		if share := w * sw[i] / variance; math.Abs(share-budgets[i]) > 2e-5 {
			t.Errorf("risk share %d = %.8f, want %.8f", i, share, budgets[i])
		}
	}
	for _, b := range [][]float64{{.5, .5}, {0, .5, .5}, {-.1, .5, .6}, {math.NaN(), .5, .5}, {math.Inf(1), .5, .5}, {.1, .1, .1}} {
		if _, err := Solve(panel, 12, Spec{Objective: RiskParity, RiskBudgets: b}); err == nil {
			t.Errorf("accepted budgets %v", b)
		}
	}
	if _, err := Solve(panel, 12, Spec{Objective: MinVolatility, RiskBudgets: budgets}); err == nil {
		t.Fatal("budget on unrelated objective accepted")
	}
	if _, err := Solve([][]float64{{.01, .01, .01}, {.02, -.02, .03}}, 12, Spec{Objective: RiskParity}); err == nil {
		t.Fatal("risk parity accepted zero variance asset")
	}
}

func TestFrontierReusesSolve(t *testing.T) {
	panel := [][]float64{{.03, -.02, .04, -.01, .02, -.02}, {.003, -.001, .004, -.002, .003, -.001}}
	caps := []float64{.0001, .02, .2}
	spec := Spec{Objective: MaxReturn, MinWeight: .05}
	points, err := Frontier(panel, 12, spec, caps)
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != len(caps) {
		t.Fatal("lost frontier points")
	}
	for i, p := range points {
		sp := spec
		sp.Limits.MaxVolatility = caps[i]
		want, err := Solve(panel, 12, sp)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(p.Result, want) || p.MaxVolatility != caps[i] {
			t.Fatalf("point %d differs from Solve", i)
		}
		measured, err := Evaluate(panel, want.Weights, 12)
		if err != nil {
			t.Fatal(err)
		}
		if p.Measured != measured {
			t.Fatalf("measurement mismatch: %+v, %+v", p.Measured, measured)
		}
	}
	if points[0].Result.Feasible {
		t.Fatal("impossible volatility cap marked feasible")
	}
	if !points[2].Result.Feasible {
		t.Fatal("loose cap not feasible")
	}
	for _, bad := range [][]float64{nil, {0}, {-.1}, {math.NaN()}, {math.Inf(1)}} {
		if _, err := Frontier(panel, 12, spec, bad); err == nil {
			t.Fatalf("accepted caps %v", bad)
		}
	}
	if _, err := Frontier(panel, 12, Spec{Objective: RiskParity}, caps); err == nil {
		t.Fatal("accepted incompatible objective")
	}
	if _, err := Frontier(panel, 12, Spec{Objective: MaxReturn, Limits: Limits{MaxVolatility: .1}}, caps); err == nil {
		t.Fatal("silently replaced explicit cap")
	}
}

func TestFitEvaluateTrainingIndependence(t *testing.T) {
	training := [][]float64{{.02, -.01, .03, -.01}, {.01, -.02, .01, .02}}
	test := [][]float64{{-.03, -.02, .01, .02}, {.04, .03, -.01, .01}}
	spec := Spec{Objective: MinVolatility}
	got, err := FitEvaluate(training, test, 12, spec)
	if err != nil {
		t.Fatal(err)
	}
	other, err := FitEvaluate(training, [][]float64{{.5, -.4, .5}, {-.4, .5, -.4}}, 12, spec)
	if err != nil {
		t.Fatal(err)
	}
	direct, err := Solve(training, 12, spec)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Fit, direct) || !reflect.DeepEqual(got.Fit, other.Fit) || got.InSample != other.InSample {
		t.Fatal("test data affected training")
	}
	if got.OutOfSample == other.OutOfSample {
		t.Fatal("test data did not affect evaluation")
	}
	in, _ := Evaluate(training, direct.Weights, 12)
	out, _ := Evaluate(test, direct.Weights, 12)
	if got.InSample != in || got.OutOfSample != out {
		t.Fatal("in/out sample labels swapped")
	}
	for _, bad := range [][][]float64{nil, {{.1, .2}}, {{.1, .2}, {math.NaN(), .1}}, {{.1}, {.1}}} {
		if _, err := FitEvaluate(training, bad, 12, spec); err == nil {
			t.Fatal("accepted invalid test panel")
		}
	}
	for _, w := range [][]float64{nil, {1}, {.5, math.NaN()}, {-.1, 1.1}, {.2, .2}} {
		if _, err := Evaluate(training, w, 12); err == nil {
			t.Errorf("accepted evaluation weights %v", w)
		}
	}
}

func TestRiskFreeSharpeAndSortino(t *testing.T) {
	panel := [][]float64{{.003, .005, .004, .002}, {.03, -.01, .04, -.02}}
	zero, err := Solve(panel, 12, Spec{Objective: MaxSharpe})
	if err != nil {
		t.Fatal(err)
	}
	raised, err := Solve(panel, 12, Spec{Objective: MaxSharpe, RiskFree: .06})
	if err != nil {
		t.Fatal(err)
	}
	if raised.Weights[1] <= zero.Weights[1] {
		t.Fatalf("risk-free rate did not reduce cash-like attraction: %v -> %v", zero.Weights, raised.Weights)
	}
	measured, err := EvaluateWithRiskFree(panel, raised.Weights, 12, .06)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(raised.Sharpe-measured.Sharpe) > 1e-10 {
		t.Fatalf("solve/evaluate Sharpe differ: %g, %g", raised.Sharpe, measured.Sharpe)
	}
	for _, obj := range []Objective{MaxSharpe, MaxSortino} {
		r, err := Solve(panel, 12, Spec{Objective: obj, RiskFree: .06, Limits: Limits{MaxVolatility: .2}})
		if err != nil {
			t.Fatal(err)
		}
		if !r.Feasible {
			t.Fatal("ordinary constrained solve became infeasible")
		}
	}
	for _, rf := range []float64{math.NaN(), math.Inf(1)} {
		if _, err := Solve(panel, 12, Spec{Objective: MaxSharpe, RiskFree: rf}); err == nil {
			t.Fatal("invalid risk-free rate accepted")
		}
		if _, err := EvaluateWithRiskFree(panel, []float64{.5, .5}, 12, rf); err == nil {
			t.Fatal("invalid evaluation rate accepted")
		}
	}
	if _, err := Solve(panel, 12, Spec{Objective: RiskParity, RiskFree: .01}); err == nil {
		t.Fatal("unrelated objective silently accepted rate")
	}
}

func TestRiskParityBoundsFeasibility(t *testing.T) {
	panel := [][]float64{{.01, .01, -.01, -.01}, {.02, -.02, .02, -.02}}
	for _, cap := range []float64{.1, .5, .8} {
		r, err := Solve(panel, 12, Spec{Objective: RiskParity, MaxWeight: cap})
		if err != nil {
			t.Fatal(err)
		}
		if want := cap >= 2.0/3; r.Feasible != want {
			t.Fatalf("cap %g: Feasible=%v, want %v", cap, r.Feasible, want)
		}
		if !r.Feasible && r.FeasibilityReason == "" {
			t.Fatal("missing violated-bound reason")
		}
	}
}

func TestRiskBudgetUnattainableNotFeasible(t *testing.T) {
	returns := [][]float64{{.1, -.1, .1, -.1}, {-.1, .1, -.1, .1}}
	for _, budgets := range [][]float64{nil, {.8, .2}} {
		result, err := Solve(returns, 12, Spec{Objective: RiskParity, RiskBudgets: budgets})
		if err != nil {
			t.Fatal(err)
		}
		if result.Feasible || result.FeasibilityReason == "" {
			t.Fatalf("risk budgets falsely certified on cancelling assets: %+v", result)
		}
	}
}

func TestSortinoUndefinedAtSelectedTarget(t *testing.T) {
	result, err := Solve([][]float64{{-.01, .02, .03}}, 12, Spec{Objective: MaxSortino, RiskFree: -.24})
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(result.Sortino) {
		t.Fatalf("no downside at selected target, want undefined Sortino; got %v", result.Sortino)
	}
}
