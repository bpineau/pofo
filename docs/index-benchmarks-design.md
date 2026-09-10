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

## Keeping the MSCI anchors current (`cmd/gen-msci-refdata`, 2026-09)

`MSCIWORLD-USD` is a manual month-end export from curvo.eu, and so are
`DEVEXUS-USD` (MSCI World ex USA) and `EM-USD` (MSCI Emerging Markets). Having
no generator, all three froze at the day of the last export: on 2026-09-10 they
still stopped at 2026-05, and with them the LEVEL of every reconstruction they
anchor (`MSCIWORLD`, `URTH`, `IWDA`, `WPEA`, the VT/VTI legs). `make
msci-refdata` closes that gap without touching a single exported point.

### The tail policy

The export always wins where it exists. The generator reads the current CSV,
keeps every point strictly before the month named by the `# tail-from:` header,
throws away whatever tail a previous run appended, and rebuilds it from the
proxy. So the Curvo boundary is minted once, on the first run, and frozen for
good; the tail is recomputed from the freshest quotes rather than accumulated;
and a re-run with no new month is a no-op on the data. Only COMPLETE months are
appended (the proxy's own last month is dropped whatever the day of the month),
which is why these references stop at the last finished month while the funds
anchored on them are current to yesterday through their real quotes.

### The proxy and its bands

The proxy is a catalogued ETF tracking the SAME index, quoted in the series'
currency, read through the normal client (adjusted closes are total returns).
Its ongoing charge is added back, because the reference is an index and is gross
of any fund fee. Every band below was measured before it was set, and the
measurement is printed on every run and written into the file's
`# tail-source:` header:

| band | value | why |
|---|---|---|
| currency | must match | a EUR line would splice an FX return into the index |
| overlap | >= 24 months | below that the tracking difference is noise |
| graded window | last 120 months of the overlap | these ETFs' launch years carry visibly bad vendor prints (a flat month against a -9.6 % index in 2010), and a proxy used at the front edge should be graded on the regime the tail is drawn from |
| tracking difference | \|TD\| <= 0.50 %/yr | a tracker of the right index still BEATS the net index gross of fees, by exactly what its domicile reclaims: +0.27 %/yr measured for Dublin-domiciled IWDA on MSCI World (treaty rate on the 70 % US sleeve), +0.41 for US-listed URTH (no US withholding at all, where the net index assumes the maximum). The band admits that; a neighbouring index misses by whole points |
| monthly correlation | >= 0.98 | a path sanity floor, not a second level gate: see the stub below |
| appended month | \|r\| <= 25 % | catches a scale break or a bad print in the tail itself |

### The month-boundary stub, and why selection is by rmse

An ETF's month-end price is struck at its own exchange's close, on its own
holiday calendar, while a global index is struck at each constituent market's
local close. A London line tracking MSCI World therefore carries a
month-boundary stub: in August 2026 the LSE was shut on the 31st and IWDA.L
printed +4.09 % for the month against +2.48 % for the MSCI World price index
struck at the index times, a 1.6-point error that only reverses the following
month. That stub caps a London line's monthly correlation near 0.985 and is
what 0.98 exists to accept.

Selection among the candidates that pass is therefore by smallest monthly
RMSE, not by smallest tracking difference: three months of URTH's +0.41 %/yr
level bias cost 0.10 %, where one misaligned month-end costs ten times that.
The RMSE weighs both, and it picks the proxy struck closest to the index's own
strike times. Measured on the graded window (2026-09-10):

| series | proxy | +TER | TD | corr | rmse | verdict |
|---|---|---|---|---|---|---|
| MSCIWORLD-USD | URTH (US-listed) | 0.24 | +0.410 %/yr | 0.9987 | 0.219 %/mo | picked |
| MSCIWORLD-USD | IWDA.L / IE00B4L5Y983 | 0.20 | +0.266 %/yr | 0.9868 | 0.705 %/mo | passes, worse path |
| DEVEXUS-USD | EXUS.L / IE0006WW1TQ4 | 0.15 | -0.019 %/yr | 0.9966 | 0.297 %/mo | picked, sole candidate |
| EM-USD | XMME.L / IE00BTJRMP35 | 0.18 | -0.217 %/yr | 0.9887 | 0.785 %/mo | picked, sole candidate |

The stub scales with the weight of the markets whose close the ETF's exchange
misses, which is why it is worst for a London World line (70 % US), mild for a
London ex-US line, and middling for a London EM line whose Asian constituents
closed hours before. No blend is used: blending URTH with IWDA.L would average
a faithful path with a stubby one and buy nothing.

Two candidates deliberately refused, for the record: ACWI ex USA trackers
(`ACWX`) for `DEVEXUS-USD`, since they hold emerging markets, and IMI trackers
(`EIMI`, `IEMG`) for `EM-USD`, since they add small caps. Both would pass the
correlation floor and fail the tracking difference, which is the division of
labour between the two gates. Yahoo does publish the MSCI net total-return
indices themselves (`^990100-USD-NETR` and siblings) but serves no history for
them, one quote only, so they cannot be the proxy.


## Testing

- `CanonicalID` / resolution: every spelling (`MSCIWORLD`, `MSCI-WORLD`,
  lower-case, `SP500`, `SP-500`, and `...SIM`) maps to the index id.
- A network-free fetch test (fake simdata FS + `stubAllBases`) proves bare
  `MSCIWORLD` returns the long USD series with zero HTTP calls.
- Golden CAGR/vol sanity for both reconstructions against the reference index.
- `cmd/gen-msci-refdata` grades and appends on synthetic fixtures: the append
  chains onto the last anchor and dates every point on the calendar month end,
  each guard refuses on its own, the proxy's incomplete month never reaches the
  file, and a round trip through the CSV is idempotent (the export era comes
  back byte for byte and the rebuilt tail is identical).

## Trade-offs

- `MSCIWORLD`/`SP500` are deliberately non-investable. Net TR sits marginally
  below a hypothetical zero-cost investor (dividend withholding tax) but is the
  standard published benchmark, and matches the refdata we already ship.
- One small fetch-path branch is the whole code cost; the SIM convention is
  untouched for every other asset.
