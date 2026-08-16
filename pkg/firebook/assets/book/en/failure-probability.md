# Failure probability: reading it, choosing it, and not letting it run you
<!-- source: ruine-et-probabilites @ a00366b22d13 -->

Every retirement simulator boils its verdict down to a single number: the failure probability, or its mirror image, the success rate. It is the most watched number in the field, and the most misread. People treat it like a weather forecast ("5% risk, that's fine"), compare it across tools that are not measuring the same thing, ask it for a precision it does not have, and forget that it describes a world in which nobody ever reacts.

This page teaches you to read that number the way a professional does: what it measures exactly, how to pick a threshold you can live with, why 2% and 8% are often indistinguishable, and how simulated ruin differs from the real thing.

::: cle The definition, no shortcuts
The failure probability of a plan is the fraction of simulated futures (or of replayed historical windows) in which the portfolio hits zero **before** the end of the horizon, under a withdrawal rule applied mechanically, with no human reaction at all. It is a property of the **pair** plan plus model, never of the plan alone: change the model and the number changes, sometimes by a factor of three, without your plan moving an inch ([[historical-vs-parametric]]).
:::

## What the number measures, and what it does not

Take the definition apart, phrase by phrase. Each phrase hides a trap.

**"The fraction of simulated futures."** The number is a frequency inside a population of scenarios that a model made up. The model changes everything. Draw the years independently and you understate the clusters of bad ones ([[sequence-of-returns]]). Replay American history and you inherit its optimism ([[the-trinity-study]], [[simulator-traps]]). Replay the world sample and you take in countries and eras that may be harsher than any future you plausibly face ([[anarkulova-cederburg]]). None of them is "the true one". That is why serious simulators show the same plan under several models at once, and why the right reading is the **range** they draw, never one model standing alone.

**"The portfolio hits zero."** Simulated ruin is binary and terminal. It does not tell failure at 71 apart from failure at 94, and it scores a plan that touches zero in its final month as a failure while a plan that ends with EUR 15,000 in hand is a success. Two plans at 5% can hide very different lives: one fails early and hard, the other runs out of breath at the very end with Social Security underneath it. So look at two more things. **When** the failures happen, and how much wealth is left at the median at the end of the horizon ([[reading-a-fan-chart]]).

**"With no human reaction at all."** The least realistic assumption, and the most useful. Unrealistic: nobody keeps drawing EUR 40,000 indexed while the portfolio slides from EUR 1M to EUR 150k; they would have cut their spending years earlier ([[when-to-worry]]). Useful: precisely because the simulated rule is blind, the number measures the plan's **intrinsic** robustness, with no credit taken for a flexibility you have yet to show. A flexible plan gets simulated with its own flexible rule ([[floor-and-ceiling]], [[morningstar-guardrails]]). Failure then drops, and what you have to look at is the standard of living delivered. Flexibility does not remove the pain, it moves it into years of smaller spending.

**And what it blithely ignores**: your mortality (ruin at 97 concerns very few people, which is why it pays to read the number alongside mortality tables, [[horizon-and-life-expectancy]]), the real lumpiness of spending, the fine print of taxes, and every safety net that sits outside the model (family, a house, a return to work).

## Choosing your threshold: there is no universal right answer

How much failure should you accept: 1%, 5%, 10%? The question looks technical. It is personal, and it turns on three things.

**How good your safety nets are.** Simulated ruin assumes you have nowhere to turn. Someone in their forties who is employable, owns a home, has a pension coming and a family that would help can rationally accept 10 to 15% **simulated** failure, because **real** ruin (ending your life destitute with nothing to fall back on) is far rarer than the number suggests. Someone at 60 with no meaningful pension, no property and no way back to work should read the number almost literally, and aim low.

**What a year of margin costs.** Going from 5% failure to 2% typically takes 10 to 20% more capital, which is two to four more years of work. Going from 5% to 10% hands them back. The threshold is the price of a trade between two risks ([[one-more-year]]): put a number on what each point of failure costs you, or gives you back, in working years, and the discussion stops being abstract. The general frame for trades like this one (utility, certainty equivalent, regret) is laid out in [[deciding-under-uncertainty]].

**How your plan fails.** Look at when the failing scenarios fail. An 8% built out of late failures (after 85, with a pension covering the floor) is more comfortable than a 4% built out of collapses at 70.

::: astuce What practitioners actually do
Serious financial planners, Michael Kitces first among them, have converged on a working range of 10 to 20% simulated failure for plans that come with safety nets and an adjustment rule, and they keep repeating that a 100% success rate is not a healthy goal. It almost always means you worked years too long and will die at your wealthiest. Morningstar calibrates its recommendations to 90% success, meaning 10% failure, over 30 years ([[morningstar-guardrails]]). A solver that sizes a plan (target capital, exit date, sustainable withdrawal) needs an acceptable failure rate as an input. Set it yourself. 10% is a reasonable starting point: lower it when your safety nets are thin, raise it when they are solid and your withdrawal rule knows how to adjust.
:::

