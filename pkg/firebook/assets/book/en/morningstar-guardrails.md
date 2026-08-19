# Modern guardrails (Morningstar): the state of the art
<!-- source: guardrails-morningstar @ a3cad1683161 -->

The guardrails family did not stop at Guyton-Klinger ([[guyton-klinger]]). Two bodies of work rebuilt it after 2020. The first is Morningstar's annual *The State of Retirement Income*, now the institutional reference of the field. It is the report the press and advisers quote every winter when they announce that the safe rate has moved again. The second is the risk-based guardrails of Michael Kitces and Derek Tharp. They throw out the 2006 indicator, the current withdrawal rate, and put the right one in its place: the recomputed success probability of the plan.

Together they are the state of the art of flexible withdrawal. Morningstar ranks the strategy first year after year on total spending and stability together, and serious practitioners build on the same architecture. This article takes both apart. Morningstar's method and its numbers come first, including the history of its recommended rate (3.3%, then 4.0%, then 3.7%, then 3.9%), a lesson in itself. Next, the exact machinery of the risk-based guardrails, and why their indicator is the better one. Then what they really cost, model dependence above all. A seven-year walk-through, with a table and a chart, shows what the rule makes you live: the pace of the decisions and the size of the swings. The last section is the executable version, with the simulator as the instrument, and the parameters to write down before you run the rule.

::: cle What changed since 2006
Two shifts. The **judging criterion** first: Morningstar grades the rules on the total spending actually lived and on its variability, bequest included, not on the success rate alone. That is the criterion that convicted Guyton-Klinger ([[withdrawal-strategies-overview]]). Then the **indicator**: Kitces and Tharp fire their adjustments off the recomputed success probability of the whole plan (remaining horizon, pensions still to come, allocation), not off the raw withdrawal-to-portfolio ratio, which is blind to all of it. The architecture is the one from 2006, a corridor and adjustments of plus or minus 10%, but the judge is honest now and the thermometer reads true.
:::

::: admin How to run it
- **You no longer watch a ratio, you watch the whole plan.** Two formulations exist, and it helps to know they are equivalent. The first: rerun the simulator on today's numbers and read the success probability. The second: divide this year's spending by your **total wealth** (portfolio plus the present value of the pensions still to come) and compare it, not with your starting rate, but with the rate still **safe for the horizon you have left**, tabulated once and for all under your assumptions. Both say the same thing. The second costs a thousand times less to compute, which is how tools actually implement it.
- **Indexation.** Same as the fixed rule: the spending level is re-indexed to inflation each year, and the adjustments of plus or minus 10% apply on top of it, to the current level.
- **Frequency.** Annual, but with **hysteresis**: a breach triggers nothing until the next review confirms it. That is what replaces the question of timing here, and it works better, since a confirmation filters the noise that a choice of date never will ([[the-annual-review]]).
- **Thresholds.** Plus or minus 20% around the safe rate, or, read in success probability, a corridor of roughly 85 to 99% around a 90% target, with adjustments of plus or minus 10%. None of these numbers is derived, and the practitioners say so themselves: they get calibrated on your plan. The only check worth doing is to verify in simulation that the cut you wrote really brings the plan back inside the corridor, and that the raise you wrote does not push it out.
- **Floor, and ceiling.** The floor is the same as in 2006. The **ceiling** on raises is not optional here, and that is the difference with the ancestor: the center of the band takes off as the horizon shortens, since a five-year horizon is "safe" at 20% a year. A version with no ceiling lifts the standard of living year after year, up to a level the plan no longer funds.
- **In your head.** Impossible, and the method owns that limit. It needs a simulator, or at the very least a precomputed table of the safe rate. That is exactly what the better indicator costs.
:::

## Morningstar: the annual reference, and what it settles

Since 2021, Morningstar's retirement team (Christine Benz, Jeffrey Ptak, John Rekenthaler, later Amy Arnott and Jason Kephart) has published a report every December that redoes the sustainable withdrawal rate from scratch. Three choices of method set it apart from the whole historical literature. It runs on forward-looking returns, the thirty-year capital market assumptions of Morningstar Investment Management ([[expected-returns]]), not on past averages. It judges on 90% success over 30 years, the practitioner standard, more forgiving than ERN's worst case ([[failure-probability]]). And, above all, it compares the flexible rules on one four-dimensional grid: the initial rate they allow, total spending over the life of the plan, income volatility, and the final bequest.

