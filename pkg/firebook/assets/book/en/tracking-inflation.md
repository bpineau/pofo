# Tracking inflation: the indexes, and yours
<!-- source: suivre-inflation @ c4d2465faff8 -->

The whole plan lives in constant money. Withdrawals get indexed, simulations run in real terms, the floor is a level of purchasing power ([[inflation-and-withdrawal-rates]]). But constant against **what**?

"Inflation" is not a raw fact. It is a constructed measure, built on a basket that is not yours. The gap between the official index and **your** personal inflation looks tiny in any single year. Compounded over thirty years it can be worth as much as a point of withdrawal rate. This chapter is the measurement manual. It starts with the indexes a US reader actually meets, CPI-U and its siblings, how they are built and what their biggest blind spot is, owner housing. Then the honest argument about how reliable they are. Then the central question, **your** inflation, why a retiree's runs above the index and how to estimate yours in an evening. Then what the market expects, the breakevens, the only inflation forecast with money riding on it. And last the operational translation: what to index withdrawals to, and how to set the spending drift in a simulator.

::: cle Three inflations to keep apart
There are three. **Official** inflation is the CPI, the inflation of the statistics, of Social Security and of TIPS. The inflation of **your basket** weights health care, services, energy and leisure your way, typically 0.2 to 0.5 points above the official number for a retiree. The inflation of your **plan** is the one you choose to apply to your withdrawals: official plus an explicit health drift, which is the honest setting. Confuse the three and you get plans that hold up in constant money, made of the wrong money.
:::

## The indexes: what they are, what is in them

**CPI-U, the headline index (BLS).** It prices the spending of urban consumers, about 93% of the US population, from tens of thousands of quotes gathered every month across roughly 75 urban areas, in a few hundred item categories weighted by their share of an average household budget. This is the number in the headlines. It is also the one TIPS principal follows, in its not-seasonally-adjusted form ([[inflation-linked-bonds]]), and the one most indexation clauses point at.

**CPI-W, the Social Security index.** Same price collection, narrower population: urban wage earners and clerical workers, a bit under a third of Americans. It carries one enormous job. The annual COLA is the change in the third-quarter average of CPI-W against the same quarter a year earlier, which is where the 2.8% raise paid in 2026 came from. Its basket leans a little more on transportation and a little less on health care than CPI-U does, which is exactly backwards for a retiree.

**PCE, the Fed's index (BEA).** The Fed's 2% target is a PCE target, not a CPI target. PCE builds its weights from what businesses report selling rather than what households report buying, re-weights faster when people substitute, and counts the health care that insurers and Medicare pay on your behalf. It normally runs a few tenths a year below CPI. Read it to understand the Fed. Do not index anything to it: nothing you own is linked to it.

**CPI-E, the research index for the elderly.** BLS also computes an experimental index over households headed by someone 62 or older. From 1982 to 2014 it ran about 0.2 points a year above CPI-W. Over 2002 to 2021 the gap almost vanished, as medical inflation slowed. Nothing is indexed to it, and bills to run the COLA on it come back to Congress every few years. Treat it as evidence about the drift, not as an index you can use.

**What the basket handles badly: owner housing.** This is the most important technical point. Shelter is the biggest line in the CPI, around a third of the basket, and most of that is owners' equivalent rent, roughly a quarter of the whole index on its own. OER is an imputation: what a homeowner would pay to rent the house they already live in. House prices, mortgage interest and property taxes left the index in 1983, when BLS moved homeowners to rental equivalence on the grounds that a house is partly an investment. So the index tracks the flow of shelter services and ignores the price of getting into the asset. That cuts both ways for you. A retiree who owns the roof outright ([[real-estate-in-retirement]]) has a personal inflation structurally gentler than the index, because his largest line is switched off. A saver still saving for a house watches the only inflation that matters to him run away, and the index says nothing about it. The same logic holds for everything the index averages and you do not consume. And the imputation is slow: OER follows market rents with a lag of a year or more, so the index is still catching up to a rent wave long after it has broken.

