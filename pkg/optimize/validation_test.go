package optimize

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestSolveValidatesPublicInputs(t *testing.T) {
	panel := [][]float64{{.01, -.02, .03}, {-.02, .01, .04}}
	cases := []struct {
		name string
		spec Spec
	}{
		{"unknown single asset", Spec{Objective: "unknown"}},
		{"negative cap", Spec{Objective: MaxSharpe, MaxWeight: -.1}},
		{"nan cap", Spec{Objective: MaxSharpe, MaxWeight: math.NaN()}},
		{"large cap", Spec{Objective: MaxSharpe, MaxWeight: 1.1}},
		{"negative floor", Spec{Objective: MinVolatility, MinWeight: -.1}},
		{"nan limit", Spec{Objective: MaxReturn, Limits: Limits{MaxVolatility: math.NaN()}}},
		{"negative limit", Spec{Objective: MaxReturn, Limits: Limits{MinReturn: -.1}}},
		{"risk parity limits", Spec{Objective: RiskParity, Limits: Limits{MaxVolatility: .1}}},
		{"cwarp via solve", Spec{Objective: CWARP}},
		{"unresolved bounds", Spec{Objective: MaxSharpe, Bounds: map[string][2]float64{"X": {0, .8}}}},
		{"short bounds", Spec{Objective: MaxSharpe, Upper: []float64{.8}}},
		{"invalid bounds", Spec{Objective: MaxSharpe, Lower: []float64{-.1, 0}}},
		{"infinite bounds", Spec{Objective: MaxSharpe, Upper: []float64{math.Inf(1), 1}}},
		{"foreign prior", Spec{Objective: MaxSharpe, Prior: []float64{.5, .5}}},
		{"unresolved view", Spec{Objective: BlackLitterman, Prior: []float64{.5, .5}, Views: []View{{Asset: "X", Confidence: .5}}}},
		{"negative prior", Spec{Objective: BlackLitterman, Prior: []float64{-.5, 1.5}}},
		{"nan prior", Spec{Objective: BlackLitterman, Prior: []float64{math.NaN(), .5}}},
		{"nan prior return", Spec{Objective: BlackLitterman, Prior: []float64{.5, .5}, PriorReturn: math.NaN()}},
		{"train", Spec{Objective: MaxSharpe, Train: Window{Start: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Solve(panel, 12, tc.spec); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	if _, err := Solve(panel[:1], 12, Spec{Objective: "unknown"}); err == nil {
		t.Fatal("single-asset shortcut accepted unknown objective")
	}
	for _, x := range []float64{math.NaN(), math.Inf(1), -1, -1.1} {
		if _, err := Solve([][]float64{{.01, x}}, 12, Spec{Objective: MinVolatility}); err == nil {
			t.Errorf("accepted return %v", x)
		}
	}
}

func TestCWARPValidationAndFeasible(t *testing.T) {
	repl, a, b := cwarpSeries()
	panel := [][]float64{a, b}
	for _, spec := range []Spec{
		{Objective: MaxSharpe}, {Objective: CWARP, MinWeight: .1},
		{Objective: CWARP, Lower: []float64{.1, 0}},
		{Objective: CWARP, Limits: Limits{MaxVolatility: .1}},
		{Objective: CWARP, MaxWeight: math.NaN()},
	} {
		if _, err := SolveCWARP(panel, repl, 12, spec); err == nil {
			t.Errorf("accepted %+v", spec)
		}
	}
	bad := append([]float64(nil), repl...)
	bad[0] = math.NaN()
	if _, err := SolveCWARP(panel, bad, 12, Spec{Objective: CWARP}); err == nil {
		t.Fatal("accepted NaN replacement")
	}
	r, err := SolveCWARP(panel, repl, 12, Spec{Objective: CWARP})
	if err != nil || !r.Feasible {
		t.Fatalf("CWARP feasible = %v, %v", r.Feasible, err)
	}
}

func TestRiskParityLegacyBounds(t *testing.T) {
	// Historically all weight bounds are ignored for risk parity, including
	// a scalar cap too small to hold any allocation. Preserve that contract.
	panel := [][]float64{{.01, -.01, .01, -.01}, {.02, .02, -.02, -.02}}
	r, err := Solve(panel, 12, Spec{Objective: RiskParity, MaxWeight: .1, MinWeight: .2})
	if err != nil || r.Feasible || r.FeasibilityReason == "" {
		t.Fatalf("legacy bounds: %+v, %v", r, err)
	}
}

func TestPathStatsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		r   [][]float64
		w   []float64
		ppy float64
	}{
		{[][]float64{{.01, .02}, {.01}}, []float64{.5, .5}, 12},
		{[][]float64{{.01, .02}}, []float64{math.NaN()}, 12},
		{[][]float64{{.01, .02}}, []float64{1}, math.NaN()},
	} {
		a, b, c := PathStats(tc.r, tc.w, tc.ppy)
		if !math.IsNaN(a) || !math.IsNaN(b) || !math.IsNaN(c) {
			t.Fatal("invalid input must return NaNs")
		}
	}
}

func TestResolvedViewValidation(t *testing.T) {
	for _, v := range []View{{Asset: "A", Return: math.NaN(), Confidence: .5}, {Asset: "A", Confidence: math.NaN()}} {
		s := Spec{Objective: BlackLitterman, Prior: []float64{.5, .5}, Views: []View{v}}
		if err := s.Resolve([][]string{{"A"}, {"B"}}); err != nil {
			t.Fatal(err)
		}
		if _, err := Solve([][]float64{{.01, -.01}, {-.02, .02}}, 12, s); err == nil {
			t.Fatal("accepted invalid view")
		}
	}
	s := Spec{Objective: MaxSharpe, Bounds: map[string][2]float64{"A": {0, .8}}}
	if _, err := Solve([][]float64{{.01, -.01}, {-.02, .02}}, 12, s); err == nil || !strings.Contains(err.Error(), "Resolve") {
		t.Fatalf("unresolved bound: %v", err)
	}
}

// Tiny but finite returns used to create an enormous inverse-variance start
// and spend forever expanding a projection bracket one unit at a time.
func TestSolveTinyFiniteReturns(t *testing.T) {
	r, err := Solve([][]float64{{1e-150, -1e-150, 2e-150}, {2e-150, 1e-150, -1e-150}}, 12, Spec{Objective: MaxSharpe})
	if err == nil {
		if err := checkWeights(r.Weights, 2, false); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSolveExtremeRiskFreeDoesNotHang(t *testing.T) {
	_, _ = Solve([][]float64{{.03, -.02, .04}, {.003, -.001, .004}}, 12, Spec{Objective: MaxSharpe, RiskFree: 1e200})
}
