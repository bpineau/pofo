# The long Treasury references (ZROZ, TREASURY-LONG-YIELD, the two total-return series)

Dated record of two reworks of the bundled US Treasury history, five days apart
and on the same foundation.

The first (2026-09-17) is the 25+ year STRIPS backcast: why the geared coupon
fund it used to be could not work outside the window it was fitted on, what
replaced it, the validation it passed and the bounds it ships under.

The second (2026-09-19) is the pair of constant-maturity TOTAL-RETURN series the
Vanguard Treasury donors are extended with, `TREASURY-LONG-USD` and
`TREASURY-INT-USD`: one of them had a seven-year hole and both were dated and
sampled wrong. That section is at the bottom.

The engine's godoc (`go doc ./pkg/simgen` on `TreasuryZeroTR` and `TreasuryTR`),
`pkg/simgen/strips.go` and the generator's godoc
(`go doc ./cmd/gen-tyield-refdata`) are the operational descriptions; this file
keeps the evidence.

## The defect

`ZROZ` (PIMCO 25+ Year Zero Coupon U.S. Treasury, real quotes from 2009-11,
published effective duration 26.0) was reconstructed as

    cash + 1.65 x (VUSTX - cash)

a long COUPON mutual fund geared over the bill rate. The multiple is the ratio
of the strip's duration to the donor's, and that ratio is not a constant: a
coupon bond's duration SHRINKS as its yield rises, because the early coupons
weigh more in the present value, while a zero's duration is its maturity at any
yield. Modified durations, semiannual convention:

| Long yield | 22y par bond | 27y strip | ratio |
|---|---|---|---|
| 2 % | 17.7 | 26.7 | 1.51 |
| 3 % | 16.0 | 26.6 | **1.66** |
| 5 % | 13.3 | 26.3 | 1.99 |
| 7 % | 11.1 | 26.1 | 2.34 |
| 9 % | 9.5 | 25.8 | 2.72 |
| 12 % | 7.7 | 25.5 | **3.31** |
| 15 % | 6.4 | 25.1 | 3.93 |

Against the eras of the bundled long-yield history (mean long yield, and the
gearing it calls for):

| Era | Mean long yield | g(y) | 1.65/g |
|---|---|---|---|
| 1953-1959 | 3.3 % | 1.70 | 0.97 |
| 1960-1969 | 4.9 % | 1.92 | 0.86 |
| 1970-1979 | 7.6 % | 2.45 | 0.67 |
| **1980-1985** | **12.0 %** | **3.34** | **0.49** |
| 1990-1999 | 7.0 % | 2.27 | 0.73 |
| 2000-2009 | 5.0 % | 1.99 | 0.83 |
| **2010-2019** | **3.2 %** | **1.65** | **1.00** |
| 2020-2026 | 3.5 % | 1.74 | 0.95 |

1.65 is exactly right for the 2010s, which is the decade it was fitted on, and
it is HALF of what 1980-1985 demands. The file's live window could not see it:
over 2009-2026 the geared form and a properly priced strip agree closely, which
is what a multiple fitted there does. Outside it the file reported an instrument
roughly half as risky as the one it names.

## What ships

**`pkg/datasets/refdata/TREASURY-LONG-YIELD.csv`** (`make tyield-refdata`,
`cmd/gen-tyield-refdata`): the US long Treasury constant-maturity PAR yield,
annualized percent, 16 268 observations, 1953-04-30 to date, business-daily from
1962-01-02 and monthly before. A rate, not a price. Three segments of the
Federal Reserve's H.15 selected interest rates, read through the DBnomics
mirror:

- the spine, 1977-02-15 onwards: the 30-year constant maturity, business daily.
  The H.15 row is continuous across the 2002-02 to 2006-02 suspension of the
  30-year issue, where the figure is the Treasury's long-term (25 years and
  above) average instead of a 30-year bond.
- 1962-01-02 to 1977-02-14: the 20-year constant maturity, business daily.
- 1953-04 to 1961-12: the 20-year constant maturity, monthly.

