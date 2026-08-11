package compare

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// SweepOptions configures Sweep.
type SweepOptions struct {
	// Weights is the grid each holding's weight is moved across, as
	// fractions. Nil means DefaultSweepGrid.
	Weights []float64
	// RebalanceDays is passed to portfolio.Simulate, so the sweep answers
	// the question in the file's own terms; negative means the caller's
	// default (90).
	RebalanceDays int
}

// DefaultSweepGrid is the weight grid Sweep uses when none is given: 0 to
// 45 % in 5-point steps, wide enough to reach both sides of any sane sleeve
// and coarse enough to stay readable.
var DefaultSweepGrid = []float64{0, .05, .10, .15, .20, .25, .30, .35, .40, .45}

// SweepPoint is one simulated allocation: the swept holding at Weight, the
// others keeping their relative proportions.
type SweepPoint struct {
	Weight  float64 // the swept holding's weight, as a fraction
	Written bool    // this is the weight the file actually holds
	Stats   metrics.Stats
	Worst5y float64 // worst rolling 5-year CAGR, NaN when the window is too short
}

// SweepHolding is one holding's whole sweep.
type SweepHolding struct {
	ID      string
	Symbol  string
	Written float64 // the weight written in the file, as a fraction
	Points  []SweepPoint
}

// Sweep answers "what does each line's weight buy, and what does it cost".
// For every holding it re-runs the simulation at a range of weights for that
// line, the OTHER lines keeping their relative proportions (the only
// renormalization that needs no further justification: funding the move from
// one named sleeve, or from cash, would be a different experiment). The
// written weight is inserted into the grid and flagged, so the file's own
// choice can be read next to its neighbours.
//
// It is what produces a portfolio file's "sane range" evidence: one call,
// one table per sleeve, showing where CAGR, volatility, Sharpe, drawdown,
// recovery time and the worst five-year stretch turn.
//
// The simulation is the real one (portfolio.Simulate, with the portfolio's
// rebalancing period, fees and flows), not a blend of daily returns: the
// question is what would have happened had another number been written in
// the file.
func Sweep(p *portfolio.Portfolio, opt SweepOptions) ([]SweepHolding, error) {
	if len(p.Assets) < 2 {
		return nil, fmt.Errorf("sweep needs at least two holdings, got %d", len(p.Assets))
	}
	grid := opt.Weights
	if len(grid) == 0 {
		grid = DefaultSweepGrid
	}
	rebalance := opt.RebalanceDays
	if rebalance < 0 {
		rebalance = 90
	}

	out := make([]SweepHolding, 0, len(p.Assets))
	for i, a := range p.Assets {
		h := SweepHolding{ID: a.ID, Symbol: a.Symbol, Written: a.Weight}
		for _, w := range weightGrid(grid, a.Weight) {
			cand, ok := reweighted(p, i, w)
			if !ok {
				continue
			}
			sim, err := portfolio.Simulate(cand, rebalance)
			if err != nil {
				return nil, fmt.Errorf("sweep %s at %.0f %%: %w", a.ID, w*100, err)
			}
			st, err := metrics.Compute(sim.Dates, sim.Index)
			if err != nil {
				return nil, fmt.Errorf("sweep %s at %.0f %%: %w", a.ID, w*100, err)
			}
			pt := SweepPoint{Weight: w, Written: w == a.Weight, Stats: st, Worst5y: math.NaN()}
			if worst, _, _, _, ok := metrics.RollingCAGR(sim.Dates, sim.Index, 5); ok {
				pt.Worst5y = worst
			}
			h.Points = append(h.Points, pt)
		}
		out = append(out, h)
	}
	return out, nil
}

// weightGrid merges the written weight into the grid, in order and without
// duplicating a point the grid already covers.
func weightGrid(grid []float64, written float64) []float64 {
	out := make([]float64, 0, len(grid)+1)
	inserted := false
	for _, w := range grid {
		if !inserted && written < w-1e-9 {
			out = append(out, written)
			inserted = true
		}
		if math.Abs(w-written) < 1e-9 {
			inserted = true
		}
		out = append(out, w)
	}
	if !inserted {
		out = append(out, written)
	}
	return out
}

// reweighted copies p with asset i at weight w, the other holdings scaled to
// hold the portfolio's TOTAL weight constant while preserving their relative
// proportions. The total, not 1: under "#meta leverage:on" the written
// weights deliberately sum to more (or less) than 100 %, and renormalizing
// them would quietly sweep the leverage too. It reports false when the move
// is impossible: nothing left to absorb the residual, or a weight the total
// cannot cover.
func reweighted(p *portfolio.Portfolio, i int, w float64) (*portfolio.Portfolio, bool) {
	total, rest := 0.0, 0.0
	for j, a := range p.Assets {
		total += a.Weight
		if j != i {
			rest += a.Weight
		}
	}
	if w < 0 || w > total {
		return nil, false
	}
	if rest <= 0 && w < total {
		return nil, false // nothing to scale the residual onto
	}
	cp := *p
	cp.Assets = make([]portfolio.Asset, len(p.Assets))
	copy(cp.Assets, p.Assets)
	scale := 0.0
	if rest > 0 {
		scale = (total - w) / rest
	}
	for j := range cp.Assets {
		if j == i {
			cp.Assets[j].Weight = w
			continue
		}
		cp.Assets[j].Weight = p.Assets[j].Weight * scale
	}
	return &cp, true
}
