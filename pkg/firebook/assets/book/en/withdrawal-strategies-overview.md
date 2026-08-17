# Withdrawal strategies: the map before the territory
<!-- source: panorama-strategies-retrait @ 4598f040315a -->

You have a portfolio, a horizon and an income to fund. One decision now governs the whole withdrawal phase: what rule, exactly, will you draw by? How much in year one? And above all, how will that amount respond to markets, to inflation, and to the years going by?

A dozen or so named answers exist: Bengen, the fixed percentage, Guyton-Klinger, VPW, CAPE rules, Morningstar guardrails, ABW/TPAW, Vanguard's floor and ceiling, annuities. The literature has been comparing them for twenty-five years. The good news is that the zoo is smaller than it looks. All of it turns on a single three-way tradeoff. Once you have the map, each rule becomes a point you can place at a glance, with predictable strengths and predictable pathologies.

This article is that map. It lays out the impossible triangle of withdrawals, the two extreme rules that bound the space, the five families and the article devoted to each, the six criteria that let you score a rule honestly, and how to read the decumulation frontier, the chart that places every rule against every other. The ten articles that follow take each rule apart. This one keeps you from getting lost among them.

::: cle The impossible triangle of withdrawals
Every withdrawal rule promises three good things: a **stable income** (the same standard of living every year), **capital safety** (never running the portfolio down to zero) and a **high income** (spending a generous share of the expected return). None can deliver all three at once. On a risky asset it is mathematically impossible. So every named strategy is a choice about what to give up. Bengen gives up safety. The fixed percentage gives up stability. A tiny withdrawal gives up the level. The clever rules are simply finer ways of spreading that sacrifice around. "Which rule is best?" has no answer until you say which sacrifice you are willing to make.
:::

::: figure withdrawal-frontier
The decumulation frontier: every rule is a point between failure probability and the variability of the standard of living. The two extremes (Bengen, the fixed percentage) bound the arc, and the modern rules move toward the ideal corner by listening to more information.
:::

## The two pure extremes, and the frontier between them

Start with the two strategies that bound the whole space. Each one delivers two corners of the triangle perfectly, by abandoning the third.

