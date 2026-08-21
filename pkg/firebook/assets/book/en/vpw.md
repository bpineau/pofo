# VPW, the Bogleheads' variable percentage withdrawal
<!-- source: vpw @ e5ad4f417a13 -->

VPW ("Variable Percentage Withdrawal") is the Bogleheads community's answer to a precise question: how do you spend a portfolio with no forecast, no risk of ruin, and no dying on a pile of gold? Their solution rests on one idea. You withdraw a percentage of the portfolio, as with a fixed percentage ([[fixed-percentage]]), but that percentage rises with age along a table computed once and for all. It is about 3.9% at 40, 4.8% at 65, 7% at 80, up to 100% in the final year.

That rising percentage is not a hack. It is the amortization formula of a loan, applied backwards to your portfolio. VPW is the historical bridge between the proportional family and the actuarial one ([[amortization-based-withdrawal]]), and one of the most widely used strategies in the real FIRE world. This article lays out its exact mechanics, enough to recompute the table yourself. It sets out its design choices, deliberate and debatable, starting with its fixed assumed returns. It describes its remarkable properties and its pathology, the same standard-of-living swings as any percentage rule, plus an end-of-life hump. Finally it presents the guardrail its authors impose, the loss tolerance test that almost everyone skips, then where it stands against modern ABW and how to put it through a simulator.

::: cle The idea in one sentence
Every year, VPW withdraws the percentage that would exhaust the portfolio exactly over the remaining years, assuming future returns equal to reference values fixed in advance. It is a loan payment, recomputed each year on the current capital and the remaining horizon. Young, the horizon is long and the percentage is low. Old, the horizon shortens and the percentage climbs. The portfolio is spent deliberately, never exhausted early, never hoarded by accident.
:::

::: admin How to run it
- **The rate applies to the current portfolio minus the pension bridge.** The bridge, meaning the missing pension years discounted and held in bonds, is set aside **before** you apply the percentage. This is the most common VPW implementation error: applying the table to the whole portfolio, when that portfolio also has fifteen or twenty years of pension to manufacture. The rest is an ordinary percentage, with no memory of the starting capital.
- **Indexation.** Nothing to write, as with any percentage of the current portfolio.
- **Frequency.** Annual: read the table, multiply.
- **Parameters.** Not thresholds but two numbers. The assumed return g (5.0% real for stocks, 1.9% for bonds, blended in proportion to your allocation) and the horizon n (100 minus your age). They are carved in by doctrine, which is a deliberate choice and not an oversight: the table is not renegotiated. The one defensible adjustment is to shave a point off g in an expensive market.
- **Floor.** External and mandatory, checked by the loss tolerance test: the income served under a "stocks −50%" assumption must stay above the floor.
- **In your head.** w = g / (1 − (1+g) to the power −n), that is, a loan payment. One implementation detail is worth knowing, because it explains most of the differences between two tables: depending on whether the withdrawal is taken at the start or the end of the year, you do or do not divide the result by (1 + g), which moves the rate by about 4% in relative terms. In practice, read the official table rather than recomputing it, but knowing where it comes from lets you check a suspicious figure.
:::

## Where it comes from, and the philosophy

VPW was born on the Bogleheads forum in the early 2010s, the work of the contributor "longinvest". Its doctrine rests on three refusals, very much in the Boglehead spirit. A refusal of ruin first, hence no fixed amount ([[fixed-inflation-adjusted-withdrawal]]). A refusal of the accidental bequest next, because the cautious 4% dies rich three times out of four, for want of having lived ([[spending-in-retirement]]). And a refusal of forecasting: no expected returns recomputed every year, no CAPE, no parameters to argue about, only one table, published and carved in. The strategy comes with a spreadsheet (the "VPW worksheet") maintained by the community, which also handles pension bridges, as we will see. It is one of the most polished free tools in the FIRE world.

## The exact mechanics: the loan formula, reversed

At the heart sits the annuity formula, the one behind every loan payment. For a capital C, a horizon of n remaining years and an assumed growth rate g, the constant payment that exhausts C in exactly n years is:

> withdrawal = C × g / (1 − (1 + g)^(−n))

VPW tabulates that ratio (withdrawal divided by capital) for each age, with n running from your current age to 100. The rate g is fixed once and for all, per asset class. The current table uses 5.0% real for global stocks and 1.9% real for bonds, blended in proportion to your allocation. A 60/40 therefore assumes about 3.8% real. Here is an extract of the table's logic for a 60/40:

| Age | Years left | VPW % |
|---|---|---|
| 40 | 60 | ~3.9% |
| 50 | 50 | ~4.1% |
| 65 | 35 | ~4.8% |
| 75 | 25 | ~5.7% |
| 85 | 15 | ~7.9% |
| 99 | 1 | 100% |

