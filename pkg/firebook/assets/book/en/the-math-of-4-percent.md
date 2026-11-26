# Why 4%? The mathematical anatomy of the rule
<!-- source: les-maths-du-4-pourcent @ e3fb62326312 -->

[[the-4-percent-rule]] tells the story of where the rule came from, and [[the-trinity-study]] how it was established. This chapter answers the question almost nobody asks: why that number? Why not 6%, when US stocks have returned 6 to 7% real on average? Why not 2%, since anything can happen? The 4% is no magic constant that fell out of a backtest. You can read it **in three tiers**, comparing the withdrawal at a constant return with what the worst historical starts could sustain. Once you see the cascade, the rule stops being a dogma to believe or debunk. It becomes a result whose workings you understand, and whose direction you can predict the day an assumption changes.

This is probably the most teachable chapter in the book: three tiers, one addition, two subtractions, nothing beyond a calculator. When you are done, you will be able to rebuild the number in your head, say exactly what would push it up or down, and understand why it has held up so well across a century that included 1929, 1966 and 2000.

::: cle The idea in one sentence
With a **constant** 4% real return, you can withdraw 4% of the starting portfolio at the end of every year indefinitely, or about 5.8% a year for thirty years. In Bengen's US study, a withdrawal starting around 4% and then adjusted for inflation survived the difficult thirty-year periods. The gap from 5.8% reflects both the returns those retirees earned and the order in which they earned them.
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

Start with $100 and withdraw $10 at the end of each year. If the market falls 20% then rises 25%, you have $70 left after the first withdrawal and $77.50 after the second. Reverse those returns and you have $115, then $82. With no withdrawals both orders would have taken you back to $100. With withdrawals, an early loss leaves less money to benefit from the recovery ([[sequence-of-returns]]).

Bengen compared people who retired in different years in the United States. They experienced different thirty-year returns and faced high inflation at different stages of retirement. The roughly 1.8-point gap between our theoretical payout and the historical withdrawal rate reflects those differences. To isolate the effect of order alone, keep the same real returns and rearrange them, as in the $100 example.

::: figure cascade-4pct
The first two steps assume a constant return. The last connects that calculation to the US historical result. The bar labelled “sequence” also reflects the level of returns during difficult periods.
:::

::: exemple The whole cascade, back of the envelope
Historical 60/40, 30 years, a fixed indexed withdrawal. Real arithmetic return about 4.6%; take out volatility drag and you get about 4% geometric (tier 1); add the 30-year amortization bonus and you get about 5.8% (tier 2); subtract the gap to the worst historical start and you land on about 4.0% (tier 3). There is the rule, rebuilt. Now make it breathe. A 50-year horizon: the bonus melts, the penalty eases a little, about 3.4% (the −1.8 of tier 3 is calibrated on thirty years). A world sample instead of the US case alone ([[anarkulova-cederburg]]): 0.5 to 1 point off. A high CAPE on the day you start ([[valuations-and-cape]]): a point shaved off tier 1. Fees of 0.5%: 0.5 off, almost one for one. A flexible rule with a floor: 0.3 to 0.5 back on. Every argument about "the real number" is an argument about a single tier, and it gets settled tier by tier, not by slogans.
:::

::: figure clavier-leviers
Each assumption moved on its own, starting from the cascade's reference plan: historical 60/40, 30 years, a fixed indexed withdrawal, so 4.0%. The ranges are the ones used in this chapter, and the sign of each bar says which way the lever pushes.
:::

## Why it holds up (and what would break it)

A rate chosen to survive the most difficult US starting date often leaves considerable wealth to households that retired in better years ([[reading-a-fan-chart]]). Why keep that margin? When you retire, you do not know which group you will join. Another household's ending wealth cannot cover your expenses if your own retirement gets off to a poor start.

Which makes the list of things that would break it short and specific. A regime **outside the sample** for your own country: a retiree in Tokyo in 1990, or one in Buenos Aires, lived through worse than the worst American case ([[international-diversification]], [[simulator-traps]]). **Frictions** you never budgeted for: the cascade is computed gross, and fees and taxes come straight off tier 1, almost one for one ([[building-it-with-us-etfs]], [[us-taxes-in-the-withdrawal-phase]]). **Rigid indexation** over a very long horizon, the combination with the least slack in it. And **quitting halfway**, statistically the number one killer, against which the cascade can do nothing ([[the-psychology-of-spending]]).

::: science What modern reconstructions confirm
The studies listed below answer different questions. Early Retirement Now compares withdrawals by starting valuation and asset allocation ([[the-ern-series]]). Blanchett and his coauthors examine how spending can adjust over retirement. Anarkulova and Cederburg broaden the sample of markets. Before comparing two published rates, check the horizon, the markets included and the definition of failure.
:::

## The essentials

- The 4% rebuilds in three tiers: about 4% of real geometric return on a 60/40 (after volatility drag), +1.8 points of amortization bonus over 30 years, −1.8 points of difference from the worst historical start (both the level and order of returns). The near-perfect cancellation of bonus and penalty is the rule's "miracle".
- Every tier has its own levers. The horizon works on the bonus (at 50 years, only +0.7). Volatility and flexibility work on the penalty. Valuations, fees and the sample work on the return tier, and fees almost one for one.
- To see the effect of order by itself, rearrange the same returns while keeping withdrawal amounts and dates fixed.
- What breaks the cascade is known: a regime outside your country's sample, frictions never budgeted for, rigid indexation over a very long horizon, and quitting along the way.
- Being able to redo this arithmetic in your head beats every debate: when someone announces that 4% is dead, or that 5% works fine, ask which tier they moved, and by how much.

---

## Going further

- William Bengen, [“Determining Withdrawal Rates Using Historical Data”](https://www.financialplanningassociation.org/article/journal/MAR04-determining-withdrawal-rates-using-historical-data) (1994). The charts show how long portfolios lasted by starting year and initial withdrawal rate.
- Early Retirement Now, the SWR series (the parts on the CAPE and on the horizon in particular): the cascade recomputed on modern data ([[the-ern-series]]).
- Blanchett, Kowara & Chen, "Optimal Withdrawal Strategy for Retirement Income Portfolios" (2012) and Morningstar's State of Retirement Income reports: the amortization bonus in rates by age.
- In this book: [[arithmetic-vs-geometric-returns]] (tier 1 in detail), [[amortization-based-withdrawal]] (tier 2 turned into a strategy), [[sequence-of-returns]] (tier 3), [[the-4-percent-rule]] (the rule told as a story) and [[the-trinity-study]] (the rule measured).