The two 20-year segments are mapped onto the spine's curve point by
`y30 = +0.2126 + 0.9638 * y20`, least squares over the 10 704 days the two
points overlap (residual sd 0.157 pt), then shifted by +0.143 pt so the mapped
head meets the spine exactly at the junction. The map is not cosmetic: the long
end is not parallel to the 20-year point and was INVERTED through the high-rate
years (the 30-year ran 0.15 pt BELOW the 20-year when the 20-year was above 9 %,
and 0.11 pt above it below 5 %), so a constant spread would have the sign
backwards in the era the series exists to cover.

**`simgen.TreasuryZeroTR`**: the constant-maturity zero-coupon engine, the
sibling of `TreasuryTR` and sharing its loop (`constantMaturityTR`). Each step
holds a fresh T-year zero priced at `(1+y/2)^(-2T)`, ages it by the step's
length and reprices it at the next yield; carry comes out as the pull to par,
there is no coupon and nothing to reinvest. The semiannual convention is the one
Treasury yields are quoted in and it matters at this maturity: discounting
annually would put the modified duration at n/(1+y) rather than n/(1+y/2), 24.1
instead of 25.5 years at a 12 % yield.

**The `ZROZ` recipe**: a constant 27-year strip on that yield, net of the fund's
own 0.15 %/yr, with real quotes grafted from 2009-11. 27 years is the average
maturity of the 25 to 30 year paper the fund's index holds, an a-priori figure
and not a fitted one; at the yields of the live window it is a modified duration
of 26.5 against the 26.0 the fund publishes. The cash leg the geared form needed
is gone: a strip owned outright earns no bill rate on top of itself.

## Validation

Month-end returns, engine WITHOUT its real graft, against the real funds over
the whole window each shares with it (`pofo -verify-simdata ZROZ` reports the
same thing on daily data).

| Reference | Engine | Months | Monthly corr | Vol engine/real | Level gap |
|---|---|---|---|---|---|
| ZROZ | old, 1.65x VUSTX | 202 | 0.978 | 0.954 | -0.14 %/yr |
| ZROZ | **27y strip** | 202 | **0.988** | 0.943 | -0.21 %/yr |
| EDV | old, 1.65x VUSTX | 224 | 0.970 | 0.986 | -0.12 %/yr |
| EDV | **27y strip** | 224 | **0.990** | 1.035 | -0.39 %/yr |

`EDV` (Vanguard Extended Duration Treasury, 20-30 year STRIPS, real from
2008-01, published duration ~24.3) is the second reference, and it is the more
telling of the two: the incumbent multiple was never fitted to it, and one
a-priori maturity now fits both funds' shapes at 0.99.

What moved, and what did not:

- **Shape improves on both funds**, monthly and weekly (daily correlation goes
  the other way, 0.965 to 0.958: a CMT yield is snapped in the afternoon while
  the fund closes at four, so a one-day mismatch is expected and washes out at
  weekly, 0.972 to 0.985).
- **Level and volatility are a fraction of a point worse on the window the
  incumbent was fitted to**, which is the window in the table. That is the price
  of not fitting: 1.65 was chosen there, and one constant maturity now has to
  serve two funds and seventy years.
- The level gap's SIGN is the one the approximations predict. The engine
  discounts a strip at a PAR yield, while a strip trades at a zero rate, which
  on an upward-sloping curve sits above the par yield of the same maturity; and
  the constant-maturity convention reprices the aged strip at its original curve
  point rather than the shorter one it rolled down to. Both omit return, and the
  engine indeed runs COLD by 0.21 pt/yr against ZROZ and 0.39 against EDV. The
  gap is NOT corrected: a constant added to the carry to close a measured return
  gap is the fudge this repository refuses everywhere else.
- The residual volatility shortfall is not a duration misfit and must not be
  tuned away. At the duration each fund publishes, ZROZ's own monthly returns
  imply a yield path 4.0 % more volatile than this series' and EDV's 1.4 % more,
  which is what a par-to-zero transform does at the long end. Lengthening the
  maturity to 28 years would put the ZROZ vol ratio at 0.978 and the EDV one at
  1.074, i.e. buy one fund's volatility by overshooting the other's, while both
  funds' published durations are SHORTER than 27 years, not longer.

## What the deep past now looks like

Month-end statistics of the engine, against the incumbent's own shipped file
over the same windows (its pre-2009 segment IS the incumbent engine):

