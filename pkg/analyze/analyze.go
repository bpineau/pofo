package analyze

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// DefaultRebalance is the rebalancing period, in calendar days, a portfolio
// is simulated with when neither Options.Rebalance nor the spec's own
// RebalanceDays sets one: the pofo CLI's default, stated once.
const DefaultRebalance = 90

// Source is what analyze needs of a data client. *marketdata.Client
// satisfies it; tests supply a fake, and so can a consumer with its own
// store.
// If a Source also implements Client.Load's signature, analysis uses Load:
// a marketdata.Client therefore accepts bundled reference IDs and CSV paths
// by the same contract as an exploration script. SeriesSource needs neither.
//
// FetchExtended follows marketdata.Client.FetchExtended: an identifier with
// the "SIM" suffix asks for the backcast-extended history, FetchOptions.From
// and To bound the window, Currency converts. Fees returns a TER in PERCENT
// per year for a bare identifier (no SIM suffix), ok false when unknown.
type Source interface {
	// FetchExtended returns the series of id, backcast-extended under the
	// SIM suffix, windowed and converted per opt.
	FetchExtended(ctx context.Context, id string, opt marketdata.FetchOptions) (*marketdata.Series, error)
	// Fees returns the TER of a bare id in PERCENT per year, ok false when
	// unknown.
	Fees(ctx context.Context, id string) (float64, bool)
}

var _ Source = (*marketdata.Client)(nil)

// Options shapes a study. The zero value studies everything available, in
// each series' native currency, real quotes only, with no benchmark.
type Options struct {
	// Simulation supplies a complete simulation configuration, including
	// calendar rebalancing and transaction costs. Nil uses the legacy
	// Rebalance/spec/default precedence. With a nonnil value, Rebalance must
	// be zero and the spec must not declare RebalanceDays: conflicting
	// instructions are errors. Zero SimOptions means no scheduled rebalance.
	Simulation *portfolio.SimOptions
	// Inflation is a price-level index, chosen independently of Currency.
	// When set, Real contains purchasing-power statistics on the same window.
	Inflation *marketdata.Series
	// RiskFree is an annual arithmetic return as a fraction, used by Sharpe
	// and Sortino. It cannot be combined with RiskFreeIndex.
	RiskFree float64
	// RiskFreeIndex is a total-return cash index in the evaluation currency.
	// Excess contains arithmetic excess returns; nominal Sharpe and Sortino
	// use those returns. Use marketdata.CashIndex for annual quoted rates.
	RiskFreeIndex *marketdata.Series
	// Adjustment controls coverage and staleness for Inflation/RiskFreeIndex.
	Adjustment marketdata.AdjustmentOptions
	// Financing is an annual rate in PERCENT used by leveraged portfolios.
	// FinancingID loads that rate from Source instead. They are exclusive;
	// both unset retain the legacy ^IRX default. Neither is a cash index.
	Financing   *marketdata.Series
	FinancingID string
	// CorrelationFrequency selects shared return periods for the correlation
	// matrix. Zero uses common quote dates; Monthly suits different closing
	// times. No forward-filled returns enter the matrix.
	CorrelationFrequency marketdata.Frequency
	Currency             string    // evaluation currency (ISO code); "" = native
	Sim                  bool      // splice bundled backcasts (the SIM convention, see Asset and Portfolio)
	Benchmark            string    // identifier for the relative statistics; "" = none
	From, To             time.Time // study window; zero = open on that side
	Rebalance            int       // rebalancing period in days for a portfolio; 0 = DefaultRebalance, negative = never
	Simdata              fs.FS     // simulated histories for the SIM convention; nil = the embedded bundle
	ExactOnly            bool      // marketdata.FetchOptions.ExactOnly: no name-matched instrument
}

// fetchOptions is the marketdata request every holding and asset of a study
// is fetched with.
func (o Options) fetchOptions() marketdata.FetchOptions {
	return marketdata.FetchOptions{
		From:      o.From,
		To:        o.To,
		Simdata:   o.Simdata,
		Currency:  o.Currency,
		ExactOnly: o.ExactOnly,
	}
}

