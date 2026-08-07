package web

import (
	"fmt"
	"math"
	"sync"

	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// The risk-based guardrail (Kitces & Tharp, industrialised by Morningstar)
// needs one thing the 2006 rule does not: a table of what is still safe as the
// horizon shortens. This file builds that table.
//
// Two decisions make it cheap and honest.
//
// It is SCALE-INVARIANT. The table is solved on a pension-free plan of unit
// capital, so one rate per remaining horizon applies at any wealth level; the
// pensions the household is owed enter on the other side of the comparison, as
// discounted wealth (decumul.RiskGuardrails reads the rate on portfolio + PV of
// the cashflows still to come, the same total-wealth view ABW takes). Without
// that split the safe rate would depend on the pension-to-capital ratio, which
// moves along every path, and no table could exist.
//
// It is solved under the model the reader SELECTED in the strip, the same lens
// every detail section runs on, and not per column of that strip: a guardrail
// table is written down once and lived with, so what a table costs when the
// world turns out harsher than the model it was written on is exactly the
// reading the strip then offers. Planning on the broad-sample century
// therefore buys a conservative table; planning on a rosy fitted history buys
// a table that will let spending ratchet up and then be caught out, which is
// the honest depiction of that choice rather than a flaw of the rule. That is deliberate: a guardrail table is
// written down once and lived with, so seeing how a table written under your
// central case behaves inside the broad-sample century or a lost decade is the
// useful reading, not a table that quietly rewrites itself per world.

// riskAnchors is how many horizons are actually solved; the table interpolates
// between them. The safe rate is a smooth, monotone function of the horizon
// left, so five anchors reproduce it to a few basis points at a twentieth of
// the cost of solving all forty.
const riskAnchors = 5

// riskSolvePaths is the path count of the table's solves. The table is a
// planning input, not a headline: a few basis points of Monte-Carlo noise on
// the band's centre is invisible next to the +-20 % band itself.
const riskSolvePaths = 600

var riskTableCache sync.Map // key -> []float64

// safeRateTable returns the risk-based guardrail's band centre for each year of
// the horizon: the withdrawal rate (share of total wealth) that still meets the
// ruin target over the years remaining. Cached per assumption set, since every
// section of the page rebuilds the same plan.
func (pr Params) safeRateTable() []float64 {
	if pr.Years <= 0 {
		return nil
	}
	target := pr.TargetRuin
	if target <= 0 {
		target = 0.05
	}
	mu, sigma, df := pr.tableAssumptions()
	months := 0
	if pr.panel != nil {
		months = pr.panel.Periods()
	}
	key := fmt.Sprintf("%d|%.5f|%.5f|%.2f|%.4f|%.4f|%.3f|%.4f|%d|%s|%d|%v",
		pr.Years, mu, sigma, df, target, pr.TaxRate,
		pr.BufferYears, pr.BufferReturn, pr.BufferStopYear,
		pr.central(), months, pr.Weights)
	if v, ok := riskTableCache.Load(key); ok {
		return v.([]float64)
	}

	// Solve the anchors, longest horizon first.
	horizons := make([]int, 0, riskAnchors)
	rates := make([]float64, 0, riskAnchors)
	for i := range riskAnchors {
		n := pr.Years - pr.Years*i/riskAnchors
		if i == riskAnchors-1 {
			n = max(1, pr.Years/(2*riskAnchors)) // the short end, where the rate runs away
		}
		if n < 1 {
			n = 1
		}
		if len(horizons) > 0 && n == horizons[len(horizons)-1] {
			continue
		}
		horizons = append(horizons, n)
		rates = append(rates, pr.safeRateAt(n, target))
	}

	out := make([]float64, pr.Years)
	for k := range out {
		out[k] = interpRate(horizons, rates, pr.Years-k)
	}
	riskTableCache.Store(key, out)
	return out
}

// tableAssumptions is the return model the table is solved on: the blended,
// CAPE-anchored central case when the handler stamped it (the same one the
// page's central column uses, so the household's table cannot be more
// optimistic than the plan it guards), the raw sliders otherwise.
func (pr Params) tableAssumptions() (mu, sigma, df float64) {
	if pr.tableSigma > 0 {
		return pr.tableMu, pr.tableSigma, pr.tableDf
	}
	return pr.Mu, pr.Sigma, pr.Df
}

// withCentral stamps the blended central assumptions onto the params, so every
// endpoint solves the guardrail table on the same model the page plans with.
func (pr Params) withCentral(panel *scenario.Panel) Params {
	pr.tableMu, pr.tableSigma, pr.tableDf = centralParams(pr, panel)
	pr.panel = panel
	return pr
}

// safeRateAt solves the safe withdrawal rate for one remaining horizon, on the
// unit-capital, pension-free, fixed-rule plan the table is defined on.
func (pr Params) safeRateAt(years int, target float64) float64 {
	const unit = 1e6
	mu, sigma, df := pr.tableAssumptions()
	src := scenario.Source(scenario.ParametricSource{Mu: mu, Sigma: sigma, Df: df, Periods: years})
	if pr.tableSigma > 0 { // server-side context available: use the selected lens
		src = pr.detailSource(pr.panel, years)
	}
	p := decumul.Plan{
		Capital:    unit,
		NeedAnnual: unit * 0.04,
		Years:      years,
		Buffer:     decumul.BufferSleeve{Years: pr.BufferYears, RealReturn: pr.BufferReturn, RefillStopYear: pr.BufferStopYear},
		Tax:        decumul.CTOFlatTax{Rate: pr.TaxRate},
		Source:     src,
	}
	// A one-year horizon has no risk to speak of: the whole pot is spendable,
	// and the solver's upper bound would clip an answer of 100 %.
	safe := p.Solve(target, decumul.WithdrawalAxis(0, unit*0.60), riskSolvePaths, simWorkers, 11)
	return safe / unit
}

// interpRate reads the table's anchors (horizons descending) at one remaining
// horizon, linearly between the two surrounding anchors and flat outside them.
func interpRate(horizons []int, rates []float64, n int) float64 {
	if len(horizons) == 0 {
		return 0
	}
	if n >= horizons[0] {
		return rates[0]
	}
	for i := 1; i < len(horizons); i++ {
		if n >= horizons[i] {
			lo, hi := horizons[i], horizons[i-1]
			t := float64(n-lo) / float64(hi-lo)
			return rates[i] + t*(rates[i-1]-rates[i])
		}
	}
	return rates[len(rates)-1]
}

// pvRate is the real rate the risk-based sensor discounts future cashflows
// with: the central geometric return, kept inside sane bounds so an extreme
// slider cannot turn a pension into a fortune or a rounding error.
func (pr Params) pvRate() float64 {
	mu, sigma, _ := pr.tableAssumptions()
	g := mu - sigma*sigma/2
	return math.Min(0.06, math.Max(0.005, g))
}
