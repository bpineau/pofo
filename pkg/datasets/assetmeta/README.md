# Asset metadata: the bundled catalog

`assets.json` is the **single source of truth** for the assets pofo
bundles. It carries both how to fetch each asset (the resolution fields) and
what it is (the descriptive metadata). `pkg/marketdata` loads its catalog
from this file (embedded via `datasets.AssetMeta()`); `pkg/suggest` reads the
descriptive fields from the same file. To add or change a bundled asset, edit
this file only.

The descriptive data was collected from issuer factsheets/KIIDs, justETF,
Morningstar and index providers. Breakdowns are approximate (whole percents)
and dated; they describe the instrument, not a precise point-in-time holding.
Treat `confidence: medium|low` records with extra care and refresh as needed.

## Schema (`assets.json`: array of objects)

Resolution fields (consumed by `pkg/marketdata`):

| field | meaning |
|---|---|
| `id` | canonical identifier (ticker or ISIN); the key |
| `isin` | ISIN, or `null`/empty for indices/spot/futures |
| `aliases` | extra identifiers accepted in portfolio files (e.g. `GOLD`, `NTSX`) |
| `name` | display name |
| `ucits` | `true` for UCITS funds/ETFs (ETCs, US funds, indices are not) |
| `eu_retail` | `true` when an EU/French retail investor can actually buy it: every UCITS fund, plus EU-listed products with a PRIIPs KID (gold/commodity ETCs, listed closed-end funds like BH Macro). `false` for US-listed funds without a KID. Omitted for non-tradable series (`index` benchmarks, spot, futures). This is the buyability flag; `ucits` alone understates it (no gold product can be UCITS, yet ETCs are freely buyable) |
| `source` | quote provider: `yahoo`, `ft`, `morningstar`, `stooq`, or `index` (non-investable benchmark served from its embedded reconstruction, no live symbol) |
| `symbol` | provider symbol (Yahoo/Stooq ticker or Morningstar id); empty for FT and `index` |
| `xid` | FT internal id; empty otherwise |
| `fees` | pinned ongoing charge (TER), percent per year; `0` = unknown, or genuinely fee-free for an `index` benchmark |

Descriptive fields (consumed by `pkg/suggest`):

