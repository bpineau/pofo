package metrics

import (
	"fmt"
	"math"
	"time"
)

// ComputeOptions supplies explicit conventions for ComputeWith.
type ComputeOptions struct {
	// PeriodsPerYear overrides date-based inference. Zero infers the cadence.
	PeriodsPerYear float64
	// RiskFree is an annual arithmetic risk-free return, as a fraction.
	// It affects Sharpe and the downside target of Sortino, not CAGR.
	RiskFree float64
}

// ComputeWith is Compute with an explicit cadence and risk-free return.
// The zero options preserve Compute's conventions. Non-finite options and
// a negative cadence are errors. All other validation is shared with Compute.
func ComputeWith(dates []time.Time, values []float64, opt ComputeOptions) (Stats, error) {
	if math.IsNaN(opt.PeriodsPerYear) || math.IsInf(opt.PeriodsPerYear, 0) || opt.PeriodsPerYear < 0 || math.IsNaN(opt.RiskFree) || math.IsInf(opt.RiskFree, 0) {
		return Stats{}, fmt.Errorf("metrics: invalid cadence or risk-free return")
	}
	s, err := Compute(dates, values)
	if err != nil {
		return Stats{}, err
	}
	if opt.PeriodsPerYear > 0 {
		s.PeriodsPerYear = opt.PeriodsPerYear
	}
	r := Returns(values)
	s.Volatility = Volatility(r, s.PeriodsPerYear)
	s.Sharpe = Sharpe(r, opt.RiskFree, s.PeriodsPerYear)
	s.Sortino = Sortino(r, opt.RiskFree, s.PeriodsPerYear)
	return s, nil
}
