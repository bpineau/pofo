# FIRE rolling holdout diagnostic

## Protocol

`examples/code/fireholdout.go` evaluates the current FIRE models on the bundled
real US 60/40 reference. It is a retrospective diagnostic of conditional
model distributions, not a validated forecast or a recommended withdrawal
rate. The reference uses S&P 500 and 5-year Treasuries, annual rebalancing and
US CPI. The current bundle ends in December 2025; the first complete calendar
year is 1954.

For each January origin, fit on the immediately preceding T complete calendar
years and evaluate against the following H complete calendar years. No test
return enters `FitParametric` or `PrepareModel`. Short-history blending uses
the current fixed priors; those conventions and the revised/reconstructed
reference are not historical data vintages. This is therefore not an
as-of-date forecasting backtest of what an investor could have known then.

Three models are evaluated:

- **Central:** fit annual arithmetic mean, annualized monthly volatility and
  tail degrees of freedom on the training window; retain the current
  horizon-dependent short-history blend.
- **Bootstrap:** the current stationary bootstrap, mean block 24 months,
  from training months only; compound each 12-month group.
- **Historical:** enumerate every complete H-year monthly cohort within the
  training window, with equal weights. Unavailable when T < H. With T = H,
  the entire distribution is one path, regardless of a Monte Carlo path
  count. With T = 35 and H = 20, there are 181 overlapping training cohorts.

CAPE and broad_sample are excluded because their current inputs are not
vintage-cut. Stress and lost_decade are scenario assumptions, outside this
three-model comparison. No model parameter, spending level or block size is
chosen against the held-out answers.

Initial wealth is 1 purchasing-power unit. Each origin/model is run both
with no withdrawals (`growth`) and constant beginning-of-year withdrawals
of 0.04 (`fixed`). There are no taxes, fees, pension, buffer or mortality.
The two plans reuse exactly the same return draws. Observed paths use the
same decumulation kernel as simulations. Ruin means an unfunded withdrawal,
not merely a low terminal balance.

Outputs include all train/test dates, raw fits, effective central parameters,
per-origin seeds, training cohort counts, realized terminal wealth and ruin,
P5/P50/P95 terminal wealth, predicted ruin and aggregate diagnostics. Tail
counts use strict inequalities; ties are inside. At a mass of zero wealth,
nominal quantiles do not imply exact 5% tails. The no-withdrawal distribution
helps distinguish return-model misses from the effects of withdrawals.

The Brier score is the average squared difference between predicted ruin
probability and the observed zero/one ruin indicator. It is descriptive here:
origins overlap and outcomes are dependent. A score of zero on a window
without any failures does not validate a model's tail risk.

## Reproduction and checks

From the module root:

```sh
go run examples/code/fireholdout.go -train 20 -years 20 -paths 6000 > /tmp/holdout-20-20.json
go run examples/code/fireholdout.go -train 10 -years 35 -paths 6000 > /tmp/holdout-10-35.json
go run examples/code/fireholdout.go -train 20 -years 35 -paths 6000 > /tmp/holdout-20-35.json
go run examples/code/fireholdout.go -train 35 -years 20 -paths 6000 > /tmp/holdout-35-20.json
go run examples/code/fireholdout.go -train 35 -years 35 -paths 6000 > /tmp/holdout-35-35.json
go run examples/code/fireholdout.go -train 10 -years 10 -paths 6000 > /tmp/holdout-10-10.json
```

The base seed is 61001; each origin adds its calendar year. Four workers
are the default. Historical results enumerate cohorts exactly; they have
no simulation error. `-stride H` selects disjoint test windows beginning
at the first eligible origin. The program always uses bundled data offline.

`TestFIREHoldoutExample` runs the example with different worker counts and
requires byte-identical JSON. It checks complete, disjoint train/test dates,
recomputes training fits from independent date selection, compares realized
paths against the reference's separate calendar-year API, and verifies that
the sole historical prediction when T = H is the training cohort. The test
checks contracts rather than frozen historical amounts.

## Recorded results

These runs use 6,000 draws for central/bootstrap, seed 61001 plus origin year,
and the bundle through 2025. Counts below describe the no-withdrawal terminal
wealth **outside** the model's [P5,P95] interval; they are counts of overlapping
origins, not independent observations.

