package analyze_test

import (
	"context"
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/analyze"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// A 60/40 in two lines: build the spec in code, study it. Against live data
// the source is a marketdata.Client,
//
//	src := marketdata.NewClient(marketdata.DefaultCacheDir())
//
// here it is an offline fake serving synthetic series, so the example runs
// without the network.
func Example_sixtyForty() {
	ctx := context.Background()
	src := newFake()

	spec, _ := portfolio.NewSpec("60/40",
		portfolio.Line{ID: "IWDA", Weight: 0.6}, // IE00B4L5Y983
		portfolio.Line{ID: "AGGH", Weight: 0.4}) // IE00BDBRDM35
	study, err := analyze.Portfolio(ctx, src, spec, analyze.Options{Currency: "EUR"})
	if err != nil {
		panic(err)
	}

	st := study.Stats
	fmt.Printf("CAGR %.1f %%, volatility %.1f %%, max drawdown %.1f %%\n", st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)
	fmt.Printf("correlation %s/%s: %.2f\n", study.Aligned.IDs[0], study.Aligned.IDs[1], study.Correlation[0][1])
	// Output:
	// CAGR 7.8 %, volatility 9.2 %, max drawdown -14.1 %
	// correlation IWDA/AGGH: -0.16
}

// Asset dissects one asset on its longest window inside the options' bounds:
// statistics, calendar tables, drawdown episodes and, with a benchmark, the
// relative statistics.
func ExampleAsset() {
	ctx := context.Background()
	src := newFake()

	a, err := analyze.Asset(ctx, src, "IWDA", analyze.Options{
		Currency:  "EUR",
		Benchmark: "MSCIWORLD",
		From:      time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s (%s), %s to %s, in %s\n", a.ID, a.Meta.Name,
		a.Stats.Start.Format(time.DateOnly), a.Stats.End.Format(time.DateOnly), a.Series.Currency)
	fmt.Printf("CAGR %.1f %%, volatility %.1f %%, beta %.2f\n", a.Stats.CAGR*100, a.Stats.Volatility*100, a.Relative.Beta)
	for _, y := range a.Years[:2] {
		fmt.Printf("%d: %+.1f %% (partial: %v)\n", y.End.Year(), y.Return*100, y.Partial)
	}
	fmt.Printf("%d drawdown episodes, the deepest %.1f %%\n", len(a.Drawdowns), a.Stats.MaxDrawdown*100)
	// Output:
	// IWDA (iShares Core MSCI World UCITS ETF USD (Acc)), 2015-01-01 to 2019-12-31, in EUR
	// CAGR 13.1 %, volatility 15.4 %, beta 0.92
	// 2015: +28.6 % (partial: true)
	// 2016: +17.3 % (partial: false)
	// 37 drawdown episodes, the deepest -23.9 %
}

// Portfolio simulates the spec on the window every holding quotes, studies
// each holding on that same window, and ties them together: correlation,
// risk and return attribution, look-through composition, and the warnings
// that say what the numbers cannot.
func ExamplePortfolio() {
	ctx := context.Background()
	src := newFake()

	spec, _ := portfolio.NewSpec("three funds",
		portfolio.Line{ID: "IWDA", Weight: 0.5}, // IE00B4L5Y983
		portfolio.Line{ID: "VWRL", Weight: 0.2}, // IE00B3RBWM25
		portfolio.Line{ID: "IGLN", Weight: 0.3}) // IE00B4ND3602
	study, err := analyze.Portfolio(ctx, src, spec, analyze.Options{Currency: "EUR"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("window %s to %s, rebalanced every %d days\n",
		study.Stats.Start.Format(time.DateOnly), study.Stats.End.Format(time.DateOnly), analyze.DefaultRebalance)
	for i, h := range study.Holdings {
		fmt.Printf("%-4s CAGR %5.1f %%  risk share %4.0f %%  return share %4.0f %%\n",
			h.ID, h.Stats.CAGR*100, study.Attribution.Risk[i]*100, study.Attribution.Return[i]*100)
	}
	fmt.Printf("portfolio CAGR %.1f %%, equity %.0f %% of capital\n", study.Stats.CAGR*100, study.Composition.Equity*100)
	for _, w := range study.Warnings {
		fmt.Println("warning:", w)
	}
	// Output:
	// window 2013-01-02 to 2019-12-31, rebalanced every 90 days
	// IWDA CAGR   9.8 %  risk share   60 %  return share   49 %
	// VWRL CAGR   9.1 %  risk share   22 %  return share   18 %
	// IGLN CAGR  10.7 %  risk share   18 %  return share   33 %
	// portfolio CAGR 10.6 %, equity 70 % of capital
	// warning: VWRL: distributing share class quoted as a NAV (ft): the income it pays out is missing from every statistic
}
