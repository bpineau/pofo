package compare

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// ExampleComparison_HTMLPage renders a tiny two-portfolio comparison to a
// report.Page. Real callers build the Comparison with Compute (which fetches
// prices); here the columns are fabricated so the example stays deterministic
// and offline. The Page carries one section per column.
func ExampleComparison_HTMLPage() {
	dates := months(6)
	col := func(name, color string) *column {
		values := make([]float64, len(dates))
		for i := range values {
			values[i] = 100 + float64(i)
		}
		stats, _ := metrics.Compute(dates, values)
		return &column{
			p:             &portfolio.Portfolio{Name: name},
			sim:           &portfolio.SimResult{Dates: dates, Values: values},
			color:         color,
			rebalanceDays: 90,
			currency:      "EUR",
			specName:      name,
			winDates:      dates,
			winValues:     values,
			stats:         stats,
		}
	}
	cols := []*column{col("Alpha", "#1f6f78"), col("Beta", "#b8563e")}
	c := newTestComparison(cols, nil, dates[0], dates[len(dates)-1], nil, Options{Rebalance: 90})

	page := c.HTMLPage(Decoration{})
	fmt.Println(len(page.Portfolios))
	// Output: 2
}

// ExampleSweep shows what one line's weight buys and costs: the other
// holdings keep their relative proportions, and the real simulation is re-run
// at every point of the grid. It is how a portfolio file's per-sleeve "sane
// range" is established.
func ExampleSweep() {
	dates := make([]time.Time, 120)
	for i, d := 0, time.Date(2000, 1, 31, 0, 0, 0, 0, time.UTC); i < len(dates); i++ {
		dates[i] = d
		d = d.AddDate(0, 1, 0)
	}
	grower := make([]marketdata.Point, len(dates))
	flat := make([]marketdata.Point, len(dates))
	for i, d := range dates {
		grower[i] = marketdata.Point{Date: d, Close: 100 * math.Pow(1.006, float64(i))}
		flat[i] = marketdata.Point{Date: d, Close: 100}
	}
	p := &portfolio.Portfolio{Name: "two lines", Assets: []portfolio.Asset{
		{ID: "GROW", Symbol: "GROW", Weight: 0.6, Fees: -1,
			Series: &marketdata.Series{Symbol: "GROW", Currency: "EUR", Points: grower}},
		{ID: "FLAT", Symbol: "FLAT", Weight: 0.4, Fees: -1,
			Series: &marketdata.Series{Symbol: "FLAT", Currency: "EUR", Points: flat}},
	}}

	holdings, err := Sweep(p, SweepOptions{Weights: []float64{0.25, 0.5, 0.75}})
	if err != nil {
		log.Fatal(err)
	}
	for _, pt := range holdings[0].Points {
		mark := ""
		if pt.Written {
			mark = " (as written)"
		}
		fmt.Printf("GROW at %.0f %%: CAGR %.1f %%%s\n", pt.Weight*100, pt.Stats.CAGR*100, mark)
	}
	// Output:
	// GROW at 25 %: CAGR 2.4 %
	// GROW at 50 %: CAGR 4.3 %
	// GROW at 60 %: CAGR 5.0 % (as written)
	// GROW at 75 %: CAGR 6.0 %
}
