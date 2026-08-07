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
// steps 1–3 (identifier → canonical id); Client.Fetch runs the whole
// pipeline (and Lookup returns a catalogued asset's full metadata):
//
//  1. the built-in aliases (GOLD → XAUUSD, BHMG → GG00BQBFY362, …);
//  2. the embedded ticker → ISIN list of European ETFs and mutual funds
//     (FundISIN);
//  3. the built-in catalog of pinned resolutions (Lookup, backed by
//     datasets.Catalog), which makes common assets deterministic and
//     independent of search engines;
//  4. otherwise, a multi-source resolution: every candidate from the Yahoo
//     search ("fund" entries first), then the Financial Times, then the
//     Morningstar identifier discovered via Boursorama; the series with
//     the deepest history wins, and the resolution is cached.
//
// # Sources
//
// Yahoo Finance (adjusted closes), Stooq (fallback for plain tickers,
// major indices and major currency crosses), Financial
// Times and Morningstar (NAVs of European funds). Downloads are cached on
// disk (JSON, one file per instrument); a failed refresh serves the stale
// data with a warning rather than failing.
//
// Eurostat serves the Harmonised Index of Consumer Prices under the
// "^HICP-<geo>" identifiers (^HICP-FR France, ^HICP-EA euro area, …): the
// monthly all-items index (2015=100) is interpolated to a smooth daily curve,
// so an inflation series behaves like any other (a chart, a CAGR that reads as
// average inflation, drawdowns that mark deflation episodes, a deflator for
// real-return work). It carries no currency. A monthly snapshot is embedded in
// the binary as an offline fallback (used only when the live API is unreachable
// and nothing is cached), so an inflation series is always available.
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
//   - CanonicalID normalizes any accepted identifier (alias, ISIN, ticker
//     from the embedded list) to its canonical form;
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
