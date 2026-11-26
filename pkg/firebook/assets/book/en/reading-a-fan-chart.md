# Reading a fan chart without fooling yourself
<!-- source: lire-un-fan-chart @ 2144bbcd17c0 -->

How much might you have left at age 75? Find 75 on the horizontal axis of the fan chart and read the amounts above it. The dark band contains the middle half of the simulated balances at that age. The line down the middle marks the median.

To see how one household got there, switch your attention to a single thin line. It might sit above the median at 65 and below the dark band at 75.

The first reading compares households at one date. The second follows one retirement through time.

::: cle What a 90% band means
The band has new boundaries at every date. A household can leave it one year and return the next. To find out how many households stay inside throughout retirement, you would need to follow each path from start to finish.
:::

## Anatomy of a fan

::: figure fan-anatomy
A plan with a 45-year horizon. The nested bands are **quantiles by date** (25-75%, 5-95%), not scenarios. The median rises. And the two example paths cross the bands: the one that ends ruined started out near the top, proof that a band is not a path.
:::

Take the fan above, piece by piece.

**The nested bands.** From darkest to lightest, they cover widening percentile intervals around the median: here the middle half of the futures (25-75) and nearly all of them (5-95). At each date, the thousands of simulated wealth levels are sorted and the quantiles plotted. So a band is not a "cautious" scenario or an "optimistic" one: it is a statistic of the whole population, recomputed from scratch at every time step.

**The median line.** At any date, the median divides the balances into two halves. Households change rank from year to year, so there is no reason for any one retirement to follow the line. The mean answers a different question. A few very large balances can pull it far above what the household in the middle holds.

**The example paths.** The two thin lines in the figure, green and red, are individual futures, drawn for one reason only: to show that a real path crosses the bands instead of running along one. The green one zigzags its way to prosperity. The red one starts in the pack, spends its first few years above the median even, then collapses into ruin.

Some charts select eight paths by sorting all households by their final balance. They show the first, the last and six evenly spaced ranks in between. Even if just one household in a thousand runs out, one of the eight lines can be red. To find the failure rate, divide the number of failures by the total number of simulations. The sample lines show *how* a failure unfolded.

**Zero and the scale.** At zero, the portfolio is empty. To see whether the household can still cover its bills, look at its pension income and the size of the next withdrawal ([[failure-probability]]). A chart may cap the axis at, say, ten times the starting capital. A few very wealthy households then fall outside the frame, while households running low on reserves are easier to see.

## Why the fan has the shape it has

Read three features separately: how wide the bands are, where the middle household stands and how close the lowest balances are to zero.

**Why it widens.** Without withdrawals, if annual log returns are independent and have the same variance, the standard deviation of cumulative log returns grows with the square root of time (√t). With withdrawals, the chart also reflects the effect of selling assets on the remaining balance. Its width alone cannot tell you whether bad years tend to cluster ([[monte-carlo-strengths-and-limits]]).

**A rising median.** In the figure, the middle household's balance grows even as some households run out. Both outcomes can occur in the same model. Conversely, a household may plan to spend down capital until a pension starts. Its median balance can fall while its planned spending remains fully funded ([[horizon-and-life-expectancy]]).

**The bottom of the fan dives first, or it does not.** This is the most useful feature to read, and it is worth seeing on two plans side by side.

::: figure fan-two-plans
Both plans cover the same horizon. On the left, the household has a cash reserve, spending it can adjust and income that starts early. On the right, it has less room to manoeuvre. Compare the two fifth percentiles after ten years.
:::

In the first few years the 5th percentile moves closer to zero faster on the right. To see why, open a path that started badly and look at the withdrawals after the loss. To measure the effect of the cash reserve, repeat the comparison with only the reserve changed. The two plans in the figure also differ in income and spending flexibility ([[sequence-of-returns]]).

::: science Point-in-time percentiles against paths: the percentile path fallacy
The household at the 5th percentile in year ten may be a different household from the one at that rank in year five. To find out what happened to the households in trouble early on, select all paths below the 10th percentile in year five. Follow those same paths to the end and count how many ran out. The fan alone does not give you that conditional probability ([[when-to-worry]]).
:::

## The five reading mistakes, and their fixes

**Mistake 1: reading the median as a promise.** "The simulator says I will have $2.8M at 75." No. It says half the simulated futures clear that number if the assumptions hold. Fix: the median is for ranking plans against each other and for calibrating bequest and spending decisions ([[spending-in-retirement]]); safety decisions get made on the bottom of the fan and on failure probability.

**Mistake 2: picking your band off a menu.** "I plan on the 25th percentile, that is prudent." A point-in-time percentile is not a livable scenario (the percentile path fallacy above), and prudence by percentile is not consistent over time. Fix: prudence gets set by the **models**, planning between an honest central case and a pessimistic one ([[historical-vs-parametric]]), and by the failure rate you accept, never by wandering around inside the bands.

