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
//     lookback, vol target, leverage) for replicating trend strategies: the
//     signal is refreshed every Rebalance days, risk is rescaled every day
//     against an exponentially weighted covariance (CovHalfLife);
//   - AnchorTrend puts a trend reconstruction's month-to-month path back on a
//     bundled record while the engine keeps supplying the daily texture, which
//     lifts the monthly agreement with the real funds from ~0.4 to ~0.7. Three
//     such records are bundled: the net managed-futures composite
//     (NetTrendAnchor) for a diversified fund, the net pure-trend composite
//     (PureTrendAnchor) for a trend overlay, and the gross academic factor
//     (GrossTrendAnchor), kept as a shape yardstick. The two net ones settle
//     the level as well as the path, and a build that takes its level from a
//     reference stops where that reference does (AnchorStart). See
//     docs/trend-reconstruction-design.md;
//   - DonorChain assembles a young fund's past out of REAL records of the same
//     trade instead of a reconstruction, nearest first: another manager's fund
//     NAVs, or, for a fund whose whole programme replicates a published index,
//     that index itself. Each donor is volatility-matched to the target and
//     lifted to the target's own fee load (see feeAligned in the recipes); a
//     donor that does not quote daily is projected onto the engine's calendar
//     first. The file starts at the deepest donor and nothing is shipped behind
//     it;
//   - financed serves the USD overnight financing rate (usdOvernight: ^SOFR
//     from 2018-04, effective fed funds from 1954, the T-bill rate before), the
//     rate a FUTURES-based leg pays. Two rates run through these
//     reconstructions and they are not interchangeable: an overlay finances at
//     the overnight rate, a collateral sleeve earns the bill rate, and the two
//     have differed by 0.02 to 1.15 points a year depending on the decade;
//   - Audit / AuditAll replay a recipe's engine WITHOUT the real quotes it
//     splices in and grade it against them over their overlap: two verdicts
//     (level, does it earn the return; path, does it move with the asset),
//     the donor chain junction by junction, and the curves to plot. This is
//     what "pofo -verify-simdata" renders;
//   - Validate measures daily and weekly correlation, beta, tracking error
//     and CAGR against the real series; WithRefData serves the bundled
//     reference series (datasets.Refdata, e.g. MSCIWORLD-USD, SP500-USD) and
//     any extra local CSVs (dev -refdata) before the network;
//   - extend/longBack splice a long real proxy behind a short component leg
//     (VFINX→S&P 500 ~1871, VTMGX→MSCI World ex-US ~1969, VEIEX→MSCI EM ~1988,
//     VFITX/VUSTX→constant-maturity Treasury TR ~1953, GC=F→LBMA gold ~1968,
//     CL=F→WTI ~1946, ^IRX→3-month T-bill ~1934, GBPUSD=X→FRED daily ~1971,
//     DFSVX→Ken French small value ~1963), so a multi-leg reconstruction
//     reaches back to its youngest leg's first quote (BuildFrame's start);
//     dailyShape then blends a real daily series of the same market into a
//     monthly proxy (anchors keep the levels, the shape supplies the day-to-day
//     variance), so long backcasts stay honest at daily-statistics frequency,
//     and longBackFee charges a gross proxy what its grossness is worth;
//   - the bundled recipes (All, Find) assemble these building blocks for
//     NTSX, NTSG, URTH, IWDA, VT, RSSB, GDE, XAUUSD, ZPRV, the Avantis Global
//     Small Cap Value ETF, SHY, IEF, TLT, ZROZ, DBMF, DBMFE, KMLM, the AQR
//     Managed Futures UCITS fund, CTA and the Winton Trend-Equity fund, among
//     others.
//
// # Units
//
// Beware of the unit conventions: fees passed to Composite and TSMOMConfig
// are FRACTIONS per year (0.0085 = 0.85 %/yr), as are
// volatility targets (0.10 = 10 %), whereas the portfolio package and
// marketdata.Fees express fees in PERCENT per year (0.85 = 0.85 %/yr).
// Rate series (^IRX, ^FVX, ^TNX, ^TYX, the policy and money-market family of
// marketdata.RateSymbols, and usdOvernight) are annualized percent levels and
// are converted to daily accruals by BuildFrame.
//
// # Four rules that keep coming back
//
// Every recipe here is its own measurement, but four corrections have now
// been made often enough to state once (each recipe's comment carries its own
// numbers):
//
//   - A DONOR OF THE WRONG DURATION is geared by the ratio of its volatility
//     to the target's, measured on their whole overlap and at MONTHLY
//     frequency when the target's daily quotes carry exchange noise. For two
//     bond series on the same curve, what separates them is duration and
//     volatility is proportional to it, so the ratio is the duration ratio:
//     TLT on VUSTX measures 1.131 where the shipped gearing read off a
//     duration table is 1.133. Never gear to minimize tracking error instead
//     (it is always lower, because with correlation under one it pays to
//     under-risk), and never read the beta of a validation line as a target:
//     beta is correlation times the volatility ratio, so it too rewards
//     under-risking.
//   - AN FX-HEDGED SHARE CLASS earns the domestic cash rate on its WHOLE
//     capital, not on the fraction its duration weight leaves over: the hedge
//     covers the position, not the leftover. A hedged leg is
//     w × (local − foreign cash) + 1.00 × domestic cash.
//   - A STAND-IN CARRIES ITS OWN PRICE LIST, and the difference belongs to the
//     wrapper rather than to the trade. A donor fund is lifted or charged the
//     difference in published ongoing charges, read off fee tables and never off
//     an observed return gap (a gap between two managers contains their skill).
//     A target's WHOLE ongoing charge on top of a fund donor bills the backcast
//     twice, since a NAV already arrives net of its own manager's cut: what is
//     due is the difference, floored at zero (feeGap, alignedComposite). It is a
//     schedule and not a constant, because a leg extended by longBack rides a
//     fee-free index or CMT reconstruction before the donor fund's own quotes,
//     and over that era the whole charge IS due.
//     An ACADEMIC FACTOR has no price list at all: it pays no fee, no commission
//     and no spread, so it owes a haircut measured on its overlap with the fund
//     it stands in for, and the smaller the stocks the larger that is
//     (longBackFee: 1.0 %/yr for the small-value factor behind DFSVX, measured
//     over 399 common months).
//   - A RATE IS NOT INTERCHANGEABLE WITH ANOTHER RATE of a different tenor or
//     security. A futures overlay finances overnight (usdOvernight) and a
//     collateral sleeve earns the bill rate; a fund tracking an overnight
//     accrual is not rebuilt from a 3-month interbank index (0.16 points a
//     year apart over 1999-2025). Splicing rate levels needs no rescaling,
//     only the right order (ESTR before EONIA, which was defined as ESTR plus
//     8.5 bp).
package simgen
