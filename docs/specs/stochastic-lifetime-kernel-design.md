# Stochastic-lifetime kernel: design

Status: shipped 2026-08-22 (`pkg/decumul`, opt-in via `Plan.Lifetime`).

## 1. What was missing

Until now every simulated decumulation path ran a FIXED horizon
(`Plan.Years`) and mortality entered only afterwards, as a posterior
weighting: `Ensemble.LifeCurve(surv)` crosses the ruin timing with a survival
curve to produce the alive / broke / dead decomposition, and the web layer
weights each ruined path by `CoupleSurvival(age, RuinYear)` to get a
"ever alive and broke" figure.

That weighting is exact for one question and structurally unable to answer
three others.

- It IS exact for alive-ruin **when mortality does not change the path**. If
  the household spends the same whether one or two members are alive, and
  every income keeps paying to the end, then a path's trajectory is
  independent of the draw of death, and
  `P(broke while alive) = E_paths[S(RuinYear)]` exactly. The kernel and the
  weighting must agree there, and a test pins that (§7).
- It CANNOT price an annuity. An annuity insures the risk of outliving the
  plan, and that risk does not exist inside a fixed horizon. A sweep run on
  2026-07-29 over ages 60 to 75 and rates 3.5 to 5.5 % could not reproduce the
  documented trade-off (annuitising raises headline ruin while improving the
  worst late-life outcomes): both readings always moved together, because
  under a fixed horizon the annuity is simply a cashflow bought at a fee. The
  figure that would have shown the trade-off was dropped for that reason
  (the 2026-07 figure backlog, since deleted), with "would need a
  stochastic-lifetime kernel" as the recorded verdict. This is that kernel.
- It has no **estate**. Terminal wealth at a fixed horizon is the wealth of a
  household that is, in the model, still alive at exactly that year. What a
  plan actually leaves is the wealth at the drawn death, which is a different
  distribution: earlier deaths leave more.
- It has no **couple dynamics**. A survivor spends less than a couple and
  loses part of the household's pensions. Both effects are first-order for a
  French household (the reversion rate is a large fraction of a pension) and
  neither survives a posterior weighting.

The chantier is therefore a library-completeness feature: a decumulation
engine that draws the lifetime inside each path is the actuarially right
product, and the fixed horizon becomes a special case of it.

## 2. The one decision that shapes everything

**The kernel draws the death; the household never sees it.**

Every spending rule keeps planning over `PlanHorizon` (defaulting to
`Years`), never over the drawn lifespan. Amortization would otherwise
liquidate exactly on the drawn death date, risk-based guardrails would band
against a horizon the retiree cannot know, and the whole ensemble would
describe a clairvoyant household. The drawn lifespan is an outcome of the
world, not an input to the policy.

The corollary is that `Years` acquires two jobs, which are then split:

- `Years` is the simulation length and the array length. A household drawn to
  outlive it is **censored**: recorded alive at the horizon, its estate read
  there.
- `PlanHorizon` (new, `0` = `Years`) is the horizon the amortization rule
  plans over.

Without the split, a mortal plan would have to choose between truncating the
longevity tail (set `Years` to age 95 and lose exactly the tail an annuity
insures) and starving the ABW household (set `Years` to age 110 and amortize
over a horizon nobody plans for). With it, a plan runs to age 110 and
amortizes to 95.

## 3. API shape

Opt-in through one nil-able field, so every existing caller, golden and
snapshot is untouched by construction:

```go
type Plan struct {
    // ...
    Lifetime    *Lifetime // nil = the fixed horizon (the historical behaviour)
    Annuity     *Annuity  // nil = no annuitisation
    PlanHorizon int       // horizon the amortization rule plans over; 0 = Years
}
```

### 3.1 Mortality inputs

```go
// MortalityLaw is any survival law for one life.
type MortalityLaw interface {
    Survival(age, years float64) float64
}

type Life struct {
    Age float64       // age at the plan's year 0
    Law MortalityLaw  // nil = FrenchMortality
}

type Lifetime struct {
    Self          Life
    Partner       *Life   // nil = a single-life household
    SurvivorSpend float64 // budget factor after the first death; 0 = 1 (no change)
}
```

