package compare

import (
	"fmt"

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
