# Why 4%? The mathematical anatomy of the rule
<!-- source: les-maths-du-4-pourcent @ dd6b6cabe46f -->

[[the-4-percent-rule]] tells the story of where the rule came from, and [[the-trinity-study]] how it was established. This chapter answers the question almost nobody asks: why that number? Why not 6%, when US stocks have returned 6 to 7% real on average? Why not 2%, since anything can happen? The 4% is no magic constant that fell out of a backtest. It is the **residue of a three-tier calculation**, and every tier has a name, a sign and a ballpark size. Once you see the cascade, the rule stops being a dogma to believe or debunk. It becomes a result you can recompute yourself the day an assumption changes.

This is probably the most teachable chapter in the book: three tiers, one addition, two subtractions, nothing beyond a calculator. When you are done, you will be able to rebuild the number in your head, say exactly what would push it up or down, and understand why it has held up so well across a century that included 1929, 1966 and 2000.

::: cle The idea in one sentence
The safe withdrawal rate is the portfolio's real geometric return, **plus** a bonus because you are allowed to spend the capital itself (the horizon ends), **minus** a penalty because returns arrive out of order and you sell into the down years (sequence risk). On a historical 60/40 over 30 years: 4 + 1.8 − 1.8, or about 4%. The whole mystery of the rule sits in one fact. The bonus and the penalty, two large opposing forces, cancel almost exactly.
:::

## Tier 1: the perpetuity, or real return as the natural ceiling

Start with the simplest world there is: a portfolio with a constant, known return, and an immortal retiree who never wants to touch the principal. The answer fits on one line. The sustainable withdrawal is exactly the portfolio's **real** return. Real, because the withdrawal has to keep pace with inflation ([[inflation-and-withdrawal-rates]]). Any reasoning done in nominal terms is an optical illusion that costs 2 to 3 points.

That leaves the return itself to pin down. A global 60/40 has historically delivered about 4.6% real, as an **arithmetic** mean. But capital does not compound at the arithmetic mean. It compounds at the **geometric** one, cut down by volatility drag, roughly half the variance ([[arithmetic-vs-geometric-returns]]). Call it **4% real geometric** for a historical 60/40. First marker planted. A cautious immortal with a classic portfolio can take about 4%. That number is still a return, not yet a withdrawal rate. Its numerical match with the final rule is a coincidence, and a confusing one, because the next two tiers are about to cancel each other almost perfectly.

## Tier 2: the amortization bonus, or the right to die

Nobody is immortal. Over a 30-year horizon you are allowed to eat the capital itself, not just its fruit. Lending gives the exact formula, a mortgage payment run backwards: at 4% real over 30 years, the withdrawal that lands exactly on zero on the last day is about 5.8% a year. The right to finish at zero is therefore worth +1.8 points. That is the **amortization bonus**, and it is the entire logic of the ABW and VPW methods, which recompute it every year ([[amortization-based-withdrawal]], [[vpw]]).

The bonus also explains why the horizon matters so much ([[horizon-and-life-expectancy]]). Stretch it to 50 years and the same formula gives 4.7%: the bonus melts to +0.7. Stretch it to infinity and it disappears. Very early retirement does not lose rate to some mystery. It simply loses the right to amortize fast.

::: figure amortissement-horizon
The withdrawal that pure amortization arithmetic funds at 4% real, by the number of years to cover. The curve falls fast, then flattens onto its asymptote, the return alone. The first twenty years carry most of the bonus, while stretching from a 50-year horizon to eternity costs only seven tenths of a point. Very early retirement does pay for its horizon, but far less than people think.
:::

## Tier 3: the sequence penalty, or what disorder costs

Tier 2 assumed a constant return. Real returns arrive in a jumble. A fixed withdrawal turns that jumble into a lopsided risk, because selling into a decline destroys capital that will never see the rebound ([[sequence-of-returns]]). Once you are withdrawing, the ending result no longer depends on the geometric mean alone. It depends on the order too, and the first years count several times more than the last ones. The rule is meant to survive the worst orders on record (1929, 1966), not the average one, so the worst disorder has to be paid for up front.

How much does it cost? That is Bengen's central finding. Across US history the worst start year could carry only about 4%, while the amortization math at the average return promised 5.8%. The historical sequence penalty is therefore worth about −1.8 points. ([[arithmetic-vs-geometric-returns]] budgets 1 to 1.5 points for the same phenomenon; it measures from the bare geometric return, without going through amortization, which is why the two figures differ.) The penalty is not a universal constant. It grows with the portfolio's volatility: an all-stock plan has both a higher average return and a heavier penalty, which is what produces the allocation plateau ([[stock-bond-allocation]]). And it shrinks with anything that cushions the down years. Diversification and rebalancing force you to buy low during a crash ([[why-diversification-works]]). Flexibility helps too, since cutting 10% in a red year buys back part of the penalty ([[flexibility-in-practice]]).

::: figure cascade-4pct
The 4% cascade: about 4% of real geometric return, plus the amortization bonus, minus the sequence penalty. The two forces cancel almost exactly, and that is the whole "mystery" of the rule.
:::