`Gompertz` already satisfies `MortalityLaw`, so `FrenchMortality` is the
bundled default and no caller has to know the interface exists. The interface
is one method because that is all the sampler needs: lifespans are drawn by
inverse transform on the survival function itself (§5), so a life table, a
cohort-projected law or a stress law drop in without any new plumbing.

`SurvivorSpend` uses `0 = 1` rather than a pointer: a survivor needing
literally nothing is not a value anybody means, so the sentinel is
unambiguous (unlike `BufferSleeve.DrawThreshold`, where an explicit zero is
meaningful and the field is therefore a pointer). The literature's figure is
around 0.7.

### 3.2 Whose income is it

`Cashflow` gains two fields whose zero values are exactly today's behaviour:

```go
type Cashflow struct {
    FromYear, ToYear int
    Annual           float64
    Owner            Owner   // zero value Household: paid while anyone is alive
    Reversion        float64 // share continuing to the survivor after Owner's death
}
```

`Owner` is `Household` (zero), `Self` or `Partner`. A French pension with a
54 % reversion is `{Owner: Self, Reversion: 0.54}`. A `Household` flow ignores
`Reversion` and simply stops when the household does. Every plan written
before this change is a plan of `Household` flows, which is what a
fixed-horizon plan meant.

### 3.3 Annuitisation

```go
type Annuity struct {
    Year  int     // plan year the annuity is bought
    Share float64 // share of the PORTFOLIO converted at that date, in [0,1]
    Rate  float64 // real rate the insurer prices at
    Load  float64 // insurer margin as a share of the fair income; 0 = actuarially fair
    Joint bool    // pays while either member is alive; else it stops at Self's death
    Law   MortalityLaw // the law the insurer prices with; nil = the household's own
}
```

Three deliberate choices.

`Load` is the **margin** (0 = fair), not the retained fraction. The existing
`AnnuityIncome(..., load)` takes the retained fraction (0.90 keeps 90 %),
which is a live misreading risk in a field where "a 10 % load" is the normal
phrase; the new struct states the margin and converts internally. The older
signature stays as it is, since it is what the web layer and its tests call.

`Share` is a share of the portfolio **at the purchase date**, not of the
initial capital, so a deferred purchase (`Year: 10`) converts what is
actually there. The premium is raised by selling from the tax pockets, so the
gross-up and the realised tax are the same machinery a withdrawal uses: an
annuity bought out of a taxable account really does cost its capital-gains
tax, and the kernel now says so.

`Law` exists because annuitants are not the general population: French
insurers price on the generational annuitant table TGF05 (`FrenchAnnuitants`,
calibrated in §5), which is much longer-lived than the INSEE period table
(31.6 years left at 65 against 21.7). Pricing on the household's own law
and calling the difference "load" conflates two effects that move
differently with age. One optional field keeps them separable.

An `Annuity` needs a `Lifetime` and is ignored without one: it has no price
without an age and nothing to insure without a drawn lifespan. The pairing is
obvious enough that it is stated rather than enforced.

### 3.4 Draws, and why `SimulateOn` changed shape

A stochastic-lifetime plan has two independent sources of randomness. The
solvers and sweeps (`Solve`, `CapitalForRuin`, `Sweep1D`) exist to reuse ONE
draw across many evaluations, so that Monte-Carlo noise cannot break the
monotonicity a bisection needs. Both sources must therefore be reusable
together:

```go
type Draws struct {
    Returns []scenario.Sequence
    Lives   []Lives // nil when the plan has no Lifetime
}

func (p Plan) Draw(nPaths, workers int, seed uint64) Draws
func (p Plan) SimulateOn(d Draws, workers int) Ensemble
```

