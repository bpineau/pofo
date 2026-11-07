package analyze_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func TestPortfolioCorrelationSharedIntervals(t *testing.T) {
	daily := &marketdata.Series{Symbol: "DAILY", Currency: "USD"}
	weekly := &marketdata.Series{Symbol: "WEEKLY", Currency: "USD"}
	for i, d := range calendar() {
		p := marketdata.Point{Date: d, Close: 100 + float64(i%13)}
		daily.Points = append(daily.Points, p)
		if d.Weekday() == time.Friday {
			weekly.Points = append(weekly.Points, p)
		}
	}
	src := &fakeSource{series: map[string]*marketdata.Series{"DAILY": daily, "WEEKLY": weekly}}
	sp, _ := portfolio.NewSpec("same path", portfolio.Line{ID: "DAILY", Weight: .5}, portfolio.Line{ID: "WEEKLY", Weight: .5})
	got, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Correlation[0][1]-1) > 1e-12 {
		t.Fatalf("same path correlation %v, want 1", got.Correlation[0][1])
	}
}

func TestAssetLoadsBundledReferences(t *testing.T) {
	c := marketdata.NewClient("")
	c.Offline = true
	s, err := analyze.Asset(context.Background(), c, "SP500-USD", analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Series.Len() < 100 {
		t.Fatalf("short study: %d", s.Series.Len())
	}
}

func TestSeriesStudyRealExcessAndJSON(t *testing.T) {
	ds := []time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)}
	asset, _ := marketdata.NewSeries("ASSET", ds, []float64{100, 110, 121})
	cpi, _ := marketdata.NewSeries("CPI", ds, []float64{100, 105, 110.25})
	cash, _ := marketdata.NewSeries("CASH", ds, []float64{1, 1.02, 1.0404})
	s, err := analyze.Series(asset, nil, analyze.Options{Inflation: cpi, RiskFreeIndex: cash})
	if err != nil {
		t.Fatal(err)
	}
	if s.Real == nil || math.Abs(s.Real.Series.Last().Close-109.75056689342404) > 1e-10 {
		t.Fatalf("real study: %+v", s.Real)
	}
	if s.Excess == nil || math.Abs(s.Excess.Series.Last().Close-1.1664) > 1e-10 {
		t.Fatalf("excess study: %+v", s.Excess)
	}
	flat, _ := marketdata.NewSeries("FLAT", ds, []float64{100, 100, 100})
	fs, err := analyze.Series(flat, nil, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(fs)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte(`"Sharpe":null`)) {
		t.Fatalf("missing-value semantics: %s", b)
	}
}

