package metrics_test

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/metrics"
)

func ExampleShrinkageCovariance() {
	// Sample variances are 1 and 4; sample covariance is 2.
	cov, err := metrics.ShrinkageCovariance([][]float64{{1, 2, 3}, {2, 4, 6}}, .25)
	if err != nil {
		panic(err)
	}
	fmt.Printf("variances %.1f / %.1f, covariance %.1f\n", cov[0][0], cov[1][1], cov[0][1])
	// Output:
	// variances 1.0 / 4.0, covariance 1.5
}

func ExampleRegressHAC() {
	// Synthetic ordered observations, with heteroskedastic and serial residuals.
	x := []float64{-3, -2, -1, 1, 2, 3}
	y := []float64{1.5, 3, -1.5, -.5, 5, 4.5}
	fit, err := metrics.RegressHAC(y, 1, x)
	if err != nil {
		panic(err)
	}
	fmt.Printf("alpha %.2f (HAC SE %.3f), beta %.2f (HAC SE %.3f)\n", fit.Alpha.Value, fit.Alpha.SE, fit.Betas[0].Value, fit.Betas[0].SE)
	// Output:
	// alpha 2.00 (HAC SE 0.898), beta 0.50 (HAC SE 0.274)
}

func ExampleBlockBootstrap() {
	// Synthetic monthly observations; two-month circular blocks retain
	// consecutive pairs. This measures estimation uncertainty, not a forecast.
	returns := []float64{.02, -.01, .03, .01, -.02, .04, .01, -.01, .02, .01, -.03, .02}
	result, err := metrics.BlockBootstrap(returns, metrics.BootstrapOptions{
		PeriodsPerYear: 12, BlockLength: 2, Repetitions: 1000, Confidence: .9, Seed: 7,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("CAGR %.2f %%, 90 %% interval [%.2f, %.2f] %%\n", result.CAGR.Estimate*100, result.CAGR.Lower*100, result.CAGR.Upper*100)
	fmt.Printf("defined Sharpe replicates: %d/%d\n", result.Sharpe.Defined, result.Options.Repetitions)
	// Output:
	// CAGR 9.12 %, 90 % interval [-0.22, 19.31] %
	// defined Sharpe replicates: 1000/1000
}

func ExampleBrinson() {
	// Two aligned sectors, with a 60/40 portfolio and 50/50 benchmark.
	attribution, err := metrics.Brinson(
		[]float64{.6, .4}, []float64{.5, .5},
		[]float64{.12, .02}, []float64{.10, .04},
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("active %.1f pt: allocation %.1f, selection cancels %v, interaction %.1f\n", attribution.ActiveReturn*100, attribution.Allocation*100, math.Abs(attribution.Selection) < 1e-12, attribution.Interaction*100)
	// Output:
	// active 1.0 pt: allocation 0.6, selection cancels true, interaction 0.4
}