`DrawPaths` (returns only) is replaced by `Draw`, and `SimulateOn` takes
`Draws` instead of `[]scenario.Sequence`. That is a compile-time break on
purpose: it is the only change that makes it IMPOSSIBLE to simulate a mortal
plan against a set of draws that carries no lifespans. The alternative,
keeping the old signature and silently running the fixed horizon whenever a
`Lifetime` plan reached `SimulateOn`, is precisely the class of bug this
package cannot afford. Four call sites in `pkg/decumul/web` were mechanically
updated.

A `Draws` assembled by hand with returns but no lifespans is still handled
rather than punished: `SimulateOn` fills the missing lifespans deterministically
from a fixed internal seed, so the result is correct and reproducible, and the
solvers still see common random numbers.

`Simulate(nPaths, workers, seed)` is unchanged and remains the obvious entry
point; it now draws lifespans too when the plan has a `Lifetime`.

## 4. Outputs

### 4.1 On each path

```go
type PathResult struct {
    // ...
    LifeYears int     // whole years the household lived (capped at the horizon)
    Outlived  bool    // it was still alive at the horizon (censored)
    Estate    float64 // total real wealth at the household's end
    Annuity   float64 // cumulative real annuity income received
    Premium   float64 // net premium actually converted
}
```

With no `Lifetime`, `LifeYears == Years`, `Outlived == true` and
`Estate == Wealth[Years]`: the fixed horizon is the special case where the
household is certain to reach the end, and every one of these fields keeps a
correct meaning rather than a sentinel. That is what makes "fixed horizon is
a special case" true in the code and not only in the prose.

`Ruined` keeps its name and gains its exact meaning. With a drawn lifetime,
the kernel simply stops at death, so a withdrawal can only fail while someone
is alive: `Ruined` IS broke-while-alive, counted rather than approximated.

**Series after death are frozen, not zeroed.** `Wealth[k]` for `k > LifeYears`
holds the estate and `Spend[k]` holds 0. Zeroing would make an ordinary death
indistinguishable from a ruin in every downstream statistic (drawdowns,
underwater spells, terminal wealth). Freezing keeps `Outcome`'s terminal
quantiles reading the estate distribution, which is the right answer in
lifetime mode and byte-identical to today's in fixed-horizon mode. The
statistics that ARE lifetime statistics were bounded by `LifeYears`
explicitly: years underwater, the worst-decade window, recovery spells, and
every spending statistic (`SpendCV`, `SpendStats`, `SpendBands`), the last of
which now reads per year across the households still ALIVE that year, which
is the conditional a reader wants.

### 4.2 Across the ensemble

```go
func (e Ensemble) LifeOutcome() LifeOutcome
func (e Ensemble) LifeStates() []LifePoint
func (e Ensemble) Estates() []float64
```

`LifeOutcome` carries the alive-ruin probability, the median lived years and
the censored share, the broke-years distribution (mean and p95 of years lived
after running out), the estate distribution (p5/p50/p90 plus the share
leaving nothing) and the mean lifetime income.

That income figure is the TOTAL, portfolio withdrawals plus the pensions and
annuity received, and it has to be. `PathResult.Spend` records only what the
portfolio delivered, since income is netted off the budget before anything is
sold; reading it alone would show an annuitised household's income collapsing
by exactly the amount the insurer started paying it. `PathResult.Received`
was added to carry the other half. The spending VOLATILITY stays on
`SpendCV`/`SpendBands`, where it belongs: a pension and a real annuity do not
move, so all of the swing is in the portfolio-funded part those already
measure.

`LifeStates` is the exact counterpart of `LifeCurve`: the same
`[]LifePoint` shape, counted from the drawn deaths instead of weighted by a
survival curve afterwards. On a plan with no annuity, no survivor adjustment
and household-owned income the two must agree, and §7 measures by how much.
`LifeCurve` stays: it is correct, cheaper, and it is what the existing web
view calls.

## 5. Drawing a lifespan

A lifespan is drawn by inverse transform on the integer survival table:

    n = min{ t >= 0 : S(t) <= u },   u ~ Uniform(0,1)

