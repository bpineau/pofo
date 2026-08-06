# Monthly sampling for the historical return models — design

Status: design, self-approved for planning (2026-06-28). Follow-up to
`decumulation-fire-design.md`.

## 1. Goal

The historical return models (`bootstrap`, `cohorts`) currently resample
**annual** real returns. With ~20 years of history that is only ~20 data
points: the bootstrap captures little intra-year regime structure, and the
cohorts model has almost no usable windows (1 cohort at a 20-year horizon, 6
at 15 years), so it often falls back to the "history too short" note.

Sample at **monthly** frequency instead (~240 points): the bootstrap captures
intra-year regimes and cross-asset correlations (1970s inflation, GFC, 2022),
and cohorts gain many windows (≈61 at a 15-year horizon). This is the
ChatGPT-suggested "monthly block bootstrap" that motivated the original work.

**Scope decision: sampling only.** The decumulation kernel stays **annual**
(annual real withdrawal), so the golden tests that pin it to the Python
reference are untouched. Monthly returns are compounded into annual returns
*before* the kernel sees them. A fully monthly kernel (monthly withdrawals,
intra-year sequence-of-returns) is explicitly out of scope for this iteration.

## 2. Constraints and fit

- Stdlib only, real euros, reuse existing packages (unchanged conventions).
- **`decumul` and its golden tests do not change.** All new behaviour lives
  in `pkg/scenario` (one small generic wrapper) and the `pkg/decumul/web`
  portfolio adapter.
- The `scenario.Source` contract is preserved: a Source still returns an
  **annual** `Sequence` of length `Years`.

## 3. Architecture

The key seam is a generic frequency-aggregation wrapper. The existing
`BlockBootstrap`, `StationaryBootstrap` and `HistoricalCohorts` run unchanged
over a *monthly* panel (with `Periods = Years*12`, block lengths in months),
producing a monthly path; a wrapper compounds each block of 12 monthly returns
into one annual return, yielding the annual `Sequence` the kernel expects.

```
pkg/scenario/
  + Annualize(s Sequence, group int) Sequence   // compound consecutive
                                                 // groups of `group` returns:
                                                 // out[k] = Π(1+s[k*group+j]) − 1
  + Compounded{Inner Source; Group int}          // Source wrapper:
                                                 //   Draw = Annualize(Inner.Draw, Group)
                                                 //   Len  = Inner.Len() / Group
```

`Compounded` is generic (any frequency → any coarser frequency) and reuses all
existing samplers verbatim. It is the only new engine code.

```
pkg/decumul/web/
  BuildMonthlyPanel(assets, hicp)   // last close of each MONTH, deflated by
                                     // ^HICP-FR via scenario.Deflate; common
                                     // window in months. Replaces the annual
                                     // BuildPanel.
  FitParametric(panel, weights)      // annualises the monthly panel (combine →
                                     // Annualize(_,12) → mean/stdev) to seed
                                     // the annual mu/sigma sliders.
  source(panel)                      // historical models built over the
                                     // monthly panel with Periods = Years*12,
                                     // wrapped in Compounded{Group: 12}.
```

**Data flow (portfolio mode):** holdings fetched (SIM-extended) → last close
per month, deflated by `^HICP-FR` → monthly `scenario.Panel` → `source()`
builds a monthly `StationaryBootstrap` / `HistoricalCohorts`, wraps it in
`Compounded{12}` → the `decumul` kernel receives annual sequences exactly as
before. The parametric model is unchanged (annual, from sliders seeded by the
annualised fit).

**Defaults:** the historical bootstrap becomes a **stationary bootstrap with a
mean block of 24 months** (preserves regimes, avoids fixed-block artefacts),
replacing the current fixed 5-year block. HICP is sampled at or before each
month-end by the existing `Deflate`, so its native (monthly) granularity needs
no special handling.

## 4. Integration detail: cohorts availability note

`computeFrom` currently type-asserts `p.Source.(scenario.HistoricalCohorts)`
to detect a horizon longer than the history and show an honest note instead of
all-zero (certain-ruin) paths. With monthly sampling the source becomes
`Compounded{Inner: HistoricalCohorts}`, so the check must unwrap the wrapper:

```go
src := p.Source
if c, ok := src.(scenario.Compounded); ok {
    src = c.Inner
}
if hc, ok := src.(scenario.HistoricalCohorts); ok && hc.Count() == 0 {
    // note: not enough history for this horizon
}
```

`Compounded.Inner` is therefore an exported field. The note wording reports the
history length in **years** (`Inner.Panel.Periods()/12`).

## 5. Testing

- `Annualize`: a known monthly sequence compounds correctly (12 zeros → one 0;
  a known product → the expected annual return); length is `len/group`.
- `Compounded`: wrapping a stub monthly Source yields annual paths of
  `Len()/Group` and the compounded values.
- `BuildMonthlyPanel` + `FitParametric`: a synthetic monthly series with known
  drift and zero inflation yields the expected annualised mu; the panel has the
  expected number of monthly periods.
- Cohorts availability: over a ~240-month panel, a 15-year horizon yields many
  windows (≈61), versus ≤6 with annual sampling — asserted via the inner
  `HistoricalCohorts.Count()`.
- The `decumul` golden tests are unchanged and must still pass.

## 6. Caveats

- Monthly real returns are sampled from last-of-month closes deflated by HICP;
  this is total-return only where the underlying quotes are total-return (the
  existing simdata/proxy caveats carry over).
- Cohorts still cannot exceed the available history (≈20 years): the existing
  note covers that case, now expressed in years from the monthly panel.
- Annual compounding of a resampled monthly path is the realised annual return
  of that synthetic path; block boundaries need not align with year boundaries,
  which is correct and intended.
```
