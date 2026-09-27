# Portfolio files

A portfolio is a plain text file, one holding per line. The command reads it,
the web app builds the same text from a URL, and the library parses it with
`portfolio.Parse`. [`examples/portfolios/`](../../examples/portfolios/README.md)
holds about forty ready-made ones.

## The format

```text
# Golden Butterfly, modernized
#meta rebalance:365 sim:on

20  IWDA          # total-market equity
20  ZPRV          # US small-cap value
20  TLT           # long US Treasuries
20  SHY           # short US Treasuries
20  GOLD          # gold
```

Each line reads `<weight in %> <identifier> [TER in %/year]`:

- the **weight** is a percentage; a decimal comma is accepted (`25,5`);
- the **identifier** is a ticker (`VTI`, `IWDA`, `CW8`), an ISIN
  (`IE00B4L5Y983`) or a catalog alias (`GOLD`, `NTSG`);
- the optional **TER** overrides the ongoing charge pofo would look up.

Everything after a `#` is a comment, and blank lines are ignored. The
portfolio takes the file's name, without its extension.

Weights that do not sum to 100 are normalized, with a warning. A weight above
100 % is refused unless the file declares leverage (below).

## Identifiers

pofo resolves an identifier in order: its built-in aliases, an embedded list
of European fund tickers, the curated catalog, then a search across its quote
sources, the deepest history winning. The report prints each resolution
(`resolved IWDA -> ...`); read that line first when a figure surprises you.

Some special names work wherever an identifier does:

| Identifier | Series |
|---|---|
| `^GSPC`, `^NDX`, `^DJI` | S&P 500, Nasdaq-100, Dow Jones price indices |
| `SP500`, `MSCIWORLD` | the S&P 500 and MSCI World total-return indices, fee-free, bundled |
| `^HICP-FR`, `^HICP-<geo>`, `^CPI-US` | inflation indices (Eurostat HICP, US CPI-U), bundled |
| `USDEUR=X`, any `<AAA><BBB>=X` | an exchange rate, quoted in the second currency |
| `GOLD` (`XAUUSD`), `XAGUSD`, `CL=F` | gold, silver, WTI crude |
| `^VIX`, `^IRX`, `^TNX`, `^ESTR`, `^SOFR`... | volatility and interest-rate **levels** |

