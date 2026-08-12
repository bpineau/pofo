// The -sweep mode: what each holding's weight buys, and what it costs.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/bpineau/pofo/pkg/compare"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// runSweep prints, for every holding of every portfolio, how the whole
// portfolio behaves as that one line's weight moves across a grid, the other
// lines keeping their relative proportions. It is the evidence behind a
// portfolio file's "sane range" per sleeve: the weight where the worst
// five-year stretch peaks, the one where the drawdown starts deepening, and
// how flat (or not) the answer is in between.
func runSweep(ctx context.Context, c *marketdata.Client, specs []*portfolio.Spec, opt *options, step float64) error {
	if len(specs) == 0 {
		return errors.New("-sweep needs a portfolio (a file or -assets)")
	}
	if step <= 0 || step > 50 {
		return fmt.Errorf("-sweep-step must be a percentage in (0,50], got %g", step)
	}
	var grid []float64
	for w := 0.0; w <= 45+1e-9; w += step {
		grid = append(grid, w/100)
	}
	for _, spec := range specs {
		if err := sweepPortfolio(ctx, c, spec, opt, grid); err != nil {
			return fmt.Errorf("portfolio %s: %w", spec.Name, err)
		}
	}
	return nil
}

func sweepPortfolio(ctx context.Context, c *marketdata.Client, spec *portfolio.Spec, opt *options, grid []float64) error {
	// A "#meta leverage:on" book borrows, and borrowed money is financed at
	// the cash rate plus a spread: exactly what pkg/compare does, because a
	// nil Cash means a flat 0 %/yr and would hand every row of the table free
	// financing. The rate is an annualized percent LEVEL, not a price, so it
	// is fetched raw (Fetch) and never through fetchAsset, whose currency
	// conversion would multiply a yield by an FX cross.
	var cashRate *marketdata.Series
	if spec.Leverage {
		cr, err := c.Fetch(ctx, "^IRX", opt.start)
		if err != nil {
			log.Printf("warning: financing rate ^IRX unavailable (%v), leverage financed at 0 %%", err)
		} else {
			cashRate = cr
		}
	}
	p, err := portfolio.Build(spec, portfolio.BuildOptions{
		Fetch: func(id string) (*marketdata.Series, error) {
			return fetchAsset(ctx, c, id, opt)
		},
		Cash:         cashRate,
		BorrowSpread: 1.0, // default: cash + 1 %/yr, as in pkg/compare
		BaseCurrency: opt.currency,
	})
	if err != nil {
		return err
	}
	rebalance := spec.RebalanceDays
	if rebalance < 0 {
		rebalance = opt.rebalance
	}
	sim, err := portfolio.Simulate(p, rebalance)
	if err != nil {
		return err
	}
	holdings, err := compare.Sweep(p, compare.SweepOptions{Weights: grid, RebalanceDays: rebalance})
	if err != nil {
		return err
	}

	fmt.Printf("\n%s: one line at a time, the others keeping their proportions\n", spec.Name)
	fmt.Printf("%s → %s, rebalanced every %d days, %s\n",
		sim.Dates[0].Format("2006-01-02"), sim.Dates[len(sim.Dates)-1].Format("2006-01-02"),
		rebalance, opt.currency)
	for _, h := range holdings {
		const row = "  %8s %9s %9s %7s %9s %7s %9s%s\n"
		fmt.Printf("\n%s (written %.1f %%)\n", h.ID, h.Written*100)
		fmt.Printf(row, "weight", "CAGR", "vol", "Sharpe", "maxDD", "TTR", "worst5y", "")
		for _, pt := range h.Points {
			mark := ""
			if pt.Written {
				mark = "  <- written"
			}
			fmt.Printf(row,
				fmt.Sprintf("%.1f %%", pt.Weight*100),
				pct(pt.Stats.CAGR), pct(pt.Stats.Volatility),
				fmt.Sprintf("%.2f", pt.Stats.Sharpe),
				pct(pt.Stats.MaxDrawdown), ttr(pt.Stats), pct(pt.Worst5y), mark)
		}
	}
	fmt.Println()
	return nil
}

// pct renders a fraction as a percentage, "-" when undefined.
func pct(v float64) string {
	if math.IsNaN(v) {
		return "-"
	}
	return fmt.Sprintf("%.2f %%", v*100)
}

// ttr renders the longest recovery in years, "+" when still underwater at the
// end of the window.
func ttr(s metrics.Stats) string {
	if s.TTRDays == 0 {
		return "-"
	}
	out := fmt.Sprintf("%.1f y", float64(s.TTRDays)/365.25)
	if s.TTROngoing {
		out += "+"
	}
	return out
}
