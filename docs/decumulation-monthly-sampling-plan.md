# Monthly sampling Implementation Plan

> Follow-up to `decumulation-fire-plan.md`. TDD, commit+push per task. Golden tests in `pkg/decumul` must stay green throughout.

**Goal:** Sample the historical return models (bootstrap, cohorts) monthly and compound to the annual returns the unchanged decumul kernel expects.

**Architecture:** A generic `scenario.Compounded` wrapper aggregates a monthly inner Source into annual; the web adapter builds a monthly panel and wraps the historical samplers.

## Global Constraints

- Stdlib only; real euros; `pkg/decumul` and its golden tests are not modified.
- Verify each task: `go test ./...` and `go vet ./...` from the repo root.

---

## Task 1: scenario.Annualize and Compounded

**Files:** Create `pkg/scenario/compound.go`, `pkg/scenario/compound_test.go`.

**Produces:** `func Annualize(s Sequence, group int) Sequence`; `type Compounded struct { Inner Source; Group int }` implementing `Source` (`Draw` = Annualize(Inner.Draw, Group); `Len` = Inner.Len()/Group).

- [ ] Test: `Annualize([12 zeros],12) == [0]`; `Annualize([0.1,0.1 ...],...)` compounds to the product; `Compounded` over a stub monthly Source yields annual length and values.
- [ ] Implement `Annualize` (product of (1+r) over each group, minus 1) and `Compounded`.
- [ ] `go test ./pkg/scenario/ && go vet ./pkg/scenario/`; commit+push.

## Task 2: BuildMonthlyPanel and annualised FitParametric

**Files:** Modify `pkg/decumul/web/portfolio.go`, `pkg/decumul/web/portfolio_test.go`.

**Produces:** `func BuildMonthlyPanel(assets []AssetSeries, hicp []marketdata.Point) (scenario.Panel, error)` (last close per month, deflated); `FitParametric` annualises the monthly panel (combine → `Annualize(_,12)` → mean/stdev) so mu/sigma are annual. `lastPerMonth` replaces `lastPerYear`. Remove the now-unused annual `BuildPanel`.

- [ ] Test: a synthetic monthly series (~36 months, +x%/month, flat HICP) gives the expected number of monthly periods and an annualised mu within tolerance.
- [ ] Implement; `go test ./pkg/decumul/web/ && go vet ./...`; commit+push.

## Task 3: monthly source() and cohorts-note unwrap

**Files:** Modify `pkg/decumul/web/model.go`, `pkg/decumul/web/model_test.go`.

**Produces:** `source(panel)` builds the historical samplers over the monthly panel with `Periods = Years*12` (stationary bootstrap, mean block 24 months; cohorts), wrapped in `Compounded{Group:12}`. `computeFrom` unwraps `Compounded` before the `HistoricalCohorts.Count()==0` note; the note reports history length in years (`Inner.Panel.Periods()/12`).

- [ ] Test: over a ~240-month panel a 15-year horizon yields many cohort windows (>30) via the inner `Count()`; a 30-year horizon still returns the note.
- [ ] Implement; `go test ./pkg/decumul/web/`; commit+push.

## Task 4: cmd runFire uses BuildMonthlyPanel

**Files:** Modify `cmd/pofo/main.go`.

- [ ] Replace `web.BuildPanel` with `web.BuildMonthlyPanel` in `runFire`.
- [ ] `go build ./cmd/pofo && go vet ./...`; live smoke-test the three models on `60 NTSGSIM / 25 DBMFESIM / 15 XAUUSD` (cohorts now available at a 15y horizon); commit+push.

## Task 5: docs

**Files:** Modify `README.md` (note the historical models sample monthly).

- [ ] One-line update in the FIRE section; `go test ./...`; commit+push.

## Self-review

- Coverage: §3 Annualize/Compounded → T1; BuildMonthlyPanel/fit → T2; source()+note → T3; cmd → T4; docs → T5. §4 unwrap → T3. §5 tests → per task. Golden tests untouched (verified each task).
- Types: `Annualize(Sequence,int) Sequence`, `Compounded{Inner Source; Group int}`, `BuildMonthlyPanel(...) (scenario.Panel, error)` consistent across tasks.
```
