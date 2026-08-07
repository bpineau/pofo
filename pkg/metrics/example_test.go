package metrics_test

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

// Compute derives every statistic from a series of dated values.
func ExampleCompute() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := make([]time.Time, 5)
	for i := range dates {
		dates[i] = start.AddDate(0, 0, i)
	}
	stats, err := metrics.Compute(dates, []float64{100, 110, 99, 104, 108})
	if err != nil {
		panic(err)
	}
	fmt.Printf("MaxDrawdown: %.1f %%\n", stats.MaxDrawdown*100)
	fmt.Printf("TTR: %d days (ongoing: %v)\n", stats.TTRDays, stats.TTROngoing)
	// Output:
	// MaxDrawdown: -10.0 %
	// TTR: 3 days (ongoing: true)
}

// CWARP scores whether overlaying an asset on a replacement portfolio (here
// equity beta) improves its risk-adjusted returns. An anti-correlated sleeve
// with positive carry scores above zero.
func ExampleCWARP() {
	equity := make([]float64, 250)
	diversifier := make([]float64, 250)
	for i := range equity {
		equity[i] = 0.001 + 0.006*math.Sin(float64(i)*0.3)
		if i >= 100 && i < 115 {
			equity[i] = -0.010 // a drawdown
		}
		diversifier[i] = -equity[i] + 0.0007 // hedge plus carry
	}
	score, ok := metrics.CWARP(diversifier, equity, metrics.CWARPParams{})
	fmt.Printf("improves the portfolio: %v\n", ok && score > 0)
	// Output:
	// improves the portfolio: true
}

// ReturnToMaxDrawdown is the Calmar-style ratio of annualized growth to the
// worst peak-to-trough loss, the return-to-drawdown building block CWARP and
// the optimizer reuse.
func ExampleReturnToMaxDrawdown() {
	returns := make([]float64, 250)
	for i := range returns {
		returns[i] = 0.002 // steady gains…
		if i >= 100 && i < 110 {
			returns[i] = -0.02 // …interrupted by a drawdown
		}
	}
	r, ok := metrics.ReturnToMaxDrawdown(returns, 0)
	fmt.Printf("defined: %v, positive: %v\n", ok, r > 0)
	// Output:
	// defined: true, positive: true
}

// Ulcer measures how painful the drawdowns were (depth and duration), and
// WorstRollingReturn the worst outcome over any window of the given length:
// the two underwater-robustness quantities the decumulation-minded optimizer
// objectives (min-ulcer, max-worst-5y) target.
func ExampleUlcer() {
	returns := make([]float64, 300)
	for i := range returns {
		returns[i] = 0.001
		if i >= 100 && i < 130 {
			returns[i] = -0.01 // a prolonged drawdown
		}
	}
	worst, _ := metrics.WorstRollingReturn(returns, 252)
	fmt.Printf("Ulcer > 0: %v, worst 1y return negative: %v\n",
		metrics.Ulcer(returns) > 0, worst < 0)
	// Output:
	// Ulcer > 0: true, worst 1y return negative: true
}

// Beta regresses a series' daily returns on a benchmark's, matching
// observations by date.
func ExampleBeta() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	n := 40
	dates := make([]time.Time, n)
	bench := make([]float64, n)
	port := make([]float64, n)
	bench[0], port[0] = 100, 100
	for i := 1; i < n; i++ {
		dates[i-1] = start.AddDate(0, 0, i-1)
		r := 0.01 * float64(i%5-2)
		bench[i] = bench[i-1] * (1 + r)
		port[i] = port[i-1] * (1 + 2*r) // exactly twice the benchmark
	}
	dates[n-1] = start.AddDate(0, 0, n-1)
	beta, ok := metrics.Beta(dates, port, dates, bench)
	fmt.Printf("beta=%.1f ok=%v\n", beta, ok)
	// Output:
	// beta=2.0 ok=true
}

// TWR neutralizes external flows: a deposit is not performance. Here the
// market gains 10 % on day 2, then a 100 deposit lands on day 3: the
// money-agnostic return stays +10 %.
func ExampleTWR() {
	d0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := []time.Time{d0, d0.AddDate(0, 0, 1), d0.AddDate(0, 0, 2)}
	values := []float64{100, 110, 210}
	flows := []metrics.Flow{{Date: dates[2], Amount: 100}}
	twr, ok := metrics.TWR(dates, values, flows)
	fmt.Printf("ok=%v TWR=%.1f %%\n", ok, twr*100)
	// Output:
	// ok=true TWR=10.0 %
}

// IRR weighs each cash flow by its date: money invested early counts more
// than money added late. Flows are signed from the investor's standpoint
// (negative going in, positive coming out).
func ExampleIRR() {
	d0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	dates := []time.Time{d0, d0.AddDate(1, 0, 0)}
	flows := []float64{-1000, -1000} // initial capital, then one contribution
	irr, ok := metrics.IRR(dates, flows, d0.AddDate(2, 0, 0), 2200)
	fmt.Printf("ok=%v IRR=%.1f %%/yr\n", ok, irr*100)
	// Output:
	// ok=true IRR=6.5 %/yr
}