so `P(n > t) = P(u < S(t)) = S(t)` exactly at every integer year, for ANY
decreasing survival function. No closed-form inversion, no rejection loop, and
the table is precomputed once per `Simulate` call (`Years+1` entries per
life), which turns a draw into a binary search rather than a few dozen `exp`
calls. Beyond the table, the household is censored.

`n` is the number of whole years lived: the household consumes years
`0 .. n-1` and its estate is `Wealth[n]`. A death partway through a year still
consumes that year's spending, which matches the kernel's begin-of-year
withdrawal convention (an annuity-due). That convention is also what makes the
money-back identity of §7 exact rather than approximate:

    E[ number of payments ] = sum_{t>=0} P(n > t) = sum_{t>=0} S(t) = AnnuityFactor at rate 0

### Calibration of the bundled law

**Reference.** INSEE's triennial France period table for 2023-2025
(provisional results published 2026-01-13 with the "Bilan démographique
2025", https://www.insee.fr/fr/statistiques/8721189, file
`6_FR_Table_mortalite_triennale.xlsx`, sheet "2025"), whose "Espérance de vie"
columns are the complete expectation at exact age x, for both sexes
("Ensemble"), women and men. `Gompertz.LifeExpectancy` computes the same
quantity (the integral of `Survival`), so the two compare directly.

**Fit.** Mode and dispersion minimise the squared error of the remaining life
expectancy at every single age from 50 to 90, on each column, then rounded to
a tenth (the rounding costs under 0.01 year of rms):

| law | Gompertz | rms 50-90 | worst 50-90 |
|---|---|---|---|
| `FrenchMortality` (Ensemble) | {89.8, 9.3} | 0.25 y | 0.43 y (age 90) |
| `FrenchWomen` | {91.9, 8.3} | 0.24 y | 0.37 y (age 90) |
| `FrenchMen` | {87.5, 9.9} | 0.25 y | 0.48 y (age 90) |

Remaining years, model versus INSEE (`TestBundledLawsMatchINSEE` holds every
age 50, 55, ..., 90 within 0.5 year):

| age | before {88, 10} | Ensemble fit / INSEE | women fit / INSEE | men fit / INSEE |
|---|---|---|---|---|
| 50 | 33.18 | 35.04 / 34.68 | 37.40 / 37.05 | 32.74 / 32.37 |
| 60 | 24.26 | 25.83 / 25.86 | 27.88 / 27.91 | 23.83 / 23.81 |
| 65 | 20.13 | 21.51 / 21.73 | 23.33 / 23.55 | 19.71 / 19.87 |
| 70 | 16.30 | 17.46 / 17.81 | 19.00 / 19.34 | 15.90 / 16.19 |
| 80 | 9.82 | 10.50 / 10.59 | 11.39 / 11.47 | 9.48 / 9.52 |

The residual has the two-parameter law's shape: about +0.35 year at 50, -0.3
around 70, +0.4 at 90. A Makeham term or a bundled table would remove it; at
a fifth of the gap it replaced, it was not worth the extra parameter.

**Before.** `Gompertz{88, 10}` (the calibration until 2026-10) ran **1.5 to
1.6 years short** of the Ensemble column from 50 to 70 and 0.8 year short at
80: it sat on the MEN's figures and flattered every plan. Measured on the
package examples (20 000 paths, same seeds), the refit moves a 60-year-old
couple's alive-ruin from 11 % to 14 % on `ExamplePlan_Simulate_lifetime`, the
65-year-old couple of `ExampleAnnuity` from 30 % to 33 % without an annuity
and from 24 % to 29 % with a third annuitised at 65, and the spending solved
at a 5 % alive-ruin on `ExamplePlan_Solve` from 24 000 to 23 500 a year.

**What still flatters.** It is a PERIOD law: it carries no future mortality
improvement, where the generational tables an insurer prices with
(TGH05/TGF05) run materially longer. It is unisex by default; `FrenchWomen`
and `FrenchMen` are there for a household that wants each member on their
own column (`Life.Law`). A household wanting the conservative reading should
raise `Mode`. The godoc says this at the constant, and the design records it
here so nobody rediscovers it as a bug.

