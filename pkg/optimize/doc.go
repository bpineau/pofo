// Package optimize computes portfolio weights that optimize a risk/return
// objective from the historical returns of the candidate assets.
//
// Eight objectives are supported:
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
//   - MinUlcer ("min-ulcer"): minimizes the Ulcer Index (root-mean-square
//     drawdown), shortening and shallowing the underwater periods that are
//     hard to sit through.
//   - MaxWorst5y ("max-worst-5y"): maximizes the worst rolling five-year
//     return, the robust worst-case medium-term outcome.
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
// # How each objective is solved, and how much to trust it
//
// MaxSharpe, MinVolatility and RiskParity read only the mean vector and the
// covariance matrix; MaxSortino, ReturnToDrawdown, MinUlcer, MaxWorst5y and
// CWARP depend on the whole return path. All of them are minimized by
// projected gradient descent over the capped simplex, from several
// deterministic starting points. They differ in what that buys:
//
//   - MinVolatility and RiskParity are convex programs, so their answer is
//     the global optimum.
//   - MaxSharpe is not concave in the weights, but the UNCAPPED tangency
//     portfolio is still obtained exactly, via Schaible's transformation
//     (minimizing yᵀΣy under muᵀy = 1 and y ≥ 0 is convex, and w = y/Σyᵢ
//     maximizes the ratio); that exact solution is one of the starting
//     points, so an uncapped run returns the global optimum and a capped one
//     starts from the right neighbourhood.
//   - MaxSortino, ReturnToDrawdown, MinUlcer, MaxWorst5y and CWARP are
//     non-convex AND non-smooth (they depend on the realized drawdown path),
//     so their weights are a good allocation, not a certified optimum.
//
// CWARP also needs a replacement series and is solved by SolveCWARP, not Solve.
//
// Conventions match pkg/metrics: simple daily returns, 252 trading days per
// year and a risk-free rate of 0. Note that a zero risk-free rate moves the
// tangency point: with rf = 0 the ratio rewards any positive return, so
// cash-like assets score far better than they would against a realistic rate.
//
// # Estimation error, which matters more than the solver
//
// The estimates in Result are in-sample: they describe the optimization
// window, and the realized figures after simulation, with rebalancing and
// fees, will differ. That gap is not a rounding detail for the objectives
// that use expected returns. The standard error of a mean return depends on
// the CALENDAR SPAN of the sample and not at all on how finely it is sampled
// (Merton, "On Estimating the Expected Return on the Market", JFE 1980), so
// feeding the optimizer daily data sharpens the covariance and adds nothing
// to the means. MaxSharpe is driven by exactly the quantity the data
// estimates worst, which is why mean-variance optimization is a known error
// maximizer (Michaud, "The Markowitz Optimization Enigma", FAJ 1989) and why
// a naive equal weighting is hard to beat out of sample (DeMiguel, Garlappi &
// Uppal, RFS 2009). Prefer RiskParity or MinVolatility, which never touch the
// means, when the weights are meant to be held rather than studied; read a
// MaxSharpe fit as a description of the window it was fitted on.
package optimize
