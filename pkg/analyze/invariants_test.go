package analyze_test

import (
	"context"
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func TestPortfolioPriceScaleAndAssetPermutation(t *testing.T) {
	a := walk(t, "A", 520)
	b := scaled(t, a, "B", 1)
	for i := range b.Points {
		b.Points[i].Close *= 1 + .05*math.Sin(float64(i)/10)
	}
	study := func(scaleA, scaleB float64, reverse bool) *analyze.PortfolioStudy {
		src := analyze.SeriesSource{"A": scaled(t, a, "A", scaleA), "B": scaled(t, b, "B", scaleB)}
		lines := []portfolio.Line{{ID: "A", Weight: .3}, {ID: "B", Weight: .7}}
		if reverse {
			lines[0], lines[1] = lines[1], lines[0]
		}
		sp, err := portfolio.NewSpec("invented allocation", lines...)
		if err != nil {
			t.Fatal(err)
		}
		out, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	want := study(1, 1, false)
	for _, got := range []*analyze.PortfolioStudy{study(100, .01, false), study(.01, 100, true)} {
		for i, v := range want.Sim.Index {
			if math.Abs(v-got.Sim.Index[i]) > 1e-9 {
				t.Fatalf("index changed at %d", i)
			}
		}
		for name, pair := range map[string][2]float64{"CAGR": {want.Stats.CAGR, got.Stats.CAGR}, "volatility": {want.Stats.Volatility, got.Stats.Volatility}, "Sharpe": {want.Stats.Sharpe, got.Stats.Sharpe}, "drawdown": {want.Stats.MaxDrawdown, got.Stats.MaxDrawdown}} {
			if math.Abs(pair[0]-pair[1]) > 1e-10 {
				t.Errorf("%s changed: %v", name, pair)
			}
		}
	}
}

func TestExternalFlowsDoNotBecomeSingleAssetPerformance(t *testing.T) {
	s := walk(t, "A", 520)
	src := analyze.SeriesSource{"A": s}
	sp, err := portfolio.NewSpec("invented allocation", portfolio.Line{ID: "A", Weight: 1})
	if err != nil {
		t.Fatal(err)
	}
	sp.Capital = 1000
	base, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	sp.Contribute = portfolio.Flow{Amount: 50, Period: portfolio.Monthly}
	flow, err := analyze.Portfolio(context.Background(), src, sp, analyze.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !(flow.Sim.Values[len(flow.Sim.Values)-1] > base.Sim.Values[len(base.Sim.Values)-1]) {
		t.Fatal("contribution did not increase money outcome")
	}
	if math.Abs(flow.Stats.CAGR-base.Stats.CAGR) > 1e-12 || math.Abs(flow.Stats.Sharpe-base.Stats.Sharpe) > 1e-10 {
		t.Fatal("external flows changed time-weighted statistics")
	}
}
