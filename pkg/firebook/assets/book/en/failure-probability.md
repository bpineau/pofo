# Failure probability: reading it, choosing it, and not letting it run you
<!-- source: ruine-et-probabilites @ 94ca3b2c5690 -->

Every retirement simulator boils its verdict down to a single number: the failure probability, or its mirror image, the success rate. It is the most watched number in the field, and the most misread. People treat it like a weather forecast ("5% risk, that's fine"), compare it across tools that are not measuring the same thing, ask it for a precision it does not have, and forget that it counts only the responses built into the rule.

This page teaches you to read that number the way a professional does: what it measures exactly, how to pick a threshold you can live with, why the same plan can show 2% or 8% under different assumptions, and how simulated ruin differs from the real thing.

::: cle The definition, no shortcuts
In pofo a path fails in the first year the portfolio cannot deliver the withdrawal the spending rule calls for after household income is taken into account. The failure rate is the share of paths that encounter at least one such year within the chosen horizon. It depends on both the **plan** and the **model** that produces the scenarios ([[historical-vs-parametric]]).
:::

## What the number measures, and what it does not

Take the definition apart, phrase by phrase. Each phrase hides a trap.

**"The fraction of simulated futures."** The number is a frequency inside a population of scenarios that a model made up. The model changes everything. Draw the years independently and you understate the clusters of bad ones ([[sequence-of-returns]]). Replay American history and you inherit its optimism ([[the-trinity-study]], [[simulator-traps]]). Replay the world sample and you take in countries and eras that may be harsher than any future you plausibly face ([[anarkulova-cederburg]]). None of them is "the true one". That is why serious simulators show a plan under several models at once. Their figures do not form a confidence interval: each asks a different question, and some do not even look at your portfolio. Read them one by one, never just one.

**"A withdrawal cannot be funded."** Suppose the final year calls for a $15,000 withdrawal and exactly $15,000 remains. The household pays every bill and ends at zero: the plan succeeds. Even with another year to live and an empty portfolio, a pension could cover the whole budget. If it leaves a $1,000 shortfall, that is the year pofo counts as a failure. Next ask **when** the shortfall occurs and how much spending goes unfunded ([[reading-a-fan-chart]]).

**"Under the stated rule."** With fixed withdrawals, the calculation keeps paying the inflation-adjusted amount after a downturn. If your plan is to cut travel before touching essentials, simulate a rule that does that ([[floor-and-ceiling]], [[morningstar-guardrails]]). It may lower the failure rate at the price of several years with fewer vacations. Look at the amount of spending delivered and the duration of the cuts to see what that flexibility cost.

**And what it blithely ignores**: your mortality if the horizon is fixed (ruin at 97 concerns very few people, which is why it pays to read the number alongside mortality tables, [[horizon-and-life-expectancy]]), the real lumpiness of spending, the fine print of taxes, and every safety net that sits outside the model (family, a house, a return to work).

## Choosing your threshold: there is no universal right answer

How much failure should you accept: 1%, 5%, 10%? The question looks technical. It is personal, and it turns on three things.

**Your fallback resources.** A pension already entered in the model reduces the withdrawals you need. Counting it again as a fallback overstates your margin. For a resource still outside the model, such as returning to work, consider at what age it would still be available and how much income it could provide.

**What a year of margin costs.** Going from 5% failure to 2% typically takes 10 to 20% more capital, which is two to four more years of work. Going from 5% to 10% hands them back. The threshold is the price of a trade between two risks ([[one-more-year]]): put a number on what each point of failure costs you, or gives you back, in working years, and the discussion stops being abstract. The general frame for trades like this one (utility, certainty equivalent, regret) is laid out in [[deciding-under-uncertainty]].

**How your plan fails.** Look at when the failing scenarios fail. An 8% built out of late failures (after 85, with a pension covering the floor) is more comfortable than a 4% built out of collapses at 70.

::: astuce What practitioners actually do
A solver that calculates a savings target or retirement date needs a failure threshold. Morningstar, for example, uses 90% success over thirty years in its baseline scenario ([[morningstar-guardrails]]). To choose your own threshold, open a few paths that fail. What budget could you live on in those paths, and for how many years? Which bills would you still be unable to pay?
:::

## False precision: why the same plan can show 2% or 8%

The simulator prints "4.7%" and the mind files away a jeweler's precision. There is none, for three reasons that stack.

**Sampling noise.** With 2,000 independent paths and a failure probability of 5% in the model, roughly 95 runs out of 100 give a failure rate between 4% and 6%. Seeing zero failures in a finite sample does not establish that the model's risk is zero. This uncertainty calculation does not directly apply to overlapping historical windows.

