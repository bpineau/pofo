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

// CapWeighted builds the same kind of index without rebalancing it, the way a
// cap-weighted one behaves: the weights are the published split on the anchor
// date and drift with the legs' own returns on either side of it. Held at
// today's split all along, the same legs earn more: the leg that outran the
// other gets its end weight over a past in which it weighed less.
func ExampleCapWeighted() {
	fetch := fakeFetcher{
		"US":    mkSeries("US", 300, 0.0008),
		"WORLD": mkSeries("WORLD", 300, 0.0002),
	}
	fr, err := BuildFrame(fetch, []string{"US", "WORLD"}, day(0))
	if err != nil {
		panic(err)
	}
	legs := []Leg{{ID: "US", Weight: 0.40}, {ID: "WORLD", Weight: 0.60}}
	anchor := fr.Dates[len(fr.Dates)-1] // the split was published on the last date
	drifting, err := CapWeighted(fr, legs, anchor, 0)
	if err != nil {
		panic(err)
	}
	fixed, err := Composite(fr, legs, "", 0)
	if err != nil {
		panic(err)
	}
	last := len(fr.Dates) - 1
	fmt.Printf("cap-weighted %.2f, fixed at the anchor's split %.2f\n", drifting[last], fixed[last])
	// Output:
	// cap-weighted 113.62, fixed at the anchor's split 114.06
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

// TreasuryZeroTR prices a constant-maturity zero-coupon Treasury (a STRIPS
// ladder) off a long yield series, which is how the 25+ year STRIPS
// reconstruction reaches 1953. A par coupon bond of the same maturity is much
// shorter once yields are high, which is the whole reason the two engines are
// separate: here a one-point fall in a 12 % yield.
func ExampleTreasuryZeroTR() {
	yields := &marketdata.Series{Points: []marketdata.Point{
		{Date: time.Date(1981, 9, 1, 0, 0, 0, 0, time.UTC), Close: 12},
		{Date: time.Date(1981, 10, 1, 0, 0, 0, 0, time.UTC), Close: 11},
	}}
	strip := TreasuryZeroTR("27y STRIP", yields, 27, 0.0015)
	par := TreasuryTR("22y par bond", yields, 22, 0.0015)
	fmt.Printf("strip %+.1f%%, par bond %+.1f%%\n",
		strip.Last().Close-100, par.Last().Close-100)
	// Output:
	// strip +30.2%, par bond +9.1%
}
