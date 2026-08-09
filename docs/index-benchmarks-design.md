# Fee-free index benchmarks (`MSCIWORLD`, `SP500`, `BTOP50`, `BTOP50E`)

## Problem

Users type `MSCIWORLD` / `SP500` expecting "the index" as a benchmark. Before
this, `MSCIWORLD` had no mapping and fuzzy-resolved to unrelated funds (fixed
by the resolution relevance gate); aliasing it to a real ETF (IWDA, VUAA) is
better but wrong in spirit: an ETF bleeds its TER every year, so it is not the
index. A benchmark should be the pure index total return, gross of fund fees,
with the full long history, by default (no `SIM` suffix).

## Design

### A new asset kind: `source: "index"`

A catalog entry may declare `"source": "index"`. Such an entry has no tradable
listing: `isin` empty, `symbol` empty, `fees` 0. Its price series *is* the
embedded daily total-return reconstruction `pkg/datasets/simdata/<id>.csv`.

`Client.fetch` gains one branch before the ISIN/ticker dispatch: when the
canonical id resolves to an `index` entry, it serves the simdata series
directly (`ReadSimdataFS(datasets.Simdata(), id)`), in the series' native
currency (USD), skipping all network resolution. Everything downstream is
unchanged: `FetchExtended` converts the currency and trims as for any USD
asset, so `-a MSCIWORLD` (EUR) works. The result:

- **bare id = long** (no `SIM` needed); the `SIM` suffix is a harmless no-op,
- **fee-free**: the reconstruction applies a 0 TER,
- **non-investable**: `fees: 0`, no ISIN. The investable UCITS ETFs stay under
  their own ids (`IWDA`/`IE00B4L5Y983`, `VUAA`/`IE00BFMXXD54`).

### The two entries

| id | index | reconstruction | history |
|---|---|---|---|
| `MSCIWORLD` | MSCI World Net TR (USD) | `MSCIWORLD-USD` refdata levels, MSCI World price-index daily shape (`^990100`), 0 TER | ~1969 |
| `SP500` | S&P 500 Total Return (USD) | `SP500-USD` refdata levels, `^GSPC` daily shape, 0 TER | ~1871/1962 |

Aliases: `MSCI-WORLD` -> `MSCIWORLD`, `SP-500` -> `SP500` (case folded by
`CanonicalID`). The temporary ETF aliases added earlier (`MSCIWORLD`,
`MSCI-WORLD` on IE00B4L5Y983; `SP-500` on IE00BFMXXD54, and the pre-existing
`SP500` there) are removed so the new entry ids win with no collision.

### Recipes + simdata

`msciworldIndexRecipe` reuses `msciWorld(0.0, fallback)`; `sp500IndexRecipe`
uses a parallel `sp500Index()` builder (SP500-USD anchors + `^GSPC` daily
shape, no fee). `make simdata` / `-gen-simdata MSCIWORLD SP500` writes the two
CSVs. Validation: correlation ~1.0 against the matching ETF, with an expected
CAGR gap of about the ETF's TER (that gap is the point), plus CAGR/vol sanity
against the reference index.

### The managed-futures pair (`BTOP50`, `BTOP50E`, added 2026-08)

Same kind, different job. Every managed-futures RECONSTRUCTION stops at
1996-03, the first NAV of the deepest real donor, so no book carrying a trend
sleeve could be tested through the 1987 crash, 1990 or the 1994 bond rout.
`BTOP50` serves the monthly Barclay BTOP50 net composite (`TREND-NET-USD`
refdata, 1986-12, already net of each constituent manager's fees) with the
daily texture of the net pure-trend composite (`TREND-PURE-NET-USD`, 2000-01),
exactly as `MSCIWORLD` is monthly before its daily shape donor opens in 1972.
`BTOP50E` is the same index hedged into EUR by the standard identity (local
total return less USD cash plus euro cash, the euro leg on the deep chain that
reaches the German money market), because every trend line a European
household can actually buy is a EUR class or EUR-hedged.

Two properties make this a benchmark rather than a reconstruction, and both
matter: nothing is rescaled to a fund's volatility target (the index is served
at its own ~9.3 % in USD, ~8 % hedged, against the ~15 % a UCITS trend fund
runs), and nothing is grafted. Rescaling an index to a fund's target is
precisely what discredited an earlier tail over this period; see "The tail that
was removed" in `docs/trend-reconstruction-design.md`. A sleeve held through
this line therefore carries roughly half the risk of the real one, which is
the price of the extra decade and the safe direction to err in. Measured over
1996-2026 on `examples/risk-budget-decumulation-longhist.txt`, substituting the
index for the two fund lines at equal weight costs 0.74 points of CAGR and
moves the drawdown by 0.06.

## Testing

- `CanonicalID` / resolution: every spelling (`MSCIWORLD`, `MSCI-WORLD`,
  lower-case, `SP500`, `SP-500`, and `...SIM`) maps to the index id.
- A network-free fetch test (fake simdata FS + `stubAllBases`) proves bare
  `MSCIWORLD` returns the long USD series with zero HTTP calls.
- Golden CAGR/vol sanity for both reconstructions against the reference index.

## Trade-offs

- `MSCIWORLD`/`SP500` are deliberately non-investable. Net TR sits marginally
  below a hypothetical zero-cost investor (dividend withholding tax) but is the
  standard published benchmark, and matches the refdata we already ship.
- One small fetch-path branch is the whole code cost; the SIM convention is
  untouched for every other asset.