**Parameter sensitivity** is far worse. Shave 0.5 point off the expected real return, a refinement nobody can actually estimate ([[expected-returns]]), and failure can double. The thickness of the tails (the degrees of freedom of the Student-t, [[fat-tails]]) moves it again. Your parameters are uncertain, so your failure probability is at least as uncertain.

**The choice of model dominates everything else.** The same plan can print 2% on historical windows, 5% on the central model, 9% under sequence stress and 14% on the broad sample. None of them is wrong; they answer different questions ("what if the future looks like the history of my own holdings, or like a prudently calibrated i.i.d. world, or like that same world with sticky bears, or what if I had been a domestic 60/40 saver in one of the 16 developed countries of the past century"). That last one does not look at your portfolio: it is a crash test, not a measure of your risk.

::: figure trois-bruits
The three noises put on one scale, for a single reference plan: the figure on screen, the two points of uncertainty the draw leaves, the three and a half the parameters leave, and the thirty-five the choice of model leaves. The bar you stare at while believing you are reading a probability is the shortest of the three, by a wide margin.
:::


One rule follows: **read failure probability as a ranking, not as a measurement**. It compares beautifully (plan A is more robust than plan B; this lever cuts risk more than that one; even the pessimistic model stays acceptable) and measures poorly ("my real risk is 4.7%"). The decimals are noise. The gaps between scenarios, and between models, are signal.

::: exemple A decision framed properly
The plan: $1.2M, $42,000 a year, a 45-year horizon, a pension of $12,000 a year from 66. Read model by model: 1% on historical windows, 4% central, 7% under sequence stress, 11% on the broad sample. The call: the central case and the stress sit under 10%; the broad sample is above 10%, but its failures land after 80, with the pension already in payment, and a spending floor of $34,000 is livable. The pension covers only $12,000 of it: the other $22,000 still comes from the portfolio, hence the written rule. Verdict: the plan is acceptable, provided the rule is written down. If the current withdrawal rate goes above 5% (portfolio below about $840k before the pension, about $600k after, once the portfolio only has to fund $30,000), spending drops to the floor until the rate is back under 4.5%. The same analysis with early failures, or with a floor nobody could live on, would have ended somewhere else: one more year, or 10% less spending.
:::

## Real ruin looks nothing like simulated ruin

A failure percentage leaves out the timing. The chart below restores it for a household retiring in 1966. As wealth falls, the planned withdrawal takes a larger share of the balance. That ratio gives the household a signal it can monitor at its annual review. Sometimes there is time to act. A sudden market loss or a large unexpected expense can shorten the warning period ([[when-to-worry]], [[the-annual-review]]).

::: figure preavis-1966
Retirement starts in 1966 with pofo's US 60/40 portfolio and a 4% initial withdrawal, then inflation adjustments. Above, remaining wealth. Below, the annual withdrawal as a percentage of the balance. The warning turns red in 1974 with $673,000 of the initial million still invested. It stays red in 1975. The portfolio runs out in 1994. Those dates reflect this particular path and the chosen thresholds ([[when-to-worry]]).
:::

That is the real reason failure probability, read properly, is an instrument of **design** and not a source of dread. It is there to compare plans and to size margins before you leave. Once you have left, it gives way to piloting: a few simple indicators, written thresholds, prepared answers. A plan at 8% failure with an attentive pilot can be safer than a plan at 3% with a sleeping one.

## The essentials

- Failure probability is a property of the pair plan plus model: read several models, each for the question it asks, never one model on its own.
- The calculation follows the decisions specified by the rule, including any spending cuts you built into it.
- Pick your threshold from the safety nets you really have and from the price of margin in working years.
- Read it as a ranking: the gaps compare, the decimals lie. Compare two plans using the same assumptions.
- Deterioration can be visible years ahead. What protects you after you leave is not a lower number, it is a written piloting routine ([[when-to-worry]]).

---

## Going further

- Early Retirement Now, SWR Series part 11 ("Six Criteria to Grade Dynamic Withdrawal Rules") and part 46 ("The Need for Precision in an Uncertain World"): [earlyretirementnow.com](https://earlyretirementnow.com) ([[the-ern-series]]).
- Michael Kitces, "Flexible Spending Rules To Avoid FIREing At 4%" and "Does the Probability of Success Drive Your Retirement Plan?" ([kitces.com](https://www.kitces.com)): the practitioner's reading of the success rate.
- Derek Tharp and Michael Kitces on probability-of-success-driven guardrails: piloting rather than a static number.
- In this book: [[historical-vs-parametric]] (why models disagree about the same plan) and [[under-the-hood]] (how the simulator computes this number and sets it against mortality tables).
