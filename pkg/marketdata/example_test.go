package marketdata_test

import (
	"context"
	"errors"
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

// PlausibleID judges an identifier on its shape alone, the first gate a
// caller applies to an identifier it did not vet itself: a checksummed ISIN
// or a plausible exchange ticker, nothing else.
func ExamplePlausibleID() {
	fmt.Println(marketdata.PlausibleID("IWDA.AS"))
	fmt.Println(marketdata.PlausibleID("IE00B4L5Y983"))
	fmt.Println(marketdata.PlausibleID("IE00B4L5Y984")) // wrong check digit
	fmt.Println(marketdata.PlausibleID("MSCI WORLD"))
	// Output:
	// true
	// true
	// false
	// false
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

// Example_fetchFailure shows how to tell the two failures of a fetch apart,
// which any caller answering someone else's identifier has to do: nothing
// quotes it (permanent, and the error names it) or a source did not answer
// (transient, so the same request may work later).
// (Not run: requires the network.)
func Example_fetchFailure() {
	client := marketdata.NewClient(marketdata.DefaultCacheDir())

	_, err := client.Fetch(context.Background(), "ZQXWVT", time.Time{})
	var unknown *marketdata.UnknownIdentifierError
	switch {
	case err == nil:
		fmt.Println("quoted after all")
	case errors.As(err, &unknown):
		fmt.Printf("no source quotes %s\n", unknown.ID) // errors.Is(err, marketdata.ErrUnknownIdentifier)
	default:
		fmt.Printf("try again later: %v\n", err)
	}
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

// ExampleClient_LatestBatchExtended values US lines before the opening bell:
// with the extended-hours opt-in, a pre-market print beats yesterday's close,
// and Quote.Session says which session the price belongs to so the display can
// label it rather than pass it off as a close.
// (Not run: requires the network.)
func ExampleClient_LatestBatchExtended() {
	client := marketdata.NewClient(marketdata.DefaultCacheDir())

	for id, q := range client.LatestBatchExtended(context.Background(), []string{"DDOG", "VT", "IWDA.AS"}) {
		label := map[string]string{"pre": "pre-market", "post": "after hours", "regular": "regular session"}[q.Session]
		if label == "" {
			label = "last close" // a fund NAV or a daily close names no session
		}
		fmt.Printf("%s %.2f %s (%s, %s)\n", id, q.Price, q.Currency, label, q.Time.Format("15:04"))
	}
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

// SampleAt reads one series onto a calendar it does not own: an exogenous
// level (here a financing rate) joined to the assets' trading days without
// adding any of its own, and held flat rather than zero before its history.
func ExampleSampleAt() {
	day := func(i int) time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	rate := &marketdata.Series{Symbol: "^IRX", Points: []marketdata.Point{
		{Date: day(2), Close: 5.25}, {Date: day(4), Close: 5.00},
	}}
	levels, before := marketdata.SampleAt(rate, []time.Time{day(0), day(2), day(3), day(5)})
	fmt.Println(levels, before.Format("2006-01-02"))
	// Output:
	// [5.25 5.25 5.25 5] 2024-01-03
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

// Trim restricts a series to a window; a zero bound is open on that side, and
// dividends are clipped along with the points.
func ExampleTrim() {
	day := func(i int) time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	s := &marketdata.Series{Symbol: "DEMO", Currency: "USD", Points: []marketdata.Point{
		{Date: day(0), Close: 100}, {Date: day(1), Close: 101},
		{Date: day(2), Close: 102}, {Date: day(3), Close: 103},
	}, Dividends: []marketdata.Dividend{
		{Date: day(0), Amount: 0.5}, {Date: day(2), Amount: 0.6},
	}}
	window := marketdata.Trim(s, day(1), day(2))
	fmt.Println(len(window.Points), window.First().Close, window.Last().Close, len(window.Dividends))
	open := marketdata.Trim(s, day(2), time.Time{})
	fmt.Println(len(open.Points), open.First().Close)
	// Output:
	// 2 101 102 1
	// 2 102
}

// ExtendBack splices a long-history proxy in front of an asset's own quotes,
// rescaled so the two agree on the day they meet. SimulatedBefore marks the
// frontier, so a reader always knows where the real data starts.
func ExampleExtendBack() {
	day := func(y int) time.Time { return time.Date(y, 1, 3, 0, 0, 0, 0, time.UTC) }
	asset := &marketdata.Series{Symbol: "VOO", Points: []marketdata.Point{
		{Date: day(2010), Close: 100}, {Date: day(2011), Close: 110},
	}}
	proxy := &marketdata.Series{Symbol: "^GSPC", Points: []marketdata.Point{
		{Date: day(2000), Close: 25}, {Date: day(2005), Close: 40}, {Date: day(2010), Close: 50},
	}}
	fmt.Println(marketdata.ExtendBack(asset, proxy))
	for _, p := range asset.Points {
		fmt.Println(p.Date.Format("2006-01-02"), p.Close)
	}
	fmt.Println(asset.ProxySymbol, asset.SimulatedBefore.Format("2006-01-02"))
	// Output:
	// true
	// 2000-01-03 50
	// 2005-01-03 80
	// 2010-01-03 100
	// 2011-01-03 110
	// ^GSPC 2010-01-03
}

// LooksDistributing spots a distributing share class in a fund's name. The
// warning matters because a distributing NAV series is a PRICE return: the
// income it pays out is missing from every statistic computed on it.
func ExampleLooksDistributing() {
	for _, name := range []string{
		"iShares Core MSCI World UCITS ETF USD (Acc)",
		"iShares $ Treasury Bond 20+yr UCITS ETF (Dist)",
	} {
		fmt.Println(marketdata.LooksDistributing(name), name)
	}
	// Output:
	// false iShares Core MSCI World UCITS ETF USD (Acc)
	// true iShares $ Treasury Bond 20+yr UCITS ETF (Dist)
}

// A rate symbol is an annualized percent LEVEL, not a price: read the
// registry to offer them, and never feed one to a return computation.
// A London listing is quoted in pence and labelled "GBp", a code that folds
// into "GBP" in any case-insensitive comparison. Fetch rescales such a series
// on the way in, so nothing this package serves ever carries a sub-unit;
// IsMinorUnit is there for a caller checking records of its own.
func ExampleIsMinorUnit() {
	for _, code := range []string{"GBP", "GBp", "GBX", "ZAc", "EUR"} {
		fmt.Println(code, marketdata.IsMinorUnit(code))
	}
	// Output:
	// GBP false
	// GBp true
	// GBX true
	// ZAc true
	// EUR false
}

func ExampleRateName() {
	fmt.Println(marketdata.RateName("^ESTR"))
	fmt.Println(marketdata.RateName("^NOSUCHRATE") == "")
	// Output:
	// Euro short-term rate (ESTR, overnight)
	// true
}

// WithoutEstimates removes a nowcast tail: the days a fund's proxy stood in
// for it, after its last published value. Every consumer that stores or
// validates data reads through it, so an estimate is never shipped.
func ExampleSeries_WithoutEstimates() {
	day := func(i int) time.Time { return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	s := &marketdata.Series{Symbol: "ERESMONDEM", EstimatedFrom: day(2), EstimateProxy: "URTH",
		Points: []marketdata.Point{
			{Date: day(0), Close: 50}, {Date: day(1), Close: 51},
			{Date: day(2), Close: 51.4}, {Date: day(3), Close: 51.8},
		}}
	published := s.WithoutEstimates()
	fmt.Println(len(s.Points), s.EstimateProxy)
	fmt.Println(len(published.Points), published.Last().Date.Format("2006-01-02"), published.EstimatedFrom.IsZero())
	// Output:
	// 4 URTH
	// 2 2026-09-02 true
}

// NewSeries wraps a consumer's own data in a Series, which then hands the
// parallel slices every pkg/metrics function takes: Dates and Values, and
// Returns as fractions.
func ExampleNewSeries() {
	dates := []time.Time{
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
	}
	s, err := marketdata.NewSeries("MY-BOOK", dates, []float64{1000, 1010, 999.9})
	if err != nil {
		panic(err)
	}
	fmt.Println(s.Len(), s.Values())
	r := s.Returns()
	fmt.Printf("%+.2f%% %+.2f%%\n", 100*r[0], 100*r[1])

	_, err = marketdata.NewSeries("MY-BOOK", []time.Time{dates[1], dates[0]}, []float64{1, 2})
	fmt.Println(err)
	// Output:
	// 3 [1000 1010 999.9]
	// +1.00% -1.00%
	// marketdata: NewSeries MY-BOOK: date 1 (2024-01-02) does not follow date 0 (2024-01-03)
}

// Rebase scales a series so it starts at a chosen level, 100 here, leaving
// every return, the metadata and the dividends (cash) as they were.
func ExampleSeries_Rebase() {
	day := func(i int) time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	s := &marketdata.Series{Symbol: "DEMO", Currency: "EUR", Points: []marketdata.Point{
		{Date: day(1), Close: 40}, {Date: day(2), Close: 42}, {Date: day(3), Close: 41},
	}, Dividends: []marketdata.Dividend{{Date: day(2), Amount: 0.3}}}
	r := s.Rebase(100)
	fmt.Println(r.Values(), r.Currency, r.Dividends[0].Amount)
	// Output:
	// [100 105 102.5] EUR 0.3
}

// Resample keeps the last TRADING close of each calendar period, dated on
// that close: March 2024 ends on a weekend after Good Friday, so its month-end
// is Thursday the 28th. The unfinished last month is kept, on its own date.
func ExampleSeries_Resample() {
	d := func(m time.Month, day int) time.Time { return time.Date(2024, m, day, 0, 0, 0, 0, time.UTC) }
	s, _ := marketdata.NewSeries("DEMO",
		[]time.Time{d(2, 28), d(2, 29), d(3, 27), d(3, 28), d(4, 1), d(4, 2)},
		[]float64{100, 101, 104, 105, 103, 106})
	for _, p := range s.Resample(marketdata.Monthly).Points {
		fmt.Println(p.Date.Format("2006-01-02"), p.Close)
	}
	fmt.Println(s.Resample(marketdata.Quarterly).Len())
	// Output:
	// 2024-02-29 101
	// 2024-03-28 105
	// 2024-04-02 106
	// 2
}

// CommonWindow is the stretch on which every series quotes: from the latest
// first quote to the earliest last one.
func ExampleCommonWindow() {
	d := func(day int) time.Time { return time.Date(2024, 1, day, 0, 0, 0, 0, time.UTC) }
	old, _ := marketdata.NewSeries("OLD", []time.Time{d(2), d(3), d(4), d(5)}, []float64{1, 2, 3, 4})
	young, _ := marketdata.NewSeries("YOUNG", []time.Time{d(4), d(5), d(8)}, []float64{1, 2, 3})
	start, end, ok := marketdata.CommonWindow(old, young)
	fmt.Println(start.Format("2006-01-02"), end.Format("2006-01-02"), ok)
	// Output:
	// 2024-01-04 2024-01-05 true
}

// AlignSeries puts several series on one calendar, starting by default where
// all of them quote, and refuses a start that would forward-fill zeros where
// Align would have done it in silence. Returns hands the per-asset returns a
// correlation or a covariance takes.
func ExampleAlignSeries() {
	d := func(day int) time.Time { return time.Date(2024, 1, day, 0, 0, 0, 0, time.UTC) }
	old, _ := marketdata.NewSeries("OLD", []time.Time{d(2), d(3), d(4), d(5)}, []float64{100, 101, 102, 103})
	young, _ := marketdata.NewSeries("YOUNG", []time.Time{d(3), d(5)}, []float64{50, 55})

	a, err := marketdata.AlignSeries([]*marketdata.Series{old, young}, time.Time{}, time.Time{})
	if err != nil {
		panic(err)
	}
	fmt.Println(a.IDs, len(a.Dates), a.Dates[0].Format("2006-01-02"))
	fmt.Println(a.Levels[1], len(a.Returns()[1]))

	_, err = marketdata.AlignSeries([]*marketdata.Series{old, young}, d(2), time.Time{})
	fmt.Println(err)
	// Output:
	// [OLD YOUNG] 3 2024-01-03
	// [50 50 55] 2
	// marketdata: AlignSeries: YOUNG starts 2024-01-03, after the window's start 2024-01-02
}
