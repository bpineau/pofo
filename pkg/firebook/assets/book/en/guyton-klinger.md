# Guyton-Klinger: the original guardrails, their power and their limits
<!-- source: guyton-klinger @ 33aecfd5ad24 -->

In 2006, the financial planner Jonathan Guyton and the computer scientist William Klinger published the most influential dynamic withdrawal rule ever written. Their set of decision rules promised to lift the initial withdrawal rate to 5.2 or 5.6%, against Bengen's 4%, in exchange for a few adjustments to the standard of living. The success was enormous. "Guardrails" became the common name of a whole family of strategies, and the rule is still one of the most widely used by US advisers.

Then modern research, ERN first among them, opened the hood and found the hidden flaw. In bad vintages the cuts repeat year after year and real income sinks 30 to 45% for decades. The portfolio's survival was bought with a failure of the standard of living, and the tables in the original article never showed it. This article tells both halves of the story: the exact rules, which get quoted far more often than they get stated correctly; where the apparent power came from, what exactly goes wrong, and the proof of it; and the modern fixes, the floor first among them, plus the parameters worth defending if you adopt the rule anyway.

::: cle The principle in one sentence
Guyton-Klinger is a fixed inflation-adjusted withdrawal that watches itself. As long as the current withdrawal rate (the withdrawal divided by the portfolio) stays inside a corridor of plus or minus 20% around the initial rate, you live exactly as you would under Bengen. If it breaks out on the high side, the portfolio has fallen too far, and you cut the withdrawal by 10%. If it breaks out on the low side, the portfolio has run away, and you raise the withdrawal by 10%. The greatness of the design is that it turns vague flexibility ("I'll be careful if things go badly") into written machinery. Its limit is that nothing bounds the **number** of cuts.
:::

::: admin How to run it
- **What you measure moves, the thresholds do not.** That is the subtlety of the rule, and the source of almost every botched replication. Each year you recompute the **current** rate, this year's planned withdrawal divided by today's portfolio, and you compare it with two thresholds fixed once and for all at the start: 1.2 times the initial rate and 0.8 times the initial rate. The 10% adjustment then applies to the **current withdrawal**, not to the initial one, so cuts compound: two cuts make −19%, three make −27%.
- **What the band means in real money.** Plus or minus 20% of a rate tells you nothing. At unchanged real spending, the lower guardrail fires when the real portfolio has fallen to 83% of its starting value, and the upper one when it has climbed to 125%. That is the quantity the rule really watches, and that is the one to form an opinion about.
- **Indexation.** The amount is re-indexed to inflation every year, except in the year after a negative portfolio return when the current rate sits above the initial rate (rule 2). The article also caps indexation at 6% a year.
- **Frequency.** Annual, on a fixed date. The date is not neutral: with a threshold rule, two identical households that check two months apart do not make the same decisions, and the gap can be measured ([[the-annual-review]]).
- **The thresholds are not derived, they are calibrated.** Nowhere does the 2006 article justify the 20% or the 10%. They are the values that came out of simulations on 1928 to 2004 and 1973 to 2004 data, three allocations, 40-year horizons. So they can legitimately move: a band of 15 to 25%, adjustments of 5 to 10%, with the research leaning toward smaller and more frequent steps.
- **Floor.** Essential, and not a comfort option: without it, nothing bounds the number of cuts, which is exactly the pathology described below. Planning value, 75 to 80% of the initial withdrawal.
- **In your head.** If the withdrawal divided by the portfolio is above 1.2 times the initial rate, multiply the withdrawal by 0.9 without going below the floor. If it is below 0.8 times the initial rate, multiply by 1.1. Otherwise, do nothing.
:::

## The exact rules, for once

The 2006 article ("Decision Rules and Maximum Initial Withdrawal Rates") defines four rules, almost always truncated when they are quoted. Here they are in full, because the details are the strategy.

**1. The Portfolio Management Rule.** This is the order of sales. You fund the withdrawal first from cash and from the gains of whatever outperformed. You never sell stocks in a year they fell: those sales wait for the recovery, while bonds and cash bridge the gap. It is a small procedural buffer built into the rule ([[cash-buffer]]), often forgotten when people replicate the strategy, and part of the measured benefit comes from it.

**2. The Withdrawal Rule.** The inflation adjustment is skipped in the years that follow a negative portfolio return, if the current withdrawal rate sits above the initial rate. This is exactly the indexation freeze after a red year ([[fixed-inflation-adjusted-withdrawal]]), in conditional form. The article also caps indexation at 6% a year.

**3. The Capital Preservation Rule, the lower guardrail.** If the current withdrawal rate goes above 120% of the initial rate (say an initial rate of 5% and a current rate above 6%), the withdrawal is cut by 10%. It stops applying in the last fifteen years of the horizon. Cutting at 82 to protect a portfolio that has only ten years left to last would make no sense.

