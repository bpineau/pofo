# NTSZ (Eurozone Efficient Core) backcast: design and epistemic ledger

## What NTSZ is

`NTSZ` = WisdomTree **Eurozone** Efficient Core UCITS ETF (`IE000OV4XWA3`, EUR
Acc), launched **2025-09-30**. It is the eurozone sibling of the US (`NTSX`) and
global (`NTSG`) Efficient Core funds: a capital-efficient **90/60** portfolio
(90% eurozone equities + 60% notional euro-denominated government bond futures),
i.e. a ~1.5x-levered eurozone 60/40 in a single fund, TER **0.20%/yr**.

Because the fund is only months old, a backcast is the whole point: it lets the
CLI compare `NTSZSIM` against `NTSXSIM`/`NTSGSIM` over a common multi-decade
history.

## The structural difference from NTSX/NTSG

NTSX reaches back to 1953 by leaning on long-running **US** index funds
(Vanguard) and USD refdata (S&P 500 to 1871, CMT Treasuries to 1953), and NTSG to
1969 on the MSCI World reconstruction plus a four-currency bond basket
(`docs/ntsg-global-efficient-core-design.md`). NTSZ is **euro-native end to end**, and no comparable deep euro building blocks
existed in the repo. So the deep tail is assembled from four new bundled
reference series, all sourced from **DBnomics** (free, key-less; the same mirror
the macro panel uses) by `cmd/gen-euro-refdata`. The pofo binary never fetches
OECD or the ECB; it embeds the CSVs.

