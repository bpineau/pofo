# Monte Carlo: strengths, weaknesses, and how to use it well
<!-- source: monte-carlo-forces-faiblesses @ 1d1b25984835 -->

Behind every failure probability, every wealth fan, every "your plan works in 94% of cases", sits the same machine: Monte Carlo simulation. It generates thousands of possible futures, then counts what happens in them. It is the central tool of modern planning, the engine under every serious simulator. It is a beautiful instrument, as long as you know what it actually does. It does not predict the future. It works out the consequences of your assumptions, with a rigor no back-of-the-envelope calculation can match.

This page takes the machine apart. Where the method came from and how it runs, step by step. What it does better than the alternatives: historical replay, closed-form formulas, and gut feel. Its four structural weaknesses and how much each one really costs you. Then how to use it sensibly, which is what separates the reader who informs a decision from the one who lets a random number generator tell them a story. Two articles carry this one forward: the model families that feed the machine ([[historical-vs-parametric]]) and the corrections that make it relevant ([[making-monte-carlo-relevant]]).

::: cle What a simulation really is
A Monte Carlo run holds no information about the future. It holds three things: your market assumptions (a distribution of returns), your plan (capital, withdrawals, rules) and a die. Its output, the failure probability, is a **theorem**. It says: "if the world draws its years from this distribution, then this plan fails x% of the time". All the value sits in the if. Used well, Monte Carlo is a microscope trained on your assumptions. Used badly, it is a machine for laundering hopes into three decimal places.
:::

## Where it comes from, and how it actually works

The method was born in 1946 at Los Alamos. Stanislaw Ulam, recovering from an illness, was playing solitaire and wondered what fraction of games could be won. The exact combinatorics were hopeless. It struck him that he could simply play a large number of games and count. With von Neumann and Metropolis, the idea became a method. They named it after the casino where Ulam's uncle lost his money, and it cracked the neutron diffusion calculations of the bomb. The principle is general: when a system is too complex for a formula, run it thousands of times and look at the distribution of outcomes.

A retiree's problem is exactly that kind of system. A withdrawal plan has memory. The withdrawal in year 12 depends on the portfolio in year 12, which depends on the whole sequence before it, on the spending rules, on the buffer, on taxes, and on the pension that starts in year 15. No closed-form formula captures that once the plan has any realism to it. Simulation does. Simulators carry different amounts of detail, but the most complete ones (TPAW Planner or Portfolio Visualizer, say) run every path like this:

1. **Draw a sequence of real returns** for the whole horizon, from the market model you picked: independent draws from a parametric law, ideally a Student-t (a bell-shaped cousin of the normal with fat tails, where extreme years stay far more frequent), blocks of history for a bootstrap, a real window for a historical cohort, Markov regimes for a sequence stress ([[historical-vs-parametric]]).
2. **Run year 1**: apply the return to the portfolio, compute the withdrawal under the active rule (fixed indexed, flexible, guardrails, VPW, ABW and the rest, grossing every sale up for tax), draw on the buffer or refill it under its own rules, collect pensions and side income once they have started.
3. **Repeat** year after year to the end of the horizon, or until the money runs out (ruin, with its date).
4. **Record everything**: ruin or not, the date of ruin, ending wealth, the spending actually delivered each year, the time spent underwater.

Then start over a few thousand times and count. The share of ruined paths gives the failure probability. The year-by-year wealth gives the fan ([[reading-a-fan-chart]]). The spending delivered gives the distribution of the standard of living actually lived, and so on down the list. There is nothing else in the box: exact accounting applied to futures drawn out of a hat. (How that loop is really implemented is in [[under-the-hood]].)

## What Monte Carlo does better than anything else

To size the tool up, set it against its three rivals.