Levels are not prices. They chart fine, but a return computed on them is
meaningless, so keep them out of weighted portfolios; chart rates with
`pofo -rates` instead. The [data guide](data.md#special-identifiers) has the
full list.

## Longer history: the SIM suffix

A bare identifier (`DBMF`, `NTSG`) uses the fund's real quotes only, so its
history starts at its launch. Add `SIM` (`DBMFSIM`) to put a reconstructed
past in front of the real quotes. Real quotes always win where they exist,
and the report marks every simulated span.

Three ways ask for it without editing each line:

| Where | How |
|---|---|
| a whole file | `#meta sim:on` |
| a whole run | `pofo -simulate` (`-b`) |
| a `/view` link | `sim=on` |

A holding with no reconstruction keeps its real quotes. `-no-simulate`
overrides all of them. The [data guide](data.md#backcasts) explains how the
past is rebuilt and graded.

## Directives: `#meta`

`#meta key:value` lines set per-portfolio options. Several may share a line,
separated by spaces.

| Directive | Effect |
|---|---|
| `rebalance:N` | rebalance to the target weights every N days; `0` = never. Overrides `-rebalance` |
| `sim:on` | backcast every holding (see above) |
| `extra-fees:X` | fees in %/yr on the whole portfolio (a wrapper, a mandate, a broker), deducted daily. Synonym: `envelope-fees` |
| `leverage:on` | keep the weights as written, up to 500 % in total; the rest is cash |
| `borrow-spread:X` | the financing spread in %/yr over cash when cash is negative (default `1.0`) |
| `capital:X` | the starting amount; required for flows, and it adds the money rows to the report |
| `contribute:A/P` | add A every period P: `week`, `month`, `quarter` or `year` |
| `withdraw:A/P` | take out A, or `A%` of the current value, every period P |
| `currencies:USD,EUR` | evaluate the portfolio in several currencies at once, one column each |
| `optimize:OBJ[,...]` | let an optimizer choose the weights (see [Optimizing weights](#optimizing-weights)) |

### Fees

Fund ongoing charges (TER) are **already in the prices**, so the simulation
never deducts them again. pofo looks each one up (and caches it for six
months) only to show it; `-no-fees` skips the lookup. `extra-fees` is
different: that cost is not in any price, so it is deducted.

### Leverage

With `leverage:on`, the weights stay as written and `100 - sum` becomes a cash
line. Positive cash earns the short rate (`^IRX`); negative cash is borrowed
at that rate plus `borrow-spread`. A portfolio whose value reaches zero is
ruined, and the report says so. Leverage does not combine with `optimize`.

```text
#meta leverage:on borrow-spread:0.8
90  VOO
60  IEF
```

### Contributions and withdrawals

Flows are invested or sold pro rata on the first trading day of each period.
The statistics stay on a time-weighted index, so flows do not distort
returns. The money rows (capital, contributed or withdrawn, final value and a
money-weighted IRR) follow the actual cash.

```text
#meta capital:10000 contribute:500/month
60  IWDA
40  AGGH
```

### Several currencies

`currencies:USD,EUR` turns one portfolio into one column per currency, each
with its own inflation deflator. The gap between the columns is the currency
risk. It does not combine with `optimize`.

## Optimizing weights

`#meta optimize:OBJ` lets an optimizer choose the weights. The report shows
`name (as written)` next to `name (OBJ)`, with the computed weights and their
statistics in a note. The weights are fitted on the past: read them as a
starting point, not a promise.

| `OBJ` | Long-only weights that... |
|---|---|
| `max-sharpe` | maximize return over volatility |
| `min-volatility` | minimize variance |
| `risk-parity` | equalize each holding's contribution to risk |
| `max-sortino` | maximize return over downside deviation |
| `return-to-drawdown` | maximize return over the maximum drawdown |
| `min-ulcer` | minimize the Ulcer Index (depth and duration underwater) |
| `max-worst-5y` | maximize the worst rolling five-year return |
| `max-return` | maximize CAGR; pair it with a limit, or it picks one line |
| `cwarp` | best diversify the benchmark (see [CWARP](#cwarp)) |
| `black-litterman` | start from what your own weights imply, then apply your views (see [Black-Litterman](#black-litterman)) |

`risk-parity` uses only the covariance, not past returns, so it chases the
backtest least. `max-sharpe` leans hardest on the past winner: always cap it.
For a withdrawal phase, `min-ulcer` and `max-worst-5y` target the discomfort
that matters there.

### Constraints

Constraints follow the objective, comma-separated:

| Constraint | Effect |
|---|---|
| `max-weight:25`, `min-weight:5` | cap or floor every line, in % |
| `bounds:NTSG:15-30` | a range for one line; repeat it per line; either end may be omitted (`bounds:GDE:-25`) |
| `max-vol:9.5` | a volatility cap, %/yr |
| `min-return:10.5` | a CAGR floor, %/yr |
| `max-drawdown:20` | a drawdown budget, % |
| `train:..2015` | fit on that window only (`START..END`, each end a year, a date, or empty) |

`risk-parity` and `cwarp` cannot enforce the three limits, so those
combinations are refused when the file is read. An unconstrained optimum is
usually a corner; bounds you can defend line by line belong in the problem.

A limit also asks the question most people mean. "The most return under 9.5 %
volatility" is sturdier than "the best Sharpe", because Sharpe here uses a
zero risk-free rate, which flatters any cash-like line.

### Judging out of sample: `train:`

With `train:`, the optimizer sees only that window, while the report measures
the weights over the whole period. From
[`optimized-constrained.txt`](../../examples/portfolios/optimized-constrained.txt):

```text
#meta optimize:max-return,max-vol:12,min-weight:5,bounds:NTSG:10-35,bounds:GOLD:5-25,train:..2015
```

```text
$ pofo -cli examples/portfolios/optimized-constrained.txt
optimized-constrained (max-return): weights computed by the optimizer (max-return
under vol ≤ 12.0 %) over 1996-03-27→2015-12-31: NTSG 10.0 %, ZPRV 23.8 %, DBMFE
30.4 %, GOLD 5.0 %, IDTL 30.8 %, in-sample CAGR 10.5 %/yr, volatility 12.0 %,
deepest drawdown -15.4 %, Sharpe 0.89; over 2016-01-01→2026-09-24, which it did
not see, CAGR 6.4 %/yr, volatility 10.6 %, deepest drawdown -20.5 %
```

Fitted, the weights promised 10.5 %/yr. On the ten years they never saw, they
delivered 6.4 %/yr with a deeper drawdown, and over the whole window the
weights as written did better (9.9 %/yr against 9.4). An optimizer run
without `train:` never shows you that gap.

### Black-Litterman

The objectives above read expected returns off the sample, which is the
quantity a price history estimates worst. `black-litterman` anchors them on
the portfolio you wrote instead. It works in three steps:

1. **Reverse optimization** turns your weights into the returns they
   implicitly expect, line by line.
2. Your **views** revise those returns, each with a confidence.
3. The weights maximize the result, under the same bounds and limits as any
   objective.

```text
#meta optimize:black-litterman,prior-return:4.6,view:IGLN:2,view:DBMFE>DTLA:3@70,bounds:IGLN:10-20
```

| Constraint | Meaning |
|---|---|
| `view:ID:Q@C` | ID earns Q %/yr, at C % confidence (50 by default) |
| `view:A>B:Q@C` | A beats B by Q points a year |
| `prior-return:R` | the return you expect from your weights as a whole; without it, a Sharpe of 0.4 is assumed |

With no view at all, you get your own weights back, exactly, and the note
lists the returns they imply. That half is often the more useful one. The
prior is your file rather than market capitalization, because trend funds,
stacked funds and gold have no capitalization weight.

Two cautions. The risk-free rate is zero, so state views as excess returns
over cash when that is what you mean. And the model prices the mean only: a
line held as a crisis hedge gets sized by its view alone, which tells you what
the hedge costs in expected return.

### CWARP

CWARP (Artemis Capital, 2020) scores whether an asset improves an existing
portfolio when layered on top at 25 % of its notional. Positive helps,
negative hurts. It combines the improvement in Sortino and in
return-to-drawdown, so it rewards low correlation and positive skew, where
Sharpe does not.

The reference portfolio is the benchmark (`-benchmark`, `^GSPC` by default).
The report shows a CWARP row for each portfolio and a CWARP column for each
holding.

`optimize:cwarp` finds the best **satellite** for the benchmark, not a good
standalone portfolio. It typically loads gold, long bonds and trend, which
look weak alone, so its standalone statistics look worse by design. Use it
when you already hold equity beta; use `max-sharpe` or `min-volatility` for a
complete allocation. It accepts `max-weight` only, and its solver is a
multi-start heuristic.

## From Go

`portfolio.Parse` reads the same format; `portfolio.NewSpec` builds a
portfolio in code, with weights as fractions. See the
[portfolios guide](library/portfolios.md).
