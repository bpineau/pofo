# The command line

`pofo` has one default mode, the comparison report, and a set of modes that
each answer one question and exit. This page goes through them by task.
`./pofo -h` lists every flag.

| I want to... | Mode |
|---|---|
| compare portfolios or funds | `pofo FILE...`, `pofo -assets A,B` |
| see it in the terminal | add `-cli` |
| see what each weight buys | `pofo -sweep FILE` |
| find what a portfolio is missing | `pofo -coverage FILE`, `pofo -suggest FILE` |
| plan a retirement (FIRE) | `pofo -fire [FILE]` |
| run the web app | `pofo -serve` |
| get the series themselves, as CSV | `pofo -dump A,B` |
| measure one series against another | `pofo -pair A,B` |
| chart interest rates | `pofo -rates ^ESTR,^SOFR` |
| check the data | `pofo -verify-data`, `pofo -verify-simdata` |

## Compare portfolios

Give it portfolio files, and it writes a self-contained HTML report and opens
it:

```sh
./pofo examples/portfolios/golden-butterfly.txt examples/portfolios/tradi-60-40.txt
./pofo -out report.html -no-open my-portfolio.txt
```

`-assets` (or `-a`) treats each identifier as a portfolio 100 % invested in
it, which compares funds without writing a file. It combines with files:

```sh
./pofo -assets WPEA,NTSG,CSPX
./pofo -assets VOO my-portfolio.txt
```

The report puts the comparison first: the growth curves, the statistics table
and the drawdowns. Each portfolio then has its own folded section:

- its performance curve and a realized-contribution timeline (which holding
  carried the trailing twelve months);
- look-through composition pies: geography, currency exposure, equity
  sectors, asset type, with stacked funds opened into their legs;
- its macro-regime coverage (growth, deflation, inflation, crisis);
- a risk budget: each asset class's share of the variance next to its share
  of the capital. The two often differ by a factor of three.

Hover any chart for the exact figures.

### In the terminal

`-cli` prints the curves and the statistics table instead of writing HTML.
`-width` sets the chart width.

```sh
./pofo -cli -assets VOO,IWDA
./pofo -offline -cli -currency USD -assets SP500,MSCIWORLD
```

### Options that shape any comparison

