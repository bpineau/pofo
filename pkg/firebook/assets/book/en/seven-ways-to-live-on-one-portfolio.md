# Seven ways to live on one portfolio
<!-- source: sept-facons-de-vivre @ a822829ec735 -->

A withdrawal rule has a name, not a face. Everyone knows that Bengen pays a fixed amount, that VPW pays a percentage, that guardrails cut when the current rate runs away. Far fewer people know what a life lived under each of those rules looks like. How much you get in year one, how much in year ten, whether the cut lands all at once or stretches over fifteen years, and what is left in the account the day you stop counting.

That is the only thing this page sets out to show. It freezes everything that can be frozen, takes three retirements that actually happened, and writes down what each rule would have paid, year after year. No rule wins. Failure rates live elsewhere, in the pages that simulate thousands of futures ([[monte-carlo-strengths-and-limits]], [[failure-probability]]). Here there is no chance at all, only three sequences of returns that happened and seven ways to get through them.

## The setup

One household, seven times. It starts with **EUR 600,000** and wants **EUR 2,000 a month**, EUR 24,000 a year, which makes an initial withdrawal rate of 4.0%, exactly Bengen's number ([[the-4-percent-rule]]). It draws no pension, earns nothing on the side, pays no tax and keeps no cash. Every one of those simplifications flatters the seven rules the same way, so none of them changes the comparison. They only make each number a little more generous than real life ([[cash-buffer]]).

The portfolio is the standard 60/40: 60% S&P 500, 40% five-year Treasuries, rebalanced every January, deflated by CPI. It is not a recommendation. It is the yardstick the whole withdrawal literature was written on, from Bengen to the Trinity study to the modern guardrails ([[the-trinity-study]]), which makes it the fairest ground for comparing rules that were calibrated on it.

::: cle Everything is in constant euros
Every figure on this page, income and capital alike, is stated in today's purchasing power. Inflation has already been taken out. So a flat line does not mean an identical check every year, it means an identical standard of living every year: the nominal check was indexed along the way. That is the convention of the whole book and it is indispensable here, because the decade from 1973 to 1982 would be unreadable otherwise ([[inflation-and-withdrawal-rates]]).
:::

The plan horizon is **40 years**, always. That detail matters more than it looks. Two of the seven rules reason on the horizon that is left: amortization spreads the capital over the remaining years, and the risk-based guardrails watch the rate still sustainable to that horizon. Telling them that retirement conveniently stops where the data stops would make them far more generous than they are. So the plan runs forty years in every case, and the page shows only the years the available history actually covers. Those two rules also need a return assumption to price their payment; they get a generic assumption, 4.5% real arithmetic at 10% volatility, never the returns of the era being replayed. No retiree in 1973 knew what was coming, and a rule fed from the rear-view mirror is no longer the rule the literature describes.

The seven rules, from the most rigid to the most closely tied to the portfolio:

| Rule | What it does | Article |
|---|---|---|
| Fixed (inflation-adjusted) | The same real amount every year | [[fixed-inflation-adjusted-withdrawal]] |
| Flex −10% | The fixed rule, minus 10% when the portfolio is 20% below its peak | [[flexibility-in-practice]] |
| Guardrails (GK) | ±10% when the current rate leaves a corridor around the initial rate | [[guyton-klinger]] |
| Risk-based guardrails | Same corridor, centered on the rate still safe over the remaining horizon | [[morningstar-guardrails]] |
| Bounded % | A percentage of the portfolio, never moving more than +5% or −2.5% a year | [[floor-and-ceiling]] |
| Amortization (ABW) | The payment that runs the capital exactly to zero over the remaining horizon | [[amortization-based-withdrawal]] |
| % of portfolio (VPW) | A fixed share of what is left, every year | [[vpw]] |

## January 1973: the crisis first

::: figure replay-marche-1973
The portfolio itself, with not a single withdrawal taken. The 1973 to 1974 crash wipes out 38% in real terms, then the decade's inflation eats what is left, so the capital is still a third below its starting value nine years in. It takes until 1983 to get durably back above where it began. Over the forty years, though, the portfolio returns 4.7% real a year.
:::

This is the sequence all of these rules were invented for ([[sequence-of-returns]]). The average return over the forty years is fine. It is the order that kills. A withdrawal taken during a crash sells a much bigger slice of the portfolio, since the portfolio has shrunk while the amount withdrawn has not. Those shares leave at the bottom and never take part in the recovery. The same crash fifteen years later would find a portfolio that has already banked its best years and has far fewer years left to fund.

::: figure replay-revenus-1973
Seven lives inside the same crisis. The fixed rule is the straight line: it pays EUR 24,000 for forty years without flinching. The other six cut, and the shape of the cut is what tells them apart. The bounded percentage steps down over some fifteen years and the household barely feels the slide. Both guardrails dive fast and deep, then climb very high. VPW tracks the market with no shock absorber, year after year.
:::

