# pofo

Go tool to visualize and compare investment portfolios over time, plus
reusable libraries to fetch price histories, compute risk/return metrics
and produce SVG charts.

The CLI reads allocation files, downloads price histories (Yahoo Finance,
Financial Times, Morningstar, Stooq, ECB), rebuilds the missing past (proxies and
simulated data), simulates each portfolio with periodic rebalancing and
generates a self-contained HTML report opened in the browser (comparison and
statistics front and center; per-portfolio sections collapsed, each with a
performance curve, a realized-contribution timeline (who carried the trailing
12 months, stacked around zero under a macro-regime strip), four look-through
composition pies (geography, currency exposure, equity sectors, asset type
with stacked funds opened into their legs), its macro-regime coverage with
each bar split by contributing holding, and the per-regime realized
contributions that mirror it; hover any chart for exact figures).

## Usage

```sh
go build ./cmd/pofo                       # self-contained binary (datasets embedded)
./pofo my-portfolio.txt other.txt         # HTML report in /tmp + open
./pofo -assets WPEA,NTSG,CSPX             # compare individual assets (100% each)
./pofo -cli -assets VOO,IWDA              # quick check in the terminal
./pofo -warmup                            # pre-warm the catalog cache
./pofo -gen-simdata                       # regenerate pkg/datasets/simdata (then rebuild)
```

The binary can be installed anywhere: simulated histories and reference
series are **embedded at build time** (`go:embed` of `pkg/datasets/`), and the
quote cache lives in the standard user cache directory
(`~/Library/Caches/pofo` on macOS, `~/.cache/pofo` on Linux).