| Window | Vol, strip | Vol, incumbent | Worst drawdown, strip | Worst drawdown, incumbent |
|---|---|---|---|---|
| 1950s (from 1953-04) | 7.3 % | n/a | -23 % | n/a |
| 1960s | 10.7 % | 6.7 % | -37 % | -28 % |
| 1970s | 19.7 % | 11.2 % | -44 % | -30 % |
| 1980s | 44.6 % | 19.8 % | -69 % | -40 % |
| 1990s | 20.3 % | 12.8 % | -37 % | -20 % |
| 2000s | 22.9 % | 16.6 % | -38 % | -22 % |
| 2010s | 19.2 % | 21.4 % | -28 % | -29 % |
| 2020s | 21.0 % | 20.8 % | -61 % | -62 % |
| 1953 to 2009-11 | 25.0 % | 14.7 % | **-80 %** | **-46 %** |
| 2002 to 2009-11 | 23.5 % | 17.3 % | -35 % | -19 % |

The two eras where the fund's own quotes govern (the 2010s and 2020s) agree to
within a point of volatility and a point of drawdown, which is the control: the
engine is not globally hot, it is the deep past that was cold.

The -80 % is arithmetic, not a surprise. The long yield went from 4.1 % in
January 1962 to 15.2 % in September 1981; a 27-year zero repriced from 4.1 % to
15.2 % loses 94 % of its price, and nineteen years of carry at 4 to 15 % gives
back a factor of about four, which lands near -77 % before the monthly path's
own peak is accounted for. No reconstruction of a 26-duration zero can come out
of that era down 46 %.

## Bounds and known residuals

- The file starts 1953-04, where H.15's 20-year point starts. The head is
  MONTHLY before 1962-01: 105 month-end observations, mapped from the 20-year
  point, so per-observation statistics over a window that straddles 1962 mix two
  cadences. Read the monthly columns there.
- Par yield in, zero rate out, and no roll-down: both understate the return, by
  0.2 to 0.4 pt/yr as measured above. Closing them honestly needs a zero curve,
  which cannot be bootstrapped from two par points.
