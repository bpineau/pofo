package remote

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func analysisOptions(in AnalysisInput, defaults, allowed []string, r *Result) (analyze.Options, error) {
	currency := def(in.Currency, "EUR", "currency", r)
	history := def(in.History, "observed", "history", r)
	policy := def(in.DataPolicy, "cached_or_fetch", "data_policy", r)
	if !slices.Contains(evaluationCurrencies, currency) {
		return analyze.Options{}, invalid("currency", "Unsupported evaluation currency.")
	}
	if history != "observed" && history != "extended" {
		return analyze.Options{}, invalid("history", "Expected observed or extended.")
	}
	if policy != "cache_only" && policy != "cached_or_fetch" {
		return analyze.Options{}, invalid("data_policy", "Expected cache_only or cached_or_fetch.")
	}
	sections := in.Sections
	if sections == nil {
		sections = defaults
		r.Meta.Defaulted = append(r.Meta.Defaulted, Default{"sections", sections})
	}
	seen := map[string]bool{}
	for _, v := range sections {
		if !slices.Contains(allowed, v) || seen[v] {
			return analyze.Options{}, invalid("sections", "Sections must be distinct supported names.")
		}
		seen[v] = true
	}
	r.Meta.Sections = sections
	r.Meta.Currency = currency
	opt := analyze.Options{Currency: currency, Sim: history == "extended", Benchmark: in.BenchmarkID, ExactOnly: true, CorrelationFrequency: marketdata.Monthly}
	for _, f := range []struct {
		s, path string
		dst     *time.Time
	}{{in.From, "from", &opt.From}, {in.To, "to", &opt.To}} {
		if f.s != "" {
			v, e := time.Parse(time.DateOnly, f.s)
			if e != nil || v.Year() < 1800 || v.After(time.Now().UTC()) {
				return opt, invalid(f.path, "Use an ISO date between 1800-01-01 and today.")
			}
			*f.dst = v
		}
	}
	if !opt.From.IsZero() && !opt.To.IsZero() && !opt.From.Before(opt.To) {
		return opt, invalid("from", "The requested interval must have positive duration.")
	}
	if in.BenchmarkID != "" {
		if err := checkIDs([]string{in.BenchmarkID}, in.History, "benchmark_id"); err != nil {
			return opt, err
		}
		if nonInvestment(in.BenchmarkID) {
			return opt, invalid("benchmark_id", "A rate or inflation level cannot be a total-return benchmark.")
		}
	}
	r.Meta.RequestedWindow = &Window{in.From, in.To}
	return opt, nil
}
func section(r *Result, s string) bool { return slices.Contains(r.Meta.Sections, s) }
func analysisContext(ctx context.Context, in AnalysisInput) context.Context {
	if in.DataPolicy != nil && *in.DataPolicy == "cache_only" {
		return marketdata.WithCacheOnly(context.WithValue(ctx, localOnlyKey{}, true))
	}
	return ctx
}
func nonInvestment(id string) bool {
	b := baseID(id)
	return marketdata.RateName(b) != "" || strings.HasPrefix(b, "^HICP") || strings.HasPrefix(b, "^CPI") || slices.Contains([]string{"^IRX", "^FVX", "^TNX", "^TYX", "TBILL-3M", "TREASURY-LONG-YIELD"}, b)
}
func statistics(s metrics.Stats, n int) *Statistics {
	return &Statistics{Number(100 * s.CAGR), Number(100 * s.Volatility), Number(100 * s.MaxDrawdown), Number(s.Sharpe), Number(s.Sortino), s.Years, n, s.PeriodsPerYear}
}
func provenance(s *marketdata.Series) Provenance {
	return Provenance{s.Symbol, s.Source, s.Currency, Window{date(s.First().Date), date(s.Last().Date)}, s.FullySimulated, date(s.SimulatedBefore), s.ProxySymbol}
}
func annual(xs []metrics.PeriodReturn) []AnnualReturn {
	out := make([]AnnualReturn, 0, len(xs))
	for _, x := range xs {
		out = append(out, AnnualReturn{Window{date(x.Start), date(x.End)}, Number(x.Return * 100), x.Partial})
	}
	return out
}
func drawdowns(xs []metrics.Episode) []Drawdown {
	xs = slices.Clone(xs)
	slices.SortStableFunc(xs, func(a, b metrics.Episode) int {
		if a.Depth < b.Depth {
			return -1
		}
		if a.Depth > b.Depth {
			return 1
		}
		return 0
	})
	out := make([]Drawdown, 0, min(10, len(xs)))
	for _, x := range xs[:min(10, len(xs))] {
		out = append(out, Drawdown{date(x.PeakDate), date(x.TroughDate), date(x.RecoverDate), Number(x.Depth * 100), x.Ongoing})
	}
	return out
}
func relative(x *metrics.Relative, id string, own Window, bench *marketdata.Series) *RelativeStatistics {
	if id == "" {
		return nil
	}
	n := Number(math.NaN())
	r := &RelativeStatistics{BenchmarkID: id, Beta: n, AlphaPct: n, InformationRatio: n, UpCapturePct: n, DownCapturePct: n}
	if bench != nil {
		r.Window = Window{max(own.From, date(bench.First().Date)), min(own.To, date(bench.Last().Date))}
	}
	if x != nil {
		r.Beta = Number(x.Beta)
		r.AlphaPct = Number(x.Alpha * 100)
		r.InformationRatio = Number(x.InfoRatio)
		r.UpCapturePct = Number(x.UpCapture * 100)
		r.DownCapturePct = Number(x.DownCapture * 100)
	}
	return r
}
func assetProjection(a *analyze.AssetStudy, r *Result, requested string, bench *marketdata.Series) AssetData {
	d := AssetData{RequestedID: requested, Provenance: ptr(provenance(a.Series))}
	if section(r, "identity") || section(r, "holdings") {
		i := Identity{ID: a.Series.Symbol, Name: a.Series.Name, QuoteCurrency: a.Series.Currency}
		if a.HasMeta {
			i = identity(a.Meta)
		}
		d.Identity = &i
	}
	if section(r, "statistics") || section(r, "holdings") {
		d.Statistics = statistics(a.Stats, len(a.Series.Points))
	}
	if section(r, "metadata") && a.HasMeta {
		m := a.Meta
		d.Metadata = &AssetMetadata{m.Geography, m.Sectors, m.CurrencyExposure, m.AsOf, m.Notes, m.Sources}
	}
	if section(r, "annual_returns") {
		d.AnnualReturns = annual(a.Years)
	}
	if section(r, "drawdowns") {
		d.Drawdowns = ptr(drawdowns(a.Drawdowns))
	}
	d.Relative = relative(a.Relative, rBenchmark(r), d.Provenance.Window, bench)
	return d
}