| Refdata (new) | Content | Source | Span |
|---|---|---|---|
| `EMU-EUR` | eurozone equity **net TR** | OECD euro-area share-price index `EA20.M.SHARE` (price) grossed by a constant net dividend yield | ~1986 |
| `EUROGOV-EUR` | euro govt bond **TR** (monthly) | OECD euro-area long-term yield `EA20.M.IRLT` → `TreasuryTR` (10y) | ~1970 |
| `EUROGOV-DAILY` | euro govt bond TR (daily shape) | ECB daily 10y yield curve `B.U2.EUR.4F.G_N_A.SV_C_YM.SR_10Y` → `TreasuryTR` | ~2004 |
| `EUROGOV-LONG-EUR` | **long** euro govt bond TR (25+, monthly) | month-ends of the real ECB 25y curve from 2004-09; before it, the OECD long-term yield mapped to a 25y yield (`0.571+0.962×10y`, calibrated on that same curve) → `TreasuryTR` (24y par, dur ~17), rebased at the junction | ~1970 |
| `EUROGOV-LONG-DAILY` | long euro govt bond TR (daily shape) | ECB daily 25y yield curve `B.U2.EUR.4F.G_N_A.SV_C_YM.SR_25Y` → `TreasuryTR` (24y par) | ~2004 |
| `DECASH-EUR` | German 3M money-market accrual | OECD German 3M interbank `DEU.M.IR3TIB` (the Bundesbank's Frankfurt three-month money, FIBOR from 1991), rolled monthly at the convention it is quoted in: simple, German 360/360 to 1990-06 then act/360 | 1960-1994 |
| `EURCASH-EUR` | euro-area 3M cash TR index | ECB monthly EURIBOR 3M `FM.M.U2.EUR.RT.MM.EURIBOR3MD_.HSTA`, rolled monthly at the convention it is quoted in: simple, act/360 | 1994-> |

## Source dataflow: the 2026-08 migration off OECD MEI

The three OECD inputs used to be read from the legacy `OECD/MEI` dataflow. That
dataflow **stopped being updated in 2024-01** and went on answering HTTP 200 for
two and a half years, so the bundled euro references silently carried a
two-year-old tail. On **2026-08-19** the generator was moved to the current
short-term-statistics dataflow `OECD/DSD_STES@DF_FINMARK`, the one
`cmd/gen-gbond-refdata` already read:

| was (`OECD/MEI`, ends 2024-01) | is (`OECD/DSD_STES@DF_FINMARK`) |
|---|---|
| `EA19.IRLTLT01.ST.M` | `EA20.M.IRLT.PA._Z._Z._Z._Z.N` |
| `EA19.SPASTT01.IXOB.M` | `EA20.M.SHARE.IX._Z._Z._Z._Z.N` |
| `DEU.IR3TIB01.ST.M` | `DEU.M.IR3TIB.PA._Z._Z._Z._Z.N` |

**EA20 rather than EA19**, because the euro area has had 20 members since
Croatia joined in 2023 and only the EA20 aggregate is still being extended: the
new dataflow publishes both, EA20 runs four months further (2026-05 against
2026-01), and over their entire common history (673 monthly yields from 1970-01,
470 share-price points from 1986-12) the two agree to within **1.5e-4 percentage
point** on the yield and to the last published digit on the equity index. The
choice therefore costs nothing in comparability and buys the live tail.

What the migration changed, measured old file against new file on the shared
window (levels rebased at the first common date):

- `EMU-EUR`, `DECASH-EUR`: identical to **1e-6** over 446 and 420 months (each
  measured against itself under the convention it was then built at; what
  re-quoting `DECASH-EUR` moved in 2026-08 is the cash-conventions section
  below). The dataflow move alone revised nothing.
- `EUROGOV-EUR` / `EUROGOV-LONG-EUR`: cumulative deviation under **0.92 %** and
  **0.23 %** over 54 years, from OECD revisions of a few basis points on the
  monthly yield (the largest single month is 2010-12, 0.11 pt of return).
- the **daily** pair was not merely extended but **repaired**. The files shipped
  before this pass carried flat runs from a degraded fetch: `EUROGOV-DAILY` held
  one constant level for **698 trading days** (2019-04 to 2022-01, the whole
  covid round trip) and `EUROGOV-LONG-DAILY` for 114. The ECB series they come
  from is clean today and the rebuilt files have **no repeated level at all**.
  Nothing reported this at the time, which is why the generator now checks it.
- the new OECD tail, 2024-01 to 2026-05, is sane against outside knowledge: the
  euro-area 10-year lands at **3.49 %** (the ECB AAA curve is at 3.21 %, a
  28 bp aggregate-over-AAA spread), and `EMU-EUR` returns **+16.3 %/yr** with
  10.7 %/yr monthly volatility against **+16.9 %/yr** and 10.7 % for the real
  EZU in EUR over the very same window: 0.6 pt/yr apart on the level and within
  0.1 pt on the risk.

The generator now validates every series before writing it (`-check`, on by
default), the way `cmd/gen-gbond-refdata` does: freshness (an OECD series more
than a year stale, or an ECB one more than a quarter, fails: this exact failure
went unnoticed for two years), flat runs, `EUROGOV-EUR` against the bundled
`BUND-EUR` over 1999-2010 (corr 0.955, CAGR gap -0.26 pt), the daily curve
against the monthly yield (vol ratio 1.09), the long reconstruction against
DBXG's realized volatility (13.99 %/yr), the synthesized long yield against the
ECB curve it was fitted on (it grades the deep tail, which is all it feeds), the
long splice against the curve it samples (264 month-ends, bit-exact), and the equity gross-up against the EZU overlap
`netDivYield` was calibrated on (3.05 %/yr, the calibration target to two
decimals).

Both calibrations were re-derived from the live data and **neither moved**: the
25-year-on-10-year regression over the full 2004-2026 ECB curve still fits
`0.5711 + 0.9615x` to four digits, and `EMU-EUR`'s 2001-2023 CAGR is still
3.05 %/yr, exactly the EZU net-TR figure `netDivYield = 2.2 %` was set against.

The `EUROGOV-LONG-*` pair proxies the **25+** segment of the euro sovereign
curve (`dbxgRecipe`, backing DBXG). It is a genuine long-bond reconstruction:
the yield path is run through `TreasuryTR` at a long maturity, so the level and
convexity come from real bond pricing, not a levered shorter bond. The maturity
(24y par ≈ modified duration ~17) is trimmed below the 25y curve point so the
daily reconstruction matches DBXG's ~14.4%/yr realized volatility over
2007-2026, because the fitted long curve is slightly more volatile than the
traded fund per year of maturity. DBXG's own quotes cover 2007-> (grafted);
before that the ECB 25y curve gives both the level and the daily shape back to
2004-09, and the OECD-10y-derived monthly tail carries it to ~1970. Reconstructing the 25+ from a levered 10y was
rejected: leveraging the excess return compounds a sustained bond bull far too
richly (~20%/yr over the 1981-2004 disinflation vs ~13%/yr for a real long bond).

## The cash legs: what each rate is quoted in

A short rate is a number plus a convention, and the convention is a property of
the publication, not a modelling choice. Reading a simple money-market quote as
if it were already compounded to the year costs about `r^2/2` a year (0.17 %/yr
at 6 %); reading a 360-day year as a calendar one costs another 1.46 % of the
rate. Both legs have now been re-quoted at what their publishers state, and both
had it wrong before: `DECASH-EUR` was compounded as an **effective annual yield**
until **2026-08-20**, and `EURCASH-EUR` was rolled simple but on a **365.25-day
year** until **2026-08-22**, where EMMI publishes Euribor act/360.

### The pre-euro leg (`DECASH-EUR`, fixed 2026-08-20)

The evidence, gathered before the change and all of it primary:

- **What the series is.** `DEU.M.IR3TIB` is the Bundesbank's own money-market
  quotation for three-month funds at the Frankfurt banking centre, FIBOR from
  1991-01. Ten annual averages read off the Bundesbank's long time series
  (*III. Geld- und Kapitalmarkt*, table "Geldmarktzinsen bis 1998") match the
  means of the OECD monthly series to **under 0.005 pt** (1965-1974), which is
  the rounding of the monthly figures being averaged. The OECD series is a
  republication, so its convention is the Bundesbank's.
- **What that convention is.** The Bundesbank's own footnote on the series:
  "Bis Juni 1990 berechnet nach der deutschen Zinsmethode 360/360 Tage, ab Juli
  1990 nach der Zinsmethode act/360." So the rate is **simple for the term of
  the deposit** throughout, on a 360-day year: the German 360/360 method (every
  month exactly one twelfth of the quoted rate) to 1990-06, act/360 after. FIBOR,
  which takes the series over in 1991-01, was quoted act/360 too. **No sub-era
  is quoted as an effective annual yield**, so there was nothing to preserve.
- **What it measures.** Under the shipped convention the index grew at
  6.31 %/yr over 1960-1994 against a mean quoted rate of 6.34 %, i.e. it paid
  slightly LESS than the rate it was built from, which no rolled deposit does.
  It now grows at **6.54 %/yr**: the rate, plus the roll compounding
  (+0.19 pt in the 1970s, +0.20 pt in the 1980s) and the act/360 uplift after
  1990-07 (+0.47 pt over 1990-1994).

Measured, new file against old, on the 420 months they both cover (dates
identical): **+0.230 pt/yr**, **+7.84 %** cumulative over 35 years, worst single
month +0.109 pt (1981-03). The generator now gates the convention itself rather
than the level: every full calendar year of the index is compared with what the
documented convention pays on that year's mean quoted rate, restated
independently of the builder, and the two agree to **0.008 pt** at worst over 34
years (gate 0.10 pt). Reading the same quote as an effective annual yield lands
0.2 to 0.7 pt below in the high-rate years, so the check names the mistake that
was there.

`DECASH-EUR` is rebased onto `EURCASH-EUR` where the two meet, so the +7.84 %
level change reaches **nothing**: only pre-1994 RETURNS travel downstream, and
the young end of every reconstruction is untouched (end-level ratio 1.000000 for
both euro-native lines). The four reconstructions that reach behind 1994 and
finance or earn euro cash move by, over their whole history: XEON
**+0.115 pt/yr** (it IS euro cash), the EUR-hedged BTOP50 index **+0.058 pt/yr**
(hedging earns euro cash), NTSZ **-0.029 pt/yr** and NTSG **-0.009 pt/yr** (both
finance a bond overlay at it, 0.60 and 0.066 of notional against a 0.10 cash
sleeve, so a dearer cash rate is a net cost). Every audit verdict is unchanged:
the graded overlaps all start in 2007 or later, decades after the last month
`DECASH-EUR` touches.

One thing was measured then and deliberately deferred: `EURCASH-EUR` accrued
EURIBOR over a 365.25-day year while EURIBOR is itself quoted act/360, worth
1.46 % of the rate (0.03 %/yr at its 2.3 %/yr average). It was left alone
because it is a change to the LIVE cash leg every hedged recipe reads, not to a
pre-euro tail, and it belonged with its own regeneration of those lines. That is
the section below. Note it was never the deep tail's problem: the German 360/360
era is 366 of `DECASH-EUR`'s 420 months, and on a monthly grid that convention
IS one twelfth of the rate a month, which the generator computes exactly.

### The live leg (`EURCASH-EUR`, fixed 2026-08-22)

- **What the series is.** `FM.M.U2.EUR.RT.MM.EURIBOR3MD_.HSTA` is the ECB's
  monthly history of the 3-month Euribor fixing, the same rate the repository
  already reads live as `^EURIBOR3M`. Euribor is administered by EMMI.
- **What that convention is.** EMMI's own *Benchmark Determination Methodology
  for Euribor* (D0016F-2019, in force 2024-12-05), paragraph 43: "The published
  Euribor rates follow euro money market conventions, that is, spot settlement
  (T+2), the TARGET2 calendar, an **Actual/360 day count convention**, and
  modified following business day with month-end adjustment convention." So the
  rate is simple for the term of the deposit, on a 360-day year, and reading it
  on a calendar year is short by 1.46 % of the rate. The pre-1999 months of the
  series are the euro's national-money-market ancestors, quoted act/360 as well
  (the Frankfurt three-month rate had moved to act/360 in 1990-07, per the
  section above), so the whole span takes one convention.
- **What it measures.** Over 1994-01 to 2026-07 the index grew at **2.2705 %/yr**
  on the calendar year; it now grows at **2.3040 %/yr**, i.e. **+0.0334 pt/yr**,
  **+1.068 %** cumulative over 32.5 years, worst single month +0.0094 pt
  (1995-04). That is the deferred estimate confirmed, not a surprise.

The generator gates it the way it gates the German leg: every full calendar year
of the index against what act/360 pays on that year's mean quoted rate, restated
independently of the builder, agreeing to **0.005 pt** at worst over 32 years
(gate 0.10). The check against the previously shipped file was extended rather
than dropped: a day-count re-quote multiplies every step's interest by a
constant and changes nothing else, so the check also compares the rebuilt index
against the shipped file **restated** that way, and it agreed to **4.4e-9**,
the rounding of the six decimals the file carries. That is the proof that the
day count was all that moved, and a refresh that drifts for any other reason
still fails.

Downstream, the twelve reconstructions that read the leg move over their whole
history by, from the re-quote alone: `DTLETR` **+0.034 %/yr** (EUR-hedged on the
full notional from 1994, the largest exposure to the leg), `MFEH` **+0.030**,
`ERNX` **+0.028** (it IS euro cash), the three EUR-hedged AQR classes
**+0.026**, `42C0` **+0.019**, the EUR-hedged BTOP50 index **+0.010**, `XEON`
**+0.006**, `CHSN` **+0.005**, `NTSG` **0.000**, and `NTSZ` **-0.014** (it
finances a 0.60 bond overlay at cash against a 0.10 cash sleeve, so a dearer
cash rate is a net cost). Every `-verify-simdata` level and path verdict is
unchanged, the graded windows moving by 0.02 to 0.04 pt of CAGR at most.

## The recipe (`ntszRecipe` / `ntszBuild`)

```
0.90 × equity  +  0.60 × (bond − EUR cash, futures overlay)  +  0.10 × EUR cash
fee 0.20%/yr, charged only where the donors' own charges do not already cover
it (EZU 0.50% and EUNH 0.07% come to 0.492% blended from 2000, so the whole
0.20% falls on the pre-2000 reference era alone) ; real NTSZ grafted from
inception (2025-10)
```

- **Equity leg** (`ntszEquityEUR`): the real MSCI Eurozone ETF (`EZU`, US-listed,
  USD, 2000->) re-expressed in EUR at the EURUSD spot (the same USD->EUR identity
  as the unhedged DBMFE leg), then extended before EZU with `EMU-EUR`. EZU is the
  deepest *real* eurozone equity series available; the euro-native `EMU-EUR`
  supplies the pre-2000 tail. This leg sets the composite floor at **~1986**.
- **Bond leg**: the real iShares Core Euro Govt Bond ETF (`EUNH.DE`, EUR, 2009->)
  extended by `EUROGOV-EUR` (with the ECB `EUROGOV-DAILY` shape from 2004),
  financed at EUR cash (`Excess` leg).
- **Cash leg** (`eurCashDeep`): the euro money-market index `EURCASH-EUR` carried
  to daily granularity (`eurCashDaily`, 1994->) and extended before the euro by
  `DECASH-EUR` (Germany was the anchor economy, the DM the reference currency).
  `EURCASH-EUR` was a hand-built file with no generator until **2026-08-20**,
  which is why it froze at 2026-01 when its FRED source became unreachable (FRED
  times out from the generation environment and DBnomics has dropped the FRED
  provider). It is now produced here, from the ECB's own EURIBOR history: the
  rebuilt index reproduces the shipped one to **4e-4** over the 385 months they
  share, and the generator refuses to write it if that ever exceeds 2e-3. What
  each leg's rate is quoted in, and what re-quoting `DECASH-EUR` moved, is the
  cash-conventions section above.

`ntszBuild` pre-builds the equity and deep-cash legs and serves them to the
standard frame/`Composite` machinery under synthetic ids (`injected` fetcher);
the bond leg reaches back through the ordinary `extend()`/`longBack` splice.

## Depth ceiling: why ~1986 and not the 1970s

The bond and cash legs reach 1970 and 1960 cheaply. The **equity leg is the
binding constraint**:

- A credible eurozone equity **total return** only goes back to the OECD euro-area
  share-price index (`EA20.M.SHARE`, ~1986). MSCI EMU net TR via Curvo starts
  even later (~1998) and is a manual export, not fetchable.
- Reaching the 1970s would require fabricating a synthetic pre-euro "eurozone
  equity in EUR" (aggregating pre-euro national markets in a currency that did
  not exist), or using Germany alone as the proxy. Both are epistemically weak
  and would overstate the backcast's authority, so they are **deliberately not
  done**. The composite starts where a real, broad euro-area equity TR does.

