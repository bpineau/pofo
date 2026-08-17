# Historical windows, bootstrap, parametric: the three families of models
<!-- source: historique-vs-parametrique @ 7eb651ee729b -->

A retirement simulator runs plans through generated futures ([[monte-carlo-strengths-and-limits]]). Everything turns on where those futures come from. There are only three broad answers, three families of models. You can replay history as it happened (historical windows). You can resample it, shuffled (the bootstrap). Or you can generate synthetic returns from a handful of parameters (parametric models).

Each family really answers a different question. Each has its virtues and its blind spots. And when their verdicts disagree about the same plan, that is not a glitch: it is the most valuable information you will get. This page takes all three in turn. For each one, the exact mechanics, the implementation choices that matter, the strengths and the traps. At the end comes the combined reading, which says what to believe for which question.

::: cle Three families, three questions
Historical windows: "how would my plan have done if I had retired on each date in the past?" Bootstrap: "and in **plausible** histories, made of the same ingredients as the past but put together differently?" Parametric: "and in a world whose mean, volatility and tails I set myself?" None of them answers "what is going to happen?"; together they **bracket** that unanswerable question. Which is why serious simulators show several families side by side instead of one verdict.
:::

## Family 1: historical windows (replay, cohorts)

**The mechanics.** This is Bengen's founding method ([[the-trinity-study]]). Take the real return series of your portfolio, or of an index. Then run the plan from every possible start date: the window from January 1975 to December 2019, then February 1975 to January 2020, and so on. Each window is a "vintage", a cohort. The failure rate is the fraction of vintages that ran out of money.

**The implementation choices that matter.** The sampling step, first. Starting once a month rather than once a year multiplies the windows by twelve and preserves the within-year sequences; that is the standard practice, and ERN's ([[the-ern-series]]). Then honesty about the horizon. When the history is shorter than the plan (20 years of data for a 45-year retirement), no complete window exists and the failure rate simply cannot be computed. A serious tool says so and refuses to answer, instead of quietly extrapolating. Last, the directed variant. Rather than running every start date, you replay the worst ones: the United States in 1929, 1966 and 2000, Japan in 1990. Real vintages picked for the ordeal, and the most vivid version of the family ([[the-trinity-study]]).

**The strengths.** Absolute fidelity to what happened. The correlations between assets, the clusters of crises, the crash-inflation-recovery chains: it is all in there, because nothing is modeled and everything is quoted. This is the only model whose every path really occurred. Hence a teaching power nothing else matches (watching your plan cross 1966 says more than a thousand probabilities) and an excellent detector of fragility against real regimes.

**The traps.** There are three, and they are serious. First, the tiny sample. A hundred years of data hold only three or four independent long retirements. The windows overlap massively: the 2008 crash shows up in 350 monthly windows. So the "failure rate" carries enormous error bars, which its display never shows. Second, the ceiling of what happened. The worst of the past is treated as the worst there is, when nothing guarantees the worst is behind us. Third, the bias of the available window. Your funds' history usually covers the recent decades, which have been kind. Historical windows therefore read as an optimistic bound, never as the verdict.

## Family 2: the bootstrap (block resampling)

**The mechanics.** The bootstrap answers the tiny-sample problem. Rather than replaying history in order, you cut it up, draw pieces at random with replacement, and glue them back into synthetic histories. Drawing month by month would destroy the clusters and the trends, since you would be back to i.i.d. So you draw blocks several years long. The reference variant is the stationary bootstrap (Politis-Romano, 1994): blocks of random length, an average of 24 months being a common choice, which avoids the cutting artifacts of fixed-size blocks. Each simulated path is then a history that never happened, though every two-year piece of it did happen, with its internal correlations and much of its memory.

**One engine, two questions.** Everything depends on the panel you reshuffle. Applied to your own funds' history at your current weights, the bootstrap builds thousands of variants of what your holdings actually lived. Applied to a long multi-country panel, it changes nature. That is the Anarkulova-Cederburg method: resample the developed world's century (the academic Jorda-Schularick-Taylor panel covers 16 countries from 1870 to 2020), drawing each block inside a single country so the great disasters (1929, the 1970s, Japan) survive intact in the paths, on a domestic 60/40 portfolio in line with the literature ([[anarkulova-cederburg]]). Same mathematical family, two different questions: "my funds, reshuffled" against "the developed century, reshuffled".

**The strengths.** The best trade-off between fidelity and variety. You keep the correlations and the short memory of the real thing, through the blocks, and you get thousands of distinct paths, through the reshuffling. The broad sample adds depth on top, with regimes your funds' history has never seen. This is the family modern research prefers for long-horizon risk, and it is Anarkulova-Cederburg's.

**The traps.** Memory beyond the block is destroyed. A seven-year bear market cannot come out of two-year blocks, short of the bad luck of several dark draws in a row. Mean reversion in valuations, spread over decades ([[valuations-and-cape]]), disappears too. The ingredients, next, are still those of the available past: the bootstrap reshuffles, it invents nothing. Applied to your funds' history alone, it inherits family 1's window bias, since variety in the draws does not fix poverty in the ingredients. Finally, block length is a real parameter. Too short and you kill the clusters; too long and you fall back into replay and its thin sample. The 24-month average follows the practice of the literature: long enough to hold a typical recession, short enough to diversify.

## Family 3: the parametric models (Student-t, and regimes)