**4. The Prosperity Rule, the upper guardrail.** If the current rate drops below 80% of the initial rate, the portfolio has run away, and the withdrawal is raised by 10%. It is the sister of Kitces's ratchet, but reversible: a later cut can take the raise back.

The whole thing hangs together and can actually be run. Every January 1, one ratio to compute, three comparisons, at most one adjustment of 10% either way. The article's promise was a big one. With these rules, a portfolio at 65% stocks supported an initial rate of 5.2 to 5.6% with 99% "success" over 40 years. That is a point and a half better than Bengen, a multiple of 18x or 19x instead of 25x ([[how-much-you-need]]). The enthusiasm is easy to understand: on the face of it, five years of work saved.

## Why it looked so strong, and where the flaw sits

Where does that free point and a half come from? From three sources, and they are not equally respectable. The first is legitimate. Rules 1 and 2 are genuine improvements, and nearly painless ones: the order of sales and the indexation freeze are worth about 0.3 to 0.5 point together, as later research confirmed. The second is an artifact of its time. The 2006 simulations ran on friendly US data and on horizons of 40 years at most ([[simulator-traps]]). The third is the construction flaw, and it deserves a hard look. The article's "success" measures the survival of the portfolio, not the survival of the standard of living, and the cuts of rule 3 are **unlimited in number**.

Run the pathology on the textbook vintage ([[the-trinity-study]], 1966). A hostile regime sets in ([[market-regimes]]). The portfolio falls, the current rate crosses 120% of the initial rate, and the 10% cut lands. The next year the portfolio has fallen again, because bear markets persist. The current rate crosses the threshold again, so another cut lands. And so on.

In ERN's simulations (parts 9 and 10 of the series, [[the-ern-series]]), retirements starting between 1966 and 1969 take cascades of cuts under Guyton-Klinger that push real income 35 to 45% below its starting level and hold it there for ten to twenty years. The portfolio, meanwhile, survives beautifully. And it survives precisely because the retiree was put on a diet for two decades. The success rate reads 99%; the life lived reads a generation of lean years. The psychological asymmetry is cruel. Every 10% cut arrives after a down year, when morale is already at its lowest, and the prosperity rule hands the raises back only years later.

::: figure gk-cascade-1966
The pathology, simulated on the 1966 vintage with an initial rate that is perfectly reasonable at 4.3%. With no floor, five cuts follow one another as the thresholds get crossed, and real income ends at $18.5k, down 57%, where it stays four years before climbing slowly back. With a floor at 78% of the starting level, the descent stops at $33.5k and the household spends the crossing, the years inside the bear market, right there. The capital never worried for a moment: with no floor it ends bigger than it started. **The success rate was measuring the survival of the portfolio, not the survival of the standard of living.**
:::

Modern research does not conclude that Guyton-Klinger fails. It concludes that its failure number cannot be compared with the failure number of a fixed rule. A 1% failure probability under GK means "even cutting without limit, 1% of the paths run out". That is a different event, and a far worse one, than a 5% failure under Bengen. Comparing the two failure rates without reading the income delivered is the central mistake of this whole part of the book ([[withdrawal-strategies-overview]]). Which puts one simple demand on whatever tool you use. As soon as a dynamic rule is switched on, failure probability is no longer enough: you have to read the distribution of the spending actually delivered, year by year, beside it. A simulator that returns nothing but a success rate for a rule with unlimited cuts is hiding half the answer ([[simulator-traps]]).

::: attention The initial rate was the real culprit
The pathology is made worse by a marketing choice in the original article: that initial rate of 5.2 to 5.6%. Starting that high means entering the corridor already close to the lower guardrail, and one mediocre decade is enough to set off the cascade. The same rules with an initial rate of 4 to 4.5% cut rarely, and briefly. The lesson holds for every flexible rule. Flexibility lets you start a little above Bengen, not a point and a half above ([[flexibility-in-practice]]). Spend the whole benefit of your flexibility up front and you no longer have flexibility, only a deferred austerity program.
:::

## The modern fixes: put a bottom under the descent

Posterity repaired Guyton-Klinger in three ways, from the simple patch to the outright replacement.

**The floor (the fix you cannot skip).** It forbids the cuts to push income below a set percentage of the starting level. The current planning value is 75 to 80%, to be lined up with your real floor once you have established it ([[how-much-you-need]]). The effect is clear: the descent is bounded, and the generation of lean years becomes, at worst, a few years down 20 to 25%. The price is honest and unavoidable. Bounding the descent recreates failure: if the floor itself is unsustainable in a given path, the portfolio runs out. The failure number then becomes a real number again, comparable with any other, and that is exactly as it should be.

Remember that the floor is not one more rule laid on top of the guardrails. It is a parameter of the rule itself. A complete guardrails spending policy is described by four numbers, not two: the initial rate, the width of the corridor, the size of the adjustment and the floor. A tool that does not expose the fourth is not simulating the rule you will actually run.

