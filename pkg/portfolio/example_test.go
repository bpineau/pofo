package portfolio_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// Parse reads a portfolio description: "<weight %> <identifier> [TER %/year]",
// everything after a # being a comment.
func ExampleParse() {
	spec, err := portfolio.Parse("my-portfolio", strings.NewReader(`
# Comment lines and blank lines are ignored.
60   VTI    0.03            # optional TER, then a free-text comment
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

// A "#meta currencies" directive asks the CLI to evaluate the portfolio in
// several base currencies at once, one comparison column each.
func ExampleParse_currencies() {
	spec, _ := portfolio.Parse("dragon", strings.NewReader(
		"#meta currencies:USD,EUR\n60 NTSGSIM\n40 XAUUSDSIM\n"))
	fmt.Println(spec.Currencies)
	// Output: [USD EUR]
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

// With a starting capital and periodic flows, Simulate tracks two series:
// Values follows the money (contributions included), while Index is the
// time-weighted return, the one to use for statistics and comparisons.
func ExampleSimulate_flows() {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &marketdata.Series{Symbol: "FLAT"}
	for i := range 95 {
		s.Points = append(s.Points, marketdata.Point{Date: start.AddDate(0, 0, i), Close: 10})
	}
	p := &portfolio.Portfolio{
		Name:       "dca",
		Assets:     []portfolio.Asset{{ID: "FLAT", Symbol: "FLAT", Weight: 1, Fees: -1, Series: s}},
		Capital:    1000,
		Contribute: portfolio.Flow{Amount: 100, Period: portfolio.Monthly},
	}
	sim, err := portfolio.Simulate(p, 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("contributed %.0f, final value %.0f, index %.0f\n",
		sim.Contributed, sim.Values[len(sim.Values)-1], sim.Index[len(sim.Index)-1])
	// Output:
	// contributed 300, final value 1300, index 100
}

// Build turns a parsed Spec into a simulatable Portfolio through a fetch
// callback. Against live data the callback is one line on a
// marketdata.Client: client.FetchExtended(id, marketdata.FetchOptions{Currency: "EUR"});
// here it serves synthetic series so the example runs offline.
func ExampleBuild() {
	spec, err := portfolio.Parse("demo", strings.NewReader(`
#meta rebalance:30
60 EQ
40 BD 0.15
`))
	if err != nil {
		panic(err)
	}
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	fetch := func(id string) (*marketdata.Series, error) {
		s := &marketdata.Series{Symbol: id, Currency: "EUR"}
		for i := range 60 {
			s.Points = append(s.Points, marketdata.Point{Date: start.AddDate(0, 0, i), Close: 100 + float64(i)})
		}
		return s, nil
	}
	p, err := portfolio.Build(spec, portfolio.BuildOptions{Fetch: fetch, BaseCurrency: "EUR"})
	if err != nil {
		panic(err)
	}
	days := spec.RebalanceDays // -1 would mean "apply your own default"
	sim, err := portfolio.Simulate(p, days)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: %d assets, %d points, final index %.1f\n",
		p.Name, len(p.Assets), len(sim.Index), sim.Index[len(sim.Index)-1])
	// Output:
	// demo: 2 assets, 60 points, final index 159.0
}
