# Web UI "Instrument" Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Re-skin pofo's two HTML surfaces (FIRE explorer + report tearsheet) to the approved "instrument" identity (design doc `docs/webui-instrument-redesign.md`), with embedded fonts, custom controls and a validated chart palette.

**Architecture:** Token values change inside `pkg/webui/theme.css` while variable NAMES stay identical, so `pkg/report` and `pkg/decumul/web` re-skin automatically; fonts ship as go:embed WOFF2 exposed as a data-URI stylesheet (`webui.FontsCSS`); the chart palette changes at its single source (`pkg/chart/svg.go`) and its one mirror (`app.js` PAL).

**Tech Stack:** Go stdlib only, go:embed, hand-written CSS/JS, headless Chrome for visual verification.

**Reference implementation:** the approved mock, committed at
`docs/superpowers/plans/2026-07-03-webui-instrument-mock.html` (`.dC`
scoped styles = direction C). Copy values from there, not from memory.

## Global Constraints

- Stdlib only; go:embed data is fine, Go dependencies are not.
- `make check` green before every commit; `make golden` untouched (no computation changes).
- Tests never touch the network.
- Risk colors (`--good/--warn/--bad`) strictly semantic, never decorative.
- All figures in the mono face with `tabular-nums`.
- No em-dash anywhere. English code/docs.
- Charts in one row: same pixel height, aligned plot edges; SVG rendered at its slot's real pixel size.
- Validated categorical palette order (do not reorder):
  `#0880A8 #C2452B #4C63D2 #B45309 #6D28D9 #35803B #BE185D #2B9BD0`
  (dataviz validator: all checks pass on #FFFFFF, worst adjacent ΔE 15.0).

---

### Task 1: Embedded fonts (pkg/webui)

**Files:**
- Create: `pkg/webui/fonts/` (5 WOFF2: InstrumentSans-{400,500,600}, SplineSansMono-{400,600}, latin subsets), `pkg/webui/fonts/OFL.txt`, `pkg/webui/fonts.go`, `pkg/webui/fonts_test.go`
- Modify: `pkg/webui/theme.go` (doc comment: fonts now embedded)

**Interfaces:**
- Produces: `webui.FontsCSS string` (complete `@font-face` CSS, data: URIs), consumed by Task 3.

- [ ] **Step 1: Fetch the latin WOFF2 subsets** (dev machine, one-off; tests stay offline). Use the Google Fonts css2 endpoint with a Chrome UA, keep only the `/* latin */` blocks:
  `family=Instrument+Sans:wght@400;500;600` and `family=Spline+Sans+Mono:wght@400;600`.
  Save as `pkg/webui/fonts/instrumentsans-{400,500,600}.woff2`, `pkg/webui/fonts/splinesansmono-{400,600}.woff2`. Add the OFL 1.1 text as `fonts/OFL.txt` with a header naming both families and their copyright lines.

- [ ] **Step 2: Write the failing test** (`pkg/webui/fonts_test.go`):

```go
func TestFontsCSS(t *testing.T) {
	for _, want := range []string{
		"font-family:'Instrument Sans';font-style:normal;font-weight:400",
		"font-weight:500", "font-weight:600",
		"font-family:'Spline Sans Mono'",
		"data:font/woff2;base64,",
	} {
		if !strings.Contains(FontsCSS, want) {
			t.Errorf("FontsCSS missing %q", want)
		}
	}
}
```

- [ ] **Step 3: Implement `fonts.go`**: `//go:embed fonts/*.woff2` into an `embed.FS`; build `FontsCSS` in an init-time loop over a small table `{file, family, weight}` emitting `@font-face{font-family:'%s';font-style:normal;font-weight:%d;font-display:swap;src:url(data:font/woff2;base64,%s) format('woff2')}` with `base64.StdEncoding`. Godoc explains the licensing (OFL, file shipped) and the no-network rationale.

- [ ] **Step 4: `go test ./pkg/webui/` passes; commit** `webui: embed Instrument Sans and Spline Sans Mono (OFL) as FontsCSS`.

### Task 2: Re-skin theme.css (tokens + shared components)

**Files:**
- Modify: `pkg/webui/theme.css` (keep every existing variable NAME; values from the design doc table)

- [ ] **Step 1: Swap token values** to the instrument set (design doc table): bg #F3F4F6, surface #FFFFFF, surface-2 #EEF0F3, ink #16181D, ink-soft #4A5160, muted #7A8294, faint #A8AEBC, line #E3E6EB, line-strong #CDD2DA, accent #0B7285, accent-ink #0A6376, accent-wash #E5F2F5, good #0C8A47/#0A7038, warn #C77E17/#9E6410, bad #D2402F/#B23425, washes at ~9% alpha; `--sans:"Instrument Sans",system-ui,...`, `--mono:"Spline Sans Mono",ui-monospace,...`; `--r:12px;--r-sm:7px`; `--sh:0 1px 2px rgba(22,24,29,.04),0 8px 24px -18px rgba(22,24,29,.25)`.

- [ ] **Step 2: Add the shared control styles** (new sections in theme.css, lifted from the mock's `.dC` rules): fully custom `input[type=range]` (4px track, accent-filled left via `--fill`, 16px ringed thumb, both -webkit and -moz), `.seg` segmented control, `.well` value readout, themed number inputs/buttons, focus ring. Delete nothing that report/app.css still reference (grep `var(--` in `pkg/report/html.go` and `app.css` to confirm names).

- [ ] **Step 3: `make check`; commit** `webui: instrument identity tokens and machined shared controls`.

### Task 3: Serve and inline the fonts

**Files:**
- Modify: `pkg/decumul/web/server.go` (serve `/fonts.css`), `pkg/decumul/web/server_test.go`, `pkg/decumul/web/assets/index.html` (`<link rel="stylesheet" href="fonts.css">` before theme.css), `pkg/report/html.go` (inline `webui.FontsCSS` into the document `<style>` next to Theme)

- [ ] **Step 1: Extend `server_test.go`**: GET `/fonts.css` returns 200, `text/css`, body containing `Instrument Sans` (fails first).
- [ ] **Step 2: Implement the handler** (mirror of the `/theme.css` one) and the two inclusions.
- [ ] **Step 3: `make check`; commit** `webui: ship the embedded fonts to both HTML surfaces`.

### Task 4: Chart palette and chrome (pkg/chart)

**Files:**
- Modify: `pkg/chart/svg.go` (defaultPalette + godoc; grid/axis/label neutrals if hexed there), `pkg/chart/fan.go` (bandFill `#2E4BE0` -> `#0B7285`), `pkg/chart/term.go` (ansiPalette remapped: `[31,166,62,172,93,28,162,39]`)

**Interfaces:**
- Produces: `PaletteColor(i)` returning the validated order above; Task 5 mirrors it.

- [ ] **Step 1: Update `defaultPalette`** to the Global-Constraints order, with a godoc note that the order is CVD-validated and must not be casually permuted.
- [ ] **Step 2: Sweep remaining quant-desk hexes** in pkg/chart (`grep -rn "2E4BE0\|0E9384\|E19000\|7A5AF8\|E8622C\|0BA5EC\|067647\|C11574" pkg/chart`) and remap each to its same-slot successor.
- [ ] **Step 3: `make check && make golden`; commit** `chart: instrument series palette (CVD-validated order)`.

### Task 5: FIRE explorer views (index.html, app.css, app.js)

**Files:**
- Modify: `pkg/decumul/web/assets/index.html` (hero verdict -> instrument readout: eyebrow, big mono figure, status chip, tolerance gauge, sentence), `app.css` (panels/matrix/rail per mock `.dC`), `app.js` (PAL mirror, slider `--fill` painting, gauge painting, chart chrome colors, frontier/end-label de-collision, row alignment)

- [ ] **Step 1: Hero markup** (from the mock, adapted to live IDs kept for app.js):

```html
<div class="hero-verdict">
  <span class="eyebrow">Ruin probability · central case</span>
  <div class="big"><span class="n" id="ruinBig"></span><span class="u">of simulated futures fail</span></div>
  <span class="chip" id="ruinChip"></span>
  <div class="gauge"><div class="fill" id="gaugeFill"></div><div class="lim" id="gaugeLim"></div></div>
  <div class="gauge-l"><span>0%</span><span id="gaugeTol"></span><span id="gaugeMax"></span></div>
  <p id="verdict"></p>
  <div class="herobar">…acceptable-ruin input + confidence chip as today…</div>
</div>
```

- [ ] **Step 2: app.js**: paint `#ruinBig/#ruinChip/#gaugeFill(lim)` from the model result (chip classes good/warn/bad by comparison to the target); on every range input set `--fill` = (v-min)/(max-min) and the readout text; PAL = Global-Constraints order; chart ink/grid/axis constants updated (`#EDF0F3` grid, `#CDD2DA` axis, `#7A8294` labels); frontier and fan end labels pass through a de-collision sort (>= 12px apart, as in the mock's `drawFrontier`).
- [ ] **Step 3: app.css**: port the mock's `.dC` component values (rail groups, ticks ruler, seg presets, matrix selected column `--accent-wash` + accent-ruled header, KPI tiles, frames). Alignment rules: every `.duo`/`#riskgrid` row uses `align-items:stretch` and fixed chart heights per row so plot boxes match; keep `#fansGrid` 2x2 equal cells.
- [ ] **Step 4: `make check`; visual pass** (Task 6 script); commit `decumul/web: instrument re-skin of the FIRE explorer`.

### Task 6: Visual verification + cross-repo check

- [ ] **Step 1:** `make build && ./pofo -fire examples/…` (pick the doc'd example), headless-Chrome screenshot at 1500px and 390px; compare against the mock for: slider rendering, hero readout, matrix, chart row alignment/proportions (Ben's explicit requirement), no horizontal scroll.
- [ ] **Step 2:** generate a report HTML (`./pofo` compare mode with `-html`), screenshot, check tokens+fonts landed and nothing regressed.
- [ ] **Step 3:** `cd ../finador && go build ./... && go test ./...` (charts consumer; API unchanged so this should be green).
- [ ] **Step 4:** fix what the screenshots reveal; re-run `make check`.

### Task 7: Docs and closure

- [ ] **Step 1:** update `pkg/webui/doc.go` + theme.css header comment (identity story), `pkg/chart/doc.go` palette note, README screenshot note if any.
- [ ] **Step 2:** `make check`; final commit `webui: document the instrument identity`; push; update memory (webui theme memory: instrument shipped).

## Self-review

- Spec coverage: tokens (T2), fonts (T1+T3), controls (T2+T5), hero (T5), charts+palette (T4+T5), alignment (T5+T6), report (T3+T6), finador (T6), non-goals respected (no golden change; term.go remap is cosmetic ANSI, allowed as chart chrome). Gaps: none found.
- Placeholders: the only "as today" reference (herobar) points at existing checked-in markup, not future work.
- Type consistency: `webui.FontsCSS` name used in T1 and T3; palette order repeated verbatim in T4/T5 via Global Constraints.
