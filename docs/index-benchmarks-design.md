# Fee-free index benchmarks (`MSCIWORLD`, `SP500`)

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