The `-assets` option treats each identifier as a portfolio invested 100% in
it, handy for comparing ETFs against each other without writing a file. It
can be combined with portfolio files.

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
| `optimize:OBJ[,max-weight:C]` | let the optimizer choose the weights for objective `OBJ` (table below), optionally capping each asset at C% |
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
| `cwarp` | maximize CWARP vs the `-benchmark` (a diversifier selector, see [CWARP](#cwarp)) |

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
file just as easily).

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
df, horizon, pension, spending rules and the French taxes. The dashboard
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

The explorer also embeds **the FIRE book** at `/book/fr/` (a small `book`
link sits at the bottom of the "How this machine works" fold): a
French-language handbook of decumulation, withdrawal strategies, resilient
portfolios, buffers and French taxation, written as cross-linked articles
and served straight from the binary (`pkg/firebook`; an English translation
will join at `/book/en/` later).

The reusable pieces live in the library: `pkg/scenario` (return-path
generation) and `pkg/decumul` (the withdrawal engine, FIRE outcome metrics
and sweeps), with the thin web layer under `pkg/decumul/web`; the book is its
own package, `pkg/firebook`, mountable by any server.

## Web app

`pofo -serve` starts the whole tool as one local web app, four surfaces on a
single port:

| URL | Surface |
|---|---|
| `/` | the **hub**: the bundled example portfolios, tick any and compare them |
| `/view` | the **visualizer**: the same HTML comparison report the CLI writes, addressed by a shareable URL |
| `/fire/` | the **FIRE simulator** (`-fire`, mounted under a prefix) |
| `/book/fr/` | the **FIRE book**, with a small nav bar back to the other surfaces |

```sh
./pofo -serve                             # http://127.0.0.1:8787/
./pofo -serve -listen 127.0.0.1:9000      # a different port
./pofo -serve examples/dragon-decumulation-household.txt  # seed the FIRE panel from a file
```

`-listen` defaults to `127.0.0.1:8787` (loopback only). Portfolio file
arguments feed the FIRE simulator's historical models, exactly as they do
for `-fire`.

The visualizer is driven entirely by its query string, so a comparison is a
link you can bookmark or share:

```
/view?ex=dragon-decumulation-household&ex=claude-dragonlite
/view?p=NTSG:60,IGLN:20,IBCI:20!sim:on&currency=EUR
```

`ex=` names a bundled example (repeat it to stack several). `p=` is an ad-hoc
portfolio, `ID:WEIGHT` pairs comma-separated, with `!key:value` meta
directives appended (`!` replaces the file format's `;`, which a query string
cannot carry). Global options (`start`, `end`, `rebalance`, `currency`,
`bench`, `sim`) mirror the CLI flags. Up to six portfolios per page, twenty
holdings each.

`p=` identifiers are **catalog-only**: the tool resolves them from the
embedded catalog (ids, ISINs, aliases, bundled fund tickers, the `SIM` suffix
allowed) and never fetches an arbitrary or unknown identifier on behalf of an
anonymous visitor, so a raw quote symbol outside the catalog is rejected
before any network call. `ex=` files carry no such limit; they are the vetted
builds shipped in the binary.

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
| `-start` | `2006-01-01` | desired start date |
| `-benchmark` | `^GSPC` | reference for Beta, capture ratios and the CWARP replacement |
| `-currency` | `EUR` | convert every series (and the benchmark) to this currency; empty disables |
| `-cache-age` | `720h` (1 month) | cache freshness before re-downloading |
| `-assets` | | list `A,B,C`: each asset compared as a 100% portfolio |
| `-cli` | | curves and summary table in the terminal, no HTML |
| `-width` | `$COLUMNS` or 100 | width of the `-cli` chart (wider = more granularity) |
| `-warmup` | | pre-warm the built-in asset catalog then exit |
| `-verify-data` | | data doctor: check the referenced assets' quotes (or the whole catalog) for anomalies (bad points, gaps, stale feeds), then exit |
| `-suggest` | | recommend catalog assets to add for better regime coverage, flag redundant holdings, then exit |
| `-coverage` | | offline advisor: show which regimes/factors a portfolio misses and the catalog assets that fill them, then exit |
| `-fire` | | open the local decumulation/FIRE explorer (sliders, ruin curves), optionally for a portfolio file, then serve until stopped |
| `-serve` | | serve the whole web app (hub, visualizer, FIRE simulator, book) on one port until stopped |
| `-listen` | `127.0.0.1:8787` | listen address for `-serve` (loopback by default) |
| `-framework` | `regimes` | classification for coverage and `-suggest`: `regimes` (macro quadrants) or `factors` (risk factors) |
| `-no-open`, `-no-simulate` | | do not open the browser / ignore SIM suffixes |

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
- **Cache**: 1 month by default; a failed refresh **serves the stale data**
  with a stderr warning (charts may stop before today), and never deletes
  anything.
- **History extension** (`…SIM` identifiers only): first the
  `pkg/datasets/simdata/` files (below), otherwise a known proxy (VOO→^GSPC,
  BND→VBMFX, …), rescaled to the first real quote. The report flags every
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
| `^HICP-FR`, `^HICP-<geo>` | Eurostat inflation index (all-items HICP) | 1955→ (FR), bundled |
| `^CPI-US` | US CPI-U inflation index (FRED) | 1913→, bundled |
| `USDEUR=X`, any `<AAA><BBB>=X` | FX cross, quoted in the second currency | 1971→ (euro crosses), bundled |
| `XAUUSD` (alias `GOLD`), `XAGUSD` | gold / silver spot (via futures) | `GOLDSIM` reaches 1968 |
| `CL=F` | WTI crude oil continuous futures | |

For example, `pofo -cli -assets '^VIX'` charts the VIX, and
`pofo -assets 'USDEUR=X,^CPI-US'` compares the dollar and US inflation.
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
(`Quote.Live == false`). When Yahoo is down or throttled, the call degrades
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
```

## Simulated data (pkg/datasets/simdata/)

Complex assets (90/60 funds, managed futures…) are rebuilt by `pkg/simgen`
from long-history building blocks, validated against their real quotes,
then stored as self-documenting CSVs (method, validation, date) in
`pkg/datasets/simdata/`:

```sh
./pofo -gen-simdata                   # regenerate everything (then make build to re-embed)
./pofo -gen-simdata -dry NTSX         # validate without writing
```

Every series is built **only from quotes the tool itself can fetch**
(Vanguard/Yahoo funds with decades of history, the `^IRX` cash rate, gold and
oil futures) combined by the in-house composite, TSMOM trend, and regression
backcast engines; no third-party data is bundled. External index series are
used solely to cross-check quality during development, never shipped.

Bundled recipes and measured quality (daily / weekly correlation of returns
vs the real series; the real series is always grafted on top of the
simulation wherever it exists):

| Asset | Method (building blocks) | Validation (daily / weekly corr) |
|---|---|---|
| NTSX (UCITS) | 0.90×VFINX + 0.60×(VFITX−cash) + 0.10×cash (1991→) | 0.96 / 0.99 |
| NTSG (UCITS) | global 90/60 US/intl variant (1999→) | 0.39 / 0.86 (thin LSE listing, short overlap) |
| NTSZ (UCITS) | eurozone 90/60: 0.90×EZU(EUR) + 0.60×(euro govt−cash) + 0.10×cash, euro-native refdata (1986→) | 0.63 / 0.93 (fund launched 2025-09, short overlap; use monthly stats) |
| URTH (MSCI World) | 0.60×VFINX + 0.40×VTMGX (1999→) | 0.90 / 0.97 |
| IWDA (MSCI World) | 0.60×VFINX + 0.40×VTMGX (1999→) | 0.60 / 0.85 (GBP listing, short overlap) |
| VT (total world) | 0.60×VFINX + 0.30×VTMGX + 0.10×VEIEX (1999→) | 0.98 / 0.99 |
| RSSB (100/100 stocks+bonds) | VT composite + 1.0×(VFITX−cash) (1999→) | 0.95 / 0.99 |
| ZPRV (US small-cap value, UCITS) | DFSVX (DFA US Small Cap Value, 1993→), real grafted 2015 | 0.67 / 0.91 |
| SHY (1-3y Treasury) | VFISX short Treasury (1991→) | 0.81 / 0.89 |
| IEF (7-10y Treasury) | VFITX intermediate Treasury (1991→) | 0.95 / 0.96 |
| TLT (20+y Treasury) | VUSTX long Treasury (1986→) | 0.98 / 0.99 |
| ZROZ (25+y STRIPS) | 1.65×(VUSTX−cash) (1986→) | 0.97 / 0.97 |
| DBMF (managed futures) | 12-month TSMOM on a 7-market basket (2001→) | 0.52 / 0.55 |
| KMLM (managed futures) | 12-month TSMOM, 15% target vol (2001→) | 0.35 / 0.32 |
| CTA (managed futures) | 12-month TSMOM, 10% target vol (2001→) | 0.20 / 0.24 |
| Winton Trend-Equity (UCITS) | 0.60×VFINX + 0.40×VTMGX + 0.50×TSMOM trend (2001→) | 0.65 / 0.84 |
| Amundi Volatility, BH Macro | regression backcast **rejected** (R² 0.20 / 0.00) | real history only (2007→) |

Managed-futures correlations are modest: each fund runs a faster, partly
discretionary strategy that a single 12-month TSMOM rule only approximates.
The lower fidelity is accepted in exchange for full self-generation.
Discretionary strategies that cannot be honestly replicated with factors are
rejected below an R² floor rather than shipped as invented data; the matching
`SIM` identifiers then simply fall back to the real (shorter) history.

## Using it as a library

The repository is also a toolkit for writing other portfolio-processing
applications. Layout:

```
pkg/marketdata/   data: resolution (aliases, ISIN, catalog), multi-provider
                  sources, cache, fees, simdata, alignment
pkg/metrics/      statistics (CAGR, Sharpe, Sortino, drawdowns, Beta, CWARP, IRR…)
pkg/optimize/     weights for max-sharpe / min-volatility / risk-parity /
                  max-sortino / return-to-drawdown / min-ulcer / max-worst-5y /
                  cwarp
pkg/suggest/      regime coverage, look-through composition, redundancy and
                  gap-filling suggestions
pkg/chart/        SVG charts (Line, Bars, Heatmap) and terminal (Term)
pkg/portfolio/    allocation file format + rebalanced simulation with
                  per-holding return attribution
pkg/report/       HTML and text rendering of the comparison model
pkg/simgen/       history reconstruction (composites, TSMOM, backcasts)
pkg/scenario/     return-path generation (parametric, bootstrap, cohorts)
pkg/decumul/      decumulation/FIRE engine + metrics + sweeps; web/ live UI
pkg/datasets/     versioned data (embedded at build time) and its QA:
  assetmeta/        catalog asset metadata (classes, factors, regimes…)
  simdata/          permanent simulated histories (spliced at runtime)
  golden/           golden tests + frozen fixtures vs external references
cmd/              the pofo binary (report, warmup, gen-simdata)
```

Everything consumable as a library lives under `pkg/`, the bundled data
(catalog and simulated histories) included, via `pkg/datasets`; `cmd/` only
contains the CLI wiring.

Each package has its documentation page, calculation conventions included
(`go doc github.com/bpineau/pofo/pkg/metrics`), and runnable examples:

```go
import (
	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// Fetch a price history (transparent resolution + caching). FetchExtended
// is the do-what-I-mean variant: it also splices simulated history behind
// "…SIM" identifiers and converts the currency, exactly like the CLI;
// Fetch is the raw real-quotes-only building block underneath.
client := marketdata.NewClient(marketdata.DefaultCacheDir())
series, err := client.FetchExtended(ctx, "NTSGSIM", marketdata.FetchOptions{Currency: "EUR"})

// Compute CAGR, Sharpe, Sortino, Ulcer, MaxDD, TTR, Beta…
stats, err := metrics.Compute(dates, values)

// Render a standalone SVG.
svg := chart.Line(chart.Options{Title: "Comparison"}, []chart.Series{{Name: "P1", Dates: dates, Values: values}})

// The core path in three calls: parse a portfolio file, build it (each
// holding fetched through the callback), then simulate with 90-day
// rebalancing. Statistics chain on sim.Index via metrics.Compute.
spec, _ := portfolio.ParseFile("p.txt")
p, _ := portfolio.Build(spec, portfolio.BuildOptions{
	Fetch: func(id string) (*marketdata.Series, error) {
		return client.FetchExtended(ctx, id, marketdata.FetchOptions{Currency: "EUR"})
	},
})
sim, _ := portfolio.Simulate(p, 90)

// Read the bundled asset catalog as typed datasets.Asset records (name, TER,
// UCITS, geography, sectors, asset class…); AssetMeta() returns the same data
// as raw JSON if you prefer your own struct.
for _, a := range datasets.Catalog() {
	_ = a.Name // a.Fees, a.Geography, a.AssetClass, a.UCITS…
}
// Resolve a ticker / alias / ISIN to its full record in one call:
iwda, ok := marketdata.Lookup("IWDA") // → (datasets.Asset, true)
_ = iwda.Fees                         // 0.20  (percent/yr)
```

- `datasets`: the versioned data embedded at build time; `Catalog()` returns
  the typed asset list (`Asset`), `AssetMeta()` the same data as raw JSON.
- `marketdata`: resolution (aliases, ISIN, catalog), `Lookup` for an asset's
  full metadata, `Resolve` to inspect the resolved source/symbol, multi-source
  daily downloads, `Intraday` for the live 5-minute path, `Latest` for the
  freshest quote, cache, simdata, proxies; `FetchExtended` bundles the whole
  per-asset pipeline in one call.
- `suggest`: regime/factor coverage, look-through composition splits (asset
  classes, geography, currency exposure, equity sectors, duration) and
  gap-filling (consumes `datasets.Asset`).
- `metrics`: statistics over value series (returns, drawdowns, Beta).
- `chart`: pure-stdlib inline SVG charts.
- `portfolio`: allocation file parsing, `Build` (spec + fetch callback →
  simulatable portfolio) and rebalanced simulation; `Simulate` attributes
  each day's return to its holdings (`Contributions`, `MonthlyContributions`).
- `report`: HTML report rendering.
- `simgen`: reconstruction engine (linear composites, TSMOM
  trend-following engine, regression backcasts) and validated recipes, all
  built from fetchable quotes only.

## Known limitations

- Price-index proxies (^GSPC, ^NDX…) omit dividends over the simulated
  portion; managed-futures replications (corr ≈ 0.3–0.5) reflect those
  strategies' regime, not their daily positions.
- Assets whose quote currency cannot be determined are left unconverted
  (flagged in the report warnings).

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
