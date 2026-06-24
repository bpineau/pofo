// Package optimize computes portfolio weights that optimize a risk/return
// objective from the historical returns of the candidate assets.
//
// Three objectives are supported:
//
//   - MaxSharpe ("max-sharpe"): the tangency portfolio, maximizing the
//     ratio of expected return to volatility.
//   - MinVolatility ("min-volatility"): the lowest-variance portfolio.
//   - RiskParity ("risk-parity"): every asset contributes the same share
//     of the total risk.
//
// Weights are long-only (no short selling) and sum to 1. An optional
// per-asset cap (MaxWeight) bounds concentration for MaxSharpe and
// MinVolatility; it does not apply to RiskParity, whose weights follow
// directly from the equal-risk condition.
//
// Conventions match pkg/metrics: simple daily returns, 252 trading days per
// year and a risk-free rate of 0. The estimates returned in Result are
// in-sample (they describe the optimization window); the realized figures
// after simulation, with rebalancing and fees, will differ.
package optimize
