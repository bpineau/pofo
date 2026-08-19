# NTSG (Global Efficient Core) backcast: the global bond basket

## What NTSG is

`NTSG` = WisdomTree **Global** Efficient Core UCITS ETF (`IE00077IIPQ8`, USD
Acc), launched **2024-11-05**, TER **0.25 %/yr**. Like its US (`NTSX`) and
eurozone (`NTSZ`) siblings it is a capital-efficient **90/60**: 90 % equities +
60 % notional government bond futures + 10 % T-bill collateral, i.e. roughly a
1.5x-levered 60/40 in one line. What makes it different from both siblings is
that **neither leg is single-country**: the equity book is developed-markets
global and the bond sleeve is a basket of futures in four currencies.

The first reconstruction treated it as an American fund with an international
tilt (a frozen 60/40 US / developed-ex-US equity split, and the whole bond
overlay on one US intermediate Treasury fund). Both halves of that were wrong,
and this document records what replaced them and how each number was measured.

## What the fund actually holds

| Leg | The fund | What the reconstruction uses |
|---|---|---|
| Equity, 90 % | WisdomTree Global Efficient Core Index (Bloomberg `WTNTSGN`, net TR): a proprietary top-1500 developed cap-weighted universe over the same 21 countries as MSCI World, with ESG exclusions | MSCI World net TR: real **IWDA** (`IE00B4L5Y983`) from 2009-09, `MSCIWORLD-USD` + the `^990100-USD-STRD` daily shape before (1969-12) |
| Bonds, 60 % | government bond futures, **~80 % US / 11 % German / 6 % Japanese / 3 % British** notional, each country near a ten-year duration | four local excess-return sleeves at those weights (`pkg/simgen/globalbond.go`) |
| Collateral, 10 % | T-bills | `^IRX` (extended by `TBILL-3M` to 1934) |

The country notionals come from the audited statement of investments at
31/12/2025 (79.7 / 11.2 / 6.1 / 3.0), corroborated by the period averages of the
SFDR annex (77.9 / 11.0 / 5.9 / 3.0). They are rounded to 80/11/6/3.

## The two errors that were fixed, and what each was worth

Measured on the fund's own overlap (2024-11 to 2026-08, 436 trading days), on
the engine alone, before any real quote is grafted:

| Engine | CAGR vs the fund | daily corr | weekly | monthly | TE/vol |
|---|---|---|---|---|---|
| 0.54×VFINX + 0.36×VTMGX + 0.60×(VFITX−cash) + 0.10×cash | **+1.87 %/yr** | 0.41 | 0.86 | 0.94 | 1.12 |
| MSCI World equity leg, old US-only bond overlay | -0.77 %/yr | 0.92 | 0.98 | - | - |
| MSCI World equity leg + four-currency bond basket | **-1.65 %/yr** | **0.92** | **0.98** | **0.97** | **0.39** |

**The equity geography.** A frozen 60/40 US / developed-ex-US split is twelve
points of US underweight against the ~72 % the fund's own universe carries, and
2024-2026 happened to be a period when developed-ex-US beat the US by twelve
points a year: the old engine collected a large windfall from an error, which is
the worst way for a reconstruction to look right. Using MSCI World removes the
bet outright, and the repository already carries that index to 1969.

**The pricing calendar.** The fund is struck at a European valuation point. Its
old donors were US mutual fund NAVs struck hours later, and no amount of
weighting fixes a day-shifted series: daily correlation was 0.41. IWDA closes in
London, in the same session, and the daily correlation goes to 0.92. This is the
single largest effect in the table and it is a *calendar* fix, not a modelling
one.

**The bond basket.** See below. It costs 0.9 points a year over this particular
window, because 2024-2026 was a bond bear market and the corrected sleeve owns
more interest-rate risk than the old one did.

## The bond basket

### Local excess returns, no currency anywhere

A bond future is entered at no cost, so its holder earns the bond's return less
the financing of the notional, **in the bond's own currency**; and the fund rolls
the currency exposure of the foreign notionals away with spot/next forwards,
whose points are exactly the interest differential. What survives is the local
excess return and nothing else.

So the German sleeve is `BUND-EUR − euro overnight`, in euros, and it is never
converted to dollars and never given a carry term. Adding either would count the
same interest differential twice, once inside the forward points and once
outside them: the standard way to make a foreign bond sleeve look richer than it
was.

### The US sleeve's duration is measured, not assumed

The old overlay put the whole sleeve on `VFITX`, whose measured effective
duration is **4.90 years** (regressed on daily 5-year CMT yield changes,
2015-2026, 2869 observations). The fund's ladder is longer than that. Three
independent readings put the pair at **0.78 VFITX / 0.22 VUSTX**:

