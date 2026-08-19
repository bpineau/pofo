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
| `EUROGOV-LONG-EUR` | **long** euro govt bond TR (25+, monthly) | OECD long-term yield mapped to a 25y yield (`0.571+0.962×10y`, calibrated on the 2004-2026 ECB curve) → `TreasuryTR` (24y par, dur ~17) | ~1970 |
| `EUROGOV-LONG-DAILY` | long euro govt bond TR (daily shape) | ECB daily 25y yield curve `B.U2.EUR.4F.G_N_A.SV_C_YM.SR_25Y` → `TreasuryTR` (24y par) | ~2004 |
| `DECASH-EUR` | German 3M money-market accrual | OECD German 3M interbank `DEU.M.IR3TIB`, compounded | 1960-1994 |

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

- `EMU-EUR`, `DECASH-EUR`: identical to **1e-6** over 446 and 420 months. The
  dataflow move alone revised nothing.
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
ECB curve it was fitted on, and the equity gross-up against the EZU overlap
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
before that the daily ECB 25y series shapes it to 2004, and the OECD-10y-derived
monthly tail carries it to ~1970. Reconstructing the 25+ from a levered 10y was
rejected: leveraging the excess return compounds a sustained bond bull far too
richly (~20%/yr over the 1981-2004 disinflation vs ~13%/yr for a real long bond).

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
- **The long monthly anchors govern the level even where a real long curve
  exists.** `euroGovLongDaily` takes `EUROGOV-LONG-EUR` as anchors and
  `EUROGOV-LONG-DAILY` only as intra-month shape, so the 25-year yield
  *synthesized from the 10-year* sets the level over 2004-2026 too, where the
  ECB publishes the real 25-year point. The affine map has a slope of 0.96, so
  it cannot reproduce a **steepening**: over 2024-01 to 2026-05 the synthesized
  long bond returns -0.3 %/yr where the real ECB 25-year curve returns
  -5.1 %/yr (the same window refits the regression at a slope of 1.43). This was
  invisible while the OECD tail stopped in 2024-01; extending it moved `MTH`'s
  level verdict from ok to warn (engine -2.30 %/yr vs real -4.19 %/yr over
  2018-2026) and widened `DBXG`'s gap from +0.01 to +0.65 %/yr, both still on
  windows the SIM consumer never sees (real quotes are grafted from 2018-09 and
  2007-08). Anchoring the post-2004 level on the real curve instead is the
  obvious fix and is **not** done here: it is a recipe change, not a data
  refresh.
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
make euro-refdata   # rebuild EMU-EUR / EUROGOV-EUR{,-DAILY} / EUROGOV-LONG-EUR{,-DAILY} / DECASH-EUR (network)
make simdata        # rebuild pkg/datasets/simdata/, including IE000OV4XWA3
```
