# FIRE model sensitivity and historical coverage

## What the history badge measures

For H years and M consecutive monthly returns, the historical source contains
max(0, M - 12H + 1) start windows. It samples these windows with replacement.
With M = 12H every draw repeats the same path: increasing the path count adds
no historical evidence. The maximum number of non-overlapping windows is
floor(M / 12H); non-overlap itself does not establish statistical independence.

The simulator therefore labels history as parametric, short or full-horizon.
It does not infer forecast confidence from history length. Short history can
still feed the bootstrap, but resampling cannot manufacture missing regimes.

## Sensitivity experiment

The bundled `replay.Reference` US 60/40 index, deflated by US CPI and rebalanced
annually, supplies monthly real returns from 1953-04-30 to 2025-12-31
(872 observations). Calibration windows of 10, 20, 35 and 60 years all end on
2025-12-31. The full record starts in April; `FitParametric` compounds its
successive twelve-month groups, rather than relabeling them calendar years.

The invented plan starts with 1,000,000 purchasing-power units and spends
40,000 per year over 35 years. It has no tax, fees, pension, buffer or mortality.
Annual withdrawals use the kernel's beginning-of-period convention. Fixed
spending is solved at a conditional 5% ruin target on [0, 100,000], with an
absolute amount tolerance of 5. This is a model diagnostic, not a recommended
withdrawal rate or a calibrated forecast for a euro investor.

Reproduction uses `marketdata.NewPanel(Monthly)` on the reference index,
`fire.FitParametric`, `fire.PrepareModel` and `fire.BuildPlan`. Each condition
uses 6,000 paths with seed 61001 and four workers; the verified bounded solve
is `Plan.SolveOnContext`. The selected amount is evaluated again with
`RuinProbOnContext` on 20,000 paths with independent seed 78001. This checks
sampling stability within the same model, not predictive validity on new
market history. Comparisons reuse seeds, and fixed/adaptive policy comparisons
reuse the exact same `Draws`.

Rounded fixed spending at the 5% conditional target:

| Calibration window | Central | Bootstrap (24-month mean block) | Historical |
|---|---:|---:|---:|
| 10 years | 34,045 | 48,700 | unavailable |
| 20 years | 33,456 | 42,587 | unavailable |
| 35 years | 45,917 | 43,585 | 74,686 |
| 60 years | 36,594 | 34,042 | 39,078 |

The 35-year historical result rests on exactly one start window. Its simulated
ruin probability jumps directly from zero to one as spending increases; there
is no smooth 5% historical quantile to estimate. The 60-year window has 301
overlapping starts but fits only one non-overlapping 35-year window.

## Which assumptions move the answer

- Bootstrap mean blocks of 6, 12, 24, 60 and 120 months move the 10-year
  history's solved spending from 45,013 to 55,756. With 60 years of history,
  the same grid spans 33,829 to 34,705. These are sensitivity ranges, not
  confidence intervals, and a longer block is not uniformly more conservative.
- On the full record, central spending is 39,969. Moving the arithmetic real
  mean by minus/plus one percentage point gives 34,586 / 45,657. Scaling
  volatility by 0.8 / 1.2 gives 44,818 / 35,159. Student-t degrees of freedom
  3, 5 and 30 produce 39,038, 39,038 and 40,436 (rounding to whole units).
- The bundled CAPE observation is 40.36 as of 2026-10-01. Applying the existing
  earnings-yield anchor changes the full-record central result to 25,381.
  This measures the consequence of that assumption; it does not validate
  applying a US equity valuation anchor to a mixed stock/bond portfolio.
- Keeping the full-record central model fixed and repeating seeds 61001
  through 61009 yields solved amounts from 39,969 to 40,982. Their independent
  validation ruin estimates range from 4.56% to 5.55%. The amount-search
  tolerance is far smaller than sampling uncertainty; a sample-feasible
  candidate is not guaranteed to remain below 5% on another draw.

At the same initial 40,000 budget, the full-record central Guyton-Klinger
policy has 0.017% sampled ruin versus 5.07% for fixed spending. However,
43.85% of adaptive paths cut spending at least once and spending's coefficient
of variation is 0.467. Under the CAPE anchor, 90.53% of adaptive paths cut at
least once. Comparing ruin alone discards the policy's lifestyle cost.

These observations justify qualifying the interface and measuring assumptions
separately from sampling noise. They do not justify retuning the six models,
ranking their probabilities as equally plausible forecasts, or optimizing
adaptive capital by a monotonic bisection.
