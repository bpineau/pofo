// Package optimize computes portfolio weights that optimize a risk/return
// objective from the historical returns of the candidate assets.
//
// Ten objectives are supported:
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
//   - MaxReturn ("max-return"): maximizes the portfolio's own CAGR (the
//     geometric return of the blended path). Alone it is degenerate, since
//     the whole budget goes to the single best-performing asset; it earns
//     its keep under a Limit, where it is the frontier point: the most
//     return reachable inside a volatility or drawdown budget.
//   - BlackLitterman ("black-litterman"): the weights that maximize the
//     mean-variance utility of the returns the portfolio's own weights
//     imply, updated by the owner's views. It is the one objective that
//     does not read its expected returns off the sample; see below.
//   - CWARP ("cwarp"): the blend that best improves a replacement portfolio
//     (a benchmark) when overlaid on it. It is solved by SolveCWARP, which
//     takes the replacement's returns as an extra argument; the objective is
//     non-convex and non-smooth (it depends on the combined drawdown), so the
//     solver is a multi-start heuristic and its weights are a good allocation
//     rather than a certified optimum.
//
// Weights are long-only (no short selling) and sum to 1. Concentration is
// bounded for every objective except RiskParity (whose weights follow
// directly from the equal-risk condition) by, in increasing order of
// precision: MaxWeight, a cap on every asset; MinWeight, a floor on every
// asset, which keeps each line in the book instead of letting the search
// drop the ones that only pay off out of sample; and Bounds, a range per
// named asset. Bounds arrive keyed by identifier and are turned into the
// per-position Lower/Upper the solver reads by Spec.Resolve, which the
// CALLER runs: resolving identifiers is pkg/marketdata's job. CWARP is the
// exception on the box too: SolveCWARP caps weights (MaxWeight) but knows
// nothing of MinWeight or Bounds, which ParseSpec therefore refuses.
//
// Bounds matter more than they look. An unconstrained optimum is a corner
// solution, and a corner fitted on one window is the least durable thing an
// optimizer produces. The ranges a portfolio's owner can defend line by line,
// for reasons the backtest cannot see, belong in the problem, not in a
// footnote to its answer.
//
// # Constraints, and why they beat a scalar objective
//
// Limits express the question most portfolio work actually asks: not "the
// best Sharpe" but "the most return without going above X volatility", or
// "the least volatility that still returns Y". Three are available
// (MaxVolatility, MinReturn, MaxDrawdown) and they compose:
//
//	optimize.Spec{Objective: optimize.MaxReturn,
//	    Limits: optimize.Limits{MaxVolatility: 0.095}}
//
// They apply to every objective but RiskParity and CWARP (BlackLitterman
// accepts every constraint those two refuse): those two are
// solved by code that never sees Limits (the equal-risk condition in closed
// form, and SolveCWARP's own search, which honours MaxWeight alone), so
// ParseSpec refuses to combine them with a limit, and CWARP with a floor or
// a per-line range, instead of dropping the constraint in silence.
//
// Any limit (or any per-asset bound) routes the solve through one penalized
// path search, whatever the objective; without them the closed forms answer
// exactly as they always did. Result.Feasible reports whether the answer
// actually meets the limits: when nothing can, the weights returned are the
// least-violating point found, and saying so is the point of the flag.
//
// Reaching for a constraint rather than a ratio also sidesteps a trap of this
// toolkit's conventions: Sharpe here runs at a ZERO risk-free rate, so any
// cash-like sleeve buys ratio for free. A euro-linker sleeve can raise a
// portfolio's measured Sharpe while cutting two points of CAGR. A volatility
// cap asks the question that was meant.
//
// # Black-Litterman: the file's weights as the prior, views as the input
//
// Every objective above that touches expected returns reads them off the
// sample, which estimates them worst (see "Estimation error" below), so the
// answer is a corner driven by whichever line was lucky over the window.
// BlackLitterman keeps the means in the problem but anchors them.
//
// The PRIOR is the portfolio file itself: Spec.Prior holds the weights the
// owner already defends, filled by the caller in asset order, and Solve
// refuses anything that is not an allocation. There is no market-
// capitalization anchor here, because the books this toolkit serves are made
// of lines that have none (a managed-futures fund, a stacked 90/60, an
// inflation-linked sleeve, gold). Reverse optimization then answers a
// question an owner rarely asks, "what return does my allocation implicitly
// expect from each line": ImpliedReturns returns pi = lambda*Sigma*w.
//
// The VIEWS are the input. Each is a belief about one line ("gold earns 2 %
// a year") or about a pair ("trend beats long duration by 3 points"),
// carrying a Confidence in (0,1). Posterior blends them into the implied
// returns by the closed form
//
//	Omega = diag((1/c_k - 1) * (P tau Sigma P')_kk)
//	mu    = pi + tau*Sigma*P' (P tau Sigma P' + Omega)^-1 (Q - P pi)
//
// where a confidence of 0.5 is He and Litterman's own Omega, near 0 switches
// a view off and near 1 makes it certain. The scalar tau is fixed at 0.05 and
// exposed nowhere because Omega is proportional to it, so it cancels exactly.
// The posterior COVARIANCE is not used: Sigma stays the risk model, the
// common practitioner choice. Weights then maximize mu'w - (lambda/2)w'Sigma w
// over the same box simplex every other objective uses, so bounds, floors and
// the feasibility limits all compose.
//
// The identity that makes the objective legible: with NO view, mu = pi and
// the long-only optimum is EXACTLY the prior, since the utility's gradient
// vanishes there. So "black-litterman" on a file with no view returns the
// file's own weights and reports what returns they imply, which is a reading
// worth having on its own. Maximizing the Sharpe ratio of the posterior would
// not have that property, which is why the utility form was chosen.
//
// The RISK AVERSION sets the scale of pi, and absolute views are read against
// that scale, so it matters. Spec.PriorReturn ("prior-return:4.6") states the
// return the owner expects from the prior allocation as a whole and fixes
// lambda at that return over the prior's variance; without it, the prior is
// assumed to earn DefaultPriorSharpe. He and Litterman's delta = 2.5 is not
// the default, and DefaultPriorSharpe says why.
//
// Two traps. The zero risk-free rate of this toolkit reaches here too: a 3 %
// view on a 3 %-volatility sleeve is a Sharpe of 1.0 in the utility's eyes
// and the weight follows, so state views as excess returns over cash when
// that is what is meant. And what this model prices is the MEAN: a line held
// for its crisis covariance or against a liability has a reason the model
// cannot read, and will be sized by its view alone. That is a feature, since
// it shows what the hedge costs in expected return, as long as the reader
// knows it.
//
// # Fitting on one window, judging on another
//
// Spec.Train carries the "train:" window parsed from a portfolio file. This
// package IGNORES IT: Solve is date-free, and slicing the returns is the
// caller's job (pkg/compare fits on the slice, then measures the resulting
// weights over the whole window, so what the report shows is out-of-sample).
// It travels in Spec because it arrives in the same directive, and one
// parser beats two.
//
// # How each objective is solved, and how much to trust it
//
// MaxSharpe, MinVolatility, RiskParity and BlackLitterman read only the mean
// vector and the covariance matrix; MaxSortino, ReturnToDrawdown, MinUlcer,
// MaxWorst5y and CWARP depend on the whole return path. All of them are minimized by
// projected gradient descent over the capped simplex, from several
// deterministic starting points. They differ in what that buys:
//
//   - MinVolatility, RiskParity and BlackLitterman are convex programs, so
//     their answer is the global optimum.
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
// MaxSharpe fit as a description of the window it was fitted on. Or state the
// means instead of estimating them: BlackLitterman is the objective for an
// owner who can say what they believe and how sure they are.
package optimize
