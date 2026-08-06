package portfolio_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/bpineau/portfodor/pkg/marketdata"
	"github.com/bpineau/portfodor/pkg/metrics"
	"github.com/bpineau/portfodor/pkg/portfolio"
)

// Parse reads a portfolio description: "<weight %> <identifier>
// [free text]", everything after a # being a comment.
func ExampleParse() {
	spec, err := portfolio.Parse("my-portfolio", strings.NewReader(`
# Comment lines and blank lines are ignored.
60   VTI    US stocks       # free text accepted
25,5 IE00B4L5Y983           # decimal comma accepted
14.5 GLD
`))
	if err != nil {
		panic(err)
	}
	for _, h := range spec.Holdings {
		fmt.Printf("%5.1f %% %s\n", h.Weight*100, h.ID)
	}
	// Output:
	//  60.0 % VTI
	//  25.5 % IE00B4L5Y983
	//  14.5 % GLD
}

// Simulate replays a portfolio (base 100, periodic rebalancing) on series
// obtained from marketdata or built by hand. Weights are FRACTIONS summing
// to 1; fees fields are PERCENT per year. The result chains directly into
// metrics.Compute.
func ExampleSimulate() {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	mk := func(symbol string, daily float64) *marketdata.Series {
		s := &marketdata.Series{Symbol: symbol}
		v := 100.0
		for i := range 504 { // ~2 years of trading days
			s.Points = append(s.Points, marketdata.Point{Date: start.AddDate(0, 0, i), Close: v})
			v *= 1 + daily
		}
		return s
	}
	p := &portfolio.Portfolio{
		Name: "60/40",
		Assets: []portfolio.Asset{
			{ID: "EQ", Symbol: "EQ", Weight: 0.60, Fees: -1, Series: mk("EQ", 0.0004)},
			{ID: "BD", Symbol: "BD", Weight: 0.40, Fees: -1, Series: mk("BD", 0.0001)},
		},
	}
	sim, err := portfolio.Simulate(p, 90) // rebalance every 90 calendar days
	if err != nil {
		panic(err)
	}
	stats, err := metrics.Compute(sim.Dates, sim.Values)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d points, final value %.0f, max drawdown %.1f %%\n",
		len(sim.Values), sim.Values[len(sim.Values)-1], stats.MaxDrawdown*100)
	// Output:
	// 504 points, final value 115, max drawdown 0.0 %
}
