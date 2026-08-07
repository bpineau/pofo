# Sequence of returns risk: the retiree's real enemy
<!-- source: sequence-des-rendements @ ac06601809cf -->

Two retirees start with the same million, the same portfolio, the same EUR 40,000 inflation-indexed withdrawal. Over thirty years their portfolios earn exactly the same average return.

One ends up with two million. The other is broke at 78. The only difference: the **order** in which the same returns arrived.

That is sequence of returns risk, and it is the central idea of this whole book. It explains why the safe withdrawal rate sits so far below average returns. It explains why the first ten years of retirement dominate everything. And it explains why most of the clever protections in the field are really anti-sequence weapons: glidepaths, cash buffers, flexibility, part-time income. By the end of this page you will know how to recognize the risk, how to measure it on your own plan, and what the full map of defenses looks like.

::: cle The fundamental asymmetry
With no withdrawals, the order of returns is irrelevant: +30% then −20%, or −20% then +30%, leave you with the same capital (adding logarithms commutes). With withdrawals, the order decides everything: every euro you take out during a trough is a euro sold at the worst price, permanently removed from the rebound. A retiree converts temporary losses into permanent ones, euro for euro with what they withdraw. That is why the very same portfolio is far riskier in the withdrawal phase than in the accumulation phase.
:::

::: figure sequence-risk
Two retirees, the same average return over thirty years and identical withdrawals: the one who takes the crash early runs dry, the one who takes it late survives. Only the order of the returns differs.
:::

## The mechanism, in an example you will not forget

Take three years of real returns: +20%, +10%, −25%. Geometric mean: near enough zero (1.20 × 1.10 × 0.75 = 0.99, or −0.3% a year). Starting capital EUR 1,000,000, EUR 40,000 withdrawn at the start of each year.

**Favorable sequence (the crash last)**: +20%, +10%, −25%.

| Year | Before draw | After draw | Return | Year end |
|---|---|---|---|---|
| 1 | 1,000,000 | 960,000 | +20% | 1,152,000 |
| 2 | 1,152,000 | 1,112,000 | +10% | 1,223,200 |
| 3 | 1,223,200 | 1,183,200 | −25% | 887,400 |

**Unfavorable sequence (the crash first)**: −25%, +10%, +20%.

| Year | Before draw | After draw | Return | Year end |
|---|---|---|---|---|
| 1 | 1,000,000 | 960,000 | −25% | 720,000 |
| 2 | 720,000 | 680,000 | +10% | 748,000 |
| 3 | 748,000 | 708,000 | +20% | 849,600 |

Same returns, same average, same withdrawals: EUR 887,400 against EUR 849,600. Almost EUR 38,000 apart, a full year of spending, in three years flat. Now stretch that mechanism across an entire opening decade of bear market. Add withdrawals that eat a growing share of a shrinking capital every year. What you get is the distance between the 1966 and the 1982 vintages of US history. The first runs out before its horizon ends, the second finishes several times richer than it started. And the first was not short of returns: its portfolio compounded at 4.2% a year in real terms over thirty years, more than the 4% it was drawing. Its ruin turned on the order alone, a first decade at −1.2% a year against +11.4% for the 1982 vintage ([[the-trinity-study]]).

::: figure millesimes-1966-1982
The same plan (EUR 1M, EUR 40k a year indexed to inflation) run on the real US 60/40, once from 1966 and once from 1982, both curves rebased to their starting day. Reconstructed for this book from the S&P 500, 5-year Treasuries and CPI-U.
:::

Here is the intuition worth keeping. In accumulation an early crash is a gift: you buy cheap for years. In withdrawal it is a hemorrhage: you sell cheap for years. The same event flips sign with the direction of the cash flows. Which is why your glorious record as a saver, "I sat through 2008 and 2020 without blinking", proves nothing about your exposure as a retiree. You were simply on the right side of the flows.

## The fragile window: the first ten years

Sequence risk is not spread evenly through time. It piles up massively at the start of retirement, for a mechanical reason. That is when the capital to protect is largest, when the most withdrawals still lie ahead, and when the path has the most years left to diverge. A crash in year 25 of a 30-year retirement is nearly painless: most of the withdrawals are behind you, and the capital needed to finish is small. The same crash in year 2 governs everything that follows.