::: figure vpw-table
On top, the table: the rate starts just above the assumed return, stays almost flat for twenty-five years, then takes off as the horizon shortens. Below, the life that table produces for a household retiring at 40 with $1M, if the market delivers exactly the assumed returns. Income is perfectly flat and capital melts to zero at 100: that is the annuity property, not an accident. The dotted line shows the rule's other face. A 30% crash at 70 takes income from $38.5k down to $27.0k **and it never comes back**, because VPW smooths nothing: it recomputes.
:::

Two properties of the formula are worth pausing on. First, over a long horizon the percentage tends toward g itself. At a 60-year horizon you withdraw barely more than the assumed growth and capital is nearly preserved. The VPW of someone retiring at 40 is therefore, in practice, an improved fixed percentage, and its climb with age only starts to bite after 65 or 70. Second, that final climb is the deliberate consumption of capital. Dying at zero at 100 is a design choice, not an accident. So it calls for a treatment of longevity risk: VPW doctrine recommends annuitizing part of the portfolio around 80, to cover the years beyond the table ([[annuities-and-safety-first]], [[horizon-and-life-expectancy]]).

**The pension bridge** is the worksheet's other practical innovation. Before your pensions start, VPW sets aside, virtually, the capital needed to "manufacture" the missing pension through the bridge years, for example 15 years × $15,000 for a pension starting at 67. It invests that in bonds and applies the percentage only to what is left. It is the uncovered phase / covered phase split of [[horizon-and-life-expectancy]], made operational: the permanent need is amortized, the temporary need is provisioned.

::: figure vpw-pont
The bridge, on the household in the example below. The lower band is the income that does not exist yet: twenty slices of $21.6k drawn from the $356k of bonds set aside, then the pension itself, which takes over without the household noticing. The upper band is VPW applied to the rest. **The top edge is flat: at 67 income does not step up, it simply changes hands.** Without the bridge, the same household would live twenty years on $64.3k before jumping to $85.9k, which is too little while it is young and too late once it is not.
:::

## What VPW gets right, and what it accepts getting wrong

**The wins.** VPW inherits every virtue of the percentage family ([[fixed-percentage]]): running the capital to zero is impossible, it is countercyclical, it self-corrects when return assumptions turn out wrong. To that it adds awareness of the horizon. Where the fixed percentage hoards forever, VPW dares to spend, and its average lifetime consumption ranks among the highest of any rule. It is the anti-"dying rich" strategy par excellence. Its governance, finally, is remarkable: a printed table, one ratio a year, no parameter to reopen. The rule thus outlives both its author and the years of panic ([[the-psychology-of-spending]]).

**The accepted misses.** The first belongs to the whole family: income follows the portfolio. VPW smooths nothing by construction, because the doctrine treats smoothing as debt in disguise. A 30% fall in the portfolio therefore means a 30% fall in income the following year. Hence the guardrail the doctrine imposes and everyone skips, the **loss tolerance test**. Before adopting VPW, compute your income under a "stocks −50%" assumption, which the worksheet displays at all times, and check that it still covers your floor ([[how-much-you-need]]). If it does not, VPW itself tells you to cut the equity share or to cover the floor another way, with a pension or an annuity. This is a strategy that demands either an external floor or genuine elasticity, exactly like its proportional parent.

::: figure vpw-test-de-perte
The test the doctrine imposes, on the couple covered below. On the left the pension bridge sleeps in bonds: the crash bites only the VPW sleeve and the household stays above its comfort level. On the right, with no bridge, the same rule serves more for twenty years, then the same crash drops it below comfort. It is the two-bar demonstration that the bridge is not a refinement of the worksheet but the rule's condition of admissibility.
:::

**The debatable miss: fixed assumed returns.** The table's g, 5% real for stocks, is a very long-run historical average. It is worth the same in 2013, in 2021 (CAPE 38!) and in 2026. That is a coherent philosophical choice, the choice not to forecast, but an expensive one when the market is dear. VPW then withdraws more than valuations promise ([[valuations-and-cape]], [[expected-returns]]), and the adjustment comes after the fact, as income falls when the disappointment lands. The ABW/TPAW family makes the opposite choice, plugging in current expected returns, CAPE included: more accurate in expectation, but more dependent on models. That is the dividing line between the two cousins ([[amortization-based-withdrawal]]).