**The series of headline rates** is an epistemology lesson on its own: 3.3% (2021, extreme valuations, zero rates), 3.8% (2022, stocks deflated), 4.0% (2023, bond yields restored), 3.7% (2024, stocks expensive again), 3.9% (2025, this time from a review of the method rather than from the markets, [[expected-returns]]). The "safe" rate is not a constant. It breathes with the conditions you start in ([[valuations-and-cape]]), and a serious institution accepts republishing it. Which sets the use to make of it. Morningstar's number for the year is the best free second opinion for calibrating a 30-year plan. For a FIRE horizon it gets marked down, as usual ([[horizon-and-life-expectancy]]).

**The ranking of the flexible rules** barely moves from one edition to the next. The fixed inflation-adjusted rule sacrifices the most spending. A simple indexation freeze after a red year ([[fixed-inflation-adjusted-withdrawal]]) offers the best benefit for the simplicity. The RMD schedule, the minimum withdrawal the IRS forces on retirees and the regulatory cousin of VPW ([[vpw]]), maximizes spending at the price of a bumpy income. And the **guardrails come first on the combined ranking**: the highest initial rate allowed (often 5% and up in their editions), the highest total spending at tolerable variability, at the price of the smallest bequest and of the complexity. Morningstar's guardrails are a Guyton-Klinger that has calmed down. A corridor of plus or minus 20% on the current rate, adjustments of 10%, but cuts capped in frequency and raises slowed. The lessons of the 2006 pathology are baked in ([[guyton-klinger]]).

Honesty demands repeating the warning that comes with the ranking: it describes a 65-year-old retiring for 30 years, with Social Security in the background. Retiring at 45 borrows the **architecture**, never the rates. A guardrail starting at 5% for 50 years in an expensive market is still reckless, however good the guardrail ([[flexibility-in-practice]]).

## Kitces and Tharp: changing the indicator

The second body of work goes after what was still wrong with the whole family: the current withdrawal rate is a **bad** thermometer. It ignores the horizon you have left (a current rate of 6% at 85 is healthy, at 52 it is serious). It ignores future flows (the same 6% two years before the pensions start is harmless, [[pensions-and-other-income]]). It ignores the allocation and the valuations. So the 2006 guardrail cuts retirees who needed no cut, and reassures plans that are already doomed.

Kitces and Tharp's proposal, developed at [kitces.com](https://www.kitces.com) and industrialized in tools like Income Lab, fits in one sentence: **watch the success probability of the whole plan** directly, recomputed with a simulator, and put the guardrails on that. It is a Bayesian update of the plan, really. Every year of markets and spending actually lived revises the belief that the plan works, and the rule acts on the revised belief instead of on a number frozen in year 1. The standard architecture:

