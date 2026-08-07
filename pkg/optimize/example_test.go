package optimize_test

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/optimize"
)

// Solve computes the weights for an objective from the assets' aligned daily
// returns. Here two uncorrelated assets (one volatile, one calm) are
// balanced for minimum variance: the calmer asset gets the larger weight.
func ExampleSolve() {
	// Asset A swings ±2 %, asset B ±1 %, with zero covariance.
	a := []float64{0.02, -0.02, 0.02, -0.02}
	b := []float64{0.01, 0.01, -0.01, -0.01}

	res, err := optimize.Solve([][]float64{a, b}, optimize.Spec{Objective: optimize.MinVolatility})
	if err != nil {
		panic(err)
	}
	fmt.Printf("A %.0f %%, B %.0f %%\n", res.Weights[0]*100, res.Weights[1]*100)
	// Output:
	// A 20 %, B 80 %
}

// SolveCWARP finds the blend that best improves a replacement portfolio (a
// benchmark). Offered equity beta and an anti-correlated diversifier, it loads
// the diversifier and reaches a positive CWARP.
func ExampleSolveCWARP() {
	repl := make([]float64, 300)
	equity := make([]float64, 300)
	diversifier := make([]float64, 300)
	for i := range repl {
		repl[i] = 0.001 + 0.006*math.Sin(float64(i)*0.3)
		if i >= 100 && i < 115 {
			repl[i] = -0.010 // a drawdown
		}
		equity[i] = repl[i]                // more equity beta
		diversifier[i] = -repl[i] + 0.0007 // hedge plus carry
	}
	res, err := optimize.SolveCWARP([][]float64{equity, diversifier}, repl, optimize.Spec{Objective: optimize.CWARP})
	if err != nil {
		panic(err)
	}
	fmt.Printf("diversifier favored: %v, CWARP positive: %v\n",
		res.Weights[1] > res.Weights[0], res.CWARP > 0)
	// Output:
	// diversifier favored: true, CWARP positive: true
}

// ParseSpec reads the value of a "#meta optimize:" directive.
func ExampleParseSpec() {
	spec, err := optimize.ParseSpec("max-sharpe,max-weight:40")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s, cap %.0f %%\n", spec.Objective, spec.MaxWeight*100)
	// Output:
	// max-sharpe, cap 40 %
}