### Calibration of the insurer's law (`FrenchAnnuitants`, 2026-10-21)

**What an insurer must price on.** The arrêté of 1 August 2006
(JORFTEXT000000820127, art. 2) homologates, for annuity contracts from
2007-01-01, TGF05 for women and TGH05 for men: generational tables of
survivors by age for every birth year 1900 to 2005, built on insured
populations and projected forward from 2005. Article A132-18 of the Code des
assurances (version in force since 2017-09-07) lets a tariff use those
tables or an insurer's own certified table, never below the homologated
tariff, and when ONE table serves every insured, as a unisex tariff must
since 2012-12-21, "celle-ci correspond à la table appropriée conduisant au
tarif le plus prudent". For an annuity the prudent table is the
longer-lived one, TGF05, and the government's answer to written question
41093 (JO AN, 2021-12-28, p. 9213) confirms the market's practice and that it
leaves insurers a technical profit, a large share of which they must hand
back as profit sharing within eight years.

**Source of the numbers.** The survivor counts were read from the
spreadsheet reproduction published by Oract Advisory
(https://www.oractadvisory.com/ressources/, "TGH-TGF-05"), checked point for
point against a second public reproduction (logement-seniors.com, "Les
Tables de mortalité TGF05 et TGH05": all 24 printed values of generations
1950 and 2000 at ages 60, 65, 80, 90, 100 and 110 agree exactly); the
tables also ship in the CASdatasets R package as `freTGF05`/`freTGH05`.
Remaining life expectancy is the curtate expectation off the counts plus
half a year; annuity factors are annuity-due at a 1 % real rate, the
kernel's convention.

**Fit.** The same method as the population law: Mode and dispersion
minimise the squared error of the remaining life expectancy at every age
from 60 to 100, rounded to a tenth. Generation 1978 is the reference cohort
(a buyer born 1975 to 1980 annuitising around 2040):

| table, generation 1978 | Gompertz | rms 60-100 | worst 60-100 | e65 fit / table | joint factor at 65, fit / table |
|---|---|---|---|---|---|
| TGF05 (`FrenchAnnuitants`) | {100.7, 7.7} | 0.17 y | 0.47 y (100) | 31.64 / 31.62 | 31.10 / 31.12 |
| TGH05 | {96.9, 8.4} | 0.21 y | 0.60 y (100) | 27.86 / 27.90 | 28.56 / 28.52 |

On what an annuity actually prices, the fit is closer still: the joint-life
factor over two TGF05 lives is within 0.25 % of the table at every age from
60 to 75 (34.46 / 34.43 at 60, 33.14 / 33.12 at 62, 27.61 / 27.65 at 70), and
the single-life one within 0.65 %. The two-parameter tail is thinner than the
table past 100 (2.9 years left at 110 against 1.8 for the law), which moves
no factor that matters. A mixed couple priced sex by sex (a TGH05 man and a
TGF05 woman) would have a joint factor 3.6 % below the two-TGF05 one at 65
(29.99 against 31.12): that is the price of the unisex rule, and the page
charges it because the law does.

**Cohort drift.** Each later birth year adds about 0.15 year of modal age:
TGF05 generation 1960 fits {98.1, 8.1}, 1965 {98.8, 8.0}, 1975 {100.3, 7.8},
1980 {101.0, 7.7}, 1985 {101.7, 7.6}, all within 0.2 year rms. A buyer born
in 1960 is therefore quoted about 6 % more income at 65 than the reference
cohort (about 1.7 % of the joint factor per five years of birth date). The page
does not know the reader's birth year, so it prices the reference cohort.

**Before and after.** The web page priced on `Gompertz{92, 9}`, a figure
chosen in 2026-08 as "3.3 years above the population law of the time" and
never sourced against TGH05/TGF05. It gave 23.4 years left at 65 where the
table gives 31.6: it priced the insurer's longevity as barely above the
population's, and quoted about a fifth too much income per euro. Income per
100 000 of premium, joint life on a same-age couple, 1 % real rate:

| purchase age | fair, 92/9 -> TGF05 | 10 % margin, 92/9 -> TGF05 |
|---|---|---|
| 62 | 3 659 -> 3 018 (-17.5 %) | 3 293 -> 2 716 |
| 65 | 3 966 -> 3 215 (-18.9 %) | 3 569 -> 2 893 |

On the §8 household (a couple of 65 on a million at 42 k, µ 3.5 %, σ 12 %,
Student-t df 5, 40 years, no tax, no buffer, half the sleeve annuitised at 65,
20 000 paths, measured through `/api/lifecycle`'s cards on the current
population law), the readout moves from "pays 3.6 % vs withdraws 4.2 %,
alive-ruin 27.0 % -> 40.4 %, median estate 419 k -> 85 k" to "pays 2.9 % vs
4.2 %, alive-ruin 27.0 % -> 55.1 %, median estate 419 k -> 0". `ExampleAnnuity`
(a third at 75, 2 % real, 10 % load; it used `Gompertz{91, 10}`, equally
unsourced) moves from 35 % alive-ruin and a 0.19 M estate to 42 % and 0.10 M,
its payout from 5.53 % to 4.25 % of the premium.

**What is still assumed.** The insurer's margin (10 % by default, 2 % to 25 %
on the slider) and the 1 % real pricing rate are planning figures, not
sourced quotes; the profit sharing the regulatory table's prudence generates
is not returned to the household, so the page quotes the contractual floor of
an indexed annuity rather than its expected payout. The book's
`credits-mortalite` plate still draws the annuitant pool on
`Gompertz{91, 10}`.

