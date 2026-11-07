package metrics_test

import (
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/metrics"
)

func ExampleComputeWith() {
	dates := []time.Time{
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
	}
	stats, err := metrics.ComputeWith(dates, []float64{100, 110, 99}, metrics.ComputeOptions{
		PeriodsPerYear: 12, RiskFree: .06,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%.0f periods/year, Sharpe %.3f\n", stats.PeriodsPerYear, stats.Sharpe)
	// Output: 12 periods/year, Sharpe -0.122
}
