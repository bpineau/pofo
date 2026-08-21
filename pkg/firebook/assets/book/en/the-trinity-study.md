# Bengen, the Trinity study, and the birth of the safe withdrawal rate
<!-- source: etude-trinity @ 6910af79a617 -->

Before 1994, one question got the same wrong answer with impressive consistency: how much can I take out of my portfolio every year? "The average return, obviously." Stocks returned 10% historically, so withdraw 8%, said professional advisers with a straight face.

William Bengen, and then the Trinity study, showed why that answer ruins people, and invented the method that still organizes the whole field: replay history. This page covers what those founding papers actually established, how their method works, what was brilliant about it, and what has aged.

It is a piece of intellectual history as much as a technical one: the concepts introduced here (vintage, SAFEMAX, success rate) run through the rest of the book.

::: cle The founding reversal
Bengen's contribution is not the number "4%". It is the demonstration that the sustainable withdrawal rate does **not** depend on the average return, but on the worst run of returns and inflation the retiree lives through, above all in the first ten years. The average American vintage carried more than 6%; the 1966 vintage carried about 4%. Planning means planning for the tail, not for the mean. The whole modern subject flows from that reversal ([[sequence-of-returns]]).
:::

## Why the "average return" answer ruins people

The intuition "stocks make 10%, so I can withdraw 8%" makes two mistakes, and they compound.

The first is confusing the arithmetic average with the growth that actually compounds. Volatility makes a portfolio grow more slowly than its annual average ([[arithmetic-vs-geometric-returns]]), and inflation takes another 2 to 3 points off. The engine of a balanced portfolio is a geometric **real** return of 3 to 5%, not 10.

The second is deadlier. A retiree who withdraws a fixed amount sells more shares when prices are low. Two sequences with the **same** average then produce opposite fortunes, depending on whether the bad years land early or late ([[sequence-of-returns]]). The average tells you almost nothing; the order tells you almost everything.

In the early 1990s Bengen, an MIT-trained engineer turned financial adviser, kept meeting clients who had been sold the 8% number. Rather than answer one opinion with another, he did what nobody had published. He tested it.

## Bengen 1994: the vintage method

