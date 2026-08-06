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
//   - Validate measures daily and weekly correlation, beta, tracking error
//     and CAGR against the real series; WithRefData optionally serves extra
//     local reference CSVs (dev -refdata) before any network source;
//   - the bundled recipes (All, Find) assemble these building blocks, from
//     fetchable quotes only, for NTSX, NTSG, URTH, IWDA, VT, RSSB, ZPRV,
//     SHY, IEF, TLT, ZROZ, DBMF, KMLM, CTA and the Winton Trend-Equity fund.
//
// # Units
//
// Beware of the unit conventions: fees passed to Composite and TSMOMConfig
// are FRACTIONS per year (0.0085 = 0.85 %/yr), as are
// volatility targets (0.10 = 10 %), whereas the portfolio package and
// marketdata.Fees express fees in PERCENT per year (0.85 = 0.85 %/yr).
// Rate series (^IRX, ^FVX, ^TNX, ^TYX) are annualized percent levels and
// are converted to daily accruals by BuildFrame.
package simgen
