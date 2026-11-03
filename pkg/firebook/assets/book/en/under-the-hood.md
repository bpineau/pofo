# Under the hood: how the simulator computes this book
<!-- source: la-machine-pofo @ c2b70f88bff4 -->

Every chapter of this book leans on pofo's FIRE simulator. The last article returns the favor: here is what the machine computes, with no black box left in it. Where its data comes from (the backfilled histories, the century-long panel, the living CAPE). How the central model gets built, from your own holdings to parameters fitted and then blended. What each column of the results table draws on, in a technical rundown of the six lenses. How the simulation core runs a plan year by year: spending rules, the tax that grosses up every sale, the buffer and its thresholds, the flows. And what each section of the page computes. Then, in the spirit of the whole book, the **limits**, owned up to. A tool whose simplifications you cannot see is an oracle, and this book has said enough against oracles ([[simulator-traps]], [[monte-carlo-strengths-and-limits]]). This is also the most version-dependent chapter in the book, because the machine keeps changing. The page's own "How this machine works" panel is the authority on its exact state on any given day ([[using-the-fire-simulator]]).

::: cle The design philosophy, in three choices
**One**: everything in **real** terms. The series are deflated by the HICP and withdrawals are set in constant purchasing power. Average inflation is in the machine by construction, and the risk of an inflationary episode shows up in the regime columns ([[inflation-and-withdrawal-rates]]). **Two**: the weight of the evidence rather than a verdict. Four families of models run side by side, two of which your dials cannot touch, so the data keeps the last word ([[historical-vs-parametric]]). **Three**: the real plan rather than a caricature. The simulation runs your actual spending rules, the actual tax on sales, the actual buffer. Plan complexity is free in a simulation, so the machine does not skimp on it ([[making-monte-carlo-relevant]]).
:::

## The data: what the machine knows about the world

**The long history of your funds.** In portfolio mode, pofo extends every holding back beyond its quoted history. The SIM extensions splice on documented backcasts: long indices, reconstructed composites. Each recipe is calibrated and versioned. Managed futures, to take one, are a vol-targeted TSMOM backcast ([[managed-futures]]). Everything is then converted into euros and **deflated** by the HICP (`^HICP-FR`). What comes out is a matrix of **monthly real returns** for your holdings, decades deep.

**A century across sixteen countries.** The broad-sample model carries the Jorda-Schularick-Taylor academic panel (annual real stock and bond returns for 16 developed countries, 1870 to 2020): the raw material for replaying national fates ([[anarkulova-cederburg]]).

**A living CAPE.** Shiller's series (1881 to today) is kept current. §00 of the simulator and the CAPE anchor both feed on it ([[valuations-and-cape]]).

## The central model's pipeline

Five steps run from your holdings to the central Student-t model ([[making-monte-carlo-relevant]] argues each one).

1. **The monthly real panel**, at the weights your holding sliders sit at **right now**.
2. **The fit.** The machine computes μ (the realized mean, annualized), σ (the monthly dispersion scaled to a year) and df (the inverse of the monthly kurtosis, so the tails of **your** funds, [[fat-tails]]). Computed that way, σ lands below the annualized daily volatility the brochures quote.
3. **Blending.** The dials are pulled toward the broad-sample prior (μ 4.5%, σ 13%, df 4, a global equity portfolio), in proportion to the share of the horizon your history does not cover, and never past 50/50. The blend applies to dials you set by hand as well: with 20 years of history for a 40-year horizon, a σ entered at 7% is simulated at 10%. Your portfolio reaches the central model through its statistics, never through its sequence.
4. **The optional anchors.** The CAPE anchor replaces the mean with what today's valuations imply, before any blending, and leaves σ and df as they are. The "Broad-sample prior" button, for its part, imposes that prior at once by rewriting all three dials (nothing to do with the Broad-sample column). Neither touches the historical columns, broad sample included.
5. **The draw.** Independent Student-t years are drawn and then compounded, so volatility drag falls out on its own ([[arithmetic-vs-geometric-returns]]).

**The six columns, exactly what each one does** ([[historical-vs-parametric]] for how to read them):

