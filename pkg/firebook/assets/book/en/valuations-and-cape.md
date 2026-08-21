# Valuations, the CAPE, and what they say about your withdrawal rate
<!-- source: valorisations-et-cape @ 0cc5844ab5e1 -->

After sequence risk, here is the most important empirical fact in the field. It fits in one sentence. The worst retirement start dates in history, 1929, 1966, 2000, Japan in 1990, all share one feature. It is not bad luck. It is that the market was historically expensive on the day they left. The valuation level on the day of your first withdrawal is the best known predictor of the withdrawal rate that vintage can carry.

The standard gauge of that expensiveness is Shiller's CAPE. This page covers it end to end. What it is and how it is computed. Why it predicts returns, and how well. The measured link between the CAPE and the withdrawal rate. The serious criticisms, because there are some. And above all the four concrete ways to use it in a plan, starting with anchoring your expected return to today's valuation. By the end you will read today's CAPE like a retiree, not like a trader. The question is not "should I sell?" but "what can I promise my plan?"

::: cle The idea in one sentence
The price you pay for a stream of earnings sets the return that stream can pay you. Buying expensive means accepting lower future returns. A retiree's fate is decided in the first decade ([[sequence-of-returns]]). So leaving when the market is expensive means the decisive decade starts with a compressed expected return and a higher risk of a correction: the double penalty. The CAPE does not predict crashes. It measures the size of the promise the market can keep.
:::

::: figure cape-swr
The withdrawal rate that would have survived every vintage, plotted against the CAPE at the start (ballpark numbers, US data). Leaving expensive compresses the sustainable rate; leaving cheap widens it.
:::

## The CAPE: definition, computation, origin

CAPE stands for *Cyclically Adjusted Price-to-Earnings ratio*. Robert Shiller and John Campbell proposed the recipe in 1988, borrowing from Graham and Dodd, who in 1934 already advised averaging earnings "over five to ten years, preferably ten":

1. Take the price of the index (the S&P 500, historically).
2. Divide it not by this year's earnings (the plain P/E) but by the average real earnings of the last ten years, each year first restated in constant currency (adjusted for inflation).

Why ten years? Because a single year's earnings make a terrible denominator. In a recession they collapse, which inflates the P/E at the worst possible moment and makes the market look "expensive" at the bottom of 2009, exactly backwards. At the top of a cycle they are flattered by margins that cannot last. Ten years cover a full business cycle. The denominator then estimates what companies normally earn, and the ratio really measures what you are paying for that earning power.

