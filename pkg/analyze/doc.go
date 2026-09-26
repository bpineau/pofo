// Package analyze is the high-level, numbers-only face of the library: one
// call studies an asset (Asset) or a portfolio (Portfolio) and returns every
// number the comparison report is drawn from, as plain values, with nothing
// rendered.
//
//	spec, _ := portfolio.NewSpec("60/40",
//		portfolio.Line{ID: "IWDA", Weight: 0.6},
//		portfolio.Line{ID: "AGGH", Weight: 0.4})
//	study, _ := analyze.Portfolio(ctx, client, spec, analyze.Options{Currency: "EUR"})
//	fmt.Println(study.Stats.CAGR, study.Correlation[0][1], study.Attribution.Risk)
//
// It is the pipeline marketdata -> portfolio -> metrics -> suggest wired once,
// with its traps closed: the holdings are aligned with marketdata.AlignSeries
// (never Align's zero fill), each holding is studied on the portfolio's own
// window, the SIM convention and the fee lookup follow the CLI. The data comes
// from a Source, which *marketdata.Client satisfies; a consumer with its own
// store, or a test, supplies another.
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
