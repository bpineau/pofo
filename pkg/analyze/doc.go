// Package analyze is the high-level, numbers-only face of the library: one
// call studies an asset or a portfolio and returns every number the
// comparison report is drawn from, as plain values, with nothing rendered;
// another measures one series against its reference.
//
// # Start here
//
//   - [Asset] studies one identifier on its longest window: [metrics.Stats],
//     calendar years and months, drawdown episodes, relative statistics
//     against a benchmark ([AssetStudy]).
//   - [Portfolio] builds and simulates a [portfolio.Spec] and studies it: the
//     statistics, each holding on the same window, the correlation matrix,
//     the risk and return attribution, the look-through composition
//     ([PortfolioStudy]).
//   - [Pair] measures a candidate series against a reference (a backcast
//     against its fund, a file against its previous version), and
//     [PairStudy.WriteText] prints the result.
//
// Asset and Portfolio read their data through a [Source], which
// [*marketdata.Client] satisfies; Pair takes two [*marketdata.Series].
//
//	spec, _ := portfolio.NewSpec("60/40",
//		portfolio.Line{ID: "IWDA", Weight: 0.6},
//		portfolio.Line{ID: "AGGH", Weight: 0.4})
//	study, _ := analyze.Portfolio(ctx, client, spec, analyze.Options{Currency: "EUR"})
//	fmt.Println(study.Stats.CAGR, study.Correlation[0][1], study.Attribution.Risk)
//
// It is the pipeline marketdata -> portfolio -> metrics -> suggest wired once,
// with its traps closed: the holdings are aligned with
// [marketdata.AlignSeries] (never Align's zero fill), each holding is studied
// on the portfolio's own window, the SIM convention and the fee lookup follow
// the CLI. A consumer with its own store, or a test, supplies another Source.
//
// # Units
//
// Every return, rate, weight and share is a FRACTION (0.07 = +7 %, 0.6 = 60 %),
// as in pkg/metrics, whose Stats keep their two exceptions: Ulcer is in
// percent points and CWARP in percent. Composition values are fractions of
// the portfolio's capital, Sectors fractions of its equity sleeve. The
// percent conventions of pkg/portfolio (RawWeight, TERs in percent per year)
// are the spec's and stay there.
//
// # Windows
//
// Asset studies the longest window its data covers, clipped to
// [Options.From, Options.To]. Portfolio studies Simulate's window, the dates
// on which every holding quotes, inside the same bounds, and studies each
// holding on THAT window, on its own quoting calendar, so the holdings'
// statistics compare with each other and with the portfolio's. A holding's
// Stats.Start is therefore the window's first day unless it did not quote
// on it.
//
// Nominal statistics only: an inflation-adjusted reading needs a consumer
// price index per currency, which pkg/compare owns.
//
// # Pairs
//
// Pair compares a candidate series with a reference, the question behind
// every data check: a reconstruction against the real fund, a fund against
// its index, a refreshed bundled file against the version it replaces (read
// through marketdata.ReadCSV from "git show HEAD~1:<path>"). It takes two
// series, not identifiers, so it works on anything a caller holds, and reads
// them on the window both cover:
//
//   - level and identity, at any cadence, off each series' close at or
//     before each bound: the CAGR gap with its standard error (a gap inside
//     two of them measures nothing), the level gap once both are rebased,
//     the ratio A/B on the first date and the first shared date it moves (a
//     rescaled copy forgiven), the dates only one side holds, and the
//     calendar years side by side, which chain to the window's total;
//   - a Monthly block on the calendar months both quote (a
//     marketdata.Panel, so a fund's last trading close meets an index's
//     calendar month-end) and, when both series are finer than monthly, a
//     Daily block on the dates both quote: correlation, volatilities and
//     their ratio, tracking error (metrics.TrackingError), beta and alpha
//     (metrics.Regress), and the periods where the two disagree most, dated.
//     PairOptions.LeadLag ranks the daily ones by metrics.LeadLagGaps, which
//     forgives a one-session difference in closing times;
//   - Warnings for what makes a figure unreliable: a window under two years,
//     a currency or cadence mismatch, each side's own caveats (a junction, a
//     reconstructed or estimated stretch), monthly dates on the first of the
//     month (a close labelled by the month it opens reads one month off).
//
// PairStudy.WriteText prints it all as aligned text; the struct marshals to
// JSON with no NaN in it, a figure that cannot be measured being left zero
// with a warning or a nil block. The pofo CLI serves it as -pair.
//
// # Errors and warnings
//
// An identifier no source quotes, in a spec, as the asset or as the
// benchmark, is an error that satisfies errors.Is(err,
// marketdata.ErrUnknownIdentifier). So is a request no study can honor: an
// empty identifier, a window ending before it starts, a spec pkg/portfolio
// refuses to build or simulate.
//
// Every other data problem is a line in Warnings, never a silent number:
// points before a date reconstructed rather than quoted (SIM), a nowcast
// tail, a distributing share class quoted as a NAV (a price return whose
// income is missing), closes that are not dividend-adjusted, a definition
// change inside the window, a currency left unconverted, a benchmark that
// failed to load or shares too few dates, a financing rate held flat or
// missing, a portfolio wiped out. A PortfolioStudy carries the portfolio's
// own warnings and every holding's, the latter prefixed with its identifier.
//
// One problem is not seen here: marketdata.Client holds an FX rate flat
// before its history when converting, and reports it only in its log
// (Client.Logf), not on the series.
package analyze