**The income-first extreme: the fixed inflation-adjusted amount** (Bengen's rule, [[fixed-inflation-adjusted-withdrawal]]). You take EUR X in year one, then the same amount adjusted for inflation, whatever happens. Income is perfectly stable, by construction. The price is that ruin becomes possible: if markets disappoint for long enough, the portfolio runs out before you do ([[the-4-percent-rule]], [[sequence-of-returns]]). All the risk piles onto one event, binary and far away.

**The capital-first extreme: the fixed percentage** ([[fixed-percentage]]). Each year you take Y% of the portfolio as it stands. Ruin becomes impossible, by construction: Y% of something is never all of it. In exchange, income tracks the market one for one. A portfolio down 35% buys a standard of living down 35%, possibly for years. All the risk spreads out over daily life.

Between those two poles lies a continuum, and it gives this whole part its key chart, the **decumulation frontier**. Every rule becomes a point on it, fixed by two coordinates and no more: its failure probability, and the swing in living standard it imposes. Bengen sits at the bottom right, maximum failure and zero variability. The fixed percentage sits at the top left, zero failure and maximum variability. The rules in between trace an arc. The picture makes twenty-five years of research visible at once. You do not remove the risk from a withdrawal plan, you choose its shape: a rare and brutal bankruptcy on one side, frequent but bounded adjustments on the other.

The rules that listen are not magic. They simply sit better on the arc, closer to the ideal corner where failure and variability are both low. They get there because they listen to information the extremes ignore: the current portfolio, valuations, the horizon that is left.

There is nothing mysterious about the chart, and you can draw it yourself. A withdrawal engine only ever runs one spending policy at a time, so you get the frontier by running the full plan once per candidate rule and writing down the pair it returns, failure and variability. The scatter you accumulate is your plan's frontier. A few tools plot it for you, but any simulator that lets you swap the withdrawal rule will give it to you in a handful of runs. Two precautions keep the result readable. Hold the inputs and the market model strictly fixed from run to run, or you end up comparing rules and worlds at the same time. And settle on one measure of variability once and for all: the standard deviation of annual real spending, the spread of the worst quartile, the count of years spent below the comfort floor, any of them works, but mixing them across rules makes the numbers meaningless.

## The full taxonomy

Here is the zoo, sorted into five families by the information each rule listens to. The table is the map, and each rule gets its own article.

| Family | What it listens to | Named rules | Article |
|---|---|---|---|
| **Fixed** | Nothing (the original plan) | Bengen's 4%, variants with partial indexation | [[fixed-inflation-adjusted-withdrawal]] |
| **Proportional** | The current portfolio | Fixed percentage, the Bogleheads' VPW (a percentage that rises with age) | [[fixed-percentage]], [[vpw]] |
| **Guardrails** (the fixed rule that bends) | The current portfolio, through thresholds | Guyton-Klinger (the classic), floor and ceiling, Vanguard's dynamic spending, the modern Morningstar and Kitces-Tharp guardrails | [[guyton-klinger]], [[floor-and-ceiling]], [[morningstar-guardrails]] |
| **Actuarial / amortization** | The portfolio, the remaining horizon and the expected return | ABW, TPAW, the RMD tables, ERN's dynamic CAPE rules | [[amortization-based-withdrawal]], [[cape-based-rules]] |
| **Guaranteed floor** (safety-first) | Nothing: it takes the floor out of the portfolio | Life annuities, a ladder of inflation-linked bonds for the essentials plus a portfolio for the rest | [[annuities-and-safety-first]], [[bond-ladders]] |

::: figure familles-information
The table above, read as a ladder. The columns fill up from top to bottom, and the governance gauge rises with them, which is the tradeoff of this whole part in one picture. Note the fifth row, set apart on purpose: like the first, it listens to nothing, and yet it looks nothing like it. That is the sign it is not playing the same game.
:::

Three things stand out.

**The more information a rule listens to, the better it sits on the frontier, and the more governance it demands.** The fixed rule demands nothing, and learns nothing. Guardrails demand that you take cuts decided in advance, while calm, at the worst emotional moment. Amortization demands an annual recalculation and an income that is officially variable. The gain in position on the frontier is real and measurable. The cost is behavioral, and it is real too ([[the-psychology-of-spending]]).

**Families 3 and 4 have won the recent research.** The modern consensus (Morningstar, Kitces, ERN, the Bogleheads, the actuarial literature) has moved. The pure fixed rule teaches well but is dominated. The pure percentage is indestructible and unlivable. The rules that listen, well-bounded guardrails and smoothed amortization, buy the best trade between failure and the life you actually live. The literature prefers the ABW/TPAW family for its internal consistency: it is the only one that can neither run dry early nor die on a pile of gold ([[amortization-based-withdrawal]]). Practitioners still prefer guardrails, because a client can read them ([[morningstar-guardrails]]).

**Family 5 changes the game.** Safety-first is not hunting for a better point on the frontier. It takes part of the spending out of the triangle altogether, the vital floor, now covered by an annuity or a pension. The portfolio then carries only the comfort layer, where variability is tolerable. Wade Pfau is the school's contemporary theorist, and his line sums it up: retirement is not a portfolio problem, it is an asset-liability matching problem. Two camps shape the whole professional debate, those who optimize the probabilistic frontier (probability-based) and those who guarantee the floor first (safety-first). The best practical answer is usually a hybrid, and most real plans already are one: Social Security covers the floor, the portfolio funds comfort ([[pensions-and-other-income]], [[us-healthcare-and-social-security]]).

## Scoring a rule honestly: the six criteria

Comparing rules takes more than a success rate. Judge on that alone and the fixed percentage always "wins" (zero failure!) while sweeping its pathology under the rug. Here is the scorecard, inherited from ERN (Part 11) and from practice ([[the-ern-series]]):

1. **Failure** (at a given horizon, under a given model): the classic criterion, necessary and never sufficient ([[failure-probability]]).
2. **The average standard of living delivered**: how much the rule actually lets you spend, on average, over the life of the plan. Two rules with the same failure probability can differ by 15% in total consumption.
3. **The distribution of living standards in the bad scenarios**: the criterion that really separates the rules. What does the rule deliver in the worst quartile, how many years below the comfort floor, and how far below? Answering means leaving the aggregates behind and reading the spending delivered year by year, along the harshest paths. That reading is what unmasked Guyton-Klinger: superb success rates paid for by decades cut 30 to 45% in the 1966 vintage ([[guyton-klinger]], [[flexibility-in-practice]]).
4. **The bequest** (the distribution of ending wealth): some rules die rich every time (the cautious fixed rule), others spend everything (amortization over an exact horizon). Neither is "right", but you should know which one you signed up for ([[spending-in-retirement]]).
5. **Governance**: can a human under stress run this rule, and can a surviving spouse? A seven-parameter rule recomputed every month is a promise that it will not be followed ([[couples-and-family]], [[the-annual-review]]).
6. **Robustness to bad assumptions**: what becomes of the rule if μ was overstated by a point? The rules that listen correct themselves along the way, which is their deepest strength. The fixed rule absorbs the error in silence, all the way to the cliff.

::: attention The trap in success-rate comparisons
You will see tables everywhere along the lines of "rule X succeeds 98% of the time against 91% for Bengen". Train the right reflex: ask what the rule did to spending to buy those points. Any flexible rule can reach a 100% success rate by cutting hard enough for long enough. The portfolio's success is then paid for by the failure of your standard of living, which the table does not show. The only honest comparison has two dimensions, failure and the life actually lived, which is to say the frontier. A ranking built on a single number cannot be honest, however good the simulations behind it.
:::

## The map in action: one plan under four rules

Nothing beats a worked case. The plan: EUR 1.5M, a comfort need of EUR 54,000 a year (3.6%), a floor set at EUR 42,000 ([[how-much-you-need]]), a 45-year horizon, and a pension of EUR 15,000 a year starting in year 17. Here are the four points on the frontier (indicative numbers from the central model, run your own):

- **Bengen, EUR 54,000 indexed**: failure about 9%. Perfectly stable income, right up to the cliff. Comfortable median bequest. The reference point.
- **Fixed percentage, 3.6%**: failure 0%. But the worst quartile spends years below the EUR 42,000 floor as soon as the first hostile regime arrives, and income swings by plus or minus 25% from one decade to the next. Unlivable for this household.
- **Guardrails (a −10% cut when the current withdrawal rate passes 4.5%, floor at 78% of comfort)**: failure about 3%, income stable most of the time, and in the bad quartile two to four cuts, never below the stated floor. Average lifetime spending comes out 4% below Bengen's.
- **ABW (amortized over the remaining horizon, central return, smoothed)**: failure structurally near zero, the highest average total consumption of the four (the rule dares to spend what the others hoard), income officially variable but bounded by the pension and by the smoothing, and a small bequest, accepted up front.

The lesson is not that ABW wins. It is that the answer depends on the household. This one has a high floor (78% of comfort) and wants to spare the spouse who will have to run the plan, so plain guardrails take it. A flexible household with no interest in leaving a bequest would have taken ABW. A household that flinches at any variation would have taken Bengen at 3.2%, with more capital. The full selection work, criterion by criterion and profile by profile, is [[choosing-your-strategy]], the closing article of this part.

## How to read the rest of this part

The articles follow the map. The extremes come first, because they bound everything: [[fixed-inflation-adjusted-withdrawal]], then [[fixed-percentage]]. Guardrails come next, from the historical classic ([[guyton-klinger]], worth reading for the mechanism and for its pathologies) to the modern versions ([[floor-and-ceiling]] for the Vanguard family, [[morningstar-guardrails]] for the practitioner state of the art). Then the actuarial family: [[vpw]] (the bridge from proportional to actuarial), [[amortization-based-withdrawal]] (the literature's favorite) and [[cape-based-rules]] (valuations as information, [[valuations-and-cape]]). Then the change of game: [[annuities-and-safety-first]] (the guaranteed floor). Then the seven rules side by side, on retirements that actually happened, year by year ([[seven-ways-to-live-on-one-portfolio]]), and the decision summary, [[choosing-your-strategy]]. Each article gives the exact mechanics, the parameters the research recommends, the known pathologies, and how to run the rule in practice. All of them fit in a few lines of arithmetic, and a serious simulator will swap one for another without changing anything else in the plan. That is exactly what you need to place them on the frontier one at a time.

## The essentials

- The impossible triangle, stable income, safe capital, high income: every rule gives up one corner, and "the best rule" does not exist until you say which sacrifice you are choosing.
- The two extremes bound the space: Bengen (stable, ruin possible) and the fixed percentage (indestructible, unlivable). Every other rule lives on the frontier between them, where each comes down to a pair of numbers, failure and the variability of the standard of living.
- Five families, sorted by what they listen to: fixed, proportional, guardrails, actuarial, guaranteed floor. Recent research favors well-bounded guardrails and amortization, practitioners favor how readable guardrails are, and the guaranteed floor changes the game instead of moving along the frontier.
- Six criteria for scoring a rule: failure, average standard of living, the life lived in the worst quartile (the one that separates them), bequest, governance, robustness to bad assumptions. Any table showing only a success rate is hiding what matters.
- The same plan changes its best rule depending on the floor, the bequest you want and your tolerance for adjustments. The reasoned selection is in [[choosing-your-strategy]].

---

## Going further

- Early Retirement Now, Part 11 (the criteria for scoring dynamic rules): the founding scorecard ([[the-ern-series]]).
- Wade Pfau, *Safety-First Retirement Planning* (2019): the liability-matching school, against the probability-based one.
- Morningstar, *The State of Retirement Income*: the annual comparison of the rules, scored on the life actually lived ([[morningstar-guardrails]]).
- Bogleheads wiki, "Withdrawal methods": the community inventory, well maintained.
- In this book: [[seven-ways-to-live-on-one-portfolio]] (the seven rules replayed on three real retirements), [[choosing-your-strategy]] (the reasoned selection, criterion by criterion) and [[under-the-hood]] (how a simulator runs these rules and draws the frontier).
