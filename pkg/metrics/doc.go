// Package metrics computes risk and return statistics for a series of
// dated values: CAGR, volatility, Sharpe, Sortino, Ulcer Index, Max
// Drawdown, TTR (time to recovery), Beta against a benchmark, and the
// daily-versus-monthly volatility term structure (the Lo-MacKinlay
// variance ratio).
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
// benchmark's by date. VarianceRatio resamples to month-end closes and
// reports the volatility at both frequencies plus their ratio, revealing
// the autocorrelation the daily statistics hide. Returns and Mean are
// exposed as building blocks.
//
// # External flows
//
// When a series carries external contributions and withdrawals (a savings
// account, a wealth tracker), Compute's raw figures would mistake them for
// performance. TWR chain-links flow-neutralized daily returns, FlowReturns
// yields the flow-adjusted daily returns (weekend points dropped, so
// calendar-daily forward-filled series keep an honest volatility), and
// Volatility, Sharpe and Sortino accept those returns with an explicit
// annual risk-free rate (Compute's convention stays rf = 0). Annualize
// turns a cumulative return over a day span into a compound annual rate,
// and IRR solves the money-weighted rate of the flows themselves.
//
// These computations are locked down by the golden package's benchmark
// tests, which check them against external references (official S&P 500
// TR annual returns, canonical drawdowns).
package metrics
