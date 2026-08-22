// Package marketdata fetches, caches and post-processes historical asset
// prices (daily closes) from public sources, addressed by ticker, ISIN or
// alias.
//
// Client.Fetch is the base entry point (resolution, download, disk cache);
// Client.FetchExtended is the do-what-I-mean one, adding the SIM-suffix
// history extension (bundled simulated series, long-history proxies) and
// currency conversion, i.e. the exact per-asset pipeline of the pofo CLI:
//
//	client := marketdata.NewClient(marketdata.DefaultCacheDir())
//	s, err := client.FetchExtended(ctx, "NTSGSIM", marketdata.FetchOptions{Currency: "EUR"})
//
// Every step stays independently reachable (Fetch, ReadSimdataFS,
// ExtendBack, ConvertCurrency, Trim) for callers that need to deviate.
//
// # Resolution
//
// An identifier goes through the following steps. CanonicalID applies
// steps 1-3 (identifier → canonical id); Client.Fetch runs the whole
// pipeline (and Lookup returns a catalogued asset's full metadata):
//
//  1. the built-in aliases (GOLD → XAUUSD, BHMG → GG00BQBFY362, …);
//  2. the embedded ticker → ISIN list of European ETFs and mutual funds
//     (FundISIN);
//  3. the built-in catalog of pinned resolutions (Lookup, backed by
//     datasets.Catalog), which makes common assets deterministic and
//     independent of search engines;
//  4. otherwise, a multi-source resolution: every candidate from the Yahoo
//     search ("fund" entries first), then the Financial Times, then a
//     Morningstar identifier, looked up in Morningstar's own fund screener
//     and, failing that, in Boursorama's search; the series with the
//     deepest history wins, and the resolution is cached.
//
// # Sources
//
// Yahoo Finance (adjusted closes), Stooq (fallback for plain tickers,
// major indices and major currency crosses), the ECB reference rates
// (second fallback for the currency crosses, daily since 1999), CBOE
// (fallback for ^VIX, full official history since 1990), Financial
// Times and Morningstar (NAVs of European funds). Downloads are cached on
// disk (JSON, one file per instrument); a failed refresh serves the stale
// data with a warning rather than failing.
//
// FT and Morningstar NAVs are both a PRICE return: what a distributing share
// class pays out is missing from them, silently, and no source corrects it
// (LooksDistributing warns, see the note on Series below). They also carry
// the same fund's history on their own dates: a weekly-dealing class measured
// through both stamps the same NAVs days apart, which moves the level by low
// double digits over decades without either source being wrong. Prefer one
// source per instrument, and never splice segments of both into one series. A few symbols additionally
// carry a bundled snapshot served as a last resort when every source fails
// and nothing is cached: ^VIX (daily, 1990→), the inflation indices (see
// below) and the euro crosses (the long daily ECU/DM/EUR proxy, 1971→, the
// same one that extends a live cross back in time).
//
// Eurostat serves the Harmonised Index of Consumer Prices under the
// "^HICP-<geo>" identifiers (^HICP-FR France, ^HICP-EA euro area, …): the
// monthly all-items index (2015=100) is interpolated to a smooth daily curve,
// so an inflation series behaves like any other (a chart, a CAGR that reads as
// average inflation, drawdowns that mark deflation episodes, a deflator for
// real-return work). It carries no currency. ^HICP-FR embeds a monthly snapshot
// (1955→) in the binary and is served offline-first from it: a normal run never
// downloads it. The live Eurostat API is consulted only under
// Client.RefreshInflation (set by "pofo -warmup"), which refreshes the disk
// cache a later run then prefers. Geographies without a bundled snapshot
// (^HICP-EA, …) keep the live path.
//
// "^CPI-US" is the dollar sibling: the US CPI-U all-items index (1982-84=100,
// monthly since 1913), embedded and served offline-first the same way, with the
// live FRED series fetched only on refresh.
//
// # Policy and money-market rates
//
// A last family carries the rates that SET the yields above: "^ESTR", its
// discontinued predecessor "^EONIA" and "^EURIBOR3M" on the euro side,
// "^ECB-DFR" and "^ECB-MRO" for the ECB's own policy rates, "^SOFR",
// "^FEDFUNDS" and "^FED-TARGET" on the dollar side. They
// come from the institution that publishes each one (the ECB Data Portal
// through DBnomics, the New York Fed, FRED), carry no currency, and are
// annualized percent LEVELS that may be zero or negative. Like the yield
// symbols they chart fine and never belong in a return computation; see
// rates.go for the registry and "pofo -rates" for the reader.
//
// # Dividends and raw closes
//
// Series.Dividends lists the cash distributions the source reported
// (ex-date, per-share amount in the quote currency); currency conversion
// reprices them alongside the points. The default close column is
// ADJUSTED (dividends reinvested): pairing it with Dividends would count
// income twice. Valuation-style consumers (holdings priced at market,
// dividends booked as cash) set FetchOptions.Raw to get the unadjusted
// (split-adjusted only) closes instead; raw series are cached as their own
// entries and cannot combine with the SIM extension, which is total-return
// by construction.
//
// # Intraday
//
// Client.Intraday fetches the current trading day's price path for an
// instrument. The call is live and stateless: the client performs no
// intraday caching, so the caller is responsible for throttling and
// storing results when needed. Yahoo Finance is the only intraday source;
// if the identifier does not resolve to a Yahoo symbol, Intraday returns
// ErrNotCovered (check with errors.Is). The mapping from an IntradaySeries
// to a chart is caller-side: iterate IntradaySeries.Points and copy
// IntradayPoint.Time into Dates and IntradayPoint.Close into Values on a chart.Series
// before passing it to chart.Line.
//
// # Latest quote
//
// Client.Latest returns the most recent price of an instrument as a Quote: the
// live Yahoo regular-market price (Quote.Live true) when the instrument is
// Yahoo-quoted, otherwise the last daily close (Quote.Live false), which for an
// FT or Morningstar fund is its latest NAV. Like Intraday the live path is
// stateless, so a caller valuing a portfolio repeatedly keeps its own
// short-TTL cache; the fallback inherits the whole Fetch resilience (Stooq,
// FT/Morningstar re-resolution, stale on-disk cache), so Latest still answers
// through a Yahoo outage or offline. Pair it with Client.FXRate to express
// the price in a display currency.
//
// # Data repair
//
// Every fetched daily series goes through a conservative cleaning pass
// before being cached, in this order: isolated one-session prints far from
// both their neighbours are dropped (a collapse that recovers, or a spike
// that gives it all back: on a French holiday Yahoo fills CL2.PA from the
// fund's pre-split line, ~300x above the sessions around it), leading
// provider placeholders go with them, a single persistent denomination
// break (pence vs pounds splices) is mended, one-session round trips no
// asset of the class could have made are dropped, and currency crosses lose
// isolated self-cancelling spikes (a Yahoo bad print, not a market move).
// Anything ambiguous is left untouched for the -verify-data doctor to flag
// for human review; rate symbols (^IRX, …) are exempt because their
// legitimate extremes look like artefacts.
//
// The round-trip pass is the strictest of them: it needs a full reversal
// within two sessions, six standard deviations of LOCAL volatility on each
// leg, and a leg beyond what the asset class's Band makes ordinary. All
// three, so 1987-10-19, March 2020 and October 2008 come through whole
// while a provider's isolated +21.9 %/-17.6 % on an emerging-market bond
// fund does not.
//
// # The data doctor
//
// Verify judges a series on its own: non-positive prices, suspicious moves,
// calendar gaps, flat runs, staleness, each against the series' own cadence.
// VerifyAsset adds what only the catalog record can say, and is what
// -verify-data (and `make verify-catalog`, over the whole catalog) runs:
//
//   - PLAUSIBILITY: volatility, CAGR, largest one-session move and deepest
//     drawdown against the asset class's Band (ClassBand), scaled by the
//     record's leverage. This is the check that names, in one line, what a
//     green fetch hides: an aggregate-bond fund at 20 %/yr volatility is
//     being served through a foreign-currency line.
//   - IDENTITY: the served currency against the record's currency field, the
//     served share-class name against its distribution field, and the first
//     quote against since, the share class's official launch date.
//
// Nothing here is repaired. The catalog is curated by hand, so the doctor
// names the disagreement and leaves the verdict to a human; several true
// findings are permanent, a fund whose provider serves its predecessor's
// history really does start before its own inception.
//
// # Simulated data
//
// ReadSimdata/WriteSimdata read and write the permanent simulated histories
// (pkg/datasets/simdata/) produced by the simgen package; ExtendBack splices
// those series, or a proxy (ProxySymbol), in front of the real quotes.
// The "SIM suffix" convention (DBMFSIM = DBMF with simulated extension) is
// decoded by SplitSim. Client.FetchExtended packages all of this into one
// call; the pieces stay public for custom pipelines.
//
// # Toolbox
//
//   - Align merges the trading calendars of several series (union of
//     dates, forward-filled prices);
//   - Client.Fees returns an asset's published TER (pinned catalog, disk
//     cache, otherwise FT tearsheets and justETF);
//   - UCITSFlag/GuessUCITS and LooksDistributing qualify funds;
//   - ClassBand returns an asset class's plausibility Band, the table the
//     doctor judges by and the cleaning pass is gated on;
//   - CanonicalID normalizes any accepted identifier (alias, ISIN, ticker
//     from the embedded list) to its canonical form; KnownLocal reports
//     whether it resolves without a network lookup, and LocalCatalog
//     enumerates that whole set, one entry per canonical id;
//   - IsISIN validates an ISIN, check digit included;
//   - Client.ConvertCurrency reprices a whole Series into a target currency
//     using daily Yahoo FX crosses; the earliest known rate is held flat
//     before the FX history starts;
//   - Client.Latest returns the freshest known price (a Quote) for an
//     identifier, the live Yahoo market price when available, otherwise the
//     last daily close;
//   - Client.Resolve returns a Resolution describing the instrument pofo
//     would quote for an identifier (ticker, ISIN or alias), using the catalog and
//     on-disk cache first, then the same multi-source search Fetch uses;
//     calling Resolve before Fetch lets callers inspect the resolved
//     source and symbol, and the result is cached so a subsequent Fetch
//     reuses the same work.
package marketdata
