package metrics

import (
	"fmt"
	"math"
)

// BrinsonSector decomposes one sector's contribution to active return.
// Every effect is a single-period fraction, not a share of the active total.
type BrinsonSector struct {
	// Allocation is (portfolio weight - benchmark weight) * (sector benchmark return - total benchmark return).
	Allocation float64
	// Selection is benchmark weight * (portfolio sector return - benchmark sector return).
	Selection float64
	// Interaction is the product of the sector's weight and return differences.
	Interaction float64
}

// BrinsonAttribution is a single-period Brinson-Fachler decomposition of a
// portfolio against its benchmark, with sectors aligned by input position.
type BrinsonAttribution struct {
	// PortfolioReturn is the weighted sum of portfolio sector returns.
	PortfolioReturn float64
	// BenchmarkReturn is the weighted sum of benchmark sector returns.
	BenchmarkReturn float64
	// ActiveReturn is PortfolioReturn minus BenchmarkReturn.
	ActiveReturn float64
	// Allocation is the sum of sector allocation effects.
	Allocation float64
	// Selection is the sum of sector selection effects.
	Selection float64
	// Interaction is the sum of sector interaction effects.
	Interaction float64
	// Sectors contains effects in the supplied sector order.
	Sectors []BrinsonSector
}

// Brinson decomposes a single period's portfolio-minus-benchmark return into
// Brinson-Fachler allocation, selection and interaction. The four slices must
// name the same sectors in the same order. The caller maps positions to names.
// Both weight vectors must be finite, nonnegative fractions summing to 1
// (tolerance 1e-9); sector returns are finite fractions for that same period.
// Zero weights are supported, and no division by a sector weight occurs.
//
// The three effects reconcile to active return up to floating-point rounding.
// A sector absent from one side still needs a finite return convention for
// that side: changing it may shift the three effects while their sum remains
// the same. This is arithmetic single-period attribution, with no claim about
// multi-period geometric linking, currency attribution, fees or cash flows.
func Brinson(portfolioWeights, benchmarkWeights, portfolioReturns, benchmarkReturns []float64) (BrinsonAttribution, error) {
	n := len(portfolioWeights)
	if n == 0 || len(benchmarkWeights) != n || len(portfolioReturns) != n || len(benchmarkReturns) != n {
		return BrinsonAttribution{}, fmt.Errorf("metrics: Brinson: need four nonempty aligned sector slices")
	}
	for side, weights := range [][]float64{portfolioWeights, benchmarkWeights} {
		sum := 0.0
		for i, w := range weights {
			if isNaNOrInf(w) || w < 0 || w > 1 {
				return BrinsonAttribution{}, fmt.Errorf("metrics: Brinson: side %d sector %d weight must be a finite fraction in [0,1]", side, i)
			}
			sum += w
		}
		if math.Abs(sum-1) > 1e-9 {
			return BrinsonAttribution{}, fmt.Errorf("metrics: Brinson: side %d weights sum to %g, expected 1", side, sum)
		}
	}
	out := BrinsonAttribution{Sectors: make([]BrinsonSector, n)}
	for i := range n {
		if isNaNOrInf(portfolioReturns[i]) || isNaNOrInf(benchmarkReturns[i]) {
			return BrinsonAttribution{}, fmt.Errorf("metrics: Brinson: sector %d returns must be finite", i)
		}
		out.PortfolioReturn += portfolioWeights[i] * portfolioReturns[i]
		out.BenchmarkReturn += benchmarkWeights[i] * benchmarkReturns[i]
	}
	out.ActiveReturn = out.PortfolioReturn - out.BenchmarkReturn
	for i := range n {
		dw := portfolioWeights[i] - benchmarkWeights[i]
		dr := portfolioReturns[i] - benchmarkReturns[i]
		s := BrinsonSector{Allocation: dw * (benchmarkReturns[i] - out.BenchmarkReturn), Selection: benchmarkWeights[i] * dr, Interaction: dw * dr}
		out.Sectors[i] = s
		out.Allocation += s.Allocation
		out.Selection += s.Selection
		out.Interaction += s.Interaction
	}
	for _, v := range []float64{out.PortfolioReturn, out.BenchmarkReturn, out.ActiveReturn, out.Allocation, out.Selection, out.Interaction} {
		if isNaNOrInf(v) {
			return BrinsonAttribution{}, fmt.Errorf("metrics: Brinson: attribution overflows")
		}
	}
	return out, nil
}
