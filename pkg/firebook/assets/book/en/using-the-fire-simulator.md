# Using the FIRE simulator
<!-- source: utiliser-la-page-fire @ ba32f89b87ac -->

pofo's FIRE simulator is a laboratory for a retirement plan, and it runs in a browser. You describe your situation there: capital, spending, age, pension, spending rules. It then runs that one plan through several market models, from the one that follows your own funds most closely to the harshest the century has on record. This article is the full manual. It says which order to read the sections in, what each group of parameters controls, and above all how to **read** what you are looking at. The tool is built to inform a decision, not to hand down a verdict.

::: cle The idea behind the tool
A single ruin figure is one model's opinion ([[failure-probability]]). So the page shows the same plan through several lenses at once. On one side, the models built on **data**: your own funds replayed or resampled, and a century of 16 developed countries. On the other, the models built on **parameters**: the central Student-t, sequence stress, lost decade. What you read is the spread between the columns. The decision to make is the one that still holds up in the pessimistic ones.

The first two lenses need a portfolio. Without one they simply disappear, and the page says so under the gauge: the Confidence chip then reads "Parametric models only". That leaves the parametric models, plus the 16-country century, which needs no portfolio at all.
:::

## Getting started: with a portfolio, or without one

**With no portfolio**, the page opens in parametric mode, where the market is three dials and nothing else: real return μ, volatility σ, and tail thickness df. This is sandbox mode, and it is the right place to learn the mechanics and size things roughly. It starts on a textbook case that describes nobody: EUR 600,000 of capital, EUR 24,000 of net spending a year, retirement at 40, a 42-year horizon, a EUR 12,000 pension from year 25, a three-year buffer, and a market at μ 5%, σ 11%, df 5. Those are demo numbers. Replace them with yours before you read anything into the output.

**With your portfolio**, the tool rebuilds the long real history of **your** holdings (through the `SIM` extensions), deflates it into constant euros, and gets two things out of it. First the central model's parameters: μ, σ and df fitted to your funds, then blended cautiously toward a world prior (more on that below). Then two purely historical models, historical windows and block bootstrap ([[historical-vs-parametric]]). You can also drag each holding's weight and watch the ruin figure recompute live. This is the mode worth the trouble: the two data columns exist at no other price.

On the web, the portfolio travels in the URL: `/firesimulator/e/<example>/` loads one of the example portfolios, `/firesimulator/p/<composition>/` a composition you write by hand, in the same grammar as the comparison page. From the command line, `pofo -fire` opens the bare simulator and `pofo -fire portfolio.txt` opens it with that portfolio loaded.

Every amount on the page is **real** (constant purchasing power: inflation is already out, so a euro in year 30 buys what a euro buys today) and **net of tax** (tax is modeled separately, in the Taxes group). EUR 60,000 a year means EUR 5,000 a month of today's spending, forever ([[arithmetic-vs-geometric-returns]]).

## The dashboard: the model strip

As soon as you scroll, the plan bar pinned to the top of the page reduces each model to a colored chip (green is safe, red is catastrophic at your spending level). The main table lines the columns up: Historical windows, Block bootstrap, the central Student-t, Sequence stress, Broad-sample (the one column that ignores your portfolio entirely and shows raw markets), Lost decade. **The table is a selector.** Click a column and **every** detail section below recomputes under that lens; the amber underline marks the active column. Read them in this order.

1. **Your funds' own history** (windows, bootstrap): the optimistic bound. Your funds lived through one window, and it was usually a kind one.
2. **The central Student-t**: the default planning case, fitted to your funds and then pulled toward caution.
3. **Sequence stress**: the same average return, but the bad years arrive in clusters. The gap to the central case is what a bad sequence costs in your plan ([[sequence-of-returns]]).
4. **The broad sample**: a full century of 16 developed countries (1870 to 2020) in a domestic 60/40, disasters included ([[anarkulova-cederburg]]). It borrows nothing from your portfolio and nothing from the dials, which makes it the honest estimate of long-horizon risk.
5. **The lost decade**: a Japan-style tail scenario. The job is to make it **survivable**, not to argue it is unlikely.

The suggested rule: plan between the central case and the broad sample. Use sequence stress to test the order of returns, and the lost decade as a crash test.

Next to the table, the gauge shows the selected model's ruin against the ruin you are willing to accept (the Acceptable ruin field, 4% by default). Every solver on the page is solved against that threshold ([[failure-probability]] to pick it well).

