// Package metrics computes risk and return statistics for a series of
// dated values: CAGR, volatility, Sharpe, Sortino, Ulcer Index, Max
// Drawdown, TTR (time to recovery) and Beta against a benchmark.
//
// # Conventions
//
// Knowing the conventions is essential to compare the results with other
// tools:
//
//   - series are daily closes; volatility and ratios are annualized over
//     252 trading days;
//   - Sharpe and Sortino use a zero risk-free rate (like Curvo);
//     PortfolioVisualizer and LazyPortfolioETF use T-bills and monthly
//     data; their Sharpe ratios come out ≈ 0.10–0.15 lower;
//   - Max Drawdown, Ulcer and TTR are measured on daily closes, harsher
//     than monthly-step tools (COVID 2020: −33.7 % on daily closes
//     versus ≈ −20 % on monthly closes);
//   - the CAGR uses 365.25-day years between the first and the last
//     date.
//
// The main entry point is Compute. Beta pairs daily returns with the
// benchmark's by date. Returns and Mean are exposed as building blocks.
//
// These computations are locked down by the golden package's benchmark
// tests, which check them against external references (official S&P 500
// TR annual returns, canonical drawdowns).
package metrics
