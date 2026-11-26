# Reading a fan chart without fooling yourself
<!-- source: lire-un-fan-chart @ ddddedc5bcc2 -->

The densest chart in all of retirement planning is the wealth fan, the fan chart: thousands of simulated paths boiled down to percentile bands that widen with time, a few example paths, a zero line. It is also the most misread chart in the business. People see promises where there are frequencies, scenarios where there are point-in-time quantiles, and the uncertainty of the world where what dominates is sometimes the uncertainty of the model.

That chart sits at the heart of every serious FIRE simulator (FICalc, cFIREsim, Portfolio Visualizer), and the same conventions turn up far beyond retirement planning: the Bank of England, which popularized the format for its inflation forecasts, the IPCC's climate projections, forecasting statistics as a whole. This article teaches you to read one like a professional, with a real fan in front of you: its exact anatomy, why it has the shape it has (every feature of its geometry is a probability fact in disguise), the five classic reading mistakes and their fixes, and the other fans that round out the wealth fan.

After this page, a fan reads like a sentence.

::: cle The reversal to make
A fan chart does not show futures. For each date, it shows the **distribution** of the states possible on that date. The 90% band in year 20 says "in year 20, 90% of the simulated paths land in this slice". It says nothing about how any one path got there, and nothing about where it goes next. The fan is a stack of cross-sections, not a bundle of paths; real paths zigzag across the bands their whole lives. Every reading mistake comes from confusing those two objects.
:::

## Anatomy of a fan

::: figure fan-anatomy
A plan with a 45-year horizon. The nested bands are **quantiles by date** (25-75%, 5-95%), not scenarios. The median rises. And the two example paths cross the bands: the one that ends ruined started out near the top, proof that a band is not a path.
:::

Take the fan above, piece by piece.

**The nested bands.** From darkest to lightest, they cover widening percentile intervals around the median: here the middle half of the futures (25-75) and nearly all of them (5-95). At each date, the thousands of simulated wealth levels are sorted and the quantiles plotted. So a band is not a "cautious" scenario or an "optimistic" one: it is a statistic of the whole population, recomputed from scratch at every time step.

**The median line.** This is the 50th percentile: at each date, half the futures sit above it and half below. Two subtleties. First, the median is not a path. No future runs along it, and staying on it would take a miracle of regularity. Second, it is not the average either. Compounded wealth distributions are badly skewed, bounded at zero below and unbounded above, so the average gets dragged up by the opulent scenarios. When a salesman says "on average you will end up with $4M", he is usually quoting the mean because the mean flatters. The median is the one that tells you what will plausibly happen to you.

**The example paths.** The two thin lines in the figure, green and red, are individual futures, drawn for one reason only: to show that a real path crosses the bands instead of running along one. The green one zigzags its way to prosperity. The red one starts in the pack, spends its first few years above the median even, then collapses into ruin.

Some charts select eight paths by sorting all households by their final balance. They show the first, the last and six evenly spaced ranks in between. Even if just one household in a thousand runs out, one of the eight lines can be red. To find the failure rate, divide the number of failures by the total number of simulations. The sample lines show *how* a failure unfolded.

**The zero line and the clipped top.** At zero the portfolio is empty. The plan fails if a scheduled withdrawal goes unpaid ([[failure-probability]]). The vertical axis, on the other hand, is better clipped, at ten times the starting capital for instance. Without that, the scenarios compounding for 30 or 40 years, which can reach 20 to 50 times the original stake, would visually flatten the zone around zero, the one that motivates the whole exercise. It is a deliberate choice: give up the spectacle of the upper tail to keep the lower tail legible, the tail that decides.

## Why the fan has the shape it has

Every geometric feature of the fan is a theorem in disguise. Know them and you can diagnose a plan at a glance.

**It widens.** For memoryless returns, the spread of cumulative returns, counted in logarithms, grows as the square root of time (√t). In dollars, compounding stretches the fan much faster at the top, and withdrawals dig out its bottom, because a bad start opens a gap that fixed draws then amplify. So do not look for market memory in its width alone: the bottom is what speaks. Real markets, with their mild pull of valuations back toward the mean, widen a little less at very long horizons. That is one reason i.i.d. models slightly over-disperse the distant future ([[monte-carlo-strengths-and-limits]]).

**A rising median.** In the figure, the middle household's balance grows even as some households run out. Both outcomes can occur in the same model. Conversely, a household may plan to spend down capital until a pension starts. Its median balance can fall while its planned spending remains fully funded ([[horizon-and-life-expectancy]]).

**The bottom of the fan dives first, or it does not.** This is the most useful feature to read, and it is worth seeing on two plans side by side.

::: figure fan-two-plans
Two plans, same horizon. On the left, a **defended** plan (buffer, flexibility, early income): the 5th percentile sinks slowly and stays positive. On the right, a **stretched** plan with no margins: the 5th percentile dives toward zero inside the first decade. The slope of the bottom of the fan over those first ten years gives a good idea of your exposure to sequence risk.
:::