**Quality adjustment (hedonics).** When the 2026 phone does more than the 2020 one at the same price, the index records a price **cut**. The method is defensible, since what is being measured is the price of the service delivered. It is argued about in practice, because you cannot buy "less quality" and your outlay does not fall. The honest debate about reliability fits in two sentences. The US indexes are serious, audited, and no manipulation has ever been demonstrated. And they measure an average, quality-adjusted basket that can drift for a long time from one household's spending. Both sentences are true, which is what the next section is about.

**Where to read them.** BLS publishes CPI monthly on bls.gov, with the detail by item category and the relative-importance tables that say what each line weighs. FRED carries every series, CPI-U, CPI-W, PCE, the category indexes and the breakevens below, in a form you can chart in a minute. Weights and methods do move: read the current release rather than a number you remember.

## Your inflation: why a retiree's drifts above

The spread of personal inflation rates around the index is documented. BLS publishes the index by item category and by region, and researchers rebuild household-level indexes from the expenditure survey; the gaps between household types run to a few tenths of a point a year. And the **retiree** profile stacks up the unfavorable weights:

- **Health care and long-term care.** This line grows as a share of the budget with age, and its bill runs structurally above the CPI. What you feel is not the published medical care index, which is measured largely at what insurers pay and moves slowly. It is your out-of-pocket share, starting with premiums: the standard Medicare Part B premium has compounded at roughly 4.2% a year over the past two decades, from $88.50 a month in 2006 to $202.90 for 2026, and that last step was 9.7% in one year. Then come home help and facilities ([[spending-in-retirement]]). Premiums are reset every year, so read the current figure before planning with it.
- **Services in general** (help at home, maintenance, restaurants, insurance). They are labor-intensive and local, so their prices follow wages, historically about a point a year above goods. And a retiree's budget holds more services than average.
- **What gets cheaper no longer concerns him as much.** Falling prices for technology and manufactured goods pull the index down, and they weigh little in a retiree's basket.

The order of magnitude that comes out matches the senior-index evidence: the CPI-E ran 0.2 to 0.3 points a year above the general index over long stretches. A retired household's inflation therefore beats the index by something like 0.2 to 0.5 points a year, more at advanced ages. Compounded over thirty years, 0.3 points is worth about 9 to 10% of purchasing power, a year and a half of spending. This is not a refinement, it is a line item of the plan.

::: figure ecart-compose
What indexing on the official index alone fails to pay, piled up year after year and counted in years of today's spending (compounding arithmetic, over 30 years). At the end of the road the price level has diverged by only 9 to 16%, but one to two whole years of spending were never funded.
:::

**Estimating yours in an evening.** Pull the 12 to 24 months of statements you have already sorted ([[how-much-you-need]]). Weight your real categories. Cross them with the BLS indexes by item category, services, food, energy, all published monthly and all charted on FRED, and use your own premium notices for health care. The weighted average is your backward-looking inflation. You are not after the decimal. The point is to know whether you are an "index" household, an "index plus 0.3" or an "index plus 0.6", and to hand that number to the plan.

::: astuce Breakevens: the only forecast with money riding on it
Where do you read what "the market" expects? In the inflation **breakevens** ([[inflation-linked-bonds]]), the gap between the nominal yield of a Treasury and the real yield of a TIPS of the same maturity. The 10-year breakeven, about 2.3% in mid-2026 and published daily by FRED, is the average expectation of investors with billions riding on it. It is imperfect, since it carries risk and liquidity premiums. It is still infinitely more disciplined than surveys and op-eds. Its use in a plan is **not** market timing, it is calibration. If your simulation implicitly assumes 2% and breakevens climb durably toward 2.7%, that is a new fact and it belongs in the annual review. And when someone sells you a product "because hyperinflation is coming", the breakevens are the polite answer ([[hyperinflation-and-extremes]]).
:::

## Tracking it in practice, and setting the plan

**The minimal dashboard** takes ten minutes at the annual review ([[the-annual-review]]). First number, official inflation over the last twelve months, published monthly by BLS: that is the number that indexed your withdrawals. Second number, your own backward-looking inflation, the evening's calculation redone at a glance. Third number, the 10-year breakeven, the calibration of expectations. Everything else is noise for a 40-year plan: this month's print, the core-versus-headline argument, the headlines themselves. You fly inflation by the year, not by the monthly release.