## The sections, in reading order

**§00 Where we are in the cycle**: today's Shiller CAPE, placed against a century of history. Not a timing signal: a reminder that where you start shapes the first decade, and the first decade decides ([[valuations-and-cape]]).

**§01 Simulated futures**: the wealth fans under four lenses side by side, with sample paths drawn in and the failed ones in red. How to read a fan without fooling yourself: [[reading-a-fan-chart]].

**§02 The retirements that actually happened**: your plan replayed at the worst start dates of the century, the US in 1929, 1966 and 2000, Japan in 1990. Real vintages, not draws ([[the-trinity-study]]).

**§03 The decisive decade**: ruin broken out by what the first ten years of each scenario returned. Sequence risk made visible on **your** plan ([[sequence-of-returns]]).

**§04 The spending you actually live**: what your spending rules really deliver, the standard of living lived year by year and who pays for it (portfolio, pension or buffer). This section matters the moment you switch on a flexible rule. Ruin drops, and here is where you see the **price** you paid for it, counted in years of reduced spending ([[flexibility-in-practice]]).

**§05 Alive, broke or gone**: ruin crossed with a couple's odds of still being alive (the page uses French life tables). Going broke at 61 and going broke at 92 are not the same event. This is where a raw ruin figure gets put back in proportion, or does not.

**§06 What moves the risk**: the frontier of withdrawal rules, where each rule is one point, ruin against how much the standard of living swings ([[withdrawal-strategies-overview]]), then the levers ranked by sensitivity. You see what really moves your risk, and what does not.

**§07 Buffer & recovery**: the cash buffer trade-off, ruin and ending wealth by number of buffer years ([[cash-buffer]]), and the distribution of years spent underwater. It shows how long the crossings last, the stretches inside a bear market that the buffer is there to cover.

**§08 Plan detail** and **§09 Reaching your target**: the plan in numbers under the selected model, and the solver menu. The menu lists the equivalent moves (more capital, less spending, one more year, more pension) that would each bring your ruin back under your threshold. This is the negotiate-with-yourself section, and it puts a price on every margin ([[one-more-year]]).

## The parameters panel, group by group

The parameters button opens the panel. Every control has a plain-language hover. Here is the map, and the traps.

**Your situation**: Deployed capital (your home and your emergency fund excluded), Age at retirement (which drives §05), Horizon (years) (plan **past** your life expectancy: going from 40 to 50 years nearly doubles ruin, [[horizon-and-life-expectancy]]), and Net spending /yr (real, net of tax, since the tax friction is modeled separately).

**Pension & side income**: Pension /yr (net real) and Pension starts in year, then Side income /yr and Side income until year for the temporary income of the early years. Work the pension number out beforehand, as a range: a stressed case, a central case counting only what you have already earned, and whatever your official statement says ([[us-healthcare-and-social-security]] for reading a US one). Early side income is the best sequence insurance there is ([[pensions-and-other-income]]). The pension is the plan's second-biggest sensitivity after spending, so do not leave these at zero out of misplaced caution ([[ten-plan-wrecking-mistakes]]).

**Spending policy**: the strategic heart of the panel. Cut in downturns is the reversible cut applied while the portfolio is in drawdown, and a 15% cut roughly halves ruin. Also cut above this WR adds a second trigger on the current withdrawal rate. Then come the Guyton-Klinger guardrails and their floor ([[guyton-klinger]]), and the risk-based guardrails, which track the rate that is still safe for the horizon you have left ([[morningstar-guardrails]]). Then the upward ratchet, the structural spending drift and the retirement smile ([[spending-in-retirement]]), pure VPW ([[vpw]]), ABW/TPAW amortization ([[amortization-based-withdrawal]]), the Vanguard-style bounded percentage ([[floor-and-ceiling]]) and the inflation-linked annuity ([[annuities-and-safety-first]]). **Exactly one policy runs at a time**: the page makes you choose, and the rule you keep sets the amount withdrawn each year. The engine underneath applies a fixed order of priority for anyone driving it from the library: ABW/TPAW amortization, then the bounded percentage, then VPW, then the risk-based guardrails, then Guyton-Klinger, then the fixed rule with its flexible cut and its ratchet. The "How this machine works" panel takes them one by one, and part IV of this book gives each of them an article ([[choosing-your-strategy]]).

