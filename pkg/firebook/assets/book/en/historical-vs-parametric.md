# Historical windows, bootstrap, parametric: the three families of models
<!-- source: historique-vs-parametrique @ 38a466828ac9 -->

How would your plan have held up through the inflation of the 1970s? A historical replay can show you what the returns of the time would have done to your savings. To find out what happens when several bad stretches come back to back, you need a way to change their order. Choose the model to match the question ([[monte-carlo-strengths-and-limits]]).

We'll compare three ways to build a history of returns. Keep spending, pension income and the length of retirement the same across all three so that the differences are easier to trace.

::: cle Three families, three questions
A replay follows the dates in their original order. A bootstrap joins stretches drawn from the data. A parametric model draws returns from a distribution whose properties you specify. All three can miss the same risk.
:::

## Family 1: historical windows (replay, cohorts)

**The mechanics.** This is Bengen's founding method ([[the-trinity-study]]). Take the real return series of your portfolio, or of an index. Then run the plan from every possible start date: the window from January 1975 to December 2019, then February 1975 to January 2020, and so on. Each window is a "vintage", a cohort. The failure rate is the fraction of vintages that ran out of money.

**Choosing start dates.** With 45 years of monthly returns and a 45-year retirement, there is one complete window. One more month of data gives you a second start date. The two households would have lived through almost exactly the same markets. Starting in every month, as ERN does ([[the-ern-series]]), shows the effect of timing. With only 20 years of data, a 45-year plan has no complete historical replay.

**What you learn.** Starting in 1966 preserves the order of stock returns and inflation as they occurred. The market prices are historical. The household's sales, spending and remaining balance are what your rules would have produced.

**What you miss.** A century contains only three non-overlapping 30-year periods. The same crash turns up in many monthly windows, seen each time from a different starting date. The failure rate describes those starts in that market. A particularly prosperous sample can flatter a plan, and a particularly difficult sample can make it look worse.

## Family 2: the bootstrap (block resampling)

**How it works.** Pick a month from the record and copy that month's returns for all the assets together. Continue with the following months, then start again at a randomly chosen date. The same block can turn up more than once. A block might preserve a stock market slump and its recovery, along with what bonds did over those same months. Where two blocks join, the historical sequence breaks.

In Politis and Romano's stationary bootstrap the length of each block varies. With a mean of 24 months, each month brings a one-in-24 chance of restarting at a random date. A three-year crisis can appear in full, be cut short or be missing altogether from a path.

**One engine, two questions.** Everything depends on the panel you reshuffle. Applied to your own funds' history at your current weights, the bootstrap builds thousands of variants of what your holdings actually lived. Applied to a long multi-country panel, it changes nature. That is the spirit of the Anarkulova-Cederburg method (38 countries in their study): resample the developed world's century, here on the academic Jorda-Schularick-Taylor panel (16 countries, 1870 to 2020), drawing each block inside a single country and making it long (ten years on average), so the great disasters (1929, the 1970s, Japan) largely survive in the paths, on a domestic 60/40 portfolio in line with the literature ([[anarkulova-cederburg]]). Same mathematical family, two different questions: "my funds, reshuffled" against "the developed century, reshuffled".

**What you learn.** Changing the order of blocks can bring together losses that originally occurred decades apart. You can see what a different sequence of familiar returns would do to your plan.

**What you miss.** Ten thousand simulated retirements from a twenty-year record still rest on twenty years of observations. Short blocks break up long trends more often. Long blocks preserve more of those trends but allow fewer combinations. If the record has no sustained inflation shock, drawing from it more often cannot tell you how your funds would have behaved through one ([[valuations-and-cape]]).

## Family 3: the parametric models (Student-t, and regimes)

**The mechanics.** You drop the raw data. You describe the world with a few parameters and draw from them. The simplest version is Gaussian i.i.d. (a mean, a volatility, independent annual draws), the one most commercial simulators use. It has two fixable defects: tails that are too thin, and no memory. A serious parametric model fixes the first by drawing from a three-parameter Student-t: μ (the mean), σ (the volatility) and df (how fat the tails are). At df 5 and equal volatility, a −30% real year is about three times more likely than under a normal law ([[fat-tails]]). The most careful tools fit the three parameters to your funds' data, and blend them toward the broad-sample prior (world equities, not your funds) when the horizon runs past the history ([[making-monte-carlo-relevant]]).

The second defect, the missing memory, gives rise to the sub-family of regime models, parametric but sequenced. A Markov chain alternates a "normal" state and a "bear" state, with transition probabilities that make those bear markets persistent: entering one is rare, staying in it is likely, and bad years arrive in clusters of about three years. That is the principle behind a sequence stress. Built properly, it keeps the same long-run mean as the central model: the stress measures the risk of order, not a hidden pessimism about the level. Only the volatility gets concentrated into episodes. Its extreme variant is the lost decade: a Japan-1990 bear market, long and deep, deliberately left uncompensated, with the mean pulled down. That one is a tail scenario to survive, not an expectation.

**The strengths.** Transparency and control, first: three explicit numbers, no black box. You can plug in forward-looking expectations ([[expected-returns]]) or the CAPE anchor ([[valuations-and-cape]]), and test "what if σ rose by two points". Generality, next: parametric models explore worlds history never produced, which neither replay nor the bootstrap can do. And isolating causes, last: the central/stress pair, with the same mean but its bad years bunched together (and a little deeper), shows what an unlucky sequence costs in your plan ([[sequence-of-returns]]). No other family makes that experiment so easy.

**The traps.** They mirror the strengths. Everything rests on three numbers nobody knows, and the input sensitivity described in [[monte-carlo-strengths-and-limits]] is at its worst here. The structure you pick, i.i.d. or a two-state Markov chain, is still a caricature of the real thing: no mean reversion in valuations, no stochastic correlation between assets (the portfolio is aggregated before the draw), and inflation left implicit (everything is in real terms). A parametric model is a laboratory instrument: perfect for controlled experiments, never to be confused with the world.

