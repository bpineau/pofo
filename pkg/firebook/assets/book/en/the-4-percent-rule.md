# The 4% rule in ten minutes
<!-- source: la-regle-des-4-pourcents @ 5b27abd60734 -->

It is the most famous rule in personal finance: withdraw 4% of your capital in the first year of retirement, raise that amount with inflation every year after, and your portfolio will last thirty years. It is so famous that people call it "the rule" and leave it at that.

This page lays out what the rule actually says (it is subtler than the barstool version), where its numbers came from, what it quietly assumes, and why the state of the art now treats it as an excellent place to start and a bad place to stop. By the end you will know how to use it for what it is worth: a ballpark, not a plan. And if you want to see why that number holds mathematically, and what would move it, the arithmetic is taken apart layer by layer in [[the-math-of-4-percent]].

::: cle What the rule says exactly
Year 1: withdraw 4% of the starting capital (EUR 40,000 on EUR 1,000,000). Every year after: withdraw the **same** amount adjusted for inflation (EUR 41,200 if inflation ran at 3%), whatever the portfolio does. In Bengen's data, across every 30-year US window since 1926, a portfolio holding 50 to 75% stocks was never emptied by that regime. "Safe" here means one thing: it never failed in the observed American past. Nothing more.
:::

## The mechanics, step by step

Meet Camille, who retires with EUR 1,000,000, invested 60% in global stocks and 40% in bonds.

1. **Year 1**: she withdraws EUR 40,000 (4% of 1,000,000).
2. **Year 2**: inflation ran at 2.5%. She withdraws 40,000 × 1.025 = EUR 41,000. It makes no difference whether her portfolio gained 15% or lost 20%: the withdrawal is the same.
3. **Year 3 and after**: same logic, last year's withdrawal adjusted for inflation.

Three properties follow straight from that machinery, and they explain everything else in the subject.

**Purchasing power is constant.** That is the rule's great virtue: your standard of living never depends on the mood of the markets. It is an annuity contract you sign with your own portfolio.

**The effective withdrawal rate, though, floats.** If Camille's portfolio drops to EUR 700,000 after a crash in year 2, her EUR 41,000 is suddenly 5.9% of capital. The rule does not care, and that is exactly where the risk lives.

**The "4%" is counted once and never again.** The rate applies to the starting capital, on the day you leave, and to nothing else. Hence a well-known paradox: two identical neighbors, one who left in 2021 with EUR 1M and draws EUR 40k, the other who left in 2022 after a crash with EUR 800k and draws EUR 32k, take out different amounts although their portfolios are now worth roughly the same. That is not a footnote. It shows the rule is a simplification of something deeper: the withdrawal rate depends on the valuations you start from ([[valuations-and-cape]]).

## Where the numbers came from

The rule has two founding acts, covered in [[the-trinity-study]].

**Bengen, 1994.** William Bengen replays every possible US retirement since 1926: leaving in 1926, in 1927, in 1928, and so on. For each vintage he computes how many years a 50/50 portfolio of stocks and government bonds would have lasted under a given indexed withdrawal. The result: the worst vintage in the record, 1966, right before fifteen years of flat markets and high inflation, carried an initial withdrawal of 4.15% for thirty years. Bengen would later name that floor SAFEMAX. The round "4%" comes from there. It is the rate of the **worst** case in American history, not an average: the median vintage carries close to 6%.

::: figure millesimes-soutenables
Each bar is one retirement date: the highest initial rate a US 50/50, real and rebalanced every year, would have carried for thirty years (Jorda-Schularick-Taylor panel, returns deflated by US CPI, vintages 1926 to 1991). The floor does land on 1966, but at 3.67% here: Bengen's 4.15% comes out of his own reconstruction, built on other series.
:::

