# Deciding under uncertainty: utility, Kelly, regret, and robust choices
<!-- source: decider-sous-incertitude @ 8d8c6ee56571 -->

This book spends a lot of pages answering questions like "which strategy maximizes the withdrawal rate" or "which allocation minimizes failure". This page steps back and asks the question that comes before them: **what makes a decision good when the future is unknown?** A retiree has a problem that averages ignore. You live one path. A simulator draws ten thousand, life draws one, and there is no second game. Maximizing an average computed over ten thousand worlds is not automatically the right goal for someone who will only ever live in one of them.

Decision theory is a century of work across economics, mathematics and psychology, and it answers that question directly. Expected gain turns out to be a poor criterion, and utility is what fixes it. Two risky plans can be compared in a single honest number, the certainty equivalent. The Kelly criterion, seductive as it is, does not transfer to a retirement as written. Regret and robustness, which sound like excuses, are grown-up criteria. All of it condenses into a protocol of five rules. This is the most conceptual page in the book and probably the most profitable one: picking the wrong criterion costs more than getting the arithmetic wrong.

::: cle The idea in one sentence
Between two plans, do not pick the one with the better average. Pick the one whose whole distribution you prefer, weighted by what each level of wealth is actually worth to you. One more euro when everything is going well is worth almost nothing. One more euro in a path headed for ruin is worth everything. That asymmetry is the whole of decision theory, and nearly every cautious recommendation in this book falls straight out of it.
:::

## Utility: why the average is a bad judge

The classic starting point is a puzzle posed in 1713, the St. Petersburg paradox. It is a coin-flip game with an infinite expected payoff that nobody would pay more than a few dozen euros to play. Daniel Bernoulli's answer sets up everything that follows. What counts is not money but the **utility** of money, and utility grows more and more slowly. Going from EUR 500,000 to EUR 1M changes a life. Going from EUR 10M to EUR 10.5M changes nothing. That **concavity** has a direct consequence: at equal expected value, less spread means more utility. Risk aversion is not an emotion, it is a theorem.

For a retiree the concavity is extreme and lopsided. The bottom of the distribution destroys the standard of living, the dignity, the options: those are the paths that graze or hit ruin ([[failure-probability]]). The top, where you die on a pile of gold, adds almost nothing beyond a bigger bequest ([[spending-in-retirement]]). A criterion that weighs those two zones equally, which is exactly what an average does, is structurally wrong for the job. Hence a habit that runs through this book: judge plans on the 5th percentile and the median, never on the mean ([[reading-a-fan-chart]]).

The tool that makes all of this usable is the **certainty equivalent**, the guaranteed amount you would take in exchange for a risky distribution. Take two plans with the same average. One pays EUR 42,500 a year for certain, the other pays EUR 20,000 or EUR 65,000 depending on the markets. To a reasonably cautious person, the second is "worth" maybe EUR 36,000 guaranteed. The certainty equivalent turns any lottery into one comparable number, and it drops faster the worse the plan's bad scenarios are. You never have to compute it formally. Asking "what guaranteed income would this risky plan be worth to me?" is already the right mental move, and it deflates plans with a pretty average and disastrous bad years on the spot.

::: figure utilite-ce
On a concave utility curve, a 50/50 lottery between EUR 20k and EUR 65k is worth less than its expected value of EUR 42.5k. Its certainty equivalent falls to about EUR 36k. The gap between the two is the price of the risk.
:::

## Tolerance and capacity: two dials, not one

Everyday language runs together two things a decision has to keep apart. Risk **tolerance** is psychological: the drop you can sit through without panicking or selling ([[the-psychology-of-spending]]). Risk **capacity** is objective: the drop your plan absorbs without breaking, however calm you happen to be. A phlegmatic ex-trader running a tight 4.5% withdrawal has high tolerance and low capacity. An anxious person sitting on 50 times their spending has the opposite. The rule for combining them is simple and has no exceptions: the smaller of the two governs. Capacity can be computed, and that is exactly what a simulator does. Tolerance, unfortunately, is discovered mostly in real declines. That is what makes the fit tests each strategy comes with worth running ([[choosing-your-strategy]]), along with questions asked while you are calm, such as "can your withdrawal drop 20% in a single year?".

::: science Kelly: the brilliant criterion you should not follow
The Kelly criterion (1956) answers a real question elegantly. What fraction of your capital should you stake on a favorable bet to maximize long-run geometric growth ([[arithmetic-vs-geometric-returns]])? Applied to markets, it typically calls for equity exposure of 120 to 200%.

Three of its assumptions are disqualifying for a retirement. It assumes an infinite horizon, while a retiree has 30 to 50 years and dated spending to fund. It is indifferent to drawdowns, and full Kelly rides calmly through 50% declines (odds of about 1 in 2 of meeting one) and 90% declines (1 in 10), which withdrawals turn into ruin through sequence risk ([[sequence-of-returns]]). And it needs the true parameters. A small estimation error on the high side of Kelly destroys growth instead of maximizing it, because the error is asymmetric.

The lesson worth keeping is **fractional Kelly**. Serious practitioners bet half Kelly or less, trading a little growth for a lot of peace. The equity position of a well-built withdrawal plan lands, without anyone aiming for it, somewhere around a quarter to a third of Kelly. If a product or a blog sells you full Kelly for your retirement, it read the formula and not the assumptions.
:::