## Calibration and known limitations (the ledger)

- **Equity net-dividend constant (`netDivYield = 2.2%/yr`).** `EA20.M.SHARE` is
  a price index; the gross-up to net TR uses a constant calibrated on the EZU
  (net TR, in EUR) overlap: EZU's EUR CAGR over 2001-2023 is ~3.05%/yr vs
  ~0.84%/yr for the price index, a ~2.2%/yr gap (dividends + universe drift). A
  constant modestly understates the richer pre-2000 dividend yield, so the deep
  equity tail is, if anything, **conservative**. Only the pre-2000 return drift
  depends on it (the level is rescaled where EZU takes over).
- **Daily statistics are inflated on the equity leg pre-2025.** EZU is US-close
  (16:00 ET); dividing by the async EURUSD close does not perfectly cancel
  intraday, so the reconstructed EUR equity carries extra day-to-day noise. It
  shows up as a wider daily-vs-monthly volatility gap than NTSX/NTSG (NTSZ daily
  vol ~21.5% vs monthly ~15.7%, against ~19%/15% for NTSX). **Use monthly/weekly
  statistics for NTSZ**; the level path, CAGR and drawdowns are unaffected. This
  is the same async-pricing caveat documented for the TIPS-hedged and unhedged
  DBMFE recipes.
