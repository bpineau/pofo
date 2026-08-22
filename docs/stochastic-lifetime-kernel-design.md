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
  (`fire-book-illustrations-2026-07.md` §10.3), with "would need a
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
insurers price on generational annuitant tables (TGH/TGF-05), which are much
longer-lived than the INSEE period table. Pricing on the household's own law
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

`FrenchMortality = Gompertz{Mode: 88, Dispersion: 10}` gives a remaining life
expectancy at 65 of **20.1 years**. INSEE's 2025 period table gives 19.7 for
men and 23.4 for women, so the bundled law sits essentially on the male
figure and about 1.5 years below the unisex average. It is also a PERIOD law:
it carries no future mortality improvement, where the generational tables an
insurer prices with (TGH/TGF-05) run materially longer.

Both gaps point the same way, and it is the flattering way: the bundled law
**understates** longevity, therefore understates the risk an annuity insures.
A household wanting the conservative reading should raise `Mode`;
`Gompertz{91, 10}` gives 22.6 years at 65, in the region of an annuitant
table. The godoc says this at the constant, and the design records it here so
nobody rediscovers it as a bug.

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
instead, a third at 75 on an annuitant table with a 10 % load, and it pays
3.6 % where the plan withdraws 4.2 %, so both readings get worse. Neither
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
  at a **1 % real rate** on an **annuitant** table (`Gompertz{92, 9}`, about
  three to four years of remaining life expectancy above the bundled population
  law, the order of magnitude of the TGH/TGF-05 gap).
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
  117 k. Break-even sits near a payout equal to the plan's withdrawal rate.
  The truncation caveat is named on the card: the plan stops at the reader's
  horizon, so payments the annuity would still make after it are neither
  collected nor counted.
  The sensitivity tornado deliberately gets no annuity bar (see
  `docs/decumulation-fire-design.md` § 6).
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
  `pkg/datasets` job with its own validation spec, not a kernel job. The
  two-parameter law is within about 1.5 years of remaining life expectancy at
  the planning ages, which is smaller than the spread between the male and
  female tables it averages.
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
