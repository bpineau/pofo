package compare

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// fetchIn fetches one asset in an explicit target currency, honoring the
// window and SIM/simdata toggles carried by the options. It is the library
// counterpart of the CLI's fetchAssetIn.
func (opt Options) fetchIn(ctx context.Context, client *marketdata.Client, id, currency string) (*marketdata.Series, error) {
	return client.FetchExtended(ctx, id, marketdata.FetchOptions{
		From:     opt.Start,
		To:       opt.End,
		NoSim:    opt.NoSim,
		Simdata:  opt.Simdata,
		Currency: currency,
	})
}

// Compute runs the whole comparison pipeline for already-parsed specs: quotes
// and benchmark fetches, portfolio builds, simulations, the common window, and
// nominal/real statistics.
func Compute(ctx context.Context, client *marketdata.Client, specs []*portfolio.Spec, opt Options) (*Comparison, error) {
	// Download every distinct (currency, asset) once. A "#meta currencies"
	// directive evaluates the same portfolio in several currencies.
	seriesByCur := map[string]map[string]*marketdata.Series{}
	resolved := map[string]bool{} // report each id's resolved instrument once
	for _, spec := range specs {
		for _, cur := range effectiveCurrencies(spec, opt.Currency) {
			m := seriesByCur[cur]
			if m == nil {
				m = map[string]*marketdata.Series{}
				seriesByCur[cur] = m
			}
			for _, h := range spec.Holdings {
				// "#meta sim:on" backcasts every holding: fetch (and cache)
				// its SIM variant, keyed by the same id Build will request.
				// -no-simulate is honored downstream in FetchExtended (NoSim),
				// which fetches real quotes for a SIM id, so the flag still
				// wins over the meta with no extra handling here.
				fetchID := portfolio.SimFetchID(h.ID, spec.Sim)
				if _, ok := m[fetchID]; ok {
					continue
				}
				s, err := opt.fetchIn(ctx, client, fetchID, cur)
				if err != nil {
					return nil, fmt.Errorf("portfolio %s, asset %q (%s): %w", spec.Name, h.ID, cur, err)
				}
				m[fetchID] = s
				// Surface what each identifier resolved to: a fuzzy source match
				// can return a wrong instrument (e.g. "SP500" -> an S&P sector
				// sub-index), and a silent mismatch is how delirious numbers slip
				// through. Show it once so the user can catch it.
				if !resolved[h.ID] {
					resolved[h.ID] = true
					log.Printf("resolved %s -> %q [%s, %s]", h.ID, s.Name, s.Source, s.Currency)
				}
			}
		}
	}

	// Benchmark for Beta/CWARP, best effort, memoized per currency. The chart's
	// reference curve uses the default currency (benchIn(opt.Currency)).
	benchCache := map[string]*marketdata.Series{}
	benchIn := func(cur string) *marketdata.Series {
		if opt.Benchmark == "" {
			return nil
		}
		if b, ok := benchCache[cur]; ok {
			return b
		}
		b, err := client.FetchExtended(ctx, opt.Benchmark, marketdata.FetchOptions{
			From: opt.Start, NoSim: true, Currency: cur,
		})
		if err != nil {
			log.Printf("warning: benchmark %s unavailable in %s (no Beta): %v", opt.Benchmark, cur, err)
			b = nil
		}
		benchCache[cur] = b
		return b
	}
	bench := benchIn(opt.Currency)

	// Simulate each portfolio; a "#meta rebalance:N" directive overrides
	// the CLI default for that portfolio only.
	var feesFor func(string) (float64, bool)
	if !opt.NoFees {
		feesFor = func(id string) (float64, bool) {
			base, _ := marketdata.SplitSim(id)
			return client.Fees(ctx, base)
		}
	}
	// The financing rate (leverage) is only fetched when needed.
	var cashRate *marketdata.Series
	for _, spec := range specs {
		if spec.Leverage {
			cr, err := client.Fetch(ctx, "^IRX", opt.Start)
			if err != nil {
				log.Printf("warning: financing rate ^IRX unavailable (%v), leverage financed at 0 %%", err)
			} else {
				cashRate = cr
			}
			break
		}
	}

	results := make([]*column, 0, len(specs))
	simulateInto := func(p *portfolio.Portfolio, spec *portfolio.Spec, currency string) error {
		days := opt.Rebalance
		if spec.RebalanceDays >= 0 {
			days = spec.RebalanceDays
		}
		sim, err := portfolio.Simulate(p, days)
		if err != nil {
			return fmt.Errorf("portfolio %s: %w", p.Name, err)
		}
		if sim.Ruined {
			cause := "the leveraged exposure exhausted the net value"
			if p.Withdraw.Active() && !p.Leverage {
				cause = "withdrawals exhausted the capital"
			}
			when := sim.Dates[len(sim.Dates)-1].Format("2006-01-02")
			log.Printf("warning: portfolio %s wiped out on %s, series truncated", p.Name, when)
			p.Warnings = append(p.Warnings, fmt.Sprintf(
				"capital wiped out on %s: %s; the series stops there", when, cause))
		}
		results = append(results, &column{p: p, sim: sim, color: chart.PaletteColor(len(results)), rebalanceDays: days, currency: currency, specName: spec.Name})
		return nil
	}
	for _, spec := range specs {
		for _, cur := range effectiveCurrencies(spec, opt.Currency) {
			p, err := portfolio.Build(spec, portfolio.BuildOptions{
				Fetch:        func(id string) (*marketdata.Series, error) { return seriesByCur[cur][id], nil },
				Fees:         feesFor,
				Cash:         cashRate,
				BorrowSpread: 1.0, // default: cash + 1 %/yr
				BaseCurrency: cur,
			})
			if err != nil {
				return nil, err
			}
			// Multi-currency: tag each column with its currency.
			if len(spec.Currencies) > 0 {
				p.Name = fmt.Sprintf("%s (%s)", p.Name, cur)
			}
			// An optimized portfolio is shown next to its written weights, so
			// the optimizer's choice can be compared with the baseline.
			// (Optimize and currencies cannot be combined, so cur is unique here.)
			if spec.Optimize != nil {
				pOpt, note, err := optimizedPortfolio(p, spec, benchIn(cur))
				if err != nil {
					return nil, fmt.Errorf("portfolio %s: %w", spec.Name, err)
				}
				p.Name = spec.Name + " (as written)"
				if err := simulateInto(p, spec, cur); err != nil {
					return nil, err
				}
				if err := simulateInto(pOpt, spec, cur); err != nil {
					return nil, err
				}
				results[len(results)-1].note = note
				continue
			}
			if err := simulateInto(p, spec, cur); err != nil {
				return nil, err
			}
		}
	}

	// Common window across portfolios: statistics and the comparison chart
	// must cover the same period to be meaningful.
	commonStart := results[0].sim.Dates[0]
	commonEnd := results[0].sim.Dates[len(results[0].sim.Dates)-1]
	for _, r := range results[1:] {
		if f := r.sim.Dates[0]; f.After(commonStart) {
			commonStart = f
		}
		if l := r.sim.Dates[len(r.sim.Dates)-1]; l.Before(commonEnd) {
			commonEnd = l
		}
	}
	if !commonStart.Before(commonEnd) {
		return nil, errors.New("no common period across the portfolios")
	}
	// Consumer-price index per currency, memoized, to report drawdowns/TTR and
	// real stats in purchasing-power terms alongside the nominal ones.
	// Best-effort: a currency without a wired CPI simply has no real columns.
	deflatorCache := map[string]*marketdata.Series{}
	deflatorIn := func(cur string) (*marketdata.Series, bool) {
		if s, ok := deflatorCache[cur]; ok {
			return s, s != nil
		}
		s, ok := inflationSeries(ctx, client, cur, commonStart)
		if !ok {
			s = nil
		}
		deflatorCache[cur] = s
		return s, s != nil
	}
	for _, r := range results {
		i, j := window(r.sim.Dates, commonStart, commonEnd)
		if j-i < 2 {
			return nil, fmt.Errorf("portfolio %s: too few points in the common window", r.p.Name)
		}
		r.winDates = r.sim.Dates[i:j]
		r.winValues = rebase(r.sim.Index[i:j])
		st, err := metrics.Compute(r.winDates, r.winValues)
		if err != nil {
			return nil, fmt.Errorf("portfolio %s: %w", r.p.Name, err)
		}
		if d, ok := deflatorIn(r.currency); ok {
			if rs, err := metrics.Compute(r.winDates, deflate(r.winDates, r.winValues, d)); err == nil {
				r.realStats, r.hasReal = rs, true
			}
		}
		if b := benchIn(r.currency); b != nil {
			bd, bv := seriesSlices(b)
			if rel, ok := metrics.VsBenchmark(r.winDates, r.winValues, bd, bv); ok {
				st.Beta, st.HasBeta = rel.Beta, true
				r.rel, r.hasRel = rel, true
			}
			if c, ok := metrics.CWARPvs(r.winDates, r.winValues, bd, bv, metrics.CWARPParams{}); ok {
				st.CWARP, st.HasCWARP = c, true
			}
		}
		r.vts, r.hasVTS = metrics.VarianceRatio(r.winDates, r.winValues)
		r.stats = st
	}

	assetMeta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		log.Printf("warning: asset metadata unavailable (%v), regime coverage omitted", err)
	}

	return &Comparison{columns: results, bench: bench, commonStart: commonStart, commonEnd: commonEnd, meta: assetMeta, opt: opt}, nil
}
