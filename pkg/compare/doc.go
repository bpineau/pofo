// Package compare computes the portfolio comparison model and assembles the
// report Page from it.
//
// It is the presentation-neutral core that sits between the library pipeline
// (marketdata, portfolio, metrics, suggest) and the renderers (report, chart):
// Compute fetches each spec in its base currency, runs the shared simulation,
// aligns every column on the common window, and folds the per-portfolio compute
// records into a Comparison. HTMLPage then turns that Comparison into a
// report.Page ready for report.Render.
//
// The package returns models, never I/O: it has no knowledge of the web server,
// the terminal, or CLI flags. All caller intent arrives through Options (base
// currency, benchmark, window, rebalancing, SIM/fee toggles, embedded simdata,
// suggestion framework) and Decoration (skin CSS, site chrome). REAL versus
// nominal accounting is handled inside (deflation by the base currency's CPI),
// so callers get both nominal and inflation-adjusted statistics without wiring
// the deflator themselves.
//
// The pipeline shape is:
//
//	Compute(...) -> *Comparison -> HTMLPage / StatRows / Columns
//
// Comparison keeps its per-column compute records private; accessors
// (CommonStart, CommonEnd, Columns) expose the narrow public view a caller
// needs without leaking the internal record.
package compare
