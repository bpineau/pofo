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
| `source` | quote provider: `yahoo`, `ft`, `morningstar` or `stooq` |
| `symbol` | provider symbol (Yahoo/Stooq ticker or Morningstar id); empty for FT |
| `xid` | FT internal id; empty otherwise |
| `fees` | pinned ongoing charge (TER), percent per year; `0` = unknown |

Descriptive fields (consumed by `pkg/suggest`):

| field | meaning |
|---|---|
| `asset_class` | `equity`, `government-bond`, `corporate-bond`, `aggregate-bond`, `inflation-linked-bond`, `money-market`, `gold`, `broad-commodity`, `managed-futures`, `long-volatility`, `tail-risk`, `multi-asset`, `real-estate`, `other` |
| `underlying` | one-line plain description of what it holds |
| `benchmark_index` | the index it tracks, or `active (...)` / `null` |
| `strategy` | `physical-replication`, `synthetic-swap`, `active`, `futures-overlay`, `leveraged-2x`, `trend-following`, `long-volatility`, `other` |
| `geography` | approximate region weights (percent), `{ "Global developed": 100 }`, or `null` when not meaningful (gold, broad managed futures, money market) |
| `sectors` | approximate equity sector weights (percent), or `null` for non-equity |
| `currency` | quote currency used for fetching and FX conversion (ISO code) |
| `distribution` | `accumulating`, `distributing`, `n/a` |
| `leverage` | `1.0` normal; `2.0` for 2× daily; embedded notional for capital-efficient funds (e.g. `1.5` for a 90/60 structure) |
| `notes` | one line on the asset's portfolio role / the market regime it serves |
| `confidence` | `high`, `medium`, `low`: confidence in the breakdowns |
| `sources` | reference URLs |

## Adding an asset

This file is curated. Add a record with the resolution fields (a working
`source`+`symbol`/`xid` and `currency`) and the descriptive fields, then
verify it fetches cleanly with `pofo -verify-data -assets <id>`. The
resolution fields make it part of the bundle (one `-warmup` away); the
descriptive fields feed `-coverage` and `-suggest`.
