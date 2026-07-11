# FIRE explorer — critique and enrichment drop (2026-07-11)

Status: approved (autonomous session, per standing instructions). Follow-up to
`docs/decumulation-fire-program-2026-07.md` and the terminal redesign
(`docs/fire-redesign-terminal-design.md`).

Ben's brief: analyse and critique the `-fire` mode, propose improvements and
above all complements; make the dashboard richer and better illustrated while
keeping the clean, professional design. frontend-design skill mandatory before
touching the UI (dataviz before any new chart).

## 1. Assessment

The engine is genuinely state of the art for a personal tool: multi-model
epistemic strip (Student-t / sequence stress / JST broad-sample / lost decade),
CAPE anchoring, four withdrawal rules (fixed, flex+WR trigger, Guyton-Klinger,
VPW) plus ratchet, French multi-envelope taxes, Gompertz mortality and the
Rich-Broke-Dead decomposition, glidepath, partial annuitisation, buffer bucket
rule. Most of the canonical literature is already in. The critique therefore
concentrates on (a) bugs that currently poison the portfolio mode, (b) data
staleness, (c) chart/readability defects, (d) the handful of genuinely missing
*illustrations* of sequence-of-returns risk, the literature's core insight.

### Bugs found (live run, `pofo -fire examples/fire-decumulation-core.txt`)

1. **CRITICAL — `runFire` ignores `#meta sim:on`.** The 2026-07-10 de-suffix
   campaign (a43133d, 5de1e5e) moved every example to bare ids + `#meta
   sim:on`; `portfolio.Build` honours the flag but `runFire` fetches
   `h.ID` raw. The FIRE panel therefore only sees the *real* quote overlap of
   NTSG/DBMFE/KMLM/... — under 12 common months — instead of the deep SIM
   backcasts. Consequences observed: `FitParametric` returns `Fit{0,0,0}`,
   the UI seeds µ=0/σ=0/df=0, the central case shows **58% ruin** at a 3.3%
   withdrawal (the recurring doom-bug signature), §06 shows 100% ruin,
   -100%/yr worst CAGR, and the Block-bootstrap column compounds a handful of
   months into a 6.4 G€ median wealth with safe WR pinned at the 15% solver
   cap. Fix: honour `spec.Sim` in `runFire` (same SIM-suffix + fallback rule
   as `Build`), and guard the degenerate cases anyway (2-4 below).
2. **Degenerate-fit guard.** `FitParametric` silently returns zeros for
   panels shorter than ~2 years and the UI seeds the sliders with them. Zeros
   must never reach the sliders: keep the parametric defaults, and say why.
3. **Garbage-in columns.** The Historical-windows / Block-bootstrap columns
   are shown (and solved, 20 bisection steps) even when the panel holds only a
   few months. Below a minimum common history (24 months) drop them with an
   explicit note instead of printing nonsense.
4. **Saturating stats.** With any ruined path, "Worst 10y real CAGR (min)"
   prints -100.0%/yr and CDaR 100.0% — true but uninformative. Compute these
   detail stats on surviving paths (label them so), keep ruin itself as the
   failure measure.
5. **Wealth formatting.** "3114k€" / "5451800k€" must render as M€ past 1M.

### Data staleness

6. **CAPE bundled series ends 2023-09 (30.8) and is presented as "Valuation
   now".** The datahub mirror stopped computing PE10 in 2023; multpl.com
   carries Shiller's current monthly CAPE (checked live: 42.18 on
   2026-07-10 vs 30.81 bundled — implied real 2.4%, not the 3.2% shown; the
   percentile moves from 95th to ~99th). Fix: `gen-cape` gains a multpl
   monthly-table fallback appended after the mirror's last PE10 month;
   `make cape` regenerates; the UI prints the as-of date and warns when the
   snapshot is older than 12 months (the tool must never silently present a
   3-year-old valuation as "now").

### Chart / dataviz defects (all in the healthy parametric mode)

7. **Buffer arbitrage is a dual-axis chart** (`chart.LineDual`) — banned by
   the project's own dataviz rules (fire-redesign doc: "No dual-axis charts").
   With ruin ~12% on a 0-100 axis the curve reads flat; the interior optimum
   the copy promises is invisible. Replace with two stacked single-axis
   panels sharing the x axis (ruin %, terminal median), auto-scaled.
8. **§02 spending fan is confusing when no flex rule is active**: a flat
   60 k€ line with a ruin-tail blob at the right end. It needs the pension
   overlay to be visible and a solvency-conditional band, plus a clear
   "fixed rule = flat by construction" empty-state note.
9. **"Why plans fail" shares can sum to 101%** (independent rounding).
10. **The model strip prints Median wealth in k€ with no thousands grouping**
    (see 5) and safe-spend sub-percentages at one decimal where two matter
    (0.7% vs 0.72% at 1.8 M€ is 4 k€/yr).

### What the literature says is missing (the complements)

The tool quantifies sequence risk but never *shows* it. The three additions
below are the highest-value illustrations per unit of code, all engine-side
computable from existing pieces:

- **A — The decisive decade (SoRR made visible).** Bucket paths by realised
  first-10-years real CAGR (quintiles); show ruin probability and median
  terminal wealth per bucket. The point (Kitces, ERN): the same average
  return with a bad first decade is what kills a plan — conditional ruin in
  the worst quintile is typically several times the headline. One new
  `Ensemble` method + one CategoryBars chart + copy.
- **B — Infamous vintages (historical cohort replay).** Replay the *actual*
  named worst retirement start dates through the user's plan: US 1929, US
  1966 (the SWR-defining cohort), US 2000, Japan 1990, plus the user's
  best/worst broad-sample draw. The JST table is bundled with years; a
  year-indexed variant of the parser gives each vintage's return sequence;
  the kernel runs it deterministically (1 path). Chart: one multi-line wealth
  chart, direct-labelled, ruin year marked; cards with "ruined in year N /
  survived with X". This is the cFIREsim-style storytelling the page lacks,
  and it makes "sequence stress" concrete with real dates.
- **C — Where the money comes from (income layers).** A stacked area of the
  median path's yearly funding mix: portfolio withdrawals vs pension vs side
  income vs annuity floor, with the spending line on top. Makes the plan's
  shape tangible (the pre-pension gap years are visibly the dangerous ones)
  and explains *why* pension timing dominates the sensitivity tornado.
  Deterministic from the Plan schedule + median withdrawal profile
  (`SpendBands` + cashflows), StackedArea chart.

Considered and deliberately not in this drop: SWR×horizon heatmap (the two
§04 curves already carry the content), growth-inflation regime lens (own
program, see darcet doc), wealth-trajectory ruin-cause classifier (tracked in
program doc), utility/floor metrics (declined earlier by Ben).

## 2. Design

### Engine (`pkg/decumul`, `pkg/decumul/web`)

- `Ensemble.DecadeBuckets(k int)` (or similar): per-quintile ruin/median
  terminal, bucketed on each path's first-decade realised real CAGR. Annual
  kernel already stores yearly wealth; first-decade return needs the path's
  return sequence — store realised first-10y growth on `PathResult` (cheap).
- `web.vintages`: parse the JST CSV once with years
  (`iso,year → series`), expose named vintages
  (`USA-1929, USA-1966, USA-2000, JPN-1990`) as `scenario.Fixed` single
  sequences (new trivial Source: a fixed Sequence); truncate at data end and
  label partial replays. Kernel runs 1 path deterministically.
- `web.incomeLayers`: median withdrawal per year from `SpendBands` at p50
  net of cashflows; layers = annuity, pension, side income, portfolio.
- New endpoints, same POST-a-Params shape: `/api/decade`, `/api/vintages`,
  `/api/income` (or folded into existing responses where a section already
  fetches — decade+vintages join §01/§03's payloads only if latency allows;
  default: separate endpoints on the fast lane).

### Fixes

- `runFire`: `if spec.Sim { id = simSuffix(h.ID) }` with the same
  fallback-to-bare on fetch error as `portfolio.Build` (reuse the exported
  helper if there is one; extract one if not).
- `FitParametric`: return `(Fit, ok)` or keep zero-value + `Fit.Zero()`
  check at call sites; `/api/meta` omits mu/sigma/df when not fitted and the
  JS keeps defaults; add `panelMonths` to `/api/meta` so the UI can explain.
- `Models`: skip Historical/Block columns when `panel.Periods() < 24`.
- Outcome detail stats conditioned on surviving paths, labels adjusted.
- `fmtWealth` JS + Go card formatting: `≥ 1M€ → x.xx M€`.
- Buffer arbitrage: two stacked `chart.Line` panels (shared x), drop
  `LineDual` use here (primitive stays for the report).
- `gen-cape`: multpl fallback (User-Agent required), `asOf` surfaced, UI
  staleness warning. Regenerate the dataset (network available, verified).
- Causes shares: largest-remainder rounding to sum to 100.

### UI (§ layout)

frontend-design + dataviz skills run before implementation (hard
requirement). Planned placement, keeping the terminal identity and the
numbered-section rhythm:

- §01 gains the vintage replay as a second row: "Simulated futures" (fans)
  then "The futures that actually happened" (vintages chart + outcome cards).
- New §02 "The decisive decade" (decade buckets) between the fans and the
  spending section — it answers the fans' question ("why do some cones
  die?").
- §Income layers joins §02-spending ("The spending you actually live") as its
  left/companion panel — that section becomes "Spending & where it comes
  from", duo layout.
- Renumber the § codes accordingly (pure labels).
- Hero/strip formatting fixes, CAPE as-of + staleness chip in §00.

### Validation

- Unit tests per new engine method (buckets, vintages parser, income
  layers); httptest smoke for new endpoints; formatting tests.
- Golden calibration anchors must not move (no calculation change to the
  kernel itself).
- Live verification of both modes (parametric + portfolio) with screenshots,
  before/after; sanity: portfolio mode central ruin returns to plausible
  single digits at 3% WR with the SIM panel restored, block-bootstrap median
  wealth back to sane magnitudes.
- `make check` green; commit to master.