**Against pure historical replay** (Bengen's method, [[the-trinity-study]]). American history holds barely a hundred years, and the 30-year windows cut out of it overlap almost completely: that comes to three or four genuinely independent retirements. Replay answers "how would my plan have done in the past?". By construction it can say nothing about futures that resemble no past. And it treats the worst historical vintage as a bound, when nothing says the worst has already happened. Monte Carlo generates tens of thousands of synthetic years instead. It explores the space of possibilities around your assumptions, not just the one path that happened. The two complement each other, which is why good simulators show them side by side.

**Against closed-form formulas** (analytic ruin expectations, rules of thumb). Formulas need unrealistic assumptions to stay solvable: normal returns, proportional withdrawals, no rules at all. Add a guardrail floor, a deferred pension, a buffer with a refill rule and tax on sales, and only simulation keeps up. That is its distinctive strength: plan complexity is free. A realistic rule costs a few lines of code, never an approximation.

**Against intuition**, last and most important. The human mind is catastrophically bad at compounding randomness over 40 years. Nobody has a feel for what 11% volatility does to a 3.8% withdrawal over 45 years, with a pension landing in year 16. The simulation does, exactly. Its psychological work cuts both ways. It sobers up the optimists, because the fan holds ruined paths even under "good" assumptions. And it settles the anxious, because the median of a sensible plan is surprisingly plush ([[the-psychology-of-spending]]).

## The four structural weaknesses, and how much each one matters

::: figure mc-entrees-vs-tirages
The same plan, one setting changed at a time. On top, μ moves half a point up, then half a point down. Failure goes from 4.1% to 8.7%, a factor of 2.1, and no data series can settle which of the three values is right. Below, the number of paths goes up tenfold with the assumptions frozen. The result does not move; only the uncertainty bar shrinks, from plus or minus 1.5 points at 1,000 paths to plus or minus 0.5 at 10,000. **All the precision that matters is on the input side.**
:::

**Weakness 1: garbage in, garbage out, with leverage.** The output is hypersensitive to the inputs. Half a point of difference on μ, statistically undetectable ([[expected-returns]]), can double the failure probability. The simulator even amplifies the apparent precision: three decimals of failure, computed off a μ known to within a point. Severity: the worst of the four, and entirely manageable. Input discipline (forward-looking calibration, blending, anchors) and reading in ranges are the answer ([[failure-probability]]). This is the whole reason for running several models.

**Weakness 2: independent draws.** Naive Monte Carlo draws each year independently of the last, coin flip by coin flip, with no memory. Real markets have memory. Bad years clump together into recessions and multi-year bear markets. Valuations create decade-long trends ([[valuations-and-cape]]). Volatility comes in clusters. The consequence is precise. At the same mean and variance, the i.i.d. model understates the odds of long mediocre stretches, the ones that kill retirees ([[sequence-of-returns]]). It also overstates dispersion a little at very long horizons, because real markets mean-revert, which pulls the 30-year fan back in. Severity: real but quantifiable, worth a few points of failure probability. The fixes exist, and a good simulator carries three: blocks, Markov regimes and cohorts ([[making-monte-carlo-relevant]]).

**Weakness 3: the shape of the distribution.** Plenty of commercial simulators draw their returns from a normal distribution. It assigns catastrophic years (−35% real) astronomically small probabilities, so 2008 becomes an "impossible" event. A Gaussian Monte Carlo is structurally optimistic about the tails. Severity: serious, but it corrects well. Draw from a Student-t whose df is fitted to the kurtosis of observed returns (kurtosis measures how fat a distribution's tails are: the higher it runs, the more frequent the extreme years). Disasters then live inside the model instead of sitting in its blind spot ([[fat-tails]]). Check this before you trust a tool: many do not even document the law they draw from ([[simulator-traps]]).

**Weakness 4: what is not in the model does not exist.** No draw contains confiscation, hyperinflation wiping out the currency you count in ([[hyperinflation-and-extremes]]), a broker's fraud, a divorce, or long-term care at EUR 300 a day. The model simulates the market risk of a plan whose parameters never change. Life's risks stay off screen. Severity: structural and irreducible. Which is why a simulator's output replaces neither your margins ([[ten-plan-wrecking-mistakes]]), nor insurance, nor piloting along the way ([[when-to-worry]]).

::: science How many paths do you need?
The sampling error on a probability p estimated over N paths is about √(p(1−p)/N). At 4,000 paths and p = 5%, it comes to plus or minus 0.3 points. That is far finer than the 2 to 3 points of uncertainty in the inputs. Going to 10,000 smooths the sensitivity curves and the extreme percentiles of the fans, which helps when you compare rules closely. Past that you are polishing a figure whose uncertainty comes from somewhere else. The real precision of a Monte Carlo never lives in N. It lives in the inputs and in the structure of the model. Be wary of tools that advertise "100,000 simulations" as proof of seriousness. That confuses the resolution of the photo with the focus.
:::

## Using it well: eight rules

Here is the practical summary, forged by the literature (Kitces has written a great deal on the right use of Monte Carlo from the advisor's side).

**1. Spend ten times more care on the inputs than on reading the outputs.** An hour on the real budget and the calibration of μ ([[how-much-you-need]], [[expected-returns]]) beats ten hours contemplating fans.

**2. Never read a single model.** The range between models (central, sequence stress, broad sample, historical) is the information. One model on its own is just an opinion ([[failure-probability]]).

**3. Read it as a ranking.** Comparing (plan A against plan B, lever X against lever Y) is what the tool is good at. Measuring ("my risk is 4.7%") is what it is bad at. The gaps are signal, the decimals noise.

**4. Use it as a what-if machine, not an oracle.** That is its real calling. What if I leave two years earlier? What if inflation adds half a point to spending? What if the pension is cut by 20%? The best tools industrialize this with an equivalent-moves solver, which converts every margin into euros, into years, or into flexibility.

**5. Look at the paths, not just the aggregates.** A probability summarizes. The sample paths in the fan, the red ones included, show how failure arrives: fast, through an early crash, or slowly, through erosion ([[reading-a-fan-chart]]). The failure mode dictates the defense, not the failure rate.

**6. Simulate your real rules, not the rigid caricature.** Then look at the standard of living delivered, not just survival. A flexible rule always "succeeds" in the ruin sense. The real question becomes what it makes you live on ([[flexibility-in-practice]]).

**7. Re-simulate rarely.** The plan gets re-simulated at the annual review ([[the-annual-review]]) or on an event, never every week. The output moves with the markets; your plan should not.

**8. Decide on the pessimistic scenarios, live on the central one.** The plan has to be acceptable under the harsh models, the broad sample and the sequence stress. But once the decision is made, the central case is the one describing your likely life. Mixing the two registers produces either reckless retirees or permanently anxious ones.

::: exemple One session, used properly
The question: "can I leave in 2027, or do I have to wait until 2029?". The bad version: run the simulator on 2027, get 93.8%, decide it clears. The good version: freeze the audited inputs, then compare the two plans under several models. A typical result: 2027 gives 4%/7%/11%/6% failure (central, stress, broad sample, historical), 2029 gives 2%/4%/6%/3%. The gap between the two dates, about 5 points in the harsh models, is what two more years of working life buy in safety. An equivalent-moves solver shows that EUR 800 a month of side income over the first five years buys roughly the same gap. The decision (leave in 2027 with a gentle part-time job, [[going-back-to-work]]) is not in the simulator. It sits in a trade-off the simulator turned into numbers, and that is the whole difference between informing a choice and being told a story.
:::

## The essentials

- Monte Carlo is your assumptions plus your plan plus a die, run through exact accounting over thousands of futures: a conditional theorem, never a prediction.
- Its strengths: exploring beyond the past that happened, absorbing the complexity of real plans for free (rules, pension, buffer, taxes), and correcting an intuition that cannot compound 40 years of randomness.
- Its weaknesses, by severity: sensitivity to the inputs (managed by discipline and several models), independent draws (fixed by blocks, regimes and cohorts), Gaussian tails (fixed by the Student-t), and what sits outside the model (irreducible: margins and piloting).
- 4,000 to 10,000 paths are enough: a Monte Carlo's precision comes from its inputs, never from its N.
- The eight rules boil down to one: use it to compare choices under several models, with audited inputs. Never to hear a reassuring number to three decimals.

---

## Going further

- Michael Kitces, "The Problem With Monte Carlo Analyses In Retirement Projections" and the related pieces ([kitces.com](https://www.kitces.com)): the right use, seen from the advisory desk.
- Early Retirement Now, part 8 (the technical appendix) and part 46 (false precision) ([[the-ern-series]]).
- Metropolis and Ulam, "The Monte Carlo Method" (1949): the founding paper, and surprisingly readable.
- Next in this book: [[historical-vs-parametric]] (the three families of return sources), [[fat-tails]] (the Student-t choice), [[making-monte-carlo-relevant]] (blending, regimes, stress) and [[reading-a-fan-chart]] (reading the outputs).
