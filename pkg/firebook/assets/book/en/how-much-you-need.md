# How much you need
<!-- source: combien-il-vous-faut @ 3c0207c058d3 -->

"You need 25 times your annual spending." The formula fits on one line. Almost everyone still gets it wrong. The mistake is never in the multiplication: it is in the two numbers being multiplied. Which spending, exactly? And which multiple, for your horizon, your portfolio, your taxes and your other income?

This page walks the whole calculation, from bank statement to capital target, and adds the corrections the barstool version leaves out. By the end you will be able to produce **your** number, with an honest range around it instead of false precision.

::: cle The real formula
Target capital = (the gross annual spending of the life you are aiming at, tax on the withdrawals included, minus the discounted value of your non-portfolio income) × a multiple between 25 and 33, chosen to fit your horizon, your portfolio and your flexibility. Every term in that sentence gets a section below.
:::

## Step 1: what you actually spend

The multiple amplifies everything. Miss EUR 200 a month and your target comes out EUR 60,000 to 80,000 short, which is one to two years of work. Your spending estimate therefore deserves more care than your choice of withdrawal rate.

**Start from the statements, not from memory.** Export 12 months, ideally 24, from every account and every card, and sort the lines. The gap between what people think they spend and what they spend is almost always 15 to 25%, and always in the same direction.

**Put the lumpy items back in.** A car every 8 years, a furnace every 15, a roof every 30, gifts, the one big trip, the vet: annualize all of it. A working rule for a homeowner is that long-run upkeep costs 1 to 2% of the property's value every year, even in a year that "cost nothing".

**Budget the life you are aiming at, not the one you have.** FIRE changes the shape of your spending. Less commuting, fewer meals bought out for lack of time, fewer of the small consolations a hard job pays for. More hobbies, more travel, more free hours to fill. And health coverage entirely on your own tab, which for an early retiree is usually the largest new line in the budget and the one most often underestimated ([[us-healthcare-and-social-security]]).

**Think in age steps.** Real spending in retirement traces a smile: high and travel-heavy at the start, calmer in the middle, then rising again late in life with health care and long-term care ([[spending-in-retirement]]). For the initial sizing, plan on the high early plateau and let the natural downslope work in your favor.

::: astuce Three budgets, not one
Write down three numbers: the **floor** budget (what cannot be compressed: housing, food, health, insurance, taxes), the **comfort** budget (the life you are aiming at) and the **dream** budget (with the extras). The gap between floor and comfort is the flexibility you can call on in a bad decade ([[flexibility-in-practice]]). Modern withdrawal strategies turn it into an explicit parameter ([[floor-and-ceiling]]). A plan whose floor sits at 90% of comfort is fragile; at 70%, it is robust.
:::

## Step 2: the tax correction, the costliest thing people forget

The founding studies leave taxes out, and so does the 25x multiple. Your withdrawals will be taxed in part. You are aiming at a net standard of living, but the withdrawal that funds it starts out gross.

How much is taken depends on which account the money sits in, on how much of each withdrawal is gain rather than your own capital, and on the rules where you live. That plumbing changes from country to country, and the US version of it is in [[us-taxes-in-the-withdrawal-phase]]. The shape of the correction is the same everywhere. Only the gain inside a withdrawal is taxed. That embedded gain fraction climbs as the years pass. And a sheltered account defers or shrinks the bill without always erasing it. A retiree drawing a modest income can end up paying very little. One drawing a large income from an account holding decades of unrealized gains pays real money.

**Turn it into a single number.** Call it the friction: the share of a gross withdrawal that never reaches your bank account. Estimate it, write it down as an assumption, and gross your target spending up by it. A few percent for a modest, well-sheltered plan; more, sometimes a good deal more, for a large one drawn from a taxable account. Then replace the assumption with your own figure once you have run the numbers on your own accounts, and revisit it at the annual review, because it drifts upward as gains pile up.