| Rule | Average income | Variability | Worst year | Lean years | Ending capital |
|---|---|---|---|---|---|
| Fixed (inflation-adjusted) | 24.0 | 0% | 24.0 | 0 | 19 |
| Flex −10% | 22.0 | 4% | 21.6 | 33 | 344 |
| Guardrails (GK) | 23.5 | 33% | 14.2 | 23 | 781 |
| Risk-based guardrails | 25.6 | 39% | 12.8 | 19 | 762 |
| Bounded % | 21.3 | 16% | 16.0 | 28 | 575 |
| Amortization (ABW) | 30.9 | 35% | 15.4 | 13 | 0 |
| % of portfolio (VPW) | 24.5 | 34% | 12.5 | 19 | 734 |

Income in EUR thousands a year, ending capital in EUR thousands, all in constant euros. Variability is the coefficient of variation of annual income, its spread measured against its own mean, and it is the yardstick for how much the rule moves the household's life around. Lean years count the years lived below the EUR 24,000 planned, out of forty.

The fixed rule held. It paid what it owed forty years running and ends with EUR 19,000 in the account, three percent of what it started with. One more market year in the wrong direction and that column would be telling a story of ruin. That is the true nature of this rule, and the right word is cliff rather than risk. Everything is fine until the second nothing is.

The two guardrails are the exact opposite. They saved the capital, more than EUR 750,000 at the finish, but they had the household living on EUR 14,200 and then EUR 12,800 through most of the 1970s and 1980s. Those are cuts of 40 to 47%, held for more than a decade. The success rate of those columns is perfect and the standard of living they describe is anything but ([[guyton-klinger]]).

In between sit two honest and very different answers. The bounded percentage goes all the way down to EUR 16,000, but it takes sixteen years to get there, in steps of 2.5% a year. Nobody lives a 2.5% cut as a crisis; you live it as a slightly tight year, and that is exactly what the rule is for ([[floor-and-ceiling]]). Amortization pays the highest average income of the seven, EUR 30,900 a year, and it ends at zero. That is not a failure, it is the definition: every year it priced the payment that exhausts the capital over the remaining horizon, and the horizon ended ([[amortization-based-withdrawal]]).

## January 1985: the opposite problem

::: figure replay-marche-1985
Forty years of tailwind. The portfolio takes 1987, 2000, 2008 and 2022 from a comfortable position and ends up twelve times larger in real terms, 6.5% a year. The first decade returns 8.4% a year, and that decade decides everything that follows.
:::

Nobody runs short of money in this retirement. So the interesting question flips, and it matters at least as much as the previous one. Who spent what the portfolio could obviously afford, and who spent forty years living small on a fortune they never touched?

::: figure replay-revenus-1985
Seven lives inside the same tailwind. The fixed rule and its flexible variant stay glued to their EUR 24,000 line while the other five climb. Amortization reaches EUR 82,000 at the 2022 peak. Guardrails, the bounded percentage and VPW settle around EUR 47,000 on average, twice the original plan.
:::

| Rule | Average income | Variability | Worst year | Lean years | Ending capital |
|---|---|---|---|---|---|
| Fixed (inflation-adjusted) | 24.0 | 0% | 24.0 | 0 | 3,829 |
| Flex −10% | 23.9 | 2% | 21.6 | 2 | 3,843 |
| Guardrails (GK) | 46.6 | 26% | 24.0 | 0 | 1,656 |
| Risk-based guardrails | 32.9 | 16% | 21.6 | 7 | 3,053 |
| Bounded % | 47.2 | 25% | 24.0 | 0 | 1,571 |
| Amortization (ABW) | 58.3 | 23% | 29.1 | 0 | 0 |
| % of portfolio (VPW) | 46.4 | 22% | 24.0 | 0 | 1,448 |

The fixed rule dies with EUR 3.83M of constant euros in the account, after forty years of living on EUR 2,000 a month. It multiplied its capital by six and did nothing with it. Nobody calls that a failure, because the word ruin does not apply, and yet the household gave up half of its retirement. Amortization lived the same period on EUR 58,300 a year on average, two and a half times better, and finishes at zero because that was the deal.

Here is the hidden price of rigidity, and no success rate will show it to you. A rule that never goes up is a rule that will not know how to spend a good market. And since most retirements land on markets that turn out fine ([[deciding-under-uncertainty]]), that is the common case, not the exotic one.

