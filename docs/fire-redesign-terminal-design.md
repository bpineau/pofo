# FIRE explorer redesign — terminal direction

Ground-up redesign of the `pofo -fire` page. Direction chosen with Ben
(2026-07-05): **terminal instrument** (dark, dense, amber-phosphor). Mocks:
`fire-terminal-redesign` artifact (ec1a81c4). Process: frontend-design →
dataviz (both done) → implementation.

## Identity

Warm-charcoal ground, amber-phosphor accent, mono-forward typography, a
Bloomberg-terminal vernacular made modern. Single-theme dark (a committed
world). Signature: the amber verdict readout with a faint underglow, and a
terminal status-line footer / command-echo header.

- **Type:** mono leads all data, labels and readouts
  (`ui-monospace,"SF Mono",Menlo`); a system sans for descriptive prose only.
- **Layout:** settings fold into a **top drawer** (command bar echoes the params),
  so the analysis column runs full-width. Dense, large, rich; sections keyed
  `§NN`. Charts get real margins and proportion (the current ones are the fault).

## Color tokens (chrome / surfaces)

```
--bg      #100E0A   warm near-black ground
--panel   #17130D   card / chart surface
--well    #221B11   inset (gauge track, meter)
--line    #2B2317   hairline grid
--line2   #3C3121   structural rule
--ink     #EFE7D6   primary text
--ink2    #B4A991   secondary text
--muted   #7E7458   axis / caption
--amber   #F5A623   HERO readout / accent (text + glow only, not a chart mark)
```

## Chart palette (dataviz-validated, dark surface #17130D)

Ran `scripts/validate_palette.js --mode dark --surface #17130D`: all six checks
PASS (lightness band 0.48–0.67, chroma ≥ 0.10, worst-adjacent CVD ΔE 43, contrast
≥ 3:1). **Categorical marks must use the deep set, not the bright hero amber.**

```
categorical (fixed order, entity identity, never cycled):
  1 #C4820F  amber/gold   (Student-t / central / primary series)
  2 #1B95B4  teal-cyan    (secondary series; fan median & bands)
  3 #2F9463  green
  4 #7A5FC0  violet
  5 #BE4E6C  rose
status (reserved, with icon/label, never reused as a series):
  good #34A46E   warn #D98A2E   bad #D2503F
```

Rule: **bright `#F5A623` = hero readout text/glow; deep `#C4820F` = amber chart
mark.** Text (values/labels/legends) wears ink tokens, never a series color.

## dataviz corrections to honour

- **No dual-axis charts.** The CAPE-history chart shows CAPE on one axis only;
  the implied real return is a readout, not a second y-scale.
- Legend present for ≥ 2 series; ≤ 4 also direct-labelled; single series named by
  the title (no legend box). Thin marks, 4px rounded data-ends, ≥ 8px markers,
  2px surface gaps between fills, recessive grid.
- Add a hover/crosshair layer to the live web charts (vanilla JS, no lib); the
  static report keeps plain SVG.

## The reworked "Where we are in the cycle" section (§00)

The old lone radial gauge was the worst offender ("dessin d'enfant"). Replace it
with **two charts**:

1. **Valuation strip** — a horizontal cheap→rich gradient scale with percentile
   ticks (median, today), a bright amber "today" marker, and a big amber implied-
   real-return readout (`1/CAPE`). Needs `capeSnapshot` (already in `pkg/decumul/web`).
2. **CAPE since 1881** — a single-axis line of the bundled Shiller series
   (`datasets.CAPE()`), today marked, historical median dashed, the 1929/2000
   peaks annotated. New `pkg/chart` line use + a `/api/cape` (or meta) SVG.

## Implementation scope (big; own plan/spec)

1. `pkg/chart/theme.go` — dark token values (the reskin surface is ready); update
   the chart snapshot golden deliberately.
2. `pkg/decumul/web/assets` — rewrite `app.css` (dark terminal), `index.html`
   (top drawer + `§` sections + 2-chart cycle section), `app.js` (drawer toggle,
   hover layer, wire the CAPE strip/history).
3. New chart uses/endpoints: valuation strip, CAPE-history line.
4. Rework proportions/margins of the existing chart primitives to the mock.

See [[fire-page-redesign]] memory, `docs/decumulation-fire-program-2026-07.md`.
