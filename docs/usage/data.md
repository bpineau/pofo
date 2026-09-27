# Data and backcasts

Every number pofo prints rests on a price series. This page says where the
series come from, what the bundled data holds, how a young fund gets a longer
past, and how to check any of it.

## Two sources: quotes and the bundle

| | Quotes | Bundled data |
|---|---|---|
| What | real daily prices of funds, stocks, indices, currencies | long reference series and reconstructed histories |
| From | Yahoo Finance, Financial Times, Morningstar, Stooq, the ECB, Eurostat, FRED | embedded in the binary and the Go module at build time |
| Network | downloaded once, then cached | never |
| Listed by | `pofo -verify-data`, the catalog | `pofo -dump list` |

`-offline` (on the command line) or `Client.Offline` (in Go) forbids
downloads: the quote cache answers whatever its age, then the bundle, and
anything else is an error.

### Quotes and the cache

pofo resolves an identifier through its aliases, an embedded list of European
fund tickers, a curated catalog of pinned resolutions, then a search across
its sources; the deepest history wins. Closes are **adjusted**: dividends are
reinvested, so a series is a total return.

Quotes are cached in the standard user cache directory for a month by default
(`-cache-age`). A failed refresh serves the stale data with a warning and
never deletes anything. The cache format is private: read cached quotes with
`pofo -dump -offline`, or an offline client in Go, never from the files.

One trap: the published price of a **distributing** share class is a price
return, missing the income it paid out. The report warns about it; nothing
corrects it.

### What is bundled

| Directory under `pkg/datasets/` | What it holds |
|---|---|
| `refdata/` | long reference series: the S&P 500 since 1871, the US total market since 1926, MSCI World and its regions since 1969, Treasury yields and total returns since 1953, T-bills, gold, rolled crude, the Bloomberg Commodity Index since 1991, euro, German, Japanese and British bonds and cash, trend and insurance-linked composites |
| `simdata/` | reconstructed histories of catalog funds and indices (the `SIM` series) |
| `assetmeta/` | the curated catalog: identity, currency, fees, asset class, regimes |
| `broadsample/`, `cape/`, `macropanel/` | research panels: developed-market real returns, the Shiller CAPE, OECD macro series |

```sh
./pofo -dump list | head -4
```

```text
ID                       KIND     FIRST       LAST        POINTS  NAME
BTOP50                   simdata  1986-12-31  2026-07-31  7092    Barclay BTOP50 managed futures (index, net of manager fees)
BTOP50E                  simdata  1986-12-31  2026-09-18  10334   Barclay BTOP50 managed futures, hedged to EUR (index)
BUND-DAILY               refdata  1997-08-07  2026-07-03  7336    German government bond total return (10-year benchmark, EUR, daily)
```

### Currencies and inflation

The command converts every series to one currency (`-currency`, euros by
default) with daily exchange rates: Yahoo first, then Stooq and the ECB
reference rates. The euro crosses reach back to 1971 through a bundled
ECU and Deutsche Mark proxy; before an exchange rate exists, the first known
rate is held flat, with a warning. The library keeps native currencies until
you call `Client.ConvertCurrency`.

Real (inflation-adjusted) statistics deflate by the Eurostat HICP for euro
reports (`^HICP-FR`, extended to 1955 through the OECD CPI) and by the US
CPI-U for dollar reports (`^CPI-US`, since 1913).

## Special identifiers

These work wherever an identifier does: portfolio files, `-assets`, `-dump`,
`Client.FetchExtended`.

| Identifier | Series | History |
|---|---|---|
| `SP500`, `MSCIWORLD` | S&P 500 and MSCI World total return, fee-free indices | 1871, 1969, bundled |
| `^GSPC` | S&P 500 price index | 1927 |
| `^NDX`, `^DJI`, `^IXIC` | Nasdaq-100, Dow Jones, Nasdaq Composite | |
| `^VIX` | CBOE volatility index, a level in percent | 1990, bundled |
| `^BCOM` | Bloomberg Commodity Index, excess return (no collateral); Yahoo withdrew it, so it is served from the bundled `BCOM-ER-USD` | 1991, bundled |
| `^IRX`, `^FVX`, `^TNX`, `^TYX` | US Treasury yields: 13 weeks, 5, 10 and 30 years | |
| `^ESTR`, `^EONIA`, `^EURIBOR3M` | euro money-market rates | 2019, 1999 to 2021, 1994 |
| `^ECB-DFR`, `^ECB-MRO` | ECB policy rates | 1999 |
| `^SOFR`, `^FEDFUNDS`, `^FED-TARGET` | US money-market and policy rates | 2018, 1954, 2008 |
| `^HICP-FR`, `^HICP-<geo>` | Eurostat inflation index | 1955 for France, bundled |
| `^CPI-US` | US CPI-U inflation index | 1913, bundled |
| `USDEUR=X`, any `<AAA><BBB>=X` | an exchange rate, quoted in the second currency | 1971 for euro crosses, bundled |
| `XAUUSD` (`GOLD`), `XAGUSD` | gold and silver spot | `GOLDSIM` from 1968 |
| `CL=F` | WTI crude oil, continuous futures | |

