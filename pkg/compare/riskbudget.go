package compare

import (
	"fmt"
	"math"
	"sort"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/report"
	"github.com/bpineau/pofo/pkg/suggest"
)

// The risk-budget block: what share of the portfolio's capital, risk and
// realized return each asset class accounts for, side by side.
//
// It answers a question the composition pies cannot. A pie says where the
// money sits; this says where the RISK sits, and the two routinely disagree by
// a factor of three, because a class's share of risk depends on its volatility
// and on how it moves with the rest of the book, not on its weight. A
// "balanced across regimes" allocation is usually balanced in capital only:
// the equity sleeve of a four-engine portfolio commonly carries two thirds of
// its variance. That gap is the finding, so the block is built to make it
// visible at a glance rather than to be derived by the reader.
//
// The arithmetic is exact (metrics.Attribute, an Euler decomposition of
// variance and a plain sum for the return), and it runs on the simulation's
// per-holding MONTHLY contributions, so drifting weights, rebalancing and
// embedded leverage are already in it, without the asynchronous-pricing bias a
// daily view carries across time zones. Only the grouping is approximate, and in one
// documented way: a stacked fund's risk is attributed to its classes pro rata
// of its look-through capital split, the same split the pies use, which
// understates the equity leg of a 90/60 (equity carries more of that fund's
// variance than its notional share). Holdings and classes are otherwise
// treated identically.

// riskBudgetRows builds the capital/risk/return table for one portfolio,
// grouped by look-through asset class. It returns nil when the simulation is
// too short to estimate a covariance or when no catalog metadata is available
// (the classes would all be "unknown", which teaches nothing).
func riskBudgetRows(r *column, meta map[string]suggest.Meta) []report.RiskRow {
	if r == nil || r.sim == nil || len(meta) == 0 {
		return nil
	}
	// MONTHLY contributions, not daily: the book holds funds priced at
	// different market closes (a US mutual-fund NAV against a European
	// listing), and asynchronous pricing depresses measured cross-correlation,
	// which would hand each holding a share of variance closer to its own
	// standalone volatility than to its real co-movement with the book. Folding
	// to months removes most of that bias, at the cost of fewer observations,
	// which a covariance over this many holdings can afford.
	_, monthly := r.sim.MonthlyContributions()
	att, err := metrics.Attribute(monthly)
	if err != nil || len(att.Risk) != len(r.p.Assets) {
		return nil
	}

	capital := map[string]float64{}
	risk := map[string]float64{}
	ret := map[string]float64{}
	for i, a := range r.p.Assets {
		for class, share := range classSplit(a, meta) {
			capital[class] += a.Weight * share
			risk[class] += att.Risk[i] * share
			ret[class] += att.Return[i] * share
		}
	}

	classes := make([]string, 0, len(risk))
	for c := range risk {
		classes = append(classes, c)
	}
	// Largest risk share first: the block exists to show what drives the book.
	sort.Slice(classes, func(i, j int) bool {
		if risk[classes[i]] != risk[classes[j]] {
			return risk[classes[i]] > risk[classes[j]]
		}
		return classes[i] < classes[j]
	})

	// The track is scaled to the largest share it must show, so the bar and
	// the reference tick share one axis and the eye compares them directly.
	scale := 0.0
	for _, c := range classes {
		scale = math.Max(scale, math.Max(risk[c], capital[c]))
	}
	if scale <= 0 {
		return nil
	}

	rows := make([]report.RiskRow, 0, len(classes))
	for _, c := range classes {
		rows = append(rows, report.RiskRow{
			Label:       prettyClass(c),
			Capital:     pctShare(capital[c]),
			Risk:        pctShare(risk[c]),
			Return:      pctShare(ret[c]),
			RiskWidth:   100 * math.Max(0, risk[c]) / scale,
			CapitalMark: 100 * capital[c] / scale,
			Negative:    risk[c] < 0,
		})
	}
	return rows
}

// pctShare formats a fraction as a percent share, at the one decimal the
// comparison deserves and no more.
func pctShare(v float64) string { return fmt.Sprintf("%.1f %%", 100*v) }

// classSplit is one asset's look-through asset-class decomposition, as
// fractions of that asset summing to 1: the per-holding form of
// suggest.AssetClassSplit, so the block and the composition pie tell the same
// story about what a fund is made of.
func classSplit(a portfolio.Asset, meta map[string]suggest.Meta) map[string]float64 {
	base, _ := marketdata.SplitSim(a.ID)
	m, _, ok := metaFor(meta, a.ID)
	switch {
	case !ok:
		return map[string]float64{suggest.BucketUnknown: 1}
	case len(m.Exposures) > 0:
		total := 0.0
		for _, notional := range m.Exposures {
			total += notional
		}
		if total <= 0 {
			return map[string]float64{suggest.BucketUnknown: 1}
		}
		out := make(map[string]float64, len(m.Exposures))
		for class, notional := range m.Exposures {
			out[class] = notional / total
		}
		return out
	case m.AssetClass != "":
		return map[string]float64{m.AssetClass: 1}
	default:
		_ = base
		return map[string]float64{suggest.BucketUnknown: 1}
	}
}

// riskBudgetNote is the caption the block carries. The three columns are NOT
// equally trustworthy and saying so is the point: capital is a decision, risk
// is a statistic that holds up out of sample, realized return is the one that
// does not.
func riskBudgetNote(start, end string) string {
	return fmt.Sprintf(
		"Risk budget over %s→%s: share of capital (your choice), of variance "+
			"(Euler decomposition of the simulated contributions) and of realized return. "+
			"Read the first two together: a class whose risk share far exceeds its capital "+
			"share is what actually drives the portfolio. Read the third with care and never "+
			"as an efficiency ratio: covariances estimate far better than means, so the risk "+
			"split is comparatively stable across windows while the return split moves with "+
			"every regime, and an insurance sleeve (gold, long duration) is MEANT to show a "+
			"small or negative return share against a real risk share. A stacked fund's shares "+
			"are split across its classes pro rata of notional.",
		start, end)
}
