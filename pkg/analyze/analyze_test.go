package analyze_test

import (
	"bytes"
	"context"
	"errors"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
	"github.com/bpineau/pofo/pkg/suggest"
)

var ctx = context.Background()

func mustSpec(t *testing.T, lines ...portfolio.Line) *portfolio.Spec {
	t.Helper()
	spec, err := portfolio.NewSpec("test", lines...)
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

func parseSpec(t *testing.T, text string) *portfolio.Spec {
	t.Helper()
	spec, err := portfolio.Parse("test", strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	return spec
}

// hasWarning reports whether a warning line contains every fragment.
func hasWarning(warnings []string, fragments ...string) bool {
	return slices.ContainsFunc(warnings, func(w string) bool {
		for _, f := range fragments {
			if !strings.Contains(w, f) {
				return false
			}
		}
		return true
	})
}

func threeFunds(t *testing.T) *portfolio.Spec {
	return mustSpec(t,
		portfolio.Line{ID: "IWDA", Weight: 0.4, Fees: -1},
		portfolio.Line{ID: "VWRL", Weight: 0.3, Fees: -1},
		portfolio.Line{ID: "IGLN", Weight: 0.3, Fees: -1})
}

// The portfolio's window is the latest start, and every holding is studied
// on it.
func TestPortfolioWindow(t *testing.T) {
	study, err := analyze.Portfolio(ctx, newFake(), threeFunds(t), analyze.Options{Benchmark: "MSCIWORLD"})
	if err != nil {
		t.Fatal(err)
	}
	start, end := study.Sim.Dates[0], study.Sim.Dates[len(study.Sim.Dates)-1]
	if !start.Equal(igRealStart) {
		t.Errorf("window starts %s, want the late holding's first quote %s", start, igRealStart)
	}
	if !study.Stats.Start.Equal(start) || !study.Stats.End.Equal(end) {
		t.Errorf("portfolio stats on %s..%s, want the simulation's %s..%s", study.Stats.Start, study.Stats.End, start, end)
	}
	if len(study.Holdings) != 3 {
		t.Fatalf("%d holdings, want 3", len(study.Holdings))
	}
	for i, h := range study.Holdings {
		if h.ID != study.Spec.Holdings[i].ID {
			t.Errorf("holding %d is %s, want spec order (%s)", i, h.ID, study.Spec.Holdings[i].ID)
		}
		if !h.Stats.Start.Equal(start) || !h.Stats.End.Equal(end) {
			t.Errorf("%s studied on %s..%s, want %s..%s", h.ID, h.Stats.Start, h.Stats.End, start, end)
		}
		if h.Relative == nil || !h.Stats.HasBeta {
			t.Errorf("%s: no relative statistics against the benchmark", h.ID)
		}
		if len(h.Years) == 0 || len(h.Months) == 0 {
			t.Errorf("%s: empty calendar tables", h.ID)
		}
		if !h.HasMeta {
			t.Errorf("%s: catalog record not found", h.ID)
		}
	}
	if study.Relative == nil || !study.Stats.HasBeta || !study.Stats.HasCWARP {
		t.Error("portfolio: no relative statistics against the benchmark")
	}
	if !hasWarning(study.Warnings, "VWRL: ", "distributing share class") {
		t.Errorf("warnings %q lack the distributing share class", study.Warnings)
	}
	// The distributing line's dividends travel with its series, clipped to
	// the window.
	vw := study.Holdings[1].Series
	if len(vw.Dividends) == 0 || vw.Dividends[0].Date.Before(start) {
		t.Errorf("VWRL dividends not carried on the window: %v", vw.Dividends)
	}
	// Fees the spec does not declare come from the source.
	if got := study.Portfolio.Assets[0].Fees; got != 0.2 {
		t.Errorf("IWDA fees %v, want the source's 0.2", got)
	}
}

// Options.Sim turns the spec's SIM flag on: the late holding's backcast opens
// the window earlier, with a warning, and a holding with no backcast keeps
// its real quotes.
func TestPortfolioSim(t *testing.T) {
	src := newFake()
	spec := threeFunds(t)
	study, err := analyze.Portfolio(ctx, src, spec, analyze.Options{Sim: true})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Sim {
		t.Error("the caller's spec was modified")
	}
	if !study.Spec.Sim {
		t.Error("the studied spec does not carry Sim")
	}
	if got := study.Sim.Dates[0]; !got.Equal(calStart) {
		t.Errorf("window starts %s, want the backcast's %s", got, calStart)
	}
	if !hasWarning(study.Warnings, "IGLN: ", "simulated before 2013-01-02 via simdata") {
		t.Errorf("warnings %q lack the backcast", study.Warnings)
	}
	for _, id := range []string{"IWDASIM", "VWRLSIM", "IGLNSIM"} {
		if !slices.Contains(src.fetched, id) {
			t.Errorf("%s not fetched (fetched %v)", id, src.fetched)
		}
	}
}

// Correlation is metrics.Corr of the aligned daily returns, and the aligned
// calendar is the simulation's.
func TestPortfolioCorrelation(t *testing.T) {
	study, err := analyze.Portfolio(ctx, newFake(), threeFunds(t), analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	al := study.Aligned
	if !slices.EqualFunc(al.Dates, study.Sim.Dates, time.Time.Equal) {
		t.Error("aligned calendar differs from the simulation's")
	}
	if want := []string{"IWDA", "VWRL", "IGLN"}; !slices.Equal(al.IDs, want) {
		t.Errorf("aligned IDs %v, want %v", al.IDs, want)
	}
	ret := al.Returns()
	for i := range ret {
		for j := range ret {
			want := metrics.Corr(ret[i], ret[j])
			if i == j {
				want = 1
			}
			if got := study.Correlation[i][j]; math.Abs(got-want) > 1e-12 {
				t.Errorf("Correlation[%d][%d] = %v, want %v", i, j, got, want)
			}
		}
	}
	if c := study.Correlation[0][1]; c < 0.5 {
		t.Errorf("the two world-equity lines correlate at %.2f, want a strong link", c)
	}
}

// The attribution shares each sum to one, and they are metrics.Attribute of
// the simulation's monthly contributions.
func TestPortfolioAttribution(t *testing.T) {
	study, err := analyze.Portfolio(ctx, newFake(), threeFunds(t), analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var risk, ret float64
	for i := range study.Holdings {
		risk += study.Attribution.Risk[i]
		ret += study.Attribution.Return[i]
	}
	if math.Abs(risk-1) > 1e-9 || math.Abs(ret-1) > 1e-9 {
		t.Errorf("shares sum to %v (risk) and %v (return), want 1", risk, ret)
	}
	_, monthly := study.Sim.MonthlyContributions()
	want, err := metrics.Attribute(monthly)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(study.Attribution, want) {
		t.Errorf("attribution %+v, want %+v", study.Attribution, want)
	}
}

// The composition is suggest's splits over the holdings built the way the
// comparison report builds them, number for number.
func TestPortfolioCompositionMatchesSuggest(t *testing.T) {
	spec := mustSpec(t,
		portfolio.Line{ID: "IWDA", Weight: 0.5, Fees: -1},
		portfolio.Line{ID: "IGLN", Weight: 0.2, Fees: -1},
		portfolio.Line{ID: "AGGH", Weight: 0.3, Fees: -1}) // not in the catalog
	study, err := analyze.Portfolio(ctx, newFake(), spec, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}

	meta, err := suggest.LoadMeta(bytes.NewReader(datasets.AssetMeta()))
	if err != nil {
		t.Fatal(err)
	}
	var h []suggest.Holding
	for _, a := range study.Portfolio.Assets {
		base, _ := marketdata.SplitSim(a.ID)
		m, ok := meta[marketdata.CanonicalID(base)]
		if !ok {
			m, ok = meta[base]
		}
		h = append(h, suggest.Holding{ID: base, Weight: a.Weight, Meta: m, HasMeta: ok})
	}
	sectors, equity := suggest.EquitySectorSplit(h)
	coverage, unclassified := suggest.Coverage(h, suggest.RegimeFramework())
	want := analyze.Composition{
		AssetClass:   suggest.AssetClassSplit(h),
		Geography:    suggest.GeographySplit(h),
		Currency:     suggest.CurrencySplit(h),
		Sectors:      sectors,
		Equity:       equity,
		Duration:     suggest.DurationSplit(h),
		Coverage:     coverage,
		Unclassified: unclassified,
	}
	if !reflect.DeepEqual(study.Composition, want) {
		t.Errorf("composition\n got %+v\nwant %+v", study.Composition, want)
	}
	if math.Abs(study.Composition.Unclassified-0.3) > 1e-12 {
		t.Errorf("unclassified %v, want the uncatalogued line's 0.3", study.Composition.Unclassified)
	}
	if study.Holdings[2].HasMeta {
		t.Error("AGGH has a catalog record in the study, none in the catalog")
	}
}

func TestUnknownIdentifier(t *testing.T) {
	src := newFake()
	if _, err := analyze.Asset(ctx, src, "NOPE", analyze.Options{}); !errors.Is(err, marketdata.ErrUnknownIdentifier) {
		t.Errorf("Asset: err = %v, want ErrUnknownIdentifier", err)
	}
	if _, err := analyze.Asset(ctx, src, "NOPE", analyze.Options{Sim: true}); !errors.Is(err, marketdata.ErrUnknownIdentifier) {
		t.Errorf("Asset with Sim: err = %v, want ErrUnknownIdentifier", err)
	}
	spec := mustSpec(t, portfolio.Line{ID: "IWDA", Weight: 0.5, Fees: -1}, portfolio.Line{ID: "NOPE", Weight: 0.5, Fees: -1})
	if _, err := analyze.Portfolio(ctx, src, spec, analyze.Options{Sim: true}); !errors.Is(err, marketdata.ErrUnknownIdentifier) {
		t.Errorf("Portfolio: err = %v, want ErrUnknownIdentifier", err)
	}
	if _, err := analyze.Asset(ctx, src, "IWDA", analyze.Options{Benchmark: "NOPE"}); !errors.Is(err, marketdata.ErrUnknownIdentifier) {
		t.Errorf("unknown benchmark: err = %v, want ErrUnknownIdentifier", err)
	}
}

func TestRelativeOnlyWithBenchmark(t *testing.T) {
	src := newFake()
	a, err := analyze.Asset(ctx, src, "IWDA", analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if a.Relative != nil || a.Stats.HasBeta {
		t.Error("Asset without benchmark: relative statistics set")
	}
	a, err = analyze.Asset(ctx, src, "IWDA", analyze.Options{Benchmark: "MSCIWORLD"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Relative == nil || !a.Stats.HasBeta || a.Relative.Beta != a.Stats.Beta {
		t.Errorf("Asset with benchmark: Relative %v, Stats.Beta %v", a.Relative, a.Stats.Beta)
	}
	if !slices.Contains(src.fetched, "MSCIWORLD") {
		t.Error("benchmark not fetched")
	}
	ps, err := analyze.Portfolio(ctx, src, threeFunds(t), analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if ps.Relative != nil || ps.Holdings[0].Relative != nil {
		t.Error("Portfolio without benchmark: relative statistics set")
	}
}

// A benchmark the source cannot serve right now costs the relative
// statistics, with a warning, not the study.
func TestBenchmarkOutage(t *testing.T) {
	src := newFake()
	src.outage["MSCIWORLD"] = true
	a, err := analyze.Asset(ctx, src, "IWDA", analyze.Options{Benchmark: "MSCIWORLD"})
	if err != nil {
		t.Fatal(err)
	}
	if a.Relative != nil || !hasWarning(a.Warnings, "benchmark MSCIWORLD unavailable") {
		t.Errorf("Relative %v, warnings %q", a.Relative, a.Warnings)
	}
}

// A benchmark sharing no date with the study says so.
func TestBenchmarkNoOverlap(t *testing.T) {
	src := newFake()
	src.series["MSCIWORLD"] = marketdata.Trim(src.series["MSCIWORLD"], time.Time{}, time.Date(2012, 12, 31, 0, 0, 0, 0, time.UTC))
	ps, err := analyze.Portfolio(ctx, src, threeFunds(t), analyze.Options{Benchmark: "MSCIWORLD"})
	if err != nil {
		t.Fatal(err)
	}
	if ps.Relative != nil || !hasWarning(ps.Warnings, "too few common dates") ||
		!hasWarning(ps.Warnings, "IWDA: benchmark MSCIWORLD: too few common dates") {
		t.Errorf("Relative %v, warnings %q", ps.Relative, ps.Warnings)
	}
}

// The spec's own rebalancing period wins over the options', which win over
// the default; a negative option means never.
func TestRebalancePrecedence(t *testing.T) {
	cases := []struct {
		name     string
		specDays int
		optDays  int
		want     int
	}{
		{"default", -1, 0, analyze.DefaultRebalance},
		{"options", -1, 30, 30},
		{"options never", -1, -1, 0},
		{"spec wins", 0, 30, 0},
		{"spec wins over default", 365, 0, 365},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			spec := threeFunds(t)
			spec.RebalanceDays = c.specDays
			study, err := analyze.Portfolio(ctx, newFake(), spec, analyze.Options{Rebalance: c.optDays})
			if err != nil {
				t.Fatal(err)
			}
			last := func(r *portfolio.SimResult) float64 { return r.Index[len(r.Index)-1] }
			for _, days := range []int{0, 30, 90, 365} {
				ref, err := portfolio.Simulate(study.Portfolio, days)
				if err != nil {
					t.Fatal(err)
				}
				if same := last(ref) == last(study.Sim); same != (days == c.want) {
					t.Errorf("simulated like %d days: %v, want %d days", days, same, c.want)
				}
			}
		})
	}
}

// Asset clips its window to the options' bounds and converts.
func TestAssetWindow(t *testing.T) {
	from := time.Date(2012, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2016, 6, 30, 0, 0, 0, 0, time.UTC)
	a, err := analyze.Asset(ctx, newFake(), " VWRL ", analyze.Options{From: from, To: to, Currency: "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != "VWRL" || a.Series.Currency != "EUR" {
		t.Errorf("ID %q, currency %q", a.ID, a.Series.Currency)
	}
	if a.Stats.Start.Before(from) || a.Stats.End.After(to) || a.Series.First().Date.Before(from) {
		t.Errorf("studied %s..%s, outside %s..%s", a.Stats.Start, a.Stats.End, from, to)
	}
	if got := a.Years[0]; !got.Partial || got.End.Year() != 2012 {
		t.Errorf("first year %+v, want a partial 2012", got)
	}
	if n := len(a.Years); n != 5 {
		t.Errorf("%d calendar years, want 5 (2012 to 2016)", n)
	}
	if !slices.IsSortedFunc(a.Drawdowns, func(x, y metrics.Episode) int { return x.PeakDate.Compare(y.PeakDate) }) {
		t.Error("drawdowns not chronological")
	}
	if !hasWarning(a.Warnings, "distributing share class") {
		t.Errorf("warnings %q lack the distributing share class", a.Warnings)
	}
	// The longest window: without bounds the study covers the whole series.
	full, err := analyze.Asset(ctx, newFake(), "IGLN", analyze.Options{Sim: true})
	if err != nil {
		t.Fatal(err)
	}
	if !full.Stats.Start.Equal(calStart) || !full.Stats.End.Equal(calEnd) {
		t.Errorf("unbounded study on %s..%s, want %s..%s", full.Stats.Start, full.Stats.End, calStart, calEnd)
	}
	if !hasWarning(full.Warnings, "simulated before 2013-01-02") {
		t.Errorf("warnings %q lack the backcast", full.Warnings)
	}
}

// A SIM request that fails falls back to the real quotes, as Build does.
func TestAssetSimFallback(t *testing.T) {
	src := newFake()
	src.failSim["IWDA"] = true
	a, err := analyze.Asset(ctx, src, "IWDA", analyze.Options{Sim: true})
	if err != nil {
		t.Fatal(err)
	}
	if !hasWarning(a.Warnings, "no simulated history") {
		t.Errorf("warnings %q lack the fallback", a.Warnings)
	}
	if !slices.Equal(src.fetched, []string{"IWDASIM", "IWDA"}) {
		t.Errorf("fetched %v, want the SIM request then the real one", src.fetched)
	}
}

func TestRefusedRequests(t *testing.T) {
	src := newFake()
	if _, err := analyze.Asset(ctx, src, "  ", analyze.Options{}); err == nil {
		t.Error("empty identifier accepted")
	}
	bad := analyze.Options{From: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC)}
	if _, err := analyze.Asset(ctx, src, "IWDA", bad); err == nil {
		t.Error("Asset: inverted window accepted")
	}
	if _, err := analyze.Portfolio(ctx, src, threeFunds(t), bad); err == nil {
		t.Error("Portfolio: inverted window accepted")
	}
	if _, err := analyze.Portfolio(ctx, src, nil, analyze.Options{}); err == nil {
		t.Error("nil spec accepted")
	}
	// A window the data does not reach leaves fewer than two quotes.
	late := analyze.Options{From: calEnd}
	if _, err := analyze.Asset(ctx, src, "IWDA", late); err == nil {
		t.Error("one-quote window accepted")
	}
}

// The spec's directives the study does not act on are warnings, and a levered
// spec whose financing rate the source lacks is financed at zero, with a
// warning.
func TestSpecDirectivesWarned(t *testing.T) {
	src := newFake()
	opt := parseSpec(t, "#meta optimize:max-sharpe\n60 IWDA\n40 AGGH\n")
	study, err := analyze.Portfolio(ctx, src, opt, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !hasWarning(study.Warnings, "#meta optimize is not run") {
		t.Errorf("warnings %q lack the optimize note", study.Warnings)
	}
	cur := parseSpec(t, "#meta currencies:USD,EUR\n60 IWDA\n40 AGGH\n")
	if study, err = analyze.Portfolio(ctx, src, cur, analyze.Options{Currency: "EUR"}); err != nil {
		t.Fatal(err)
	}
	if !hasWarning(study.Warnings, "#meta currencies is not expanded", "studied in EUR") {
		t.Errorf("warnings %q lack the currencies note", study.Warnings)
	}
	lev := parseSpec(t, "#meta leverage:on\n100 IWDA\n50 AGGH\n")
	if study, err = analyze.Portfolio(ctx, src, lev, analyze.Options{}); err != nil {
		t.Fatal(err)
	}
	if !hasWarning(study.Warnings, "financing rate ^IRX unavailable") || !study.Portfolio.Leverage {
		t.Errorf("warnings %q lack the financing note", study.Warnings)
	}
}

// A portfolio its withdrawals exhaust is studied up to the ruin, and says so.
func TestPortfolioRuin(t *testing.T) {
	spec := parseSpec(t, "#meta capital:1000\n#meta withdraw:100/month\n100 AGGH\n")
	study, err := analyze.Portfolio(ctx, newFake(), spec, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !study.Sim.Ruined || !hasWarning(study.Warnings, "capital wiped out", "withdrawals exhausted the capital") {
		t.Errorf("ruined %v, warnings %q", study.Sim.Ruined, study.Warnings)
	}
}
