package simgen

import (
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// Composite builds a constant-weight base-100 index from arbitrary
// components, here a 90/60 equities/bonds blend with an "excess" leg
// (futures) financed at the cash rate and 0.20%/yr fees.
func ExampleComposite() {
	fetch := fakeFetcher{
		"ACTIONS": mkSeries("ACTIONS", 300, 0.0008),
		"OBLIG":   mkSeries("OBLIG", 300, 0.0002),
		"^IRX":    mkLevels("^IRX", 300, 3.0), // annualized rate in %
	}
	fr, err := BuildFrame(fetch, []string{"ACTIONS", "OBLIG", "^IRX"}, day(0))
	if err != nil {
		panic(err)
	}
	values, err := Composite(fr, []Leg{
		{ID: "ACTIONS", Weight: 0.90},
		{ID: "OBLIG", Weight: 0.60, Excess: true},
		{ID: "^IRX", Weight: 0.10},
	}, "^IRX", 0.0020)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d points, base %.0f\n", len(values), values[0])
	// Output:
	// 300 points, base 100
}

// TSMOM replays a configurable time-series momentum strategy on a basket
// of markets. (API example; not executed.)
func Example_tsmom() {
	var fetch Fetcher // e.g. marketdata.NewClient("data")
	fr, _ := BuildFrame(fetch, []string{"^IRX", "VFINX", "GC=F"}, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	values, start, _ := TSMOM(fr, TSMOMConfig{
		Markets:  []string{"VFINX", "GC=F"},
		CashID:   "^IRX",
		Lookback: 252, VolWindow: 63, Rebalance: 21,
		TargetVol: 0.10, MaxLeverage: 2,
	})
	_ = values[start:]
}

// Audit grades a recipe's engine against the asset's real quotes over the
// window where they overlap, which is the only window a backcast can be
// judged on. The CLI wraps it as "pofo -verify-simdata".
func ExampleAudit() {
	real := quoted(mkWobbly("FUND", 900, 4e-4, 0.01))
	fetch := fakeFetcher{"FUND": real}
	recipe := Recipe{
		ID: "FUND", Name: "a fund", Method: "canned",
		// A reconstruction that levers the fund's own path by 30 %.
		Build: func(Fetcher, time.Time) (*marketdata.Series, error) {
			return scaled("engine", real, 1.3), nil
		},
	}
	a := Audit(fetch, recipe)
	fmt.Printf("level=%s path=%s monthly=%.2f\n", a.Level, a.Path, a.MonthlyCorr)
	// Output:
	// level=bad path=ok monthly=1.00
}