## The three families in common simulators

Knowing which is which changes how you use any tool, because a verdict only means something once you know which family produced it. The inventory below describes the models the most widely used simulators implement, as their documentation presents them at the time of writing (tools evolve, so check theirs). It is an inventory, not a ranking. An excellent single-family tool is worth more than a catch-all, and a missing family signals a design choice, not a flaw. One disclosure is owed to the reader: the simulator that comes with this book is not one its author can judge from the outside, which is one more reason to cross-check its verdicts against an independent tool.

| Tool | Families | Worth knowing |
|---|---|---|
| cFIREsim | 1 (US replay since 1871) | Free. Shiller data for stocks, bonds and gold, configurable spending plans. The benchmark for simplicity. |
| ERN SWR Toolbox | 1 (monthly US replay since 1871, every start month) | Free. A spreadsheet to copy, auditable formula by formula, side flows and the series' CAPE rules included ([[the-ern-series]]). You have to get your hands into the spreadsheet. |
| FI Calc | 1 (US replay since 1871) | Free. The richest set of ready-made withdrawal rules (a dozen of them, guardrails and VPW included), carefully taught. |
| This book's simulator | 1, 2 and 3 (monthly windows of your own funds; bootstrap of your funds, and the broad sample of the 16-country century; fitted Student-t, sequence stress, lost decade) | The tool built alongside this book. The youngest on the list, taxes reduced to a single blended effective rate, no persistent-volatility model. |
| Portfolio Visualizer | 2 and 3 (draws of historical years, no blocks; normal, Student-t, GARCH, forward-looking expectations) | The GARCH captures volatility clustering, which is rare. The historical draw is year by year, so it has no memory. A good share of the features has gone paid. |
| Rich, Broke or Dead | 1 (US cycles since 1871) | Free. Crosses every cycle with mortality tables: the most vivid "rich, broke or dead" picture of the genre. |
| TPAW Planner | 1 and 2 (historical sequences, and historical draws recentered on today's expectations, 1/CAPE and real rates) | Free. The reference implementation of amortization (ABW), and the only one on the list that anchors its expectations to valuations by default. |

Before comparing two tools, write down the country, period and asset mix each uses. Replaying a US portfolio since 1871 and resampling your own funds since launch changes several assumptions at once. Also check whether the tool uses the observed returns directly or shifts them to match a chosen expected return. That detail explains why changing a return assumption leaves some results untouched and moves others.

## Reading the three families together

That leaves the real question: what do you do when the families disagree, which is the normal state of affairs? Here is the grid, disagreement by disagreement.

**History looks better than the central model.** Compare the average return in the sample with the return assumed by the model. The gap might reflect an unusually good period, a conservative assumption or a difference in the risks represented. Change one assumption at a time and rerun the same plan.

**The central case acceptable, the sequence stress clearly worse.** Your plan is exposed to the order of returns: a high initial withdrawal, little flexibility, no early income. This is not a problem of expected level but of structure. The defenses are the ones in the anti-sequence table ([[sequence-of-returns]]): flexibility in writing, a cash buffer, a glide path, income in the first years.

**Only the international model shows frequent failures.** It applies your spending and pension to a domestic 60/40 portfolio across different countries and periods. Examine the difficult episodes. To test what diversifying your own assets would achieve, return to a model that uses your portfolio weights ([[anarkulova-cederburg]], [[all-weather-portfolios]]).

**The plan survives even a lost decade.** Check how much the household could spend along the way. If essential spending held up, try a concrete change, such as retiring a year earlier. Surviving a single stress test does not establish that you are saving too much ([[one-more-year]], [[spending-in-retirement]]).

Keep each result alongside the question it can answer. A failure rate computed on your funds and one computed on a domestic 60/40 portfolio can both be informative while measuring different risks.

::: exemple A useful disagreement
You have 15 years of prices for your funds and expect a 40-year retirement. The historical-window entry should be blank. A bootstrap can assemble 40 years from those 15, but every block comes from the same short record. To test the plan's sensitivity, keep your portfolio and lower the expected return in the parametric model. Check when essential spending first falls short. You now know which part of the answer came from the data and which depended on an assumption.
:::

## The essentials

- Three families: replay (windows and cohorts), reshuffle (block bootstrap, including the broad sample over the 16-country century), generate (parametric Student-t, and its regime variants for memory).
- Each one answers a different question, and their traps are complementary: a tiny sample and the ceiling of what happened (1), memory truncated at the block and ingredients taken from the past (2), sensitivity to the inputs and a caricature of structure (3).
- A pure replay uses the returns from the dates you choose. Models that shift historical returns to a chosen mean combine resampling with a forward-looking assumption.
- A gap between models gives you an assumption to investigate. Differences in returns, holdings or sample periods can produce the same symptom.
- For each difficult scenario, check how long the household could pay its essential bills and what funded them.

---

## Going further

- Politis & Romano, [The Stationary Bootstrap (1994)](https://users.ssc.wisc.edu/~bhansen/718/Politis%20Romano.pdf), on how the random-length blocks are constructed.
- Anarkulova, Cederburg, O'Doherty & Sias (2023): the block bootstrap applied to the developed century, the reference for the broad-sample model ([[anarkulova-cederburg]]).
- Early Retirement Now, part 8: the systematic monthly replay method ([[the-ern-series]]).
- In this book: [[fat-tails]] (the Student-t choice in detail), [[making-monte-carlo-relevant]] (blending and the anchors of the central model), [[under-the-hood]] (exactly how each of these models is implemented).