**Market model**: Real growth return is μ, the **arithmetic** real mean of the growth engine, and the geometric return you actually live is about μ − σ²/2. Volatility (long-horizon) is σ, lower than the one-year number the brochures quote. Tail df sets how fat the tails are: at df 5 a catastrophic year is roughly ten times likelier than a normal law allows ([[fat-tails]]). In portfolio mode all three come pre-filled from your funds, then blended toward a cautious world prior in proportion to how far the horizon runs past the history ([[making-monte-carlo-relevant]]). A prior, in statistics, is what you assume before you even look at your own data: here, what a century of developed markets suggests when your funds' history is too short to settle the question on its own. Two anchors earn their place. Broad-sample prior rewrites the dials with the century's cautious assumptions, and Anchor return to today's valuation (CAPE) replaces the mean alone with 1/CAPE, the return compression that current valuations imply ([[cape-based-rules]]). The rising-equity glidepath ([[glidepaths]]) and monthly withdrawals sit in this group too.

**Cash buffer**: Buffer (years of spending), three by default, its real return, and the year refills stop. Mind the convention: the buffer is **carved out of** your starting capital, never added on top ([[cash-buffer]], [[refilling-the-buffer]]).

**Taxes**: Tax on gains is charged on the gain share of every sale, by grossing the sale up. Withdrawing EUR 60k net sells more than EUR 60k of assets, and the effective burden climbs as unrealized gains build up. The 32.8% preset is one country's blended rate and nothing more: replace it with your own blended effective rate across your accounts ([[us-taxes-in-the-withdrawal-phase]]).

**Simulation**: Simulated paths sets how many paths each model draws, 2,000 by default and 10,000 at most. At 2,000 the ruin figure moves by roughly plus or minus 0.7 point from one run to the next with nothing changed. That is sampling noise, and it is the first reason never to read the second decimal ([[failure-probability]]).

::: astuce A typical session, in six moves
1) Enter your situation, your pension and your audited real spending ([[how-much-you-need]]). 2) Read the strip: the spread of ruin figures, column by column. 3) Click Broad-sample and look at §02 and §03: does your plan survive the real disasters? 4) Switch on the spending rule you are considering and read §04: is the standard of living it delivers acceptable in the bad quarter? 5) Open §09: what would it cost to bring ruin under your threshold, and which lever is cheapest for you? 6) Write down the configuration you settled on, and its thresholds, in your written plan ([[building-your-plan]]). Honest timing: an hour the first time, ten minutes at the annual review ([[the-annual-review]]).
:::

::: attention The three classic misuses
Pushing μ up because your fund did better: a fund's short history is a favorable window, not an expectation ([[simulator-traps]]), and the blend toward the prior exists precisely to stop that. Looking only at the green column: a plan that is acceptable only in the optimistic model is not acceptable. And mistaking decimals for signal: read the **gaps**, between columns, between scenarios, between levers, and never the second decimal ([[failure-probability]]).
:::

## What the page does not do

Several things are deliberately out of scope. No forecast: no model here predicts anything, they bracket what is possible. No account-by-account tax detail: one blended rate, which you calibrate yourself. No illiquid assets: a rental property is modeled as side income ([[real-estate-in-retirement]]). And no advice: the tool explores assumptions, the decisions stay yours. The models, the data feeding them and the methodological caveats are spelled out in the two fold-out panels at the bottom of the page, and at length in [[under-the-hood]].

## The essentials

- Two modes: the parametric sandbox, or the mode calibrated on your own funds, which is the only one that unlocks the data columns. Everything is in real euros, and spending is net of tax.
- The table at the top is a selector: click a column and the whole page recomputes under that lens. Plan between the central case and the broad sample.
- Reading order: the strip (the spread), §02 and §03 (the real disasters, and the sequence), §04 (what flexibility costs in lived spending), §09 (the price of each margin).
- The most powerful levers, in the usual order: spending and how flexible it is, the pension and early side income, and only then the portfolio.
- The market parameters come pre-filled and are deliberately pulled toward caution. The broad-sample and CAPE anchors are guardrails, not exotic options.

---

## Going further

- The two fold-out panels on the page itself: "How this machine works" (every control, every model) and "Method & honest caveats".
- The long version of the plumbing ([[under-the-hood]]), the families of models ([[historical-vs-parametric]]) and how to read the fans ([[reading-a-fan-chart]]).
- pofo's **readme**, section "Decumulation / FIRE analysis", for the command-line options.
