// Package optimize computes portfolio weights that optimize a risk/return
// objective from the historical returns of the candidate assets.
//
// Six objectives are supported:
//
//   - MaxSharpe ("max-sharpe"): the tangency portfolio, maximizing the
//     ratio of expected return to volatility.
//   - MinVolatility ("min-volatility"): the lowest-variance portfolio.
//   - RiskParity ("risk-parity"): every asset contributes the same share
//     of the total risk.
//   - MaxSortino ("max-sortino"): maximizes the portfolio's own Sortino
//     ratio (return over downside deviation), rewarding assets that cut the
//     downside, i.e. non-correlation and positive skew.
//   - ReturnToDrawdown ("return-to-drawdown"): maximizes the portfolio's own
//     return-to-maximum-drawdown (a Calmar-style ratio), rewarding shallow
//     drawdowns.
//   - CWARP ("cwarp"): the blend that best improves a replacement portfolio
//     (a benchmark) when overlaid on it. It is solved by SolveCWARP, which
//     takes the replacement's returns as an extra argument; the objective is
//     non-convex and non-smooth (it depends on the combined drawdown), so the
//     solver is a multi-start heuristic and its weights are a good allocation
//     rather than a certified optimum.
//
// Weights are long-only (no short selling) and sum to 1. An optional
// per-asset cap (MaxWeight) bounds concentration for every objective except
// RiskParity, whose weights follow directly from the equal-risk condition.
//
// MaxSharpe and MinVolatility optimize the mean and covariance in closed-ish
// form; MaxSortino, ReturnToDrawdown and CWARP depend on the whole return
// path, so they are solved by the same multi-start heuristic as MaxSharpe
// (their weights are a good allocation, not a certified optimum). CWARP also
// needs a replacement series and is solved by SolveCWARP, not Solve.
//
// Conventions match pkg/metrics: simple daily returns, 252 trading days per
// year and a risk-free rate of 0. The estimates returned in Result are
// in-sample (they describe the optimization window); the realized figures
// after simulation, with rebalancing and fees, will differ.
package optimize