- **Historical windows**: every possible start month in your panel opens a window replayed as it happened (with an honest refusal when your history is shorter than the horizon).
- **Block bootstrap**: a stationary bootstrap of your panel, with random block lengths averaging 24 months.
- **Student-t**: the central case of the pipeline above.
- **Sequence stress**: a two-state Markov chain (sticky bears: about 19% of years in a bear market, in episodes of roughly three years, volatility 1.5 times higher inside them, **long-run mean preserved**). Overall volatility rises a little along the way, so the gap to the central case measures mostly what a bad sequence costs, not that alone ([[sequence-of-returns]]).
- **Broad-sample**: a block bootstrap of the JST panel in a domestic 60/40, blocks of ten years on average, each drawn from a single country, which can change from one block to the next. Missing years (wars, mostly) cut the series. Your portfolio and your dials are ignored, deliberately: it is an external yardstick, not a measure of your plan.
- **Lost decade**: a sticky Japan-style bear regime, where you spend a third of the time on average, in ten-year episodes, at μ − 0.8σ with volatility 1.4 times higher. The long-run mean therefore falls below μ, and the more so the higher the volatility you enter: for μ 4.5%, the bear regime returns +1.3% real at σ 4%, but −5.9% at σ 13%. It is a built endurance test, not the probability of a Japanese decade for your portfolio.

## The core: a plan run year by year

Every simulated path is complete accounting, under the simplified conventions described here ([[monte-carlo-strengths-and-limits]]). At each time step, a year by default and a month when monthly withdrawals are switched on, the core does four things, in this order. The withdrawal happens at the start of the period: 100 of capital, a need of 10 and a +10% return leave 99 at year end, not 100.

1. **The need** for the period is set by whichever rule is active, then reduced by the flows arriving at that date (the pension from its start year, side income until its end year, an annuity if part of the capital was converted). Each rule has its own formula: the fixed inflation-adjusted rule serves the same real need every year; the flexible cut bites when drawdown crosses its threshold; guardrails move spending by plus or minus 10% at the edges of the corridor, with their floor; the fixed percentage takes a share of the current portfolio; ABW reprices the annuity on wealth plus discounted pensions; the bounded rule follows the Vanguard corridor; the ratchet and the retirement smile ride on top of whatever comes out ([[withdrawal-strategies-overview]]).
2. **The source** of the withdrawal is picked. The buffer pays first when drawdown is past its threshold (10% by default), the sleeves otherwise. The buffer then refills gradually as calm returns, never above its ceiling, and stops refilling in the year you chose ([[refilling-the-buffer]]: the mechanics described there are literally the ones in the code).
3. **Tax** is applied. Every sale is **grossed up** so the net lands where it should: the rate on the dial hits the gain share of the sale. That share is set at the start (50% by default, since no long accumulation has a cost basis equal to its value), then follows the path, rising with gains and with the inflation that erodes the real cost basis, the way a weighted-average cost basis behaves. Optional tax wrappers each carry their own rate and are drained in a fixed order, taxable account first, then the PEA, then assurance-vie: a simple order, not an optimization.
4. **The return** for the period is applied to what is left in the sleeves.

Ruin is recorded in the first year the planned withdrawal can no longer be funded in full, with its date. Everything else is recorded too: wealth, the spending delivered, time spent underwater. The machine draws 2,000 paths by default, and the dial goes to 10,000.

## The sections: what each one computes

A quick tour, with the reading advice for each one in the manual ([[using-the-fire-simulator]]).

- **§00**: today's CAPE against its own century.
- **§01**: the wealth fans under the four lenses (percentiles date by date, eight paths at regular ranks, the red ones counting ruin, the axis clipped at 10x, [[reading-a-fan-chart]]).
- **§02**: your plan replayed at the famous start dates on the long data (1929, 1966, 2000, Japan 1990).
- **§03**: ruin broken out by what the first decade of each path returned (the sequence made visible).
- **§04**: the distribution of the standard of living actually **delivered**, and who pays for it (the judge of flexible rules).
- **§05**: a couple's mortality crossed with ruin ("alive, broke or gone"; the page uses French life tables), the distribution of bequests, and the **causes** of failure (an early crash, a slow grind or longevity, the diagnosis that picks the defense).
- **§06**: the frontier of the rules (ruin against how much the lived standard of living swings) and the sensitivity of each lever.
- **§07**: the buffer trade-off (a sweep from 0 to 10 years) and the histogram of years spent underwater.
- **§08 and §09**: the plan in numbers under the selected model (in the models table, ruin follows your rule, but the safe rate is solved as a fixed withdrawal), then the **solver**, the equivalent moves that each bring ruin back under your threshold (the anti-OMY menu, [[one-more-year]]).

