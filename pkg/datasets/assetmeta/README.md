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

## Adding an asset

This file is curated. Add a record with the resolution fields (a working
`source`+`symbol`/`xid` and `currency`) and the descriptive fields, then
verify it fetches cleanly with `pofo -verify-data -assets <id>`. The
resolution fields make it part of the bundle (one `-warmup` away); the
descriptive fields feed `-coverage` and `-suggest`.
