// Package compare computes the portfolio comparison model and assembles the
// report Page from it.
//
// It is presentation over pkg/analyze, the numbers layer, and it sits between
// that layer and the renderers (report, chart). Every column Compute builds is
// an analyze.Portfolio study (fetch, build, simulation, per-holding studies,
// composition, risk attribution), made through one memoizing source so a
// series two columns share is fetched once; Comparison.Studies hands those
// studies to a caller who wants the numbers behind every chart rather than the
// page. What compare adds is what only a comparison of several columns owns:
// the "#meta currencies" expansion (one study per currency), the "#meta
// optimize" column (the optimizer runs here, then its weights are studied like
// any others), the benchmark, the window common to every column, and the
// nominal and real (CPI-deflated) statistics on that window. A column whose
// window is its whole simulation reads its study's own statistics; a column
// the common window cuts is measured again on that slice of the same series.
// HTMLPage then turns the Comparison into a report.Page ready for
// report.Render.
//
// The package returns models, never I/O: it has no knowledge of the web server,
// the terminal, or CLI flags. All caller intent arrives through Options (base
// currency, benchmark, window, rebalancing, SIM/fee toggles, embedded simdata,
// suggestion framework, and ExactForeign, which resolves identifiers outside
// the bundled catalog exactly so a typo composed by an untrusted hand fails
// instead of quoting an unrelated fund) and Decoration (skin CSS, site
// chrome). REAL versus nominal accounting is handled inside (deflation by the
// base currency's CPI), so callers get both nominal and inflation-adjusted
// statistics without wiring the deflator themselves.
//
// The pipeline shape is:
//
//	Compute(...) -> *Comparison -> HTMLPage / StatRows / Columns / Studies
//
// Each portfolio's detail section is assembled from three blocks that answer
// three different questions about the same holdings: the composition pies say
// where the MONEY sits (look-through geography, currency, sectors, asset
// class), the coverage bars say which macro regimes it claims to cover a
// priori, and the risk budget says where the RISK actually sits, next to the
// capital and realized-return shares of the same classes. The three routinely
// disagree, and that disagreement is the reason all three are shown.
//
// Comparison keeps its per-column compute records private; accessors
// (CommonStart, CommonEnd, Columns) expose the narrow public view a caller
// needs without leaking the internal record, and Studies the numbers layer
// underneath it, one analyze.PortfolioStudy per column.
//
// # Optimized columns
//
// A spec carrying "#meta optimize:" produces TWO columns, the weights as
// written and the weights the optimizer chose, so the two can be read side by
// side. This package resolves the per-asset bounds (which arrive keyed by
// identifier), applies the "train:" window by slicing the returns to it, and
// writes Column.Note: which objective ran under which limits, over which
// window, the weights it landed on, and, when the fit used only part of the
// history, how those weights behaved over the stretch they did not see. That
// last clause is what keeps an optimizer honest, so it is built here rather
// than left to each renderer.
//
// # Sweep
//
// Sweep answers the neighbouring question: not "which weights are best" but
// "what does each weight buy, and what does it cost". It re-runs the real
// simulation with one holding's weight moved across a grid, the others
// keeping their relative proportions, and reports CAGR, volatility, Sharpe,
// drawdown, recovery time and the worst five-year stretch at every point. It
// is the evidence behind a portfolio file's per-sleeve "sane range", and it
// is deliberately model-free: no optimizer, no objective, just the
// consequences of a number.
package compare
