package remote

import (
	"context"
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func fixtureSource() analyze.SeriesSource {
	src := analyze.SeriesSource{}
	for j, id := range []string{"AAA", "BBB", "BENCH"} {
		s := &marketdata.Series{Symbol: id, Currency: "EUR", Source: "fixture"}
		v := 100.
		for k := 0; k < 84; k++ {
			v *= 1 + .005 + .03*math.Sin(float64(k+j))
			s.Points = append(s.Points, marketdata.Point{Date: time.Date(2018, time.Month(1+k), 28, 0, 0, 0, 0, time.UTC), Close: v})
		}
		if j == 1 {
			s.Points = s.Points[12:72]
		}
		src[id] = s
	}
	return src
}
func TestAssetDirectParityAndNull(t *testing.T) {
	src := fixtureSource()
	s := New(src, "test")
	result, err := s.Execute(context.Background(), "analyze_assets", []byte(`{"ids":["AAA"],"benchmark_id":"BENCH","sections":["identity","statistics","annual_returns","drawdowns"]}`))
	if err != nil {
		t.Fatal(err)
	}
	a := result.Data.(AssetsData).Assets[0]
	direct, err := analyze.Asset(context.Background(), src, "AAA", analyze.Options{Currency: "EUR", Benchmark: "BENCH", ExactOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if float64(a.Statistics.CAGR) != direct.Stats.CAGR*100 || a.Relative == nil || float64(a.Relative.Beta) != direct.Relative.Beta {
		t.Fatalf("parity: %+v", a)
	}
	b, err := json.Marshal(Number(math.NaN()))
	if err != nil || string(b) != "null" {
		t.Fatalf("undefined: %s %v", b, err)
	}
}
func TestPortfoliosCommonWindowFeeAndBuyHold(t *testing.T) {
	src := fixtureSource()
	s := New(src, "test")
	input := `{"portfolios":[{"holdings":[{"id":"AAA","weight_pct":60},{"id":"BBB","weight_pct":40}],"rebalance_days":0,"annual_fee_pct":1},{"holdings":[{"id":"AAA","weight_pct":100}]}],"sections":["statistics","composition","correlations","attribution"]}`
	r, err := s.Execute(context.Background(), "analyze_portfolio", []byte(input))
	if err != nil {
		t.Fatal(err)
	}
	out := r.Data.(PortfoliosData)
	if r.Meta.EffectiveWindow.From != "2019-01-28" || r.Meta.EffectiveWindow.To != "2023-12-28" {
		t.Fatalf("window %+v", r.Meta.EffectiveWindow)
	}
	spec, err := portfolio.NewSpecWith("", portfolio.SpecOptions{RebalanceDays: ptr(0), EnvelopeFees: ptr(1.)}, portfolio.Line{ID: "AAA", Weight: .6}, portfolio.Line{ID: "BBB", Weight: .4})
	if err != nil {
		t.Fatal(err)
	}
	d, err := analyze.Portfolio(context.Background(), src, spec, analyze.Options{Currency: "EUR", CorrelationFrequency: marketdata.Monthly, ExactOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if float64(out.Portfolios[0].Statistics.CAGR) != d.Stats.CAGR*100 {
		t.Fatal("fee/buy-hold parity")
	}
	if out.Portfolios[0].Statistics.Years != out.Portfolios[1].Statistics.Years {
		t.Fatal("different windows")
	}
	if _, err = json.Marshal(r); err != nil {
		t.Fatal(err)
	}
}
func TestIdentityOnlyDoesNotLoad(t *testing.T) {
	s := New(nil, "test")
	r, err := s.Execute(context.Background(), "analyze_assets", []byte(`{"ids":["IWDA"],"sections":["identity","metadata"],"data_policy":"cache_only"}`))
	if err != nil {
		t.Fatal(err)
	}
	d := r.Data.(AssetsData).Assets[0]
	if d.Identity == nil || d.Statistics != nil || d.Provenance != nil {
		t.Fatalf("%+v", d)
	}
}
func TestAtomicPortfolioValidation(t *testing.T) {
	s := New(nil, "test")
	called := false
	s.BeforeFetch = func(context.Context, string) error { called = true; return nil }
	_, err := s.Execute(context.Background(), "analyze_portfolio", []byte(`{"portfolios":[{"holdings":[{"id":"IWDA","weight_pct":100}]},{"holdings":[{"id":"IWDA","weight_pct":99}]}]}`))
	if err == nil || called {
		t.Fatalf("validation did IO: err=%v calls=%v", err, called)
	}
}

func TestComparisonRejectsDisjointQuotedCalendars(t *testing.T) {
	src := fixtureSource()
	for i := range src["BBB"].Points {
		src["BBB"].Points[i].Date = src["BBB"].Points[i].Date.AddDate(0, 0, 1)
	}
	s := New(src, "test")
	_, err := s.Execute(context.Background(), "analyze_portfolio", []byte(`{"portfolios":[{"holdings":[{"id":"AAA","weight_pct":100}]},{"holdings":[{"id":"BBB","weight_pct":100}]}]}`))
	if err == nil || publicError(err).Code != "no_common_period" {
		t.Fatalf("compared different calendars: %v", err)
	}
}
