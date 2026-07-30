package optimize_test

import (
	"fmt"

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
