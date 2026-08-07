# FIRE increment A — broad-sample empirical panel + sanity guard

Part of the FIRE improvement program (see the assessment thread). This is the
first shippable increment: give the decumulation engine an *empirical* long-run
sequence-risk model, and a guard that flags implausible return calibrations.

## Why

Today every model is synthetic (Student-t, mean-preserving MarkovRegime,
bootstraps over a short 1999+ portfolio window). None sees the persistent
multi-decade real bear markets that actually cause ruin (Japan post-1990,
1929-32, the 1970s). Anarkulova/Cederburg show a fixed 4% rule fails far more
often on broad 1870+ data than US backtests imply. We bundle a real broad-sample
panel and bootstrap from it, so "sequence risk" becomes empirical, and the
literature anchor is automatic.

## Data source (verified reachable)

JST Macrohistory Database R6 (Jordà, Schularick, Taylor): 18 advanced economies,
1870-2020, annual, with equity total return (`eq_tr`), bond total return
(`bond_tr`), bill rate (`bill_rate`), CPI and nominal GDP per country. Free to
use with citation. The `.xlsx` download redirects to a CDN and returns 200; it is
a zip of XML, parseable with stdlib `archive/zip` + `encoding/xml` (no
dependency), so it fits the generator pattern.

## Decisions (made, not open)

- **World aggregation: GDP-weighted** across the 18 countries per year (matches
  the broad-sample spirit; equal-weight over-counts small markets). Missing
  country-years are dropped from that year's weights.
- **Assets: equity, bonds, bills** (three panel rows), not equity-only: bonds
  are needed for the allocation/glidepath work in increment D and for an honest
  "more bonds is not automatically safer" comparison.
- **Real returns**: nominal total return deflated by the same country's CPI,
  `(1+nominal)/(1+cpi_infl) - 1`, before aggregation.
- **Generator lives behind a `make` target** (`make broadsample`), a small Go
  program under `internal/` or `cmd/`, mirroring `make simdata`. The runtime
  never fetches JST; it reads the committed CSV via `go:embed`.

## Deliverables

### 1. Bundled data
`pkg/datasets/broadsample/world-real.csv`, columns `year,equity,bond,bill`
(real annual returns as fractions), 1870-2020, with a header comment carrying
the JST citation and the regen command. Embedded via the existing
`pkg/datasets` `go:embed`.

### 2. Generator
`make broadsample`: downloads the JST xlsx, parses the relevant sheet with
`archive/zip`+`encoding/xml`, computes per-country real returns, GDP-weights to
a world series, writes the CSV. Documented, reproducible, network-only at
generation time. A `-dry` mode prints coverage without writing.

### 3. Loader + Source (`pkg/scenario`)
- `BroadSamplePanel() (Panel, error)` reads the embedded CSV into a `Panel`
  (rows equity/bond/bill, `Weights` defaulting to 100% equity).
- No new Source type: `StationaryBootstrap{Panel: bs, MeanBlock: 8, Periods: N}`
  already gives the empirical sequence-risk path. A thin
  `BroadSampleSource(weights, periods) Source` convenience wraps it.

### 4. Web wiring (`pkg/decumul/web`)
Add a **"Broad sample (1870+)"** model column to `/api/models` and the model
strip, driven by the bundled panel at the user's weights. It sits beside the
Student-t / Regime / Conservative columns as the empirical counterpoint. Uses
the existing multi-model plumbing; no UI redesign.

### 5. Sanity guard (`pkg/decumul`)
`func Plausibility(e Ensemble) []string` (or on the model summary): computes the
ensemble's effective real **geometric** mean and the 30y safe-WR at 95%, and
returns a warning string when either leaves a documented anchor band
(geometric mean outside [0%, 8%], 30y safe-WR outside [2%, 5%]). Surfaced in the
existing confidence/caveat area. This locks the "doom-loop" regression (effective
real return silently too low) that has recurred; a test asserts a deliberately
broken calibration trips it and a sane one does not.

## Testing

- `scenario`: `BroadSamplePanel` loads, has ~150 periods, three rows, plausible
  moments (world real equity geometric ~5%/yr, vol ~18%); a bootstrap over it
  preserves the panel mean within tolerance.
- `decumul`: `Plausibility` trips on a −2% geometric source, passes on the
  central Student-t; a golden-style check that broad-sample ruin at fixed 4%/30y
  lands in the Anarkulova-class band (higher than the US-fit model).
- `make check` and `make golden` green; the generator is not run in tests
  (network-free), the committed CSV is the fixture.

## Docs

- `pkg/scenario/doc.go`: the empirical panel and its provenance/citation.
- `pkg/datasets/broadsample/README` (schema + regen recipe + citation).
- CLAUDE.md map row for `pkg/datasets/broadsample` and the `make broadsample`
  target; README note that the FIRE explorer offers a broad-sample model.
- The `long-history-data-sources` memory gains the JST provenance.

## Out of scope (later increments)

CAPE conditioning (B), VPW/frontier/cause-of-ruin (C), glidepath/annuity (D),
the visual-system pass (E). This increment only adds the empirical model and the
guard; it reuses the existing charts and model strip.
