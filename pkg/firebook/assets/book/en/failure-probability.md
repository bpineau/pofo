# Failure probability: reading it, choosing it, and not letting it run you
<!-- source: ruine-et-probabilites @ 31cbc065d010 -->

A simulation reports a 95% success rate. In the other 5%, did the household miss its last withdrawal by a few dollars, or did it face fifteen years of essential bills it could not pay? The percentage alone cannot tell you.

To interpret it, start with two questions. What event does the tool count as a failure? What does your plan do when conditions deteriorate?

::: cle The definition, no shortcuts
In pofo a path fails in the first year the portfolio cannot deliver the withdrawal the spending rule calls for after household income is taken into account. The failure rate is the share of paths that encounter at least one such year within the chosen horizon. It depends on both the **plan** and the **model** that produces the scenarios ([[historical-vs-parametric]]).
:::

## What the number measures, and what it does not

Take the definition apart, phrase by phrase. Each phrase hides a trap.

**"The fraction of simulated futures."** The number is a frequency inside a population of scenarios that a model made up. The model changes everything. Draw the years independently and you understate the clusters of bad ones ([[sequence-of-returns]]). Replay American history and you inherit its optimism ([[the-trinity-study]], [[simulator-traps]]). Replay the world sample and you take in countries and eras that may be harsher than any future you plausibly face ([[anarkulova-cederburg]]). None of them is "the true one". That is why serious simulators show a plan under several models at once. Their figures do not form a confidence interval: each asks a different question, and some do not even look at your portfolio. Read them one by one, never just one.

**“A withdrawal cannot be funded.”** Suppose the final year calls for a $15,000 withdrawal and exactly $15,000 remains. The household pays every bill and ends at zero: the plan succeeds. Even with another year to live and an empty portfolio, a pension could cover the whole budget. If it leaves a $1,000 shortfall, that is the year pofo counts as a failure. Next ask **when** the shortfall occurs and how much spending goes unfunded ([[reading-a-fan-chart]]).

**“Under the stated rule.”** With fixed withdrawals, the calculation keeps paying the inflation-adjusted amount after a downturn. If your plan is to cut travel before touching essentials, simulate a rule that does that ([[floor-and-ceiling]], [[morningstar-guardrails]]). It may lower the failure rate at the price of several years with fewer vacations. Look at the amount of spending delivered and the duration of the cuts to see what that flexibility cost.

**How long the money must last.** With a fixed forty-year horizon, every household must fund forty years. With simulated lifespans, a shortfall only counts while the household is still alive ([[horizon-and-life-expectancy]]). In either case, check which income, extraordinary expenses and fallback resources the calculation includes.

## Choosing your threshold: there is no universal right answer

How much failure should you accept: 1%, 5%, 10%? The question looks technical. It is personal, and it turns on three things.

**Your fallback resources.** A pension already entered in the model reduces the withdrawals you need. Counting it again as a fallback overstates your margin. For a resource still outside the model, such as returning to work, consider at what age it would still be available and how much income it could provide.

**What a year of margin costs.** Going from 5% failure to 2% typically takes 10 to 20% more capital, which is two to four more years of work. Going from 5% to 10% hands them back. The threshold is the price of a trade between two risks ([[one-more-year]]): put a number on what each point of failure costs you, or gives you back, in working years, and the discussion stops being abstract. The general frame for trades like this one (utility, certainty equivalent, regret) is laid out in [[deciding-under-uncertainty]].

**How your plan fails.** Look at when the failing scenarios fail. An 8% built out of late failures (after 85, with a pension covering the floor) is more comfortable than a 4% built out of collapses at 70.

::: astuce What practitioners actually do
A solver that calculates a savings target or retirement date needs a failure threshold. Morningstar, for example, uses 90% success over thirty years in its baseline scenario ([[morningstar-guardrails]]). To choose your own threshold, open a few paths that fail. What budget could you live on in those paths, and for how many years? Which bills would you still be unable to pay?
:::

## Where does the uncertainty in the percentage come from?

The simulator prints "4.7%" and the mind files away a jeweler's precision. There is none, for three reasons that stack.

**Sampling noise.** With 2,000 independent paths and a failure probability of 5% in the model, roughly 95 runs out of 100 give a failure rate between 4% and 6%. Seeing zero failures in a finite sample does not establish that the model's risk is zero. This uncertainty calculation does not directly apply to overlapping historical windows.

**Sensitivity to assumptions.** Keep the simulation method and lower the expected real return by half a percentage point ([[expected-returns]]). The withdrawals now put more pressure on the portfolio. Repeat the exercise with more frequent extreme losses ([[fat-tails]]). These comparisons show which assumptions matter most to your plan.

**Choice of model.** Historical windows replay observed years. A parametric model draws returns under explicit assumptions. A stress test across countries may use a standard domestic portfolio different from yours. A result of 2% in the first test and 14% in the last describes two different experiments. The range from 2% to 14% therefore has no interpretation as a confidence interval.

::: figure trois-bruits
The three noises put on one scale, for a single reference plan: the figure on screen, the two points of uncertainty the draw leaves, the three and a half the parameters leave, and the thirty-five the choice of model leaves. The bar you stare at while believing you are reading a probability is the shortest of the three, by a wide margin.
:::


Compare plans **on the same scenarios** first. If their ranking changes when you change the expected return or the model, your preference depends on that assumption. If the ranking holds, you have a reason to take it seriously. A difference between 2% and 8% can be quite clear within a given model even if real-world risk remains uncertain.

::: exemple A decision framed properly
A household plans to spend $42,000 a year and will receive a $12,000 pension starting at 66. Before the pension, the portfolio must supply $42,000 a year. Afterward it must supply $30,000. If the household can trim its budget to $34,000, it will still need $22,000 from the portfolio after the pension starts. Simulate the rule that triggers the cut and count the years spent at $34,000 in the difficult paths. The household can then decide whether the lower failure rate is worth the reduction in living standards.
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
- To choose a threshold, consider the size and duration of the shortfalls, your available fallback resources and the cost of a larger margin.
- Distinguish variation caused by the draw from variation caused by the assumptions. Running more simulations reduces the former.
- During retirement, written thresholds can help you respond to a persistent rise in the current withdrawal rate ([[when-to-worry]]).

---

## Going further

- Early Retirement Now, SWR Series part 11 ("Six Criteria to Grade Dynamic Withdrawal Rules") and part 46 ("The Need for Precision in an Uncertain World"): [earlyretirementnow.com](https://earlyretirementnow.com) ([[the-ern-series]]).
- Michael Kitces, "Flexible Spending Rules To Avoid FIREing At 4%" and "Does the Probability of Success Drive Your Retirement Plan?" ([kitces.com](https://www.kitces.com)): the practitioner's reading of the success rate.
- Derek Tharp and Michael Kitces on probability-of-success-driven guardrails: piloting rather than a static number.
- In this book: [[historical-vs-parametric]] (why models disagree about the same plan) and [[under-the-hood]] (how the simulator computes this number and sets it against mortality tables).