- **Validation is thin by construction.** The real fund has ~months of history,
  so the overlap barely clears the 60-point floor: weekly corr ~0.93, daily ~0.63,
  and the annualized-CAGR comparison over 9 months is noise. The value here is
  the deep reconstruction, not a tight tracking claim; the real quotes are
  grafted from inception regardless.
- **The long series' level now follows the real ECB curve (fixed 2026-08-19).**
  `euroGovLongDaily` takes `EUROGOV-LONG-EUR` as anchors and
  `EUROGOV-LONG-DAILY` only as intra-month shape, so whatever the monthly file
  carries sets the level. It used to carry a 25-year yield *synthesized from the
  10-year* over 2004-2026 too, where the ECB publishes the real 25-year point.
  The affine map has a slope of 0.96, so it cannot reproduce a **steepening**:
  over 2024-01 to 2026-05 the synthesized long bond returned -0.3 %/yr where the
  real ECB 25-year curve returns -5.1 %/yr (the same window refits the
  regression at a slope of 1.43). This was invisible while the OECD tail stopped
  in 2024-01; extending it moved `MTH`'s level verdict to warn and widened
  `DBXG`'s gap from +0.01 to +0.65 %/yr.
  `cmd/gen-euro-refdata` now builds the monthly file the way a donor chain is
  built: **the real curve where it exists, the synthesis only in front of it**.
  From the ECB curve's first month (**2004-09-30**) every monthly anchor is that
  curve's own month-end level, so anchors and shape are the same series and the
  reconstruction reproduces the real long bond exactly; the synthesized tail
  (~1970 to 2004-08, its published levels unchanged to the last digit) has the
  real era rebased onto it by one constant factor, so the junction shows no
  level jump and the seam keeps the one synthesized month return (+1.14 %) it
  has no substitute for. Measured old file against new: the pre-junction 416
  months are **identical**, the real era returns **+3.07 %/yr against +3.31 %/yr**
  synthesized (widest rebased path deviation +31 % in 2012, -4 % at the end),
  and the 2024-2026 tail now returns **-4.96 %/yr**, the real curve's own
  number. `-verify-simdata` on the same windows: `DBXG` **+0.65 → -0.05 %/yr**
  (monthly corr 0.51 → 0.93, drift +12.9 % → -0.9 %), `MTH` **+1.89 →
  -0.54 %/yr** (corr 0.56 → 0.96, drift +15.4 % → -4.2 %), level verdicts ok on
  both. The affine map was **not** refitted: 2004-2026 is the only window where
  a real 25-year yield exists at all, and it now governs the deep tail alone.