An index symbol (`^...`) is resolved by its symbol only: when its quote
fails, pofo never falls back to a fund found by name, which could be any
product that carries the index's short name.

Yields, rates, `^VIX` and the inflation indices are **levels**, not prices.
They chart fine and give good regime context, but a return computed on them
reads as nonsense, so keep them out of weighted portfolios. `pofo -rates`
charts rate levels; `pofo -rates list` names them.

## Backcasts

Most funds worth studying are young. A backcast (a reconstruction of a fund's
missing past) lets a study reach the decades that hold the real tests: 1973,
the 1980 rates, 2000, 2008.

### The SIM convention

- `IWDA` is the fund's real quotes only, from its launch.
- `IWDASIM` puts the bundled reconstruction in front of them. Real quotes
  always win where they exist, and `Series.SimulatedBefore` marks the join.
- `#meta sim:on` (a file), `-simulate` (a run) and `sim=on` (a `/view` link)
  ask for it wholesale; `-no-simulate` turns it off everywhere.

A fund with no reconstruction falls back to a known total-return proxy
(`VOO` from the S&P 500, `BND` from its mutual-fund sibling), rescaled to its
first real quote, or else keeps its real quotes.

### What is reconstructed

| Family | Examples | Reaches back to |
|---|---|---|
| World and US equity | `IWDASIM`, `URTHSIM`, `VTSIM`, `VTISIM` | 1969 (MSCI World), 1987 (all-world), 1962 (US total market) |
| Treasuries | `TLTSIM`, `IEFSIM`, `ZROZSIM`, `SHYSIM` | 1953 to 1991 |
| Gold | `GOLDSIM` | 1968 |
| Capital-efficient (stacked) funds | `NTSGSIM`, `NTSZSIM`, `GDESIM`, `RSSBSIM`, `RSSTSIM` | 1968 to 2000 |
| Small-cap value | `ZPRVSIM`, `AVWSSIM` | 1963, 1994 |
| Managed futures | `DBMFSIM`, `KMLMSIM`, `CTASIM`, the AQR classes | 1996 |
| Insurance-linked | `ILSFUND` | 2005 |

`pofo -dump list` gives the exact list and dates.

### How they are built

Each recipe combines real series: the fund's own older share classes, the
index it tracks, mutual funds with decades of history, bundled references.
A few principles hold throughout.

- **Evidence bounds length.** A history stops where its evidence stops. The
  managed-futures chains end at their deepest real fund NAV (1996); nothing
  is extrapolated in front of that.
- **Fees are charged once.** A donor fund's price is already net of its own
  fee, so a recipe charges only the difference to the target's fee. An
  academic factor pays no fee and gets a measured haircut instead.
- **Weights follow the market.** A world-equity blend drifts with its legs'
  returns from a published split, rather than using today's split backwards.
- **No honest donor, no backcast.** A long-volatility fund or a discretionary
  macro fund trades its own book, which no factor mix explains, so they keep
  their real quotes only.

Managed futures are the hardest case: a fifty-market programme cannot be
replicated day by day. Their reconstructions aim for the right month and the
right level, taken from composites of real funds net of fees, with an engine
supplying the daily texture. Monthly agreement with the real funds is good;
daily agreement is modest by construction.

The design records, with every measurement, are in [`docs/specs/`](../specs/)
(for example the
[managed-futures field guide](../specs/trend-reconstruction-design.md) and the
[zero-coupon Treasury record](../specs/long-treasury-zero-coupon-design.md)).

### How far to trust a backcast

`pofo -verify-simdata` replays each recipe **without** the real quotes it
normally splices in, and lays the result over those quotes where they exist.
That overlap is the only window on which a reconstruction can be judged.

