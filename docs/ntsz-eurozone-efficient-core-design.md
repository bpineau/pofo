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

NTSX and NTSG reach back to 1953/1969 by leaning on long-running **US** index
funds (Vanguard) and USD refdata (S&P 500 to 1871, CMT Treasuries to 1953). NTSZ
is **euro-native end to end**, and no comparable deep euro building blocks
existed in the repo. So the deep tail is assembled from four new bundled
reference series, all sourced from **DBnomics** (free, key-less; the same mirror
the macro panel uses) by `cmd/gen-euro-refdata`. The pofo binary never fetches
OECD or the ECB; it embeds the CSVs.

| Refdata (new) | Content | Source | Span |
|---|---|---|---|
| `EMU-EUR` | eurozone equity **net TR** | OECD euro-area share-price index `EA19.SPASTT01` (price) grossed by a constant net dividend yield | ~1986 |
| `EUROGOV-EUR` | euro govt bond **TR** (monthly) | OECD euro-area 10y yield `EA19.IRLTLT01` → `TreasuryTR` (10y) | ~1970 |
| `EUROGOV-DAILY` | euro govt bond TR (daily shape) | ECB daily 10y yield curve `B.U2.EUR.4F.G_N_A.SV_C_YM.SR_10Y` → `TreasuryTR` | ~2004 |
| `DECASH-EUR` | German 3M money-market accrual | OECD German 3M interbank `DEU.IR3TIB01`, compounded | 1960-1994 |

## The recipe (`ntszRecipe` / `ntszBuild`)

```
0.90 × equity  +  0.60 × (bond − EUR cash, futures overlay)  +  0.10 × EUR cash
fee 0.20%/yr ; real NTSZ grafted from inception (2025-10)
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
  share-price index (`EA19.SPASTT01`, ~1986). MSCI EMU net TR via Curvo starts
  even later (~1998) and is a manual export, not fetchable.
- Reaching the 1970s would require fabricating a synthetic pre-euro "eurozone
  equity in EUR" (aggregating pre-euro national markets in a currency that did
  not exist), or using Germany alone as the proxy. Both are epistemically weak
  and would overstate the backcast's authority, so they are **deliberately not
  done**. The composite starts where a real, broad euro-area equity TR does.

## Calibration and known limitations (the ledger)

- **Equity net-dividend constant (`netDivYield = 2.2%/yr`).** `EA19.SPASTT01` is
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
- **Bond duration.** `EUROGOV-*` reconstruct a 10y benchmark (matching the OECD/ECB
  yield tenor); the real `EUNH.DE` is a broad eurozone govt basket (duration
  ~7-8y). The small mismatch is rescaled/absorbed at the 2009 splice.

## Sanity check (1986-12 -> 2026, EUR)

| | NTSX | NTSG | NTSZ |
|---|---|---|---|
| CAGR | 11.0% | 9.8% | **8.2%** |
| Vol (monthly, ann.) | 15.3% | 13.5% | **15.7%** |
| Max drawdown | -50% | -47% | **-55%** |
| Worst rolling 5y | -6.3% | -5.2% | **-9.2%** |

The eurozone 90/60 trails the US and global versions and draws down harder,
exactly the "lost decade + leverage" story one expects, with a clean monthly
volatility close to its NTSX peer.

## Regeneration

```sh
make euro-refdata   # rebuild EMU-EUR / EUROGOV-EUR / EUROGOV-DAILY / DECASH-EUR (network)
make simdata        # rebuild pkg/datasets/simdata/, including IE000OV4XWA3
```