**Book.** One plate freezes the law: `vivant-ruine-parti` (article
`horizon-et-esperance-de-vie`, `TestVivantRuinePartiMatchesTheEngine`), which
`make figure-drift` reports stale until it is refrozen.

Couple survival treats the two lives as **independent** draws from their own
laws. Real couples are positively dependent (shared environment, shared
shocks, the broken-heart effect), so independence overstates the probability
that at least one member is still alive, which again overstates the plan's
longevity load rather than flattering it. Modelling the dependence needs a
copula or a common-shock factor plus a calibration nobody in this repo can
source, and it would move the results by less than the period-versus-cohort
gap above. Independence is the right stopping point, stated rather than
hidden.

## 6. Kernel changes

Both kernels (`RunPath` and `RunPathMonthly`) take the same per-path `life`
value, whose zero value is immortal, so the mortal and fixed-horizon code
paths are one code path:

1. the loop runs to `life.end()`, which is `Years` without a `Lifetime`;
2. needs-based spending levels are scaled by `life.spendFactor(k)` (1 while
   the household is whole, `SurvivorSpend` after the first death). Wealth-based
   budgets (VPW, bounded percent, ABW) are NOT scaled: their budget is a share
   of what exists, not a restatement of a need;
3. income is netted per owner and per reversion rather than in bulk;
4. at `Annuity.Year` the premium is raised by selling from the tax pockets and
   converted into a lifelong real income at the price `AnnuityFactor` quotes
   for the covered lives at the age then reached;
5. after the loop, wealth is filled forward with the estate.

Nothing above executes when `Lifetime` and `Annuity` are nil beyond one nil
check per path, and the goldens confirm the arithmetic did not move.

## 7. Sanity anchors

Three anchors, each a test, each with a number.

**Mortality off reproduces the fixed horizon exactly.** A `Lifetime` whose law
never kills (a stub returning survival 1) must produce path-by-path
identical results to the same plan with `Lifetime` nil: same wealth series,
same ruin years, same spending. Not a tolerance, an equality.