Look at the 5th percentile over the first ten years, the colored line at the bottom of each fan. That is the fragile window made visible ([[sequence-of-returns]]). A well-defended plan shows a fan bottom that sinks slowly and then recovers; a plan with a stretched withdrawal rate and no margins shows a 5th percentile diving toward zero by years 5 to 10. Two fans of comparable width can hide opposite exposures: everything is in the **slope of the bottom, early on**.

::: science Point-in-time percentiles against paths: the percentile path fallacy
The deepest conceptual error of the fan chart deserves its own box, and the anatomy figure shows it. The "5th percentile" line is **not** a scenario. It is a seam stitched from thousands of different scenarios, each holding that rank for a moment. A genuine worst-decile path usually visits the middle band in some years. The other way around, the path that ends ruined has often spent its early years in the upper half, exactly like the red path in the anatomy figure (two good years before the plunge: that is the 2000 vintage, [[the-trinity-study]]). The practical consequence: you cannot read "if I am below the 10th percentile in year 5, I will end up ruined" off the fan. The fan carries no conditional probabilities along paths. That question, where do I stand and what should I conclude, is legitimate and crucial, but it takes another tool: watching thresholds on your current withdrawal rate ([[when-to-worry]]), not staring at the fan.
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

## Beyond wealth: the other fans

In the simulators that offer them, the same distribution-by-date format supports three other readings, and each answers a question the wealth fan leaves alone.

**The fan of delivered spending.** The moment your plan runs a flexible rule (guardrails, VPW, ABW, [[withdrawal-strategies-overview]]), wealth is no longer enough: the question becomes what you will actually live on. A fan of the standard of living year by year has a bottom of its own: the standard of living in the bad quarter of futures, for how many years, paid for by what (portfolio, pension or buffer). This is the view that separates withdrawal rules. A rule that "never fails" while its 10th percentile of spending sits 35% below plan for twelve years has not removed the risk. It has moved it out of bankruptcy and into daily life ([[flexibility-in-practice]]).

**Final wealth.** A histogram of the closing balances shows how often wealth ended up high or low. If the simulation stops at age 90, those are balances at age 90. Estimating the inheritance at death also requires accounting for the age at death ([[horizon-and-life-expectancy]], [[spending-in-retirement]]).

**The causes of failure.** Among the ruined futures alone, the shape of the path tells you the failure mode: an early crash (wealth halved in the first ten years, the sequence disaster), a slow grind (the portfolio never took off, the lost decade), or longevity (the plan prospered, peaked, and the retirement outlived it). Three modes, three different defenses: a glide path and a buffer for the first ([[glidepaths]]), regime assets for the second, annuities for the third ([[annuities-and-safety-first]]). This is the view that turns "6% failure" into an actionable diagnosis.

::: exemple A full reading, in four looks
The plan: $1.6M, $55,000 a year with guardrails, a 48-year horizon. Look 1, the fans of the different models side by side: similar shapes, but the bottom of the broad-sample fan sinks noticeably faster. So the choice of model weighs more than luck, with one caveat: the broad sample does not simulate this portfolio, but a domestic 60/40. Look 2, the bottom of the central fan over ten years: a gentle slope. The red path shows what happened to withdrawals after the first decline. Look 3, the fan of delivered spending: the 25th percentile stays six years 12% below the comfort level. It is not necessarily the same household each year: how long a cut is actually lived shows path by path, by opening a few futures from the bottom. That is the real price of the guardrails, and it is judged acceptable against the floor already set ([[how-much-you-need]]). Look 4, the causes: three quarters of the remaining failures are slow grinds. The defense to test first is therefore not more cash, but building blocks against a lost decade ([[all-weather-portfolios]]). Four looks, four informed decisions. That is what a well-read fan chart delivers in two minutes.
:::

## The essentials

- A fan chart is a stack of distributions by date, not a bundle of futures: the bands are point-in-time quantiles and real paths zigzag across them; the median is neither a path nor the average.
- The geometry talks: a plunging fan bottom means withdrawals widening the gap, the slope of the fan bottom over the first ten years gives your sequence exposure (see the two plans compared).
- Paths selected by final rank let you examine both successes and failures. Calculate the failure rate across all simulations.
- The five mistakes: median as a promise, percentile as a scenario, forgetting the model the fan is conditional on (read the gap between fans first), the illusion of scale, and pixels mistaken for probabilities.
- The other fans complete the picture: delivered spending (the real judge of flexible rules), bequests (ruin seen from the other end), causes of failure (the diagnosis that picks the defense).

---

## Going further

- Bank of England, *Inflation Report fan charts* (the historical methodology note): where the format and its conventions come from.
- Early Retirement Now, Part 46 (the false precision of simulator output) ([[the-ern-series]]).
- Multi-model fans, and the fans of spending, bequests and causes: [[using-the-fire-simulator]] and [[under-the-hood]].
- The logical next step: [[simulator-traps]] (the biases upstream of the chart) and [[when-to-worry]] (the right tool for "where do I stand on my own path?").