## Regret, good enough, and robustness

Utility does not capture everything that matters. Three more ideas round out the toolbox, and all three are more respectable than their reputation.

**Regret.** Harry Markowitz, who invented portfolio optimization, said he put his own savings in a 50/50 split "to minimize my future regret" rather than at the optimum his own theory pointed to. The minimax regret criterion picks the option whose worst backward-looking regret is the smallest. It is perfectly rational when you only live one path. It explains why a cautious plan still holds stocks, out of regret at missing thirty years of gains, and why an aggressive plan still holds bonds, out of regret at the crash in year one. Those choices beat the corner solutions. Many of the lukewarm-looking answers in this book are exactly that: the size of the gold sleeve, the withdrawal corridor, the broadened 60/40. They are deliberate regret minimizations, and that is a compliment.

**Good enough.** Herbert Simon called it satisficing: look for a solution that is sufficient rather than the best one. A plan with a 90% success rate and an income that covers your life has nothing to gain from further optimization, and plenty to lose in complexity, fees and fragility. Simon's question ("good enough for what?") serves a retiree better than the optimizer's ("maximal under which model?"). The honest answer to that second one is always "under a model that is wrong" ([[monte-carlo-strengths-and-limits]]).

**Robustness.** A robust decision stays good even when the model that produced it is wrong. Portfolio optima are famously fragile ridges: move the expected return by one point and the "optimal" allocation moves by twenty. Nature grants us one mercy, though, because the surface around the optimum is flat. Between 50 and 80% stocks the SWR barely moves ([[stock-bond-allocation]]), and the same holds between 5 and 15% in gold or trend. The practice follows from that terrain. Find the plateau, settle in the middle of it rather than at the edge where one modeling error tips everything over, and walk away from any recommendation that only exists on a ridge. That is robust optimization in the version you can run without mathematics: optimize against the worst plausible assumption, not for the best estimate.

## The protocol: five rules for deciding

All of it condenses into a protocol you can apply to every decision in the plan, from the allocation to the exit date ([[one-more-year]]).

One: judge decisions on the **process**, never on the outcome of one path. A good decision can turn out badly and a stupid bet can pay off; over a single life, confusing the two is the most expensive mistake there is. Two: compare plans on the **5th percentile and the median** together, your homemade certainty equivalent, not on the average and not on the best case. Three: prefer the **middle of a plateau** to the top of a ridge, and distrust any option whose advantage vanishes when you move one assumption by a point. Four: when the choice is close, minimize the regret you can see coming, in both directions, the crash and the rally you missed. Five: write the decision down, with the conditions for revisiting it, **while you are calm** ([[building-your-plan]]), because the best decision theory in the world does not survive a call made at 11 at night in March 2020.

::: exemple Two plans, two criteria, two winners
Capital of EUR 1.2M, spending of EUR 43,000 a year. Plan A is 85% stocks with a fixed withdrawal: a 30-year median of EUR 4.1M, a 5th percentile that runs out in year 24, and a magnificent average. Plan B is 65% stocks, diversified, with a flexible corridor: a median of EUR 2.9M and a 5th percentile that ends at EUR 400,000, with the withdrawal cut to EUR 36,000 at the worst moment. On the average, A wins by a mile. On the certainty equivalent of a normally cautious person, B wins with no argument, because A's extra EUR 1.2M of median weighs less than its 5% chance of destitution. A simulator settles the question well only when you ask it the right one. The question was never "which plan pays more?" but "which distribution would you rather live in?".
:::

## The essentials

- You live one path, so an average taken over ten thousand worlds is not your criterion. Utility is, concave and brutally lopsided around ruin, and it is why every plan gets judged on the 5th percentile and the median.
- The certainty equivalent ("what guaranteed income would this risky plan be worth?") converts any two plans into comparable numbers, and it deflates a pretty average with an ugly tail every time.
- The decision obeys the smaller of two dials: tolerance, psychological and discovered in declines, and capacity, objective and computed by a simulator.
- Kelly maximizes geometric growth under assumptions (infinite horizon, no spending, known parameters) that are all false in retirement. Keep the idea of fractional Kelly, stay away from full Kelly.
- Minimized regret, satisficing and robustness (the middle of the plateau, not its edge) are grown-up criteria, not surrender. The protocol fits in five rules, and the fifth one, deciding in writing while calm, carries the other four.

---

## Going further

- Nicolas Bernoulli (1713) stated the St. Petersburg paradox and Daniel Bernoulli (1738) resolved it with utility; Von Neumann and Morgenstern for expected utility theory.
- Herbert Simon, "A Behavioral Model of Rational Choice" (1955): satisficing from the man who coined it, and won a Nobel for it.
- Edward Thorp, "The Kelly Criterion in Blackjack, Sports Betting, and the Stock Market": the best account of Kelly, assumptions included.
- Kahneman and Tversky, "Prospect Theory" (1979): loss aversion measured, and the bridge to [[the-psychology-of-spending]].
- In this book: [[failure-probability]] (picking your threshold), [[choosing-your-strategy]] (the protocol applied to withdrawal rules), [[stock-bond-allocation]] (the plateau), [[monte-carlo-strengths-and-limits]] (why every model is wrong, and how to decide anyway).
