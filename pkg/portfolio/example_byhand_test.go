package portfolio_test

import (
	"fmt"
	"math"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// The manual path, one step at a time: a spec built in code, Build through a
// fetch callback (synthetic here; against live data one line on a
// marketdata.Client, client.FetchExtended(ctx, id, opts)), a starting capital
// and a monthly contribution, the simulation, then the statistics on the
// time-weighted Index and the money-weighted return on Values and the flows.
func Example_byHand() {
	spec, _ := portfolio.NewSpec("60/40",
		portfolio.Line{ID: "IWDA", Weight: 0.6}, // IE00B4L5Y983
		portfolio.Line{ID: "AGGH", Weight: 0.4}) // IE00BDBRDM35
	p, err := portfolio.Build(spec, portfolio.BuildOptions{Fetch: synthetic})
	if err != nil {
		panic(err)
	}
	p.Capital = 10_000
	p.Contribute = portfolio.Flow{Amount: 500, Period: portfolio.Monthly}
	sim, err := portfolio.Simulate(p, 90) // rebalance every 90 days
	if err != nil {
		panic(err)
	}

	stats, _ := metrics.Compute(sim.Dates, sim.Index) // the strategy, flows stripped out
	fmt.Printf("CAGR %.1f %%, volatility %.1f %%, max drawdown %.1f %%\n", stats.CAGR*100, stats.Volatility*100, stats.MaxDrawdown*100)

	// The saver's own rate: money going in is negative, the final value closes the account.
	dates, flows := []time.Time{sim.Dates[0]}, []float64{-p.Capital}
	var booked []metrics.Flow // the same flows the other way round, for TWR
	for i, d := range sim.FlowDates {
		dates, flows = append(dates, d), append(flows, -sim.FlowAmounts[i])
		booked = append(booked, metrics.Flow{Date: d, Amount: sim.FlowAmounts[i]})
	}
	last := len(sim.Dates) - 1
	irr, _ := metrics.IRR(dates, flows, sim.Dates[last], sim.Values[last])
	twr, _ := metrics.TWR(sim.Dates, sim.Values, booked)
	fmt.Printf("put in %.0f, worth %.0f, money-weighted %.1f %%/yr\n", p.Capital+sim.Contributed, sim.Values[last], irr*100)
	fmt.Printf("time-weighted %+.1f %%, as the index says: %+.1f %%\n", twr*100, sim.Index[last]-100)
	// Output:
	// CAGR 10.7 %, volatility 9.5 %, max drawdown -5.3 %
	// put in 27500, worth 33664, money-weighted 10.2 %/yr
	// time-weighted +35.7 %, as the index says: +35.7 %
}

// synthetic serves three years of daily closes: a swinging equity line and a
// calmer bond line.
func synthetic(id string) (*marketdata.Series, error) {
	drift, swing := 0.0004, 0.012
	if id == "AGGH" {
		drift, swing = 0.0001, 0.003
	}
	var dates []time.Time
	var closes []float64
	level := 100.0
	for i := range 3 * 365 {
		dates = append(dates, time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i))
		closes = append(closes, level)
		level *= 1 + drift + swing*math.Sin(float64(i)*0.3)
	}
	return marketdata.NewSeries(id, dates, closes)
}
