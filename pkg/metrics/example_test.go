package metrics_test

import (
	"fmt"
	"time"

	"portfodor/pkg/metrics"
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
