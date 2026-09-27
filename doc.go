// Package pofo is the root of a dependency-free (standard library only) Go
// toolkit for deciding and checking how a long-horizon portfolio is built:
// market data with decades of bundled history, risk and return statistics,
// portfolio simulation and optimization, withdrawal (FIRE) studies, and the
// reconstruction of the past of young funds. The command in cmd/pofo is one
// application built on it; every capability it exposes is reusable from the
// library packages under pkg/. This package itself holds no code.
//
// # Start here
//
// Import the package that owns the question. Much of the data is bundled
// with the module (indices since 1871, yields, gold, trend and bond
// references, the backcast of every catalog fund), so most of these work
// offline.
//
//   - Load a series ([marketdata]): Bundled (embedded, no network),
//     NewClient then Client.FetchExtended (live quotes, cached on disk;
//     Client.Offline stays off the network), ReadCSV (a file of your own);
//     Client.Load tries the three in turn (a path is a file, then the
//     bundle, then the client), the one call an exploration script needs.
//   - Describe it ([metrics], [analyze]): Series.Stats, which returns a
//     metrics.Stats; metrics.CalendarReturns, metrics.DrawdownEpisodes; or
//     analyze.Asset for everything at once.
//   - Compare series: analyze.Pair (a candidate against its reference: a
//     backcast against its fund, a file against its previous version);
//     marketdata.AlignSeries then metrics.CorrelationMatrix.
//   - Returns panels, blends, conditional statistics: marketdata.NewPanel,
//     then Panel.Mix (a rebalanced blend), Panel.Series, Panel.Pick with
//     metrics.LowestK (the worst months, dated), Panel.Track.
//   - Regression: metrics.Regress on panel columns.
//   - A portfolio ([portfolio]): analyze.Portfolio in one call, or
//     portfolio.Parse or NewSpec, then Build and Simulate.
//   - Optimize weights ([optimize]): ParseSpec, then Solve.
//   - FIRE and decumulation ([decumul], [scenario], [replay]): a
//     decumul.Plan over a scenario.Source, then Plan.Simulate and
//     Plan.Solve; replay.Run for history as it happened.
//   - Backcasts ([simgen]): Find and Validate; the shipped ones are read by
//     marketdata.Bundled or behind a SIM identifier.
//   - Render ([chart], [compare], [report]): chart.Line, or compare.Compute
//     then report.Render for the CLI's HTML report.
//   - Export: marketdata.WriteCSV.
//
// A complete program, offline (the package example below):
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//
//		"github.com/bpineau/pofo/pkg/marketdata"
//	)
//
//	func main() {
//		// The S&P 500 total return since 1871, bundled with the module.
//		sp500, err := marketdata.Bundled("SP500-USD")
//		if err != nil {
//			log.Fatal(err)
//		}
//		bonds, err := marketdata.Bundled("TREASURY-INT-USD")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// Monthly returns on the months both share, and a 60/40 rebalanced
//		// every month (weights are fractions).
//		p, err := marketdata.NewPanel(marketdata.Monthly, sp500, bonds)
//		if err != nil {
//			log.Fatal(err)
//		}
//		p, err = p.Mix("60/40", map[string]float64{"SP500-USD": 0.6, "TREASURY-INT-USD": 0.4})
//		if err != nil {
//			log.Fatal(err)
//		}
//		for _, id := range p.IDs {
//			s, err := p.Series(id)
//			if err != nil {
//				log.Fatal(err)
//			}
//			st, err := s.Stats()
//			if err != nil {
//				log.Fatal(err)
//			}
//			fmt.Printf("%-16s CAGR %5.2f %%/yr, volatility %4.1f %%/yr, max drawdown %5.1f %%\n",
//				id, st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)
//		}
//	}
//
// Longer programs, one question each, live in examples/lib (README.md there
// is the index): describe a series, blend, regress, the worst months,
// a reconstruction against its reference, a portfolio file simulated, a FIRE
// run, a CSV export. README.md's chapter "Using it as a library" walks the
// library by task, each snippet a verbatim copy of a runnable example, and
// every package's own documentation opens on the calls to start with. The
// same questions without writing Go: "pofo -dump", "pofo -pair", "pofo
// -verify-simdata -json", each with -offline.
//
// # Units and conventions
//
// Units are the library's number one trap, and this table holds them (the
// README's Units table says the same):
//
//	weights          FRACTION in memory (portfolio.Line, Holding.Weight, Asset.Weight,
//	                 Panel.Mix, optimize, analyze); PERCENT in portfolio files and
//	                 Holding.RawWeight
//	fees (TER)       PERCENT per year in portfolio (Line.Fees, Holding.Fees,
//	                 EnvelopeFees, BorrowSpread), marketdata Client.Fees and
//	                 datasets.Asset.Fees; FRACTION per year in simgen (and its
//	                 volatility targets) and marketdata Series.LessFee
//	returns          FRACTION everywhere (metrics, analyze, scenario, decumul:
//	                 0.04 = +4 %); scenario and decumul work in REAL terms
//	statistics       FRACTION (metrics.Stats, analyze studies), except
//	                 Stats.Ulcer in percent points and Stats.CWARP in percent
//	per period or    a return, a Regression, a Covariance and a VaR are PER
//	  annualized     PERIOD; CAGR, volatility, Sharpe, Sortino, tracking error
//	                 are ANNUALIZED (Regression.AnnualAlpha converts)
//	cadence          PERIODS PER YEAR (252 daily, 52 weekly, 12 monthly):
//	                 metrics.PeriodsPerYear, Stats.PeriodsPerYear,
//	                 Panel.PeriodsPerYear, the periodsPerYear argument of the
//	                 bare-return functions and of optimize.Solve
//	dates            every Point.Date is a session date at 00:00 UTC; series
//	                 match by exact time.Time equality; a monthly series is
//	                 labelled by its month-END
//	closes           ADJUSTED (dividends reinvested, total return) by
//	                 default; FetchOptions.Raw gives unadjusted closes with
//	                 Series.Dividends beside them (never both adjusted and
//	                 dividends); a distributing fund's NAV is a PRICE return
//	SIM suffix       "IWDA" is real quotes only; "IWDASIM" splices the bundled
//	                 backcast in front (Client.FetchExtended; Series.SimulatedBefore
//	                 marks the join); "#meta sim:on" asks it for a whole file
//	rates            annualized PERCENT LEVELS (^IRX, ^ESTR, ^SOFR..., and
//	                 portfolio.Portfolio.Cash); never a return
//	#meta directives PERCENT as written (max-vol:9, view:ID:8@70), FRACTION
//	                 once parsed into optimize.Spec
//
// Volatility and ratios annualize at each series' own cadence with a zero
// risk-free rate, the CAGR over 365.25-day years; the conventions section of
// the metrics documentation says why the figures differ from other tools'.
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
//     or alias; resolves identifiers against the embedded catalog, aligns
//     trading calendars (AlignSeries, the strict one) and cuts several
//     series into a returns Panel on the periods they share, the table a
//     statistic across series reads (blends, regressions, the other
//     series over one series' worst months). It is also the one
//     way in and out for data at rest: Bundled reads any embedded series
//     without a client, an Offline client serves the quote cache without
//     the network, ReadCSV and ReadLongCSV read CSV files, WriteCSV writes
//     them.
//   - pkg/metrics: risk/return statistics on parallel date and value slices
//     (CAGR, volatility, Sharpe, Sortino, Ulcer, max drawdown,
//     time-to-recovery, Beta, CWARP, IRR, TWR), the correlation and
//     covariance matrices, calendar returns, rolling beta and correlation,
//     historical VaR, multiple regression (Regress), dated extremes
//     (LowestK, HighestK), tracking error, lead-lag gaps (LeadLagGaps, two
//     closes struck at different hours) and the Euler risk attribution.
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
//     warnings) and renders nothing; Pair measures one series against its
//     reference (a backcast against the fund, a file against its previous
//     version) and prints itself as text.
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
//	marketdata   datasets metrics
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
// marketdata importing metrics is the first rule's consequence, not an
// exception to it: the data package hands its series to the math (Stats,
// Panel.Series, a panel's cadence), and the math never learns what a Series
// is.
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
// # API versus plumbing
//
// Most exported symbols are the API. A few exist for the data generators and
// data modes under cmd/ (writing simdata files, grading a reconstruction for
// -verify-simdata, building one bundled donor) and are not meant for
// consumers: pkg/simgen and pkg/marketdata list them in a final "Generator
// plumbing" section of their package documentation, so a reader can tell the
// toolkit from the machinery that maintains its data.
//
// [marketdata]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/marketdata
// [metrics]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/metrics
// [analyze]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/analyze
// [portfolio]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/portfolio
// [optimize]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/optimize
// [decumul]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/decumul
// [scenario]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/scenario
// [replay]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/replay
// [simgen]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/simgen
// [chart]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/chart
// [compare]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/compare
// [report]: https://pkg.go.dev/github.com/bpineau/pofo/pkg/report
package pofo
