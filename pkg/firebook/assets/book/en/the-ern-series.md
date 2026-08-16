# ERN's Safe Withdrawal Rate series: a reader's guide
<!-- source: serie-ern @ 698768d091ff -->

If this book could send you to one outside source on everything it covers, it would be this one: the Safe Withdrawal Rate Series on the blog *Early Retirement Now* (ERN). Karsten Jeske writes it, a PhD economist who worked at the Atlanta Fed and in quantitative asset management, and who retired early in 2018. More than sixty parts since 2016, every one of them built on reproducible simulations over monthly US data going back to 1871. Article after article, the series took apart the slogans of first-generation FIRE and laid down most of what is now said seriously about long retirements.

This page is a reader's guide: the major results, where to find them, and what to know about the author's leanings so you can read him well. Many articles in this book argue with the series; this one hands you the map.

::: cle Why the series matters
Before ERN, the FIRE debate was slogan against slogan ("4%, settled" against "anything can happen"). The series imposed a method: exhaustive simulations over 150 years of monthly data, every vintage, horizons of 50 to 60 years. And a standing honesty about what works, what does not, and what each fix costs. You can disagree with its conclusions, and some do ([[anarkulova-cederburg]]). You cannot go back to the slogans.
:::

## The author and the method

ERN's frame holds across the whole series. Monthly US data since 1871 (S&P composite stocks, 10-year government bonds, from Shiller). Simulated retirements starting in every possible month, over horizons of 30 to 60 years. And a close look at a criterion that is not just ruin, but ending wealth and the path too: when the failure hits, and what the retiree lived through on the way there. The tools are public. The Google Sheets toolbox of Part 28 lets you redo every calculation at home.

Two leanings are worth knowing before you start. The first sits in the data, which is American. The series inherits the optimism of that sample ([[the-trinity-study]], [[anarkulova-cederburg]]). ERN knows it, says so, and holds that 150 years of monthly data from a big market, 1929 and 1966 included, already discipline the conclusions hard. The second is a posture. ERN writes for the early retiree who refuses to depend on luck. Hence a robustness bar, often "survive the worst vintage", tougher than the 90 to 95% success rates practitioners work with ([[morningstar-guardrails]], [[failure-probability]]).

## The major results, part by part

**The bedrock: 4% was not built for you (Parts 1 to 3, 26).** This is the founding result. The 4% rule holds over 30 years. A horizon of 50 to 60 years calls for 3.25 to 3.5% if the withdrawal is rigid, meaning no flexibility at all on the amount, and the safe rate depends heavily on starting valuations (Part 3). Above a CAPE of 20, the historical SAFEMAX drops at every horizon. Part 26 ("Ten Things the Makers of the 4% Rule Don't Want You to Know") is the best popular summary ([[the-4-percent-rule]], [[valuations-and-cape]]).

**Sequence explains everything (Parts 14 and 15).** The series shows, with the numbers in hand, that the average return over 30 years matters less than the return of the first 5 to 10 years. That is the heart of [[sequence-of-returns]]. Part 53 adds the corollary: the saver and the retiree face opposite exposures to sequence.

**Flexibility is oversold (Parts 9 to 11, 23 to 25, 58).** This is the most uncomfortable block of the series, and the most useful. The Guyton-Klinger rules (Parts 9 and 10) post flattering "success" rates. In the bad vintages, those rates are paid for with decades of spending cut by 30 to 45%. Ruin is simply swapped for a long stretch of poverty that no table of success rates shows ([[guyton-klinger]]). Parts 23 to 25 and 58 generalize the finding: any realistic flexibility, bounded and livable, buys a few tenths of a point of withdrawal rate, not the advertised magic ([[flexibility-in-practice]]). That result is what pushed the whole field to show the standard of living delivered, and not just the survival of the portfolio ([[using-the-fire-simulator]]).

**The CAPE as a rule, not as a fear (Parts 18 and 54).** This is the series' constructive proposal: withdrawal rules built on the CAPE, where the rate, initial and current alike, adjusts to valuations, formalized in Part 54. It is the direct ancestor of the CAPE anchor in the FIRE simulator ([[cape-based-rules]]).

**Buckets and cash do not do what people think (Parts 12, 48, 55).** This one cuts against the bucket dogma. A cash buffer drawn down and then mechanically refilled barely lowers the risk of ruin, because cash costs in return what it saves in sequence. And the popular bucket strategies are often market timing in disguise, with no clear rule ([[cash-buffer]], [[the-bucket-strategy]], [[refilling-the-buffer]]). The nuance matters. Simulation lands in the same ballpark, with the buffer trade-off usually flat ([[cash-buffer]]). None of which takes anything away from what the buffer is worth psychologically, or for governance ([[the-psychology-of-spending]]).

**Glidepaths work (Parts 19, 20, 43).** This is the good news of the series. Starting at 60% stocks and climbing back toward 100% over ten years or so buys 0.1 to 0.3 points of withdrawal rate in the worst vintages, at a modest cost in the good ones. The protection lands squarely on the fragile window ([[glidepaths]], [[the-three-phases]]).