The research (ERN part 15 in particular, [[the-ern-series]]) puts numbers on that intuition: the correlation between a plan's final success and the returns it actually earned is overwhelming for the first 5 to 10 years and weak afterward. In practice, **three quarters** of the fate of a 40-year retirement is settled in its first decade. The analysis that shows it can be redone on any set of simulated paths. Sort the scenarios by the real return of their first ten years, then look at the distribution of final outcomes within each bucket. That is failure decomposed by decisive decade, and it tells you at once whether your plan depends on the opening or on the average.

Three practical consequences follow from that concentration in time.

**1. Protection can be temporary.** Since the danger is concentrated, expensive defenses (cautious allocations, buffers, side income) do not have to last forever. Concentrating them on the fragile window captures most of the benefit for a fraction of the cost. That is the foundation of the rising equity glidepaths of Pfau and Kitces ([[glidepaths]]): start cautious, then raise equity exposure as the window closes.

**2. Your start date is a risk parameter.** Retiring at the top of an expensive market raises the odds that the fragile window contains the crash ([[valuations-and-cape]]). It is also what makes one more year partly rational in a euphoric market, and expensive in a market that has already been purged ([[one-more-year]]).

**3. The early years of a plan are monitored differently.** A plan that crosses its first decade without major damage has, statistically, won; vigilance can relax. Useful alert thresholds are therefore dated, not uniform ([[when-to-worry]]).

::: science Measuring sequence risk on your own plan
Two readings make your exposure visible, and any serious tool supports both. The first is the decomposition by decisive decade described above. If the scenarios whose first decade falls in the worst quartile nearly all end in ruin, your plan is a bet on the sequence; if they survive battered, it is robust. The second is a comparison of two twin models. A central model draws years independently. A sequence stress keeps exactly the same long-run average but makes the bad years arrive in clusters, the way persistent bear markets do (Markov chains). Everything else is identical: withdrawal, allocation, horizon. The gap in failure rate between the two is, precisely, what the sequence costs in your plan ([[making-monte-carlo-relevant]]). A small gap signals a plan that is naturally well defended, by a low withdrawal, by flexibility, or by income. A gap of several points signals that the defenses below deserve your attention.
:::

## Why averages lie to you

Sequence risk explains a paradox that throws every beginner: how can a portfolio "averaging 5% real" support only a 3.5% withdrawal? Should you not be able to draw the average forever?

No, for two reasons that stack. The first is volatility drag. The compound growth of a volatile portfolio falls below its arithmetic mean by roughly half its variance ([[arithmetic-vs-geometric-returns]]). A "5% average" carrying 15% volatility actually compounds at about 3.9%. The second reason is the sequence. Even the geometric return is only drawable if the returns arrive smoothly. Their unevenness, combined with fixed withdrawals, eats a further premium. The safe withdrawal rate is therefore structurally below the expected geometric return, which is itself below the arithmetic mean the brochures advertise. Keep the ranking in mind: **arithmetic mean > geometric mean > sustainable withdrawal rate**. Each step typically costs 0.5 to 1.5 points.

::: attention The simulator that is too smooth
Any model that draws years independently (naive Monte Carlo, the central Student-t model included) slightly understates sequence risk: real markets come in clusters, trends and lost decades, not lottery draws ([[simulator-traps]]). Which is why a central model is never read alone. It is read next to a sequence stress and a resampling of the global century, both of which contain the chains of events that independent draws erase ([[anarkulova-cederburg]], [[historical-vs-parametric]]). The rule is blunt: if your plan is only acceptable under the central model, it is not acceptable ([[making-monte-carlo-relevant]]).
:::

## The map of defenses

Every large family of protection in this field is, at bottom, an anti-sequence weapon. Here they are in one table, with their mechanism and the chapter that covers them.