- **The target**: keep the plan near 80 to 90% success (10 to 20% simulated ruin, the practitioner's working zone, [[failure-probability]]).
- **The lower guardrail**: if the success probability falls below roughly 70 to 75% (the portfolio has taken damage, or the assumptions have degraded), cut spending by about 10%, which lifts it back above the target.
- **The upper guardrail**: if it goes above about 99% (the plan is on its way to dying rich), raise spending by about 10%.
- **The review**: annual on a fixed date, or triggered by a breach ([[the-annual-review]], [[when-to-worry]]).

Two households show why the indicator is better. Take the 62-year-old whose pension starts at 64. A crash sends his current rate jumping, so the 2006 guardrail cuts. His success probability barely moves: two bridge years to fund, and the rest is matched. The risk-based guardrail does not cut, and it is right. Now take the 48-year-old who retired early into a very expensive market. His current rate stays moderate, so the 2006 guardrail sleeps. But his simulator, anchored to valuations ([[cape-based-rules]]), watches the success probability slide: the early warning lands years before the raw ratio says anything. The risk-based indicator carries every piece of information this book has assembled (sequence, valuations, horizon, flows), because that indicator is the simulator itself.

::: figure deux-thermometres
The two households, each run through both instruments. The verdicts cross, and that is the whole argument: one corridor gives opposite decisions depending on what you feed it. The raw ratio cuts the first household, which needs no cut, and sleeps through the second, which is sliding. Look hardest at the case on the right: nothing moved on the 2006 thermometer while fourteen points of success probability disappeared. An instrument that does not move is not the same thing as a plan that is fine.
:::

::: attention The price of the better indicator: model dependence
The risk-based thermometer inherits everything fragile in the simulator behind it ([[simulator-traps]]). A success probability computed on a cheerful Gaussian model will say "all fine" right into the wall. One computed on the broad sample may cut too early. False signals are real too. The success probability moves with the markets and with every update to the assumptions, so a pair of guardrails set too tight (75/95) will yo-yo. The defenses are known. Run the indicator on several models, firing on the central case and checking against the harsh ones: the weight of the evidence, applied here. Use wide bands. Use hysteresis, which reacts only to a breach confirmed at two reviews in a row. And keep the adjustments gentle. The risk-based guardrail is the state of the art for anyone with an honest simulator and the discipline to use it. Without both, the old raw ratio with a floor under it is still defensible ([[guyton-klinger]]).
:::

## Seven years with a risk-based guardrail

All of this stays abstract until you watch the rule live. Nothing beats a walk-through, as with the Vanguard corridor ([[floor-and-ceiling]]). Take the plan used as the example below: $1.5M, comfort withdrawal $54,000 (3.6%), floor $44,000, pension in year 18. The corridor is written this way: cut 10% if simulated success falls below 85%, raise 10% if it goes above 99%, and no breach counts until two consecutive reviews confirm it (the hysteresis). The sequence is an early bear market followed by a recovery, and the probabilities read off are illustrative.

| Review | Portfolio (real) | Success read | Decision (corridor 85 to 99, hysteresis) | Withdrawal |
|---|---|---|---|---|
| 1 | 1,500,000 | 93% | inside the corridor: nothing | 54,000 |
| 2 | 1,150,000 (−20%) | 82% | first low alert: wait | 54,000 |
| 3 | 990,000 (−10%) | 76% | alert confirmed: cut 10% | **48,600** |
| 4 | 1,060,000 (+12%) | 88% | back inside the corridor: nothing | 48,600 |
| 5 | 1,120,000 (+10%) | 91% | corridor: nothing | 48,600 |
| 6 | 1,340,000 (+24%) | 99.2% | first high alert: wait | 48,600 |
| 7 | 1,430,000 (+11%) | 99.4% | confirmed: raise 10% | **53,460** |

::: figure guardrails-indicateur
The seven reviews of the table, in two aligned panels. On top, the indicator: the recomputed success probability crosses the 85 to 99% corridor, and only a breach confirmed at two reviews in a row triggers an adjustment (the hollow circles are first alerts, put on hold). Below, the income: two steps of plus or minus 10% in seven years, nowhere near the floor. The normal case is stillness.
:::

**Read the pace first.** Seven reviews, two decisions. The normal case, five years out of seven, is to do nothing, and that is by design: a well-sized guardrail stays quiet most of the time ([[the-annual-review]]). The hysteresis cost a year of delay on the cut, since year 2 was already flashing. That is its job. Had year 3 bounced, the alert would have died with no decision taken, and the income would never have moved. You trade a little responsiveness for the absence of yo-yo.

**Then read the sizes.** This answers the real question, what the rule feels like to live on. Through a severe crossing, the years spent inside a bear market (−28% real at the trough), the income took one step down of 10%, held it four years, and then climbed back. It never came close to the floor ($48,600 against $44,000). The raw percentage rule would have delivered −28% in two years ([[fixed-percentage]]); the fixed inflation-adjusted rule would have changed nothing at all, letting the silent ruin build ([[fixed-inflation-adjusted-withdrawal]]). Steps of plus or minus 10%, rare and dated, are the family's signature: decisions that come seldom but are real, where the Vanguard corridor prefers a continuous glide with no decision at all ([[floor-and-ceiling]]).

**Last, read what the indicator sees.** Between reviews 3 and 6, simulated success climbs from 76 to 99%. The markets explain only part of that. The cut itself adds to it, less spending meaning more success, and so does a factor the raw ratio will never see: every year that passes shortens the remaining horizon and brings the year-18 pension closer. Time works for the indicator. That is precisely the information the 2006 thermometer threw away, and the reason the rule was rebuilt.

## The executable version, with the simulator as the instrument

Here is the whole thing assembled for a reader of this book: risk-based guardrails you can actually run, with the simulator as your instrument.

**At design time.** Set the three thresholds. The target failure probability on the central model, 5% say, which is what the Acceptable ruin setting holds. The lower guardrail, a central-case failure above 12 to 15%, placed so that the 10% cut is enough to get back under the target. Two simulations check that calibration, the damaged plan with the cut and without it. The upper guardrail, a failure below 1% together with a very low current rate, which is the condition for a ratchet. Then write down the three numbers, the size of the adjustments (plus or minus 10% of comfort spending, never below the floor, [[how-much-you-need]]) and the hysteresis (two consecutive reviews).

**In operation.** Once a year, on a fixed date, three moves: update the capital, the real spending and the pensions, rerun the simulation, read the failure probability of the central model and of the harsh ones. Three cases follow. Inside the corridor, do nothing: the normal case is a virtue, not a disappointment. Two years running under the upper threshold, take the raise, or let the rule produce it on its own if your tool simulates it continuously. Two years running above the lower threshold, apply the written cut. Between two reviews, nothing, unless a breach is massive ([[when-to-worry]]).

**The continuously simulated version**, for anyone who would rather have the rule inside the model than run the annual procedure by hand. The principle is still Kitces and Tharp's. The band no longer follows the starting rate, but the rate still safe for the horizon you have left, tabulated once and for all under your central assumptions. And the current rate is read on total wealth, the pensions still to come discounted and included. The adjustments stay the family's own, plus or minus 10% at every breach. As with the ancestor, the floor is not a rule stacked on top of the guardrails but a parameter of the rule itself, alongside the ceiling that slows the raises ([[guyton-klinger]]). In the simulator, that is exactly how Risk-based guardrails (Kitces / Morningstar) works, with its own guardrails floor and its own raise ceiling.
Two implementation details are worth knowing. Expressing the indicator as a withdrawal rate rather than as a success probability recomputed every year gives the same signal for a fraction of the compute, one being a reading of the other. And the table of safe rates, once solved under a given model, is then kept as it is. A table written on a flattering history will therefore allow raises that a harsh model punishes later, and that gap is precisely the model risk of the paragraph above.
What is left is to look at what your version would make you live, on the distribution of the spending delivered as much as on the frontier between risk and income. Set the rule to your written parameters, then look at the worst quartile before you sign.

::: exemple Risk-based guardrails, sized in one sitting
The plan: $1.5M, comfort $54,000, floor $44,000, age 45, a pension of $16,000 in year 18. The design session starts at the comfort level: central-case failure 5% (target met), broad sample 11% (accepted, since the failures are late and the pension covers the floor). Then the lower guardrail gets tested. Simulate the plan after an early crash, capital down to $1.1M in year 3: central-case failure 14%. That is the threshold. The written cut (10%, down to $48,600) brings it back to 7%, so the adjustment is big enough and the guardrail sits in the right place. Test the upper one at $2.1M (a rich year 8): failure 0.6%. The 10% raise takes it to 1.1%, granted. The whole plan fits on half a page: three thresholds, two adjustments, one review date. And every number was **checked** in both directions instead of copied out of somebody else's article. That is the whole difference between applying a rule and owning yours.
:::

## What it changes, and what it does not

Modern guardrails are where the family of fixed rules that bend ends up. That makes them the best default for a household that wants a **stable** income most of the time, with its safety actively steered. They do not repeal the laws of physics. Flexibility still buys about 0.3 to 0.5 point of rate over the equivalent fixed rule, not a point and a half ([[flexibility-in-practice]]). The floor is still the real admission requirement. And in the final comparison the actuarial family ([[amortization-based-withdrawal]]) keeps advantages of its own, no cliff anywhere and higher spending, against a variability you accept up front. Choosing between the two winning families is the subject of [[choosing-your-strategy]].

## The essentials

- Two rebuilds: Morningstar (the honest judge, spending actually lived plus variability plus bequest, forward-looking returns, an annual report whose series of rates, 3.3, 4.0, 3.7 then 3.9%, teaches that the rate breathes) and Kitces-Tharp (the right thermometer, the success probability of the whole plan, not the raw ratio).
- In Morningstar's ranking, the calmed-down guardrails dominate on spending and stability together; the plain indexation freeze remains the best benefit for the simplicity; everything transposes as architecture, never as rates, for a FIRE horizon.
- The risk-based indicator carries horizon, pensions and valuations. It does not cut the retiree whose pension is about to start, and it warns the early retiree in an expensive market years before the ratio does. But it inherits the simulator's fragilities: several models, wide bands, hysteresis and gentle adjustments are all mandatory.
- The pace is slow: years of nothing, then one step of plus or minus 10% confirmed at two reviews. In the walk-through, a crossing at −28% is lived as a single 10% cut held four years, nowhere near the floor, with the indicator recovering afterwards through three channels (the markets, the cut, and the horizon getting shorter).
- The executable version: three failure thresholds written down (target about 5%, cut above 12 to 15%, raise below 1%), plus or minus 10% and never below the floor, an annual review, confirmation at two reviews. And every threshold **checked** by simulation in both directions.
- The family is the state of the art of steered stable income; its final rival is amortization ([[amortization-based-withdrawal]]): see [[choosing-your-strategy]].

---

## Going further

- Morningstar, *The State of Retirement Income* (annual, free at [morningstar.com](https://www.morningstar.com)): the report, its method and the comparison of the rules.
- Kitces & Tharp, "Probability-Of-Success-Driven Guardrails", and the guardrails series at [kitces.com](https://www.kitces.com); Income Lab for the industrialized version.
- Guyton & Klinger (2006) for the ancestor, and ERN parts 9 to 11 for the critique that made these rebuilds necessary ([[the-ern-series]], [[guyton-klinger]]).
- In this book: [[guyton-klinger]] (the ancestor and its four parameters), [[floor-and-ceiling]] (the continuous corridor, the other school of steered income), [[choosing-your-strategy]] (the final trade-off). To run both indicators step by step, see [[using-the-fire-simulator]].
