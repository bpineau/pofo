# The US total-market reference (`USMKT-USD`) and the VTI backcast

## The problem

A total-market fund (VTI, and any share class of the same portfolio) quotes
from 2001. Everything before that has to be reconstructed, and the
reconstruction used to be the S&P 500 (`VFINX`, itself carried back by
`SP500-USD`), on the ground that the two track at about 0.99 daily correlation.

A correlation is the wrong test for a LEVEL. The whole market is the S&P 500
plus the mid, small and micro completion, and that completion's premium is
worth roughly a point a year with a sign that changes by era:

| window | S&P 500 | total market | difference |
|---|---|---|---|
| 1992-04 to 2001-06 (the two funds' own NAVs) | 14.93 %/yr | 13.88 %/yr | -1.05 pts/yr |
| 1962-01 to 2001-06 (market factor vs the index) | 10.43 %/yr | 11.04 %/yr | +0.61 pts/yr |

Substituting the large-cap index for the whole market therefore does not add
noise, it adds a bias whose direction the backcast cannot know in advance.

## The reference

`USMKT-USD` (`cmd/gen-usmkt-refdata`, `make usmkt-refdata`) is the daily total
return of the whole US market since 1926-07, cumulated from the Fama/French
market factor: `Mkt-RF`, the excess return of the CRSP value-weighted portfolio
of all NYSE, AMEX and NASDAQ common stocks, plus `RF`, the one-month Treasury
bill return of the same day. The Ken French Data Library already stands behind
two bundled series (`USSCV-USD`, `DEVEXUS-DAILY`), so this adds a source of the
same grade rather than a new one.

It is an ACADEMIC FACTOR: gross of fees, commissions and spreads, like
`USSCV-USD` and unlike a fund NAV. What it owes before it may stand in for a
fund is charged once, by `longBackFee` in `pkg/simgen/extend.go`
(`usmktGrossCost`), never inside the file.

### Validation, before the file is written

Two references that know nothing about the factor, both mandatory (a failure
writes nothing and says why):

- the bundled S&P 500 total return (`SP500-USD`), on CALENDAR-YEAR returns over
  the whole span: same market minus a completion tail, so the years must move
  together and the means must be close. Measured: 100 years, return correlation
  0.9906, mean year -0.22 pts (the market slightly BELOW the large-cap index
  over the century, the 1930s and the post-1980 decades pulling that way; the
  largest single year is 1929, at 6.5 pts).
- the real NAVs of the fund the series extends (`VTSMX`), monthly over the
  whole overlap: same portfolio minus a wrapper. Measured: 411 months, 1992-04
  to 2026-07, monthly correlation 0.9992, 11.02 %/yr against the fund's 10.73,
  a grossness of +0.29 pts/yr. A factor that LAGGED the fund it extends would
  be refused outright, as would one more than a point a year above it.

Plus the structural checks any bundled series gets: at least 20 000 daily
returns, a first date inside the library's own first month, ascending dates,
positive levels, no single day beyond 30 % (1987-10-19 lost 20 %), and a last
return no older than 200 days (the library republishes monthly, with a lag).

### The grossness constant

`usmktGrossCost` = 0.3 %/yr, the full-overlap figure rounded, following
`USSCVGrossCost`'s doctrine. About half of it is the donor's own price list
(the Investor class charged 0.20 %/yr in the 1990s and charges 0.14 now, and
the target ETF charges 0.03), the rest being commissions, spreads and the CRSP
tail of micro caps no fund replicates share for share. The overlap does not
identify the split and the constant does not pretend it does. Per decade the
gap runs +0.80 (1992-1999), -0.12 (2000s), +0.30 (2010s), +0.56 (2020s): a
swing no stability criterion accepts as a trend, so no era is fitted.

## The recipe

`vtiRecipe` now reads the target's OWN portfolio all the way down: `VTSMX`
(the Investor share class of the fund VTI is a class of, 1992-04), extended by
`USMKT-USD` through `longBack`, with real VTI grafted from 2001-06.

Graded without its graft against the real fund (`pofo -verify-simdata VTI`,
2001-06 to 2026-09, 25.3 years):

| engine | CAGR gap | cumulated drift | monthly correlation | TE |
|---|---|---|---|---|
| `VFINX` (the S&P 500, incumbent) | -0.31 pts/yr | -6.99 % | 0.9952 | 2.39 %/yr |
| `VTSMX` + `USMKT-USD` | -0.11 pts/yr | -2.80 % | 0.9994 | 1.99 %/yr |

Better on the level and on the path at once, which is the bar a donor change
has to clear.

## Bounds and what is deliberately not done

- The bundled `VTI.csv` still starts at 1962-01, like every other
  reconstruction, because `simgen.ComponentsFrom` is where component histories
  are requested from. The reference reaches 1926-07 and any caller asking for
  more gets it; moving that global floor is a separate decision that would move
  every file in `pkg/datasets/simdata`.
- `VFINX` stays the S&P 500 donor of the recipes that MEAN the S&P 500 (`SP500`,
  `GDE`, `NTSX`, `RSST`, the world blends' US leg). Nothing about those funds is
  a total-market claim, so nothing there changes.
- The factor is not used as a benchmark asset of its own. It is a donor: it
  carries no fund's fee, and a catalog entry would invite it into portfolios
  where a fee-free market is not a thing anyone holds.
