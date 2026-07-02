# Web UI redesign: the "instrument" identity

Status: approved (Ben, 2026-07-03) from mock direction C of three
(instrument / research note / terminal). Mocks:
https://claude.ai/code/artifact/f0db4eb5-0e71-4748-b57b-592d37609a0d

## Why the previous themes failed

Two identities were rejected in a row ("warm study", then "quant desk",
commit 0be7d89). The post-mortem, confirmed against locador (a reference
Ben rates highly, ../locador/server/assets/style.css): the failures were
never the palette family but the execution. Quant-desk shipped the system
font stack, a recognizable free-UI-kit palette (Untitled UI grays), and
native `accent-color` range sliders. Locador ships embedded web fonts,
fully custom-styled controls and crafted shadows.

Three rules follow, and they are permanent for pofo's web surfaces:

1. Real embedded fonts, never the system stack for identity roles.
2. No default-rendered form controls, ever. Sliders, checkboxes, number
   inputs and buttons are all styled to the theme.
3. No palette lifted from a UI kit; neutrals are hue-biased and chosen,
   accents are validated (dataviz six checks) against our surfaces.

## The identity

Precision-instrument register: cool light ground, white panels, one petrol
accent, machined controls, monospace figures everywhere. What quant-desk
aimed for, executed at the level of a good design studio.

### Tokens (pkg/webui/theme.css, same variable names as today)

| Token | Value | Role |
|---|---|---|
| `--bg` | `#F3F4F6` | page ground, cool gray |
| `--surface` | `#FFFFFF` | panels, cards |
| `--surface-2` | `#EEF0F3` | wells, insets, zebra |
| `--ink` | `#16181D` | primary text |
| `--ink-soft` | `#4A5160` | secondary text |
| `--muted` | `#7A8294` | captions, axis labels |
| `--faint` | `#A8AEBC` | placeholders |
| `--line` | `#E3E6EB` | hairline |
| `--line-strong` | `#CDD2DA` | structural rule |
| `--accent` | `#0B7285` | petrol, the one brand color |
| `--accent-ink` | `#0A6376` | accent tuned for text on white |
| `--accent-wash` | `#E5F2F5` | selected / highlighted surfaces |
| `--good` / `--good-ink` | `#0C8A47` / `#0A7038` | risk: safe |
| `--warn` / `--warn-ink` | `#C77E17` / `#9E6410` | risk: caution |
| `--bad` / `--bad-ink` | `#D2402F` / `#B23425` | risk: danger |
| washes | derived at ~9 % alpha of each risk hue | badges, notes |

Radii 12px (panels) / 7px (small), shadow
`0 1px 2px rgba(22,24,29,.04), 0 8px 24px -18px rgba(22,24,29,.25)`
(a crisp contact line plus a soft distant drop; that pairing is what makes
panels read "finished" instead of flat).

Risk colors stay strictly semantic (dots, chips, badge text), never
decorative. Dense full-width layouts stay.

### Typography (embedded, go:embed)

- UI and display: Instrument Sans 400 / 500 / 600.
- Figures and code: Spline Sans Mono 400 / 600, `tabular-nums`
  everywhere digits align.
- Files: latin-subset WOFF2 under `pkg/webui/fonts/` with their OFL
  license; a generated `webui.FontsCSS` exposes `@font-face` rules with
  data: URIs so the report stays a single self-contained file and the
  FIRE server serves it at /fonts.css. No network fetch at runtime, no Go
  dependency (data via go:embed only). Cost: ~210 KB in the binary.

### Controls (the signature)

- Range sliders: 4px track, filled left portion in accent (JS sets
  `--fill`), 16px white thumb with 1.5px accent ring and contact shadow;
  a ruler of ticks under plan-defining sliders.
- Value readouts: mono, in a `--surface-2` well with hairline border.
- Presets: segmented control (well background, white raised active
  segment), not chips.
- Checkboxes: accent `accent-color` is acceptable there (small, themed).
- Number inputs, buttons: themed borders, focus ring in accent.

### Hero verdict (FIRE)

The verdict panel becomes an instrument readout: eyebrow, large mono ruin
figure, status chip (good/warn/bad wording + dot), a tolerance gauge (6px
well, solid `--good` fill for the measured value, ink tick at the
tolerance mark, mono scale labels), then the one-sentence plan summary
with bold figures. The model matrix keeps its columns; the central column
is washed `--accent-wash` with an accent-ruled header.

### Charts

- Categorical series: re-derive the 8-slot palette from the instrument
  hues, starting [#0880A8 petrol, #6D28D9 violet, #B45309 ochre,
  #BE185D magenta] and extending to 8; order chosen by the dataviz
  validator (CVD >= 12 target) on `#FFFFFF`; single source of truth
  pkg/chart/svg.go, mirrored in app.js PAL.
- Fan bands: petrol at .14 / .30 alpha, median line `--accent` 2px, ruin
  traces `--bad` 1.2px at .7 opacity.
- Grid `#EDF0F3` hairlines, axis `#CDD2DA`, labels `--muted` in the mono
  face; direct end labels with collision-avoidance (min 12px apart).
- Alignment and proportions (explicit requirement): charts in the same
  row share the same pixel height and margin box; each SVG renders at its
  display slot's real pixel size (no viewBox scaling that distorts label
  sizes); axis label columns inside a row use one width so plot areas
  align edge to edge.

## Scope

- `pkg/webui/`: theme.css re-skin (same var names), fonts/ + fonts.go +
  FontsCSS, doc updates.
- `pkg/decumul/web/assets/`: index.html (hero readout structure, fonts
  link), app.css (controls, panels, matrix per above), app.js (slider
  fill + readout painting, chart chrome colors, PAL, label
  de-collision, row-height alignment).
- `pkg/report/html.go`: inline FontsCSS; verify components on new tokens.
- `pkg/chart/svg.go`: new categorical palette + chart chrome neutrals.
- Non-goals: terminal (ANSI) charts, report layout rework, any change to
  computations (make golden must stay green untouched).

## Risks

- Report and FIRE page weight grows ~210 KB (fonts): accepted.
- finador reuses pofo charts: palette change is intentional and applies
  there too; verify finador still builds/tests after API-neutral change.
- Font licensing: OFL files shipped alongside; note in fonts/README.