| Train / horizon | January origins | Central | Bootstrap | Historical |
|---|---|---:|---:|---:|
| 10 / 10 years | 1964-2016, n=53 | 27/53 | 28/53 | 53/53 |
| 20 / 20 years | 1974-2006, n=33 | 18/33 | 17/33 | 33/33 |
| 10 / 35 years | 1964-1991, n=28 | 9/28 | 17/28 | unavailable |
| 20 / 35 years | 1974-1991, n=18 | 8/18 | 11/18 | unavailable |
| 35 / 20 years | 1989-2006, n=18 | 0/18 | 0/18 | 7/18 |
| 35 / 35 years | 1989-1991, n=3 | 0/3 | 0/3 | 3/3 |

At 20/20 years, central has 6 lower-tail and 12 upper-tail misses; bootstrap
has 5 and 12. Withdrawing 4% gives 18/33 and 16/33 misses respectively.
Different training lengths change the eligible origins, so comparing these
rows does not isolate the effect of training length. In particular, the
35/35 experiment has only three origins and cannot validate a 35-year tail.

Fixed 4% spending, mean predicted ruin over origins versus realized ruin:

| Train / horizon | Realized failures | Central mean probability | Bootstrap mean probability | Historical mean probability |
|---|---:|---:|---:|---:|
| 10 / 35 | 4/28 (14.29%) | 32.19% | 39.47% | unavailable |
| 20 / 35 | 0/18 | 35.85% | 48.58% | unavailable |
| 35 / 35 | 0/3 | 12.12% | 18.13% | 0% |
| 20 / 20 | 0/33 | 2.40% | 2.94% | 0% |
| 35 / 20 | 0/18 | 0.49% | 0.99% | 0% |

Aggregate pessimism can coexist with missing the particular origins that
fail. For 10/35, the four realized failures begin in 1965, 1966, 1968 and
1969. Their central probabilities are 5.18%, 8.60%, 6.10% and 14.15%; bootstrap
assigns 0%, 0.017%, 0% and 0.067%. A zero simulation count is not a proof of
zero probability. The bootstrap training windows preceding those failures
contain no equivalent of the ensuing sequence.

For 10/35 fixed spending, central/bootstrap Brier scores are 0.2800/0.4778.
For 20/20 they are 0.00166/0.00188, with no realized failures. These scores
are not an overall model ranking; horizons, eligible eras and event rates
differ, and there is only one US historical path.

## Sampling and overlap sensitivity

Repeating 20/20 with 20,000 draws and independent base seed 78001 moves mean
predicted ruin to 2.386%/2.914% for central/bootstrap. Central fixed-spending
lower-tail misses move from 6 to 7; bootstrap stays at 5. For 10/35, the
repeat gives 32.11%/39.50% mean predicted ruin and Brier 0.2799/0.4782.
The broad discrepancies persist under this simulation-noise check; individual
quantile crossings can still move.

For 20/20 with `-stride 20`, only origins 1974 and 1994 remain. Both realized
paths fall inside the central/bootstrap P5-P95 intervals, for both plans,
and neither ruins. Mean predicted fixed-spending ruin falls to 0.258%/0.100%.
This does not refute the annual-origin result: removing overlap leaves only
two selected histories. Non-overlap itself does not establish independence,
and the origin phase matters. No binomial confidence interval or forecast
calibration significance test is claimed.

## Implications

The current conditional distributions do not earn a claim of historically
validated 90% predictive coverage from this exercise. Short samples can miss
future regimes; resampling more paths does not supply those regimes. Historical
with T = H is especially unsuitable for interpreting its P5/P95 as forecast
uncertainty. Central/Bootstrap's apparent success at 35/35 rests on three
heavily overlapping origins.

These findings support retaining the history caveats, exposing assumptions,
and separating model uncertainty from Monte Carlo noise. They do not justify
retuning on these same held-out observations or choosing a withdrawal rate
from the most favorable model. Broader geographies and genuinely dated input
vintages would be needed for stronger external validation. The existing
[sensitivity record](fire-model-sensitivity.md) answers a different question:
how inputs move a model's answer, without testing later realized history.
