package analyze

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// borrowSpread is the spread over the cash rate, in percent per year, a
// levered spec pays when it sets no "#meta borrow-spread": the CLI's default.
const borrowSpread = 1.0

// financingRate is the cash rate a levered portfolio is financed at, as in
// the CLI: the 13-week US Treasury bill, an annualized percent level.
const financingRate = "^IRX"

// Composition is the look-through of a portfolio, as pkg/suggest computes it
// from the bundled catalog. Every value is a FRACTION of the portfolio's
// capital, except Sectors (fractions of the equity sleeve, whose size is
// Equity) and Duration (years per unit of capital).
type Composition struct {
	AssetClass   map[string]float64           // suggest.AssetClassSplit
	Geography    map[string]float64           // suggest.GeographySplit
	Currency     map[string]float64           // suggest.CurrencySplit: currency EXPOSURE, not quote currency
	Sectors      map[string]float64           // suggest.EquitySectorSplit, of the equity share; Equity is that share
	Equity       float64                      // the equity sleeve's notional, a fraction of capital
	Duration     suggest.DurationLedger       // suggest.DurationSplit
	Coverage     map[suggest.Category]float64 // suggest.Coverage under suggest.RegimeFramework
	Unclassified float64                      // weight of the holdings the catalog does not know
}

// PortfolioStudy is a portfolio dissected: the simulation, its statistics,
// each holding's own study on the same window, and what ties them. Returns,
// weights and shares are FRACTIONS, as in AssetStudy.
type PortfolioStudy struct {
	Spec        *portfolio.Spec        // the spec studied (a copy with Sim set under Options.Sim)
	Portfolio   *portfolio.Portfolio   // as built: series fetched, fees looked up
	Sim         *portfolio.SimResult   // the simulation; its window is the study's
	Stats       metrics.Stats          // on Sim.Index; Beta and CWARP set when the benchmark overlaps it
	Years       []metrics.PeriodReturn // calendar years of Sim.Index
	Months      []metrics.PeriodReturn // calendar months of Sim.Index
	Drawdowns   []metrics.Episode      // drawdown episodes of Sim.Index, chronological
	Relative    *metrics.Relative      // vs Options.Benchmark; nil without one or without overlap
	Holdings    []AssetStudy           // in Spec order, each on the simulation window
	Aligned     *marketdata.Aligned    // the holdings on the simulation calendar, IDs as written
	Correlation [][]float64            // of the holdings' daily returns on Aligned, Spec order
	Attribution metrics.Attribution    // risk and return shares from MONTHLY contributions, Spec order
	Composition Composition            // look-through, from the catalog
	Warnings    []string               // the portfolio's, the simulation's and every holding's (prefixed "ID: ")
}

