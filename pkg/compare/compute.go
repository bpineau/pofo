package compare

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"slices"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

// Compute runs the whole comparison pipeline for already-parsed specs. Every
// column is an analyze.Portfolio study (fetch, build, simulation, per-holding
// studies, composition, attribution: see Comparison.Studies), made through
// one memoizing source so a series shared by several columns is fetched once.
// Compute adds what only a comparison of several columns owns: the "#meta
// currencies" expansion (one study per currency), the "#meta optimize" column
// (the optimizer runs here, on the written column's built portfolio, and its
// weights are then studied like any others), the benchmark, the common window,
// and the nominal and real (CPI-deflated) statistics on that window.
func Compute(ctx context.Context, client *marketdata.Client, specs []*portfolio.Spec, opt Options) (*Comparison, error) {
	src := newSource(client, opt)

	// Fetch every holding up front, in every currency its spec is evaluated
	// in, so a failure names the portfolio, the asset and the currency before
	// any study starts; the studies then find each series in the memo.
	resolved := map[string]bool{} // report each id's resolved instrument once
	for _, spec := range specs {
		for _, cur := range effectiveCurrencies(spec, opt.Currency) {
			for _, h := range spec.Holdings {
				// "#meta sim:on" backcasts every holding: fetch its SIM
				// variant, the id portfolio.Build will request. -no-simulate
				// is honored by the source (NoSim), which fetches real quotes
				// for a SIM id, so the flag still wins over the meta.
				s, err := src.FetchExtended(ctx, portfolio.SimFetchID(h.ID, spec.Sim), src.holding(cur))
				if err != nil {
					return nil, fmt.Errorf("portfolio %s, asset %q (%s): %w", spec.Name, h.ID, cur, err)
				}
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

	// Benchmark for Beta/CWARP, best effort, per currency (the source
	// memoizes it). The chart's reference curve uses the default currency.
	warned := map[string]bool{}
	benchIn := func(cur string) *marketdata.Series {
		if opt.Benchmark == "" {
			return nil
		}
		b, err := src.FetchExtended(ctx, opt.Benchmark, src.benchmark(cur))
		if err != nil {
			if !warned[cur] {
				warned[cur] = true
				log.Printf("warning: benchmark %s unavailable in %s (no Beta): %v", opt.Benchmark, cur, err)
			}
			return nil
		}
		return b
	}
	bench := benchIn(opt.Currency)

	// The comparison's rebalancing period 0 means never; analyze reads 0 as
	// its default and a negative period as never.
	rebalance := opt.Rebalance
	if rebalance <= 0 {
		rebalance = -1
	}
	var results []*column
	for _, spec := range specs {
		for _, cur := range effectiveCurrencies(spec, opt.Currency) {
			aopt := analyze.Options{Currency: cur, From: opt.Start, To: opt.End, Rebalance: rebalance, Simdata: opt.Simdata}
			b := benchIn(cur)
			if b != nil {
				aopt.Benchmark = opt.Benchmark
			}
			name := spec.Name
			if len(spec.Currencies) > 0 {
				// Multi-currency: tag each column with its currency.
				name = fmt.Sprintf("%s (%s)", name, cur)
			}
			if spec.Optimize != nil {
				// An optimized portfolio is shown next to its written weights,
				// so the optimizer's choice can be compared with the baseline.
				// (Optimize and currencies cannot be combined, so cur is unique.)
				name = spec.Name + " (as written)"
			}
			written, err := studyColumn(ctx, src, columnSpec(spec, name), aopt, opt.Rebalance, cur, spec.Name)
			if err != nil {
				return nil, err
			}
			if spec.Optimize == nil {
				results = append(results, written)
				continue
			}
			pOpt, note, err := optimizedPortfolio(written.p, spec, b)
			if err != nil {
				return nil, fmt.Errorf("portfolio %s: %w", spec.Name, err)
			}
			optimized, err := studyColumn(ctx, src, weightedSpec(spec, pOpt), aopt, opt.Rebalance, cur, spec.Name)
			if err != nil {
				return nil, err
			}
			optimized.note = note
			results = append(results, written, optimized)
		}
	}

	// A public entry point must not panic on empty input: callers reach this
	// only with at least one spec today, but the guard keeps the API safe.
	if len(results) == 0 {
		return nil, errors.New("no portfolios to compare")
	}
	// Identity colors are assigned once the count is known: the palette picks
	// the n hues that stay apart from each other (chart.PaletteFor), which a
	// running index cannot do since column four decides what column one needs.
	pal := chart.PaletteFor(len(results))
	for i, r := range results {
		r.color = pal[i]
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
		if err := r.measure(i, j, benchIn(r.currency)); err != nil {
			return nil, fmt.Errorf("portfolio %s: %w", r.p.Name, err)
		}
		if d, ok := deflatorIn(r.currency); ok {
			if rs, err := metrics.Compute(r.winDates, deflate(r.winDates, r.winValues, d)); err == nil {
				r.realStats, r.hasReal = rs, true
			}
		}
	}

	assetMeta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		log.Printf("warning: asset metadata unavailable (%v), regime coverage omitted", err)
	}

	return &Comparison{columns: results, bench: bench, commonStart: commonStart, commonEnd: commonEnd, meta: assetMeta, opt: opt}, nil
}

// studyColumn studies one column's spec through analyze and wraps the study
// in the column record, with the report's warnings: the build's, the
// simulation's and a ruin, each also logged. rebalance is the comparison's
// default period, which the spec's "#meta rebalance:N" overrides; specName is
// the spec the column came from (the column's own name may be decorated).
func studyColumn(ctx context.Context, src *source, spec *portfolio.Spec, aopt analyze.Options, rebalance int, cur, specName string) (*column, error) {
	ps, err := analyze.Portfolio(ctx, src, spec, aopt)
	if err != nil {
		return nil, err
	}
	p, sim := ps.Portfolio, ps.Sim
	if p.Leverage && p.Cash == nil {
		log.Printf("warning: portfolio %s: financing rate unavailable, leverage financed at 0 %%", p.Name)
	}
	for _, w := range sim.Warnings {
		log.Printf("warning: portfolio %s: %s", p.Name, w)
	}
	warnings := slices.Concat(p.Warnings, sim.Warnings)
	if sim.Ruined {
		cause := "the leveraged exposure exhausted the net value"
		if p.Withdraw.Active() && !p.Leverage {
			cause = "withdrawals exhausted the capital"
		}
		when := sim.Dates[len(sim.Dates)-1].Format("2006-01-02")
		log.Printf("warning: portfolio %s wiped out on %s, series truncated", p.Name, when)
		warnings = append(warnings, fmt.Sprintf(
			"capital wiped out on %s: %s; the series stops there", when, cause))
	}
	days := rebalance
	if spec.RebalanceDays >= 0 {
		days = spec.RebalanceDays
	}
	return &column{
		study: ps, p: p, sim: sim, warnings: warnings,
		rebalanceDays: days, currency: cur, specName: specName,
	}, nil
}

// columnSpec is the spec one column studies: spec under the column's name,
// with the directives Compute acts on itself cleared (the currency list is
// expanded into columns, the optimizer runs here), so the study describes
// exactly what its column shows.
func columnSpec(spec *portfolio.Spec, name string) *portfolio.Spec {
	eff := *spec
	eff.Name = name
	eff.Currencies = nil
	eff.Optimize = nil
	return &eff
}

// weightedSpec is the spec of an optimizer's column: spec's holdings at the
// weights the optimizer gave p, under p's name. The weights are copied as
// they are, never renormalized, so the study simulates exactly them.
func weightedSpec(spec *portfolio.Spec, p *portfolio.Portfolio) *portfolio.Spec {
	eff := columnSpec(spec, p.Name)
	eff.Holdings = slices.Clone(spec.Holdings)
	for i := range eff.Holdings {
		eff.Holdings[i].Weight = p.Assets[i].Weight
		eff.Holdings[i].RawWeight = 100 * p.Assets[i].Weight
	}
	return eff
}

// measure sets the column's nominal statistics on the common window [i, j)
// of its simulation. When that window is the whole simulation, as it is for a
// lone portfolio, they are the study's own numbers; otherwise they are
// measured again on that slice of the same Sim.Index, with the calls analyze
// makes, so a shared window reads the same numbers either way.
func (r *column) measure(i, j int, bench *marketdata.Series) error {
	dates, index := r.sim.Dates[i:j], r.sim.Index[i:j]
	r.vts, r.hasVTS = metrics.VarianceRatio(dates, index)
	if r.study != nil && i == 0 && j == len(r.sim.Dates) {
		r.stats = r.study.Stats
		if rel := r.study.Relative; rel != nil {
			r.rel, r.hasRel = *rel, true
		}
		return nil
	}
	st, err := metrics.Compute(dates, index)
	if err != nil {
		return err
	}
	if bench != nil {
		bd, bv := bench.Dates(), bench.Values()
		if rel, ok := metrics.VsBenchmark(dates, index, bd, bv); ok {
			st.Beta, st.HasBeta = rel.Beta, true
			r.rel, r.hasRel = rel, true
		}
		if c, ok := metrics.CWARPvs(dates, index, bd, bv, metrics.CWARPParams{}); ok {
			st.CWARP, st.HasCWARP = c, true
		}
	}
	r.stats = st
	return nil
}