**Mistake 3: forgetting that the fan is conditional on the model.** The fan of a central parametric model and the fan of a harsh historical replay are two different objects for the same plan. The gap between those fans shows your sensitivity to the choice of model, part of the epistemic uncertainty, the uncertainty about the unknown world: we do not know which one is right. It is often wider than any single fan, and a single fan measures only aleatory uncertainty, the randomness inside a given world. The best simulators therefore show several fans, one per market model, on the same scale. When the fans look alike, the plan is robust to the choice of model. When they diverge, that gap deserves more attention than the width of any one fan, and the decision has to hold up under the harsh models. Fix: read the difference between the fans first, the shape of each one second.

**Mistake 4: the illusion of scale.** A linear axis visually flattens the early years, where everything is decided, and dramatizes the late ones; a log axis would do the opposite and make zero impossible to plot (log 0 = −∞). The usual compromise is a clipped linear axis (see the anatomy above), which keeps zero and keeps the early years legible. Fix: whatever the tool, always ask what the chosen scale amplifies and what it hides, and find the zero line before anything else.

**Mistake 5: counting pixels instead of probabilities.** The visual area of the "ruin" zone depends on how many paths are drawn, how thick the lines are, where the axis is clipped: none of that is a probability. Fix: the fan gives you the **shape** of the risk, when it hits, how, how brutally; the **numbers** come from the failure probability and the results table. You read the two together, never one for the other.

::: astuce Putting it to work: handling a real fan
Nothing beats a live fan computed on your own plan. Fifteen minutes is enough. In the FIRE simulator, the Simulated futures section shows four fans of the same plan, on the same scale, one per market model.

- Start with the central fan, on its own. Find the median, the 25-75 band, the 5-95 band, the zero line.
- Move to the harshest one, Lost decade. Compare the two fan bottoms before you compare the two medians.
- Follow one path that ends at zero. Find the first bad year and the withdrawals that followed. Then look up the failure rate for the entire simulation in the results table.
- Finally switch the spending rule, from the fixed indexed one to a flexible one. The bottom of the fan lifts, and the fan of delivered spending shows what that safety costs you in lived spending.

The same walkthrough works anywhere else, in any tool that plots a distribution of paths (TPAW Planner and Portfolio Visualizer in percentile bands, cFIREsim and FICalc as a bundle of historical vintages). You only have to rerun the calculation model by model, then keep the fans side by side in front of you.
:::

## What did households spend, leave behind or fail to fund?

The wealth fan leaves three questions open. Each calls for a different way of looking at the paths.

**The spending fan.** With a flexible rule (guardrails, VPW, ABW, [[withdrawal-strategies-overview]]), also plot what the household could actually spend each year. If the 10th percentile is 35% below budget for twelve dates in a row, the bottom of the distribution stays low. To find the length of cuts *one household* experienced, count consecutive years below its spending floor along each individual path ([[flexibility-in-practice]]).

**Final wealth.** A histogram of the closing balances shows how often wealth ended up high or low. If the simulation stops at age 90, those are balances at age 90. Estimating the inheritance at death also requires accounting for the age at death ([[horizon-and-life-expectancy]], [[spending-in-retirement]]).

**Patterns of failure.** Among households that ran out, some lost half their capital early on. Others saw it erode slowly, or exhausted a portfolio that had first grown. These patterns tell you where to look. To understand a particular failure, compare returns, withdrawals and outside income year by year ([[glidepaths]], [[annuities-and-safety-first]]).

::: exemple How long would you have to live on less?
A household budgets $55,000 a year for 48 years and agrees to cut back if markets fall. The 25th percentile of spending falls 12% below that budget for six consecutive years. Two questions matter to the household. Would spending at that level cover its essentials ([[how-much-you-need]])? And on the paths that suffered cuts, how many years in a row did they last? The household budget answers the first question. The second takes a year-by-year look at those paths.
:::

## The essentials

- A fan chart is a stack of distributions by date, not a bundle of futures: the bands are point-in-time quantiles and real paths zigzag across them; the median is neither a path nor the average.
- A rising median can coexist with failures. Compare funded spending and the remaining balance with the plan's goals.
- Paths selected by final rank let you examine both successes and failures. Calculate the failure rate across all simulations.
- The five mistakes: median as a promise, percentile as a scenario, forgetting the model the fan is conditional on (read the gap between fans first), the illusion of scale, and pixels mistaken for probabilities.
- Choose the next chart to match your question. How much spending did the plan deliver? How much wealth remained at death? What happened to households that fell short?

---

## Going further

- Bank of England, *Inflation Report fan charts* (the historical methodology note): where the format and its conventions come from.
- Early Retirement Now, Part 46 (the false precision of simulator output) ([[the-ern-series]]).
- Multi-model fans, and the fans of spending, bequests and causes: [[using-the-fire-simulator]] and [[under-the-hood]].
- The logical next step: [[simulator-traps]] (the biases upstream of the chart) and [[when-to-worry]] (the right tool for "where do I stand on my own path?").