**The weighting and the kernel agree on alive-ruin.** On a plan with no
annuity, no survivor adjustment and household income, the posterior weighting
`mean_paths S(RuinYear)` and the kernel's counted `RuinAlive` estimate the
same quantity. Measured on a couple retiring at 60 on a million at 45 k a
year, 30 000 paths, over a 55-year horizon: the fixed-horizon headline is
**88.9 %**, the survival weighting says **51.3 %** and the kernel counts
**51.1 %**, a gap of **0.15 pt**. The paired Monte-Carlo standard error is
about 0.29 pt, so the test asserts 1.0 pt (over three sigma). Curve by curve,
`LifeStates` and `LifeCurve` never differ by more than **0.6 pt** on the dead
share and **0.2 pt** on the broke share. This is the anchor that says the new
kernel did not silently change the meaning of ruin, and the 88.9-versus-51
gap is the measure of how much the fixed horizon was overstating.

**Money back at zero rate and zero load.** An annuity priced at a 0 % real
rate with no insurer margin is a pure redistribution across the group: the
expected total income it pays must equal the premium. Measured through the
kernel on 50 000 paths, a 500 000 premium pays back **500 888**, a ratio of
**1.0018**; the test asserts 2 %. It is the identity that proves the mortality
credit is really being realised (early deaths funding late ones) rather than
approximated, and it fails immediately if the draw convention and the pricing
convention disagree by a single year.

**And the trade-off itself is now measurable**, which was the point. On a
couple retiring at 65 on a million at 42 k a year, converting half the
portfolio into an actuarially fair joint annuity moves the alive-ruin from
**24.4 % to 12.1 %** and the median estate from **460 k to 367 k**: the
insurance is bought, and the bequest pays for it. Price it realistically
instead, at the web page's quote (1 % real, the TGF05 annuitant law, a 10 %
load), and it pays 2.9 % where the plan withdraws 4.2 %, so both readings get
worse (the measured figures are in §5, "Before and after"). Neither
result was reachable under a fixed horizon, where the 2026-07 sweep found the
two readings moving together at every age and rate.

`make golden` must not move: no golden plan carries a `Lifetime`, and the
draw stream for returns is unchanged.

## 8. Deferred, with the reason

- ~~**Web / `-fire` wiring.**~~ DONE 2026-08-22. The lifecycle view
  (`pkg/decumul/web/lifecycle.go`) now runs the exact kernel: a couple of two
  lives of the same age drawn from `FrenchMortality`, no survivor adjustment,
  which is exactly the household the posterior weighting assumed. The same
  `Draws` are replayed with `Lifetime` nil, so the "ignoring mortality" card is
  the paired mortality-free twin of the counted figures rather than a second
  simulation. The stack reads `LifeStates`, the terminal-wealth histogram reads
  `Estates` (wealth at the household's own end, so its upper bands thin out:
  measured on the default plan, "8M+" falls from 17.1 % to 12.8 % because most
  households die before the horizon), and one card was added for the length of
  a failure (`BrokeYearsMean / RuinAlive`).
  Since 2026-10 the stack reads `LifeBands(1, 2)`, `LifeStates` with the
  funded share split by real wealth against the path's own starting capital
  (at or under it, up to twice it, above), its dead and broke shares being
  `LifeStates`' own counts. The broke layer sits on the baseline and the dead
  one hangs from the top, so each risk reads against an axis; the x axis is
  the first member's age, the plan year riding in the hover header.
  Measured on the page's default plan, 20 000 paths: the alive-and-broke
  figure moves 1.62 % -> 1.65 %, the still-alive-at-horizon figure 40.9 % ->
  40.1 %, and the alive-broke-dead curves never differ by more than 0.80 pt on
  the dead share and 0.06 pt on the broke share, all inside the Monte-Carlo
  error of the run.