| field | meaning |
|---|---|
| `asset_class` | `equity`, `government-bond`, `corporate-bond`, `aggregate-bond`, `inflation-linked-bond`, `money-market`, `gold`, `broad-commodity`, `managed-futures`, `long-volatility`, `tail-risk`, `multi-asset`, `real-estate`, `other` |
| `underlying` | one-line plain description of what it holds |
| `benchmark_index` | the index it tracks, or `active (...)` / `null` |
| `strategy` | open vocabulary; common values: `physical-replication`, `synthetic-swap`, `active`, `futures-overlay`, `leveraged-2x`, `leveraged-3x`, `trend-following`, `long-volatility`, `multi-factor`, `systematic factor tilt`, `covered-call overlay`, `fundamentally-weighted`, `other` |
| `geography` | approximate region weights (percent), `{ "Global developed": 100 }`, or `null` when not meaningful (gold, broad managed futures, money market) |
| `sectors` | approximate equity sector weights (percent), or `null` for non-equity; for a stacked fund, describes the equity leg |
| `currency` | quote currency used for fetching and FX conversion (ISO code) |
| `currency_exposure` | optional look-through fiat exposure: currency (ISO code, plus `None` for real assets and `Dynamic` for futures books) → percent of capital; any shortfall below 100 counts as `None`. Set it only where the automatic derivation (`suggest.CurrencySplit`: hedging, asset class, geography, then quote currency) is wrong: funds denominated differently than their holdings' countries (corporate/aggregate bonds, EM hard-currency debt), mixed-region equity residuals worth resolving |
| `distribution` | `accumulating`, `distributing`, `n/a` |
| `leverage` | `1.0` normal; `2.0` for 2× daily; embedded notional for capital-efficient funds (e.g. `1.5` for a 90/60 structure) |
| `duration` | effective duration in years (fixed income); for a stacked fund, the duration of its bond exposure per unit of notional (e.g. `7.0` for a 90/60 fund's intermediate futures ladder) |
| `notes` | one line on the asset's portfolio role / the market regime it serves |
| `confidence` | `high`, `medium`, `low`: confidence in the breakdowns |
| `sources` | reference URLs |

## Controlled vocabularies

The composition consumer (`pkg/suggest/composition.go`) aggregates these fields
by canonical label, so `geography`, `sectors` and `currency_exposure` use fixed
vocabularies. `suggest.CanonRegion`, `suggest.CanonSector` and
`suggest.regionCurrency` must understand every label used here; extend those Go
maps only if you genuinely add a new country label, and keep them consistent
with this list.

### Geography

Countries use English standard names, with two abbreviations (`US`, `UK`):
`US`, `UK`, `Japan`, `Germany`, `France`, `Italy`, `Spain`, `Netherlands`,
`Ireland`, `Luxembourg`, `Greece`, `Austria`, `Belgium`, `Finland`, `Portugal`,
`Switzerland`, `Sweden`, `Denmark`, `Norway`, `Canada`, `Australia`,
`New Zealand`, `China`, `Hong Kong`, `Taiwan`, `South Korea`, `India`,
`Singapore`, `Indonesia`, `Thailand`, `Brazil`, `Mexico`, `South Africa`,
`Saudi Arabia`, `Poland` (other real countries allowed, English names).

Region buckets, only for residuals a fund does not break down finer:
`Other eurozone` (euro members not listed), `Other Europe` (European mix
spanning EUR and non-EUR: CHF, SEK, DKK, NOK, GBP), `Other developed`,
`Other emerging`, `Other`. `regionCurrency` maps eurozone labels and
`Other eurozone` to EUR; `Other Europe`, `Other developed`, `Other emerging`
and `Other` map to no single currency (the `Other` currency bucket), so an
equity fund carrying a large such residual may deserve a `currency_exposure`
override. Do not reintroduce eliminated spellings (`United States`,
`United Kingdom`, `Europe`, `Europe ex-UK`, `North America`, `Global`, `Asia`,
`Emerging Markets`, `Latin America`, `Middle East & Africa`, ...); resolve them
to the list above, splitting `North America` into `US` and `Canada`.

### Equity sectors

The 11 GICS sectors exactly, plus `Other` for residuals: `Information
Technology`, `Financials`, `Health Care`, `Consumer Discretionary`,
`Consumer Staples`, `Industrials`, `Energy`, `Materials`, `Utilities`,
`Real Estate`, `Communication Services`. For a stacked fund, `sectors`
describes the equity leg (the equity index it stacks). Bond funds keep their
own bond-sleeve labels (`US Government`, `Investment-Grade Corporate`, ...);
those are ignored by the equity pie and must not be GICS-normalized.

### Currency exposure

Keys are ISO 4217 codes, plus `None` (no fiat: real assets) and `Dynamic`
(futures book). No `Other`: resolve every residual to real currencies, or let
the below-100 shortfall count as `None`. Sums above 100 are intended for
levered funds (a 2x USD fund is `{"USD": 200}`); a stacked fund with a
managed-futures leg carries `Dynamic` for that leg's notional.

## Adding an asset

This file is curated. Add a record with the resolution fields (a working
`source`+`symbol`/`xid` and `currency`) and the descriptive fields, then
verify it fetches cleanly with `pofo -verify-data -assets <id>`. The
resolution fields make it part of the bundle (one `-warmup` away); the
descriptive fields feed `-coverage` and `-suggest`.

## Provenance and refresh recipes

Field guide for refreshing the descriptive data: per family, where the numbers
live and the dead ends. Each asset's `sources` array holds the deepest stable
URL actually used, and its `notes` line the interpretation hint.

### Geography splits
- Broad index funds (MSCI World / EAFE / EM, S&P 500, MSCI EMU): the MSCI index
  factsheet PDF (`msci.com/documents/10199/255599/<index>.pdf`, "Country
  Weights" panel) or the issuer factsheet. MSCI World country weights are also
  the currency basis for the Efficient Core and Winton equity legs.
- Active small/mid funds (Independance AM): the monthly reporting PDF carries a
  "Repartition geographique" pie and a sector table on page 2 (e.g. LU1832174962:
  `independance-am.com/wp-content/uploads/YYYY/MM/...-reporting-europe-small-...pdf`).
  That PDF is image-based: WebFetch returns binary, so open it as a PDF (the
  Read tool renders the pie and table).
- Keep region residuals as `Other eurozone/Europe/developed/emerging`; do not
  invent country detail an issuer does not publish.

### Equity-leg sectors of stacked funds
- Use the tracked index's sector weights, not a blended fund sheet. NTSG / RSSB /
  Winton = MSCI World; NTSX / RSST = S&P 500; NTSZ = MSCI EMU
  (`msci.com/resources/factsheets/index_fact_sheet/msci-emu-index.pdf`, "Sector
  Weights"). A LifeStrategy-type multi-asset sheet lumps the bond sleeve into
  "Other"; once `exposures` is set, replace `sectors` with the equity index's
  breakdown or the equity pie mis-reads it.

### Bond-leg durations
- Efficient Core family (NTSX/NTSI/NTSE/NTSG/NTSZ): a 2/5/10/30y government
  futures ladder, effective ~7y. WisdomTree publishes no single number; the
  ladder and the ~7-7.5y figure come from third-party reviews
  (`optimizedportfolio.com/ntsx`) and match aggregate-bond duration by design.
- Return Stacked (RSBT/RSSB): the bond stack targets the Bloomberg US Aggregate,
  ~6y (returnstackedetfs.com product pages / ReSolve commentary).
- LifeStrategy: global aggregate, EUR-hedged, ~6.5y.

### currency_exposure overrides
- Denomination beats geography for bonds: a EUR- or USD-denominated corporate or
  aggregate fund is 100% its denomination currency regardless of issuer country
  (the name/KID states it, even when the share class quotes in a third currency,
  e.g. IE00B3DKXQ41 quotes GBP but is 100% EUR). Hedged share classes need no
  override (rule 3 handles them). USD hard-currency EM sovereign = 100% USD; EM
  local debt = the index's local-currency basket (JPM GBI-EM Global Diversified
  weights, medium confidence, drifts with the index).
- Stacked funds: equity-leg currencies times leg notional, plus the fund's
  collateral currency for the margin pocket; bond FUTURES legs add no FX; a gold
  leg is `None` (shortfall); a managed-futures leg is `Dynamic` for its notional.

### Time sinks and dead ends
- Independance factsheets are image-only PDFs: read as PDF, not HTML.
- justETF / Boursorama / Morningstar fund pages and the spglobal sector
  dashboards are JavaScript-rendered or return 403 to automated fetch; use the
  index-provider PDF or a cached web-search result for the numbers.
- Ossiam CAPE US Sector Value (LU1079841513): the four held sectors rotate
  monthly and are not published as a stable table; any snapshot is low
  confidence (four cheapest-CAPE S&P 500 sectors at 25% each).