**The mechanics.** You drop the raw data. You describe the world with a few parameters and draw from them. The simplest version is Gaussian i.i.d. (a mean, a volatility, independent annual draws), the one most commercial simulators use. It has two fixable defects: tails that are too thin, and no memory. A serious parametric model fixes the first by drawing from a three-parameter Student-t: μ (the mean), σ (the volatility) and df (how fat the tails are). At df 5, a −30% real year is about ten times more likely than under a normal law ([[fat-tails]]). The best tools draw monthly and compound into years, fit the three parameters to your funds' data, and blend them toward a prudent world prior when the horizon runs past the history ([[making-monte-carlo-relevant]]).

The second defect, the missing memory, gives rise to the sub-family of regime models, parametric but sequenced. A Markov chain alternates a "normal" state and a "bear" state, with transition probabilities that make those bear markets persistent: entering one is rare, staying in it is likely, and bad years arrive in clusters of about three years. That is the principle behind a sequence stress. Built properly, it keeps the same long-run mean as the central model: the stress measures the risk of order, not a hidden pessimism about the level. Only the volatility gets concentrated into episodes. Its extreme variant is the lost decade: a Japan-1990 bear market, long and deep, deliberately left uncompensated, with the mean pulled down. That one is a tail scenario to survive, not an expectation.

**The strengths.** Transparency and control, first: three explicit numbers, no black box. You can plug in forward-looking expectations ([[expected-returns]]) or the CAPE anchor ([[valuations-and-cape]]), and test "what if σ rose by two points". Generality, next: parametric models explore worlds history never produced, which neither replay nor the bootstrap can do. And isolating causes, last: the central/stress pair, identical in everything but the order of the years, measures what the sequence costs in your plan ([[sequence-of-returns]]). No other family allows that controlled experiment.

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

Three remarks for using this table. First, identify the family before you read the verdict. A pure family-1 tool on American data gives you the optimistic bound of the evidence, no more and no less, and that is already a lot. Second, remember that μ, σ and df drive family 3 only; the data models ignore them, and a setting that "does nothing" is not a bug. Third, you can rebuild the multi-model evidence by hand well enough: the same plan entered into two tools from different families is a dashboard on its own. Note in passing that every replay tool named here works on American history; the multi-country century remains, to this day, the rare commodity of the inventory.

## Reading the three families together

That leaves the real question: what do you do when the families disagree, which is the normal state of affairs? Here is the grid, disagreement by disagreement.

**Historical and bootstrap optimistic, the central case harsher.** The most common case: your funds lived through a good window, and blending toward the prior pulls the central case down. The reading: the gap measures the bias of your historical window. Believe the central case, size the plan on it, and keep the historical models as the "the world keeps going as I have known it" scenario.

**The central case acceptable, the sequence stress clearly worse.** Your plan is exposed to the order of returns: a high initial withdrawal, little flexibility, no early income. This is not a problem of expected level but of structure. The defenses are the ones in the anti-sequence table ([[sequence-of-returns]]): flexibility in writing, a cash buffer, a glide path, income in the first years.

**Everything fine except the broad sample.** Your plan holds in the world of your assumptions, but not across the full developed century. The reading: look at where the broad-sample paths fail. It is almost always in the blocks of persistent inflation and in the countries falling behind. The answers are international diversification and regime assets ([[anarkulova-cederburg]], [[all-weather-portfolios]]), not necessarily more capital.

**Even the lost decade passes.** Your plan is oversized. The question is no longer ruin but opportunity cost: years of work too many, capital that will die intact ([[one-more-year]], [[spending-in-retirement]]).

The summary rule, given before but at its most useful here: **plan between the central case and the broad sample, test the order with the stress, put the plan through the lost decade, and keep the historical models as an optimistic bound and as a teaching tool.** Four families of futures, one decision.

::: exemple A useful disagreement
A real portfolio with 15 years of history (a kind window, 2010 to 2025), a plan at 3.9%. The verdicts: historical windows 0%, bootstrap 2%, central case 6%, stress 10%, broad sample 13%. A naive reader picks the model they like best. The grid reasons differently. That window holds neither a long inflation nor a lost decade, so the gap between families 1 and 2 and family 3 gives away a window bias. The plan is also sensitive to order (6 to 10, a withdrawal a touch high, zero flexibility). And the broad sample confirms the vulnerability to inflation. The decision that follows: withdrawal brought down to 3.6%, a written flexibility rule (a 10% cut above a 4.5% current withdrawal rate), and 10% of the portfolio moved into linkers and gold ([[inflation-linked-bonds]], [[gold-in-retirement]]). No single model would have produced that three-part diagnosis. The disagreement did.
:::

## The essentials

- Three families: replay (windows and cohorts), reshuffle (block bootstrap, including the broad sample over the 16-country century), generate (parametric Student-t, and its regime variants for memory).
- Each one answers a different question, and their traps are complementary: a tiny sample and the ceiling of what happened (1), memory truncated at the block and ingredients taken from the past (2), sensitivity to the inputs and a caricature of structure (3).
- μ, σ and df drive the parametric family only; the data models ignore them. Always know which family you are looking at.
- The disagreements between families are the real result: window bias, exposure to order, regime vulnerability, oversizing. Every pattern of disagreement has its diagnosis and its fix.
- The decision summary: size between the central case and the broad sample, test the order with the stress, put the plan through the lost decade, keep the historical models as an optimistic bound and as an object lesson.

---

## Going further

- Politis & Romano, "The Stationary Bootstrap" (1994): the reference method for random-length blocks.
- Anarkulova, Cederburg, O'Doherty & Sias (2023): the block bootstrap applied to the developed century, the reference for the broad-sample model ([[anarkulova-cederburg]]).
- Early Retirement Now, part 8: the systematic monthly replay method ([[the-ern-series]]).
- In this book: [[fat-tails]] (the Student-t choice in detail), [[making-monte-carlo-relevant]] (blending and the anchors of the central model), [[under-the-hood]] (exactly how each of these models is implemented).
