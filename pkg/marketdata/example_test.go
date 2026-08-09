package marketdata_test

import (
	"context"
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

	series, err := client.Fetch(context.Background(), "IWDA", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: %d quotes since %s\n",
		series.Name, len(series.Points), series.First().Date.Format("2006-01-02"))
}

// Example_fetchExtended shows the do-what-I-mean entry point: the SIM suffix
// extends the fund's history with the bundled simulated series (real quotes
// keeping priority), and the result is converted to euros. This is the exact
// per-asset pipeline of the pofo CLI, in one call.
// (Not run: requires the network.)
func Example_fetchExtended() {
	client := marketdata.NewClient(marketdata.DefaultCacheDir())

	series, err := client.FetchExtended(context.Background(), "NTSGSIM", marketdata.FetchOptions{Currency: "EUR"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s: since %s (simulated before %s)\n", series.Name,
		series.First().Date.Format("2006-01-02"),
		series.SimulatedBefore.Format("2006-01-02"))
}

// ExampleClient_Latest shows the one-call answer to "what is this asset worth
// right now": the live Yahoo market price when the instrument is Yahoo-quoted
// (Quote.Live true), otherwise its last daily close or fund NAV. FXRate turns
// the quote currency into the display one.
// (Not run: requires the network.)
func ExampleClient_Latest() {
	client := marketdata.NewClient(marketdata.DefaultCacheDir())
	ctx := context.Background()

	q, err := client.Latest(ctx, "VWCE")
	if err != nil {
		panic(err)
	}
	rate, err := client.FXRate(ctx, q.Currency, "EUR", q.Time)
	if err != nil {
		panic(err)
	}
	freshness := "close of"
	if q.Live {
		freshness = "live at"
	}
	shares := 12.0
	fmt.Printf("%.2f EUR (%s %s)\n", shares*q.Price*rate, freshness, q.Time.Format("2006-01-02 15:04"))
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
	// [warn] 2024-01-04: move of +243.1 % in one observation, beyond the 25.0 % this series can make
	// [warn] 2024-01-05: move of -70.6 % in one observation, beyond the 25.0 % this series can make
}

// Lookup resolves a ticker, alias or ISIN to the asset's full catalog
// metadata in one call.
func ExampleLookup() {
	a, ok := marketdata.Lookup("IWDA")
	fmt.Println(ok, a.ID, a.AssetClass, a.Fees)
	// Output: true IE00B4L5Y983 equity 0.2
}

// LocalCatalog enumerates the identifier set that resolves without any
// network lookup: one entry per canonical id, sorted, with its alternates.
func ExampleLocalCatalog() {
	cat := marketdata.LocalCatalog()
	fmt.Println(len(cat) > 0)
	// Output: true
}

// ClassBand exposes the plausibility bounds the data doctor judges an asset
// against: what its class can do over a whole history, and what it cannot.
// Scale widens them by a fund's notional leverage.
func ExampleClassBand() {
	b, _ := marketdata.ClassBand("aggregate-bond")
	fmt.Printf("plain      volatility up to %.0f %%/yr, worst session %.0f %%\n", b.VolHi*100, b.Move*100)
	stacked := b.Scale(1.5) // a 90/60 efficient-core sleeve
	fmt.Printf("at 1.5x    volatility up to %.0f %%/yr, worst session %.0f %%\n", stacked.VolHi*100, stacked.Move*100)
	// Output:
	// plain      volatility up to 12 %/yr, worst session 8 %
	// at 1.5x    volatility up to 18 %/yr, worst session 12 %
}

// VerifyAsset is the doctor's full pass on a catalogued identifier: the series
// hygiene Verify judges from the quotes alone, plus plausibility against the
// asset class's band and identity against the catalog record. Here a series
// pinned to an equity ETF is far too quiet to be one.
func ExampleVerifyAsset() {
	s := &marketdata.Series{Symbol: "VOO", Currency: "USD"}
	start := time.Date(2010, 9, 10, 0, 0, 0, 0, time.UTC) // days after VOO's own inception
	for i := 0; i < 500; i++ {
		s.Points = append(s.Points, marketdata.Point{
			Date:  start.AddDate(0, 0, i),
			Close: 100 * (1 + 0.0002*float64(i%7)),
		})
	}
	for _, issue := range marketdata.VerifyAsset("VOO", s, start.AddDate(0, 0, 501)) {
		fmt.Println(issue.Message)
	}
	// Output:
	// volatility 0.9 %/yr is outside the equity band [6.0, 42.0], wrong quote line?
}
