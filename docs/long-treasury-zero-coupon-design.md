# The long zero-coupon Treasury reconstruction (ZROZ, TREASURY-LONG-YIELD)

Dated record of the 2026-09-17 rework of the 25+ year Treasury STRIPS backcast:
why the geared coupon fund it used to be could not work outside the window it
was fitted on, what replaced it, the validation it passed and the bounds it
ships under. The engine's godoc (`go doc ./pkg/simgen` on `TreasuryZeroTR`),
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
annualized percent, 16 266 observations, 1953-04-30 to date, business-daily from
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
`y30 = +0.2126 + 0.9638 * y20`, least squares over the 10 702 days the two
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