**Trinity, 1998.** Three Trinity University professors (Cooley, Hubbard, Walz) turn the approach into a grid of probabilities: for each withdrawal rate, allocation and horizon, what share of the historical windows survived? The famous cell: 4%, a 50/50 portfolio, 30 years, 95 to 96% success depending on the update. That study is where the idea of a "success probability" comes from, and it still shapes every simulator ([[failure-probability]]).

::: encart The rule upside down: the 25x multiple
"Withdraw 4%" flips into "accumulate 25 times what you spend" (1/0.04 = 25). Same rule seen from the accumulation side, and that is its most useful job: turning a budget into a capital target. 3% means 33 times, 3.5% about 29 times. The full path from budget to target, taxes included, is in [[how-much-you-need]].
:::

## What the rule assumes without saying so

The marketing power of "4%" buried the long list of its assumptions. Each one deserves a look against your own situation.

| Implicit assumption | What your plan actually looks like |
|---|---|
| A 30-year horizon | Leaving at 40 means 45 to 55 years ([[horizon-and-life-expectancy]]) |
| **US** stocks and bonds, 1926 to 1995 | The best market of the century, over its best century ([[anarkulova-cederburg]]) |
| No fees | An index fund costs 0.1 to 0.4% a year, an insurance wrapper sometimes another 1% ([[building-it-with-us-etfs]]) |
| No tax on withdrawals | Tax on a withdrawal is a spending item: count it ([[us-taxes-in-the-withdrawal-phase]]) |
| Perfectly rigid spending, indexed to inflation | Real spending moves, and that is a margin you can use ([[spending-in-retirement]]) |
| No other income, ever | Pensions, side work and inheritances exist ([[pensions-and-other-income]]) |
| The retiree runs the rule mechanically for 30 years | Nobody watches a portfolio melt without reacting ([[the-psychology-of-spending]]) |
| Success = one euro left on the final day | Ending at 82 with EUR 3,000 counts as "success" for the simulator alone |

None of these assumptions kills the rule as a ballpark. Together, though, they explain why the number that comes out of a serious look at **your** situation can land well away from 4%, in either direction.

## What the state of the art says

Thirty years of research have sharpened the picture, and the results converge on four points.

**1. Over 30 years in the United States, the rule did hold.** Including through 2000-2009, the worst modern decade for stocks: the retiree of January 2000, who took two 50% crashes in their first ten years, was still solvent in 2024 under the 4% rule. The historical frame was not absurd.

**2. Over a long horizon, 4% is too high for a rigid rule.** That is the central result of the Early Retirement Now series ([[the-ern-series]]): over 50 to 60 years, the rate that would have survived every US vintage falls toward 3.25 to 3.5%. The failure probability of a rigid 4% over 50 years, measured on American history, runs around 10 to 20% depending on the allocation. Too much for a life plan.

**3. Outside the United States it is worse.** On the global sample, the broad sample of 16 developed countries since 1870 ([[anarkulova-cederburg]]), the "safe" rate of a domestic 60/40 sits well below 4%. Twentieth-century France, Japan, Germany and Italy handed their retirees sequences that American history simply does not contain. A globally diversified investor today sits somewhere between those two worlds.

**4. Starting valuations move the rate.** Every one of the worst vintages began in an expensive market, with a high CAPE. Modern rules therefore condition the initial rate on valuations ([[valuations-and-cape]], [[cape-based-rules]]). It is one of the best established improvements in the field.

::: science Where the research puts the "real" number today
For an early retirement, meaning a horizon of 45 years and up, with a globally diversified portfolio, no flexibility and no other income, the literature converges on 3.0 to 3.5% for a rigid rule (ERN, 3.25%; Morningstar 2025, over 30 years with forward-looking returns, 3.9%; Anarkulova-Cederburg, global sample, closer to 2.3 to 2.7% at the pessimistic end, a low bound that is disputed). Every protection you add lifts that number, sometimes past 4%: flexible spending ([[flexibility-in-practice]]), part-time income ([[going-back-to-work]]), a pension arriving later, guardrails ([[morningstar-guardrails]]). The safe rate is not a constant of nature. It is the output of a model, and a function of your assumptions and your margins.
:::