func TestSeriesSourceAndOptions(t *testing.T) {
	ds := []time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)}
	s, _ := marketdata.NewSeries("A", ds, []float64{1, 1.1, .99})
	s.Currency = "EUR"
	src := analyze.SeriesSource{"A": s}
	cp, err := src.FetchExtended(context.Background(), "A", marketdata.FetchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cp.Points[0].Close = 9
	if s.First().Close != 1 {
		t.Fatal("source aliased caller's points")
	}
	if _, err := analyze.Series(s, nil, analyze.Options{Currency: "USD"}); err == nil {
		t.Fatal("silent currency conversion")
	}
	if _, err := analyze.Series(nil, nil, analyze.Options{}); err == nil {
		t.Fatal("nil accepted")
	}
	if _, err := analyze.Series(s, nil, analyze.Options{Benchmark: "missing"}); err == nil {
		t.Fatal("absent benchmark accepted")
	}
	for _, o := range []analyze.Options{{RiskFree: math.NaN()}, {RiskFree: .02, RiskFreeIndex: s}, {Financing: s, FinancingID: "rate"}, {CorrelationFrequency: -1}, {Adjustment: marketdata.AdjustmentOptions{MaxAge: -1}}} {
		if _, err := analyze.Series(s, nil, o); err == nil {
			t.Fatalf("accepted %+v", o)
		}
	}
	a, err := analyze.Series(s, nil, analyze.Options{RiskFree: .06})
	if err != nil {
		t.Fatal(err)
	}
	zero, err := analyze.Series(s, nil, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs((zero.Stats.Sharpe-a.Stats.Sharpe)-.06/a.Stats.Volatility) > 1e-12 || a.Stats.CAGR != zero.Stats.CAGR {
		t.Fatal("risk-free affects wrong metrics")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := src.FetchExtended(ctx, "A", marketdata.FetchOptions{}); err == nil {
		t.Fatal("cancelled context ignored")
	}
	if _, err := src.FetchExtended(context.Background(), "missing", marketdata.FetchOptions{}); err == nil {
		t.Fatal("unknown identifier accepted")
	}
}

func TestPortfolioStudyAdjustmentsAndJSON(t *testing.T) {
	src := newFake()
	ref := src.series["IWDA"]
	ds := ref.Dates()
	levels := make([]float64, len(ds))
	for i := range levels {
		levels[i] = 100 + float64(i)/100
	}
	cpi, _ := marketdata.NewSeries("CPI", ds, levels)
	sp, _ := portfolio.NewSpec("example", portfolio.Line{ID: "IWDA", Weight: .6}, portfolio.Line{ID: "AGGH", Weight: .4})
	st, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{Inflation: cpi, RiskFree: .02})
	if err != nil {
		t.Fatal(err)
	}
	if st.Real == nil || st.Holdings[0].Real == nil {
		t.Fatal("missing real studies")
	}
	if _, err := json.Marshal(st); err != nil {
		t.Fatal(err)
	}
}

func TestPortfolioExplicitSimulation(t *testing.T) {
	src := newFake()
	spec, _ := portfolio.NewSpec("example", portfolio.Line{ID: "IWDA", Weight: .6}, portfolio.Line{ID: "AGGH", Weight: .4})
	settings := portfolio.SimOptions{RebalancePeriod: portfolio.Monthly, TransactionCost: .001}
	charged, err := analyze.Portfolio(context.Background(), src, spec, analyze.Options{Simulation: &settings})
	if err != nil {
		t.Fatal(err)
	}
	sum := 0.0
	for _, cost := range charged.Sim.Costs {
		sum += cost
	}
	if sum <= 0 || charged.Sim.Costs[0] != 0 {
		t.Fatal("execution costs not passed to simulator")
	}
	settings.TransactionCost = 0
	free, err := analyze.Portfolio(context.Background(), src, spec, analyze.Options{Simulation: &settings})
	if err != nil {
		t.Fatal(err)
	}
	if free.Stats.CAGR <= charged.Stats.CAGR {
		t.Fatal("costs do not reduce time-weighted performance")
	}
	if _, err := analyze.Portfolio(context.Background(), src, spec, analyze.Options{Simulation: &settings, Rebalance: 30}); err == nil {
		t.Fatal("conflicting rebalance options accepted")
	}
	spec.RebalanceDays = 0
	if _, err := analyze.Portfolio(context.Background(), src, spec, analyze.Options{Simulation: &settings}); err == nil {
		t.Fatal("conflicting spec directive accepted")
	}
}

func TestPortfolioExplicitFinancing(t *testing.T) {
	src := newFake()
	spec, err := portfolio.NewSpecWith("leveraged example", portfolio.SpecOptions{Leverage: true}, portfolio.Line{ID: "IWDA", Weight: .9}, portfolio.Line{ID: "AGGH", Weight: .6})
	if err != nil {
		t.Fatal(err)
	}
	ds := src.series["IWDA"].Dates()
	values := make([]float64, len(ds))
	for i := range values {
		values[i] = 2
	}
	cash, _ := marketdata.NewSeries("RATE", ds, values)
	study, err := analyze.Portfolio(context.Background(), src, spec, analyze.Options{Financing: cash})
	if err != nil {
		t.Fatal(err)
	}
	if study.Portfolio.Cash != cash || study.Portfolio.Assets[0].Weight != .9 {
		t.Fatal("financing or exposure not passed through")
	}
	if _, err := analyze.Portfolio(context.Background(), src, spec, analyze.Options{FinancingID: "not supplied"}); err == nil {
		t.Fatal("missing explicit financing was silently ignored")
	}
}

func TestPortfolioNativeCashCurrency(t *testing.T) {
	ds := []time.Time{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)}
	asset, _ := marketdata.NewSeries("A", ds, []float64{100, 110, 105})
	asset.Currency = "EUR"
	cash, _ := marketdata.NewSeries("cash", ds, []float64{1, 1.01, 1.02})
	cash.Currency = "USD"
	sp, _ := portfolio.NewSpec("example", portfolio.Line{ID: "A", Weight: 1})
	src := analyze.SeriesSource{"A": asset}
	if _, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{RiskFreeIndex: cash}); err == nil {
		t.Fatal("native EUR portfolio accepted USD cash")
	}
	cash.Currency = "EUR"
	s, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{RiskFreeIndex: cash})
	if err != nil {
		t.Fatal(err)
	}
	if s.Excess.Series.Currency != "EUR" {
		t.Fatal("native denomination lost")
	}
	asset.Currency = ""
	if _, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{RiskFreeIndex: cash}); err == nil {
		t.Fatal("unknown native currency compared with EUR cash")
	}
}

func TestFinancingIDRetainsRateBeforeWindow(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2024, 1, d, 0, 0, 0, 0, time.UTC) }
	asset, _ := marketdata.NewSeries("A", []time.Time{day(5), day(6), day(10), day(11)}, []float64{100, 110, 105, 106})
	asset.Currency = "EUR"
	rate, _ := marketdata.NewSeries("RATE", []time.Time{day(1), day(10)}, []float64{1, 10})
	sp, _ := portfolio.NewSpecWith("example", portfolio.SpecOptions{Leverage: true}, portfolio.Line{ID: "A", Weight: 2})
	src := analyze.SeriesSource{"A": asset, "RATE": rate}
	pointer, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{From: day(5), Financing: rate})
	if err != nil {
		t.Fatal(err)
	}
	named, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{From: day(5), FinancingID: "RATE"})
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range pointer.Sim.Index {
		if v != named.Sim.Index[i] {
			t.Fatalf("named financing looks ahead: %v vs %v", pointer.Sim.Index, named.Sim.Index)
		}
	}
}

func TestAdjustedPortfolioRetainsHoldingWarnings(t *testing.T) {
	src := newFake()
	src.series["IWDA"].Warnings = []string{"qualified input"}
	dates := src.series["IWDA"].Dates()
	levels := make([]float64, len(dates))
	for i := range levels {
		levels[i] = 100
	}
	cpi, _ := marketdata.NewSeries("basket", dates, levels)
	spec, _ := portfolio.NewSpec("example", portfolio.Line{ID: "IWDA", Weight: 1})
	study, err := analyze.Portfolio(context.Background(), src, spec, analyze.Options{Inflation: cpi})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range study.Real.Series.Warnings {
		if w == "IWDA: qualified input" {
			found = true
		}
	}
	if !found {
		t.Fatal("exportable real curve lost holding qualification")
	}
}
