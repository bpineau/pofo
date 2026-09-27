// Package marketdata is where every series comes from and how it is shaped
// for the math: it fetches, caches and cleans historical prices (daily
// closes) from public sources, addressed by ticker, ISIN or alias, reads the
// series the binary bundles, and cuts several series into the tables a
// statistic across series reads.
//
// # Start here
//
//   - [Bundled] reads an embedded series (an index, a yield, a cash rate, or
//     a catalog fund's SIM history under its SIM name) with no network and
//     no client; [BundledIDs] lists them.
//   - [NewClient] and [Client.FetchExtended] fetch anything quoted: the CLI's
//     per-asset pipeline (resolution, disk cache, SIM backcast, currency
//     conversion). [Client.Offline] keeps it off the network.
//   - [Client.Load] is the explorer's door: a CSV path, else the bundle, else
//     the client, one call per identifier.
//   - [Series] is what every door returns: [Series.Stats] scores it,
//     [Series.Returns], [Series.Resample], [Series.Change] and
//     [Series.LessFee] reshape it, [NewSeries] wraps a consumer's own data.
//   - [NewPanel] cuts several series into a returns table on the periods they
//     all share: [Panel.Col], [Panel.Mix], [Panel.Pick], [Panel.Track].
//     [AlignSeries] puts their levels on one calendar instead.
//   - [ReadCSV], [ReadLongCSV] and [WriteCSV] move series in and out as files.
//
// The CLI's own fetch, a fund with its backcast in front, in euros:
//
//	client := marketdata.NewClient(marketdata.DefaultCacheDir())
//	s, err := client.FetchExtended(ctx, "NTSGSIM", marketdata.FetchOptions{Currency: "EUR"})
//
// Every step stays independently reachable ([Client.Fetch], [ReadSimdataFS],
// [Client.ConvertCurrency], [Trim]) for callers that need to deviate; the
// steps that exist for the data generators are listed apart, under "Generator
// plumbing".
//
// Units: closes are ADJUSTED total-return levels in the major unit of
// [Series].Currency, every [Point].Date is a session date at 00:00 UTC,
// returns are FRACTIONS, [Series.LessFee] takes a FRACTION per year while
// [Client.Fees] answers in PERCENT per year, and the rate symbols (^IRX,
// ^ESTR...) are annualized PERCENT LEVELS, never prices.
//
// # Loading series
//
// Four doors lead to a Series, and only the first may use the network:
//
//   - Client.FetchExtended (or Fetch) for anything quoted, downloaded and
//     cached on disk under DefaultCacheDir, the directory the CLI's -data
//     defaults to. Client.Offline keeps it off the network: the cache then
//     answers whatever its age and depth, the bundled data behind it (a
//     catalog index, the backcast of a SIM identifier, the snapshots listed
//     under "Sources"), and anything else fails with an error wrapping
//     ErrOffline. The cache's JSON files are private to this package and
//     change without notice: read them through an Offline client, never
//     directly.
//   - Bundled for what the binary embeds, with no Client at all: a catalog
//     asset's SIM history (pkg/datasets/simdata, Source "simdata": the
//     reconstruction with the real quotes grafted on, as of the last
//     refresh), answering to the SIM form of a quoted fund's identifier
//     ("IWDASIM", never the bare "IWDA", which means real quotes only) and
//     to the bare id of a catalog index ("SP500"); or a reference series
//     (pkg/datasets/refdata, Source "refdata": indices, yields, cash rates,
//     NAV snapshots). BundledIDs lists them.
//   - ReadCSV for any "date,value" file, the layout of the bundled ones: a
//     file of one's own, or a bundled file as of an older commit piped out
//     of "git show". It honours the headers the bundled files carry
//     ("# junctions:" above all) and its errors name the line.
//   - ReadLongCSV for several series in one "id,date,value" file.
//
// Client.Load chains three of them for a program that explores data: a path
// is read as a file, a bare identifier the module bundles (a reference
// series, a catalog index) comes from the bundle, and anything else goes
// through FetchExtended, so a fund's bare identifier is its real quotes and
// its SIM identifier those quotes with the reconstruction in front; the
// window and the currency conversion apply to all three.
// It is the call every script of examples/code makes.
//
// WriteCSV is the way out: the long layout, each series' metadata as "#"
// comments any CSV reader can skip, values that parse back exactly. "pofo
// -dump" is the same on the command line.
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
// An index symbol ("^GSPC", "^BCOM") never reaches step 4 as a name: its
// fallback keeps only listings of the symbol itself (as
// FetchOptions.ExactOnly does for any ticker), and a cached resolution to
// another instrument is ignored, because the full-text searches match a fund
// whose name merely carries the index's short name. An index symbol whose
// feed was withdrawn is served from the bundled reference that replaces it
// ("^BCOM" from refdata BCOM-ER-USD, Yahoo having dropped the Bloomberg
// Commodity family in 2026-09), without touching the network.
//
// # When a fetch finds nothing
//
// Two failures look alike from the outside, and a caller answering a stranger
// (the web app above all) must tell them apart:
//
//   - no source quotes the identifier: every source that answered reported it
//     holds nothing under that name. Fetch returns an *UnknownIdentifierError,
//     which names the identifier and satisfies errors.Is(err,
//     ErrUnknownIdentifier). It is permanent: a typo, an invented ticker, a
//     line nothing carries.
//   - a source did not answer: a network error, a rate limit, an HTTP 5xx, an
//     unreadable payload. Nothing is marked, since the failure says nothing
//     about the identifier, and the same request may work later.
//
// The distinction is drawn the safe way round: a failure counts as evidence
// about the identifier only when the source that produced it said so, so an
// unclassified one reads as the second case rather than blaming the caller's
// identifier. ErrWrongCurrency is a third, narrower answer: the instrument
// exists, but no source quotes it natively in the requested currency
// (FetchOptions.Currency with NoConvert).
//
// # Sources
//
// Yahoo Finance (adjusted closes), Stooq (fallback for plain tickers,
// major indices and major currency crosses), the ECB reference rates
// (second fallback for the currency crosses, daily since 1999), CBOE
// (fallback for ^VIX, full official history since 1990), Financial
// Times and Morningstar (NAVs of European funds), and the airfund.io delivery
// API (the official daily NAV behind a French employee-savings fund's page,
// catalog source "airfund": such a fund has no ISIN and no listing, and its
// bundled refdata NAV snapshot answers offline). Downloads are cached on
// disk (one file per instrument, in a format private to this package); a
// failed refresh serves the stale data with a warning rather than failing. Each file records the format it was
// written under, so a fix that changes what a correct file holds can distrust
// its own past output without invalidating a whole cache (see cacheFormat).
//
// FT and Morningstar NAVs are both a PRICE return: what a distributing share
// class pays out is missing from them, silently, and no source corrects it
// (LooksDistributing warns, see the note on Series below). They also carry
// the same fund's history on their own dates: a weekly-dealing class measured
// through both stamps the same NAVs days apart, which moves the level by low
// double digits over decades without either source being wrong. Prefer one
// source per instrument, and never splice segments of both into one series.
// The cache enforces that last rule: a non-Yahoo source's history is keyed by
// the SOURCE and the identifier, so an ISIN re-resolved from one to the other
// refetches instead of reading its predecessor's file back, which the
// stale-cache fallback would otherwise have done silently during an outage.
//
// A few symbols additionally carry a bundled snapshot, served as a last
// resort when every source fails and nothing is cached: ^VIX (daily, 1990→),
// the inflation indices (see below) and the euro crosses (the long daily
// ECU/DM/EUR proxy, 1971→, the same one that extends a live cross back in
// time).
//
// # Quote units
//
// Series.Currency and Quote.Currency are always an ISO currency, and the
// numbers next to them are always in that currency's MAJOR unit. Providers
// are not: a London listing comes back from Yahoo in pence labelled "GBp" and
// from the Financial Times labelled "GBX", Johannesburg in cents ("ZAc"), the
// Chicago grain contracts in US cents ("USX"). Every such series is rescaled
// where it enters the package (units.go), dividends included, so a consumer
// never has to know which venue it came from.
//
// This is not a convenience. "GBp" and "GBP" differ by case alone, so a pence
// price satisfies every currency gate a caller can set - FetchOptions.Currency
// with NoConvert, QuoteOptions.Currency - and is then booked as pounds: a
// hundredfold valuation error that no plausibility band can see, since a
// change of unit leaves every return untouched. IsMinorUnit names a sub-unit
// code for a caller validating its own records, and the data doctor reports
// one reaching it.
//
// # Trading days
//
// Every Point.Date is the calendar day the session ran on, at 00:00 UTC. A
// provider's daily bar carries an INSTANT of that session rather than a date,
// so the day is read in the venue's own time zone: the ASX opens at 10:00 in
// Sydney, which is 23:00 UTC of the day before while Australia is on summer
// time, and reading it in UTC dated half of every Australian history one day
// early, Monday's session on a Sunday. The correction only ever moves a date
// forward, since a UTC reading is never late, only early.
//
// One family is deliberately left on the UTC calendar: a currency cross, which
// has no exchange and no trading day (see venueTimezone, which also records
// the separate dating anomaly Yahoo's crosses carry in the European summer).
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
// (^HICP-EA, …) keep the live path. The index is read from Eurostat's
// prc_hicp_minr dataset (ECOICOP 2, on its 2015=100 unit), the one Eurostat
// kept updating when it rebased the HICP in 2026 and froze the former
// prc_hicp_midx at 2025-12: a deflator that stops quietly reads every later
// month as zero inflation, so a live series trailing the calendar by more than
// three months is served with a warning.
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
// # Definition junctions
//
// Series.Junctions are the dates on which a series' publisher started
// measuring something ELSE: the levels on both sides are real, the step
// between them is not a move, and a consumer that reads a return from it
// fabricates one. They are read from a simdata file's optional
// "# junctions:" header, so the declaration travels with the data rather
// than living in whichever code happened to notice. The bundled
// TREASURY-LONG-YIELD carries one, 1973-01-04, where H.15's 20-year
// constant maturity stepped 0.71 pt overnight while the 10-year point sat
// still; simgen's constant-maturity engines skip any period spanning one.
//
// A file that STOPS by design declares it the same way, in an optional
// "# ends:" header (a date, then the reason in free text), read into
// Series.Ends: its source was discontinued (WTI-ER-USD after EIA stopped
// publishing its contract series) or it is trimmed where better data takes
// over (the daily shape series). Verify then reads an old last point as the
// whole record rather than a stale feed, and reports the file instead when its
// last point is not the declared date.
//
// # Intraday
//
// Client.Intraday fetches the current trading day's price path for an
// instrument. The call is live and stateless: the client performs no
// intraday caching, so the caller is responsible for throttling and
// storing results when needed. Yahoo Finance is the only intraday source;
// if the identifier does not resolve to a Yahoo symbol, Intraday returns
// ErrNotCovered (check with errors.Is), unless the catalog names a
// nowcast_proxy for it: a fund priced once a day and published with a lag
// (an FCPE) then gets an ESTIMATED path, its last daily value scaled by the
// proxy's intraday move converted tick by tick (IntradaySeries.Estimate,
// see nowcast.go). The same proxy extends the fund's daily series past its
// last NAV (Series.EstimatedFrom marks the tail; WithoutEstimates strips
// it, and nothing cached or shipped ever holds it). Both estimates start
// from the proxy print the last NAV was struck on: its close of that day,
// or its OPEN when the record says nowcast_anchor "open" (a fund whose
// valuation rules price its holding at the opening of the valuation day),
// with a silent fall back on the close when that day has no opening price.
// The mapping from an IntradaySeries
// to a chart is caller-side: iterate IntradaySeries.Points and copy
// IntradayPoint.Time into Dates and IntradayPoint.Close into Values on a chart.Series
// before passing it to chart.Line.
//
// # Latest quote
//
// Client.Latest returns the most recent price of an instrument as a Quote: the
// live Yahoo regular-market price (Quote.Live true) when the instrument is
// Yahoo-quoted, otherwise the last daily close (Quote.Live false), which for an
// FT or Morningstar fund is its latest NAV; a fund with a nowcast_proxy quotes
// the last tick of its estimated intraday path (Quote.Live true, Source
// "nowcast"), or its last published NAV when the proxy is unreachable. Like Intraday the live path is
// stateless, so a caller valuing a portfolio repeatedly keeps its own
// short-TTL cache; the fallback inherits the whole Fetch resilience (Stooq,
// FT/Morningstar re-resolution, stale on-disk cache), so Latest still answers
// through a Yahoo outage or offline. Pair it with Client.FXRate to express
// the price in a display currency.
//
// Client.LatestBatch quotes many identifiers in one call, and Client.LatestAny
// the first of several identifiers of one instrument that answers.
//
// # Trading sessions and extended hours
//
// Quote.Session names the session the price was struck in, and Quote.Time is
// always the instant of that print:
//
//   - "regular": Yahoo's regularMarketPrice. Time is an intraday instant while
//     the market is open, and the closing instant once it has closed - a closed
//     market still quotes, which is why Live alone never means "open".
//   - "pre" / "post": a pre-market or after-hours print, only ever returned
//     under the extended-hours opt-in below. Time is the instant of that print.
//   - "": the source names no session. A daily close (Time is that close's
//     date at 00:00 UTC), an FT or Morningstar NAV, a nowcast estimate.
//
// The default quote paths report "regular" or "", never an off-hours price:
// existing callers see the same prices they always did. Ask for extended hours
// with QuoteOptions.ExtendedHours (Client.LatestAny) or Client.LatestBatchExtended,
// and a pre-market or after-hours print then wins whenever it is STRICTLY more
// recent than the regular session's last price. That recency test is the whole
// rule: it needs no session calendar, it ignores the stale pre-market field
// Yahoo keeps serving during the regular session, and at 03:00 in New York it
// correctly reports last night's after-hours print, labelled "post", instead of
// the 16:00 close. Only US venues run these sessions; a European line, a fund
// NAV and a nowcast answer exactly as they would without the opt-in.
//
// Extended-hours quotes come from Yahoo's v7 quote API, the only endpoint
// carrying preMarketPrice/postMarketPrice (the chart meta the single-symbol
// spot path reads has neither, nor marketState). It needs a cookie+crumb pair,
// so it can fail where the plain spot call succeeds: the extended leg then
// falls back to the regular quote rather than lose it. Two caveats for the
// consumer: an off-hours print is thin and wide-spread, a price to SHOW rather
// than to book a valuation on, and Client.Intraday is deliberately left alone -
// its path stops at the regular session's bounds.
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
// # Data guards
//
// Two rules read a series as DATA, because what they hunt is invisible to a
// statistic, and each exists once. FindGaps returns the steps longer than
// the series' own local pace allows (one and a half of its steps, never
// under fourteen days): a monthly file that skips a month, a daily one
// silent for three weeks. FindSpikes returns the one-session round trips no
// instrument makes (opposite legs of 2 % and more, each beyond six local
// standard deviations, cancelling to within a third of the smaller leg).
// The data doctor reports both, and the bundle's golden guards run both
// over every embedded file. The round trip is one mechanism, SpikeRule,
// whose bars a caller sets: the fetch-time cleaner above Drops what its
// stricter rule proves, pkg/simgen's shape despiker what a floorless one
// does, and FindSpikes reports what shipped data must never carry.
//
// # The data doctor
//
// Verify judges a series on its own: non-positive prices, suspicious moves,
// calendar gaps (FindGaps), round trips (FindSpikes), flat runs, staleness,
// each against the series' own cadence (a declared Series.Ends replaces the
// staleness check).
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
// ReadSimdata and ReadSimdataFS read the permanent simulated histories
// (pkg/datasets/simdata/) produced by the simgen package, and the long
// reference series of pkg/datasets/refdata/ in the same format, through
// ReadCSV's parser; Bundled is the one-call form over the embedded copies.
// The "SIM
// suffix" convention (DBMFSIM = DBMF with simulated extension) is decoded by
// SplitSim. Client.FetchExtended packages the whole extension into one call:
// the bundled series, or a total-return proxy converted into the asset's
// currency, spliced in front of the real quotes. Writing those files and
// splicing by hand are the generators' business (see "Generator plumbing").
//
// # A series as data
//
// A Series is also the bridge to the slice-based rest of the tree (pkg/metrics
// takes parallel dates and values). NewSeries wraps a consumer's own data
// (strictly ascending dates, normalized to 00:00 UTC, finite values); Dates,
// Values and Returns (simple returns as fractions, the same numbers as
// metrics.Returns) hand out fresh slices, and Stats is metrics.Compute over
// them in one call; Rebase scales a copy to start at a chosen level; LessFee
// deducts a yearly charge on the calendar, as a FRACTION (0.0085 = 0.85 %/yr,
// NOT the percent of Client.Fees); Change is the cumulative return between
// two dates (a named episode: 2008, 2022), refusing a series that cannot
// answer for one of them; Resample keeps the last trading close of each
// calendar month, quarter or year, a month-END series like every bundled
// monthly anchor. Every one of them returns fresh slices and leaves its
// receiver alone, and a bad argument (a charge of the whole level, a negative
// frequency, and on a Panel an unknown column or period) is an error, never
// a panic.
//
// Several series meet on one calendar through AlignSeries, the strict sibling
// of Align: it starts by default at CommonWindow's start (the latest first
// quote) and returns an error naming the series where Align would forward-fill
// zeros; its Aligned result hands the per-asset returns a correlation takes.
// Align itself stays for callers that compute their own start (portfolio's
// simulation), SampleAt for an exogenous level read onto a calendar it must
// not shape, and Trim is the window operation on one series.
//
// # Returns panels
//
// Aligned and Panel answer two different questions, and the difference is
// the calendar. Aligned holds LEVELS on the UNION of the series' quoting
// days, forward-filled across the days a series did not quote: what a
// simulation needs, a portfolio being worth something every day any of its
// holdings trades. Panel holds RETURNS on the STRICT intersection of their
// periods, each return spanning the same interval in every column: what a
// statistic across series needs, since a forward-filled day reads as a zero
// return beside another series' real move and biases every correlation
// toward zero.
//
// NewPanel cuts each series into periods of a Frequency (Daily, the quote
// dates themselves, or Monthly, Quarterly, Yearly), labels every period
// CANONICALLY (the calendar month-end at 00:00 UTC, whatever day a source
// happened to close its month on, so panels from different sources join
// exactly) and keeps the labels every series has. A series' unfinished last
// period is not a period, a monthly hole inside the shared window is an
// error naming the series, and a period crossing a definition junction is
// dropped for every column (NewPanel's godoc holds the rules and why).
//
// From there, everything is a column or a selection of periods: Col hands a
// column to pkg/metrics (Mean, Regress, LowestK, Corr), R is the
// [asset][period] matrix CorrelationMatrix and Covariance take; Between keeps
// an episode, Pick a sample of periods (the reference's worst decile, found
// by metrics.LowestK); Mix appends a blend rebalanced every period, weights
// summing to 1 with any financing as an explicit cash column; Series
// rebuilds a column as a level, so Stats scores a blend in one call; Track
// reads one column against another (metrics.Track: correlation, tracking
// error and difference, beta, alpha), the figures of a replica or a
// reconstruction against what it follows; Compound(5) reads a daily panel
// at a five-session horizon; and PeriodsPerYear is the count that
// annualizes it all, 12 for a monthly panel.
//
// # Toolbox
//
//   - Align merges the trading calendars of several series (union of
//     dates, forward-filled prices); SampleAt reads ONE series onto a
//     calendar somebody else decided, for the exogenous levels (a financing
//     rate, a deflator) that must not add sessions of their own nor
//     forward-fill zeros before their history;
//   - Client.Fees returns an asset's published TER (pinned catalog, disk
//     cache, otherwise FT tearsheets and justETF);
//   - UCITSFlag/GuessUCITS and LooksDistributing qualify funds; IsMinorUnit
//     tells a venue sub-unit ("GBp", "ZAc") from a currency, for a caller
//     validating records of its own (nothing served here carries one);
//   - ClassBand returns an asset class's plausibility Band, the table the
//     doctor judges by and the cleaning pass is gated on;
//   - CanonicalID normalizes any accepted identifier (alias, ISIN, ticker
//     from the embedded list) to its canonical form; KnownLocal reports
//     whether it resolves without a network lookup, and LocalCatalog
//     enumerates that whole set, one entry per canonical id;
//   - IsISIN validates an ISIN, check digit included; PlausibleID judges any
//     identifier on its shape alone (a valid ISIN or a plausible exchange
//     ticker), the first gate for identifiers a caller did not vet itself,
//     to be paired with FetchOptions.ExactOnly so that a typo fails rather
//     than adopting a fund a full-text search merely liked;
//   - Client.Cached reports whether an identifier's history is already in
//     the disk cache, fresh and deep enough to answer without any upstream
//     request, for callers that must ration those;
//   - Client.ConvertCurrency reprices a whole Series into a target currency
//     using daily Yahoo FX crosses; the earliest known rate is held flat
//     before the FX history starts;
//   - Client.Latest returns the freshest known price (a Quote) for an
//     identifier, the live Yahoo market price when available, otherwise the
//     last daily close; Client.LatestBatchExtended and
//     QuoteOptions.ExtendedHours add the pre-market and after-hours prints,
//     labelled by Quote.Session;
//   - Client.Resolve returns a Resolution describing the instrument pofo
//     would quote for an identifier (ticker, ISIN or alias), using the catalog and
//     on-disk cache first, then the same multi-source search Fetch uses;
//     calling Resolve before Fetch lets callers inspect the resolved
//     source and symbol, and the result is cached so a subsequent Fetch
//     reuses the same work.
//
// # Generator plumbing
//
// These are exported for the data generators and data modes under cmd/ (and,
// for ExtendBack, for pkg/simgen, which the generators drive) and are not
// meant for consumers, whose entry point is Client.FetchExtended:
//
//   - WarmupIDs lists every catalog identifier, the set "pofo -warmup",
//     "pofo -verify-data" and the -suggest candidate pool walk;
//   - SimdataFile and WriteSimdata write a simdata or refdata CSV, the format
//     ReadSimdata reads back, and EndsHeader spells its "# ends:" line for the
//     generators that format their own headers;
//   - ExtendBack splices a proxy in front of a series, rescaled at the join
//     and marked by SimulatedBefore, the step FetchExtended and every simgen
//     recipe build on;
//   - ProxySymbol names the long-history proxy FetchExtended splices behind a
//     symbol.
package marketdata