::: attention The limits, faced head on
Here is the list the page's "Method & honest caveats" panel shows, with a comment on each.

- The **portfolio is aggregated** before the simulation runs. There is no fine rotation between asset classes along a path; the regime grid belongs to the design stage, and the machine then tests the mix you settled on ([[market-regimes]]).
- The **central model is i.i.d.** The stress and broad-sample columns exist precisely for that limit.
- **Tax** stays simplified: one shared gain share, a rate per wrapper that you calibrate yourself, a fixed selling order.
- Your funds' **history** is a kind window. Blending says so and the warnings say so, but no correction turns a short window into the whole story.
- **Mortality** is ignored by default, and §05 puts it back in.
- The usual items stay **out of scope**: divorce, long-term care, political tail risk, all left to named margins and to the chapters that deal with them ([[hyperinflation-and-extremes]]).

One reading rule follows from the whole book: read each column for the question it asks (your portfolio, or an external yardstick), push the horizon past your life expectancy, and read the decimals as a ranking ([[failure-probability]]).
:::

## Checking it yourself

There are three ways to audit the machine. The engine is Go code you can read: in the pofo repository, the models, the withdrawal rules and the backcast recipes are all legible and all tested, and the golden tests compare the computations against frozen external references. The book's **headline results** reproduce there in a few minutes. A SAFEMAX near 4% over 30 years falls out of the historical windows; the Trinity cliff appears when you sweep the withdrawal rate; the effect of sticky regimes shows up as the gap between the central case and the stress column; the pension effect shows up the moment you move its dial. The matching chapters give the recipe each time. And **disagreements** with other tools are explained by the audit checklist: tails, memory, frictions, sample ([[simulator-traps]]). This simulator ticks boxes that most consumer tools leave empty, and it shows the ones it leaves empty itself. That is the final test to apply to **any** tool, this one included. The question is not "is it right?", because nobody is right about 45 years, but "do we know what it does, and does it say so?".

## The essentials

- Three design choices: everything in real terms (deflated by the HICP), the weight of the evidence rather than one verdict (two columns ignore your dials, so the data keeps the last word), and the real plan simulated year by year (rules, tax, buffer, flows).
- The central model is built in five steps: the monthly real panel of **your** holdings, then the fit (μ, long-horizon σ, df for the tails), then blending toward the broad-sample prior (capped at 50/50), then the optional anchors (CAPE, the "Broad-sample prior" button), then annual Student-t draws.
- The six columns: your own windows replayed, your panel bootstrapped, the central case, the mean-preserving stress (mostly what a bad sequence costs), the JST century in a domestic 60/40 (an external yardstick), and the lost decade (an endurance test). Every disagreement between them is a diagnosis ([[historical-vs-parametric]]).
- The core grosses up every sale for tax (with a gain share you set, which then follows the path), draws on the buffer at its thresholds, and books pensions and side income on their dates: the machinery this book describes is literally the machinery in the code.
- The limits are on display: aggregation, an i.i.d. central model, simplified tax, short windows, what is out of scope. And that is the test worth applying to any tool: not being right, but saying what it does.

---

## Going further

- The "How this machine works" and "Method & honest caveats" panels of the FIRE simulator itself: the machine's exact, current state.
- The manual: [[using-the-fire-simulator]]. The reasoning behind each choice: [[making-monte-carlo-relevant]], [[historical-vs-parametric]], [[fat-tails]].
- Jorda-Schularick-Taylor (the broad-sample panel) and Shiller's website (the CAPE): the public data the machine ships with ([[the-library]]).
