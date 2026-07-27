// Package golden contains pofo's golden tests: they replay the
// simulation and the metrics on frozen real data (testdata/) and compare
// the results against hand-validated external references (official S&P 500
// TR annual returns, canonical drawdowns, statistics published by
// LazyPortfolioETF). Any drift in the computations (CAGR, volatility,
// Sharpe, Sortino, Ulcer, Max Drawdown, TTR) beyond the tolerances fails
// the suite.
//
// Three families live here:
//
//   - golden_test.go pins the COMPUTATIONS on frozen daily fixtures.
//   - refdata_test.go pins the bundled long backcast SERIES
//     (pkg/datasets/refdata) against published index returns.
//   - aqrmf_test.go pins a bundled fund series against AUDITED net asset
//     values per share, to the cent: a fund NAV has one true value per day,
//     which also proves the identifier resolved to the right share class.
package golden