- **Both cash legs are quoted right (fixed 2026-08-20 and 2026-08-22).**
  `DECASH-EUR` read its source rate as an effective annual yield, where the
  Bundesbank quotes it simple on a 360-day year; re-quoting it adds
  **+0.230 %/yr** to the pre-1994 cash path and moves the four reconstructions
  that reach behind 1994 by **+0.115 to -0.029 %/yr** over their whole history.
  `EURCASH-EUR` then read Euribor on a calendar year, where EMMI publishes it
  act/360; re-quoting it adds **+0.033 %/yr** to the euro-era cash path and
  moves the twelve reconstructions that read it by **+0.034 to -0.014 %/yr**.
  No audit verdict changed either time. Evidence and measurements are in the
  cash-conventions section above.
- **Bond duration.** `EUROGOV-*` reconstruct a 10y benchmark (matching the OECD/ECB
  yield tenor); the real `EUNH.DE` is a broad eurozone govt basket (duration
  ~7-8y). The small mismatch is rescaled/absorbed at the 2009 splice.

## Sanity check (1986-12 -> 2026, EUR)

| | NTSX | NTSG | NTSZ |
|---|---|---|---|
| CAGR | 11.1% | 9.2% | **8.4%** |
| Vol (monthly, ann.) | 15.3% | 14.2% | **15.7%** |
| Max drawdown | -50% | -49% | **-55%** |
| Worst rolling 5y | -6.3% | -5.9% | **-9.0%** |

The eurozone 90/60 trails the US and global versions and draws down harder,
exactly the "lost decade + leverage" story one expects, with a clean monthly
volatility close to its NTSX peer.

## Regeneration

```sh
make euro-refdata   # rebuild EMU-EUR / EUROGOV-EUR{,-DAILY} / EUROGOV-LONG-EUR{,-DAILY} / DECASH-EUR / EURCASH-EUR (network)
make simdata        # rebuild pkg/datasets/simdata/, including IE000OV4XWA3
```