| Defense | Anti-sequence mechanism | Where it is covered |
|---|---|---|
| Lower initial withdrawal | Draws less during any early trough | [[how-much-you-need]] |
| Flexible spending, guardrails | Draws less exactly when prices are low | [[flexibility-in-practice]], [[morningstar-guardrails]], [[guyton-klinger]] |
| Cash buffer, buckets | Sells cash rather than stocks in the trough | [[cash-buffer]], [[the-bucket-strategy]], [[refilling-the-buffer]] |
| Glidepath (bond tent) | Cuts equity exposure during the fragile window only | [[glidepaths]] |
| Part-time income early (barista) | Cuts net withdrawals during the fragile window | [[going-back-to-work]], [[pensions-and-other-income]] |
| Uncorrelated defensive assets | Softens the depth of the trough itself | [[defensive-assets]], [[gold-in-retirement]], [[managed-futures]] |
| Annuity or guaranteed floor | Takes part of your spending out of the sequence game | [[annuities-and-safety-first]] |
| Start date conditioned on valuations | Avoids opening the fragile window at the top | [[valuations-and-cape]] |

No defense is free. A low withdrawal costs years of work. Cash and annuities cost return. Flexibility costs comfort, part-time income costs freedom. Designing a plan ([[building-your-plan]], [[choosing-your-strategy]]) means buying that anti-sequence protection at the best price for your situation. A household with an already high spending floor will buy cash buffer and deferred income. A flexible household will mostly buy an adaptive withdrawal rule, the best value for money for almost everyone.

::: exemple The same retirement, with and without defenses
Base plan: EUR 1M, 60/40, a rigid EUR 40,000 a year, 45-year horizon; failure under sequence stress about 18%. Variant A, guardrails (cut to EUR 36,000 as soon as the current withdrawal rate passes 5%) gives about 7%. Variant B, three years of spending held as a cash buffer, drawn down in troughs and refilled at peaks, gives about 12%. Variant C, EUR 12,000 a year of side income for the first 8 years, gives about 8%. Variants A and C together give about 3%. The exact figures depend on the model, so test your own; the ranking, though, is robust: flexibility pays first, early income next, the buffer as a supplement.
:::

## Three misunderstandings worth clearing up

**"Sequence risk is crash risk."** No. It is the risk of a crash or an erosion in the **wrong place**. The worst US vintage is not 1929 but 1966: no spectacular crash, fifteen years of slow real suffocation by inflation ([[the-trinity-study]], [[inflation-and-withdrawal-rates]]). Purely anti-crash defenses (options, stop losses) miss that failure mode entirely; anti-sequence defenses (flexibility, income, real assets) cover it.

**"I am a passive long-term investor, so I am immune."** The immunity of passive investing holds in accumulation, where order is irrelevant. From your first withdrawal on, you are in the sequence game, however passive the portfolio. It is the classic blind spot of the seasoned saver who walks into retirement with accumulation reflexes ([[ten-plan-wrecking-mistakes]]).

**"The sequence is unpredictable, so nothing can be done."** The sequence is unpredictable; your **exposure** to it can be measured and reduced, which is the entire point of the table above. You cannot forecast the rain, so you build a roof.

## The essentials

- With withdrawals, the order of returns matters as much as their average: early losses become permanent, euro for euro with what you withdraw during the trough.
- The danger concentrates in the first 5 to 10 years, the fragile window; a plan that crosses it well has statistically won.
- A ranking to memorize: arithmetic mean > geometric mean > sustainable rate. The brochures sell you the first, you live on the third.
- Every large protection in the field is an anti-sequence weapon, each with its price; bounded flexibility offers the best protection per unit of cost for most plans.
- Measure your exposure two ways: failure decomposed by the return of the first decade, and the gap between a central model and its sequence stress. Then read [[failure-probability]] to interpret the numbers.

---

## Going further

- Early Retirement Now, SWR Series parts 14 and 15 ("Sequence of Return Risk"): the demonstration that the sequence explains most of the outcome; part 53 on hedges and the Retiree-Saver Investment Pact ([[the-ern-series]]).
- Pfau and Kitces, "Reducing Retirement Risk with a Rising Equity Glide Path", *Journal of Financial Planning*, 2014: the defense by allocation ([[glidepaths]]).
- Moshe Milevsky, "Retirement Ruin and the Sequencing of Returns": the actuarial formalization.
- In this book: [[the-math-of-4-percent]] (the sequence penalty in figures, about 1.8 points in the cascade of the 4% rule) and [[why-diversification-works]] (diversification as a remedy for the same risk).
- Also in this book: [[market-regimes]] (the persistence of bear markets, the structure a sequence stress imitates) and [[under-the-hood]] (how this machinery is actually implemented).