- regressing NTSX's own returns, less its 0.90 equity leg and its 0.10
  collateral leg, on the two donors' excess returns over 2018-08 to 2026-07
  gives 0.772 at both the weekly (399 points) and the monthly (95 points,
  R² 0.86) horizon;
- the same regression on NTSI (2021-05 on, which finances the same US ladder)
  gives 0.812 weekly and 0.833 monthly;
- `VUSTX`'s measured duration is 16.16 years (against the 30-year CMT), so a
  0.78/0.22 blend runs **7.4 years**, which is both what an equally weighted
  2/10/30-year Treasury futures ladder carries and the effective duration
  WisdomTree publishes for the sleeve (~7 years).

The caveat that belongs with the number: the donors' durations are not constant
back to 1953. Their deep tails are constant-maturity reconstructions of a 5-year
and a 20-year par bond, whose durations move with the yield level, so the blend
runs a little short in the high-yield 1970s and a little long in the 2010s.

### Where each sleeve starts, and what happens before

The weights are **renormalized over the sleeves that quote**, day by day, so the
overlay always carries a full notional. The British sleeve opens in 1978 (its
financing rate does) and the Japanese one in 1986-07 (its yield series does);
before those dates their share is carried by the sleeves that exist rather than
by an invented proxy. Substituting a Treasury for a missing JGB would say the
opposite of what a multi-currency basket is for.

The composite's own floor is ~1969-12, set by the equity leg, so the deepest
years of the file run on 88 % of the basket, then 91 %, then all of it.

## New bundled reference series (`cmd/gen-gbond-refdata`, `make gbond-refdata`)

| Refdata (new) | Content | Source | Span |
|---|---|---|---|
| `BUND-EUR` | German govt bond TR (10y benchmark, monthly) | OECD long-term yield `DEU.M.IRLT` (`DSD_STES@DF_FINMARK`) → `TreasuryTR` (10y) | 1956-05 |
| `BUND-DAILY` | the same, daily shape | Bundesbank daily term structure of listed federal securities, 10y point (`BBSIS`) → `TreasuryTR` | 1997-08 |
| `JGB-JPY` | Japanese govt bond TR (10y benchmark, **daily**) | Japanese Ministry of Finance `jgbcme_all.csv`, 10Y column → `TreasuryTR` | 1986-07 |
| `GILT-GBP` | British govt bond TR (10y benchmark, monthly) | OECD long-term yield `GBR.M.IRLT` → `TreasuryTR` (10y) | 1960-01 |
| `JPCASH-JPY` | yen call-money accrual (monthly) | OECD immediate rate `JPN.M.IRSTCI`, compounded | 1985-07 |
| `GBCASH-GBP` | sterling interbank accrual (monthly) | OECD immediate rate `GBR.M.IRSTCI`, compounded | 1978-01 |

Two source notes that cost time to find:

- the OECD series must come from the **current** short-term-statistics dataflow
  `OECD/DSD_STES@DF_FINMARK`. The legacy `OECD/MEI` dataset stopped being
  updated in **2024-01** while still answering HTTP 200;
  `cmd/gen-euro-refdata` was moved off it in 2026-08 and `cmd/gen-macropanel`
  is now the last generator reading it;
- the Ministry of Finance CSV is only served on the `/english/` path. It carries
  two header rows, dates as `YYYY/M/D` with no zero padding, and a bare `-`
  wherever a tenor did not quote: the table opens in 1974 but its 10-year column
  only opens in 1986-07.

The euro-area aggregate `EUROGOV-EUR` is deliberately **not** reused for the
German sleeve: it smears the periphery spreads of 2011-2012 that a Bund basket
never carried.

### The validation the generator runs on itself

`gen-gbond-refdata` refuses to write anything until five checks pass (house
rule: a series that downloaded cleanly has proved nothing). Measured on
2026-08-08:

| Check | Result |
|---|---|
| `BUND-EUR` vs `EUROGOV-EUR`, 1999-2010, when euro spreads were thin | CAGR 4.67 % vs 4.41 % (gap +0.26 pts/yr), monthly corr **0.955** |
| `BUND-DAILY` vs `BUND-EUR` volatility, 1998-2026 | 6.07 % vs 5.05 %/yr (ratio 1.20; the OECD publishes a monthly *average* yield, which damps the month) |
| `JGB-JPY` under yield-curve control, 2016-2021 | CAGR +0.44 %/yr, vol **2.02 %/yr**: quiet and flat, as a pinned bond must be |
| `GILT-GBP` in the 1970s | CAGR +7.95 %/yr, vol 10.51 %/yr: positive in nominal terms, far behind that decade's inflation |
| `JPCASH-JPY` / `GBCASH-GBP` | 1.22 %/yr and 5.72 %/yr over their whole spans, deepest drawdowns -0.32 % and 0.00 % |