// check refuses an options value no study can honor.
func (o Options) check() error {
	if math.IsNaN(o.RiskFree) || math.IsInf(o.RiskFree, 0) || o.RiskFree != 0 && o.RiskFreeIndex != nil {
		return errors.New("analyze: choose a finite RiskFree or a RiskFreeIndex, not both")
	}
	if o.Financing != nil && o.FinancingID != "" {
		return errors.New("analyze: choose Financing or FinancingID, not both")
	}
	if o.CorrelationFrequency < 0 || o.Adjustment.MaxAge < 0 {
		return errors.New("analyze: negative correlation frequency or maximum observation age")
	}
	if !o.From.IsZero() && !o.To.IsZero() && o.To.Before(o.From) {
		return fmt.Errorf("analyze: window ends %s, before it starts %s",
			o.To.Format(time.DateOnly), o.From.Format(time.DateOnly))
	}
	return nil
}

// AssetStudy is one asset dissected on the study window. Every return, rate
// and drawdown in it is a FRACTION (0.07 = +7 %), except metrics.Stats.Ulcer
// (percent points) and Stats.CWARP (percent), as in pkg/metrics.
type AssetStudy struct {
	Real      *AdjustedStudy         // purchasing power, when Inflation is set
	Excess    *AdjustedStudy         // arithmetic excess over RiskFreeIndex, when set
	ID        string                 // identifier as asked for (or as written in the spec)
	Meta      datasets.Asset         // catalog record, zero when HasMeta is false
	HasMeta   bool                   // the identifier is in the bundled catalog (marketdata.Lookup)
	Series    *marketdata.Series     // as evaluated: converted, spliced when asked, trimmed to the window
	Stats     metrics.Stats          // on Series; Beta and CWARP set when a benchmark overlaps it
	Years     []metrics.PeriodReturn // calendar years, the first one Partial
	Months    []metrics.PeriodReturn // calendar months, the first one Partial
	Drawdowns []metrics.Episode      // every drawdown episode, chronological
	Relative  *metrics.Relative      // vs Options.Benchmark; nil without one or without overlap
	Warnings  []string               // data problems behind the numbers (see the package doc)
}

// Asset studies one asset over the longest window its data covers, clipped
// to [opt.From, opt.To].
//
// With opt.Sim the identifier is fetched through portfolio.SimFetchID, as a
// portfolio holding would be: its backcast-extended history when the bundle
// has one, its real quotes otherwise (a failing SIM fetch falls back to the
// real one, with a warning). An identifier already written with the SIM
// suffix asks for the backcast whatever opt.Sim says.
//
// An identifier no source quotes is an error that satisfies
// errors.Is(err, marketdata.ErrUnknownIdentifier), and so is an unknown
// opt.Benchmark; a benchmark that merely failed to load is a warning, and
// leaves Relative nil. So is every other data problem (see the package doc).
func Asset(ctx context.Context, src Source, id string, opt Options) (*AssetStudy, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("analyze: empty identifier")
	}
	if err := opt.check(); err != nil {
		return nil, err
	}
	bench, warnings, err := benchmark(ctx, src, opt)
	if err != nil {
		return nil, err
	}
	fo := opt.fetchOptions()
	fetchID := portfolio.SimFetchID(id, opt.Sim)
	s, err := fetch(ctx, src, fetchID, fo)
	if err != nil && fetchID != id {
		// Same rule as portfolio.Build: Sim means "the backcast IF one
		// exists", so an asset without one is studied on its real quotes.
		if real, rerr := fetch(ctx, src, id, fo); rerr == nil {
			s, err = real, nil
			warnings = append(warnings, "no simulated history, using real quotes only")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("analyze: %s: %w", id, err)
	}
	s = marketdata.Trim(s, opt.From, opt.To)
	if s.Currency == "" && opt.Currency != "" {
		warnings = append(warnings, "unknown currency, left unconverted")
	}
	a, err := study(id, s, bench)
	if err != nil {
		return nil, fmt.Errorf("analyze: %s: %w", id, err)
	}
	if err := adjustStudy(a, opt); err != nil {
		return nil, fmt.Errorf("analyze: %s: %w", id, err)
	}
	a.Warnings = append(warnings, a.Warnings...)
	if opt.Benchmark != "" && bench != nil && a.Relative == nil {
		a.Warnings = append(a.Warnings, noOverlap(opt.Benchmark))
	}
	return a, nil
}

// study dissects one series as it is. The returned study is filled as far as
// the series allows even when err is not nil (identifier, catalog record,
// series and warnings), so a portfolio can keep a holding it has no
// statistics for.
func study(id string, s *marketdata.Series, bench *marketdata.Series) (*AssetStudy, error) {
	a := &AssetStudy{ID: id, Series: s, Warnings: seriesWarnings(s)}
	base, _ := marketdata.SplitSim(id)
	a.Meta, a.HasMeta = marketdata.Lookup(base)
	if s.Len() < 2 {
		return a, fmt.Errorf("%d quote(s) in the study window, at least 2 needed", s.Len())
	}
	dates, values := s.Dates(), s.Values()
	stats, err := s.Stats()
	if err != nil {
		return a, err
	}
	a.Stats = stats
	a.Years, a.Months, a.Drawdowns = calendar(dates, values)
	a.Relative = relative(&a.Stats, dates, values, bench)
	return a, nil
}

