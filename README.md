# pofo

Go tool to visualize and compare investment portfolios over time, plus
reusable libraries to fetch price histories, compute risk/return metrics
and produce SVG charts.

A public instance runs at [pofo.zouh.org](https://pofo.zouh.org): the
[portfolio visualizer](https://pofo.zouh.org/visualizer), a
[FIRE simulator](https://pofo.zouh.org/firesimulator/), and a book on
living off your capital in retirement, in two editions:
[Le FIRE tranquille](https://pofo.zouh.org/firebook/fr/) (French) and
[The Quiet FIRE](https://pofo.zouh.org/firebook/en/) (English), both also
downloadable as EPUB. No ads, no accounts, no tracking cookies.

The CLI reads allocation files, downloads price histories (Yahoo Finance,
Financial Times, Morningstar, Stooq, ECB), rebuilds the missing past (proxies and
simulated data), simulates each portfolio with periodic rebalancing and
generates a self-contained HTML report opened in the browser (comparison and
statistics front and center; per-portfolio sections collapsed, each with a
performance curve, a realized-contribution timeline (who carried the trailing
12 months, stacked around zero under a macro-regime strip), four look-through
composition pies (geography, currency exposure, equity sectors, asset type
with stacked funds opened into their legs), its macro-regime coverage with
each bar split by contributing holding, a risk budget showing what share of
the variance each asset class actually carries next to its share of capital
(the two commonly differ by a factor of three: a "balanced" allocation is
usually balanced in capital only), and the per-regime realized contributions
that mirror it; hover any chart for exact figures).

## Usage

```sh
go build ./cmd/pofo                       # self-contained binary (datasets embedded)
./pofo my-portfolio.txt other.txt         # HTML report in /tmp + open
./pofo -assets WPEA,NTSG,CSPX             # compare individual assets (100% each)
./pofo -cli -assets VOO,IWDA              # quick check in the terminal
./pofo -b -assets AVWS,ZPRV               # same, with every history backcast
./pofo -warmup                            # pre-warm the catalog cache
./pofo -gen-simdata                       # regenerate pkg/datasets/simdata (then rebuild)
./pofo -export-epub le-fire-tranquille.epub  # export the FIRE book as EPUB 3 (-book-lang en for English)
```

The binary can be installed anywhere: simulated histories and reference
series are **embedded at build time** (`go:embed` of `pkg/datasets/`), and the
quote cache lives in the standard user cache directory
(`~/Library/Caches/pofo` on macOS, `~/.cache/pofo` on Linux).

The `-assets` option treats each identifier as a portfolio invested 100% in
it, handy for comparing ETFs against each other without writing a file. It
can be combined with portfolio files. Add `-simulate` (`-b`) to backcast every
identifier of the run, so `-b -assets AVWS,ZPRV` says what
`-assets AVWSSIM,ZPRVSIM` says without suffixing each one.

## Portfolio file format

One line per asset: `<weight in %> <identifier> [TER in %/year]`. Everything
after a `#` is a comment, and nothing else may follow the optional fee
column; blank lines are ignored. The portfolio name is the file name without
its extension.

```
# Description, links, notes…
#meta rebalance:30   # directive: this portfolio rebalances every 30 days
#meta extra-fees:0.5 # wrapper/mandate fees, applied on top of the whole portfolio
60   VTI            # US equities
25,5 IE00B4L5Y983   # ISIN resolved automatically (decimal comma accepted)
14.5 GOLD           # built-in alias → gold XAU/USD
```

`#meta key:value` lines carry per-portfolio directives:

| Directive | Effect |
|---|---|
| `rebalance:N` | rebalance back to the target weights every N days (`0` = never) |
| `sim:on` | backcast **every** holding, as if each identifier carried the `SIM` suffix, so a file need not spell it out (and stays a one-liner to de-suffix when sharing). Identifiers already ending in `SIM` are untouched, and a holding with no simulated history falls back to its real quotes. `-no-simulate` overrides it (real data only) |
| `extra-fees:X` (or `envelope-fees:X`) | extra fees in %/yr on the **whole** portfolio (insurance/pension wrapper, mandate, broker), on top of each asset's TER. Unlike TERs (already in the quotes) they are **deducted** from the simulation |
| `leverage:on` | keep the written weights (sum up to 500%); the residual `100−sum` is a cash position, earning the short rate (`^IRX`) if positive, **financed** if negative. A NAV reaching zero is a ruin |
| `borrow-spread:X` | financing spread in %/yr over cash when the cash position is negative (default `1.0`) |
| `capital:X` | starting amount (e.g. `10000`); required for flows, and unlocks the money rows |
| `contribute:A/P` | add a fixed amount A every period P ∈ {`week`, `month`, `quarter`, `year`} |
| `withdraw:A/P` | take out A, or `A%` of the current value, every period P |
| `optimize:OBJ[,constraints]` | let the optimizer choose the weights for objective `OBJ` (table below), under optional weight bounds, feasibility limits and a fitting window (all below) |
| `currencies:USD,EUR` | evaluate the portfolio in several base currencies at once; each becomes a comparison column with its own numeraire and CPI deflator (nominal **and** real stats), so the gap between columns is the currency risk. Cannot be combined with `optimize` |

Flows are invested or sold pro rata on the first trading day of each new
period. Statistics and comparison charts stay on a **time-weighted index**
(flows don't distort returns), while the money rows (starting capital,
contributed/withdrawn, final value and a **money-weighted IRR**) follow the
actual cash. Withdrawing a depleted portfolio is a ruin, flagged in the report.

**`optimize:` objectives.** The report shows `name (as written)` beside
`name (OBJ)`, with the computed weights and their in-sample stats as a note
(fitted on the past, so a starting point, not a promise). `optimize` cannot be
combined with `leverage`.

| `OBJ` | Long-only weights that… |
|---|---|
| `max-sharpe` | maximize return / volatility (the tangency portfolio) |
| `min-volatility` | minimize variance |
| `risk-parity` | equalize each asset's risk contribution (ignores `max-weight`) |
| `max-sortino` | maximize return / downside deviation (rewards non-correlation & positive skew) |
| `return-to-drawdown` | maximize return / max drawdown (Calmar-style) |
| `min-ulcer` | minimize the Ulcer Index (depth **and** duration underwater) |
| `max-worst-5y` | maximize the worst rolling 5-year return |
| `max-return` | maximize the CAGR. Degenerate alone (everything goes to the single fastest-compounding line): pair it with a limit, where it becomes the frontier point |
| `cwarp` | maximize CWARP vs the `-benchmark` (a diversifier selector, see [CWARP](#cwarp)) |
| `black-litterman` | start from the returns the file's OWN weights imply, revise them with your views, and maximize the utility of the result (see [Black-Litterman](#black-litterman)) |

**Constraints.** They follow the objective, comma-separated, and compose:

| Constraint | Effect |
|---|---|
| `max-weight:25` | cap every line at 25 % (ignored by `risk-parity`) |
| `min-weight:5` | floor every line at 5 %, so the search cannot drop a sleeve it dislikes in sample |
| `bounds:NTSG:15-30` | a range for ONE line, repeatable; either end may be omitted (`bounds:GDE:-25`). An identifier matching no holding is an error, not a silent no-op |
| `max-vol:9.5` | volatility cap, %/yr |
| `min-return:10.5` | CAGR floor, %/yr (strictly positive) |
| `max-drawdown:20` | drawdown budget, % |
| `train:..2015` | fit the weights on that window ONLY (`START..END`, each end a year, a `YYYY-MM-DD` date, or empty) |
| `view:IGLN:2@60` | a belief, for `black-litterman` only: IGLN earns 2 %/yr, at 60 % confidence (50 by default). `view:A>B:3@70` says A beats B by 3 points. Repeatable |
| `prior-return:4.6` | for `black-litterman` only: the return you expect from the written weights as a whole, which sets the scale everything else is read against |

The three limits do not combine with `risk-parity` or `cwarp`, whose solvers
cannot enforce them (and `cwarp` takes no `min-weight` or `bounds` either, only
`max-weight`): those combinations are rejected when the file is read, rather
than silently dropped.

The limits express what portfolio work usually asks: not "the best Sharpe" but
*the most return that stays under 9.5 % volatility*, or *the least volatility
that still compounds at 10.5 %*. They also route around a trap of this
toolkit's conventions: Sharpe here runs at a **zero** risk-free rate, so any
cash-like sleeve buys ratio for free (a short-linker sleeve can lift a
portfolio's measured Sharpe while costing two points of CAGR). A volatility cap
asks the intended question. When no allocation can meet the limits, the note
says so instead of returning a plausible-looking answer.

Bounds are not decoration: an unconstrained optimum is a corner solution, and a
corner fitted on one window is the least durable thing an optimizer produces.
The ranges an owner can defend line by line, for reasons the backtest cannot
see, belong in the problem.

```
#meta optimize:max-return,max-vol:9.5,min-weight:5,bounds:NTSG:10-30,bounds:GDE:10-25
```

**`train:`, fitting on data and judging on other data.** With a training window,
the optimizer sees only that slice while the report measures the resulting
weights over the WHOLE window, so the column you read is out of sample. The
note carries both halves, and the gap is usually the story:

```
$ pofo -cli hydra.txt          # with optimize:max-return,max-vol:9.5,train:..2015
hydra (max-return): weights computed by the optimizer (max-return under vol ≤ 9.5 %)
over 1996-03-27→2015-12-31: NTSG 10.0 %, ZPRV 15.5 %, MFEH 53.0 %, RAEF 5.0 %,
GDE 10.5 %, ZROZ 5.9 %, in-sample CAGR 11.5 %/yr, volatility 9.5 %, deepest
drawdown -12.5 %, Sharpe 1.19; over 2016-01-01→2026-08-07, which it did not see,
CAGR 8.3 %/yr, volatility 10.0 %, deepest drawdown -20.1 %
```

Fitted, it promised 11.5 %/yr inside a 9.5 % volatility budget; on the eleven
years it had never seen it delivered 8.3 %/yr, 10.0 % volatility and a
drawdown 8 points deeper than the one it was fitted to avoid. An optimizer
without `train:` never shows you that.

For **decumulation**, where long underwater stretches or a bad five-year run are
hard to live through, `min-ulcer` and `max-worst-5y` target that discomfort
directly; read the effect off the report's *Ulcer Index*, *TTR* and *Worst
rolling 5y CAGR* rows. The longest-recovery TTR itself is a step function of the
weights, so it is not an objective; `min-ulcer` is its smooth, optimizable
proxy. A decumulation file might read:

```
#meta capital:500000
#meta withdraw:4%/year
#meta optimize:min-ulcer,max-weight:40
40 NTSGSIM     # equity + duration engine
25 XAUUSDSIM   # gold
20 ZROZSIM     # long-duration deflation hedge
15 DBMFESIM    # managed-futures trend
```

### Black-Litterman

Every objective above that uses expected returns reads them off the sample,
which is exactly the quantity a price history estimates worst, and the answer
is a corner built on whichever line was lucky. `black-litterman` keeps the
expected returns in the problem but anchors them on the portfolio you already
wrote.

It runs in three steps. **Reverse optimization** turns the weights in the file
into the returns those weights implicitly expect, line by line: a 15 % gold
line in a defensive book can imply 5.6 %/yr, which is a number worth seeing on
its own. Your **views** then revise them, each carrying a confidence. And the
weights maximize the utility of the result, over the same bounds and limits
every other objective accepts.

The prior is the FILE, not a market-capitalization portfolio: a
managed-futures fund, a stacked 90/60, an inflation-linked sleeve and gold
have no capitalization weight, so the canonical anchor does not exist for the
books this tool is built for. The allocation you already defend does.

```
#meta optimize:black-litterman,prior-return:4.6,view:IGLN:2,view:DBMFE>DTLA:3@70,bounds:IGLN:10-20
```

`view:ID:Q@C` says *ID earns Q percent a year*, `view:ID>ID2:Q@C` says *ID
beats ID2 by Q points a year*, and `C` is your confidence as a percentage
(50 by default, which is He and Litterman's own choice; near 0 switches the
view off, near 100 makes it a certainty). `Q` may be negative or zero, both of
which are statements. `prior-return:R` is the return you expect from the
written weights as a whole and fixes the scale the views are read against;
without it the written weights are assumed to earn a Sharpe of 0.4.

With **no view at all**, the objective returns your own weights back, exactly,
and the note reports the returns they imply. That is not a bug: reverse
optimization and optimization are inverses, and the answer to "what do my
weights already assume" is often the more useful half.

Two things to keep in mind. This toolkit runs at a **zero risk-free rate**, so
a 3 %/yr view on a 3 %-volatility sleeve reads as a Sharpe of 1.0 and the
weights follow: state views as excess returns over cash when that is what you
mean (the note flags a view on a cash-like line). And what the model prices is
the **mean**: a line held for its crisis covariance, or against a liability,
has a reason the model cannot read and will be sized by its view alone. That
is useful, since it puts a number on what the hedge costs in expected return,
as long as you read it that way.

### CWARP

**CWARP** (Cole Wins Above Replacement Portfolio, Artemis Capital Management's
"Moneyball for Modern Portfolio Theory", 2020) scores whether an asset improves
a pre-existing *replacement* portfolio when layered on top at 25% of notional,
financed by borrowing. It is the geometric average of the improvements the
overlay makes to the replacement's Sortino ratio and return-to-maximum-drawdown,
minus one, in percent: **positive helps, negative hurts**. Unlike the Sharpe
ratio it rewards non-correlation and skew, because both denominators are
measured on the *combined* series.

The replacement is the report benchmark (`-benchmark`, default `^GSPC` = equity
beta, the paper's standard; point it at a 60/40 series for that variant). CWARP
appears in the report in two forms:

- a **CWARP row** in the statistics table, scoring the whole portfolio as a 25%
  overlay on the benchmark;
- a **per-holding CWARP column** in each portfolio's asset table, scoring each
  sleeve on its own, so you can see which holdings actually diversify equity
  beta (typically gold, long duration and trend, not more equity).

`#meta optimize:cwarp` maximizes it directly, choosing the weights whose blend
best improves the benchmark; the achieved score is reported in the optimizer
note. The objective is non-convex, so the solver is a multi-start heuristic and
its weights are a good allocation rather than a certified optimum.

**Read the result correctly.** `optimize:cwarp` finds the best *diversifier of
the benchmark* (the ideal satellite sleeve to layer on top of equity beta), not
a good standalone portfolio. It therefore loads exactly the assets that look
weak on their own but shine in combination: gold, long-duration bonds, trend.
The report's CAGR / volatility / drawdown columns describe that sleeve **on its
own**, so they will almost always look worse than your written portfolio, by
design; that is the whole point of the Sharpe critique. The value is not in the
sleeve's standalone stats but in the `+CWARP` it adds when overlaid on the
benchmark. Use `optimize:cwarp` when you already hold equity beta and want the
best diversifying satellite; use `max-sharpe` / `min-volatility` for a complete
standalone allocation.

Without `#meta leverage:on`, a weight > 100% is rejected (with a hint) and sums
≠ 100% are normalized as before.
An optional third column declares an asset's TER
(e.g. `60 VOO 0.03`); otherwise it is fetched automatically (FT, justETF)
and cached for 6 months; `-no-fees` disables that lookup. The report shows
per-asset fees and a "Weighted ongoing fees" row in the statistics table.

The identifier can be a US ticker (`VTI`), a European ticker from the
embedded list (`IWDA`, `CSPX`, `CW8`…), an ISIN, or a built-in alias
(`GOLD`, `WTI`, `BHMG`, `AMUNDI-VOLATILITY`, `WINTON-TREND-EQUITY`…). If the
weights do not sum to 100, they are normalized with a warning.

**SIM convention**: a bare identifier (`DBMF`, `NTSG`, `VOO`) uses only the
asset's real quotes; the history starts at its inception date. The `SIM`
suffix (`DBMFSIM`, `NTSGSIM`, `VOOSIM`…) additionally allows extending the
uncovered period, via `pkg/datasets/simdata/` then the known proxies; real
quotes always keep priority wherever they exist. `-no-simulate` ignores SIM
suffixes globally. `#meta sim:on` applies the suffix to every holding of a
file at once, so you can drop it from each line (and add it back for the whole
file just as easily); `-simulate` (`-b`) is the same thing from the command
line, for every file and every `-assets` identifier of the run at once. Both
only ever turn the backcast on: an asset with no simulated history keeps its
real quotes, with a note in the report.

## What each weight buys: `-sweep`

`pofo -sweep portfolio.txt` takes one holding at a time, moves its weight
across a grid (0 to 45 % in 5-point steps, `-sweep-step` to refine), keeps the
other lines' **relative proportions**, and re-runs the real simulation at every
point:

```
$ pofo -sweep examples/hydra-five-engines-capital-efficient.txt

NTSG (written 18.0 %)
    weight      CAGR       vol  Sharpe     maxDD     TTR   worst5y
     0.0 %   10.96 %    9.83 %    1.07  -17.40 %   1.4 y    0.72 %
    10.0 %   10.86 %    9.79 %    1.07  -18.57 %   1.5 y    0.88 %
    18.0 %   10.77 %    9.89 %    1.05  -19.50 %   2.0 y    1.01 %  <- written
    30.0 %   10.61 %   10.28 %    1.00  -20.89 %   2.3 y    1.18 %
    45.0 %   10.39 %   11.10 %    0.92  -23.83 %   2.8 y    0.80 %
```

Read down a column and the sleeve's job becomes explicit: here more world
equity costs CAGR and Sharpe while **buying** the five-year floor, which is an
argument the backtest cannot make on its own but can price. Read across the
sleeves and you get the per-line "sane ranges" a portfolio file should be
documenting: where the worst five-year stretch peaks, where the drawdown starts
deepening, and how flat the answer is in between (usually flatter than the
optimizer's precision suggests).

Every grid point is measured over the same window, since a holding at 0 %
stays in the portfolio and keeps binding it, so the rows are comparable. The
simulation is the real one, with the file's own `rebalance:` period, fees and
flows: the question is what would have happened had a different number been
written in the file.

## Suggesting assets to add

`pofo -suggest portfolio.txt` analyses a portfolio's **macro-regime
coverage** and recommends catalog assets to add that fill the gaps. The four
regimes are the growth × inflation quadrants behind All-Weather- and
Dragon-style portfolios: `growth`, `deflation`, `inflation`, `crisis`, and
each catalog asset is mapped to the regimes it helps in from its factual tags
(asset class, strategy; see `pkg/datasets/assetmeta/`). A regime with little
weight is a gap.

It is **structure-first**: only assets that fill a gap are considered, and
each is then validated **out-of-sample**: the candidate is added at a
modest weight and a walk-forward checks that Sharpe and max-drawdown improve
*consistently across periods*, not in one lucky stretch. Because adding an
asset at a fixed weight fits nothing to the data, this measures robustness
rather than an over-fitted optimum. Suggestions are kept diverse (at most one
per asset class) and reported with the gap they fill, a suggested weight,
their correlation to the portfolio, and the out-of-sample win counts.

`-suggest` also flags **redundancies**: holdings that move almost
identically and share an asset class (three S&P 500 trackers are one bet, not
three). It prints to the terminal and exits, like `-verify-data`.

For a quick, **offline** read, `pofo -coverage portfolio.txt` shows the
same coverage chart and then, for each gap, lists the catalog assets that
fill it (grouped by asset class), with no price downloads, no ranking, just the
menu of options. Run `-suggest` afterwards to rank and validate them.

By default coverage is organized by the four **macro regimes**.
`-framework factors` switches to a **risk-factor** lens (market, size,
value, momentum, quality, term, credit, alternative, cash) for both
`-coverage` and `-suggest`. The factor mapping is coarser: this catalog
holds many diversifiers (gold, commodities, managed futures, volatility)
that are not Fama-French factors and all land in *alternative*, so the
regime view stays the default.

## Decumulation / FIRE analysis

`pofo -fire` opens a local web explorer that simulates a withdrawal
(retirement) phase and shows the **probability of ruin** as you drag sliders
for capital, spending floor, cash-buffer years, real return, volatility, tail
df, horizon, pension, spending rules and the French taxes (the rate, the share
of today's capital that is unrealised gain, and how much of it sits in a PEA or
an assurance-vie, the pockets being drained taxable account first). The dashboard
reads top to bottom as one argument: the same plan under every return model
(Student-t, sequence stress, JST broad-sample, lost decade), today's
valuation (live Shiller CAPE), the simulated wealth fans, the plan replayed
through the **worst retirements on record** (USA 1929/1966/2000, Japan 1990),
ruin decomposed by the **first decade's return** (sequence risk), the
delivered spending and its funding mix, mortality-crossed ruin
("alive, broke or gone"), the risk levers, and the buffer arbitrage. The
engine runs in Go (parallel Monte-Carlo); the page only renders.

`pofo -fire portfolio.txt` seeds the model from a real portfolio: it derives
the return assumptions and a historical real-return panel from the holdings
(reconstructed back via `SIM`, deflated by `^HICP-FR`), lets you switch
between a **parametric**, **historical bootstrap** or **historical-cohort**
projection, and drag each holding's weight to re-test ruin live. The two
historical models sample at **monthly** frequency (a stationary block
bootstrap and every actual start month), preserving intra-year regimes and
cross-asset correlations, then compound to the annual withdrawal cycle.
Everything is in real euros; the model is a fat-tailed hypothesis-exploration
tool, **not investment advice**.

The explorer also embeds **the FIRE book** ("Le FIRE tranquille") at
`/firebook/fr/` (a small link sits at the bottom of the "How this machine
works" fold): a French-language handbook of decumulation, withdrawal
strategies, resilient portfolios, buffers and French taxation, written as
cross-linked articles and served straight from the binary (`pkg/firebook`).
The complete English edition, "The Quiet FIRE", sits next to it at
`/firebook/en/`, and every page of one edition links to its counterpart in the
other.

The whole book is also downloadable as a single **EPUB 3** file for offline
reading: every mount exposes it at `le-fire-tranquille.epub` (a discreet
"Version EPUB" link sits on the book index, `the-quiet-fire.epub` on the
English one), and `pofo -export-epub le-fire-tranquille.epub` writes the same
file from the command line, `-book-lang en` the English edition.

The same mount also serves an **OPDS 1.2 catalog** at `opds.xml` (e.g.
`http://localhost:8080/firebook/fr/opds.xml`): add that URL once in an e-book
reader such as KOReader to download and later refresh the book in place, the
re-download overwriting the same file so any annotation sidecar is preserved.

Each edition publishes an **Atom feed** at `feed.xml` (e.g.
`http://localhost:8080/firebook/en/feed.xml`), one entry per article in reading
order, which a feed reader can follow instead of the index page.

The book is written to be found and to be quoted. Each page carries its own
title, description, canonical link, hreflang pair and schema.org data, and each
article is also served as **clean Markdown** at the same URL with a `.md`
suffix (`/firebook/en/what-is-fire.md`), source untouched, behind one comment
line naming the page to cite. The server publishes `/sitemap.xml`,
a `/robots.txt` that allows everything and names the AI crawlers one by one,
and an [`/llms.txt`](https://llmstxt.org) indexing both editions article by
article. Nothing here is behind a paywall: cite the page.

The reusable pieces live in the library: `pkg/scenario` (return-path
generation) and `pkg/decumul` (the withdrawal engine, FIRE outcome metrics
and sweeps), with the thin web layer under `pkg/decumul/web`; the book is its
own package, `pkg/firebook`, mountable by any server.

## Web app

`pofo -serve` starts the whole tool as one local web app, every surface on a
single port:

| URL | Surface |
|---|---|
| `/` | the **landing page**: the pofo mark, one sentence, and four cards linking the sections |
| `/visualizer` | the **portfolio visualizer**: compose portfolios or tick bundled examples and compare them |
| `/view` | the visualizer's report: the same HTML comparison the CLI writes, addressed by a shareable URL |
| `/firesimulator/` | the **FIRE simulator** (`-fire`, mounted under a prefix; old `/fire/` redirects here) |
| `/firebook/fr/` | the **FIRE book** ("Le FIRE tranquille"), with a small nav bar back to the other surfaces (old `/book/fr/` redirects here) |
| `/firebook/en/` | the English edition ("The Quiet FIRE"), cross-linked with the French one page by page |
| `/firebook/<lang>/<article>.md` | any article's **Markdown source**, served as written: the same URL as the page, plus `.md` |
| `/firebook/<lang>/feed.xml` | the edition's **Atom feed**: every article, in reading order, for a feed reader |
| `/firebook/<lang>/card.png` | the edition's **social card**, the 1200x630 image a shared link shows |
| `/sitemap.xml`, `/robots.txt`, `/llms.txt` | what crawlers and AI agents read: every page of both editions, everything allowed, and an [llms.txt](https://llmstxt.org) index of the whole site |
| `/healthz` | the **liveness probe** for an orchestrator or load balancer: `200 text/plain` `ok`, and the one route kept out of the access log so a probe every few seconds leaves no noise |

```sh
./pofo -serve                             # http://127.0.0.1:8787/
./pofo -serve -listen 127.0.0.1:9000      # a different port
./pofo -serve examples/dragon-decumulation-household.txt  # seed the FIRE panel from a file
```

**Identifiers a visitor may compose.** Everything the bundled catalog resolves
offline (ids, ISINs, aliases, embedded fund tickers, with the optional `SIM`
suffix) is free and unlimited. Anything else is fetched from the usual quote
sources, under two guards: it must *look* like an instrument (a valid ISIN,
check digit included, or a plausible exchange ticker such as `DGRO` or
`IWDA.AS`; no quote symbol, no punctuation), and it is rationed by two rolling
hourly budgets, one per visitor and one for the whole process:

```sh
./pofo -serve -serve-foreign-per-hour 10 -serve-foreign-global-per-hour 60  # the defaults
./pofo -serve -serve-foreign-per-hour 0    # catalog only, nothing else is fetched
```

Only identifiers that would really cost an upstream request are charged: one
already in the quote cache is free, so re-running a link costs nothing. A spent
budget answers `429` and says so. Such an identifier is resolved **exactly**
(no fuzzy name search), so a typo fails instead of quoting an unrelated fund;
the CLI keeps the fuzzy search, its user being able to read the resolution line.
A well-formed identifier no source quotes (`ZQXWVT` is a perfectly good ticker
shape) answers `404` and names it, since no retry would help; an upstream
outage on a real identifier stays a `500` with the detail in the server log.

A public deploy can also **push** its URLs to the search engines that speak
[IndexNow](https://www.indexnow.org/) instead of waiting to be crawled again.
Mint one unguessable key per host (8 to 128 letters, digits and dashes), serve
it, and submit after each deploy:

```sh
./pofo -serve -indexnow-key <key>                        # publishes /<key>.txt
./pofo -indexnow https://example.org -indexnow-key <key> # push the URL list
```

The list is the sitemap's, so what is pushed and what is crawlable cannot
disagree. Without `-indexnow-key` the feature is off, and nothing is ever
submitted by a running server: `-indexnow` is the only thing here that talks to
a search engine, and only when you run it.

A public deploy can likewise count its readers with
[Cloudflare Web Analytics](https://developers.cloudflare.com/web-analytics/),
which sets no cookie and stores nothing on the visitor's device:

```sh
./pofo -serve -cf-beacon-token <token>   # or POFO_CF_BEACON_TOKEN in the environment
```

The beacon then rides every HTML page the server emits: the landing page, the
visualizer, every report, both editions of the book, every simulator mount. Both
the flag and the variable empty, the feature is entirely off and the pages are
byte for byte what they were. There is no other analytics of any kind in the
program.

`-listen` defaults to `127.0.0.1:8787` (loopback only). Portfolio file
arguments feed the FIRE simulator's historical models, exactly as they do
for `-fire`.

The visualizer is driven entirely by its query string, so a comparison is a
link you can bookmark or share:

```
/view?ex=dragon-decumulation-household&ex=golden-butterfly
/view?p=NTSG:60,IGLN:20,IBCI:20!sim:on&currency=EUR
```

`ex=` names a bundled example (repeat it to stack several). `p=` is an ad-hoc
portfolio, `ID:WEIGHT` pairs comma-separated, with `!key:value` meta
directives appended (`!` replaces the file format's `;`, which a query string
cannot carry). Global options (`start`, `end`, `rebalance`, `currency`,
`bench`, `sim`) mirror the CLI flags. `currency` takes a three-letter ISO code
or `native` (keep each series in its own currency); `bench` takes an
identifier from the local catalog, the default `^GSPC`, or empty to drop Beta
(anything else is rejected before any network call). Up to six portfolios per
page, twenty holdings each.

`p=` identifiers are **catalog-only**: the tool resolves them from the
embedded catalog (ids, ISINs, aliases, bundled fund tickers, the `SIM` suffix
allowed) and never fetches an arbitrary or unknown identifier on behalf of an
anonymous visitor, so a raw quote symbol outside the catalog is rejected
before any network call. `ex=` files carry no such limit; they are the vetted
builds shipped in the binary.

Every `/view` report also carries a live **composer**: an in-page editor over
this same grammar. Add or remove holdings and portfolios, edit weights (with a
sum badge and one-click normalize), name each id from the catalog as you type,
and Fork an example into an editable copy; the URL rewrites itself as you go so
the link is always the current state, and Run renders it. It stays within the
same catalog gate and caps, so nothing you can compose escapes them.

Each portfolio in a `/view` report carries a **Simulate** link that opens the
FIRE simulator bound to that portfolio (`/firesimulator/p/<spec>/` for an ad-hoc
composition, `/firesimulator/e/<name>/` for an example). The visualizer's home
page also remembers your
default currency, rebalance and sim settings in a cookie, so it opens where you
left it; a `/view` link stays self-contained and reproduces the same report for
anyone, cookie or not.

Everything runs on the machine that started it, and the default bind is
loopback. To reach the app from your phone or another device, put it behind
your tailnet instead of opening a port:

```sh
tailscale serve 8787       # https://<machine>.<tailnet>.ts.net/ , private to your tailnet
```

## Main options

| Option | Default | Description |
|---|---|---|
| `-out` | `/tmp/pofo-<timestamp>.html` | generated HTML file |
| `-data` | standard user cache | quote cache (JSON) |
| `-simdata` | embedded in the binary | source of simulated histories (directory for dev) |
| `-rebalance` | `90` | rebalance every N calendar days (0 = never) |
| `-start`, `-end` | earliest / last available quote | the analysis window (`YYYY-MM-DD`) |
| `-benchmark` | `^GSPC` | reference for Beta, capture ratios and the CWARP replacement |
| `-currency` | `EUR` | convert every series (and the benchmark) to this currency; empty disables |
| `-cache-age` | `720h` (1 month) | cache freshness before re-downloading; the data generators (`-gen-simdata`, `-verify-simdata`) default to `24h` instead, since what they write ships inside the binary |
| `-assets`, `-a` | | list `A,B,C`: each asset compared as a 100% portfolio |
| `-simulate`, `-b` | | backcast every identifier of the run, as if each carried the `SIM` suffix |
| `-cli` | | curves and summary table in the terminal, no HTML |
| `-width` | `$COLUMNS` or 100 | width of the `-cli` chart (wider = more granularity) |
| `-warmup` | | pre-warm the built-in asset catalog then exit |
| `-verify-data` | | data doctor: check the referenced assets' quotes (or the whole catalog, `make verify-catalog`) for anomalies. Series hygiene (bad points, gaps, stale feeds, each judged against the pace the series kept at the time), plus, for a catalogued asset, plausibility against its class's volatility/CAGR/move/drawdown band and identity against its record (served currency, share class, inception). Prints a summary and exits |
| `-verify-simdata` | | reconstruction quality report: replay every recipe's engine (or the ones named as arguments) against the real quotes, write an HTML page and open it, then exit |
| `-suggest` | | recommend catalog assets to add for better regime coverage, flag redundant holdings, then exit |
| `-coverage` | | offline advisor: show which regimes/factors a portfolio misses and the catalog assets that fill them, then exit |
| `-sweep` | | per-holding weight sweep: what each line's weight buys and costs, then exit |
| `-sweep-step` | `5` | grid step, in weight percent, for `-sweep` |
| `-fire` | | open the local decumulation/FIRE explorer (sliders, ruin curves), optionally for a portfolio file, then serve until stopped |
| `-serve` | | serve the whole web app (hub, visualizer, FIRE simulator, book) on one port until stopped |
| `-export-epub` | | write one edition of the FIRE book to the given path as an EPUB 3 file, then exit |
| `-book-lang` | `fr` | with `-export-epub`: which edition to write, `fr` ("Le FIRE tranquille") or `en` ("The Quiet FIRE") |
| `-listen` | `127.0.0.1:8787` | listen address for `-serve` (loopback by default) |
| `-serve-foreign-per-hour` | `10` | with `-serve`: how many identifiers outside the bundled catalog one client may have fetched from the quote sources per hour (`0` = catalog only) |
| `-serve-foreign-global-per-hour` | `60` | with `-serve`: the same budget for the whole process, all clients together |
| `-indexnow-key` | | IndexNow ownership key: with `-serve`, publish it at `/<key>.txt`; with `-indexnow`, sign the submission (empty = off) |
| `-indexnow` | | submit every published URL of the given origin to the IndexNow search engines, then exit |
| `-cf-beacon-token` | | Cloudflare Web Analytics site token: with `-serve` or `-fire`, put the cookieless beacon on every HTML page (empty = off) |
| `-framework` | `regimes` | classification for coverage and `-suggest`: `regimes` (macro quadrants) or `factors` (risk factors) |
| `-no-open`, `-no-simulate` | | do not open the browser / ignore SIM suffixes (overrides `-simulate`) |

## Data

- **Resolution**: aliases → embedded ticker→ISIN list (European ETFs/funds)
  → built-in catalog of pinned resolutions → multi-source search
  (Yahoo, FT, Morningstar via Boursorama), the deepest series winning.
- **Inflation**: `^HICP-FR` (and `^HICP-<geo>`, e.g. `^HICP-EA`) fetch the
  Eurostat Harmonised Index of Consumer Prices (monthly; the French series is
  extended back to 1955 via the OECD CPI), interpolated to a smooth daily
  curve. `^CPI-US` is the dollar sibling: the US CPI-U from FRED, monthly
  since 1913. They chart like any asset: the CAGR reads as average inflation,
  drawdowns mark deflation episodes; they also serve as the deflator for
  real-return analysis (HICP for EUR reports, CPI-US for USD ones). Monthly
  snapshots are embedded in the binary as offline fallbacks, so the series
  are available even if Eurostat or FRED is down.
- **Currency**: every series is converted to the `-currency` (default EUR)
  using daily FX crosses (Yahoo, with Stooq then the ECB reference rates as
  fallbacks), so USD ETFs and EUR funds compare fairly; the euro crosses
  reach back to 1971 via a bundled daily ECU/DM/EUR proxy, the earliest known
  rate is held flat before the FX history starts (with a warning), and
  unconverted (unknown-currency) assets are flagged. For library consumers,
  `Client.ConvertCurrency` reprices any `Series` into a target currency via
  the same crosses.
- **Cache**: 1 month by default (`-cache-age`), a day for the data
  generators, whose output ships in the binary; a failed refresh **serves the
  stale data** with a stderr warning (charts may stop before today), and never
  deletes anything.
- **History extension** (`…SIM` identifiers only): first the
  `pkg/datasets/simdata/` files (below), otherwise a known total-return proxy
  (VOO→SP500, IWM→^RUTTR, BND→VBMFX, …), converted into the asset's quote
  currency and rescaled to its first real quote. The report flags every
  simulated portion.

### Special identifiers

Beyond tickers, ISINs and catalog aliases, a few special names work wherever
an asset identifier does (portfolio files, `-assets`, `-verify-data`, the
library `Fetch`):

| Identifier | Series | History |
|---|---|---|
| `^GSPC` | S&P 500 price index | 1927→ |
| `^NDX`, `^DJI`, `^IXIC` | Nasdaq-100, Dow Jones, Nasdaq Composite | |
| `^VIX` | CBOE Volatility Index (implied vol, percent points) | 1990→, bundled |
| `^IRX`, `^FVX`, `^TNX`, `^TYX` | US Treasury yields: 13-week, 5, 10, 30-year | |
| `^ESTR`, `^EONIA`, `^EURIBOR3M` | euro benchmark rates: ESTR overnight, its discontinued predecessor EONIA, Euribor 3-month (monthly) | 2019→ / 1999-2021 / 1994→ |
| `^ECB-DFR`, `^ECB-MRO` | ECB policy rates: deposit facility, main refinancing | 1999→ |
| `^SOFR`, `^FEDFUNDS`, `^FED-TARGET` | US benchmark and policy rates: SOFR, effective funds rate, FOMC target upper bound | 2018→ / 1954→ / 2008→ |
| `^HICP-FR`, `^HICP-<geo>` | Eurostat inflation index (all-items HICP) | 1955→ (FR), bundled |
| `^CPI-US` | US CPI-U inflation index (FRED) | 1913→, bundled |
| `USDEUR=X`, any `<AAA><BBB>=X` | FX cross, quoted in the second currency | 1971→ (euro crosses), bundled |
| `XAUUSD` (alias `GOLD`), `XAGUSD` | gold / silver spot (via futures) | `GOLDSIM` reaches 1968 |
| `CL=F` | WTI crude oil continuous futures | |

For example, `pofo -cli -assets '^VIX'` charts the VIX, and
`pofo -assets 'USDEUR=X,^CPI-US'` compares the dollar and US inflation.

The rate symbols are LEVELS in annualized percent, not prices, and the euro
ones went negative for eight years, so the portfolio machinery (`-assets`,
which builds a 100 % position and computes returns) is meaningless on them.
Chart them with the dedicated mode instead, which draws the levels and
summarises last/min/max/average per symbol:

```sh
pofo -rates ^ESTR,^EURIBOR3M,^ECB-DFR -start 2015-01-01
pofo -rates ^SOFR,^FED-TARGET,^IRX          # the dollar side
pofo -rates list                            # what is available
```

"Bundled" means the binary embeds the history (full for `^VIX`, daily for
the euro crosses, monthly anchors otherwise) and serves it as a last resort,
so these chart offline.

Mind the units: the yields, `^VIX` and the inflation indices are LEVELS, not
prices. They chart fine and their long histories make good regime context,
but they are not investable returns: keep them out of weighted portfolios
(statistics computed on them read as nonsense).

The long MSCI World history is not a symbol of its own: it backs the `SIM`
extensions, so chart `IWDASIM` or `URTHSIM` to see MSCI World back to 1969
(the bundled series is the net total return in USD).

### Intraday

`Client.Intraday` fetches the current trading day's price path (5-minute
resolution) from Yahoo Finance. The call is live and stateless: the client
performs no intraday caching, so the caller is responsible for throttling
and storing results when needed. If the identifier does not resolve to a
Yahoo symbol (for example, a fund quoted only by FT or Morningstar),
the call returns `ErrNotCovered`; test with `errors.Is`.

One exception: a fund priced once a day and published with a lag (a French
employee-savings fund such as `ERESMONDEM`, served from its official NAV feed)
whose catalog record names a `nowcast_proxy` gets an ESTIMATED path instead,
its last daily value scaled by the proxy's intraday move converted tick by
tick (`IntradaySeries.Estimate` is true, `Proxy` names the proxy). The same
proxy carries the fund's daily series past its last published NAV
(`Series.EstimatedFrom` marks the tail, `WithoutEstimates` strips it); the
estimate is never cached and never enters a bundled dataset. Both estimates
start from the proxy print the last NAV was struck on, which the record names:
its close of that day, or its opening price when the record says
`nowcast_anchor: "open"` (a fund whose valuation rules price its holding at the
opening of the valuation day, such as the single-stock FCPE `ERES_DATADOG`),
falling back on the close when that day has no opening price.

Mapping the result to a chart series is caller-side:

```go
s, err := client.Intraday("VOO")
if err != nil {
	// errors.Is(err, marketdata.ErrNotCovered) means no intraday for this asset
}
ser := chart.Series{Name: s.Name}
for _, p := range s.Points {
	ser.Dates = append(ser.Dates, p.Time)
	ser.Values = append(ser.Values, p.Close)
}
svg := chart.Line(chart.Options{Title: s.Name}, []chart.Series{ser})
```

### Latest quote

`Client.Latest` returns the most recent price of an instrument as a `Quote`,
for a live portfolio valuation. A Yahoo-quoted instrument yields its live
market price (`Quote.Live == true`); any other instrument (an FT or Morningstar
fund, whose last NAV close is its latest price) yields its last daily close
(`Quote.Live == false`); a fund with a `nowcast_proxy` yields the last tick of
its estimated intraday path (`Quote.Live == true`, `Quote.Source == "nowcast"`),
or its last published NAV when the proxy is unreachable. When Yahoo is down or throttled, the call degrades
instead of failing: retries and a second Yahoo host first, then the daily-close
path with its Stooq/ECB/FT/Morningstar fallbacks and, last, the stale on-disk
cache, so it answers for every asset and even offline.

```go
q, err := client.Latest(ctx, "VWCE")
if err != nil {
	// no usable quote for this identifier
}
rate, _ := client.FXRate(ctx, q.Currency, "EUR", q.Time)
value := shares * q.Price * rate // valuation in EUR
_ = q.Live                       // true: real-time; false: last daily close (q.Time)
_ = q.Session                    // "regular", or "" when the source names no session
```

#### Extended hours (pre-market, after hours)

`Quote.Session` names the session the price was struck in, and `Quote.Time` is
always that print's instant. The default calls only ever report `"regular"`
(Yahoo's regular-market price, timed at the close once the market has shut) or
`""` (a daily close, a fund NAV, a nowcast: no session to name).

Ask for the off-hours prints of US venues with `QuoteOptions.ExtendedHours`
(`Client.LatestAny`) or `Client.LatestBatchExtended`. A pre-market or
after-hours print then wins whenever it is strictly more recent than the regular
session's last price, and `Quote.Session` reads `"pre"` or `"post"` so the
display can label it:

```go
for id, q := range client.LatestBatchExtended(ctx, []string{"DDOG", "VT"}) {
	fmt.Println(id, q.Price, q.Session, q.Time.Format("15:04")) // DDOG 225.70 post 19:59
}
```

Only US venues run these sessions: a European line, a fund NAV and a nowcast
answer exactly as they would without the opt-in, and so does everything when
the extended leg (Yahoo's cookie-authenticated quote API) is unavailable. Such
a print is thin and wide-spread: show it, do not book a valuation on it.

## Simulated data (pkg/datasets/simdata/)

Complex assets (90/60 funds, managed futures…) are rebuilt by `pkg/simgen`
from long-history building blocks, validated against their real quotes,
then stored as self-documenting CSVs (method, validation, date) in
`pkg/datasets/simdata/`:

```sh
./pofo -gen-simdata                   # regenerate everything (then make build to re-embed)
./pofo -gen-simdata -dry NTSX         # validate without writing
./pofo -verify-simdata                # how good is every engine? HTML report, opened
./pofo -verify-simdata CTA DBMF       # just these two
```

`-verify-simdata` is the standing answer to "can this backcast be trusted".
It replays each recipe's engine **without the real quotes it normally splices
in** and lays it over those quotes on the window where they exist, which is the
only window a reconstruction can be judged on. Each asset gets a curve pair
(engine against real, base 100), a cumulative engine/real drift panel, and two
separate verdicts: **level**, does the engine earn the asset's return (a hot
engine flatters every backtest downstream), and **path**, does it move with the
asset, on the monthly correlation and on the tracking error relative to the
asset's own volatility. Recipes built as donor chains also get a chain of
custody: every junction graded on its own overlap, since the card's own
statistics can only grade the nearest one. Worst first inside each family.

Every series is built **only from quotes the tool itself can fetch**
(Vanguard/Yahoo funds with decades of history, the `^IRX` cash rate, gold and
oil futures) combined by the in-house composite, TSMOM trend and donor-chain
engines; no third-party data is bundled. External index series are
used solely to cross-check quality during development, never shipped.

Bundled recipes and measured quality (daily / weekly correlation of returns
vs the real series; the real series is always grafted on top of the
simulation wherever it exists):

| Asset | Method (building blocks) | Validation (daily / weekly corr) |
|---|---|---|
| NTSX (UCITS) | 0.90×VFINX + 0.60×(VFITX−cash) + 0.10×cash (1991→) | 0.96 / 0.99 |
| NTSG (UCITS) | global 90/60: 0.90×MSCI World net TR (real IWDA 2009→, refdata 1969→) + 0.60×a four-currency government bond futures overlay (80 % US / 11 % German / 6 % Japanese / 3 % British, each a local excess return) + 0.10×cash | 0.92 / 0.98 |
| NTSZ (UCITS) | eurozone 90/60: 0.90×EZU(EUR) + 0.60×(euro govt−cash) + 0.10×cash, euro-native refdata (1986→) | 0.63 / 0.93 (fund launched 2025-09, short overlap; use monthly stats) |
| URTH (MSCI World) | 0.60×VFINX + 0.40×VTMGX (1999→) | 0.90 / 0.97 |
| IWDA (MSCI World) | 0.60×VFINX + 0.40×VTMGX (1999→) | 0.60 / 0.85 (GBP listing, short overlap) |
| VT (total world) | cap-weighted VFINX + VTMGX + VEIEX, weights pinned to the published 2009-10-31 FTSE All-World split (40.4/46.1/13.5) and drifting with the legs' own returns (1988→) | 0.98 / 0.99 |
| RSSB (100/100 stocks+bonds) | the same cap-weighted world equity sleeve + 1.0×(VFITX−cash) (1988→) | 0.95 / 0.99 |
| ZPRV (US small-cap value, UCITS) | DFSVX (DFA US Small Cap Value, 1993→), then the Ken French small-value factor less a measured 1.0 %/yr for its grossness (1963→), real grafted 2015 | 0.67 / 0.91 |
| AVWS (global small-cap value, UCITS) | 0.70×AVUV + 0.30×AVDV, the same manager's own sleeves (2019→), then the Dimensional pair (1994→), in EUR | 0.46 / 0.92 (EUR NAV against US-close donors: read the weekly figure, monthly 0.99) |
| SHY (1-3y Treasury) | VFISX short Treasury (1991→) | 0.81 / 0.89 |
| IEF (7-10y Treasury) | VFITX intermediate Treasury (1991→) | 0.95 / 0.96 |
| TLT (20+y Treasury) | VUSTX long Treasury (1986→) | 0.98 / 0.99 |
| ZROZ (25+y STRIPS) | a constant 27-year zero-coupon Treasury repriced off the bundled long Treasury yield, net of the fund's own fee (1953-04→) | 0.96 / 0.99 (monthly 0.99; a geared coupon fund cannot stand in for a strip, see `docs/long-treasury-zero-coupon-design.md`) |
| DBMF (managed futures) | the net all-styles composite it replicates, then real NAVs back to the deepest donor (1996-03→) | 0.68 / 0.75 |
| KMLM (managed futures) | real managed-futures NAVs, 14% target vol (1996-03→) | 0.63 / 0.65 |
| CTA (managed futures) | the net pure-trend composite, then the same deepest donor, 16% target vol (1996-03→) | 0.54 / 0.54 |
| RSST (100% stocks + trend) | VFINX + 1.0×(trend−cash) overlay on a net pure-trend reference (2000→) | 0.88 / 0.89 |
| RSBT (100% bonds + trend) | VFITX + 1.0×(trend−cash), same overlay (2000→) | 0.50 / 0.47 |
| Winton Trend-Equity (UCITS) | 0.60×VFINX + 0.40×VTMGX + 0.50×trend overlay (2000→) | 0.62 / 0.84 |

A fund's ongoing charge is deducted only where its donors do not already carry
one. A mutual-fund or ETF NAV arrives net of its own manager's fee, so charging
the target's whole TER on top of it would bill the backcast twice: what each
recipe charges is the difference, floored at zero, and it steps with the eras,
the whole charge falling on the deep segment where the leg is a fee-free index
or a constant-maturity reconstruction. An academic factor pays no fee at all
and owes a haircut instead (1.0 %/yr for the small-value one, measured).

Managed-futures correlations are modest: each fund runs a faster, partly
discretionary strategy that a single 12-month TSMOM rule only approximates.
Those daily correlations understate the fidelity, because a seven-market
engine cannot reproduce a fifty-market programme day by day. What a
reconstruction owes its user is the right MONTH: a bundled reference supplies
the month-to-month path (monthly agreement with the real funds 0.60 to 0.97)
and the engine supplies the daily texture. Both references are composites of
REAL programmes, net of their managers' fees, so each settles the level as well
as the path and nothing is levelled by hand: a diversified fund reads the
all-styles one, a trend overlay the pure-trend one. The two funds that exist to
replicate a published index take that index itself as their pre-inception donor,
which fits them better than any other manager's fund does. Each vol target is the
volatility the fund itself realized: DBMF 12.2 % against 12.4 % real, KMLM 14.8
against 14.7, CTA 16.7 against 16.9, AQR 9.5 against 9.3. Each donor segment
also carries the fund's fee load rather than its own, lifted by the difference
between the two published fee tables (the donors are 1.3 to 2.7 %/yr vehicles
standing in for 0.75 to 0.90 % funds, and an index of funds carries its
constituents' 2 %); what remains of the gap is the manager's own edge, and it is
left open.

These histories are shorter than they used to be, on purpose: each stops where
its evidence stops, the chains at their deepest real donor NAV (1996-03) and
the overlays at their reference's first day (2000-01). A very reliable
twenty-plus-year backcast is worth more than a heavily simulated forty-year
one. The construction, its measurements and what it leaves open are in
`docs/trend-reconstruction-design.md`.
The lower fidelity is accepted in exchange for full self-generation.
Strategies with no honest donor are not reconstructed at all: a long-volatility
fund such as Amundi Volatility World (LU0319687124) and a discretionary macro
vehicle such as BH Macro (BHMG) trade their own book, not a factor mix, and a
factor regression on them explained 20 % and 0 % of their daily variance. They
carry no simulated history; their `SIM` identifiers simply fall back to the
real (shorter) quotes, from 2007.

## Using it as a library

Everything under `pkg/` is a standard-library-only Go toolkit, bundled data
included; `cmd/` only wires the CLI. Every snippet below is the body of the
runnable example its first comment names, so `go test ./...` keeps it true.

### Study a portfolio in ten lines

`portfolio.NewSpec` builds a spec in code (weights as fractions), and
`analyze.Portfolio` returns the numbers the report is drawn from: statistics,
each holding on the same window, correlations, the risk budget, look-through
composition, warnings. `src` is offline here; live, it is a `marketdata.Client`.

```go
// from analyze.Example_sixtyForty
ctx := context.Background()
src := newFake()

spec, _ := portfolio.NewSpec("60/40",
	portfolio.Line{ID: "IWDA", Weight: 0.6}, // IE00B4L5Y983
	portfolio.Line{ID: "AGGH", Weight: 0.4}) // IE00BDBRDM35
study, err := analyze.Portfolio(ctx, src, spec, analyze.Options{Currency: "EUR"})
if err != nil {
	panic(err)
}

st := study.Stats
fmt.Printf("CAGR %.1f %%, volatility %.1f %%, max drawdown %.1f %%\n", st.CAGR*100, st.Volatility*100, st.MaxDrawdown*100)
fmt.Printf("correlation %s/%s: %.2f\n", study.Aligned.IDs[0], study.Aligned.IDs[1], study.Correlation[0][1])
fmt.Printf("risk budget: %.0f %% / %.0f %% of the variance\n", study.Attribution.Risk[0]*100, study.Attribution.Risk[1]*100)
```

### Get a price history and its raw data

`Fetch` resolves a ticker, ISIN or alias and caches adjusted daily closes;
`FetchExtended` is the CLI's per-asset pipeline. `Raw` pairs unadjusted closes
with `Dividends` (never pair dividends with adjusted closes); `NewSeries`
wraps a consumer's own data.

```go
// from marketdata.Example_priceHistory (compiled, not run: it needs the network)
ctx := context.Background()
client := marketdata.NewClient(marketdata.DefaultCacheDir())

// Real quotes, adjusted, native currency; the slices pkg/metrics takes.
iwda, err := client.Fetch(ctx, "IWDA", time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)) // IE00B4L5Y983
if err != nil {
	panic(err)
}
dates, closes, returns := iwda.Dates(), iwda.Values(), iwda.Returns()
fmt.Println(len(dates), len(closes), len(returns), iwda.Resample(marketdata.Monthly).Len())

// The CLI's pipeline: the bundled backcast in front (SIM), in euros.
long, err := client.FetchExtended(ctx, "IWDASIM", marketdata.FetchOptions{Currency: "EUR"})
if err != nil {
	panic(err)
}
fmt.Println("simulated before", long.SimulatedBefore.Format(time.DateOnly))

// Unadjusted closes, the distributions beside them as cash.
vt, err := client.FetchExtended(ctx, "VT", marketdata.FetchOptions{Raw: true}) // US9220427424
if err != nil {
	panic(err)
}
fmt.Println(len(vt.Dividends), "distributions in", vt.Currency)

// Live: the freshest price, today's 5-minute path (ErrNotCovered off Yahoo).
if q, err := client.Latest(ctx, "IWDA"); err == nil {
	fmt.Println(q.Price, q.Currency, q.Live)
}
if today, err := client.Intraday(ctx, "IWDA"); err == nil {
	fmt.Println(len(today.Points), "ticks today")
}
```

### Dissect one asset

`analyze.Asset` studies one identifier on its longest window: statistics,
calendar years and months, drawdown episodes, and the relative statistics
against a benchmark.

```go
// from analyze.ExampleAsset (synthetic series, offline)
ctx := context.Background()
src := newFake()

a, err := analyze.Asset(ctx, src, "IWDA", analyze.Options{
	Currency:  "EUR",
	Benchmark: "MSCIWORLD",
	From:      time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
})
if err != nil {
	panic(err)
}
fmt.Printf("%s (%s), %s to %s, in %s\n", a.ID, a.Meta.Name,
	a.Stats.Start.Format(time.DateOnly), a.Stats.End.Format(time.DateOnly), a.Series.Currency)
fmt.Printf("CAGR %.1f %%, volatility %.1f %%, beta %.2f\n", a.Stats.CAGR*100, a.Stats.Volatility*100, a.Relative.Beta)
for _, y := range a.Years[:2] {
	fmt.Printf("%d: %+.1f %% (partial: %v)\n", y.End.Year(), y.Return*100, y.Partial)
}
fmt.Printf("%d drawdown episodes, the deepest %.1f %%\n", len(a.Drawdowns), a.Stats.MaxDrawdown*100)
```

### Compare several series on one calendar

`marketdata.AlignSeries` puts series on one calendar where all of them quote
(and errors where `Align` would forward-fill zeros); `pkg/metrics` takes plain
slices, so it also reads a valuation series built elsewhere.

```go
// from metrics.Example_oneCalendar (threeFunds: three synthetic series)
a, err := marketdata.AlignSeries(threeFunds(), time.Time{}, time.Time{})
if err != nil {
	panic(err)
}
r := a.Returns()
corr, cov := metrics.CorrelationMatrix(r), metrics.Covariance(r)
fmt.Printf("from %s: %s/%s %.2f, %s/%s %.2f\n", a.Dates[0].Format(time.DateOnly),
	a.IDs[0], a.IDs[1], corr[0][1], a.IDs[0], a.IDs[2], corr[0][2])
fmt.Printf("%s volatility %.1f %%/yr\n", a.IDs[0], math.Sqrt(cov[0][0]*252)*100)

// Partial flags a first year measured from the first quote; the last row
// ends on the last quote, which its End says.
for _, y := range metrics.CalendarReturns(a.Dates, a.Levels[0], 12) {
	fmt.Printf("year to %s %+5.1f %%, partial=%v\n", y.End.Format(time.DateOnly), y.Return*100, y.Partial)
}
if _, betas, ok := metrics.RollingBeta(a.Dates, a.Levels[1], a.Dates, a.Levels[0], 1); ok {
	fmt.Printf("one-year beta of %s on %s, last: %.2f\n", a.IDs[1], a.IDs[0], betas[len(betas)-1])
}
if v, ok := metrics.VaR(r[0], 0.95); ok {
	fmt.Printf("daily 95 %% VaR of %s: %.2f %%\n", a.IDs[0], v*100)
}
```

### Simulate a portfolio by hand

What `analyze.Portfolio` wires, one step at a time. With flows,
`SimResult.Index` is the time-weighted series (statistics) and `Values`
follows the money (`IRR`); `TWR` recovers the first from the second.

```go
// from portfolio.Example_byHand (synthetic: a fetch callback serving made-up series)
spec, _ := portfolio.NewSpec("60/40",
	portfolio.Line{ID: "IWDA", Weight: 0.6}, // IE00B4L5Y983
	portfolio.Line{ID: "AGGH", Weight: 0.4}) // IE00BDBRDM35
p, err := portfolio.Build(spec, portfolio.BuildOptions{Fetch: synthetic})
if err != nil {
	panic(err)
}
p.Capital = 10_000
p.Contribute = portfolio.Flow{Amount: 500, Period: portfolio.Monthly}
sim, err := portfolio.Simulate(p, 90) // rebalance every 90 days
if err != nil {
	panic(err)
}

stats, _ := metrics.Compute(sim.Dates, sim.Index) // the strategy, flows stripped out
fmt.Printf("CAGR %.1f %%, volatility %.1f %%, max drawdown %.1f %%\n", stats.CAGR*100, stats.Volatility*100, stats.MaxDrawdown*100)

// The saver's own rate: money going in is negative, the final value closes the account.
dates, flows := []time.Time{sim.Dates[0]}, []float64{-p.Capital}
var booked []metrics.Flow // the same flows the other way round, for TWR
for i, d := range sim.FlowDates {
	dates, flows = append(dates, d), append(flows, -sim.FlowAmounts[i])
	booked = append(booked, metrics.Flow{Date: d, Amount: sim.FlowAmounts[i]})
}
last := len(sim.Dates) - 1
irr, _ := metrics.IRR(dates, flows, sim.Dates[last], sim.Values[last])
twr, _ := metrics.TWR(sim.Dates, sim.Values, booked)
fmt.Printf("put in %.0f, worth %.0f, money-weighted %.1f %%/yr\n", p.Capital+sim.Contributed, sim.Values[last], irr*100)
fmt.Printf("time-weighted %+.1f %%, as the index says: %+.1f %%\n", twr*100, sim.Index[last]-100)
```

### Optimize weights

`optimize.Solve` takes aligned daily returns and a `Spec`, whose bounds and
limits bind every objective. Black-Litterman takes the file's weights as its
prior and blends `view:` beliefs into the returns they imply.

```go
// from optimize.ExampleSolve_boundedBlackLitterman (exampleReturns: synthetic daily returns)
spec, err := optimize.ParseSpec("black-litterman,view:TREND:8@70,bounds:TREND:10-40,max-vol:9")
if err != nil {
	log.Fatal(err)
}
if err := spec.Resolve([][]string{{"EQUITY"}, {"TREND"}, {"CASH"}}); err != nil {
	log.Fatal(err)
}
spec.Prior = []float64{0.5, 0.3, 0.2} // the weights written in the file

res, err := optimize.Solve(exampleReturns(750), spec)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("EQUITY %.0f %%, TREND %.0f %%, CASH %.0f %%, feasible %v\n",
	res.Weights[0]*100, res.Weights[1]*100, res.Weights[2]*100, res.Feasible)
```

### What the book is missing

`pkg/suggest` reads what holdings ARE before what they returned: regime
coverage and gaps, redundancies, candidates ranked on walk-forward windows.
Its look-through splits (`AssetClassSplit`, `CurrencySplit`...) fill
`PortfolioStudy.Composition`.

```go
// from suggest.ExampleAnalyze (deterministic synthetic returns)
const n = 500
held := make([]float64, n)
diversifier := make([]float64, n)
for i := range n {
	held[i] = 0.004 * math.Sin(float64(i)/5)
	diversifier[i] = 0.004*math.Cos(float64(i)/5) + 0.0003
}
holdings := []suggest.Holding{
	{ID: "IWDA", Weight: 1, HasMeta: true, Meta: suggest.Meta{AssetClass: "equity"}}, // IE00B4L5Y983
}
candidates := []suggest.Candidate{{
	Meta:        suggest.Meta{ID: "IGLN", AssetClass: "gold"}, // IE00B4ND3602
	PortReturns: held,
	Returns:     diversifier,
	Years:       12,
}}

res := suggest.Analyze(holdings, [][]float64{held}, candidates, suggest.DefaultOptions(), suggest.RegimeFramework())
fmt.Println("gaps:", res.Gaps)
for _, s := range res.Suggestions {
	fmt.Printf("%s at %.0f %% fills %s (%d/%d windows)\n",
		s.Meta.ID, s.Weight*100, s.Fills, s.SharpeWins, s.Windows)
}
```

### FIRE in a dozen lines

A `scenario.Source` draws real-return paths (parametric here; the bootstraps
resample history, `Deflate` makes it real); `Plan.Simulate` runs the
withdrawal kernel, `Solve` turns the question around, and `Plan.Lifetime`
draws the lifespan inside every path (`decumul.ExamplePlan_Simulate_lifetime`).

```go
// from decumul.Example_fire
p := decumul.Plan{
	Capital: 1_000_000, NeedAnnual: 32_000, Years: 35,
	Tax:    decumul.CTOFlatTax{Rate: 0.314},
	Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
}
o := p.Simulate(20_000, 4, 7).Outcome() // paths, workers, seed
fmt.Printf("ruin %.0f%%, median terminal wealth %.1f M\n", o.RuinProb*100, o.TerminalP50/1e6)

spend := p.Solve(0.05, decumul.WithdrawalAxis(10_000, 100_000), 20_000, 4, 7)
fmt.Printf("5%% ruin at %.0f a year\n", math.Round(spend/500)*500)
```

### Reconstruct a missing history

`simgen.Find` returns the recipe `pofo -gen-simdata` ships for an asset, and
`Build` runs it on any `Fetcher`. `Validate` grades a reconstruction on its
overlap with the real quotes (`simgen.ExampleValidate`); `pofo
-verify-simdata ZROZ` renders the full audit.

```go
// from simgen.ExampleFind (offline: bundled reference data only)
r, ok := simgen.Find("ZROZ")
if !ok {
	panic("no recipe")
}
s, err := r.Build(simgen.WithRefData(datasets.Refdata(), offline{}), time.Time{})
if err != nil {
	panic(err)
}
fmt.Println(r.Name)
fmt.Printf("rebuilt from %s, graded against %s\n", s.First().Date.Format(time.DateOnly), r.ValidateAgainst)
```

### Render

`compare.Compute` runs the CLI's comparison, `Studies` hands the numbers
behind every chart, `chart.Line` draws standalone SVG and `report.Render` the
HTML report. The two bundled indices keep it offline.

```go
// from compare.Example_render
client := marketdata.NewClient("") // "" = no disk cache
spec, _ := portfolio.NewSpec("world and US",
	portfolio.Line{ID: "MSCIWORLD", Weight: 0.6},
	portfolio.Line{ID: "SP500", Weight: 0.4})
cmp, err := compare.Compute(context.Background(), client, []*portfolio.Spec{spec}, compare.Options{
	Currency: "USD", NoFees: true, Rebalance: 90, Framework: suggest.RegimeFramework(),
})
if err != nil {
	panic(err)
}

st := cmp.Studies()[0]
svg := chart.Line(chart.Options{Title: st.Spec.Name, Width: 800, Height: 400}, []chart.Series{
	{Name: st.Spec.Name, Dates: st.Sim.Dates, Values: st.Sim.Index},
})

var page strings.Builder
if err := report.Render(&page, cmp.HTMLPage(compare.Decoration{})); err != nil {
	panic(err)
}
fmt.Println(len(st.Holdings), "holdings,", len(st.Correlation), "x", len(st.Correlation[0]), "correlation")
fmt.Println(strings.HasPrefix(svg, "<svg"), strings.Contains(page.String(), "</html>"))
```

### Units

| Quantity | Unit | Where |
|---|---|---|
| Weights | fraction (0.6) | `portfolio.Line`, `Holding.Weight`, `Asset.Weight`, `optimize`, `analyze` |
| Weights | percent (60) | portfolio files, `Holding.RawWeight` |
| Fees (TER) | percent per year (0.20) | `portfolio` (`Line.Fees`, `Holding.Fees`, `EnvelopeFees`, `BorrowSpread`), `marketdata.Client.Fees`, `datasets.Asset.Fees` |
| Fees, volatility targets | fraction per year (0.0020) | `simgen` |
| Returns, CAGR, volatility, drawdowns, VaR | fraction (0.04 = +4 %) | `metrics`, `analyze`, `scenario`, `decumul` (the last two in REAL terms) |
| Ulcer, CWARP | percent points, percent | `metrics.Stats.Ulcer`, `metrics.Stats.CWARP` |
| Rates (`^IRX`, `^ESTR`, `^SOFR`...) | annualized percent LEVEL | `marketdata` series, `portfolio.Portfolio.Cash`: never a return |
| `#meta` directives | percent as written (`max-vol:9`) | fractions once parsed into `optimize.Spec` |

### Where a number comes from

| Question | Where it is answered |
|---|---|
| Is the math right? | `pkg/datasets/golden` (`make golden`): the statistics replayed on frozen real data against published references, Black-Litterman against its papers' tables |
| Why does it differ from another tool? | the Conventions section of `go doc ./pkg/metrics`: 252 days, zero risk-free rate, drawdowns on daily closes, 365.25-day years |
| Is a bundled series right? | the golden package's refdata, gap and spike guards; `pofo -verify-simdata ID` for a backcast against the real quotes; `make verify-catalog` for the data doctor |
| What can a study not know? | its `Warnings`: simulated spans, distributing share classes, definition junctions, unconverted currencies |

### Packages

```
pkg/analyze/      the numbers-only studies: Asset and Portfolio in one call
pkg/marketdata/   data: resolution (aliases, ISIN, catalog), multi-provider
                  sources, cache, fees, simdata, NewSeries, AlignSeries
pkg/metrics/      statistics (CAGR, Sharpe, Sortino, drawdowns, Beta, CWARP,
                  IRR, TWR), correlation and covariance matrices, calendar
                  returns, rolling beta, VaR, risk attribution
pkg/optimize/     weights for max-sharpe / min-volatility / max-return /
                  risk-parity / max-sortino / return-to-drawdown / min-ulcer /
                  max-worst-5y / cwarp / black-litterman, under per-line bounds
                  and volatility / return / drawdown limits
pkg/suggest/      regime coverage, look-through composition, redundancy and
                  gap-filling suggestions
pkg/chart/        SVG charts (Line, Bars, Heatmap) and terminal (Term)
pkg/portfolio/    allocation file format, NewSpec, rebalanced simulation with
                  per-holding return attribution
pkg/compare/      the CLI's comparison pipeline over analyze, and its page
pkg/report/       HTML and text rendering of the comparison model
pkg/simgen/       history reconstruction (composites, TSMOM, backcasts)
pkg/scenario/     return-path generation (parametric, bootstrap, cohorts)
pkg/decumul/      decumulation/FIRE engine + metrics + sweeps + optional
                  stochastic lifetime (estates, annuities); web/ live UI
pkg/datasets/     versioned data (embedded at build time) and its QA:
  assetmeta/        catalog asset metadata (classes, factors, regimes…)
  simdata/          permanent simulated histories (spliced at runtime)
  refdata/          long reference series the backcasts are built on
  golden/           golden tests + frozen fixtures vs external references
cmd/              the pofo binary and the data generators (gen-*-refdata)
```

Each package's `go doc` page holds its conventions and more runnable
examples (`go doc github.com/bpineau/pofo/pkg/metrics`); the catalog reads as
typed records through `datasets.Catalog` and `marketdata.Lookup`. The root
`doc.go` holds the layering (which package may import which) and the units
table above. `pkg/simgen` and `pkg/marketdata` end their documentation with a
"Generator plumbing" section: the exports that serve the data generators
under `cmd/` and are not meant for consumers.

## Known limitations

- Every history proxy is a total-return series except one: QQQ before its
  1999-03 launch rides the Nasdaq-100 PRICE index (^NDX), no total-return
  Nasdaq-100 reaching further back, so that span misses the index's dividend
  yield (0.57 pt/yr on the 1999-2026 overlap; smaller then, not measured).
  QQQM and EQQQ inherit it through QQQ.
- A simulated span is a reconstruction, not the fund: `pofo -verify-simdata`
  grades each engine against the real quotes (managed-futures ones follow
  their strategy's monthly path, not its daily positions).
- A foreign identifier (outside the catalog, whose records all declare a
  currency) whose source reports no quote currency is left unconverted, and
  the report says so.

## Golden tests

`pkg/datasets/golden/` replays the simulation on frozen real data (SPY
2006-2025, URTH 2012-2025) and compares CAGR, volatility, Sharpe, Sortino,
Ulcer, Max Drawdown and TTR against validated external references (official
S&P 500 TR annual returns, canonical GFC/COVID drawdowns,
LazyPortfolioETF).
Any calculation drift beyond the tolerances fails `go test ./pkg/datasets/golden`.

## Development

```sh
make check      # gofmt check + vet + staticcheck + tests (no network)
make help       # every other target (build, golden, simdata, demo…)
```

No external dependencies: standard library only. `AGENTS.md` is the
contributor/AI-agent quick map: repository layout, unit conventions and
traps, house rules and common tasks.