Whole-span CAGRs, for the record: `BUND-EUR` 5.83 %/yr (1956-2026), `GILT-GBP`
7.10 %/yr (1960-2026), `JGB-JPY` 2.48 %/yr (1986-2026).

## A negative yield is a price, not a gap

`simgen.TreasuryTR` used to treat a yield at or below zero as missing data and
carry the index flat through it. That is unusable for Japan: the 10-year
benchmark closed **below zero on 453 days between 2016-02 and 2020-05**, and
those were years in which the bond returned a great deal. The par-bond formula
needs no floor (a bond whose coupon equals its negative yield still prices at
exactly 100); zero is the annuity factor's one singular point, and its limit
there is the undiscounted sum of the coupons plus the face. The guard is now an
epsilon at zero and a refusal below -100 %/yr, nothing else.

This also means the bundled euro series would gain from a refresh: the ECB AAA
curve behind `EUROGOV-DAILY` spent years below zero and that file, generated
before the fix, still flat-lines those days. `make euro-refdata` will correct it;
it was left out of this change so that NTSZ's behaviour would not move at the
same time.

## Independent cross-check: NTSI

`NTSI` (WisdomTree International Efficient Core, US-listed, 2021-05 on) is 90 %
developed-ex-US equity + 60 % of the same overlay + 10 % collateral, and it has
**five years** of record against NTSG's 1.7. Rebuilding it as
`0.90×VTMGX + 0.60×overlay + 0.10×cash` over its whole history (1308 trading
days):

| Overlay | CAGR engine vs real | daily corr | weekly | TE |
|---|---|---|---|---|
| old, 0.60×(VFITX−cash) | 6.99 % vs 6.67 % (**+0.32**) | 0.965 | 0.987 | 4.1 %/yr |
| US-only but duration-matched (0.78/0.22) | 6.29 % vs 6.67 % (-0.38) | 0.966 | 0.987 | 4.1 %/yr |
| the four-currency basket | 6.22 % vs 6.67 % (**-0.45**) | **0.967** | 0.987 | **4.0 %/yr** |

Read it honestly: on level, over this particular five-year bond bear market, the
duration correction moves the error from +0.32 to -0.38 points a year, which is
a wash in magnitude and a sign flip. On path it is a small improvement. The
duration blend is kept because the evidence for it is a 95-month regression with
R² 0.86 plus the fund's own published duration, and because a level test run over
a single rate regime cannot arbitrate a duration: it can only tell you which way
rates went.

NTSI is *not* wired into `ValidateAgainst`. It is a different fund with a
different equity book, and the check belongs in this document, not in the grader.

## Known limitations (the ledger)

- **Validation is thin.** The fund has 1.7 years of quotes. The `-1.65 %/yr`
  level gap is one bond bear market wide and the audit flags the window. The
  five-year NTSI check above is the better level evidence.
- **The index is a proxy.** Plain MSCI World stands in for a screened top-1500
  universe. MSCI's own screened-vs-plain comparison runs +0.36 %/yr with 0.81 %
  tracking error over ten years, so the substitution is conservative rather than
  flattering, and the fund's published book is ~3 points lighter in the US than
  the plain index for the same reason. Over 2024-2026, when ex-US led, that
  alone accounts for roughly 0.3 points a year of the remaining gap.
- **The British sleeve is monthly.** No daily gilt curve is bundled: the Bank of
  England serves its history as a workbook, and 1.8 % of net assets does not
  justify a parser. The sleeve therefore steps once a month, which is visible in
  the engine's variance ratio and in nothing else.
- **The country weights are a snapshot.** 80/11/6/3 is the basket at end-2025,
  held constant across fifty-six years. The fund's own weights drift; there is no
  history of them.
- **The monthly reference series are monthly *averages*.** OECD publishes an
  average-of-days yield, which damps within-month volatility by roughly a fifth
  (measured above on the German pair). The daily shapes cover 1997 on for
  Germany and 1986 on for Japan; the sterling sleeve and the deep German tail
  carry that damping.

## Regeneration

```sh
make gbond-refdata   # rebuild BUND-EUR{,-DAILY} / JGB-JPY / GILT-GBP / JPCASH-JPY / GBCASH-GBP (network)
make simdata         # rebuild pkg/datasets/simdata/, including IE00077IIPQ8
```

Both are part of `make refresh`, in that order.