- ~~**The annuity panel.**~~ DONE 2026-08-22. The rail's `annuityShare` control
  bought a `Cashflow` that paid for ever out of an untaxed premium, the
  fixed-horizon reading of an annuity; it now feeds `Plan.Annuity`
  (`pkg/decumul/web/annuity.go`) with two more parameters, `annuityYear` (the
  plan year of the purchase, clamped into the plan since the page starts at
  retirement) and `annuityLoad` (the insurer's margin, 2 % to 25 %, defaulting
  to 10 %, where an unset zero means that default rather than a fair annuity so
  a link written before the control cannot quietly buy one). Fixed, and stated
  in the UI rather than offered as knobs: a joint-life, inflation-linked quote
  at a **1 % real rate** on an **annuitant** table (`FrenchAnnuitants` since
  2026-10-21, TGF05 generation 1978, 31.6 years left at 65, calibrated in §5;
  until then `Gompertz{92, 9}`, 23.4 years, chosen as 3.3 years above the
  population law of the time and never sourced against TGH05/TGF05).
  **The panel lives in `/api/lifecycle` and nowhere else**, because that is the
  only view whose households can outlive their money. Everything fixed-horizon
  ignores the block, and the view carries one standing line saying why. Wiring
  it page-wide was refused: turning a slider off zero would silently redefine
  ruin everywhere from broke-by-horizon to broke-while-alive, which is exactly
  the category error the page fights between the guardrails and the fixed rule.
  The readout is three cards, no new chart: a THIRD twin runs the same draws
  and the same deaths with the purchase removed, giving `broke while alive`
  and `median estate` before and after, plus the ratio that decides their sign,
  what the annuity **pays** against what the plan **withdraws**.
  Measured on the § 7 household (a couple retiring at 65 on a million at 42 k,
  µ 3.5 %, σ 12 %, no tax, no buffer, half the sleeve annuitised at 65),
  the sign is entirely a pricing question, and the readout says so: at the
  §7 quote (fair, 2 % real, population table) it pays 4.97 % against a 4.2 %
  withdrawal and moves alive-ruin 24.0 % → 12.0 % with the estate 460 k → 367 k,
  while at the page's realistic quote (1 % real, annuitant table, 10 % margin)
  it pays 3.57 % and moves alive-ruin 24.0 % → 36.6 % with the estate 460 k →
  117 k (measured in 2026-08 on the laws of the time; on the TGF05 law it pays
  2.9 %, see §5). Break-even sits near a payout equal to the plan's
  withdrawal rate.
  The truncation caveat is named on the card: the plan stops at the reader's
  horizon, so payments the annuity would still make after it are neither
  collected nor counted.
  The sensitivity tornado deliberately gets no annuity bar (an annuity is a
  decision, not a parameter to perturb).
  The old control's URL key is KEPT rather than dropped: `annuityShare` still
  means "annuitise this share", so a link written under it still asks for the
  same thing and now gets it priced properly, with the two new keys taking
  their defaults. Its figures are not the ones its sender saw, and cannot be:
  those came from a product nobody sells.
- **Correlated couple mortality.** See §5: the effect is smaller than the
  period-versus-cohort gap already accepted, and it needs a calibration this
  repo cannot source.
- **Full life tables (INSEE, TGH/TGF-05).** The interface already accepts
  them; what is missing is bundled data and a generator, which is a
  `pkg/datasets` job with its own validation spec, not a kernel job. Since
  2026-10 the two-parameter laws sit within half a year of both references
  at the planning ages (INSEE 2023-2025 for the household, TGF05 for the
  insurer, joint annuity factors within 0.25 %), so a bundled table would
  buy less than the cohort choice it would also have to make.
- **Reversion-rate and deferred-with-no-cash-value annuities.** A joint
  annuity paying the survivor a reduced rate, and a pure longevity annuity
  (deferred, no surrender value), both change the PRICING formula, not the
  kernel. A deferred purchase is already expressible as `Annuity.Year > 0`.
- **Health-cost and long-term-care shocks.** A late-life spending shock
  correlated with survival is a real risk and a real modelling job; the
  deterministic part of it is already expressible through
  `Plan.SpendSchedule` (the retirement smile), and the stochastic part needs
  its own literature pass.
- **Mortality inside the monthly kernel at monthly granularity.** Deaths are
  drawn in whole years in both kernels. A monthly death date would change the
  spending of one year by at most half a year's need, far below the Monte-Carlo
  error of any statistic the package reports.
