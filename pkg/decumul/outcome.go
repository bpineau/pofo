package decumul

import (
	"math"
	"slices"

	"github.com/bpineau/pofo/pkg/metrics"
)

// Outcome bundles the headline decumulation statistics across an Ensemble.
// All wealth figures are real euros; rates are fractions.
//
// Every per-path statistic here is read over the household's LIVED window.
// Without a Plan.Lifetime that window is the whole horizon and the figures are
// what they always were; with one, the terminal quantiles become the ESTATE
// distribution and the path statistics stop at the drawn death instead of
// counting a frozen post-mortem tail.
type Outcome struct {
	RuinProb              float64 // share of paths that ran out
	TerminalP5            float64 // 5th-percentile terminal wealth (0 for ruined)
	TerminalP50           float64 // median terminal wealth
	MedianYearsUnderwater float64 // median years spent below the prior real high
	Worst10yCAGR          float64 // worst (min) rolling 10-year real CAGR across all paths
	Worst10yP5            float64 // 5th-percentile of paths' worst 10-year real CAGR (robust)
	CDaR                  float64 // mean of the worst 5% path drawdowns (0.30 = 30%)
	MedianCumTax          float64 // median cumulative tax paid over a path
	EffectiveTaxRate      float64 // median per-path tax / gross withdrawn (0.15 = 15%)
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
	taxes := make([]float64, len(e.Paths))
	taxRates := make([]float64, len(e.Paths))
	worsts := make([]float64, 0, len(e.Paths))
	ruined := 0
	for i, p := range e.Paths {
		lived := p.Wealth[:p.end()+1]
		terminals[i] = lived[len(lived)-1]
		if p.Ruined {
			ruined++
		}
		under, dd := pathPeakStats(lived)
		underwater[i] = float64(under)
		maxDDs[i] = dd
		taxes[i] = p.TaxPaid
		if gross := p.Withdrawn + p.TaxPaid; gross > 0 {
			taxRates[i] = p.TaxPaid / gross
		}
		if c, ok := worst10y(lived); ok {
			worsts = append(worsts, c)
		}
	}
	o.RuinProb = float64(ruined) / float64(len(e.Paths))
	q := metrics.Quantiles(terminals, 0.05, 0.50)
	o.TerminalP5, o.TerminalP50 = q[0], q[1]
	o.MedianYearsUnderwater = metrics.Quantiles(underwater, 0.50)[0]
	// Both worst-decade figures read the same sample, and both stay empty when
	// no path has a full decade to show. The minimum is taken over that sample
	// rather than against a zero seed: a plan whose every decade grew would
	// otherwise report a worst decade of 0, a decade no path ever lived.
	if len(worsts) > 0 {
		o.Worst10yCAGR = slices.Min(worsts)
		o.Worst10yP5 = metrics.Quantiles(worsts, 0.05)[0]
	}
	o.MedianCumTax = metrics.Quantiles(taxes, 0.50)[0]
	o.EffectiveTaxRate = metrics.Quantiles(taxRates, 0.50)[0]
	o.CDaR = conditionalTail(maxDDs, 0.05)
	return o
}

// pathPeakStats walks a wealth path once for the two statistics that read the
// same running peak: under is the number of points strictly below the prior
// real high, maxDD the deepest peak-to-trough loss (0.30 = 30%). A point at a
// new high is neither, which is what lets the division be skipped there.
func pathPeakStats(w []float64) (under int, maxDD float64) {
	peak := w[0]
	for _, v := range w {
		if v >= peak {
			peak = v
			continue
		}
		under++
		if peak > 0 {
			if d := 1 - v/peak; d > maxDD {
				maxDD = d
			}
		}
	}
	return under, maxDD
}

// worst10y is the lowest 10-year real CAGR found in the wealth path; ok is
// false when no decade window has a positive starting wealth (a path shorter
// than 11 points, or one already at zero throughout). A decade that ends at
// zero counts as its realised -100% return rather than as an undefined value:
// the -1 then means "lost everything over this decade", and windows that start
// after ruin (zero starting wealth) are skipped instead of conflated with it.
func worst10y(w []float64) (float64, bool) {
	// x^0.1 is increasing over the non-negative growth ratios, so the worst
	// decade is the one with the smallest ratio: the window loop compares
	// ratios and the (expensive) root is taken once, on the winner, instead of
	// once per window. Same winner, same arithmetic on it, same bits.
	ratio, ok := 0.0, false
	for i := 0; i+10 < len(w); i++ {
		if w[i] <= 0 {
			continue // window starts after ruin: undefined, skip
		}
		r := max(w[i+10], 0) / w[i] // end == 0 -> 0 -> -1 below (total loss realised)
		if !ok || r < ratio {
			ratio, ok = r, true
		}
	}
	if !ok {
		return 0, false
	}
	return math.Pow(ratio, 0.1) - 1, true
}

// conditionalTail averages the worst frac share of dds (already losses).
func conditionalTail(dds []float64, frac float64) float64 {
	if len(dds) == 0 {
		return 0
	}
	n := int(frac * float64(len(dds)))
	if n < 1 {
		n = 1
	}
	sum := 0.0
	for _, d := range metrics.TopK(dds, n) {
		sum += d
	}
	return sum / float64(n)
}
