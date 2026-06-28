package decumul

import (
	"math"
	"sort"

	"github.com/bpineau/pofo/pkg/metrics"
)

// Outcome bundles the headline decumulation statistics across an Ensemble.
// All wealth figures are real euros; rates are fractions.
type Outcome struct {
	RuinProb              float64 // share of paths that ran out
	TerminalP5            float64 // 5th-percentile terminal wealth (0 for ruined)
	TerminalP50           float64 // median terminal wealth
	MedianYearsUnderwater float64 // median years spent below the prior real high
	Worst10yCAGR          float64 // worst rolling 10-year real CAGR across paths
	CDaR                  float64 // mean of the worst 5% path drawdowns (0.30 = 30%)
}

// Outcome computes the bundle.
func (e Ensemble) Outcome() Outcome {
	var o Outcome
	if len(e.Paths) == 0 {
		return o
	}
	terminals := make([]float64, len(e.Paths))
	underwater := make([]float64, len(e.Paths))
	maxDDs := make([]float64, len(e.Paths))
	ruined, worst := 0, 0.0
	for i, p := range e.Paths {
		terminals[i] = p.Wealth[len(p.Wealth)-1]
		if p.Ruined {
			ruined++
		}
		underwater[i] = float64(yearsUnderwater(p.Wealth))
		maxDDs[i] = pathMaxDD(p.Wealth)
		if c := worst10y(p.Wealth); c < worst {
			worst = c
		}
	}
	o.RuinProb = float64(ruined) / float64(len(e.Paths))
	q := metrics.Quantiles(terminals, 0.05, 0.50)
	o.TerminalP5, o.TerminalP50 = q[0], q[1]
	o.MedianYearsUnderwater = metrics.Quantiles(underwater, 0.50)[0]
	o.Worst10yCAGR = worst
	o.CDaR = conditionalTail(maxDDs, 0.05)
	return o
}

// yearsUnderwater counts entries strictly below the running peak.
func yearsUnderwater(w []float64) int {
	peak, n := w[0], 0
	for _, v := range w {
		if v >= peak {
			peak = v
		} else {
			n++
		}
	}
	return n
}

// pathMaxDD is the deepest peak-to-trough loss of a wealth path (0.30 = 30%).
func pathMaxDD(w []float64) float64 {
	peak, dd := w[0], 0.0
	for _, v := range w {
		if v > peak {
			peak = v
		}
		if peak > 0 {
			if d := 1 - v/peak; d > dd {
				dd = d
			}
		}
	}
	return dd
}

// worst10y is the lowest 10-year real CAGR found in the wealth path, or -1
// (worst possible) when the path hits zero within a 10-year window.
func worst10y(w []float64) float64 {
	worst := 0.0
	for i := 0; i+10 < len(w); i++ {
		if w[i] <= 0 || w[i+10] <= 0 {
			return -1
		}
		c := math.Pow(w[i+10]/w[i], 0.1) - 1
		if c < worst {
			worst = c
		}
	}
	return worst
}

// conditionalTail averages the worst frac share of dds (already losses).
func conditionalTail(dds []float64, frac float64) float64 {
	if len(dds) == 0 {
		return 0
	}
	s := append([]float64(nil), dds...)
	sort.Sort(sort.Reverse(sort.Float64Slice(s)))
	n := int(frac * float64(len(s)))
	if n < 1 {
		n = 1
	}
	sum := 0.0
	for _, d := range s[:n] {
		sum += d
	}
	return sum / float64(n)
}