The idea is beautifully simple. Take the annual US data since 1926: S&P stocks, intermediate government bonds, inflation (Bengen used the Ibbotson series). Invent a retiree who leaves on January 1, 1926 with a 50/50 portfolio and an initial withdrawal of, say, 4% indexed to inflation. Run the years one at a time: return, withdrawal, rebalance. Write down how long the portfolio lasts. Start over for someone leaving in 1927, then 1928, and so on: each start year is a **vintage** (a cohort, ERN's "retirement cohort"). Then run the whole thing again at other withdrawal rates.

The result fits in one famous chart: how long the portfolio lasted, vintage by vintage, at each rate. At 3%, every vintage clears 50 years. At 4%, the worst vintage lasts 33 years, and **all** of them last at least 30. At 5%, the late-1960s vintages run dry in about 20 years. At 6%, dozens of vintages fail inside 20 years.

Bengen would later name **SAFEMAX** the highest rate that survives every vintage over the chosen horizon: about 4.15% over 30 years at 50/50. He also picked out the three worst moments to leave: 1929 (crash and deflation), 1937, and above all **1966**, which was not the worst crash but the worst **combination**, fifteen years of flat real markets while inflation kept inflating the withdrawals. The lesson matters more than the number: what kills a retiree is not the spectacular crash, it is long, grinding real erosion ([[inflation-and-withdrawal-rates]]).

His later papers (1996 to 2006) filled in the frame. The best allocation holds 50 to 75% stocks, because going lower **lowers** the safe rate: bonds alone do not stand up to inflation. Adding small caps lifts SAFEMAX, to 4.3% in 1997 with 30% of the equity sleeve in small caps, then to 4.5% in 2006 with a wider set of assets. Horizon matters too: about 4.15% over 30 years, but only about 3.5% over a very long one.

::: encart Why the method was brilliant, and what it is still worth
Thirty years on, historical replay (the "historical windows") is still one of the three big families of models that withdrawal simulators run. Its strength: it keeps everything synthetic models struggle to reproduce, real sequences (a crash **then** inflation **then** a recovery), stock-bond correlations that shift, long memory. Its weakness: it holds only the American past, a single-country sample, and the luckiest country at that, with windows that overlap (since 1926 there are only three or four genuinely independent 30-year stretches). Hence the modern corrections: a world sample ([[anarkulova-cederburg]]; by convention that link also covers the broad-sample model of the simulators, replayed on 16 countries of the JST panel, the practical cousin of the 38-country paper), bootstrapping and parametric models ([[historical-vs-parametric]]).
:::

## Trinity 1998: from a floor to a probability

Four years later, three finance professors at Trinity University in Texas, Philip Cooley, Carl Hubbard and Daniel Walz, published "Retirement Savings: Choosing a Withdrawal Rate That Is Sustainable". Same replay method, one conceptual shift: instead of the floor rate that survives **everything** (Bengen's SAFEMAX), they published a **grid of success rates**. For each combination of withdrawal rate, allocation and horizon, it gives the share of historical windows in which the portfolio ended with money left.

A slice of the grid (figures from the 2011 update, inflation-indexed withdrawals, 1926 to 2009 data):

| Initial rate | 100% stocks, 30 years | 75/25, 30 years | 50/50, 30 years | 25/75, 30 years |
|---|---|---|---|---|
| 3% | 100% | 100% | 100% | 100% |
| 4% | 98% | 100% | 96% | 80% |
| 5% | 80% | 82% | 67% | 31% |
| 6% | 62% | 60% | 51% | 22% |

Three lasting lessons come out of that grid.

**The cliff.** Between 4 and 5%, success collapses. The subject is nonlinear, and that is why "just a little more" costs so much.

**Allocation cuts one way.** Holding too few stocks is far more dangerous than holding too many: the 25/75 fails one time in five, where the 75/25 never failed at all.

**The success rate itself.** Trinity is where failure probability became the common language of the field, the one every modern simulator speaks ([[failure-probability]]).

::: attention What Trinity's "95% success" means, and what it does not
Trinity's percentage counts overlapping historical windows from the US market alone. "95%" means "95% of the retirements beginning between 1926 and 1980 would have held", not "your plan has a 95% chance of working". And "success" is counted in the weakest sense there is: a balance still positive on the last day, even if it is one dollar. The windows share their years (the 1929 crash shows up in dozens of them), the independent sample is tiny, and the future is not drawn from that urn. The probabilities modern simulators print carry cousin limits ([[simulator-traps]], [[reading-a-fan-chart]]). The defense never changes: run several models side by side and keep margins.
:::

## What the founders had not seen yet

Reading Bengen and Trinity today means admiring the method and measuring how far the field has come since. Their blind spots, each covered elsewhere in this book, map out the program of modern research.

- **The FIRE horizon.** Thirty years was the horizon of a 65-year-old. Over 45 to 55 years, the US SAFEMAX drops toward 3.25 to 3.5% ([[the-ern-series]]) and Trinity's grid no longer applies as printed ([[horizon-and-life-expectancy]]).
- **Geographic survivorship bias.** The Ibbotson data starts in 1926 in the United States, the country that won two world wars on other people's soil and came out of them a superpower. The world sample (16 countries, 1870 to 2020) tells a harder story ([[anarkulova-cederburg]]). That panel is what the "broad sample" models replay.
- **Valuations.** Bengen noted the link between starting market levels and SAFEMAX as early as 2006; formalizing it (CAPE to initial rate) came later ([[valuations-and-cape]], [[cape-based-rules]]).
- **A rigid withdrawal.** Bengen's retiree runs the rule for 30 years without looking up. The next generation of strategies (Guyton-Klinger [[guyton-klinger]], modern guardrails [[morningstar-guardrails]], amortization [[amortization-based-withdrawal]]) starts from the opposite idea: react to what you learn.
- **Fees, taxes, real spending.** Out of scope for the founders, first order in real life ([[how-much-you-need]]).

None of this refutes them. The vintage method still stands; what widened is its inputs and its frame. Bengen has kept updating his own number, upward for the classic 65-year-old American retiree holding a more diversified portfolio, while repeating that it depends on the frame. So when you hear that "the 4% rule is dead" or that "4% is far too timid", the question to ask is the same either way: in what frame, over what horizon, with what margins?

## Redo Bengen yourself

One of the method's great teaching virtues is that anyone can redo it. The classic historical simulators, FICalc and cFIREsim first among them, replay the vintage logic on the long US indexes, one withdrawal rate at a time. A simulator that reads a real portfolio replays it on the history of **your own** holdings and hands down a verdict start date by start date. Either exercise is worth the hour. Watching your plan go through 1966 or 2000 makes sequence risk concrete in a way no probability does.

::: exemple Reading a vintage
The plan: $1M, a 60/40, 4% indexed. Replayed vintage by vintage, the January 2000 start shows what a bad vintage looks like: two crashes in the first decade, the real portfolio cut in half by 2009, a recovery that never catches up with the good vintages, and capital still below half its starting value a quarter of a century later, winded but solvent. The 2009 start sails far above it. Same rule, same portfolio, same long-run average: only the **start date** differs. That is sequence risk made visible, and the best possible way into [[sequence-of-returns]].
:::

## The essentials

- Bengen (1994) invented the vintage replay and SAFEMAX: the rate that survives the **worst** start in the record, about 4.15% over 30 years in the United States. The "4%" is an American worst-case floor, not an average.
- Trinity (1998) turned that floor into a grid of success probabilities and gave the field the language of ruin; its grid shows the cliff between 4 and 5% and the danger of holding too few stocks.
- The enemy they identified is not the crash but long real erosion (the 1966 vintage): inflation on top of a flat market.
- The blind spots (FIRE horizons, the American bias, valuations, rigidity, fees and taxes) define modern research. That is what the rest of this part is about.
- The method can be redone on your own plan with any historical replay simulator: do it, one vintage lived through beats a thousand probabilities.

---

## Going further

- William Bengen, "Determining Withdrawal Rates Using Historical Data", *Journal of Financial Planning*, October 1994 (freely available on the FPA site); and *Conserving Client Portfolios During Retirement* (2006) for the synthesis.
- Cooley, Hubbard & Walz, "Retirement Savings: Choosing a Withdrawal Rate That Is Sustainable", *AAII Journal*, February 1998, and its updates (2011).
- Early Retirement Now, SWR Series parts 1 and 8 (the technical appendix to the method): [earlyretirementnow.com](https://earlyretirementnow.com) ([[the-ern-series]]).
- Wade Pfau, "An International Perspective on Safe Withdrawal Rates" (2010): the first big step outside the American frame, a prelude to [[anarkulova-cederburg]].
- In this book: [[the-math-of-4-percent]] (why Bengen's number holds mathematically, layer by layer).
