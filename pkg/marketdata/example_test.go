package marketdata_test

import (
	"fmt"
	"time"

	"github.com/bpineau/portfodor/pkg/marketdata"
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