- `-0.15 %/yr` of fee is charged on a fee-free reconstruction, which is correct
  here (the yield series carries no manager's charge) and is the fund's own
  published ongoing charge, never a residual.
- **Left alone, measured**: the coupon-to-coupon gearing the 20+ year ETFs use
  (`longTreasuryGearing`, 17/15) has the same yield dependence in principle, but
  between two coupon bonds it drifts only from 1.11 at 3 % to 1.04 at 12 %, a
  7 % effect rather than a factor of two, so `DTLA`/`DTLE` stand as they are.
- **Named, not fixed**: `marketdata.proxyFor` maps `EDV` to `VUSTX` with no
  gearing at all, which gives a 24-duration STRIPS fund the history of a
  16-duration coupon fund. `EDV` has no catalog record to hang a recipe on, so
  the strip engine is not wired to it; anyone adding that record should give it
  a `TreasuryZeroTR` recipe at a ~25-year maturity (monthly correlation 0.990,
  vol ratio 0.957 as measured above) rather than the proxy.

## The two total-return references, rebuilt 2026-09-19

`TREASURY-LONG-USD` and `TREASURY-INT-USD` are the deep proxies behind the two
Vanguard Treasury donors (`VUSTX` 1986-05, `VFITX` 1991-10) and, for the
intermediate one, the bond sleeve of the bundled US 60/40 the withdrawal replay
runs on (`pkg/replay`). Both were built from FRED's MONTHLY constant-maturity
series (`GS20`, `GS5`) through `simgen.TreasuryTR`. Both carried a defect, and
the long one carried two.

### The hole

H.15 suspended the 20-year constant maturity between 1987-01 and 1993-09. The
long file was priced off that point alone, so it had no row at all over those
seven years: it went from 1986-12 (525.92) straight to 1993-10 (840.54). Any
direct consumer read one +60 % "month" and then seven years of no volatility,
and the numbers that come out of that look plausible. It is why a FIRE book
plate had to rebuild its own long leg (`pkg/firebook/figures_matiere.go`), and
it is why `pkg/datasets/golden/gaps_test.go` now exists.

The file's own consumers did not see it, which is how it survived: the splice
behind `VUSTX` only uses the proxy BEFORE 1986-05, and every golden measures
calendar-year returns or a fifty-year CAGR, both of which a hole leaves nearly
untouched (the 1972-2021 CAGR read 8.34 %/yr with the hole and 7.63 without,
against the SBBI-era 7.6 the golden asserts).

### The dating

A FRED monthly yield is the month's AVERAGE of the daily observations, stamped
on the FIRST of the month. Both errors push the same way: the average puts every
turning point half a month early, and the label another half month early again.
What comes out has the right calendar-year returns and the wrong months. Against
the funds these files stand behind, on monthly returns:

| Series vs fund | monthly corr | vol ratio | CAGR gap |
|---|---|---|---|
| long, GS20 month-average, vs `VUSTX` 2001-2026 | 0.724 | 0.827 | -0.10 pt/yr |
| long, month-end, vs `VUSTX` 2001-2026 | **0.975** | 0.953 | -0.10 pt/yr |
| int, GS5 month-average, vs `VFITX` 2001-2026 | 0.717 | 0.743 | -0.91 pt/yr |
| int, month-end, vs `VFITX` 2001-2026 | **0.977** | 0.906 | -0.93 pt/yr |

The window is deliberately the one with no hole in it, so that what is measured
is the dating alone.

### What ships

Both series are now written by `cmd/gen-tyield-refdata`, which already owned the
long par yield, off the same H.15 points read through DBnomics and through the
same engine. One point per calendar month, each on the month's LAST quoted
yield, gap-free from 1953-04, 0.10 %/yr as before.

- **`TREASURY-INT-USD`**: a 5-year par bond on the H.15 5-year point
  (`RIFLGFCY05_N.B` business-daily from 1962-01-02, `RIFLGFCY05_N.M` monthly
  before). One curve point throughout: no map, no splice, nothing to document
  but the cadence.
- **`TREASURY-LONG-USD`**: a 20-year par bond on the bundled
  `TREASURY-LONG-YIELD`, month-end.

The long one needs its decision stated plainly, because the bond and the curve
point it is discounted at are not the same tenor. The BOND stays 20-year, which
is the maturity of the published long-term government bond record the goldens
are anchored on and the neighbourhood of what `VUSTX` holds; it is not fitted to
anything. The CURVE POINT is the long one this repository already assembles and
validates: the 30-year par yield from 1977-02, the 20-year point mapped onto it
before. So across 1987-01..1993-09 the series prices a 20-year bond at the
30-year yield, which is exactly the trade that removes the hole, since the
30-year row publishes without interruption there.

What that costs is measurable from the map itself (`y30 = +0.2126 + 0.9638·y20`,
10 704 overlapping days, residual sd 0.157 pt): the discount rate sits
`+0.2126 − 0.0362·y20` from the published 20-year point, i.e. +0.07 pt at a 4 %
yield and -0.22 pt at 12 %, averaging near zero over the record, and its moves
are 0.964 as large. Both are small next to a 10 %/yr volatility, and neither is
corrected. The alternative, inverting the map to synthesize a 20-year yield
everywhere, was refused: it would put a fitted regression between the file and
the published yield over the sixty years where the real 20-year point IS
published, to gain nothing outside the seven where it is not.

### Validation

Month-end returns against the real funds, over the whole window each shares with
the reconstruction. `VUSTX` is the reference that matters, because it quotes
from 1986-05 and therefore covers the whole of the former hole.

| Reference | Window | File | Months | Monthly corr | Vol ratio | CAGR gap |
|---|---|---|---|---|---|---|
| `VUSTX` | full overlap | old | 399 | 0.882 | 0.679 | -0.70 pt/yr |
| `VUSTX` | full overlap | **new** | 483 | **0.974** | **0.985** | **-0.06 pt/yr** |
| `VUSTX` | 1986-2000 | old | 94 | 0.964 | 0.620 | -2.89 pt/yr |
| `VUSTX` | 1986-2000 | **new** | 175 | **0.977** | 1.098 | **+0.09 pt/yr** |
| `VUSTX` | 2001-2026 | old | 304 | 0.724 | 0.827 | -0.10 pt/yr |
| `VUSTX` | 2001-2026 | **new** | 307 | **0.975** | **0.953** | -0.10 pt/yr |
| `TLT` | full overlap | old | 286 | 0.731 | 0.732 | +0.02 pt/yr |
| `TLT` | full overlap | **new** | 289 | **0.990** | **0.840** | +0.09 pt/yr |
| `VGLT` | full overlap | old | 196 | 0.757 | 0.787 | -0.15 pt/yr |
| `VGLT` | full overlap | **new** | 199 | **0.992** | **0.886** | **-0.00 pt/yr** |
| `VFITX` | full overlap | old | 415 | 0.702 | 0.753 | -0.75 pt/yr |
| `VFITX` | full overlap | **new** | 418 | **0.971** | **0.891** | -0.78 pt/yr |

The old file's 1986-2000 row is what a hole does to a statistic: 94 months
instead of 175, an apparent 36 %/yr of fund volatility because one "month" spans
seven years, and a level gap of -2.89 pts/yr that means nothing at all.

The intermediate file's level gap does not move (-0.75 to -0.78 pt/yr) and is
not a defect of the dating: `VFITX` is a 5-10 year fund and the reconstruction
is a 5-year one, deliberately the shorter of the two, and the same shortfall is
what `intTreasuryGearing` exists to correct where a recipe needs the longer
segment.

Calendar years, the rebuilt long file against the two real long funds:

| Year | old | new | `VUSTX` | `TLT` |
|---|---|---|---|---|
| 1987 | (no row) | -5.64 | -3.17 | n/a |
| 1990 | (no row) | +5.86 | +5.88 | n/a |
| 1994 | -8.91 | -8.63 | -6.99 | n/a |
| 1995 | +30.33 | +31.19 | +30.21 | n/a |
| 2008 | +25.76 | +32.94 | +22.55 | +33.95 |
| 2009 | -12.27 | -21.17 | -11.98 | -21.81 |
| 2013 | -13.37 | -10.62 | -13.03 | -13.38 |
| 2022 | -23.59 | -24.77 | -29.57 | -31.23 |

1987 and 1990 are the hole. 2008 and 2009 are the useful pair: the rebuilt
series turns in +32.94 then -21.17 where `TLT` did +33.95 then -21.81, while
`VUSTX` did half of each, which says the 20-year par bond is the longer of the
two funds at the extremes even though its full-record volatility lands on
`VUSTX`'s (ratio 0.985 against `VUSTX`, 0.840 against `TLT`).

One external published series was fetched to grade the ENGINE rather than the
maturity: Damodaran's "Annual Returns on Stock, T.Bonds and T.Bills",
`pages.stern.nyu.edu/~adamodar/New_Home_Page/datafile/histretSP.html`, read
2026-09-19, whose T.Bond column is a 10-YEAR constant-maturity total return from
1928. Rebuilding a 10-year par bond on the H.15 10-year point and comparing the
63 calendar years 1963-2025: mean absolute error 0.97 pt under the old
month-average convention and 1.02 pt under the new month-end one, with the same
+0.33 pt mean error. That is the honest result and it is the reason the goldens
never caught the dating flaw: at the calendar-year frequency the two conventions
are indistinguishable, because a half-month smear at both ends of a
December-to-December window cancels. The discrimination is monthly, and it is
the table above.

### What moved downstream

- Every recipe with a `VUSTX` or `VFITX` leg regenerated (`make simdata`, 59
  files): the change only reaches the years before those funds' own quotes, so
  no audit verdict moved (`pofo -verify-simdata TLT IDTL IEF DTLETR ZROZ NTSG`
  reads level ok / path ok on TLT and IEF as before).
- The bundled real US 60/40 (`pkg/replay`) barely moves in aggregate: 72 years
  1954-2025 at 5.533 %/yr real before and 5.529 after, yearly volatility 11.19
  against 11.20, worst year 1974 at -22.78 against -22.79. The fix redistributes
  return WITHIN years, not across the record. One `Example` output moved by a
  thousand euros of residual estate on a knife-edge vintage.
- Fifteen FIRE book plates now lag their data, all by fractions of a point (see
  `make figure-drift`). Six of them (`TestMourirRiche…`, `TestReplayFigures…`,
  `TestReplayTables…`, `TestMillesimes…`, `TestBengen…`, `TestSmoothing…`) were
  not behind `frozenAgainstData` although they recompute frozen numbers from the
  bundled data; they are now, per the standing rule that a refresh must never
  require a code change.