**Bands and doses.** The second family of settings: tighten or loosen the corridor (plus or minus 20% as standard) and the size of the adjustments (10% as standard). Narrower bands smooth the path, with adjustments that come more often and bite less. Smaller cuts (5%) repeated hurt less than 10% cuts spaced further apart. The research leans toward more frequent and gentler adjustments, which move the rule closer to continuous smoothing ([[fixed-percentage]], the Yale rule).

**What came next: risk-based guardrails.** This is the conceptual replacement, from Kitces and Derek Tharp, then industrialized by Morningstar. Instead of watching the current withdrawal rate, a crude thermometer that ignores the remaining horizon and the pensions still to come, you watch the recomputed success probability of the plan, and you adjust when it leaves its corridor. Same architecture, far better indicator. A current rate of 6% at 80 with a pension running is nothing to worry about; the same rate at 52 is serious. The risk-based guardrail knows that, the 2006 one does not. It is the state of the art of the family, and it has its own article ([[morningstar-guardrails]]).

## If you use it: the parameters worth defending

Some long FIRE plans still pick the 2006 version, for a good reason. You can run it by hand, where a risk-based guardrail forces you to restart a simulator at every review ([[morningstar-guardrails]]). Here is the configuration the post-ERN literature supports:

- **An initial rate of 4 to 4.5%**, not 5.5. Flexibility buys about 0.5 point above the equivalent fixed rule, no more ([[flexibility-in-practice]]).
- **A corridor of plus or minus 20% with 10% adjustments** (the standard), or 15% and 5% for the gentle version.
- **A floor at 75 to 80% of the initial withdrawal**, lined up with your real floor: not negotiable.
- **The indexation freeze after a red year and the order of sales** (rules 1 and 2): keep them, that is the free part.
- **No more cuts at the very end of the horizon** (the original article said so), and none once a pension covers the floor ([[pensions-and-other-income]]).
- **The review is annual, on a fixed date**: the rule is computed on January 1, not at every scare.

::: exemple The cascade, with and without a floor
The plan: $1.3M, initial rate 4.3% ($55,900), corridor plus or minus 20% (thresholds at 3.44% and 5.16%), 10% cuts. A hostile regime sets in and the real portfolio slides to $950,000 in three years. In year 3 the current rate hits 5.9%, above 5.16%, so the withdrawal is cut to $50,300. In years 4 and 5 the bear market persists and two more cuts land, taking income to $40,700, down 27%. With no floor, the 1966 path keeps going: five cuts, income around $33,000 (down 41%) for a decade. With a floor at 78% ($43,600), the third cut stops at the floor. Income then rides out the crossing down 22%, and the plan's failure probability climbs from about 1% to about 4%. That is the price, honest and visible, of refusing the unlimited diet. Redo this calculation on your own plan, varying nothing but the floor. Read the answer on the income delivered year by year as much as on the failure probability, or the version with no floor will win every time.
:::

## The essentials

- The four rules of 2006: the order of sales, the conditional indexation freeze, a 10% cut when the current rate goes above 120% of the initial one, a 10% raise when it drops below 80%. It is runnable machinery, and it invented the guardrails.
- The promise (5.2 to 5.6% initial, 99% success) rested on a flaw: cuts with no limit on their number, which bad vintages exploit until real income sits 35 to 45% below plan for decades. GK's success rate never compares with a fixed rule's unless you read the income delivered.
- The fixes: a floor at 75 to 80% (which recreates honest failure, and that is the point), gentler and more frequent adjustments, and, as a conceptual replacement, the risk-based guardrails of Kitces, Tharp and Morningstar ([[morningstar-guardrails]]).
- Defensible parameters today: an initial rate of 4 to 4.5%, a corridor of plus or minus 20%, 10% cuts, a floor lined up with your real floor, the indexation freeze kept, an annual review on a fixed date.
- To simulate it honestly you need a tool that takes all four parameters, floor included, and that shows the spending delivered as well as the failure probability. Judge on the life lived and on the trade-off between income and risk, never on the success rate alone.

---

## Going further

- Guyton & Klinger, "Decision Rules and Maximum Initial Withdrawal Rates", *Journal of Financial Planning*, 2006: the original article (readable, and instructive to reread once you know what came after).
- Early Retirement Now, parts 9 and 10: the demonstration of the pathology, with the simulations to back it ([[the-ern-series]]).
- Kitces & Tharp, "The Extraordinary Upside Potential Of Sequence Of Return Risk In Retirement" and the guardrails posts at [kitces.com](https://www.kitces.com): the risk-based descendants.
- In this book: [[morningstar-guardrails]] (the state of the art of the family), [[floor-and-ceiling]] (the other way to run a corridor), [[flexibility-in-practice]] (what flexibility can really buy).