::: exemple From net to gross
Target spending: EUR 36,000 a year, net. Assumed tax friction on the withdrawals: 12% (an assumption, to be replaced by your own). Gross spending to fund: 36,000 / 0.88, about EUR 40,900 a year. At 3.5%, the target moves from EUR 1,029,000 (the naive calculation on the net figure) to EUR 1,169,000. Forgetting the tax would have undersized the plan by EUR 140,000, roughly two years of saving for a well-paid couple.
:::

## Step 3: subtract your non-portfolio income

The portfolio does not have to fund what other income will fund. Three categories, each handled differently.

**Your state pension.** Even someone who stops at 42 has usually banked years of contributions, and will draw a pension in their sixties: smaller than a full career would have earned, but real. Ignoring it is the most common cautious mistake, and it is paid for in years of work nobody needed. Do not subtract it from the budget, though, because it does not arrive for another twenty or twenty-five years. Model it as deferred income instead: a dated flow, in constant euros, that lightens the withdrawals from the year it starts. Any serious simulator accepts such a flow. The effect on failure probability is often dramatic, because a pension relieves exactly the long scenarios, the ones where the portfolio runs dry ([[us-healthcare-and-social-security]]).

**Near-certain income**: rent from a property you own ([[real-estate-in-retirement]]), a pension already claimed, an annuity. Subtract these from the spending, after a prudence haircut for what can go wrong: vacancy, repairs, the tax they carry of their own.

**Hoped-for but unguaranteed income**: part-time work, a hobby that starts paying ([[going-back-to-work]]). Do not subtract these when you size the plan. Treat them as margin. The plan has to hold without them. All they do is soften the crossing, the years spent inside a bear market.

## Step 4: choose the multiple

The multiple is the inverse of the withdrawal rate ([[the-4-percent-rule]]). Choosing one is choosing the other, so what you are really trading is extra years of work against the risk of having to adjust later. Here are the markers research offers ([[the-ern-series]], [[morningstar-guardrails]], [[anarkulova-cederburg]]):

| Multiple | Rate | Who it fits |
|---|---|---|
| 25 (4%) | 4.0% | A horizon of 30 years or less, or a long horizon with big margins: real flexibility, a state pension coming later, likely part-time income |
| 28-29 (3.5%) | ~3.5% | The working compromise for an early exit with a globally diversified portfolio and ordinary flexibility |
| 33 (3%) | 3.0% | A very long horizon with no margin at all, a budget already at the floor, strong risk aversion, or faith in the pessimistic global sample |
| > 33 | < 3% | Rarely rational: past this point the dominant risk is no longer ruin, it is dying rich after years of work nobody needed ([[one-more-year]]) |

::: figure cible-convexite
The three rows of the table, placed on the curve that links them. Because the multiple is the exact inverse of the rate, every extra half point of caution costs more than the one before, and the "> 33" row is a property of that curve rather than an opinion.
:::

Two forces pull the choice in opposite directions, and naming them is what makes the trade honest. Pulling the rate down, the multiple up: a long horizon, expensive starting valuations ([[valuations-and-cape]]), a world wider than the American sample, fees and taxes. Pulling the rate up, the multiple down: genuine spending flexibility, future income (that pension again), the ability to go back to work, an adaptive withdrawal rule ([[choosing-your-strategy]]). One more fact belongs on that list and usually gets left off. The simulator's binary "success" is too blunt a measure. Most of the historical "failures" at 3.5% are paths where the trouble was visible ten years ahead ([[when-to-worry]]).

::: science The multiple is not a precision call, it is a posture call
The gap between 25x and 33x is typically 3 to 6 more years of work for someone saving 40 to 50% of their income. Research cannot settle it for you. It bounds the reasonable range, 25 to 33, and measures what each protection pays back in multiple. The final call is a life trade between two lopsided risks: running short of money at 80 (serious, but visible long in advance and cushioned by a pension), or trading away years of active life that never come back. That call is **yours**. Be wary of anyone who makes it for you with confidence, in either direction.
:::

## Step 5: assemble it, then stress it

The full calculation fits in five lines. Here it is on a realistic case.

::: exemple The calculation end to end
Nadia and Marc, 41 and 43, are aiming to stop at 48.

