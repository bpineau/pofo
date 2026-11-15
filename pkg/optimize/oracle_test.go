package optimize_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/optimize"
)

// Orthogonal centered rows have variance ratio 1:4. Their minimum-variance
// allocation is (4/5,1/5); positive-excess tangency weights are mu_i/var_i.
func orthogonalPanel(scale float64) [][]float64 {
	a, b := []float64{.014, .014, -.006, -.006}, []float64{.026, -.014, .026, -.014}
	for i := range a {
		a[i] *= scale
		b[i] *= scale
	}
	return [][]float64{a, b}
}

func TestAnalyticAllocationsAcrossReturnScales(t *testing.T) {
	for _, scale := range []float64{.001, .01, 1, 10} {
		for _, objective := range []optimize.Objective{optimize.MinVolatility, optimize.MaxSharpe} {
			t.Run(fmt.Sprintf("%s/%g", objective, scale), func(t *testing.T) {
				spec := optimize.Spec{Objective: objective}
				want := .8
				if objective == optimize.MaxSharpe {
					spec.RiskFree = .012 * scale
					want = 30. / 42.5
				}
				r, err := optimize.Solve(orthogonalPanel(scale), 12, spec)
				if err != nil {
					t.Fatal(err)
				}
				if !r.Feasible || math.Abs(r.Weights[0]-want) > 1e-5 {
					t.Fatalf("weights=%v feasible=%v, analytic first weight=%g", r.Weights, r.Feasible, want)
				}
			})
		}
	}
}

func TestSharpeDoesNotPreferUndefinedCashToNegativeRatios(t *testing.T) {
	panel := [][]float64{{0, 0, 0, 0}, {-.01, -.03, -.01, -.03}}
	r, err := optimize.Solve(panel, 12, optimize.Spec{Objective: optimize.MaxSharpe})
	if err != nil {
		t.Fatal(err)
	}
	if math.IsNaN(r.Sharpe) || r.Weights[1] == 0 {
		t.Fatalf("undefined cash ranked above all measurable ratios: %+v", r)
	}
}

func TestMinimumVarianceDuplicateExposure(t *testing.T) {
	panel := orthogonalPanel(1)
	panel = append(panel, append([]float64(nil), panel[0]...))
	for _, order := range [][]int{{0, 1, 2}, {1, 2, 0}, {2, 0, 1}} {
		rows := [][]float64{panel[order[0]], panel[order[1]], panel[order[2]]}
		r, err := optimize.Solve(rows, 12, optimize.Spec{Objective: optimize.MinVolatility})
		if err != nil {
			t.Fatal(err)
		}
		for i, id := range order {
			if id == 1 && math.Abs(r.Weights[i]-.2) > 1e-5 {
				t.Fatalf("duplicate exposure changed optimum: order %v weights %v", order, r.Weights)
			}
		}
	}
}