**Inflation, from both ends (Parts 41 and 51).** Part 41 asks whether low inflation on the day you leave allows a higher rate. Part 51 answers that the level on that day predicts almost nothing, because it is the inflation of the years that **follow** that shaves the sustainable rate down, by roughly 0.2 points for every point of average inflation over the first decade ([[inflation-and-withdrawal-rates]]).

**Annuities, bonds, gold, real estate, leverage (Parts 29 to 31, 34 to 36, 40, 49, 52, 56, 59, 62).** These are the asset reviews. The yield illusion of dividend portfolios, since living off your dividends is no safer (Parts 29 to 31 and 40, [[ten-plan-wrecking-mistakes]]). Gold as a partial sequence hedge (Part 34, [[gold-in-retirement]]). Real estate (Part 36, [[real-estate-in-retirement]]). Annuities and Social Security (Parts 56 and 59, [[annuities-and-safety-first]]). Leverage, sophisticated and in small doses (Parts 49 and 52, [[leverage-and-margin]]). And small-cap value in decumulation, a negative verdict (Part 62, 2025, [[factors-in-retirement]]).

**The side topics that change a life (Parts 21, 22, 27, 37, 42, 47, 60).** The mortgage in retirement, where keeping the loan is a levered bet on sequence (Part 21). "One more year", with numbers attached (Part 42, [[one-more-year]]). When to worry along the way (Part 47, [[when-to-worry]]). Getting through a bear market (Part 37, [[bear-markets-in-retirement]]). And the critique of *Die With Zero* (Part 60, [[spending-in-retirement]]).

::: astuce Where to start, depending on your question
The overview: Parts 1 and 26. "How much can I withdraw?": Parts 2 and 3, then 54 for the CAPE. "Will flexibility save me?": Parts 23 to 25, then 58. "My allocation around the exit": Parts 19 and 20. "What about buckets?": Parts 12 and 48. Each part takes 20 to 40 minutes. The whole series is a book in its own right, longer than this one, and the two fit together. ERN pushes the American simulations further; this book covers the world sample, the tax and account frame, and the plumbing of the simulator.
:::

## Reading ERN with the right filters

Three filters, and how much of each you need depends on where you live.

**The geographic filter.** If you are American, this one barely applies. The data, the tax chapters, Social Security, the 401(k) and IRA material all land in your own system, and you can read them straight ([[us-taxes-in-the-withdrawal-phase]]). One caveat survives even then: the numbers inherit the sample. A century and a half of US returns is the record of the luckiest large market there has been, and the safe rates that come out of it stay on the optimistic side ([[anarkulova-cederburg]]). Read from anywhere else, the filter does real work. The mechanisms carry over untouched, whether it is sequence, the CAPE, glidepaths or flexibility. The numbers need a check against your own currency and home market, and the tax and pension chapters have to be swapped for yours.

**The posture filter.** The bar "survive the worst vintage in the record" is a deliberate stance of caution, tougher than the 90 to 95% success practitioners settle for. If you have solid safety nets (a pension, employability, real flexibility), ERN's recommendations are a prudent bound for you, not a survival minimum ([[failure-probability]], [[how-much-you-need]]).

**The perfectionism filter.** The series proves that you can refine forever, down to the tenth of a point of SWR and the allocation within 5 points. Keep one reminder in mind as you read. The first three levers of a real plan are still audited spending, the pension counted, and a written rule ([[ten-plan-wrecking-mistakes]]). Refinement comes after those.

::: terrain What the community did with it
The series changed the FIRE conversation for good. "What's your SWR?" replaced "Have you hit 25x?". Serious simulators now show the standard of living delivered, not just the success rate. And "ERN Part N" has become common shorthand on the forums, the sign of a community that raised its own bar. One piece of advice to close on: do not read the sixty parts back to back. Take the ones that answer the question in front of you, then come back later. It is an encyclopedia, not a novel.
:::

## The essentials

- ERN's SWR series ([earlyretirementnow.com](https://earlyretirementnow.com), more than 60 parts since 2016) is the modern reference on long-horizon withdrawal: methodical, reproducible, free.
- Its structural results: 3.25 to 3.5% rigid over 50 to 60 years, the dominant role of sequence and valuations, flexibility oversold (it moves the pain rather than removing it), buckets demystified, glidepaths validated.
- Its leanings: American data, so numbers on the optimistic side, and a worst-case bar, so recommendations on the cautious side. The two partly cancel out. Know it as you read.
- An American reader takes the whole series straight, tax chapters included. Read from anywhere else, keep the mechanisms, check the numbers against your own market and currency, and replace the tax and pension chapters.
- Many of the FIRE simulator's ideas (the CAPE anchor, the standard of living delivered, sequence stress, the buffer trade-off) are in direct dialogue with this series ([[using-the-fire-simulator]], [[under-the-hood]]).

---

## Going further

- The entry point: [earlyretirementnow.com/safe-withdrawal-rate-series/](https://earlyretirementnow.com/safe-withdrawal-rate-series/) (the full table of contents, kept up to date).
- The toolbox (Part 28): the public simulation worksheet, to redo the calculations yourself.
- Karsten Jeske on podcasts (ChooseFI, Rational Reminder, Bogleheads): the spoken version of the same results, often easier to get into.
- The counterpoints in this book: [[anarkulova-cederburg]] (the sample beyond the United States) and [[morningstar-guardrails]] (the practitioner's reading, less worst-case).