// benchmark is stored only in local execution arguments, never inferred from labels.
func rBenchmark(r *Result) string {
	a, ok := r.Meta.Assumptions.(AnalysisAssumptions)
	if ok {
		return a.BenchmarkID
	}
	return ""
}
func loaded(ctx context.Context, g *guardedSource, ids []string, opt analyze.Options, r *Result) (analyze.SeriesSource, error) {
	src := analyze.SeriesSource{}
	for _, id := range ids {
		fetchID := portfolio.SimFetchID(id, opt.Sim)
		if _, ok := src[id]; ok {
			continue
		}
		s, err := g.FetchExtended(ctx, fetchID, marketdata.FetchOptions{From: opt.From, To: opt.To, Currency: opt.Currency, ExactOnly: true})
		if err != nil {
			return nil, err
		}
		if len(s.Points) < 2 {
			return nil, &Error{Code: "no_common_period", Message: "At least two usable observations are required."}
		}
		if s.Currency == "" && opt.Currency != "" {
			return nil, &Error{Code: "unsupported_configuration", Message: "Evaluation currency cannot be established for this series."}
		}
		if (opt.Sim || fetchID != id) && !s.FullySimulated && s.SimulatedBefore.IsZero() {
			warn(r, "extended_history_unavailable", "No reconstructed extension was available; using observed history.", id)
		}
		src[id] = s
		src[fetchID] = s
	}
	return src, nil
}
func (s *Service) assets(ctx context.Context, in AssetsInput, r *Result) (AssetsData, error) {
	out := AssetsData{Assets: []AssetData{}}
	opt, err := analysisOptions(in.AnalysisInput, []string{"identity", "statistics"}, []string{"identity", "statistics", "metadata", "annual_returns", "drawdowns"}, r)
	if err != nil {
		return out, err
	}
	if len(in.IDs) < 1 || len(in.IDs) > 10 {
		return out, invalid("ids", "Supply one through ten identifiers.")
	}
	if err = checkIDs(in.IDs, in.History, "ids"); err != nil {
		return out, err
	}
	quotes := section(r, "statistics") || section(r, "annual_returns") || section(r, "drawdowns") || in.BenchmarkID != ""
	for _, id := range in.IDs {
		if quotes && nonInvestment(id) {
			return out, invalid("ids", "Rate and inflation levels support identity/metadata only.")
		}
	}
	ctx = analysisContext(ctx, in.AnalysisInput)
	g := s.source()
	r.Meta.Assumptions = AnalysisAssumptions{ReturnBasis: "nominal", BenchmarkID: in.BenchmarkID}
	var bench *marketdata.Series
	if in.BenchmarkID != "" {
		b, e := loaded(ctx, g, []string{in.BenchmarkID}, opt, r)
		if e != nil {
			return out, e
		}
		bench = b[in.BenchmarkID]
	}
	for idx, id := range in.IDs {
		if !quotes {
			if a, ok := marketdata.Lookup(id); ok {
				d := AssetData{RequestedID: id}
				if section(r, "identity") {
					d.Identity = ptr(identity(a))
				}
				if section(r, "metadata") {
					d.Metadata = &AssetMetadata{a.Geography, a.Sectors, a.CurrencyExposure, a.AsOf, a.Notes, a.Sources}
				}
				out.Assets = append(out.Assets, d)
				continue
			}
			if nonInvestment(id) {
				name := marketdata.RateName(baseID(id))
				if name == "" {
					name = baseID(id)
				}
				d := AssetData{RequestedID: id}
				if section(r, "identity") {
					d.Identity = &Identity{ID: baseID(id), Name: name}
				}
				warn(r, "reference_level", "This is a rate or inflation level, not an investable total-return series.", id)
				out.Assets = append(out.Assets, d)
				continue
			}
		}
		src, e := loaded(ctx, g, []string{id}, opt, r)
		if e != nil {
			pe := publicError(e)
			pe.Path = fmt.Sprintf("ids[%d]", idx)
			return out, pe
		}
		if bench != nil {
			src[in.BenchmarkID] = bench
		}
		a, e := analyze.Asset(ctx, src, id, opt)
		if e != nil {
			return out, e
		}
		if len(a.Years) > 150 {
			return out, invalid("from", "Annual history is limited to 150 years; request a shorter interval.")
		}
		warnings(r, a.Warnings, fmt.Sprintf("ids[%d]", idx))
		out.Assets = append(out.Assets, assetProjection(a, r, id, bench))
	}
	return out, nil
}
func validateHoldings(hs []Holding, path string, history *string) error {
	if len(hs) < 1 || len(hs) > 20 {
		return invalid(path, "Supply one through twenty holdings.")
	}
	sum := 0.
	seen := map[string]bool{}
	for i, h := range hs {
		p := fmt.Sprintf("%s[%d]", path, i)
		if err := checkIDs([]string{h.ID}, history, p+".id"); err != nil {
			return err
		}
		if nonInvestment(h.ID) {
			return invalid(p+".id", "A rate or inflation level cannot be a portfolio holding.")
		}
		key := baseID(h.ID)
		if seen[key] {
			return invalid(p+".id", "Duplicate canonical holding.")
		}
		seen[key] = true
		if h.WeightPct == nil || math.IsNaN(*h.WeightPct) || math.IsInf(*h.WeightPct, 0) || *h.WeightPct < 0 || *h.WeightPct > 100 {
			return invalid(p+".weight_pct", "Weight is required and must be between 0 and 100 percent.")
		}
		sum += *h.WeightPct
	}
	if math.Abs(sum-100) > 1e-6 {
		return invalid(path, "Portfolio weights must sum to 100 percent.")
	}
	return nil
}
func (s *Service) portfolios(ctx context.Context, in PortfoliosInput, r *Result) (PortfoliosData, error) {
	out := PortfoliosData{Portfolios: []PortfolioData{}}
	opt, err := analysisOptions(in.AnalysisInput, []string{"statistics"}, []string{"statistics", "composition", "holdings", "correlations", "attribution", "annual_returns", "drawdowns"}, r)
	if err != nil {
		return out, err
	}
	if len(in.Portfolios) < 1 || len(in.Portfolios) > 4 {
		return out, invalid("portfolios", "Supply one through four allocations.")
	}
	var ids []string
	for i := range in.Portfolios {
		a := &in.Portfolios[i]
		path := fmt.Sprintf("portfolios[%d]", i)
		if utf8.RuneCountInString(a.Label) > 80 {
			return out, invalid(path+".label", "Label exceeds 80 characters.")
		}
		if err = validateHoldings(a.Holdings, path+".holdings", in.History); err != nil {
			return out, err
		}
		reb := def(a.RebalanceDays, 90, path+".rebalance_days", r)
		fee := def(a.AnnualFeePct, 0., path+".annual_fee_pct", r)
		if reb < 0 || reb > 36500 || fee < 0 || fee > 20 {
			return out, invalid(path, "Rebalance days must be 0 through 36500; annual fee 0 through 20 percent.")
		}
		a.RebalanceDays = ptr(reb)
		a.AnnualFeePct = ptr(fee)
		for _, h := range a.Holdings {
			ids = append(ids, h.ID)
		}
	}
	unique := map[string]bool{}
	for _, id := range ids {
		unique[baseID(id)] = true
	}
	if in.BenchmarkID != "" {
		unique[baseID(in.BenchmarkID)] = true
	}
	if len(unique) > 40 {
		return out, invalid("portfolios", "At most 40 unique quote dependencies including the benchmark are allowed.")
	}
	ctx = analysisContext(ctx, in.AnalysisInput)
	g := s.source()
	src, err := loaded(ctx, g, ids, opt, r)
	if err != nil {
		return out, err
	}
	// Choose shared quoted endpoints through the library's exact-period panel.
	// Merely intersecting date ranges lets disjoint calendars produce different
	// actual start/end dates when each portfolio is subsequently trimmed.
	var commonSeries []*marketdata.Series
	seenSeries := map[string]bool{}
	for _, id := range ids {
		key := canonical(id)
		if seenSeries[key] {
			continue
		}
		seenSeries[key] = true
		v := *src[id]
		v.Symbol = key
		commonSeries = append(commonSeries, &v)
	}
	panel, err := marketdata.NewPanel(marketdata.Daily, commonSeries...)
	if err != nil || panel.Len() == 0 {
		return out, &Error{Code: "no_common_period", Message: "The holdings need at least two common quoted dates for a comparable observation window."}
	}
	from, to := panel.Starts[0], panel.Ends[panel.Len()-1]
	opt.From, opt.To = from, to
	r.Meta.EffectiveWindow = &Window{date(from), date(to)}
	r.Meta.Assumptions = AnalysisAssumptions{ReturnBasis: "nominal", CorrelationFrequency: "monthly", BenchmarkID: in.BenchmarkID}
	var bench *marketdata.Series
	if in.BenchmarkID != "" {
		b, e := loaded(ctx, g, []string{in.BenchmarkID}, analyze.Options{Currency: opt.Currency, Sim: opt.Sim, ExactOnly: true}, r)
		if e != nil {
			return out, e
		}
		bench = b[in.BenchmarkID]
		src[in.BenchmarkID] = bench
	}
	for _, a := range in.Portfolios {
		var lines []portfolio.Line
		for _, h := range a.Holdings {
			lines = append(lines, portfolio.Line{ID: h.ID, Weight: *h.WeightPct / 100})
		}
		spec, e := portfolio.NewSpecWith(a.Label, portfolio.SpecOptions{RebalanceDays: a.RebalanceDays, EnvelopeFees: a.AnnualFeePct}, lines...)
		if e != nil {
			return out, invalid("portfolios", e.Error())
		}
		study, e := analyze.Portfolio(ctx, src, spec, opt)
		if e != nil {
			return out, e
		}
		warnings(r, study.Warnings, a.Label)
		d := PortfolioData{Label: a.Label, RebalanceDays: *a.RebalanceDays, AnnualFeePct: *a.AnnualFeePct, Holdings: []Holding{}, Provenance: []Provenance{}}
		for _, h := range study.Spec.Holdings {
			d.Holdings = append(d.Holdings, Holding{canonical(h.ID), ptr(h.Weight * 100)})
		}
		for _, h := range study.Holdings {
			d.Provenance = append(d.Provenance, provenance(h.Series))
		}
		if section(r, "statistics") {
			d.Statistics = statistics(study.Stats, len(study.Sim.Index))
		}
		if section(r, "composition") {
			c := study.Composition
			coverage := map[string]float64{}
			for k, v := range c.Coverage {
				coverage[string(k)] = 100 * v
			}
			d.Composition = &Composition{pctMap(c.AssetClass), pctMap(c.Geography), pctMap(c.Currency), pctMap(c.Sectors), 100 * c.Equity, coverage, 100 * c.Unclassified}
		}
		if section(r, "holdings") {
			for _, h := range study.Holdings {
				d.Assets = append(d.Assets, assetProjection(&h, r, h.ID, bench))
			}
		}
		if section(r, "correlations") {
			for _, row := range study.Correlation {
				rr := []Number{}
				for _, v := range row {
					rr = append(rr, Number(v))
				}
				d.Correlations = append(d.Correlations, rr)
			}
		}
		if section(r, "attribution") {
			d.Attribution = &Attribution{"monthly", pctSlice(study.Attribution.Risk), pctSlice(study.Attribution.Return)}
		}
		if section(r, "annual_returns") {
			if len(study.Years) > 150 {
				return out, invalid("from", "Annual history exceeds 150 years.")
			}
			d.AnnualReturns = annual(study.Years)
		}
		if section(r, "drawdowns") {
			d.Drawdowns = ptr(drawdowns(study.Drawdowns))
		}
		d.Relative = relative(study.Relative, in.BenchmarkID, *r.Meta.EffectiveWindow, bench)
		out.Portfolios = append(out.Portfolios, d)
	}
	return out, nil
}
func pctMap(src map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range src {
		out[k] = 100 * v
	}
	return out
}
func pctSlice(xs []float64) []Number {
	out := make([]Number, len(xs))
	for i, v := range xs {
		out[i] = Number(100 * v)
	}
	return out
}

var evaluationCurrencies = []string{"EUR", "USD", "GBP", "CHF", "JPY", "CAD", "AUD", "SEK", "NOK", "DKK", "CNY", "HKD", "NZD"}