func TestBoundedMinimumVarianceAcrossScales(t *testing.T) {
	for _, scale := range []float64{.001, .01, 1, 10} {
		r, err := optimize.Solve(orthogonalPanel(scale), 12, optimize.Spec{
			Objective: optimize.MinVolatility, Lower: []float64{.1, .1},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !r.Feasible || math.Abs(r.Weights[0]-.8) > 1e-5 {
			t.Errorf("scale %g weights=%v, analytic first weight=.8", scale, r.Weights)
		}
	}
}

func TestSharpeRejectsEverywhereUndefinedObjective(t *testing.T) {
	for _, n := range []int{1, 2} {
		panel := make([][]float64, n)
		for i := range panel {
			panel[i] = []float64{.01, .01, .01, .01}
		}
		for _, bounded := range []bool{false, true} {
			spec := optimize.Spec{Objective: optimize.MaxSharpe}
			if bounded {
				spec.Lower = make([]float64, n)
				for i := range spec.Lower {
					spec.Lower[i] = .1
				}
			}
			if r, err := optimize.Solve(panel, 12, spec); err == nil {
				t.Errorf("n=%d bounded=%v: undefined objective returned success: %+v", n, bounded, r)
			}
		}
	}
}

// The grid does no gradient search and scores every feasible allocation
// through the public measurement API. Its best score is a lower bound on
// the continuous optimum, not an assertion of global optimality.
func TestThreeAssetGridOracle(t *testing.T) {
	panel := make([][]float64, 3)
	for i := range panel {
		panel[i] = make([]float64, 60)
	}
	for j := range panel[0] {
		x := float64(j)
		panel[0][j] = .012 + .07*math.Sin(x*.9)
		panel[1][j] = .006 + .025*math.Cos(x*.7)
		panel[2][j] = .002 - .012*math.Sin(x*.9) + .004*math.Cos(x*1.3)
	}
	cases := []struct {
		name string
		spec optimize.Spec
	}{
		{"variance", optimize.Spec{Objective: optimize.MinVolatility}},
		{"sharpe", optimize.Spec{Objective: optimize.MaxSharpe, RiskFree: .02}},
		{"return-vol-cap", optimize.Spec{Objective: optimize.MaxReturn, Limits: optimize.Limits{MaxVolatility: .08}}},
		{"return-dd-cap", optimize.Spec{Objective: optimize.MaxReturn, Limits: optimize.Limits{MaxDrawdown: .06}}},
		{"variance-return-floor", optimize.Spec{Objective: optimize.MinVolatility, Limits: optimize.Limits{MinReturn: .07}}},
		{"narrow-box", optimize.Spec{Objective: optimize.MaxReturn, Lower: []float64{.32, .32, .32}, Upper: []float64{.35, .35, .35}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			score := func(w []float64) (float64, bool) {
				for i, x := range w {
					if len(tc.spec.Lower) > 0 && (x < tc.spec.Lower[i]-1e-9 || x > tc.spec.Upper[i]+1e-9) {
						return 0, false
					}
				}
				p, err := optimize.EvaluateWithRiskFree(panel, w, 12, tc.spec.RiskFree)
				if err != nil {
					t.Fatal(err)
				}
				lim := tc.spec.Limits
				if lim.MaxVolatility > 0 && p.Volatility > lim.MaxVolatility+1e-9 || lim.MinReturn > 0 && p.CAGR < lim.MinReturn-1e-9 || lim.MaxDrawdown > 0 && -p.MaxDrawdown > lim.MaxDrawdown+1e-9 {
					return 0, false
				}
				switch tc.spec.Objective {
				case optimize.MinVolatility:
					return -p.Volatility, true
				case optimize.MaxSharpe:
					return p.Sharpe, !math.IsNaN(p.Sharpe)
				default:
					return p.CAGR, true
				}
			}
			best := math.Inf(-1)
			var bestW []float64
			for a := 0; a <= 100; a++ {
				for b := 0; b <= 100-a; b++ {
					w := []float64{float64(a) / 100, float64(b) / 100, float64(100-a-b) / 100}
					if s, ok := score(w); ok && s > best {
						best, bestW = s, w
					}
				}
			}
			if bestW == nil {
				t.Fatal("fixture has no feasible grid allocation")
			}
			r, err := optimize.Solve(panel, 12, tc.spec)
			if err != nil {
				t.Fatal(err)
			}
			got, feasible := score(r.Weights)
			t.Logf("solver %v score %.9f; grid %v score %.9f; deficit %.9g", r.Weights, got, bestW, best, best-got)
			// The nonsmooth drawdown boundary has a measured 1.64 bp CAGR
			// deficit on this fixture. Pin a 2 bp ceiling, not global optimality.
			tolerance := 1e-4
			if tc.spec.Limits.MaxDrawdown > 0 {
				tolerance = 2e-4
			}
			if !r.Feasible || !feasible || got < best-tolerance {
				t.Fatalf("solver misses feasible grid benchmark: result %+v, measured score %g; grid %v score %g", r, got, bestW, best)
			}
		})
	}
}
