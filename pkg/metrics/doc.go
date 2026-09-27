// Package metrics computes risk and return statistics for a series of
// dated values: CAGR, volatility, Sharpe, Sortino, Ulcer Index, Max
// Drawdown, TTR (time to recovery), Beta against a benchmark, and the
// native-versus-monthly volatility term structure (the Lo-MacKinlay
// variance ratio); and over bare return samples, the correlation and
// covariance matrices, multiple regression, dated extremes, tracking error
// and the lead-lag gaps of two closes struck at different hours.
//
// It is the math of the toolkit and imports nothing of it: every function
// takes plain slices (parallel dates and values, or bare returns), so it
// reads a series built anywhere. pkg/marketdata hands its series over
// (Series.Stats, Panel.Col).
//
// # Start here
//
//   - [Compute] turns a dated value series into [Stats]: CAGR, volatility,
//     Sharpe, Sortino, Ulcer, max drawdown, time to recovery, skew and
//     kurtosis, annualized at the series' own cadence ([PeriodsPerYear]).
//   - [Returns] and [Mean] are the building blocks; [CalendarReturns] the
//     yearly or monthly table; [Drawdowns] the running drawdown, and
//     [DrawdownEpisodes] and [MaxDrawdown] the dated episodes ([Episode]).
//   - [Corr], [CorrelationMatrix] and [Covariance] relate return columns
//     measured on one calendar; [Regress] fits one on several ([Regression]);
//     [Track] measures a replica against what it follows ([Tracking]).
//   - [LowestK] and [HighestK] date a sample's extremes; [VaR] and [CVaR]
//     read its tail.
//   - [TWR] and [IRR] measure a series that carries external flows ([Flow]).
//
// Every return and statistic is a FRACTION (0.07 = +7 %), except
// [Stats].Ulcer in percent points and [Stats].CWARP in percent.
//
// # Conventions
//
// Knowing the conventions is essential to compare the results with other
// tools:
//
//   - annualization follows the data's cadence: volatility and ratios are
//     annualized at PeriodsPerYear of the series' dates, the observations it
//     holds per year snapped to a canonical count (252 on trading-day closes,
//     so every daily figure is the classic one; 52 on a weekly NAV; 12 on a
//     monthly index; 365 on a calendar that quotes weekends). Stats reports
//     the count it used. Functions that receive dates infer it; those that
//     receive bare returns (Volatility, Sharpe, Sortino, ReturnToMaxDrawdown,
//     WorstRollingReturn, CWARP) take it as an argument, TradingDaysPerYear
//     for daily returns. A series that changes cadence along the way is read
//     at its prevailing one: resample it first when that matters;
//   - Sharpe and Sortino use a zero risk-free rate (like Curvo);
//     PortfolioVisualizer and LazyPortfolioETF use T-bills and monthly
//     data; their Sharpe ratios come out ≈ 0.10-0.15 lower;
//   - Max Drawdown, Ulcer and TTR are measured on daily closes, harsher
//     than monthly-step tools (COVID 2020: −33.7 % on daily closes
//     versus ≈ −20 % on monthly closes);
//   - the CAGR uses 365.25-day years between the first and the last
//     date.
//
// The main entry point is Compute. Beta pairs returns with the benchmark's by
// date. VarianceRatio resamples a series sampled more often than monthly to
// month-end closes and reports the volatility at both frequencies plus their
// ratio, revealing the autocorrelation the daily statistics hide. Returns and
// Mean are exposed as building blocks.
//
// # CWARP
//
// CWARP (Cole Wins Above Replacement Portfolio, Artemis Capital Management)
// scores whether an asset improves a pre-existing "replacement" portfolio
// when layered on top at a fixed weight, financed by borrowing. It is the
// geometric average of the improvements it makes to the replacement's
// Sortino ratio and its return-to-maximum-drawdown, minus one, in percent:
// positive means the overlay lifts the portfolio, negative means it hurts.
// Because both denominators are measured on the combined series, CWARP
// rewards non-correlation and skew that the Sharpe ratio ignores. The
// replacement is typically equity beta or a 60/40 blend. Sortino and
// ReturnToMaxDrawdown, its two building blocks, are exported on their own:
// they are the downside-aware quantities pkg/optimize maximizes for the
// max-sortino and return-to-drawdown objectives. Ulcer (root-mean-square
// drawdown) and WorstRollingReturn (worst outcome over a fixed window) round
// out the underwater-robustness measures, behind the min-ulcer and
// max-worst-5y objectives that matter most in decumulation. The
// black-litterman objective needs none of them: it works on the mean vector
// and the covariance matrix alone.
//
// # Matrices, calendar table and tails
//
// Corr is the one Pearson correlation of the tree (suggest.Correlation
// delegates to it). CorrelationMatrix and Covariance take [asset][period]
// returns on ONE calendar, as marketdata.Aligned.Returns produces them, and
// return an error on ragged rows; Covariance is per period (multiply by the
// calendar's PeriodsPerYear, 252 for a daily one, to annualize it).
// CalendarReturns cuts a value series into calendar
// months, quarters or years (blocks of months counted from January, as
// marketdata.Series.Resample cuts them), each period chained on the previous
// period's last close and the first one flagged Partial: the table behind an
// "annual returns" row or a monthly heatmap, and the one the golden tests
// read published calendar-year returns against. RollingBeta and RollingCorr
// are Beta and Corr over RollingVol's trailing windows, on the returns Beta
// pairs by date. VaR and CVaR read a return sample's tail historically, off
// Quantiles, as POSITIVE per-period losses (0.95 = the loss exceeded in 5 %
// of periods), never annualized. Everything here takes and returns
// fractions.
//
// # Regression and dated extremes
//
// Regress is ordinary least squares with an intercept, on one regressor or
// several: a fund on its index and a currency, a strategy on two factors.
// It centers the data and factorizes the regressors by Householder QR, never
// forming X'X, and the golden package holds it to the certified values of
// the NIST StRD Longley benchmark. Its Regression is PER PERIOD (alpha a
// per-period return, ResidualSD a per-period deviation); AnnualAlpha and
// AnnualResidualVol annualize them at a cadence the caller passes, the
// PeriodsPerYear of the data. Each coefficient carries its standard error
// and t-statistic, classical ones that trust independent residuals.
//
// LowestK and HighestK return the POSITIONS of a sample's extremes, stable
// on ties and blind to NaN, so the worst months of a return column can be
// dated (a marketdata.Panel's Ends) and the other columns read on exactly
// those periods (Panel.Pick): the conditional statistic "what did the hedge
// do in the equity's worst decile". TopK stays the values-only fast path of
// a tail statistic over a large sample.
//
// # Tracking a reference
//
// TrackingError is the annualized volatility of one return column minus
// another, the risk of holding a replica, a share class or a
// reconstruction instead of what it follows. LeadLagGaps is the period by
// period disagreement between two such columns once a one-session
// difference in closing times is forgiven (a Xetra line against an index
// struck after New York): the test that tells a bad print from a time zone,
// behind both the donor repairs of pkg/simgen and the dated divergences of
// analyze.Pair.
//
// Track gathers the whole comparison of two return columns in one Tracking
// (correlation, the two volatilities and their ratio, tracking error,
// tracking difference, beta and Jensen's alpha), each figure the one its
// single-purpose function computes, and DifferenceSE says how many
// standard errors a return gap over the window is worth. analyze.Pair's
// blocks and pkg/simgen's reconstruction audit both read it, through
// marketdata.Panel.Track, so a backcast is graded by one set of formulas
// whichever tool reports it.
//
// # Attribution
//
// Attribute splits a portfolio's risk and realized return across its holdings,
// from the per-holding contribution series a simulation produces
// (portfolio.SimResult.Contributions). The return share is a plain sum; the
// risk share is the Euler decomposition of variance, Cov(c_i, r_p)/Var(r_p),
// so the shares sum to one without normalization and a holding that moves
// against the book can legitimately take a negative one. Reading the two
// together is the point: a sleeve's share of risk routinely differs from its
// share of capital by a factor of three, and an insurance sleeve is meant to
// show a small return share against a real risk share.
//
// # External flows
//
// When a series carries external contributions and withdrawals (a savings
// account, a wealth tracker), Compute's raw figures would mistake them for
// performance. TWR chain-links flow-neutralized daily returns, FlowReturns
// yields the flow-adjusted returns and their cadence (weekend points dropped,
// so calendar-daily forward-filled series keep an honest volatility and read
// as trading days), and Volatility, Sharpe and Sortino accept those returns
// with an explicit annual risk-free rate (Compute's convention stays rf = 0). Annualize
// turns a cumulative return over a day span into a compound annual rate,
// and IRR solves the money-weighted rate of the flows themselves.
//
// # Distribution shape
//
// Skewness, ExcessKurtosis, Autocorr and Histogram describe a sample's shape.
// Quantiles reads it at chosen probabilities and TopK returns its largest
// values, the raw material of a tail statistic (a CVaR, a conditional
// drawdown). Both place only the order statistics they need, by partial
// selection rather than by sorting the whole sample: reading five percentiles
// of thousands of values, which a decumulation wealth fan does once per year
// of its horizon, costs a fraction of the sort. The values are the order
// statistics either way, so the results are exactly those of a full sort.
//
// These computations are locked down by the golden package's benchmark
// tests, which check them against external references (official S&P 500
// TR annual returns, canonical drawdowns).
package metrics