**What should withdrawals be indexed to?** The canonical fixed rule indexes on the CPI ([[fixed-inflation-adjusted-withdrawal]]). That is the right **default**: observable, hard to argue with, consistent with Social Security and with TIPS. Complete it with an explicit drift. The clean setting is not "I index on CPI plus 0.4" (unverifiable), it is "I index on the CPI and I budget a real drift in spending". That is literally what a simulator does. A **real spending drift** control (in percent a year) adds a real slope to spending (0.3 to 0.5% a year is the recommended planning value for the health drift), and the retirement smile shapes the profile by age ([[spending-in-retirement]], [[using-the-fire-simulator]]). A plan simulated over 45 years with **no** drift assumes your basket will track the national average basket for half a century. It is the most common piece of disguised optimism on the subject.

**And remember what a simulator already does.** It works in **real** terms from end to end, on deflated series, with withdrawals held at constant purchasing power ([[under-the-hood]]). **Average** inflation is in the machine by construction. What the controls add is **your** gap to the average, the drift. What the regime models test is the **risk** of an episode ([[inflation-and-withdrawal-rates]] for the full mechanics).

::: exemple Denise and Paul's personal inflation
Denise (63) and Paul (66) own their home outright and spend $46,000 a year. Here is their evening of arithmetic. Six lines, each with its weight in the budget and its own rate: services, help and insurance 22% at 3% a year, health care and premiums 14% at 4.5%, leisure and travel 24% at 2%, food 18% at 2%, energy and transport 12% at 2.5%, everything else 10% at 2%. Their weighted personal inflation comes out around 2.6% while the official index runs 2.1%. The gap is 0.5 points, most of it health care, right in line with their age. On the plan side they index withdrawals on the official index (that is what the contracts say), set "Real spending drift /yr" to 0.4%, and turn the smile on (the health drift net of the travel slowdown after 80). The simulation moves: central failure goes from 3.9 to 4.8%. That is the real price of their basket, which no national index would ever have billed them for. Better to learn it at 63 than at 83.
:::

::: figure panier-contributions
Denise and Paul's evening of arithmetic, line by line. Each contribution is the weight of the line times the rise in its prices, and the six stack up to 2.63%. Their budget happens to be in dollars; the gesture is the same in any currency, on any statement you have already sorted.
:::

## The essentials

- Three inflations: the official one (CPI, the index of Social Security, of TIPS and of the deflated series a simulator runs on), yours (your basket, typically index plus 0.2 to 0.5 for a retiree, health care and services leading), and the plan's (official plus an explicit drift).
- The US indexes are serious, but they average a basket that is not yours. The main blind spot is owner housing: shelter enters as imputed rent, and house prices, mortgage interest and property taxes have been out of the index since 1983. A paid-off roof softens **your** inflation, and the index does not know it.
- The compounded gap counts. 0.3 points a year over 30 years is about 10% of purchasing power, a year and a half of spending. Estimate your basket in an evening: your statements crossed with the BLS category indexes, charted on FRED.
- The annual dashboard is three numbers: official inflation over 12 months, your backward-looking inflation, the 10-year breakeven (the only expectation with money riding on it, a calibration, never a timing signal).
- The clean setting fits on one line. Index on the CPI and budget the drift (a slope of 0.3 to 0.5% a year, plus the smile profile). A 45-year plan with no drift assumes an eternally average basket, and that is the best-disguised optimism on the subject.

---

## Going further

- [bls.gov](https://www.bls.gov): the monthly CPI, the detail by item category, the relative-importance tables and the CPI-E research index, the tools of the evening's calculation.
- [fred.stlouisfed.org](https://fred.stlouisfed.org): CPI-U, CPI-W, PCE, the category indexes and the breakevens, all chartable in a minute.
- [ssa.gov](https://www.ssa.gov): how the COLA is computed, quarter by quarter, and the figure for the current year.
- The Boskin Report (1996) and what followed it: the quality and substitution debate, documented.
- In this book: [[inflation-and-withdrawal-rates]] (the exact effect on the plan), [[spending-in-retirement]] (the drift and the smile), [[inflation-linked-bonds]] (breakevens in practice).