::: exemple The whole cascade, back of the envelope
Historical 60/40, 30 years, a fixed indexed withdrawal. Real arithmetic return about 4.6%; take out volatility drag and you get about 4% geometric (tier 1); add the 30-year amortization bonus and you get about 5.8% (tier 2); take out the penalty for the worst order on record and you land on about 4.0% (tier 3). There is the rule, rebuilt. Now make it breathe. A 50-year horizon: the bonus melts, the penalty eases a little, about 3.4% (the −1.8 of tier 3 is calibrated on thirty years). A world sample instead of the US case alone ([[anarkulova-cederburg]]): 0.5 to 1 point off. A high CAPE on the day you start ([[valuations-and-cape]]): a point shaved off tier 1. Fees of 0.5%: 0.5 off, almost one for one. A flexible rule with a floor: 0.3 to 0.5 back on. Every argument about "the real number" is an argument about a single tier, and it gets settled tier by tier, not by slogans.
:::

::: figure clavier-leviers
Each assumption moved on its own, starting from the cascade's reference plan: historical 60/40, 30 years, a fixed indexed withdrawal, so 4.0%. The ranges are the ones used in this chapter, and the sign of each bar says which way the lever pushes.
:::

## Why it holds up (and what would break it)

The real surprise is not the rule's value, it is its stability. It held through two world wars, a depression, a stagflation and two twenty-first-century crashes. Three mechanisms explain that, all of them already familiar. The first is the **tiers offsetting each other**. Stretches of weak returns leave cheap valuations behind them, and cheap valuations mean better returns next. The cascade refuels itself along the way, the mechanical version of mean reversion. The second is **rebalancing**, which turns every crash into a forced purchase at a discount and chips away at the sequence penalty exactly when it bites. The third is the **hidden margin**. The rule is calibrated on the worst case in the record, so in six vintages out of ten it ends with more capital, in real terms, than it started with ([[reading-a-fan-chart]]). The median subsidizes the tail.

Which makes the list of things that would break it short and specific. A regime **outside the sample** for your own country: a retiree in Tokyo in 1990, or one in Buenos Aires, lived through worse than the worst American case ([[international-diversification]], [[simulator-traps]]). **Frictions** you never budgeted for: the cascade is computed gross, and fees and taxes come straight off tier 1, almost one for one ([[building-it-with-us-etfs]], [[us-taxes-in-the-withdrawal-phase]]). **Rigid indexation** over a very long horizon, the combination with the least slack in it. And **quitting halfway**, statistically the number one killer, against which the cascade can do nothing ([[the-psychology-of-spending]]).

::: science What modern reconstructions confirm
The cascade is more than a teaching device. It checks out term by term. ERN's reconstructions (the SWR series, [[the-ern-series]]) find the sequence penalty in how the rate depends on the starting CAPE and on the allocation. Blanchett, Pfau and Morningstar find the amortization bonus in rates that rise with age: published safe rates climb about 0.3 to 0.5 points every time the horizon shortens by five years. And Anarkulova and Cederburg put a number on the sample term by swapping US history for the world basket (0.5 to 1 point off, depending on the criterion). When this book suggests starting around 3 to 3.5% for a long FIRE rather than the canonical 4%, that is not a mood. It is the same cascade, with every tier reset to its forward-looking value ([[expected-returns]]).
:::

## The essentials

- The 4% rebuilds in three tiers: about 4% of real geometric return on a 60/40 (after volatility drag), +1.8 points of amortization bonus over 30 years, −1.8 points of penalty for the worst order of returns on record. The near-perfect cancellation of bonus and penalty is the rule's "miracle".
- Every tier has its own levers. The horizon works on the bonus (at 50 years, only +0.7). Volatility and flexibility work on the penalty. Valuations, fees and the sample work on the return tier, and fees almost one for one.
- Its historical robustness comes from mean reversion (bad decades set up the good ones), from rebalancing (forced buying at the bottom) and from the worst-case margin (the median ends richer than it started).
- What breaks the cascade is known: a regime outside your country's sample, frictions never budgeted for, rigid indexation over a very long horizon, and quitting along the way.
- Being able to redo this arithmetic in your head beats every debate: when someone announces that 4% is dead, or that 5% works fine, ask which tier they moved, and by how much.

---

## Going further

- William Bengen, "Determining Withdrawal Rates Using Historical Data" (1994): the original paper, where the cascade is implicit in the tables.
- Early Retirement Now, the SWR series (the parts on the CAPE and on the horizon in particular): the cascade recomputed on modern data ([[the-ern-series]]).
- Blanchett, Kowara & Chen, "Optimal Withdrawal Strategy for Retirement Income Portfolios" (2012) and Morningstar's State of Retirement Income reports: the amortization bonus in rates by age.
- In this book: [[arithmetic-vs-geometric-returns]] (tier 1 in detail), [[amortization-based-withdrawal]] (tier 2 turned into a strategy), [[sequence-of-returns]] (tier 3), [[the-4-percent-rule]] (the rule told as a story) and [[the-trinity-study]] (the rule measured).