Two useful caveats. The risk-based guardrails stay well below their cousins, at EUR 32,900, because their corridor is capped at 150% of the planned level on this page; without that cap the rule climbs without limit as the horizon shortens, which is a known pathology and not a performance ([[morningstar-guardrails]]). And the −10% flex rule cuts only twice in forty years, a reminder that flexibility conditioned on a 20% drop almost never fires in a bull market. It is insurance, not an income strategy.

## January 2000: the one you are living

::: figure replay-marche-2000
The lost decade, then the catch-up. The start lands on the peak of the dot-com bubble, the first decade returns exactly zero percent real a year, 2008 arrives on a portfolio that had never recovered from 2000, and 2022 hits a thinner cushion than planned. It took thirteen years to get back to the real value of the starting capital.
:::

This retirement is not over. Twenty-six of the forty planned years have run, and that is exactly what makes it instructive: it reads the way yours might read today, halfway through, with no idea how it ends.

::: figure replay-revenus-2000
Seven lives inside the lost decade. Six rules out of seven have already had the household living below its plan, most of them for more than twenty years. Only the fixed rule held its line, and it has EUR 271,000 left to fund fourteen years at EUR 24,000.
:::

| Rule | Average income | Variability | Worst year | Lean years | Ending capital |
|---|---|---|---|---|---|
| Fixed (inflation-adjusted) | 24.0 | 0% | 24.0 | 0 | 271 |
| Flex −10% | 21.9 | 4% | 21.6 | 23 | 387 |
| Guardrails (GK) | 18.9 | 10% | 17.5 | 24 | 552 |
| Risk-based guardrails | 19.7 | 24% | 15.7 | 21 | 589 |
| Bounded % | 19.5 | 11% | 16.9 | 25 | 511 |
| Amortization (ABW) | 23.7 | 12% | 18.2 | 16 | 316 |
| % of portfolio (VPW) | 19.1 | 12% | 14.7 | 25 | 568 |

Lean years are counted here out of twenty-six, not forty.

Look at the fixed rule's row and do the arithmetic the household would do. It has EUR 271,000 left and fourteen years of plan, a current withdrawal rate of 8.9%. Not one of those fourteen years is funded by anything but luck. The columns that cut have roughly twice the capital for the same distance to go, and they paid for it by living on EUR 19,000 for twenty years. Nobody can say today who was right, and that is the heart of the matter.

::: attention What the last column does not say
A large ending capital is neither a win nor a loss until you know what you wanted it for. For a household with no heirs and nothing to pass on, the EUR 3.83M of the fixed rule in 1985 is a dead loss of standard of living. For a household that wants to leave something behind, or that fears an expensive spell of care at the end of life, it is exactly the goal. The right rule turns on a question that is not financial.
:::

## Three retirements are not a distribution

Three start dates are three draws. A rule that got through 1973 and 2000 is not proven safe; it is merely plausible on three paths from a single country, the one whose century was the kindest of all ([[anarkulova-cederburg]]). That is exactly why simulators draw thousands of futures instead of three ([[historical-vs-parametric]]).

These three were also chosen, which is a form of bias worth stating out loud. The available US history holds thirty-three forty-year windows, and the median window returns 5.5% real a year. 1973 returns 4.7% and 1985 returns 6.5%, so the two periods bracket that median case instead of flattering it. They do not exhaust the range of what can happen. There are sequences worse than 1973 elsewhere in the world, and Japan after 1990 is the most brutal example ([[hyperinflation-and-extremes]]).

So read this page for how the rules behave, not for a ranking. The behavior is stable, and it shows up in one era after another: the fixed rule never moves and either breaks all at once or not at all; the guardrails dive early, deep and for a long time; the bounded percentage glides instead of falling; amortization follows the market, rises with age and ends at zero by construction; VPW takes everything on the chin, with no shock absorber and without ever running dry.

## What to take away

Three things hold up after these three retirements, and none of them shows in a failure probability.

The first is that income stability and capital safety are the same resource, spent in two different places ([[withdrawal-strategies-overview]]). The fixed rule bought forty years of peace in 1973 and paid for it with the EUR 19,000 it had left at the end. The guardrails bought EUR 780,000 of capital and paid for it with a decade of lean years. No rule creates comfort. They move it around.

The second is that the shape of a cut counts as much as its depth. The bounded percentage and the guardrails both come down toward EUR 15,000 to 16,000 in 1973. One takes sixteen years, in steps nobody feels; the other gets there in seven, in jumps of 10%. The second is far harder to live with, and the usual metrics give it almost no credit ([[the-psychology-of-spending]]).

The third is that a rule that never goes up is a choice, not prudence. Deciding never to raise your standard of living is a bet on 1973 in a world where most retirements look like 1985. If that bet is yours, take it with your eyes open. If the real goal is not to run short, there are cheaper ways to get there, starting with an explicit floor under the cuts ([[choosing-your-strategy]]).
