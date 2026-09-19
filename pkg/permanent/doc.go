// Package permanent implements the tactical "Permanent Portfolio 2.0" allocator:
// a regime-driven tilt of Harry Browne's four sleeves (equity, long bonds, cash,
// gold), after Didier Darcet's revisit of the Permanent Portfolio.
//
// The design and the empirical evidence behind it live in
// docs/darcet-permanent-portfolio-design.md, including an epistemic ledger that
// separates what Darcet described in advance from what was reconstructed or
// fitted. Read it before retuning anything: the Params defaults below are a
// reconstruction of rules Darcet does not disclose, not validated optima.
//
// # Layering
//
// The package is pure at the core and impure only at the edge:
//
//   - Panel parses the embedded OECD macro panel (datasets.MacroPanel()) into
//     per-country monthly series. Offline, deterministic.
//   - Regime is the world macro state at a month: the growth and inflation
//     BREADTH (share of countries whose year-on-year rate is accelerating) and
//     the mean monetary slope and real short rate. Panel.Regimes derives it.
//   - Allocation is the four-sleeve target. Regime.Allocate maps a regime to
//     weights with Darcet's quadratic (1/d^2) damping. Pure math.
//   - Simulate backtests a slice of regimes against caller-supplied REAL asset
//     returns. It never fetches: the caller owns the network (deflate the four
//     sleeves to real returns, then hand them over), so the package stays
//     testable offline.
//
// # Conventions and units
//
//   - All returns are REAL (inflation removed) and MONTHLY; annualization uses
//     12 periods (distinct from pkg/metrics, which is daily/nominal). 0.01 = +1%.
//   - Breadth is a fraction in [0,1]; slope and real short rate are in
//     PERCENTAGE POINTS (2.0 = +2 pp). Allocation weights are FRACTIONS summing
//     to 1.
//   - Lookahead: Simulate drives each month's return with the most recent
//     regime dated strictly before it, and Regime itself only reads past macro
//     data (the breadth is additionally smoothed, which lags it further). That
//     is ONE month of lag by REFERENCE date, which is not the same as a lag by
//     PUBLICATION date: month M is allocated on the regime of reference month
//     M-1, whose OECD industrial production lands about forty days after that
//     month ends, i.e. around the tenth of M+1, so the allocation reads a number
//     that did not exist when it was struck. SignalConfig.ReleaseLags closes
//     that gap, PublicationLags being the honest setting (production two months
//     further back, prices one, market rates not at all). It is NOT the default:
//     Regime is also read descriptively, to say what the world's macro state was
//     in a past month (the report's regime strip), and there the reference-date
//     reading is the right one. A backtest should pass PublicationLags; it costs
//     about a third of the tactical edge over the static portfolio (globally
//     +1.29 to +0.80 points a year, and a deeper worst drawdown), measured in
//     docs/darcet-permanent-portfolio-design.md, which reports the whole battery
//     at both settings.
package permanent
