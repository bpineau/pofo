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
//     that index itself, and for the DBi family that index read through the ten
//     futures contracts the fund actually holds (DBiReplication in
//     dbireplica.go, bundled as refdata by cmd/gen-dbi-refdata: half the
//     composite, half a rolling regression of it on those ten contracts with
//     the intercept discarded, which tracks the fund better than either side). Each donor is volatility-matched to the target and
//     lifted to the target's own fee load (see feeAligned in the recipes); a
//     donor that does not quote daily is projected onto the engine's calendar
//     first, the engine's day-to-day amplitude rescaled (textureScale) so the
//     projected days carry the volatility the donor's own weeks imply. The file
//     starts at the deepest donor and nothing is shipped behind it;
//   - monthlyVolMatch is DonorChain's volatility match for the case a
//     per-observation one cannot serve: a MONTHLY donor and a weekly fund share
//     almost no observation dates, so the ratio must be measured on month-end
//     returns instead. It is what lets the insurance-linked family
//     (catbond.go: ILSFUND, ILSFUNDE and the cat bond share classes behind
//     them) stand on a monthly index; see docs/catbond-sleeve-design.md;
//   - movesOnly is the other half of that projection, for a donor that quotes
//     daily on paper and STALE in fact (a thin listing whose feed reprints the
//     previous close for days on end). The repeated prints are dropped, which
//     turns each flat run back into the gap it really is, and the surviving
//     levels become anchors for the proxy's texture (shapedSeries). No
//     rescaling is applied to such an era: a scale factor stretches a shape,
//     and this one has to be replaced. See chsnRecipe, which measures what the
//     staleness does to daily volatility and what the treatment gives back;
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
//     variance), so long backcasts stay honest at daily-statistics frequency;
//     a shape source is despiked first (a lone provider print whose two huge
//     legs cancel against a calm neighbourhood is dropped, real crash days
//     never qualify), and longBackFee charges a gross proxy what its
//     grossness is worth;
//   - globalbond.go rebuilds a MULTI-CURRENCY government bond futures basket
//     (the sleeve of the global efficient-core fund NTSG) as one excess-return
//     index: a local sleeve per currency, each netted against its OWN
//     money-market rate and carrying no FX at all, weights renormalized over
//     the sleeves that quote so the notional is always full;
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
//     under-risking. And write the gearing as cash + g × (donor − cash), never
//     as g × donor: a longer bond does not earn g times a shorter one's coupon,
//     so plain multiplication pays the file (g − 1) × the cash rate every year,
//     which is 0.9 of a point in the 1960s-80s (dtlaRecipe, corrected 2026-08).
//     The ratio is also not constant, duration shrinking as yields rise: TLT on
//     VUSTX reads 1.243 over 2002-2010 and 1.06 over 2018-2026, so measure it on
//     the whole overlap and not on the era that happens to be observable.
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
//   - A CASH-LIKE PROXY IS GRADED ON ITS LEVEL, NOT ON ITS PATH. A compounded
//     bill rate has almost no daily variance of its own, so its correlation
//     with a fund that carries any spread at all measures the spread, not the
//     reconstruction. ERNA is the case: monthly correlation 0.29 over 96
//     months, and 0.81 once the two months of the March 2020 ultrashort-credit
//     dislocation and its reversal are dropped (dropping any further month
//     changes nothing). Those two months are real, to the day, in both
//     US-listed siblings; they are the fund's investment-grade spread, which
//     the recipe deliberately omits and a bill rate cannot have. Read the level
//     verdict on such a line, and the path verdict as a description of the
//     spread that was left out.
package simgen
