// Package marketdata fetches, caches and post-processes historical asset
// prices (daily closes) from public sources, addressed by ticker, ISIN or
// alias.
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
// Yahoo Finance (adjusted closes), Stooq (ticker fallback), Financial
// Times and Morningstar (NAVs of European funds). Downloads are cached on
// disk (JSON, one file per instrument); a failed refresh serves the stale
// data with a warning rather than failing.
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
// Point.Time into Dates and Point.Close into Values on a chart.Series
// before passing it to chart.Line.
//
// # Simulated data
//
// ReadSimdata/WriteSimdata read and write the permanent simulated histories
// (pkg/datasets/simdata/) produced by the simgen package; ExtendBack splices
// those series, or a proxy (ProxySymbol), in front of the real quotes.
// The "SIM suffix" convention (DBMFSIM = DBMF with simulated extension) is
// decoded by SplitSim.
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
//   - Client.Resolve returns a Resolution describing the instrument pofo
//     would quote for an identifier (ticker, ISIN or alias), using the catalog and
//     on-disk cache first, then the same multi-source search Fetch uses;
//     calling Resolve before Fetch lets callers inspect the resolved
//     source and symbol, and the result is cached so a subsequent Fetch
//     reuses the same work.
package marketdata
