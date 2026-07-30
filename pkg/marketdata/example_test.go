package marketdata_test

import (
	"fmt"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// CanonicalID follows alias chains to the canonical identifier.
func ExampleCanonicalID() {
	fmt.Println(marketdata.CanonicalID("gold"))
	fmt.Println(marketdata.CanonicalID("AMUNDI-VOLATILITY"))
	fmt.Println(marketdata.CanonicalID("VOO"))
	// Output:
	// XAUUSD
	// LU0319687124
	// VOO
}

// FundISIN translates European ETF and mutual fund tickers to ISINs using
// the embedded correspondence list.
func ExampleFundISIN() {
	isin, ok := marketdata.FundISIN("IWDA")
	fmt.Println(isin, ok)
	// Output:
	// IE00B4L5Y983 true
}

// Example_fetch shows typical client usage: resolution
// (alias → ISIN → source), with transparent downloading and disk caching.
// (Not run: requires the network.)
func Example_fetch() {
	client := marketdata.NewClient("data")
	client.Logf = func(format string, args ...any) { /* optional logging */ }

	series, err := client.Fetch("IWDA", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s — %d quotes since %s\n",
		series.Name, len(series.Points), series.First().Date.Format("2006-01-02"))
}

// Align merges trading calendars: the union of dates from start on, with
// each series' level forward-filled across its own non-trading days.
func ExampleAlign() {
	day := func(i int) time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	a := &marketdata.Series{Symbol: "A", Points: []marketdata.Point{
		{Date: day(0), Close: 10}, {Date: day(1), Close: 11}, {Date: day(2), Close: 12},
	}}
	b := &marketdata.Series{Symbol: "B", Points: []marketdata.Point{
		{Date: day(0), Close: 100}, {Date: day(2), Close: 102}, // no quote on day 1
	}}
	dates, levels := marketdata.Align([]*marketdata.Series{a, b}, day(0), time.Time{})
	fmt.Println(len(dates), levels[0], levels[1])
	// Output:
	// 3 [10 11 12] [100 100 102]
}

// Verify is the data doctor: it flags bad points, gaps, flat stretches and
// staleness so suspect series are reviewed instead of silently skewing a
// simulation.
func ExampleVerify() {
	s := &marketdata.Series{Symbol: "DEMO"}
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	closes := []float64{100, 101, 102, 350, 103, 104} // one obviously bad point
	for i, c := range closes {
		s.Points = append(s.Points, marketdata.Point{Date: start.AddDate(0, 0, i), Close: c})
	}
	for _, issue := range marketdata.Verify(s, start.AddDate(0, 0, 7)) {
		fmt.Println(issue)
	}
	// Output:
	// [warn] 2024-01-04: daily move of +243.1 % — missed split or bad point?
	// [warn] 2024-01-05: daily move of -70.6 % — missed split or bad point?
}

// Lookup resolves a ticker, alias or ISIN to the asset's full catalog
// metadata in one call.
func ExampleLookup() {
	a, ok := marketdata.Lookup("IWDA")
	fmt.Println(ok, a.ID, a.AssetClass, a.Fees)
	// Output: true IE00B4L5Y983 equity 0.2
}
