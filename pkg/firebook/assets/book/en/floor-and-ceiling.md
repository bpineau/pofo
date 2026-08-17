# Floor and ceiling and the Vanguard rules: bounded flexibility
<!-- source: plancher-plafond @ 3fd361881015 -->

Between the fixed amount that ignores the market ([[fixed-inflation-adjusted-withdrawal]]) and the percentage that follows its every move ([[fixed-percentage]]) sits a third route, disarmingly simple. Follow the percentage, but put bounds on the movement. That is the floor-and-ceiling family. Its industrial version is Vanguard's dynamic spending rule: each year you aim at X% of the current portfolio, but real income may not rise more than 5% or fall more than 2.5% from last year.

Two bounds, nothing else. That gentle asymmetry, up twice as fast as down, is enough to turn the percentage's brutal volatility into glides you can live with, and it keeps almost all of the self-correction. This is the rule one of the two largest asset managers in the world recommends to its retired clients. It is also one of the easiest in the whole survey to run: two comparisons a year and you are done.

This article takes it apart. Two lineages first, Bengen's own floor and ceiling and then the Vanguard corridor. Then the year-by-year mechanics, what the bounds really do to the risk (they put ruin back on the table, and you have to understand that and accept it), how to choose the parameters, and where the rule stands against the guardrails and against ABW.

::: cle The Vanguard rule in three lines
1) This year's target = w × the current portfolio (w set at the start, 4% say). 2) Ceiling: the withdrawal may not exceed last year's withdrawal, in real terms, × 1.05. 3) Floor: it may not drop below × 0.975. The year's withdrawal is the target, clipped by the two bounds. In normal times you live on something very close to a straight percentage. In a crash you glide down 2.5% real a year instead of falling. In a boom you climb 5% a year instead of running away with it.
:::

::: admin How to run it
- **Two reference points, and never mix them up.** The target applies to the **current portfolio**: w × today's portfolio, with w set at the start and never revised. The bounds apply to the withdrawal **actually delivered last year**, not to last year's target and not to the initial withdrawal. That is what gives the rule its short memory and its lack of an anchor: after ten years, nothing left in the arithmetic remembers the day you retired.
- **Indexation.** Everything runs in constant euros. Re-index last year's delivered withdrawal to inflation first, then apply the +5% / −2.5% bounds to it. One consequence explains much of the rule's success: with inflation at 2 to 3%, the deepest cut allowed lands somewhere between −0.6% and +0.4% **in current euros**. The bank statement barely moves while purchasing power slips 2.5% a year. That is deliberate ([[the-psychology-of-spending]]).
- **Frequency.** Annual. This is the family where the cadence matters least, since the smoothing comes from the bounds and not from the date.
- **The thresholds are one point on a grid, not a constant of nature.** Vanguard swept the possible pairs, trading the 35-year success rate against income stability, and settled on this one, which survives 35 years about 85% of the time; its 2020 publication uses an even more timid default, +5 / −1.5. So move the bounds without guilt: (+4 / −2) for a rigid budget, (+6 / −4) for an elastic one. The asymmetry has to stay. It encodes a measured fact: a drop in income hurts more than a rise of the same size helps ([[the-psychology-of-spending]]).
- **Floor.** None built in, and this is the rule's one real flaw: its lower bound is a **speed**, not a level, so nothing stops the descent if the hostile regime drags on. The admission test stands in for it: (1 − the lower bound) to the power n has to stay above your real floor, for an n as long as a genuine hostile regime, twenty years and not six.
- **In your head.** Income delivered = w × portfolio, clipped to the range [last year's income × 0.975, last year's income × 1.05].
:::

## Two lineages, one idea

**The Bengen line: bounds on the level.** Few people know it, but the inventor of the fixed withdrawal proposed a floor-and-ceiling variant himself, back in his work of the 1990s and 2000s. Take a percentage of the current portfolio, but bound it in absolute level. Never below about 85 to 90% of the initial real withdrawal: that is the floor. Never above about 120 to 125%: that is the ceiling. You get a percentage locked inside a band fixed once and for all around the starting point. The floor guarantees a predictable minimum income, which is excellent for covering the part of the budget you cannot compress ([[how-much-you-need]]). The price: in a very bad regime you draw "at the floor" from a collapsed portfolio, and ruin becomes possible again, concentrated in the extreme scenarios. Bengen measured that this band bought about 0.25 to 0.5 of a point of initial rate over his fixed rule. Bounded flexibility was already buying its half point ([[flexibility-in-practice]]).

**The Vanguard line: a corridor on the change.** The rule Vanguard published (the "From assets to income" research, the backbone of its retirement advice since the 2010s) moves the bounds. They no longer cap the level against year 1; they cap the change from one year to the next, +5% / −2.5% in real terms. The difference runs deep. Bengen's band is anchored to the past: the initial withdrawal stays the eternal reference, the dead memory of the fixed rules. The Vanguard corridor has no anchor at all. After ten years of glides, income can sit at 70% or at 140% of where it started. It followed reality, slowly. This is asymmetric exponential smoothing of the fixed percentage, a direct cousin of the Yale rule ([[fixed-percentage]]), with different half-lives up and down.