1. **24 months of statements** give EUR 3,400 a month of real spending, lumpy items annualized.
2. **The life they are aiming at** adds EUR 350 a month of travel and leisure and EUR 220 a month of health coverage: **EUR 3,970 a month, or EUR 47,600 a year net**.
3. **Assumed tax friction on the withdrawals**, 12%: **EUR 54,100 a year gross**.
4. **Non-portfolio income**: none before 65. Pensions estimated at EUR 2,100 a month for the couple between 65 and 67, read off their official pension statements, given to the simulator rather than subtracted.
5. **Multiple**: a 45-year horizon, a global 70/30, a floor at 75% of comfort, hence **3.5%, or 28.6x**.

**Target: 54,100 × 28.6, about EUR 1,547,000.**

Check it in a simulation, at EUR 1,550,000 and EUR 54,100 a year with pensions from 66: central-case failure probability around 5%, below the 10 to 20% working zone ([[failure-probability]]) by construction, since the pension is counted and the multiple is already cautious. The same plan **without** the pensions shows about 12%, and would have needed roughly EUR 200,000 more. The state pension is "worth" four years of work here. That is why it never gets left out.
:::

::: figure cible-cascade
The same calculation, step by step, in euros of capital. The two corrections the barstool version leaves out, tax friction and the state pension, each outweigh the travel budget, and they pull in opposite directions.
:::

The number you land on is not a sacred finish line. It is the center of a range. Stress it: plus or minus 10% on the spending, plus or minus half a point on the rate, the pension pushed back two years. If the conclusion (in practice, your departure date) survives those shakes, the plan is solid. If it flips, you know which input to work on. Replaying the plan under a handful of variants is the most profitable thing you can do with a simulator, and far more useful than staring at the first verdict it printed.

## The shortcuts that fool people

**"I'll just use what I spend now."** Almost always wrong, and wrong in both directions at once: too high (a mortgage that ends, children who leave, work expenses that disappear) and too low (health coverage, medical bills, the hobbies that fill freed-up hours, the home repairs you kept putting off). Do the real work of step 1.

**"The simulator says 97%, so I'm fine."** A simulator is worth its inputs and its model, and nothing more ([[simulator-traps]]). 97% on US historical windows, with spending underestimated by 15%, is false comfort. Better practice: several models read side by side ([[historical-vs-parametric]]), audited inputs, and margin.

**"I'll count my home in my capital."** The home you live in cuts your spending, since you pay no rent, but it produces no withdrawals. Counting it in the target capital counts that effect twice. It stays a reserve of last resort (selling, downsizing, a reverse mortgage, [[real-estate-in-retirement]]): a margin, not an asset of the plan.

**"I hit the number, then I stop thinking about it."** The target moves with your life. Recompute it once a year, in ten minutes, at the annual review ([[the-annual-review]]).

## The essentials

- The target = gross spending (tax on the withdrawals included) × 25 to 33.
- The shakiest term is the spending: 24 months of statements, lumpy items annualized, health coverage added, three budgets (floor, comfort, dream).
- A state pension goes into the simulator as deferred income. Leaving it out costs years of work nobody needed.
- The multiple is a posture call between 25 and 33. 28 to 29 (3.5%) is the modern working compromise for a diversified early exit.
- The result is a range to stress, not a number to carve. Next comes [[using-the-fire-simulator]] to put it to the test, then [[withdrawal-strategies-overview]] to choose how to spend it.

---

## Going further

- Early Retirement Now, SWR Series parts 2 (preserving capital against consuming it) and 28 (the toolbox): [earlyretirementnow.com](https://earlyretirementnow.com) ([[the-ern-series]]).
- Your own pension statement, from whoever owes you the pension: the official estimate is the input for the deferred-income line of your plan. In the United States that is the Social Security Statement in a *my Social Security* account ([ssa.gov/myaccount](https://www.ssa.gov/myaccount), [[us-healthcare-and-social-security]]).
- Morningstar, *The State of Retirement Income*: the recommended rates, recomputed every year ([[morningstar-guardrails]]).
- The FIRE simulator ([[using-the-fire-simulator]]), and exactly what it computes ([[under-the-hood]]).