## False precision: 2% and 8% are often the same number

The simulator prints "4.7%" and the mind files away a jeweler's precision. There is none, for three reasons that stack.

**Sampling noise** is the smallest of them. With the 2,000 paths simulators commonly draw, a true 5% prints anywhere between 4 and 6% depending on the draw. Annoying, but bounded.

**Parameter sensitivity** is far worse. Shave 0.5 point off the expected real return, a refinement nobody can actually estimate ([[expected-returns]]), and failure can double. The thickness of the tails (the degrees of freedom of the Student-t, [[fat-tails]]) moves it again. Your parameters are uncertain, so your failure probability is at least as uncertain.

**The choice of model dominates everything else.** The same plan can print 2% on historical windows, 5% on the central model, 9% under sequence stress and 14% on the broad sample. None of them is wrong; they answer different questions ("what if the future looks like the history of my own holdings, or like a prudently calibrated i.i.d. world, or like that same world with sticky bears, or like a century across 16 developed countries").

One rule follows: **read failure probability as a ranking, not as a measurement**. It compares beautifully (plan A is more robust than plan B; this lever cuts risk more than that one; even the pessimistic model stays acceptable) and measures poorly ("my real risk is 4.7%"). The decimals are noise. The gaps between scenarios, and between models, are signal.

::: exemple A decision framed properly
The plan: EUR 1.2M, EUR 42,000 a year, a 45-year horizon, a pension of EUR 12,000 a year from 66. Read as a range: 1% on historical windows, 4% central, 7% under sequence stress, 11% on the broad sample. The call: the central case and the stress sit under 10%; the broad sample is above 10%, but its failures land after 80, with the pension already in payment, and a spending floor of EUR 34,000 is livable. Verdict: the plan is acceptable, provided the rule is written down. If the current withdrawal rate goes above 5% (portfolio below about EUR 840k), spending drops to the floor until the rate is back under 4.5%. The same analysis with early failures, or with a floor nobody could live on, would have ended somewhere else: one more year, or 10% less spending.
:::

## Real ruin looks nothing like simulated ruin

One last correction, and the one that does the most for your sleep. In the simulator, ruin is a cliff: the balance crosses zero on a Tuesday and everything stops. In life, a retirement plan fails slowly, and **visibly**: the portfolio drifts off its planned path, the current withdrawal rate climbs year after year, the warning lights turn amber a decade before the drop. Studies of the historical paths that failed agree on the timing: between the moment a doomed plan becomes statistically recognizable and the day the money actually runs out, 8 to 15 years typically go by. That is an enormous amount of notice for anyone who has written down thresholds for action ([[when-to-worry]], [[the-annual-review]]).

That is the real reason failure probability, read properly, is an instrument of **design** and not a source of dread. It is there to compare plans and to size margins before you leave. Once you have left, it gives way to piloting: a few simple indicators, written thresholds, prepared answers. A plan at 8% failure with an attentive pilot is safer than a plan at 3% with a sleeping one.

## The essentials

- Failure probability is a property of the pair plan plus model: read the range across several models, never one model on its own.
- The number assumes zero human reaction. It measures the plan's intrinsic robustness, not your fate.
- Pick your threshold from the safety nets you really have and from the price of margin in working years; 10 to 20% simulated is where practitioners work when the nets and an adjustment rule exist, and 100% success is an anti-goal.
- Read it as a ranking: the gaps compare, the decimals lie. 2% and 8% are often indistinguishable once parameter uncertainty is counted.
- Real ruin is slow and visible years ahead. What protects you after you leave is not a lower number, it is a written piloting routine ([[when-to-worry]]).

---

## Going further

- Early Retirement Now, SWR Series part 11 ("Six Criteria to Grade Withdrawal Rules") and part 46 ("The Need for Precision in an Uncertain World"): [earlyretirementnow.com](https://earlyretirementnow.com) ([[the-ern-series]]).
- Michael Kitces, "Flexible Spending Rules To Avoid FIREing At 4%" and "Is A Probability Of Success-Driven Retirement Plan Actually Riskier?" ([kitces.com](https://www.kitces.com)): the practitioner's reading of the success rate.
- Derek Tharp and Michael Kitces on probability-of-success-driven guardrails: piloting rather than a static number.
- In this book: [[historical-vs-parametric]] (why models disagree about the same plan) and [[under-the-hood]] (how the simulator computes this number and sets it against mortality tables).