// calendar is the yearly and monthly return tables and the drawdown episodes
// of a value series. Its two period lengths are positive and dates and
// values come from one series, so CalendarReturns cannot fail here.
func calendar(dates []time.Time, values []float64) (years, months []metrics.PeriodReturn, drawdowns []metrics.Episode) {
	years, _ = metrics.CalendarReturns(dates, values, int(marketdata.Yearly))
	months, _ = metrics.CalendarReturns(dates, values, int(marketdata.Monthly))
	return years, months, metrics.DrawdownEpisodes(dates, values)
}

// relative measures a value series against the benchmark, the way the
// comparison report does: Relative from metrics.VsBenchmark, whose Beta also
// fills stats, and CWARP with the default parameters. It returns nil without
// a benchmark or without enough common dates.
func relative(stats *metrics.Stats, dates []time.Time, values []float64, bench *marketdata.Series) *metrics.Relative {
	if bench == nil {
		return nil
	}
	bd, bv := bench.Dates(), bench.Values()
	var out *metrics.Relative
	if rel, ok := metrics.VsBenchmark(dates, values, bd, bv); ok {
		stats.Beta, stats.HasBeta = rel.Beta, true
		out = &rel
	}
	if c, ok := metrics.CWARPvs(dates, values, bd, bv, metrics.CWARPParams{}); ok {
		stats.CWARP, stats.HasCWARP = c, true
	}
	return out
}

// benchmark fetches opt.Benchmark on the study's window and currency, real
// quotes only: a benchmark is a reference, not a reconstruction. An unknown
// identifier is the caller's error; any other failure only costs the relative
// statistics, and says so.
func benchmark(ctx context.Context, src Source, opt Options) (*marketdata.Series, []string, error) {
	if opt.Benchmark == "" {
		return nil, nil, nil
	}
	fo := opt.fetchOptions()
	fo.NoSim = true
	b, err := fetch(ctx, src, opt.Benchmark, fo)
	switch {
	case err == nil && b.Len() >= 2:
		return b, nil, nil
	case err == nil:
		return nil, []string{fmt.Sprintf("benchmark %s: fewer than two quotes, no relative statistics", opt.Benchmark)}, nil
	case errors.Is(err, marketdata.ErrUnknownIdentifier):
		return nil, nil, fmt.Errorf("analyze: benchmark %s: %w", opt.Benchmark, err)
	case ctx.Err() != nil:
		return nil, nil, ctx.Err()
	}
	return nil, []string{fmt.Sprintf("benchmark %s unavailable, no relative statistics: %v", opt.Benchmark, err)}, nil
}

// noOverlap is the warning of a study the benchmark shares too few dates with.
func noOverlap(bench string) string {
	return fmt.Sprintf("benchmark %s: too few common dates for relative statistics", bench)
}

// seriesWarnings lists what a series, as studied, carries that its
// statistics cannot show: reconstructed or estimated points inside it, a
// price-return NAV, a definition change read as a move.
func seriesWarnings(s *marketdata.Series) []string {
	if s.Len() == 0 {
		return nil
	}
	out := s.DataWarnings()
	first, last := s.First().Date, s.Last().Date
	if sb := s.SimulatedBefore; !sb.IsZero() && first.Before(sb) {
		out = append(out, fmt.Sprintf("simulated before %s%s: the history before that date is a reconstruction, not quotes",
			sb.Format(time.DateOnly), via(s.ProxySymbol)))
	}
	if ef := s.EstimatedFrom; !ef.IsZero() && !last.Before(ef) {
		out = append(out, fmt.Sprintf("estimated from %s%s: the points from that date on are a nowcast, not published values",
			ef.Format(time.DateOnly), via(s.EstimateProxy)))
	}
	for _, j := range s.Junctions {
		if j.After(first) && !j.After(last) {
			out = append(out, fmt.Sprintf("definition change on %s: the step into that date is not a market move, yet the statistics read it as one",
				j.Format(time.DateOnly)))
		}
	}
	return out
}

// via names the proxy of a reconstructed or estimated stretch, when known.
func via(proxy string) string {
	if proxy == "" {
		return ""
	}
	return " via " + proxy
}
