// Package pofo is the root of a dependency-free Go toolkit for tracking
// and designing stock-market portfolios. The command in cmd/pofo is one
// application built on it; every capability it exposes is reusable from the
// library packages under pkg/. README.md's chapter "Using it as a library"
// walks the library by task, each snippet copied from a runnable example.
//
// # Packages
//
// The toolkit is a layered set of focused packages:
//
//   - pkg/datasets: the bundled, versioned data, embedded into the binary:
//     the curated asset catalog (assetmeta/assets.json, typed as
//     datasets.Asset), the permanent simulated histories (simdata/), the long
//     reference series they are built on (refdata/) and the three research
//     panels behind the FIRE and macro-regime work (broadsample/, cape/,
//     macropanel/). This is the single source of truth other packages read
//     from.
//   - pkg/marketdata: fetches, caches and post-processes daily, intraday and
//     latest (real-time) prices from public sources, addressed by ticker, ISIN
//     or alias; resolves identifiers against the embedded catalog and aligns
//     trading calendars (AlignSeries, the strict one).
//   - pkg/metrics: risk/return statistics on parallel date and value slices
//     (CAGR, volatility, Sharpe, Sortino, Ulcer, max drawdown,
//     time-to-recovery, Beta, CWARP, IRR, TWR), the correlation and
//     covariance matrices, calendar returns, rolling beta and correlation,
//     historical VaR, and the Euler risk attribution.
//   - pkg/optimize: long-only weights for an objective (max-sharpe,
//     min-volatility, max-return, risk-parity, max-sortino,
//     return-to-drawdown, min-ulcer, max-worst-5y, cwarp) from the assets'
//     historical returns, or from the returns the file's own weights imply
//     and the owner's views revise (black-litterman).
//   - pkg/portfolio: the allocation file format, NewSpec for a spec built in
//     code, and the rebalanced, fee-aware simulation that replays a portfolio
//     over time, attributing each day's return to its holdings.
//   - pkg/suggest: structure-first analysis: regime coverage, look-through
//     composition splits (asset classes, geography, currency, sectors,
//     duration), redundancy and out-of-sample-validated gap-filling
//     suggestions from the catalog.
//   - pkg/analyze: the high-level, numbers-only face: one call studies an
//     asset or a portfolio (statistics, calendar tables, drawdown episodes,
//     correlation, risk and return attribution, look-through composition,
//     warnings) and renders nothing.
//   - pkg/simgen: reconstruction of the missing past of complex assets
//     (capital-efficient funds, managed futures) into simdata files.
//   - pkg/scenario: synthetic real-return path generation (parametric
//     Student-t, block/stationary bootstrap, historical cohorts) behind one
//     Source interface; the input to decumulation studies.
//   - pkg/decumul: decumulation/FIRE engine over a scenario.Source: ruin
//     probability, FIRE outcome metrics, capital/buffer sizing and sweeps,
//     with a thin embedded live UI under pkg/decumul/web.
//   - pkg/replay: the same withdrawal rules run over the years as they
//     actually happened, on a bundled real US 60/40, so a rule can be
//     described by the life it delivered rather than by a failure
//     probability; consumed by the FIRE book's historical replay article.
//   - pkg/chart: dependency-free SVG and terminal charts (line, pie, bars,
//     heatmap).
//   - pkg/report: HTML and text rendering of a portfolio-comparison model.
//   - pkg/compare: the reusable comparison pipeline shared by the CLI and the
//     -serve web app: each column an analyze study, plus what only a
//     comparison owns (the common window, the benchmark, the real statistics,
//     the optimizer's column) and the HTML report Page.
//   - pkg/webui: the visual identity every HTML surface shares (design
//     tokens, embedded typefaces, favicon, the optional analytics beacon).
//   - pkg/firebook: the embedded decumulation handbook in its two editions,
//     served, exported and syndicated over four content-agnostic packages:
//     pkg/bookmd (the book-Markdown dialect), pkg/epub (EPUB 3 writer),
//     pkg/opds (OPDS 1.2 catalog) and pkg/seo (sitemap, robots.txt,
//     llms.txt, Atom, IndexNow).
//
// # Layering
//
// A package imports only packages on an earlier line; this is the graph as
// measured by go list, each package followed by what it imports of pofo:
//
//	datasets, metrics, chart, webui, bookmd, epub, opds, seo    (nothing)
//	marketdata   datasets
//	optimize     metrics
//	suggest      datasets metrics
//	scenario     marketdata
//	simgen       marketdata metrics
//	decumul      metrics scenario
//	portfolio    marketdata optimize
//	replay       datasets decumul marketdata metrics scenario
//	analyze      datasets marketdata metrics portfolio suggest
//	report       webui
//	firebook     bookmd epub opds seo webui
//	compare      analyze chart datasets marketdata metrics optimize portfolio report suggest
//	decumul/web  chart datasets decumul firebook marketdata metrics replay scenario webui
//
// Two edges are rules, not accidents: metrics imports nothing of pofo (it is
// the math, and takes a valuation series its caller built), and analyze never
// imports chart, report or compare (the numbers come before any picture).
//
// # Typical pipeline
//
// The high-level path is two calls, a spec and a study:
//
//	client := marketdata.NewClient(marketdata.DefaultCacheDir())
//	spec, _ := portfolio.NewSpec("60/40",
//		portfolio.Line{ID: "IWDA", Weight: 0.6},
//		portfolio.Line{ID: "AGGH", Weight: 0.4})
//	study, _ := analyze.Portfolio(ctx, client, spec, analyze.Options{Currency: "EUR"})
//
// Every step underneath stays reachable for a custom pipeline: resolve and
// fetch each holding's series with pkg/marketdata (splicing simulated history
// from pkg/datasets when a fund is young), build and simulate the portfolio
// with pkg/portfolio, score it with pkg/metrics, and either render it with
// pkg/chart and pkg/report or improve it with pkg/optimize and pkg/suggest:
//
//	p, _ := portfolio.Build(spec, portfolio.BuildOptions{
//		Fetch: func(id string) (*marketdata.Series, error) {
//			return client.FetchExtended(ctx, id, marketdata.FetchOptions{Currency: "EUR"})
//		},
//	})
//	sim, _ := portfolio.Simulate(p, 90)
//	stats, _ := metrics.Compute(sim.Dates, sim.Index)
//
// pkg/compare packages the CLI's whole wiring (several columns, a common
// window, nominal and real statistics, the report Page), so the CLI and the
// -serve web app share one comparison pipeline.
//
// # Conventions and units
//
// Series are daily closes; volatility and ratios annualize over 252 trading
// days with a zero risk-free rate, the CAGR over 365.25-day years. Units are
// the library's number one trap, and one table holds them:
//
//	weights          FRACTION in memory (portfolio.Line, Holding.Weight, Asset.Weight,
//	                 optimize, analyze); PERCENT in portfolio files and Holding.RawWeight
//	fees (TER)       PERCENT per year in portfolio (Line.Fees, Holding.Fees,
//	                 EnvelopeFees, BorrowSpread), marketdata Client.Fees and
//	                 datasets.Asset.Fees; FRACTION per year in simgen (and its
//	                 volatility targets)
//	returns          FRACTION everywhere (metrics, analyze, scenario, decumul:
//	                 0.04 = +4 %); scenario and decumul work in REAL terms
//	statistics       FRACTION (metrics.Stats, analyze studies), except
//	                 Stats.Ulcer in percent points and Stats.CWARP in percent
//	rates            annualized PERCENT LEVELS (^IRX, ^ESTR, ^SOFR..., and
//	                 portfolio.Portfolio.Cash); never a return
//	#meta directives PERCENT as written (max-vol:9, view:ID:8@70), FRACTION
//	                 once parsed into optimize.Spec
//
// # API versus plumbing
//
// Most exported symbols are the API. A few exist for the data generators and
// data modes under cmd/ (writing simdata files, grading a reconstruction for
// -verify-simdata, building one bundled donor) and are not meant for
// consumers: pkg/simgen and pkg/marketdata list them in a final "Generator
// plumbing" section of their package documentation, so a reader can tell the
// toolkit from the machinery that maintains its data.
//
// This package itself holds no code: start from the package docs above
// (for example, go doc github.com/bpineau/pofo/pkg/analyze).
package pofo
