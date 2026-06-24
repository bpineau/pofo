// Package golden contains pofo's golden tests: they replay the
// simulation and the metrics on frozen real data (testdata/) and compare
// the results against hand-validated external references (official S&P 500
// TR annual returns, canonical drawdowns, statistics published by
// LazyPortfolioETF). Any drift in the computations — CAGR, volatility,
// Sharpe, Sortino, Ulcer, Max Drawdown, TTR — beyond the tolerances fails
// the suite.
package golden
