// Package simgen rebuilds the missing past of complex assets (90/60 funds,
// managed futures, …) and validates each reconstruction against real
// quotes. Results are stored as permanent "simdata" files that pofo
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
//     and CAGR against the real series; WithRefData serves the bundled
//     reference series (datasets.Refdata, e.g. MSCIWORLD-USD, SP500-USD) and
//     any extra local CSVs (dev -refdata) before the network;
//   - extend/longBack splice a long real proxy behind a short component leg
//     (VFINX→S&P 500 ~1871, VTMGX→MSCI World ex-US ~1969, VEIEX→MSCI EM ~1988,
//     VFITX/VUSTX→constant-maturity Treasury TR ~1953, GC=F→LBMA gold ~1968,
//     CL=F→WTI ~1946, ^IRX→3-month T-bill ~1934, GBPUSD=X→FRED daily ~1971),
//     so a multi-leg reconstruction reaches back to its youngest leg's first
//     quote (BuildFrame's start); dailyShape then blends a real daily series
//     of the same market into a monthly proxy (anchors keep the levels, the
//     shape supplies the day-to-day variance), so long backcasts stay honest
//     at daily-statistics frequency;
//   - the bundled recipes (All, Find) assemble these building blocks for
//     NTSX, NTSG, URTH, IWDA, VT, RSSB, GDE, XAUUSD, ZPRV, SHY, IEF, TLT,
//     ZROZ, DBMF, DBMFE, KMLM, CTA and the Winton Trend-Equity fund, among
//     others.
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