::: science VPW and ABW: the same formula, two epistemologies
Mathematically, VPW is an ABW with constant assumed returns and no fine present value of future flows: the same reversed annuity. The divergence is epistemological. VPW bets that the average retiree will go less wrong with a carved table than with annual forecasts, which is behavioral robustness. ABW bets that the information in current prices beats a secular average, which is conditional accuracy. The research leans toward ABW on the numbers, but on simulations where the rule is applied flawlessly. Forum wisdom leans toward VPW among real humans. The honest choice depends on who will be running the rule twenty years from now ([[couples-and-family]], [[choosing-your-strategy]]).
:::

## Who it suits, and the settings that matter

The ideal VPW profile combines three traits: a floor covered outside the portfolio, genuine elasticity of living standard above that floor, and a taste for auditable simplicity. The floor can come from a pension already running or bridged, or from an annuity; the loss tolerance test then passes naturally. The simplicity comes from the table and the worksheet, nothing else. That is precisely the typical Bogleheads retiree. It is also the covered phase of any FIRE plan ([[horizon-and-life-expectancy]]): once pensions are running and cover the floor, VPW on the remaining portfolio is hard to beat.

In the uncovered phase of an early retirement, VPW needs two adjustments. The first is the worksheet's pension bridge, and it is mandatory. Without it, the percentage applies to a capital that also has fifteen years of pension to manufacture, and the loss test fails almost every time. The second, in an expensive market, is a manual haircut on g: using 4% instead of 5% for stocks amounts to folding in the CAPE anchor roughly, without betraying the spirit of the table.

::: astuce Putting VPW through a simulator
The official table lives in the Bogleheads worksheet. In a general-purpose simulator, two rules approximate it, to be tried one after the other, since only one spending policy drives a plan at a time.

- **A constant percentage of the portfolio.** That is VPW while the horizon is still long, the flat zone of the table. The approximation is excellent before 60 or 65.
- **Withdrawal amortized over the remaining horizon** (ABW/TPAW). That is the full reversed annuity: rising percentage, exact horizon, future flows discounted and current expected returns, valuation anchor included if the tool offers one. That is modern VPW, the one in TPAW Planner ([[amortization-based-withdrawal]]).

Either way, judge on the standard of living served year after year, never on the probability of ruin, which every proportional rule drives to zero by construction ([[flexibility-in-practice]]).
:::

::: exemple One FIRE couple's VPW, bridge included
Nora and Malik, 47, $1.6M, 60/40, pensions estimated at $21,600 a year from 67, floor $38,000, comfort $52,000. The bridge provisions the twenty missing pension years, discounted at 1.9% real, which is $356,000 held in bonds. VPW applies to the rest, $1,243,000 at 47, the 60/40 table at 4.0%, that is $50,000. Add the bridge slice, $21,600 a year, for an initial income of $71,600. Now the tolerance test, stocks −50%. The VPW sleeve falls to $870,000 and its share of income to $35,000, but the bridge does not move, since it sleeps in bonds. The household therefore still lives on $56,600, above its comfort level and far above its floor. Without the bridge, the same rule would have served $64,300 for twenty years, then $85,900 all at once at 67, and the crash would have cut it to $45,000, below comfort. Twenty years later, pensions running, VPW draws 5.0% of what is left for comfort and for projects. Two regimes, one table, zero forecasting. That is well-built VPW.
:::

## The essentials

- VPW is a loan annuity reversed. Each year it applies the percentage that would exhaust the portfolio over the remaining years, at fixed assumed returns (5% real stocks, 1.9% bonds). That percentage rises with age: about 3.9% at 40, 4.8% at 65, 100% at 99.
- It inherits from the fixed percentage (capital can never run out, countercyclical, self-correcting) and adds awareness of the horizon: deliberate spending, near-zero bequest, and the most generous average consumption in the field.
- Its requirements: the loss tolerance test (income under "stocks −50%" at or above the floor, never to be skipped), the pension bridge in the uncovered phase, and annuitization around 80 for the longevity beyond the table.
- Its dividing line with ABW: a carved table (behavioral robustness) against current returns (conditional accuracy). Same formula, two bets on whoever will run it.
- Two rules approximate it in a simulator, to be tested separately: the constant percentage (VPW at a long horizon) and withdrawal amortized over the horizon (full dynamic VPW). Judge them on the standard of living served, and in an expensive market, shave a point off g.

---

## Going further

- Bogleheads wiki, "Variable percentage withdrawal (VPW)" and the forum's "VPW forward test" thread: the doctrine, the table, the worksheet, and ten years of documented live execution.
- The VPW worksheet (official, free): pension bridges and the loss test built in.
- Early Retirement Now, part 11 (VPW scored against the other rules) ([[the-ern-series]]).
- In this book: [[fixed-percentage]] (the parent), [[amortization-based-withdrawal]] (the modern cousin), [[annuities-and-safety-first]] (the end-of-life complement), [[choosing-your-strategy]] (the final call).
