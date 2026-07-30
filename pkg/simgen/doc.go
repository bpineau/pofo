// Package simgen rebuilds the missing past of complex assets (90/60 funds,
// managed futures, …) and validates each reconstruction against real
// quotes. Results are stored as permanent "simdata" files that portfodor
// splices in front of the real histories.
//
// # Toolbox
//
//   - BuildFrame aligns the daily returns of several components (rate
//     series such as ^IRX are converted to accrual);
//   - Composite builds a base-100 index from constant weights, including
//     "excess" legs (futures) and annual fees;
//   - TSMOM is a configurable time-series momentum engine (markets,
//     lookback, vol target, leverage) for replicating trend strategies;
//   - FitBackcast regresses an asset on factors and replays the model
//     over the whole history (rejected under an R² floor: ErrUnfaithful);
//   - WithRefData serves local reference series (datasets/refdata/)
//     before
//     any network source; Validate measures daily and weekly correlation,
//     beta, tracking error and CAGR against the real series;
//   - the bundled recipes (All, Find) assemble these building blocks for
//     NTSX, NTSG, URTH, IWDA, ZROZ, IEF, TLT, XAUUSD, DBMF, KMLM, CTA and
//     the Winton Trend-Equity fund.
package simgen