// Portfolio builds, simulates and dissects a portfolio.
//
// The window is Simulate's: from the first date every holding quotes to the
// last date they all still quote, inside [opt.From, opt.To]. Each
// Holdings[i] is studied on THAT window, on its own quoting calendar, so the
// holdings' statistics compare with each other and with the portfolio's;
// Asset alone studies an asset's longest window.
//
// Each holding is fetched through portfolio.SimFetchID with the spec's Sim
// flag, which opt.Sim turns on (never off, like the CLI's -simulate): its
// backcast-extended history when the bundle has one, its real quotes
// otherwise. A TER the spec does not declare is looked up with src.Fees on
// the bare identifier; TERs are informational, already net in prices.
//
// The rebalancing period is the spec's own RebalanceDays when it sets one
// (0 or more, "#meta rebalance:N"), else opt.Rebalance, else
// DefaultRebalance: the precedence the CLI applies. A levered spec is
// financed at ^IRX (fetched from src, held flat before its history, 0 %/yr
// with a warning when it cannot be fetched) plus its borrow spread, 1 %/yr
// by default. Spec.Optimize and Spec.Currencies are not acted on: the
// written weights are studied, in opt.Currency, and a warning says so.
//
// Correlation is the Pearson correlation of the holdings' daily returns on
// Aligned (metrics.CorrelationMatrix). Attribution is metrics.Attribute over
// the simulation's MONTHLY contributions (SimResult.MonthlyContributions),
// the reading the comparison report's risk budget makes: folding to months
// removes most of the bias asynchronous closes (a US fund against a European
// listing) put into a daily covariance. Composition is the catalog
// look-through of the portfolio's weights.
//
// Errors are those of Asset (an unknown identifier satisfies
// errors.Is(err, marketdata.ErrUnknownIdentifier)), plus a spec Build or
// Simulate refuses. Everything else, down to a holding with too few quotes
// in the window for statistics of its own, is a warning.
func Portfolio(ctx context.Context, src Source, spec *portfolio.Spec, opt Options) (*PortfolioStudy, error) {
	if spec == nil {
		return nil, errors.New("analyze: nil spec")
	}
	if err := opt.check(); err != nil {
		return nil, err
	}
	eff := *spec
	eff.Sim = spec.Sim || opt.Sim
	bench, benchWarnings, err := benchmark(ctx, src, opt)
	if err != nil {
		return nil, err
	}

	var warnings []string
	if eff.Optimize != nil {
		warnings = append(warnings, "#meta optimize is not run here: the written weights are studied")
	}
	if len(eff.Currencies) > 0 {
		warnings = append(warnings, fmt.Sprintf("#meta currencies is not expanded here: the portfolio is studied in %s", currencyLabel(opt.Currency)))
	}
	fo := opt.fetchOptions()
	build := portfolio.BuildOptions{
		Fetch: func(id string) (*marketdata.Series, error) {
			s, err := src.FetchExtended(ctx, id, fo)
			if err != nil {
				return nil, err
			}
			return marketdata.Trim(s, opt.From, opt.To), nil
		},
		Fees: func(id string) (float64, bool) {
			base, _ := marketdata.SplitSim(id)
			return src.Fees(ctx, base)
		},
		BorrowSpread: borrowSpread,
		BaseCurrency: opt.Currency,
	}
	if eff.Leverage {
		// A rate is a percent LEVEL: fetched as is, never converted.
		cash, err := src.FetchExtended(ctx, financingRate, marketdata.FetchOptions{From: opt.From})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("financing rate %s unavailable (%v), leverage financed at 0 %%", financingRate, err))
		} else {
			build.Cash = cash
		}
	}
	p, err := portfolio.Build(&eff, build)
	if err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}
	sim, err := portfolio.Simulate(p, rebalanceDays(&eff, opt))
	if err != nil {
		return nil, fmt.Errorf("analyze: portfolio %s: %w", eff.Name, err)
	}
	warnings = slices.Concat(p.Warnings, warnings, sim.Warnings, ruin(p, sim), benchWarnings)

	stats, err := metrics.Compute(sim.Dates, sim.Index)
	if err != nil {
		return nil, fmt.Errorf("analyze: portfolio %s: %w", eff.Name, err)
	}
	ps := &PortfolioStudy{Spec: &eff, Portfolio: p, Sim: sim, Stats: stats}
	ps.Years, ps.Months, ps.Drawdowns = calendar(sim.Dates, sim.Index)
	ps.Relative = relative(&ps.Stats, sim.Dates, sim.Index, bench)
	if bench != nil && ps.Relative == nil {
		warnings = append(warnings, noOverlap(opt.Benchmark))
	}

	start, end := sim.Dates[0], sim.Dates[len(sim.Dates)-1]
	series := make([]*marketdata.Series, len(p.Assets))
	for i, a := range p.Assets {
		series[i] = a.Series
		h, err := study(a.ID, marketdata.Trim(a.Series, start, end), bench)
		if err != nil {
			h.Warnings = append(h.Warnings, fmt.Sprintf("no statistics on the portfolio's window: %v", err))
		} else if bench != nil && h.Relative == nil {
			h.Warnings = append(h.Warnings, noOverlap(opt.Benchmark))
		}
		ps.Holdings = append(ps.Holdings, *h)
		for _, w := range h.Warnings {
			warnings = append(warnings, a.ID+": "+w)
		}
	}

	al, err := marketdata.AlignSeries(series, start, end)
	if err != nil {
		return nil, fmt.Errorf("analyze: portfolio %s: %w", eff.Name, err)
	}
	for i, a := range p.Assets {
		al.IDs[i] = a.ID
	}
	ps.Aligned = al
	ps.Correlation = metrics.CorrelationMatrix(al.Returns())

	_, monthly := sim.MonthlyContributions()
	if att, err := metrics.Attribute(monthly); err == nil {
		ps.Attribution = att
	} else {
		warnings = append(warnings, fmt.Sprintf("no risk and return attribution: %v", err))
	}

	comp, err := composition(p.Assets)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("asset catalog unreadable, no composition: %v", err))
	}
	ps.Composition = comp
	ps.Warnings = warnings
	return ps, nil
}