The asymmetry is not decoration. It encodes a documented human preference: a cut in your standard of living hurts about twice as much as a raise of the same size feels good. It also encodes a statistical fact. Markets rise more often than they fall, so the upper bound does most of the work and keeps a lid on euphoria, while the lower bound, rarer, cushions the shocks. The (+5/−2.5) pair is Vanguard's setting, and you can tune it (see below).

## The mechanics at work: five hard years

Nothing beats walking through it. The plan: EUR 1.4M, w = 4%, initial withdrawal EUR 56,000.

| Year | Portfolio (real) | 4% target | Bounds (real) | Income delivered |
|---|---|---|---|---|
| 1 | 1,400,000 | 56,000 | (reference) | 56,000 |
| 2 | 1,190,000 (−15%) | 47,600 | ≥ 54,600 | **54,600** (floor, −2.5%) |
| 3 | 1,010,000 (−15%) | 40,400 | ≥ 53,235 | **53,235** (floor) |
| 4 | 1,090,000 (+8%) | 43,600 | ≥ 51,904 | **51,904** (floor again) |
| 5 | 1,240,000 (+14%) | 49,600 | ≤ 54,499 / ≥ 50,606 | **50,606** (floor) |
| 6 | 1,400,000 (+13%) | 56,000 | ≤ 53,136 | **53,136** (ceiling, +5%) |

::: figure corridor-1966
The same hostile vintage, three rules, one capital of EUR 1M. The fixed indexed rule delivers EUR 40k without flinching, right up to the wall: the portfolio is empty in 1994. The pure percentage never runs out, but it makes you live the crisis in real time, down 55% of income at the 1982 trough. The Vanguard corridor makes the same underlying correction, in glides of at most 2.5% a year: income falls from EUR 40k to EUR 22.8k, but over twenty-three years, and never in a single jump. **That is what the rule delivers: the same adjustment, spread out until it becomes livable.**
:::

Read years 4 to 6 closely. The portfolio recovers, but the withdrawal keeps sliding down for a while, because the 4% target is still far below it; only then does it climb back, at the capped pace. Across six years of a severe crossing, the years spent inside a bear market, with the portfolio down 28% at the trough, income never moved by more than 2.5% in a year, for a cumulative sacrifice of about 10% at worst. The same sequence under a pure percentage would have delivered −28% in two years. That is the product: glides instead of falls.

The price sits in the same table. In years 3 and 4 you draw 4.8 to 5.3% of a shrunken portfolio: the capped descent borrows from capital during the crisis. In ordinary scenarios the portfolio pays that back on the recovery. In a really long, hostile regime ([[market-regimes]]), the borrowing piles up and ruin becomes possible again. That is the key point to take away. The bounds sell back part of the percentage's anti-ruin guarantee to buy stability. So the Vanguard rule sits between the fixed rule and the percentage on the frontier ([[withdrawal-strategies-overview]]). Its failure probability, low but not zero, is a real number, in the same range as the fixed rule's. That is the whole difference from the hollow 0% of the pure percentage, or from the incomparable number Guyton-Klinger posts when it runs without a floor ([[guyton-klinger]]).

::: science What the comparisons say
In Vanguard's own studies and in independent comparisons (Morningstar tests a close variant every year, [[morningstar-guardrails]]), the bounded rule behaves like a steady B student: failure well below the fixed rule at the same initial rate, typically two to three times lower; slightly higher total consumption; income variability far below the pure percentage, a standard deviation of annual changes around 2% against about 11%; a bequest in the middle. It wins no category. It is second everywhere, with no known pathology. That is exactly its argument. The rules that win a category pay for it somewhere else, ABW on consumption, the guardrails on stability. The bounded corridor is the choice of someone who refuses to pay dearly for anything.
:::

## Choosing your parameters

Three decisions, in order of importance.

**The percentage w.** Same logic as the fixed percentage ([[fixed-percentage]]). The geometric bound applies: w has to stay below the expected real geometric return for median income to hold flat. Self-correction allows the same generosity as everywhere else, so 4 to 4.5% is defensible where the fixed rule would demand 3.25 to 3.5%. In an expensive market, mark it down ([[valuations-and-cape]]), or better, let the CAPE anchor judge your w for you.

**The pair of bounds.** (+5/−2.5) is the standard. Two variants earn their keep: (+4/−2) for tighter budgets, with an even gentler descent and slightly higher failure; (+6/−4) for elastic budgets that want to track the target more closely. The consistency rule is simple: the lower bound has to stay bearable once it compounds over several years. Six years at −2.5% a year comes to −14%. Does that still clear your real floor? That is the family's admission test, and you check it on the distribution of income delivered year after year, never on the failure probability alone.

