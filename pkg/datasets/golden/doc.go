// Package golden contains pofo's golden tests: they replay the
// simulation and the metrics on frozen real data (testdata/) and compare
// the results against hand-validated external references (official S&P 500
// TR annual returns, canonical drawdowns, statistics published by
// LazyPortfolioETF). Any drift in the computations (CAGR, volatility,
// Sharpe, Sortino, Ulcer, Max Drawdown, TTR) beyond the tolerances fails
// the suite.
//
// Five families live here:
//
//   - golden_test.go pins the COMPUTATIONS on frozen daily fixtures.
//   - refdata_test.go pins the bundled long backcast SERIES
//     (pkg/datasets/refdata) against published index returns.
//   - aqrmf_test.go pins a bundled fund series against AUDITED net asset
//     values per share, to the cent: a fund NAV has one true value per day,
//     which also proves the identifier resolved to the right share class.
//   - blacklitterman_test.go pins the optimizer's Black-Litterman step
//     (reverse optimization and the Bayesian blend) on the published tables
//     of He and Litterman (1999) and Idzorek (2005): no bundled data at all,
//     only the model's own literature.
//   - trendcadence_test.go pins the managed-futures files' donor era to a
//     CADENCE invariant: a weekly-dealing donor is projected onto a daily
//     calendar there, and a projection that gets the daily amplitude wrong
//     leaves every level right and every per-observation statistic wrong.
package golden