// rebalanceDays is the period a spec is simulated with: its own directive,
// else the options', else the default; a negative option means never (0).
func rebalanceDays(spec *portfolio.Spec, opt Options) int {
	switch {
	case spec.RebalanceDays >= 0:
		return spec.RebalanceDays
	case opt.Rebalance == 0:
		return DefaultRebalance
	case opt.Rebalance < 0:
		return 0
	}
	return opt.Rebalance
}

// ruin is the warning of a simulation that wiped the capital out, worded as
// the comparison report words it.
func ruin(p *portfolio.Portfolio, sim *portfolio.SimResult) []string {
	if !sim.Ruined {
		return nil
	}
	cause := "the leveraged exposure exhausted the net value"
	if p.Withdraw.Active() && !p.Leverage {
		cause = "withdrawals exhausted the capital"
	}
	return []string{fmt.Sprintf("capital wiped out on %s: %s; the series stops there",
		sim.Dates[len(sim.Dates)-1].Format(time.DateOnly), cause)}
}

// currencyLabel names an evaluation currency in a warning.
func currencyLabel(c string) string {
	if c == "" {
		return "each holding's native currency"
	}
	return c
}

// catalog is the bundled asset catalog keyed by id and ISIN, decoded once.
var catalog = sync.OnceValues(func() (map[string]suggest.Meta, error) {
	return suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
})

// holdings adapts a portfolio's assets to suggest holdings exactly as the
// comparison report does: the identifier's SIM suffix dropped, its catalog
// record looked up under its canonical id, then under the bare identifier.
func holdings(assets []portfolio.Asset, meta map[string]suggest.Meta) []suggest.Holding {
	out := make([]suggest.Holding, len(assets))
	for i, a := range assets {
		base, _ := marketdata.SplitSim(a.ID)
		m, ok := meta[marketdata.CanonicalID(base)]
		if !ok {
			m, ok = meta[base]
		}
		out[i] = suggest.Holding{ID: base, Weight: a.Weight, Meta: m, HasMeta: ok}
	}
	return out
}

// composition is the look-through of a portfolio's weights.
func composition(assets []portfolio.Asset) (Composition, error) {
	meta, err := catalog()
	if err != nil {
		return Composition{}, err
	}
	h := holdings(assets, meta)
	c := Composition{
		AssetClass: suggest.AssetClassSplit(h),
		Geography:  suggest.GeographySplit(h),
		Currency:   suggest.CurrencySplit(h),
		Duration:   suggest.DurationSplit(h),
	}
	c.Sectors, c.Equity = suggest.EquitySectorSplit(h)
	c.Coverage, c.Unclassified = suggest.Coverage(h, suggest.RegimeFramework())
	return c, nil
}