| Option | Default | Effect |
|---|---|---|
| `-start`, `-end` (`-s`, `-e`) | the longest common window | the analysis window, `YYYY-MM-DD` |
| `-currency` | `EUR` | convert every series, benchmark included; empty keeps native currencies |
| `-rebalance` | `90` | rebalance every N calendar days, `0` = never (a file's `#meta rebalance:` wins) |
| `-benchmark` | `^GSPC` | the reference for beta, capture ratios and CWARP; empty drops them |
| `-simulate` (`-b`) | off | backcast every identifier, as if each carried the `SIM` suffix |
| `-no-simulate` | off | real quotes only, whatever the files or `-simulate` say |
| `-no-fees` | off | do not look up each fund's ongoing charge |
| `-offline` | off | never download: the quote cache whatever its age, then the bundled data |
| `-cache-age` | `720h` | re-download quotes older than this |

## See what each weight buys: `-sweep`

`-sweep` moves one holding's weight across a grid, keeps the other lines in
their proportions and reruns the full simulation at each point:

```sh
./pofo -offline -sweep examples/portfolios/golden-butterfly.txt
```

```text
golden-butterfly: one line at a time, the others keeping their proportions
1991-10-28 → 2026-09-24, rebalanced every 365 days, EUR

IWDA (written 20.0 %)
    weight      CAGR       vol  Sharpe     maxDD     TTR   worst5y
     0.0 %    7.93 %   11.17 %    0.72  -20.31 %   3.2 y    1.14 %
    10.0 %    8.09 %   11.11 %    0.74  -22.57 %   3.2 y    1.13 %
    20.0 %    8.23 %   11.23 %    0.75  -24.81 %   3.2 y    0.25 %  <- written
    30.0 %    8.35 %   11.52 %    0.74  -27.06 %   3.5 y   -0.91 %
    40.0 %    8.45 %   11.98 %    0.72  -30.95 %   5.0 y   -2.09 %
...
```

Read down a column to see a sleeve's job: here more world equity adds a little
CAGR but costs drawdown and the worst five-year stretch. Read across sleeves
to find the sane range of each line. The grid runs from 0 to 45 % in 5-point
steps; `-sweep-step` refines it. Every row shares one window, so rows compare.

## Find what a portfolio is missing

`-coverage` is an offline read of a portfolio's macro-regime coverage, with
the catalog assets that would fill each gap:

```sh
./pofo -coverage examples/portfolios/tradi-60-40.txt
```

```text
Coverage (by weight):
  growth      ████████████          60 %
  deflation   ████████              40 %
  inflation                          0 %   ← gap
  crisis                             0 %   ← gap

To fill the gaps, the catalog offers (run -suggest to rank them):
  inflation:
    gold             DE000A0S9GB0, GLD, IE00B4ND3602 … (+3)
    managed-futures  AQMIX, BTOP50, BTOP50E … (+24)
...
```

`-suggest` then ranks those candidates. It adds each one at a modest weight
and checks, on walk-forward windows, that Sharpe and drawdown improve
consistently rather than in one lucky stretch. It keeps at most one suggestion
per asset class, and it also flags redundant holdings (three S&P 500 trackers
are one bet).

`-framework factors` switches both modes from the four macro regimes to risk
factors (market, size, value, momentum, quality, term, credit, alternative,
cash). The regime view stays the default because gold, trend and volatility
all land in the single "alternative" factor.

## Plan a retirement: `-fire`

`-fire` opens the FIRE simulator in your browser; a portfolio file seeds it
with that portfolio's own history:

```sh
./pofo -fire
./pofo -fire examples/portfolios/fire-decumulation-core.txt
```

The [FIRE guide](fire.md) explains what it computes and how to read it.

## Run the web app: `-serve`

`-serve` starts the visualizer, the FIRE simulator and the book on one local
port, `http://127.0.0.1:8787/` by default. See the [web guide](web.md).

## Get the series themselves: `-dump`

`-dump` prints series to standard output as one long CSV, for another program
to read. Log lines go to standard error.

```sh
./pofo -offline -dump TREASURY-LONG-USD -monthly -start 2020-01-01
```

```text
# TREASURY-LONG-USD name: US long-term Treasury total return (20-year par bond on the long constant-maturity yield, month-end)
id,date,value
TREASURY-LONG-USD,2020-01-31,6492.165025
TREASURY-LONG-USD,2020-02-28,6873.785348
...
```

It fetches exactly as a report would (`SIM`, `-simulate`, the cache), cuts to
`-start` and `-end`, and keeps month-end closes under `-monthly`. Each series
stays in its native currency unless `-currency` is given. Values parse back
exactly.

`-dump list` names every bundled series with its dates. Read the output with
`pandas.read_csv(path, comment="#")`, or in Go with `marketdata.ReadLongCSV`.

## Measure one series against another: `-pair`

`-pair A,B` measures a candidate A against a reference B: a backcast against
the real fund, a fund against its index, a refreshed file against its
previous version. Here a London-listed world fund against its index:

```sh
./pofo -offline -pair IWDA,MSCIWORLD -currency USD -start 2016-01-01
```

```text
window      2016-01-04 to 2026-08-31 (10.7 years)
CAGR        A +13.13 %, B +12.91 %, gap +0.23 pt/yr (standard error 0.77)
level       A ends +2.15 % from B, both rebased at the start

returns  periods  per year  corr   vol A    vol B    vol ratio  tracking  beta   alpha/yr
daily    2629     252       0.761  16.08 %  15.48 %  1.039      10.92 %   0.791  +3.12 %
monthly  126      12        0.986  13.99 %  14.66 %  0.954      2.49 %    0.941  +0.79 %
...
```

The daily correlation is low only because London closes hours before New York;
the monthly figure is the one to read. `-lead-lag` ranks the daily divergences
forgiving that one-session offset. The full study also lists the largest
divergences, dated, and the calendar years side by side. `-json` prints it as
JSON.

A side holding a slash or ending in `.csv` is a `date,value` file, which is how
a data refresh is checked against the version it replaces:

```sh
git show HEAD~1:pkg/datasets/refdata/MSCIWORLD-USD.csv > /tmp/old.csv
./pofo -offline -pair MSCIWORLD-USD,/tmp/old.csv
```

## Chart interest rates: `-rates`

Rates are levels in annualized percent, not prices, and euro rates were
negative for years, so a return computed on them means nothing. `-rates`
charts the levels and summarizes each symbol:

```sh
./pofo -rates ^ESTR,^EURIBOR3M,^ECB-DFR -start 2015-01-01
./pofo -rates list          # every available symbol
```

## Check the data

| Command | What it checks |
|---|---|
| `pofo -verify-data FILE` or `-verify-data -assets ID` | each quote series: bad points, gaps, stale feeds, implausible moves for its asset class, identity against its catalog record |
| `pofo -verify-simdata [ID...]` | each backcast replayed against the real quotes: an HTML report plus one verdict row per recipe (`-json` for the whole audit) |

```sh
./pofo -offline -no-open -verify-simdata ZROZ DBMF
```

```text
# Trend / managed futures
ID    LEVEL  PATH  MONTHLY  WEEKLY  DAILY  GAP %/YR  SE    TE/VOL  YEARS  REFERENCE
DBMF  warn   warn  0.89     0.85    0.78   -1.26     2.05  0.66    7.3    DBMF

# Capital-efficient / stacked
ID    LEVEL  PATH  MONTHLY  WEEKLY  DAILY  GAP %/YR  SE    TE/VOL  YEARS  REFERENCE
ZROZ  ok     ok    0.99     0.98    0.96   +0.03     0.82  0.29    16.9   ZROZ
```

The [data guide](data.md#how-far-to-trust-a-backcast) explains the verdicts.

## Other modes

| Command | What it does |
|---|---|
| `pofo -warmup` | download and cache the whole catalog |
| `pofo -export-epub FILE [-book-lang en]` | write the FIRE book as an EPUB 3 file |
| `pofo -gen-simdata [ID...]` | regenerate the bundled backcasts (maintainers; see the [data guide](data.md#refreshing-the-bundled-data)) |
| `pofo -indexnow ORIGIN -indexnow-key KEY` | push a public deploy's URLs to search engines (see the [web guide](web.md#running-a-public-instance)) |
| `pofo -book-drift` | list the English articles that lag their French source |