A few markers to calibrate the eye (S&P 500, Shiller's series, whose CAPE runs from 1881):

| Era | CAPE | What followed (real, 10 to 15 years) |
|---|---|---|
| 1881 to 2026 average | ~17 | ~6.5% a year over the very long run |
| 1921 (the postwar low) | ~5 | The Roaring Twenties: over 15% a year |
| 1929 (before the crash) | ~33 | Negative over ten years; one of Bengen's worst vintages, with 1966 |
| 1966 (peak of the great postwar bull market) | ~24 | About 0% real over fifteen years: the worst US start date ([[the-trinity-study]]) |
| 1982 (the stagflation low) | ~7 | The biggest bull market of the century: over 13% a year |
| December 1999 (the internet bubble) | ~44 | The all-time record; two 50% crashes in the decade that followed |
| 2009 (the financial-crisis trough) | ~13 | ~12% a year real over the following decade |
| 2024 to 2026 | 32 to 42 | Still being written; historically this zone has never delivered better than about 4% real over ten years |

A raw level never reads alone. Put it back against its century of history, and above all against the last thirty or forty years. Shiller's series is public and updated every month, so placing today's level takes a minute. Make it the first move of any planning session, before you touch a single return assumption.

## Why it predicts, and what it does not predict

The mechanism is not mystical. It is cash-flow arithmetic. Owning the stock market means owning a claim on future corporate earnings. A shareholder's long-run return breaks into three pieces. First, the earnings yield at the price you paid (roughly 1/CAPE). Second, the real growth of those earnings (historically 1.5 to 2% a year in the United States). Third, the change in valuation between purchase and sale, the expansion or contraction of the multiple. Over one year that third term swamps everything and the CAPE predicts nothing. Over ten to fifteen years it averages out, and only the first two remain, the first of which you know on the day you buy. A CAPE of 33 is an earnings yield of 3%. The "certain" part of your future real return is already capped low, whatever happens to the rest.

Empirically, the relationship is among the most robust in finance, and also more modest than what you usually read. Over the 1,242 monthly US start dates from January 1913 to June 2016, the starting CAPE explains 29% of the variance of real returns over the following ten years (regressing on 1/CAPE; 31% regressing on the log of the CAPE). At fifteen years it rises to 36%, and to 41% in logs. The number depends heavily on the window: about 43% for start dates before 1950, 25% for those after. Going back to 1881 on Shiller's series does not help, it hurts (28% at ten years).

Keep the ballpark: a third of the variance at ten years, a bit more at fifteen. That is both enormous, since nothing else does better, and far too little for timing, since at any given CAPE the range of ten-year outcomes stays wide. The honest phrasing fits on one line. The CAPE shifts the center of the distribution of future returns without narrowing it much. That is exactly the information a planner needs, and exactly the information a trader can do nothing with.

::: figure cape-dix-ans
Each dot is a start month between January 1913 and May 2016: its CAPE on the horizontal axis, the annualized real return of the S&P 500 (total return, deflated by US CPI) over the following ten years on the vertical axis. The curve is the least-squares fit of the return on 1/CAPE, R² = 0.29; the three columns show the range of outcomes at a given CAPE.
:::

Three things the CAPE does not do, worth fixing in your head before we go further:

- **It has no sense of timing.** The CAPE was already well above its historical average in early 1996. The bubble ran four more years and more than doubled prices. Shiller himself published *Irrational Exuberance* in March 2000, a legendary piece of timing, but his indicator had been flashing for years. Getting out of the market on a CAPE signal is the strategy that has ruined the most careful people.
- **It does not predict crashes.** An expensive market can deflate through a crash (2000) or through years of flat prices while earnings catch up, which is at least part of the long digestion from 2000 to 2013. The CAPE predicts weak average returns, not their choreography.
- **It does not travel naively across countries or eras.** More on that in the criticisms section.

## The CAPE and the withdrawal rate: the numbers

Now the heart of it for a retiree. A retirement is decided mostly in its first decade ([[sequence-of-returns]]), and the CAPE predicts precisely the average return of the decade that follows. So you would expect a strong link between the CAPE at the start and the maximum sustainable withdrawal rate of that vintage (SAFEMAX, [[the-trinity-study]]). There is one, and it is among the best replicated results in the literature. At ERN, the starting CAPE alone explains about 70% of the variance in the sustainable rates of US vintages, far ahead of every other candidate, including the inflation rate on the day of departure. Bengen sketched it in 2006. Kitces documented it in 2008 ("Resolving the Paradox: Is the Safe Withdrawal Rate Sometimes Too Safe?"). Pfau turned it into a regression in 2011. ERN systematized it on monthly data, in Part 3 and then Part 54 of his series ([[the-ern-series]]).

The ballpark numbers that come out of ERN's work (30-year retirements, a roughly 75/25 portfolio, US monthly data from 1871 to 2016; the exact thresholds move from one study to the next, the structure never does):

| CAPE at the start | Historical frequency | SAFEMAX (30 years) | What it means |
|---|---|---|---|
| Under 15 (cheap market) | about a third of the time | 5.5 to 13% | The 4% rule is very conservative |
| 15 to 20 | about a quarter of the time | ~4.5 to 5.5% | 4% has room to spare |
| 20 to 30 | about a third of the time | ~3.8 to 4.5% | 4% works, with no margin |
| Over 30 (expensive) | about a tenth of the time, but **often** in recent decades | ~3.2 to 3.8% | A rigid 4% sits in the historical failure zone |

And for FIRE-length horizons (50 to 60 years), ERN finds that above a CAPE of 30 the rate that would have survived every vintage drops toward 3.0 to 3.25%. Put another way, the famous modern range of "3.25 to 3.5% for an early retirement" ([[the-4-percent-rule]]) is not an all-weather average. It is already the number conditional on leaving into an expensive market, which is where most FIRE candidates stand today. The reverse holds too: someone leaving after a big bear market, at a CAPE of 15, can legitimately withdraw far more. The 1982 vintage carried more than 7%.

This result has a deep conceptual consequence. The "safe withdrawal rate" is not a constant, it is a function of the entry price. The 4% rule averages start dates that have nothing in common, and the CAPE lets you un-average them. Hence the next generation of rules, the CAPE-based family, whose canonical form (ERN Part 54) is:

> withdrawal rate = a + b × (1/CAPE)

with a of about 1.5 to 2% (the part of the withdrawal funded by earnings growth and by the rest of the portfolio) and b of about 0.4 to 0.5 (the sensitivity to the earnings yield). Take a = 1.75% and b = 0.5. At a CAPE of 20 the withdrawal comes to 4.25%. At 33 it falls to 3.27%. At 12 it climbs to 5.9%. Those rules, their dynamic behavior (the rate is recomputed every year on the current CAPE and the current portfolio, which makes them disciplined cousins of the fixed percentage) and their parameters have their own article ([[cape-based-rules]]).

::: science Why the link is so strong: the three channels
The starting CAPE drives the SAFEMAX through three channels that stack. Channel 1, the expectation. A low earnings yield mechanically means a lower average return over the decisive decade ([[arithmetic-vs-geometric-returns]]). Channel 2, mean reversion. Leaving at a CAPE of 35 means running the risk that the multiple falls back toward 20, which is a 40% cut in valuation to absorb while you are withdrawing. Leaving at 12 puts that wind at your back. Channel 3, the correlation with inflation. The great multiple-compression episodes (1966 to 1982) are often inflationary, and inflation lifts indexed withdrawals at the exact moment the portfolio is taking the hit ([[inflation-and-withdrawal-rates]]). The three worst configurations in the history of withdrawal combine all three channels.
:::

## The serious criticisms of the CAPE

An indicator this widely used has been attacked from every direction. Most of the attacks hold some truth, and knowing them keeps you clear of the two opposite mistakes: ignoring the CAPE, or reading it to the decimal.

**"Accounting standards have changed."** True. Today's GAAP earnings are not those of 1950: writedowns are more aggressive, especially since 2001, and stock options and intangibles are now expensed where they used to be capitalized. Net effect, modern earnings are understated on a constant-method basis, so the modern CAPE is overstated by a few points against century-long comparisons. Jeremy Siegel made this his central criticism, and proposed a CAPE built on NIPA earnings (the national accounts), which comes out structurally lower.

**"Buybacks distort the comparison."** Partly true. Companies now return more cash through share repurchases than through dividends. Buybacks shrink the share count, so earnings per share grow faster than they did in the dividend era. A denominator averaged over ten years lags that growth, which leaves it a little too low and the CAPE a little too high. The proposed fix is Shiller's own Total Return CAPE.

**"Interest rates justify higher multiples."** This was the dominant argument of the 2010s. At zero real rates, discounting future earnings justifies a CAPE of 30 and up. Shiller answered with the Excess CAPE Yield (ECY): 1/CAPE minus the 10-year real rate, the shareholder's premium over bonds. The ECY does put stock valuations in perspective against bonds. But for a retiree the consolation is thin. A world where stocks and bonds both promise little (2021, a CAPE of 38 and negative real rates) is a world where the sustainable withdrawal rate is low, full stop ([[expected-returns]]). And the rise in real rates through 2022 and 2023 took most of the force out of the argument anyway. At a 2% real rate, a CAPE of 35 is hard to justify again.

**"The sector mix has changed."** True. An index that is 30 to 40% technology, with high margins and low capital intensity, "deserves" a structurally higher multiple than an index of 1970 industrial conglomerates. This is hard to quantify cleanly. Mostly it argues against comparing today's absolute level with pre-1990 averages.

**The practical synthesis** of these criticisms comes in two parts. The modern US CAPE is probably overstated by 3 to 8 points in a naive century-long comparison, and the mean it reverts to is no longer 17 but something closer to 22 or 25. But, and this is the decisive point for us, those corrections move the level, not the slope. Even corrected, a CAPE of 35 sits in the expensive quintile of its own era, and "more expensive at the start means a lower SAFEMAX" survives every published correction. For planning use, which is ordinal and works in broad zones, the criticisms call for humility about the exact thresholds, not for dropping the tool.

::: attention The long-run-average fallacy
The most common misuse fits in one sentence. "The CAPE has been above its historical average since 1991 except for a few months in 2009, so it is broken, so I ignore it." That confuses two uses. As a position signal, meaning are you in the expensive quintile of your own era, the CAPE still works. 1999 and 2021 really were relative peaks, 2009 a relative trough, and the returns that followed confirmed it. As a signal of reversion to an eternal mean of 17, it has indeed been broken for thirty years. Use the rank, the percentile within the last 30 to 40 years, not the distance from the average since 1881.
:::

## The four uses in a plan, safest first

**Use 1: calibrate the plan's expected return (recommended).** This is the most direct use and the least debatable. Since 1/CAPE estimates the central component of real equity returns over the coming decade, feed that into the model instead of a historical average. The move, known as a CAPE anchor, takes one line. You replace the central model's mean, and only its mean, with the estimate implied by today's CAPE, weighted by the equity share of the portfolio, and you leave volatility and tail thickness at their usual values. The entry price shifts the center of the distribution, not its width, and the model should say nothing more than that. Few tools lay that anchor in one click, but any parametric model whose mean you can edit lets you lay it by hand ([[expected-returns]]). In an expensive market the effect is plain. Central-case failure probability rises by several points, which is information, not punishment. It is the price of your entry point, made visible. A plan that only holds without the anchor is a plan betting on "this time is different." (How the simulator lays this anchor, behind its "Anchor return to today's valuation (CAPE)" control, is spelled out in [[under-the-hood]].)

**Use 2: size the initial rate (recommended).** When you set your multiple ([[how-much-you-need]]), look at the CAPE zone. Above 30, size on a rigid 3 to 3.5%, or plan explicit margins. Below 20, which usually means you are leaving after a big bear market, congratulations, 4% and more can be defended historically. This is a coarse use, robust to every criticism above.

**Use 3: steer the withdrawal year by year (the CAPE-based rules, for the rigorous).** The withdrawal rate is recomputed every year from the current CAPE. It is more sophisticated, with excellent properties, since the withdrawal falls in bubbles and rises in troughs and the rule is countercyclical by construction, and with real demands on discipline. See the dedicated article ([[cape-based-rules]]) and the comparison with the other strategies ([[choosing-your-strategy]]).

**Use 4: shift the departure date (in small doses).** Since leaving expensive is risk factor number one, pushing back by a few quarters a departure planned at the top of a euphoria, or taking the window that opens after a market has been cleaned out, is worth real money ([[the-three-phases]], [[one-more-year]]). The limit is psychological. The CAPE can stay expensive for ten years, and "I will leave when it gets cheaper" is one more year with no exit condition. If you use this one, bound it, with a dated condition ("by this date at the latest") and a plan B (leaving with partial income, [[going-back-to-work]]).

And the forbidden use, binary portfolio timing: sell everything at a high CAPE, buy it all back at a low one. Exit strategies built on a valuation signal destroy value on average, because they miss the ends of bubbles, which are the best years, in exchange for protection that arrives far too early. Every study confirms it. The CAPE sets the plan, meaning the spending, the rate and the expectations; it does not set the portfolio. The one defensible exception, a mild one, is the glide path around the departure date ([[glidepaths]]), which can take the valuation regime into account.

::: exemple Two departures, two worlds
Twins: $1.3M, a global 60/40, a 45-year horizon, the same $45,000 a year of target spending (3.46%). Amel leaves in January 2000, at a CAPE of 44. Boris leaves in January 2010, at a CAPE of 20. With the expectation anchored on the valuation of the day, Amel's central model runs on about 2.3% expected real equity return over the decisive decade. Central-case failure probability goes above 15%, verdict: not as it stands. What is left is to price the fixes, all of them equivalent in failure probability. Cut to $39,000 (3%), or add $1,000 a month of side income for 8 years, or push the date back by 30 months. Boris, at a CAPE of 20, runs on about 5% expected and about 4% central-case failure probability: plan approved, no changes. Twenty years later, history has settled it exactly that way. The US vintage of 2000 came within a hair of the red zone and survived 4% only thanks to the 2010s. The 2009 to 2010 vintage is one of the most generous of the century. Two identical plans, two entry prices, two fates. That is all the CAPE gives you, and it was there on the day they left.
:::

## The CAPE outside the United States, and your own portfolio's CAPE

Almost everything above is calibrated on the S&P 500, because that is where the long data is. Three notes for anyone holding a global portfolio.

**National CAPEs exist** (Barclays-Shiller, Research Affiliates, StarCapital), and the link between valuation and future returns holds in every market studied, with roughly the same slope. The levels, though, do not compare naively from one country to the next: sector mix, accounting standards and governance all differ. Japan "deserved" higher CAPEs for decades, Europe lower ones. Compare each CAPE with its own history.

**A global portfolio dilutes the problem without removing it.** The US market is 60 to 70% of the world indices. When it is expensive, your world equity fund is expensive. The non-US part, structurally cheaper in recent years, lifts the aggregate earnings yield by a point or two, which is real but not decisive. With no world series as long or as clean, common practice is to take Shiller's US CAPE as the measure of global expensiveness. It is a conservative approximation, and better to own it as one than to cobble together a homemade aggregate.

**The CAPE says nothing about your bonds, your gold or your alternatives.** It measures the equity engine. The expectation for the rest of the portfolio is set another way: current real yields for bonds, one of the rare cases where the expectation is almost literally printed on the label ([[bonds-in-retirement]], [[expected-returns]]).

## The essentials

- The CAPE is price divided by average real earnings over 10 years. It measures expensiveness, and it predicts the center of real returns over 10 to 15 years (R² of about 0.3 at ten years and 0.4 at fifteen in the United States), not their timing and not their crashes.
- Every one of the worst retirement vintages starts at a high CAPE. The "safe withdrawal rate" is a function of the entry price. Above a CAPE of 30 with a long horizon, the historical zone is 3.0 to 3.25% rigid; below 15 it clears 5%.
- The criticisms (accounting, buybacks, rates, sectors) move the thresholds, not the slope. Use it ordinally, by zones, never to the decimal. Read the rank within the last 30 to 40 years, not the distance from the average since 1881.
- Four legitimate uses, safest first: calibrate the model's expectation (the CAPE anchor), size the initial rate, steer the withdrawal ([[cape-based-rules]]), and shift the departure date a little, with a dated bound. One forbidden use: binary portfolio timing.
- The habit worth building: place today's CAPE within its own history, rerun the plan with the expectation anchored on it, and if it no longer holds, price what your entry point costs, in dollars of spending, in working years, or in flexibility.

---

## Going further

- Campbell & Shiller, "Stock Prices, Earnings, and Expected Dividends" (1988); Shiller, *Irrational Exuberance* (2000, later editions with the ECY): the sources.
- The data: Robert Shiller's site (Yale) publishes the CAPE series from 1871 to today, updated monthly. It is the reference series, the one behind every number on this page; [multpl.com](https://www.multpl.com) republishes it in a format that is easy to pull.
- Early Retirement Now, SWR Series Part 3 (CAPE and SAFEMAX), Part 18 and Part 54 (the CAPE-based rules): the formalization for retirees ([[the-ern-series]]).
- Michael Kitces, "Resolving the Paradox: Is the Safe Withdrawal Rate Sometimes Too Safe?" (2008): the valuation-withdrawal link from the practitioner's side.
- Wade Pfau, "Can We Predict the Sustainable Withdrawal Rate for New Retirees?" (2011): the SAFEMAX regression on the CAPE.
- Research Affiliates (Asset Allocation Interactive) and the Global Investment Returns Yearbook: CAPEs and expected returns by country, kept up to date.
- Next in this book: [[cape-based-rules]] (the dynamic use), [[expected-returns]] (forward-looking expectations across every asset class) and [[making-monte-carlo-relevant]] (how a compressed expectation enters a model).
