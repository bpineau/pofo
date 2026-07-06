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
//   - No lookahead: Simulate drives each month's return with the most recent
//     regime dated strictly before it, and Regime itself only reads past macro
//     data (the breadth is additionally smoothed, which lags it further).
package permanent
