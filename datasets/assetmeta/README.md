# Asset metadata

Factual, human-legible metadata for every asset in the bundled catalog
(`pkg/marketdata/catalog.go`), keyed by catalog `id`. These tags describe
*what each instrument is* — asset class, underlying, geography, sector,
strategy — so the optimizer's `-suggest` mode (see `pkg/optimize` and the
roadmap) can reason about diversification and regime/factor coverage rather
than chase the best-fitting weights on past returns alone.

The data was collected from issuer factsheets/KIIDs, justETF, Morningstar
and index providers. Breakdowns are approximate (whole percents) and dated;
they describe the instrument, not a precise point-in-time holding. Treat
`confidence: medium|low` records with extra care and refresh as needed.

## Schema (`assets.json` — array of objects)

| field | meaning |
|---|---|
| `id` | catalog identifier (matches `pkg/marketdata/catalog.go`) |
| `isin` | ISIN, or `null` for indices/spot/futures |
| `asset_class` | `equity`, `government-bond`, `corporate-bond`, `aggregate-bond`, `inflation-linked-bond`, `money-market`, `gold`, `broad-commodity`, `managed-futures`, `long-volatility`, `tail-risk`, `multi-asset`, `real-estate`, `other` |
| `underlying` | one-line plain description of what it holds |
| `benchmark_index` | the index it tracks, or `active (...)` / `null` |
| `strategy` | `physical-replication`, `synthetic-swap`, `active`, `futures-overlay`, `leveraged-2x`, `trend-following`, `long-volatility`, `other` |
| `geography` | approximate region weights (percent), `{ "Global developed": 100 }`, or `null` when not meaningful (gold, broad managed futures, money market) |
| `sectors` | approximate equity sector weights (percent), or `null` for non-equity |
| `currency` | fund base/quote currency (ISO code) |
| `distribution` | `accumulating`, `distributing`, `n/a` |
| `leverage` | `1.0` normal; `2.0` for 2× daily; embedded notional for capital-efficient funds (e.g. `1.5` for a 90/60 structure) |
| `notes` | one line on the asset's portfolio role / the market regime it serves |
| `confidence` | `high`, `medium`, `low` — confidence in the breakdowns |
| `sources` | reference URLs |

## Regenerating

This file is curated. It was bootstrapped by research sub-agents (one per
batch of catalog assets) and merged with a coverage check against the
catalog IDs (106/106, no gaps or duplicates). To extend it when the catalog
grows, research the new `id`s into the same schema and re-run the coverage
check.