::: figure corridor-borne
The admission test in one picture. Each curve is your income after n straight years of falling at the lower bound; the red line is the floor of the household in the example below. The Vanguard standard lasts 7.8 years before it crosses that floor, the gentle variant nearly ten, the elastic one under five. The circle marks what the 1966 vintage actually asked of the corridor: twenty-three years of sliding in a row, down to 57% of the starting income. **Test the bound over the length of a real hostile regime, not over six polite years.**
:::

**How it interacts with outside income.** Like the whole proportional family, the rule applies to the portfolio alone. In the uncovered phase of a FIRE plan, the bridge years, the pension bridge gets provisioned separately ([[vpw]], [[horizon-and-life-expectancy]]). Once the pension covers the floor, the corridor on whatever portfolio is left becomes nearly risk-free.

::: astuce Testing the bounded corridor in a simulator
The rule is two comparisons a year, so a spreadsheet is enough to run it yourself. In a decumulation simulator, look for a bounded-percentage spending policy with its three parameters exposed, the target percentage and the two bounds: "Bounded % of portfolio (Vanguard-style)".

- **One spending policy at a time.** The corridor replaces the fixed withdrawal, the flexible cut or the guardrails; it does not stack on top of them. Test the rules one after another, on the same plan and the same market assumptions.
- **Judge it on the income delivered.** The distribution of spending year by year is the only place the compounded lower-bound test can be checked.
- **Do not ignore failure, though.** Unlike the pure percentage and the actuarial rules, the bounded corridor can empty the portfolio. A tool that reports 0% failure on this rule has almost certainly simulated a percentage with no bounds at all.
:::

## Who it fits, against the two finalists

Here is the profile the bounded corridor was made for. A household that wants ease of execution above all: one multiplication, two comparisons, no failure thresholds to watch, no actuarial machinery to run. It wants an income that never surprises, moving no more than 2.5 to 5% a year, known in advance. And it accepts being second everywhere. This is probably the best default rule for a surviving spouse who does not manage money ([[couples-and-family]]), and an excellent cruising rule for the covered phase.

Against the finalists, two trade-offs. Against the guardrails ([[morningstar-guardrails]]), the corridor trades rare steps of plus or minus 10%, and the emotional weight of a decision each time, for continuous glides that ask for no decision at all: less optimal, more livable for many people. Against ABW ([[amortization-based-withdrawal]]), it gives up consumption optimality and any awareness of the horizon, and gets governance that fits on a postcard. The general guide to picking one is [[choosing-your-strategy]].

::: exemple Calibrating your version in twenty minutes
The household: real floor EUR 41,000, comfort EUR 50,000, portfolio EUR 1.3M (target w of 3.85%), pension in 15 years. Start with the lower-bound test. A six-year crossing at the −2.5% floor takes income to 50,000 × 0.975^6, about EUR 43,100, still above the floor, so (+5/−2.5) is admissible. Now simulate. Under a central model the corridor comes out at 2.8% failure against 6.1% for a fixed EUR 50,000 withdrawal, and its worst income quartile drops 9% for five to seven years, which the household accepts. The (+4/−2) variant, tested next, gives 3.4% failure and a worst quartile of −7%: the household prefers it and writes it into the plan. Twenty minutes, two simulations, a rule they own. Compare that with the hours of argument risk-based guardrails demand, for a result close to this one in most scenarios.
:::

## The essentials

- Two lineages: Bengen's floor and ceiling, with bounds on the level around the initial withdrawal, and the Vanguard corridor, with bounds on the annual change, +5% / −2.5% in real terms. The second one, which carries no dead memory, became the standard.
- What it delivers: glides instead of falls, never more than a 2.5% real cut in a year. What it costs: the capped descent borrows from capital during a crisis, and ruin becomes possible again, low but honest, in the same range as the fixed rule's.
- Second everywhere, first nowhere, no pathology: this is the rule for someone who refuses to pay dearly for optimality of any kind, and probably the best default for a non-specialist running the plan.
- Parameters: w chosen as for a percentage (the geometric bound, 4 to 4.5% defensible), bounds of (+5/−2.5) as standard, and the admission test. The lower bound compounded over six years has to stay above the real floor.
- As easy to simulate as it is to run, provided you judge it on the income delivered as much as on failure, and compare it with the guardrails and ABW on the same plan before choosing ([[choosing-your-strategy]]).

---

## Going further

- Vanguard, "From assets to income: a goals-based approach to retirement spending" (the founding research behind the rule, free) and its updates.
- Bengen, *Conserving Client Portfolios During Retirement* (2006): the original floor-and-ceiling.
- Morningstar, *The State of Retirement Income*: the bounded variant in the annual comparison ([[morningstar-guardrails]]).
- In this book: [[fixed-percentage]] (the raw material and the Yale rule), [[guyton-klinger]] and [[morningstar-guardrails]] (the other road to a corridor), [[amortization-based-withdrawal]] (optimality against simplicity).
