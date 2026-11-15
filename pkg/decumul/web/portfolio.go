package web

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/fire"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/scenario"
)

// AssetSeries is a weighted nominal history; see fire.AssetSeries.
type AssetSeries = fire.AssetSeries

// Coverage describes the retained monthly window; see fire.Coverage.
type Coverage = fire.Coverage

// Fit holds annual real-return parameters; see fire.Fit.
type Fit = fire.Fit

// BuildMonthlyPanel delegates numerical preparation to fire.BuildMonthlyPanel.
func BuildMonthlyPanel(assets []AssetSeries, hicp []marketdata.Point) (scenario.Panel, Coverage, error) {
	return fire.BuildMonthlyPanel(assets, hicp)
}

// FitParametric delegates calibration to fire.FitParametric.
func FitParametric(panel scenario.Panel, weights []float64) Fit {
	return fire.FitParametric(panel, weights)
}

// checkWeights reports whether a posted allocation can be read against the
// panel it will be combined with.
//
// A weight vector is defined up to scale, so one that does not sum to 1 is
// simply rescaled (withCentral does it); a vector of the WRONG LENGTH is a
// different portfolio and cannot be repaired, so it is refused rather than
// padded. Padding would silently simulate an allocation nobody asked for, and
// reading it would index past scenario.Panel's asset rows, which is how a
// hand-written POST of two weights against a four-holding panel used to take
// the request's connection down with a stack trace instead of a 400.
func (pr Params) checkWeights(panel *scenario.Panel) error {
	if len(pr.Weights) == 0 || panel == nil {
		return nil // parametric mode: the allocation is not read at all
	}
	if n := len(panel.Returns); len(pr.Weights) != n {
		return fmt.Errorf("allocation carries %d weights but the portfolio holds %d holdings",
			len(pr.Weights), n)
	}
	sum := 0.0
	for _, w := range pr.Weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 {
			return fmt.Errorf("allocation weights must be finite and non-negative")
		}
		sum += w
	}
	if sum <= 0 {
		return fmt.Errorf("allocation weights are all zero")
	}
	return nil
}

// normalize scales weights to sum to 1 (returned unchanged if they sum to 0).
func normalize(w []float64) []float64 {
	sum := 0.0
	for _, x := range w {
		sum += x
	}
	if sum == 0 {
		return w
	}
	out := make([]float64, len(w))
	for i, x := range w {
		out[i] = x / sum
	}
	return out
}