```sh
./pofo -verify-simdata                  # every recipe: HTML report plus a verdict row each
./pofo -verify-simdata ZROZ DBMF        # just these
./pofo -verify-simdata -json > qa.json  # the whole audit, for a program
```

```text
ID    LEVEL  PATH  MONTHLY  WEEKLY  DAILY  GAP %/YR  SE    TE/VOL  YEARS  REFERENCE
ZROZ  ok     ok    0.99     0.98    0.96   +0.03     0.82  0.29    16.9   ZROZ
```

| Verdict | Question |
|---|---|
| **level** | does the reconstruction earn the fund's return? (a hot one flatters every backtest built on it) |
| **path** | does it move with the fund, month by month? |

The columns give the correlations at each cadence, the return gap with its
standard error, and the tracking error relative to the fund's own volatility.
The HTML report adds the curves, a drift panel, and for donor chains a grade
for every junction. `-offline` replays it from the cache, so two runs around a
recipe change compare exactly.

From Go, `simgen.Find` and `simgen.Validate` do the same (see the
[backcasts guide](library/backcasts.md)).

## How far to trust a number

| Question | Where it is answered |
|---|---|
| Is the math right? | `make golden`: statistics replayed on frozen real data (SPY 2006-2025, URTH 2012-2025) against published references, Black-Litterman against its papers' tables |
| Why does it differ from another tool? | volatility annualizes at each series' own cadence, the risk-free rate is zero, drawdowns use daily closes, a year is 365.25 days (the conventions section of `go doc github.com/bpineau/pofo/pkg/metrics`) |
| Is a bundled series sound? | the gap and spike guards (`marketdata.FindGaps`, `marketdata.FindSpikes`), run on every bundled file by the test suite; `go run examples/code/scanbundle.go` runs them on demand |
| Is a quote series sound? | `pofo -verify-data`: bad points, gaps, stale feeds, implausible moves for the asset class, identity against the catalog record |
| What can a study not know? | its warnings: simulated spans, distributing share classes, definition changes in a series, unconverted currencies |

### Known limitations

- Before its 1999 launch, `QQQ` rides the Nasdaq-100 **price** index, the only
  one reaching that far, so that span misses the dividend yield (about half a
  point a year later on). `QQQM` and `EQQQ` inherit it.
- A reconstruction is not the fund. Managed-futures ones follow their
  strategy's monthly path, not its daily positions.
- An identifier outside the catalog whose source reports no currency is left
  unconverted, and the report says so.

## Refreshing the bundled data

For maintainers. These targets need the network and move bundled files.

```sh
make refresh          # every bundled series from its live source, in dependency order
make check golden     # both must stay green: refreshing data never requires a code change
make verify-catalog   # the data doctor over the whole catalog
make figure-drift     # which of the book's frozen figures the refresh made stale
```

Every generator checks what it downloaded before it writes, and refuses
otherwise: a source that stopped updating, a frozen run, a value out of its
historical range, or (for a source that never revises, like the T-bill rate
or the gold fix) a refresh that does not reproduce the bundled history point
for point. Generators read each series at its publisher rather than through a
mirror wherever they can, because a mirror freezes behind a generator that runs
fine: the OECD series (euro, British and Japanese references, the macro panel)
come from the OECD's own SDMX API, which admits 60 downloads an hour, so each
generator asks for a whole dataflow's series in one request. To find a bundled
file that has fallen behind its source:

```sh
go run examples/code/scanbundle.go -stale 45
```

A file that stops on purpose (its source was discontinued, like EIA's WTI
futures settlements after 2024-04-05, or it is only kept until real quotes take
over) carries an `# ends:` header and is never reported stale.

To work on one reconstruction:

```sh
./pofo -gen-simdata -dry ZROZ     # build and validate, write nothing
./pofo -gen-simdata ZROZ          # write pkg/datasets/simdata/, then make build to embed it
./pofo -verify-simdata ZROZ       # grade it against the real quotes
```

```text
✓ ZROZ           1953-04-30 → 2026-09-25 (16297 points)
  corr=0.958 (weekly 0.985) beta=0.90 TE=6.7%/yr CAGR sim 1.79% vs real 1.76% (overlap 4210 d from 2009-11-04 to 2026-09-17) vs ZROZ
```

`make help` lists every per-series target.
