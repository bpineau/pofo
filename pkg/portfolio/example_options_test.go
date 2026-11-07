package portfolio_test

import (
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/portfolio"
)

func ExampleNewSpecWith() {
	spec, err := portfolio.NewSpecWith("Levered allocation", portfolio.SpecOptions{Leverage: true},
		portfolio.Line{ID: "VTI", Weight: .9}, // US9229087690
		portfolio.Line{ID: "BND", Weight: .6}) // US9219378356
	if err != nil {
		panic(err)
	}
	fmt.Printf("exposures %.0f / %.0f %%\n", spec.Holdings[0].Weight*100, spec.Holdings[1].Weight*100)
	// Output:
	// exposures 90 / 60 %
}

func ExampleSimulateWith() {
	// Synthetic prices around a calendar boundary. February's first
	// observed close executes the rebalance, without using future prices.
	dates := []time.Time{
		time.Date(2020, 1, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC),
	}
	a, err := marketdata.NewSeries("A", dates, []float64{100, 200, 200})
	if err != nil {
		panic(err)
	}
	b, err := marketdata.NewSeries("B", dates, []float64{100, 100, 100})
	if err != nil {
		panic(err)
	}
	p := &portfolio.Portfolio{Capital: 1000, Assets: []portfolio.Asset{
		{Symbol: "A", Weight: .5, Series: a}, {Symbol: "B", Weight: .5, Series: b},
	}}
	sim, err := portfolio.SimulateWith(p, portfolio.SimOptions{
		RebalancePeriod: portfolio.Monthly, TransactionCost: .01,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("January turnover %.1f %%, February turnover %.1f %%\n", sim.Turnover[1]*100, sim.Turnover[2]*100)
	fmt.Printf("cost %.2f, net value %.2f, index %.2f\n", sim.Costs[2], sim.Values[2], sim.Index[2])
	// Output:
	// January turnover 0.0 %, February turnover 33.3 %
	// cost 5.00, net value 1495.00, index 149.50
}