## So, keep it or toss it?

Keep it, in its place. The 4% rule is excellent in three roles and fails in a fourth.

**Excellent as a unit of measure.** "This annual expense needs 25 times its own size in capital" is the most useful mental reflex in the field. A recurring EUR 100 a month "costs" EUR 30,000 to 40,000 of capital, and that one calculation changes how you weigh a subscription, a car, a move.

**Excellent as a starting point for sizing.** Aim at 25 times your spending, then refine with a real model ([[using-the-fire-simulator]]) and your own situation. That is the right order. The mistake is not starting at 4%, it is stopping there.

**Excellent as a sanity check on a sales pitch.** When someone promises "8% a year, risk free", the rule tells you instantly that this is twice what the best market in history sustained under a rigid rule over thirty years. A reflex worth having.

**Bad as an actual withdrawal strategy.** Nobody should run a blind indexed withdrawal mechanically for forty years. Not because the rule is wrong, but because it ignores the information that keeps arriving: markets, health, real spending, valuations. The whole withdrawal-strategies part of this book ([[withdrawal-strategies-overview]]) is about what goes in its place, rules that listen to the portfolio ([[guyton-klinger]], [[morningstar-guardrails]], [[amortization-based-withdrawal]]). At equal capital they fund more spending, or protect the same standard of living far better.

::: attention The most common misreading
"4%, so I can withdraw 4% of the portfolio every year." No. That is the fixed percentage strategy ([[fixed-percentage]]), which can never ruin you but swings your standard of living hard. Bengen's rule indexes the initial **amount** to inflation, and lets its share of the current capital drift wherever the market takes it. The two strategies carry neither the same risks nor the same comfort, and confusing them derails every conversation on the subject.
:::

::: exemple The 4% rule against a real case
Back to Camille: EUR 1M, a global 60/40, retiring at 45, EUR 40,000 a year indexed. Here is what the three families of models ([[historical-vs-parametric]]) say over 45 years. A calibrated central parametric model typically gives a failure probability around 10 to 15%, a replay of the global sample gives more, the portfolio's own historical windows give less. At 3.4%, that is EUR 34,000 a year, the central-case failure probability drops below 5%. Keeping 4% but adding EUR 800 a month of pension from 67 gets it below 5% as well. That is the rule used well: one starting point, three levers tested, one informed decision.
:::

## The essentials

- The rule: 4% of the starting capital, then the same amount indexed to inflation; never once broken over 30 years in Bengen's US data.
- The "4" comes from the worst American vintage, 1966, and not from an average; it was already a cautious number... inside its own frame.
- Outside that frame (a long horizon, the whole world, fees, taxes, high valuations), the equivalent rigid rate is closer to 3 to 3.5%.
- Flipped into the "25x multiple", it is the best mental reflex in the field.
- As an actual strategy it is beaten by the adaptive rules: read [[withdrawal-strategies-overview]] before you carve 4% into your plan.

---

## Going further

- William Bengen, "Determining Withdrawal Rates Using Historical Data", *Journal of Financial Planning*, 1994.
- Cooley, Hubbard & Walz, "Retirement Savings: Choosing a Withdrawal Rate That Is Sustainable", *AAII Journal*, 1998 (the Trinity study, [[the-trinity-study]]).
- Early Retirement Now, SWR Series parts 1 and 26 ("Ten Things the Makers of the 4% Rule Don't Want You to Know"): the most complete modern critique ([[the-ern-series]]).
- Morningstar, *The State of Retirement Income* (annual): the recommended rate, recomputed every year with forward-looking returns ([[morningstar-guardrails]]).
- Bengen himself, in recent interviews. He now finds 4% too conservative for a more diversified portfolio, while reminding everyone that his rule only holds inside the US frame over 30 years. The two moves, his upward and the "whole world, long horizon" research downward, show how much the number depends on the frame.
