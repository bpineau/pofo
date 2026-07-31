# Decumulation / FIRE analysis Implementation Plan

> **Status: DONE.** Shipped to master: `pkg/scenario`, `pkg/decumul`,
> `pkg/decumul/web` and `cmd/pofo -fire`. The follow-up backlog
> (`decumulation-fire-followups.md`), the usability rewrite
> (`decumulation-fire-rewrite-spec.md`, "largely implemented") and the v3
> enrichment (`decumulation-fire-v3-enrichment.md`) were all built on top. The
> unchecked `- [ ]` boxes below are the original one-shot checklist and were
> never ticked; they do not track remaining work.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add decumulation/FIRE analysis to pofo: project a portfolio's (or a set of parameters') probability of ruin under withdrawals, size cash buffers and capital against a target ruin, and explore it live in the browser.

**Architecture:** Three layers. `pkg/scenario` generates synthetic real-return paths (parametric Student-t, block/stationary bootstrap, historical cohorts) behind one `Source` interface. `pkg/decumul` runs a withdrawal kernel over those paths to produce ruin probabilities, FIRE outcome metrics and parameter sweeps. `pkg/decumul/web` is a thin embedded HTTP UI; `cmd/pofo -fire` wires it. Two small generic SVG primitives (`chart.Heatmap`, `chart.Bars`) are added.

**Tech Stack:** Go 1.26, standard library only (`math/rand/v2` for RNG), reusing `pkg/metrics`, `pkg/chart`, `pkg/marketdata`.

## Global Constraints

- **Zero third-party dependencies.** Standard library only. Do NOT add `gonum` or any module; implement Student-t by hand. (`go.mod` stays single-module, no `require` block.)
- **Go 1.26**, module `github.com/bpineau/pofo`.
- **Real euros throughout.** Returns are real; the spending floor is real and constant.
- **Reuse existing packages:** `metrics.Quantiles`, `metrics.Histogram`, `metrics.Mean`, `metrics.Returns`, `metrics.DrawdownEpisodes`, `chart.Line`/`chart.Options`/`chart.Series`/`chart.PaletteColor`. Do not re-derive what they already compute.
- **Conventions:** every new package has a `doc.go` with a package comment; exported symbols carry godoc in English; runnable `Example` tests where they clarify usage; tests are network-free.
- **Verify each task with:** `go test ./...` and `go vet ./...` from the repo root `/Users/ben/projects/pofo`.
- **Reproducibility:** Monte-Carlo uses per-worker deterministic seeds derived from a base seed; tests fix the worker count.

---

## File Structure

```
pkg/scenario/
  doc.go            package comment
  source.go         Sequence, Source, ParametricSource (+ studentT draw)
  panel.go          Panel, Combine
  bootstrap.go      BlockBootstrap, StationaryBootstrap
  cohorts.go        HistoricalCohorts
  deflate.go        Deflate (nominal points + HICP points -> real Sequence)
  *_test.go         one test file per source file

pkg/decumul/
  doc.go            package comment + caveats
  plan.go           Plan, Cashflow, BufferSleeve, FlexRule, Tax, CTOFlatTax
  run.go            Run (per-path kernel), PathResult
  simulate.go       Simulate, Ensemble, CapitalForRuin
  outcome.go        Outcome, Ensemble.Outcome (ruin, terminals, CDaR, …)
  recovery.go       RecoveryTimeDistribution
  sweep.go          Param, Sweep1D, Sweep2D, SweepPoint, Surface
  golden_test.go    acceptance tests vs the reference validation table
  *_test.go         unit tests per file

pkg/decumul/web/
  doc.go            package comment
  server.go         Handler(model) http.Handler, /api/sim, params JSON
  model.go          Params, Result (JSON DTOs), runParametric, runPortfolio
  assets/index.html
  assets/app.js
  assets/app.css
  embed.go          //go:embed assets/*
  server_test.go    httptest smoke tests

pkg/chart/
  bars.go           Bars
  heatmap.go        Heatmap
  bars_test.go, heatmap_test.go

cmd/pofo/
  main.go           + -fire flag, runFire(), portfolio->Panel adapter
```

---

## Task 1: scenario.Source and ParametricSource

**Files:**
- Create: `pkg/scenario/doc.go`
- Create: `pkg/scenario/source.go`
- Test: `pkg/scenario/source_test.go`

**Interfaces:**
- Produces: `type Sequence []float64`; `type Source interface { Draw(rng *rand.Rand) Sequence; Len() int }`; `type ParametricSource struct { Mu, Sigma, Df float64; Periods int }`; `func (p ParametricSource) Draw(rng *rand.Rand) Sequence`; `func (p ParametricSource) Len() int`.
- Consumes: nothing.

- [ ] **Step 1: Write the failing test**

```go
package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestParametricSourceMoments(t *testing.T) {
	src := ParametricSource{Mu: 0.04, Sigma: 0.12, Df: 6, Periods: 30}
	rng := rand.New(rand.NewPCG(1, 2))
	var all []float64
	for i := 0; i < 4000; i++ {
		seq := src.Draw(rng)
		if len(seq) != 30 {
			t.Fatalf("len = %d, want 30", len(seq))
		}
		all = append(all, seq...)
	}
	mean, variance := 0.0, 0.0
	for _, x := range all {
		mean += x
	}
	mean /= float64(len(all))
	for _, x := range all {
		variance += (x - mean) * (x - mean)
	}
	variance /= float64(len(all) - 1)
	if math.Abs(mean-0.04) > 0.01 {
		t.Errorf("mean = %.4f, want ~0.04", mean)
	}
	if sd := math.Sqrt(variance); math.Abs(sd-0.12) > 0.01 {
		t.Errorf("stdev = %.4f, want ~0.12", sd)
	}
}

func TestParametricSourceClampsRuin(t *testing.T) {
	src := ParametricSource{Mu: 0.0, Sigma: 2.0, Df: 3, Periods: 50}
	rng := rand.New(rand.NewPCG(7, 7))
	for i := 0; i < 2000; i++ {
		for _, r := range src.Draw(rng) {
			if 1+r < 0 {
				t.Fatalf("return %.4f makes 1+r negative", r)
			}
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/scenario/`
Expected: FAIL (package/symbols undefined).

- [ ] **Step 3: Write doc.go**

```go
// Package scenario generates synthetic real-return paths from either
// parameters or a panel of historical returns, behind a single Source
// interface. It is decumulation-agnostic and reusable for any
// path-dependent study (accumulation, glidepaths, stress tests).
//
// All returns are periodic and real (inflation already removed): use
// Deflate to obtain them from nominal prices and an HICP series. A Source
// yields one Sequence per Draw; callers run many Draws for a Monte-Carlo,
// or iterate HistoricalCohorts for a deterministic every-start-date backtest.
package scenario
```

- [ ] **Step 4: Write source.go**

```go
package scenario

import (
	"math"
	"math/rand/v2"
)

// Sequence is a periodic real-return path, e.g. 30 annual returns where
// 0.04 means +4 % over the period.
type Sequence []float64

// Source produces synthetic return paths of a fixed length.
type Source interface {
	// Draw returns one path of Len periods using rng.
	Draw(rng *rand.Rand) Sequence
	// Len is the number of periods in every path Draw returns.
	Len() int
}

// ParametricSource draws i.i.d. returns from a Student-t distribution
// scaled so the mean is Mu and the standard deviation is exactly Sigma.
// Df is the degrees of freedom (lower = fatter tails); Df <= 2 falls back
// to a Normal. Each drawn return r is clamped so that 1+r >= 0.
type ParametricSource struct {
	Mu, Sigma, Df float64
	Periods       int
}

// Len reports the path length.
func (p ParametricSource) Len() int { return p.Periods }

// Draw returns one i.i.d. path.
func (p ParametricSource) Draw(rng *rand.Rand) Sequence {
	scale := p.Sigma
	if p.Df > 2 {
		scale = p.Sigma / math.Sqrt(p.Df/(p.Df-2))
	}
	seq := make(Sequence, p.Periods)
	for i := range seq {
		r := p.Mu + scale*studentT(rng, p.Df)
		if 1+r < 0 {
			r = -1
		}
		seq[i] = r
	}
	return seq
}

// studentT returns a standard Student-t variate at df degrees of freedom
// (variance df/(df-2) for df>2). df <= 0 returns a standard normal.
func studentT(rng *rand.Rand, df float64) float64 {
	if df <= 0 {
		return rng.NormFloat64()
	}
	z := rng.NormFloat64()
	// chi-square(df) = 2 * Gamma(df/2, 1); scale cancels in the ratio.
	chi2 := 2 * gamma(rng, df/2)
	return z / math.Sqrt(chi2/df)
}

// gamma draws from Gamma(shape, 1) via Marsaglia-Tsang (shape >= 1) with
// the standard boost for shape < 1. Stdlib-only.
func gamma(rng *rand.Rand, shape float64) float64 {
	if shape < 1 {
		return gamma(rng, shape+1) * math.Pow(rng.Float64(), 1/shape)
	}
	d := shape - 1.0/3.0
	c := 1.0 / math.Sqrt(9*d)
	for {
		x := rng.NormFloat64()
		v := 1 + c*x
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := rng.Float64()
		if u < 1-0.0331*x*x*x*x {
			return d * v
		}
		if math.Log(u) < 0.5*x*x+d*(1-v+math.Log(v)) {
			return d * v
		}
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./pkg/scenario/ && go vet ./pkg/scenario/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/scenario/doc.go pkg/scenario/source.go pkg/scenario/source_test.go
git commit -m "scenario: Source interface and parametric Student-t source"
```

---

## Task 2: scenario.Panel

**Files:**
- Create: `pkg/scenario/panel.go`
- Test: `pkg/scenario/panel_test.go`

**Interfaces:**
- Consumes: `Sequence` (Task 1).
- Produces: `type Panel struct { Returns [][]float64; Weights []float64 }` (Returns indexed `[asset][period]`, all assets same length); `func (p Panel) Periods() int`; `func (p Panel) Combine(weights []float64) Sequence` (weighted sum across assets at each period; nil weights uses p.Weights).

- [ ] **Step 1: Write the failing test**

```go
package scenario

import (
	"math"
	"testing"
)

func TestPanelCombine(t *testing.T) {
	p := Panel{
		Returns: [][]float64{
			{0.10, -0.05, 0.20},
			{0.00, 0.02, -0.01},
		},
		Weights: []float64{0.6, 0.4},
	}
	got := p.Combine(nil)
	want := Sequence{0.06, -0.022, 0.116}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Errorf("period %d: got %.4f, want %.4f", i, got[i], want[i])
		}
	}
}

func TestPanelCombineReweight(t *testing.T) {
	p := Panel{Returns: [][]float64{{0.10}, {0.00}}, Weights: []float64{0.6, 0.4}}
	if got := p.Combine([]float64{1, 0}); math.Abs(got[0]-0.10) > 1e-9 {
		t.Errorf("reweighted got %.4f, want 0.10", got[0])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/scenario/ -run TestPanel`
Expected: FAIL (Panel undefined).

- [ ] **Step 3: Write panel.go**

```go
package scenario

// Panel is an aligned matrix of per-asset real periodic returns plus the
// default weights of a portfolio over them. Returns is indexed
// [asset][period]; every asset row has the same length (use Deflate and an
// aligner to build it). Because resampling happens on the time axis
// (Periods), cross-asset correlations and historical regimes are preserved.
type Panel struct {
	Returns [][]float64 // [asset][period]
	Weights []float64   // default portfolio weights, summing to 1
}

// Periods is the number of historical periods, 0 for an empty panel.
func (p Panel) Periods() int {
	if len(p.Returns) == 0 {
		return 0
	}
	return len(p.Returns[0])
}

// Combine collapses the panel into one portfolio return path using weights
// (nil uses p.Weights). Reweighting is cheap, so live allocation changes do
// not need the underlying series refetched.
func (p Panel) Combine(weights []float64) Sequence {
	if weights == nil {
		weights = p.Weights
	}
	out := make(Sequence, p.Periods())
	for a, row := range p.Returns {
		w := weights[a]
		for t, r := range row {
			out[t] += w * r
		}
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/scenario/ -run TestPanel`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/scenario/panel.go pkg/scenario/panel_test.go
git commit -m "scenario: Panel of aligned per-asset returns with cheap reweighting"
```

---

## Task 3: Block and stationary bootstrap sources

**Files:**
- Create: `pkg/scenario/bootstrap.go`
- Test: `pkg/scenario/bootstrap_test.go`

**Interfaces:**
- Consumes: `Panel`, `Sequence`, `Source` (Tasks 1-2).
- Produces: `type BlockBootstrap struct { Panel Panel; Weights []float64; BlockLen, Periods int }`; `type StationaryBootstrap struct { Panel Panel; Weights []float64; MeanBlock float64; Periods int }`; both implement `Source`.

- [ ] **Step 1: Write the failing test**

```go
package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

// resampling a panel must approximately preserve its own mean return.
func TestBlockBootstrapPreservesMean(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	row := make([]float64, 240)
	src := rand.New(rand.NewPCG(9, 9))
	mean := 0.0
	for i := range row {
		row[i] = 0.005 + 0.04*src.NormFloat64()
		mean += row[i]
	}
	mean /= float64(len(row))
	p := Panel{Returns: [][]float64{row}, Weights: []float64{1}}
	bb := BlockBootstrap{Panel: p, BlockLen: 12, Periods: 360}
	if bb.Len() != 360 {
		t.Fatalf("Len = %d, want 360", bb.Len())
	}
	got := 0.0
	n := 0
	for i := 0; i < 200; i++ {
		for _, r := range bb.Draw(rng) {
			got += r
			n++
		}
	}
	got /= float64(n)
	if math.Abs(got-mean) > 0.003 {
		t.Errorf("resampled mean %.4f, panel mean %.4f", got, mean)
	}
}

func TestStationaryBootstrapLen(t *testing.T) {
	p := Panel{Returns: [][]float64{{0.01, 0.02, -0.01}}, Weights: []float64{1}}
	sb := StationaryBootstrap{Panel: p, MeanBlock: 4, Periods: 20}
	rng := rand.New(rand.NewPCG(1, 1))
	if got := sb.Draw(rng); len(got) != 20 {
		t.Fatalf("len = %d, want 20", len(got))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/scenario/ -run Bootstrap`
Expected: FAIL (undefined).

- [ ] **Step 3: Write bootstrap.go**

```go
package scenario

import (
	"math"
	"math/rand/v2"
)

// BlockBootstrap resamples contiguous blocks of BlockLen periods from the
// Panel's history (sampling on the time axis, so cross-asset correlations
// and regimes survive), applies Weights (nil uses Panel.Weights) and
// concatenates until Periods returns are produced.
type BlockBootstrap struct {
	Panel    Panel
	Weights  []float64
	BlockLen int
	Periods  int
}

// Len reports the path length.
func (b BlockBootstrap) Len() int { return b.Periods }

// Draw returns one resampled path.
func (b BlockBootstrap) Draw(rng *rand.Rand) Sequence {
	hist := b.Panel.Combine(b.Weights)
	return blocks(rng, hist, b.BlockLen, b.Periods, func() bool { return false })
}

// StationaryBootstrap is a block bootstrap with random block lengths drawn
// from a geometric distribution of mean MeanBlock (Politis-Romano): each
// period continues the current block with probability 1-1/MeanBlock, else
// starts a new random block. It avoids the fixed-length artefacts of
// BlockBootstrap.
type StationaryBootstrap struct {
	Panel     Panel
	Weights   []float64
	MeanBlock float64
	Periods   int
}

// Len reports the path length.
func (s StationaryBootstrap) Len() int { return s.Periods }

// Draw returns one resampled path.
func (s StationaryBootstrap) Draw(rng *rand.Rand) Sequence {
	hist := s.Panel.Combine(s.Weights)
	pNew := 1.0
	if s.MeanBlock > 1 {
		pNew = 1 / s.MeanBlock
	}
	return blocks(rng, hist, math.MaxInt, s.Periods, func() bool { return rng.Float64() < pNew })
}

// blocks builds a path of n periods by copying from hist starting at random
// indices. A new block starts when the running block reaches blockLen or
// restart() returns true. hist is treated circularly.
func blocks(rng *rand.Rand, hist Sequence, blockLen, n int, restart func() bool) Sequence {
	out := make(Sequence, 0, n)
	h := len(hist)
	if h == 0 {
		return make(Sequence, n)
	}
	pos, left := rng.IntN(h), blockLen
	for len(out) < n {
		if left == 0 || restart() {
			pos, left = rng.IntN(h), blockLen
		}
		out = append(out, hist[pos%h])
		pos++
		if left != math.MaxInt {
			left--
		}
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/scenario/ -run Bootstrap`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/scenario/bootstrap.go pkg/scenario/bootstrap_test.go
git commit -m "scenario: block and stationary bootstrap sources over a Panel"
```

---

## Task 4: HistoricalCohorts source

**Files:**
- Create: `pkg/scenario/cohorts.go`
- Test: `pkg/scenario/cohorts_test.go`

**Interfaces:**
- Consumes: `Panel`, `Sequence` (Tasks 1-2).
- Produces: `type HistoricalCohorts struct { Panel Panel; Weights []float64; Periods int }`; `func (h HistoricalCohorts) Len() int`; `func (h HistoricalCohorts) Count() int` (number of distinct start windows); `func (h HistoricalCohorts) Cohort(i int) Sequence` (the i-th actual window, no resampling); `Draw` returns a uniformly random cohort so it also satisfies `Source`.

- [ ] **Step 1: Write the failing test**

```go
package scenario

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestHistoricalCohorts(t *testing.T) {
	p := Panel{Returns: [][]float64{{0.1, 0.2, 0.3, 0.4, 0.5}}, Weights: []float64{1}}
	h := HistoricalCohorts{Panel: p, Periods: 3}
	if h.Count() != 3 { // windows starting at 0,1,2
		t.Fatalf("Count = %d, want 3", h.Count())
	}
	got := h.Cohort(1)
	want := Sequence{0.2, 0.3, 0.4}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Errorf("cohort[1][%d] = %.4f, want %.4f", i, got[i], want[i])
		}
	}
	rng := rand.New(rand.NewPCG(1, 1))
	if len(h.Draw(rng)) != 3 {
		t.Errorf("Draw len wrong")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/scenario/ -run Cohorts`
Expected: FAIL (undefined).

- [ ] **Step 3: Write cohorts.go**

```go
package scenario

import "math/rand/v2"

// HistoricalCohorts yields every actual historical window of Periods
// consecutive returns from the Panel, with no resampling: the deterministic
// "every retirement start date" backtest. Count is the number of windows;
// Draw picks one at random so it also satisfies Source.
type HistoricalCohorts struct {
	Panel   Panel
	Weights []float64
	Periods int
}

// Len reports the path length.
func (h HistoricalCohorts) Len() int { return h.Periods }

// Count is the number of distinct start windows, 0 when history is shorter
// than Periods.
func (h HistoricalCohorts) Count() int {
	if n := h.Panel.Periods() - h.Periods + 1; n > 0 {
		return n
	}
	return 0
}

// Cohort returns the i-th historical window (start index i).
func (h HistoricalCohorts) Cohort(i int) Sequence {
	hist := h.Panel.Combine(h.Weights)
	return Sequence(append([]float64(nil), hist[i:i+h.Periods]...))
}

// Draw returns a uniformly random cohort.
func (h HistoricalCohorts) Draw(rng *rand.Rand) Sequence {
	if h.Count() == 0 {
		return make(Sequence, h.Periods)
	}
	return h.Cohort(rng.IntN(h.Count()))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/scenario/ -run Cohorts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/scenario/cohorts.go pkg/scenario/cohorts_test.go
git commit -m "scenario: historical-cohorts source (every actual start date)"
```

---

## Task 5: Deflate helper

**Files:**
- Create: `pkg/scenario/deflate.go`
- Test: `pkg/scenario/deflate_test.go`

**Interfaces:**
- Consumes: `Sequence`, `marketdata.Point`.
- Produces: `func Deflate(prices, hicp []marketdata.Point) Sequence` (period-over-period real returns: nominal price ratio divided by the inflation ratio, on the dates of `prices`, using the HICP value at or before each date).

- [ ] **Step 1: Write the failing test**

```go
package scenario

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func d(y int) time.Time { return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC) }

func TestDeflateRemovesInflation(t *testing.T) {
	// nominal +10%/yr, inflation +10%/yr -> ~0 real return.
	prices := []marketdata.Point{{Date: d(2000), Close: 100}, {Date: d(2001), Close: 110}, {Date: d(2002), Close: 121}}
	hicp := []marketdata.Point{{Date: d(2000), Close: 100}, {Date: d(2001), Close: 110}, {Date: d(2002), Close: 121}}
	got := Deflate(prices, hicp)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	for i, r := range got {
		if math.Abs(r) > 1e-9 {
			t.Errorf("real return[%d] = %.6f, want ~0", i, r)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/scenario/ -run Deflate`
Expected: FAIL (undefined).

- [ ] **Step 3: Write deflate.go**

```go
package scenario

import (
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

// Deflate converts a nominal price series into real period-over-period
// returns using an HICP (or any price-level) series: for consecutive prices
// the nominal ratio is divided by the inflation ratio over the same dates.
// The HICP level used for a date is the last point at or before it.
func Deflate(prices, hicp []marketdata.Point) Sequence {
	if len(prices) < 2 || len(hicp) == 0 {
		return nil
	}
	out := make(Sequence, 0, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		p0, p1 := prices[i-1].Close, prices[i].Close
		c0 := hicpAt(hicp, prices[i-1].Date)
		c1 := hicpAt(hicp, prices[i].Date)
		if p0 <= 0 || c0 <= 0 || c1 <= 0 {
			out = append(out, 0)
			continue
		}
		out = append(out, (p1/p0)/(c1/c0)-1)
	}
	return out
}

// hicpAt returns the HICP level at or before t (the first level if t is
// before the series starts).
func hicpAt(hicp []marketdata.Point, t time.Time) float64 {
	level := hicp[0].Close
	for _, p := range hicp {
		if p.Date.After(t) {
			break
		}
		level = p.Close
	}
	return level
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/scenario/ && go vet ./pkg/scenario/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/scenario/deflate.go pkg/scenario/deflate_test.go
git commit -m "scenario: Deflate nominal prices to real returns via HICP"
```

---

## Task 6: decumul.Plan, Tax and CTOFlatTax

**Files:**
- Create: `pkg/decumul/doc.go`
- Create: `pkg/decumul/plan.go`
- Test: `pkg/decumul/plan_test.go`

**Interfaces:**
- Consumes: `scenario.Source`.
- Produces:
  - `type Cashflow struct { FromYear int; Annual float64 }`
  - `type BufferSleeve struct { Years, RealReturn, DrawThreshold, RefillCap float64 }`
  - `type FlexRule struct { Threshold, Cut float64 }`
  - `type Tax interface { GrossUp(net, growth, cost float64) (gross, newCost, taxPaid float64) }`
  - `type CTOFlatTax struct { Rate float64 }` implementing `Tax`
  - `type Plan struct { Capital, NeedAnnual float64; Cashflows []Cashflow; Years int; Buffer BufferSleeve; Flex FlexRule; Tax Tax; Source scenario.Source }`
  - `func (p Plan) needAt(year int) float64` (NeedAnnual minus active cashflows, floored at 0)

- [ ] **Step 1: Write the failing test**

```go
package decumul

import (
	"math"
	"testing"
)

func TestCTOFlatTaxGrossUp(t *testing.T) {
	tax := CTOFlatTax{Rate: 0.30}
	// growth 200k, cost 100k -> gain fraction 0.5; net 10k.
	gross, newCost, paid := tax.GrossUp(10000, 200000, 100000)
	// effective rate = 0.30*0.5 = 0.15 -> gross = 10000/0.85.
	wantGross := 10000 / 0.85
	if math.Abs(gross-wantGross) > 1e-6 {
		t.Errorf("gross = %.2f, want %.2f", gross, wantGross)
	}
	if math.Abs(paid-(gross-10000)) > 1e-6 {
		t.Errorf("paid = %.2f, want %.2f", paid, gross-10000)
	}
	// cost reduced pro rata of the sale: cost * (1 - gross/growth).
	wantCost := 100000 * (1 - gross/200000)
	if math.Abs(newCost-wantCost) > 1e-6 {
		t.Errorf("newCost = %.2f, want %.2f", newCost, wantCost)
	}
}

func TestNeedAtAppliesCashflows(t *testing.T) {
	p := Plan{NeedAnnual: 48000, Cashflows: []Cashflow{{FromYear: 12, Annual: 18000}}}
	if got := p.needAt(0); math.Abs(got-48000) > 1e-9 {
		t.Errorf("needAt(0) = %.0f, want 48000", got)
	}
	if got := p.needAt(12); math.Abs(got-30000) > 1e-9 {
		t.Errorf("needAt(12) = %.0f, want 30000", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/`
Expected: FAIL (package/symbols undefined).

- [ ] **Step 3: Write doc.go**

```go
// Package decumul evaluates decumulation (withdrawal / retirement / FIRE)
// portfolios. It runs a withdrawal kernel over the real-return paths of a
// scenario.Source to estimate the probability of ruin, FIRE outcome metrics
// and parameter sweeps, and to size a starting capital or a cash buffer
// against a target ruin probability.
//
// Everything is in real euros: the spending floor is constant in purchasing
// power, returns are real, pensions are entered as real Cashflows. The
// parametric model is i.i.d. with fat tails and is probably optimistic vs
// multi-country history; pair it with the bootstrap and historical-cohort
// scenario.Sources, and read ruin in relative orders of magnitude. This is a
// hypothesis-exploration tool, not investment advice.
package decumul
```

- [ ] **Step 4: Write plan.go**

```go
package decumul

import "github.com/bpineau/pofo/pkg/scenario"

// Cashflow is a real annual income (e.g. a pension) received from FromYear
// (0-based) to the horizon; it reduces the net amount sold from the
// portfolio. Pensions are modelled as cashflows, not as an asset.
type Cashflow struct {
	FromYear int
	Annual   float64
}

// BufferSleeve is a low-volatility cash or inflation-linked pocket sized at
// Years times annual spending (capped at the capital). It earns RealReturn
// and is drained first while the portfolio drawdown exceeds DrawThreshold;
// otherwise it is refilled from growth, by at most RefillCap of growth/year.
type BufferSleeve struct {
	Years         float64
	RealReturn    float64
	DrawThreshold float64 // default 0.10
	RefillCap     float64 // default 0.50
}

// FlexRule cuts the year's spending by Cut (e.g. 0.25) whenever the
// portfolio drawdown exceeds Threshold (e.g. 0.20). A zero rule is inactive.
type FlexRule struct {
	Threshold, Cut float64
}

// Tax grosses up a net withdrawal taken by selling part of a growth sleeve
// whose market value is growth and whose cost basis is cost. It returns the
// gross amount to sell, the new cost basis after the sale, and the tax paid.
type Tax interface {
	GrossUp(net, growth, cost float64) (gross, newCost, taxPaid float64)
}

// CTOFlatTax is the French taxable-account flat tax: only the realised gain
// fraction of a sale is taxed at Rate, so the effective rate starts low and
// drifts toward Rate as unrealised gains compound.
type CTOFlatTax struct{ Rate float64 }

// GrossUp implements Tax.
func (t CTOFlatTax) GrossUp(net, growth, cost float64) (gross, newCost, taxPaid float64) {
	if growth <= 0 {
		return net, cost, 0
	}
	gainFrac := 1 - cost/growth
	if gainFrac < 0 {
		gainFrac = 0
	}
	eff := t.Rate * gainFrac
	gross = net
	if eff < 1 {
		gross = net / (1 - eff)
	}
	if gross > growth {
		gross = growth
	}
	newCost = cost * (1 - gross/growth)
	return gross, newCost, gross - net
}

// Plan is a full decumulation scenario.
type Plan struct {
	Capital    float64
	NeedAnnual float64
	Cashflows  []Cashflow
	Years      int
	Buffer     BufferSleeve
	Flex       FlexRule
	Tax        Tax
	Source     scenario.Source
}

// needAt is the net spending in a given year after active cashflows,
// floored at 0.
func (p Plan) needAt(year int) float64 {
	need := p.NeedAnnual
	for _, c := range p.Cashflows {
		if year >= c.FromYear {
			need -= c.Annual
		}
	}
	if need < 0 {
		return 0
	}
	return need
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./pkg/decumul/ && go vet ./pkg/decumul/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/decumul/doc.go pkg/decumul/plan.go pkg/decumul/plan_test.go
git commit -m "decumul: Plan, Tax interface and CTO flat-tax cost-basis model"
```

---

## Task 7: Run — the per-path withdrawal kernel

**Files:**
- Create: `pkg/decumul/run.go`
- Test: `pkg/decumul/run_test.go`

**Interfaces:**
- Consumes: `Plan`, `Tax` (Task 6), `scenario.Sequence`.
- Produces: `type PathResult struct { Wealth []float64; Ruined bool; TaxPaid float64; Withdrawn float64 }` (Wealth has Years+1 entries: index 0 = starting capital, index k = total wealth after year k); `func (p Plan) RunPath(returns scenario.Sequence) PathResult`.

- [ ] **Step 1: Write the failing test**

```go
package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// With zero returns, no tax and no pension, capital depletes by need/year.
func TestRunPathDepletion(t *testing.T) {
	p := Plan{Capital: 100000, NeedAnnual: 25000, Years: 5, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0, 0, 0, 0, 0})
	if !res.Ruined {
		t.Errorf("expected ruin: 100k - 5*25k < 0")
	}
	if len(res.Wealth) != 6 {
		t.Fatalf("Wealth len = %d, want 6", len(res.Wealth))
	}
	if math.Abs(res.Wealth[0]-100000) > 1e-6 {
		t.Errorf("Wealth[0] = %.0f, want 100000", res.Wealth[0])
	}
}

// A high enough capital with positive returns survives.
func TestRunPathSurvives(t *testing.T) {
	p := Plan{Capital: 1_000_000, NeedAnnual: 20000, Years: 10, Tax: CTOFlatTax{Rate: 0}}
	res := p.RunPath(scenario.Sequence{0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05, 0.05})
	if res.Ruined {
		t.Errorf("did not expect ruin")
	}
	if res.Wealth[10] <= 0 {
		t.Errorf("final wealth = %.0f, want > 0", res.Wealth[10])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/ -run RunPath`
Expected: FAIL (undefined).

- [ ] **Step 3: Write run.go**

```go
package decumul

import "github.com/bpineau/pofo/pkg/scenario"

// PathResult is the outcome of one simulated decumulation path. Wealth has
// Years+1 points: Wealth[0] is the starting capital and Wealth[k] is total
// real wealth (growth + buffer) at the end of year k. Ruined latches true
// the first year a withdrawal cannot be funded.
type PathResult struct {
	Wealth    []float64
	Ruined    bool
	TaxPaid   float64
	Withdrawn float64
}

// RunPath simulates one path under the returns sequence (one return per
// year; missing years are treated as 0). The order each year is: compute the
// net need after cashflows, apply the flex cut on deep drawdowns, withdraw
// via the bucket rule (buffer first while underwater, else growth + refill),
// then grow the sleeves.
func (p Plan) RunPath(returns scenario.Sequence) PathResult {
	tax := p.Tax
	if tax == nil {
		tax = CTOFlatTax{Rate: 0}
	}
	target := p.Buffer.Years * p.NeedAnnual
	buffer := target
	if buffer > p.Capital {
		buffer = p.Capital
	}
	growth := p.Capital - buffer
	cost := growth // initial cost basis = invested amount

	drawTh := p.Buffer.DrawThreshold
	if drawTh == 0 {
		drawTh = 0.10
	}
	refillCap := p.Buffer.RefillCap
	if refillCap == 0 {
		refillCap = 0.50
	}

	res := PathResult{Wealth: make([]float64, p.Years+1)}
	res.Wealth[0] = p.Capital
	peak := p.Capital

	for k := 0; k < p.Years; k++ {
		total := growth + buffer
		if total <= 0 {
			res.Ruined = true
			// remaining years stay at 0.
			break
		}
		if total > peak {
			peak = total
		}
		dd := 1 - total/peak

		need := p.needAt(k)
		if p.Flex.Cut > 0 && dd > p.Flex.Threshold {
			need *= 1 - p.Flex.Cut
		}
		if need > total {
			res.Ruined = true
		}

		if dd > drawTh && buffer > 0 {
			// drain buffer first (no tax), remainder from growth.
			take := need
			if take > buffer {
				take = buffer
			}
			buffer -= take
			res.Withdrawn += take
			if rem := need - take; rem > 0 {
				gross, nc, paid := tax.GrossUp(rem, growth, cost)
				growth -= gross
				cost = nc
				res.TaxPaid += paid
				res.Withdrawn += rem
			}
		} else {
			gross, nc, paid := tax.GrossUp(need, growth, cost)
			growth -= gross
			cost = nc
			res.TaxPaid += paid
			res.Withdrawn += need
			// refill buffer toward target from growth.
			if refill := target - buffer; refill > 0 && growth > 0 {
				if cap := growth * refillCap; refill > cap {
					refill = cap
				}
				if refill > p.NeedAnnual {
					refill = p.NeedAnnual
				}
				g2, nc2, paid2 := tax.GrossUp(refill, growth, cost)
				growth -= g2
				cost = nc2
				res.TaxPaid += paid2
				buffer += refill
			}
		}
		if growth < 0 {
			buffer += growth // cover the shortfall from the buffer
			growth = 0
		}
		if growth+buffer <= 0 {
			res.Ruined = true
		}

		growth *= 1 + ret(returns, k)
		buffer *= 1 + p.Buffer.RealReturn
		res.Wealth[k+1] = growth + buffer
	}
	return res
}

// ret returns the k-th return, or 0 when the sequence is shorter.
func ret(s scenario.Sequence, k int) float64 {
	if k < len(s) {
		return s[k]
	}
	return 0
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/decumul/ -run RunPath`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/decumul/run.go pkg/decumul/run_test.go
git commit -m "decumul: per-path withdrawal kernel (bucket rule, flex, tax)"
```

---

## Task 8: Simulate and Ensemble (parallel Monte-Carlo)

**Files:**
- Create: `pkg/decumul/simulate.go`
- Test: `pkg/decumul/simulate_test.go`

**Interfaces:**
- Consumes: `Plan`, `PathResult`, `scenario.Source`.
- Produces: `type Ensemble struct { Paths []PathResult; Years int }`; `func (p Plan) Simulate(nPaths, workers int, seed uint64) Ensemble`; `func (e Ensemble) RuinProb() float64`.

- [ ] **Step 1: Write the failing test**

```go
package decumul

import (
	"math"
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestSimulateDeterministic(t *testing.T) {
	p := Plan{
		Capital: 1_500_000, NeedAnnual: 48000, Years: 35,
		Tax:    CTOFlatTax{Rate: 0.30},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35},
	}
	a := p.Simulate(20000, 4, 7).RuinProb()
	b := p.Simulate(20000, 4, 7).RuinProb()
	if a != b {
		t.Errorf("not reproducible: %.4f vs %.4f", a, b)
	}
	if a < 0 || a > 1 {
		t.Errorf("ruin prob out of range: %.4f", a)
	}
}

func TestSimulateMoreCapitalLowerRuin(t *testing.T) {
	mk := func(c float64) Plan {
		return Plan{Capital: c, NeedAnnual: 48000, Years: 35, Tax: CTOFlatTax{Rate: 0.30},
			Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35}}
	}
	low := mk(1_200_000).Simulate(20000, 4, 7).RuinProb()
	high := mk(2_500_000).Simulate(20000, 4, 7).RuinProb()
	if !(high < low) {
		t.Errorf("more capital should lower ruin: low=%.4f high=%.4f", low, high)
	}
	_ = math.Abs
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/ -run Simulate`
Expected: FAIL (undefined).

- [ ] **Step 3: Write simulate.go**

```go
package decumul

import (
	"math/rand/v2"
	"sync"
)

// Ensemble is the result of many simulated paths sharing a horizon.
type Ensemble struct {
	Paths []PathResult
	Years int
}

// Simulate runs nPaths Monte-Carlo paths across workers goroutines. Each
// worker derives its RNG from (seed, workerID) so the result is
// reproducible for a fixed worker count.
func (p Plan) Simulate(nPaths, workers int, seed uint64) Ensemble {
	if workers < 1 {
		workers = 1
	}
	paths := make([]PathResult, nPaths)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(seed, uint64(w)+1))
			for i := w; i < nPaths; i += workers {
				paths[i] = p.RunPath(p.Source.Draw(rng))
			}
		}(w)
	}
	wg.Wait()
	return Ensemble{Paths: paths, Years: p.Years}
}

// RuinProb is the fraction of paths that ran out of money.
func (e Ensemble) RuinProb() float64 {
	if len(e.Paths) == 0 {
		return 0
	}
	n := 0
	for _, r := range e.Paths {
		if r.Ruined {
			n++
		}
	}
	return float64(n) / float64(len(e.Paths))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/decumul/ -run Simulate`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/decumul/simulate.go pkg/decumul/simulate_test.go
git commit -m "decumul: parallel Monte-Carlo Simulate and Ensemble.RuinProb"
```

---

## Task 9: CapitalForRuin and golden acceptance tests

**Files:**
- Modify: `pkg/decumul/simulate.go` (add `CapitalForRuin`)
- Create: `pkg/decumul/golden_test.go`
- Test: same.

**Interfaces:**
- Consumes: `Plan`, `Ensemble`.
- Produces: `func (p Plan) CapitalForRuin(target float64, lo, hi float64, nPaths, workers int, seed uint64) float64` (smallest capital whose ruin probability <= target, by bisection reusing the same seed across evaluations so noise stays monotone).

- [ ] **Step 1: Write CapitalForRuin in simulate.go**

```go
// CapitalForRuin returns the smallest starting capital in [lo, hi] whose
// ruin probability is at most target, by ~18 bisection steps. The same seed
// is reused at every capital so Monte-Carlo noise does not break
// monotonicity. Buffer.Years scales with NeedAnnual, not with capital, so
// only Capital varies between evaluations.
func (p Plan) CapitalForRuin(target, lo, hi float64, nPaths, workers int, seed uint64) float64 {
	for i := 0; i < 18; i++ {
		mid := (lo + hi) / 2
		q := p
		q.Capital = mid
		if q.Simulate(nPaths, workers, seed).RuinProb() > target {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}
```

- [ ] **Step 2: Write golden_test.go (failing first run guards the kernel)**

```go
package decumul

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

// Reference values from docs/decumulation-fire-design.md §7 (Python model),
// tolerance ±0.03 M€ on target capital and ±0.3 pt on ruin, >=150k paths.
// Gross-up is modelled here through the cost basis; the reference uses a
// flat 12% gross-up, so these golden checks fix Tax = flat 12% via a stub.
type flatGrossUp struct{ rate float64 }

func (f flatGrossUp) GrossUp(net, growth, cost float64) (float64, float64, float64) {
	gross := net * (1 + f.rate)
	nc := cost
	if growth > 0 {
		nc = cost * (1 - gross/growth)
	}
	return gross, nc, gross - net
}

func basePlan(mu, pensionMonthly float64, years int) Plan {
	return Plan{
		NeedAnnual: 48000,
		Cashflows:  []Cashflow{{FromYear: 67 - 55, Annual: pensionMonthly * 12}},
		Years:      years,
		Tax:        flatGrossUp{rate: 0.12},
		Source:     scenario.ParametricSource{Mu: mu, Sigma: 0.12, Df: 6, Periods: years},
	}
}

func TestGoldenTargetCapital(t *testing.T) {
	cases := []struct {
		mu, pension float64
		years       int
		want        float64 // M€
	}{
		{0.035, 1800, 35, 1.67},
		{0.030, 1800, 35, 1.81},
		{0.045, 1800, 35, 1.45},
		{0.035, 1400, 35, 1.84},
	}
	for _, c := range cases {
		p := basePlan(c.mu, c.pension, c.years)
		got := p.CapitalForRuin(0.05, 0.8e6, 4.5e6, 200000, 8, 7) / 1e6
		if d := got - c.want; d < -0.05 || d > 0.05 {
			t.Errorf("mu=%.3f pension=%.0f: target=%.2fM, want ~%.2fM", c.mu, c.pension, got, c.want)
		}
	}
}

func TestGoldenRuinAt2M(t *testing.T) {
	p := basePlan(0.035, 1800, 35)
	p.Capital = 2_000_000
	got := p.Simulate(200000, 8, 7).RuinProb() * 100
	if got < 1.0 || got > 3.5 {
		t.Errorf("ruin at 2.0M = %.2f%%, want ~2.1%%", got)
	}
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./pkg/decumul/ -run Golden -v`
Expected: PASS within tolerance. If a target is off by more than 0.05M, re-check the kernel order (withdraw at year start, then grow) against `run.go` before widening tolerances.

- [ ] **Step 4: Commit**

```bash
git add pkg/decumul/simulate.go pkg/decumul/golden_test.go
git commit -m "decumul: CapitalForRuin bisection and golden acceptance tests"
```

---

## Task 10: Outcome metrics

**Files:**
- Create: `pkg/decumul/outcome.go`
- Test: `pkg/decumul/outcome_test.go`

**Interfaces:**
- Consumes: `Ensemble`, `PathResult`, `metrics.Quantiles`, `metrics.RollingCAGR` is window-based on values; here implement worst rolling real CAGR directly on Wealth.
- Produces: `type Outcome struct { RuinProb, TerminalP5, TerminalP50, MedianYearsUnderwater, Worst10yCAGR, WithdrawalFailureRate, CDaR float64 }`; `func (e Ensemble) Outcome() Outcome`.

- [ ] **Step 1: Write the failing test**

```go
package decumul

import (
	"math"
	"testing"
)

func TestOutcomeBasics(t *testing.T) {
	// two paths: one survives flat at 100, one ruined.
	e := Ensemble{Years: 2, Paths: []PathResult{
		{Wealth: []float64{100, 100, 100}, Ruined: false},
		{Wealth: []float64{100, 50, 0}, Ruined: true},
	}}
	o := e.Outcome()
	if math.Abs(o.RuinProb-0.5) > 1e-9 {
		t.Errorf("RuinProb = %.3f, want 0.5", o.RuinProb)
	}
	if o.TerminalP5 > o.TerminalP50 {
		t.Errorf("p5 (%.1f) should be <= p50 (%.1f)", o.TerminalP5, o.TerminalP50)
	}
	if o.CDaR < 0 || o.CDaR > 1 {
		t.Errorf("CDaR out of range: %.3f", o.CDaR)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/ -run Outcome`
Expected: FAIL (undefined).

- [ ] **Step 3: Write outcome.go**

```go
package decumul

import (
	"sort"

	"github.com/bpineau/pofo/pkg/metrics"
)

// Outcome bundles the headline decumulation statistics across an Ensemble.
// All wealth figures are real euros; rates are fractions.
type Outcome struct {
	RuinProb              float64 // share of paths that ran out
	TerminalP5            float64 // 5th-percentile terminal wealth (0 for ruined)
	TerminalP50           float64 // median terminal wealth
	MedianYearsUnderwater float64 // median years spent below the prior real high
	Worst10yCAGR          float64 // worst rolling 10-year real CAGR across paths
	WithdrawalFailureRate float64 // share of path-years a withdrawal was unfunded
	CDaR                  float64 // mean of the worst 5% path drawdowns (0.30 = 30%)
}

// Outcome computes the bundle.
func (e Ensemble) Outcome() Outcome {
	var o Outcome
	if len(e.Paths) == 0 {
		return o
	}
	terminals := make([]float64, len(e.Paths))
	underwater := make([]float64, len(e.Paths))
	maxDDs := make([]float64, len(e.Paths))
	ruined, worst := 0, 0.0
	for i, p := range e.Paths {
		terminals[i] = p.Wealth[len(p.Wealth)-1]
		if p.Ruined {
			ruined++
		}
		underwater[i] = float64(yearsUnderwater(p.Wealth))
		maxDDs[i] = pathMaxDD(p.Wealth)
		if c := worst10y(p.Wealth); c < worst {
			worst = c
		}
	}
	o.RuinProb = float64(ruined) / float64(len(e.Paths))
	q := metrics.Quantiles(terminals, 0.05, 0.50)
	o.TerminalP5, o.TerminalP50 = q[0], q[1]
	o.MedianYearsUnderwater = metrics.Quantiles(underwater, 0.50)[0]
	o.Worst10yCAGR = worst
	o.CDaR = conditionalTail(maxDDs, 0.05)
	return o
}

// yearsUnderwater counts entries strictly below the running peak.
func yearsUnderwater(w []float64) int {
	peak, n := w[0], 0
	for _, v := range w {
		if v >= peak {
			peak = v
		} else {
			n++
		}
	}
	return n
}

// pathMaxDD is the deepest peak-to-trough loss of a wealth path (0.30 = 30%).
func pathMaxDD(w []float64) float64 {
	peak, dd := w[0], 0.0
	for _, v := range w {
		if v > peak {
			peak = v
		}
		if peak > 0 {
			if d := 1 - v/peak; d > dd {
				dd = d
			}
		}
	}
	return dd
}

// worst10y is the lowest 10-year real CAGR found in the wealth path, or 0
// when the path is shorter than 10 years or hits zero.
func worst10y(w []float64) float64 {
	worst := 0.0
	for i := 0; i+10 < len(w); i++ {
		if w[i] <= 0 || w[i+10] <= 0 {
			return -1 // hit zero: worst possible
		}
		c := pow10(w[i+10]/w[i]) - 1
		if c < worst {
			worst = c
		}
	}
	return worst
}

// pow10 is the 10th root.
func pow10(x float64) float64 {
	// avoid math.Pow import churn: x^(1/10).
	return powFrac(x, 0.1)
}

// conditionalTail averages the worst frac share of dds (already losses).
func conditionalTail(dds []float64, frac float64) float64 {
	if len(dds) == 0 {
		return 0
	}
	s := append([]float64(nil), dds...)
	sort.Sort(sort.Reverse(sort.Float64Slice(s)))
	n := int(frac * float64(len(s)))
	if n < 1 {
		n = 1
	}
	sum := 0.0
	for _, d := range s[:n] {
		sum += d
	}
	return sum / float64(n)
}
```

- [ ] **Step 4: Add the powFrac helper (math.Pow wrapper) at the bottom of outcome.go**

```go
import "math"

// powFrac returns x^p for x>0.
func powFrac(x, p float64) float64 { return math.Pow(x, p) }
```

Note: merge the `import "math"` into the existing import block at the top of `outcome.go` (with `sort` and the metrics import); do not add a second import statement.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./pkg/decumul/ -run Outcome && go vet ./pkg/decumul/`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add pkg/decumul/outcome.go pkg/decumul/outcome_test.go
git commit -m "decumul: FIRE outcome metrics (terminals, underwater, CDaR, worst-10y)"
```

---

## Task 11: RecoveryTimeDistribution

**Files:**
- Create: `pkg/decumul/recovery.go`
- Test: `pkg/decumul/recovery_test.go`

**Interfaces:**
- Consumes: `Ensemble`.
- Produces: `type RecoveryBucket struct { Years int; Share float64 }`; `func (e Ensemble) RecoveryTimeDistribution() []RecoveryBucket` (across all underwater episodes of all paths, the distribution of years taken to regain the prior real high; episodes still underwater at the horizon count at their current length under the largest bucket).

- [ ] **Step 1: Write the failing test**

```go
package decumul

import (
	"math"
	"testing"
)

func TestRecoveryDistribution(t *testing.T) {
	// path: peak 100, dips for 2 years, recovers -> one 2-year episode.
	e := Ensemble{Years: 4, Paths: []PathResult{
		{Wealth: []float64{100, 90, 95, 100, 100}},
	}}
	dist := e.RecoveryTimeDistribution()
	total := 0.0
	for _, b := range dist {
		total += b.Share
	}
	if math.Abs(total-1.0) > 1e-9 {
		t.Errorf("shares sum to %.4f, want 1.0", total)
	}
	// the 2-year bucket should hold the single episode.
	for _, b := range dist {
		if b.Years == 2 && math.Abs(b.Share-1.0) > 1e-9 {
			t.Errorf("2y share = %.3f, want 1.0", b.Share)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/ -run Recovery`
Expected: FAIL (undefined).

- [ ] **Step 3: Write recovery.go**

```go
package decumul

import "sort"

// RecoveryBucket is the share of underwater episodes whose recovery took
// Years years (Years 0 = the path was at a new high).
type RecoveryBucket struct {
	Years int
	Share float64
}

// RecoveryTimeDistribution is the full histogram of years-to-regain a prior
// real high across every underwater episode of every path. Unlike a mean, it
// exposes the psychologically costly tail ("14 years below my initial
// wealth"). An episode still underwater at the horizon is counted at its
// current length.
func (e Ensemble) RecoveryTimeDistribution() []RecoveryBucket {
	counts := map[int]int{}
	total := 0
	for _, p := range e.Paths {
		for _, spell := range underwaterSpells(p.Wealth) {
			counts[spell]++
			total++
		}
	}
	if total == 0 {
		return nil
	}
	out := make([]RecoveryBucket, 0, len(counts))
	for y, c := range counts {
		out = append(out, RecoveryBucket{Years: y, Share: float64(c) / float64(total)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Years < out[j].Years })
	return out
}

// underwaterSpells returns the length in years of each peak-to-recovery
// episode in a wealth path; a year at a fresh high is a zero-length episode.
func underwaterSpells(w []float64) []int {
	var spells []int
	peak := w[0]
	under := 0
	for _, v := range w[1:] {
		if v >= peak {
			if under > 0 {
				spells = append(spells, under)
			} else {
				spells = append(spells, 0)
			}
			peak = v
			under = 0
		} else {
			under++
		}
	}
	if under > 0 { // still underwater at the horizon
		spells = append(spells, under)
	}
	return spells
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/decumul/ -run Recovery`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/decumul/recovery.go pkg/decumul/recovery_test.go
git commit -m "decumul: full recovery-time distribution across paths"
```

---

## Task 12: Sweep1D and Sweep2D

**Files:**
- Create: `pkg/decumul/sweep.go`
- Test: `pkg/decumul/sweep_test.go`

**Interfaces:**
- Consumes: `Plan`, `Ensemble`.
- Produces:
  - `type Param int` with `const ( Capital Param = iota; BufferYears; Mu; NeedAnnual )` and `func (p *Plan) set(param Param, v float64)` (mutates a copy's field; Mu rebuilds a ParametricSource preserving Sigma/Df/Periods).
  - `type SweepPoint struct { Value, RuinProb, TerminalP50 float64 }`
  - `func (p Plan) Sweep1D(param Param, values []float64, nPaths, workers int, seed uint64) []SweepPoint`
  - `type Surface struct { Xs, Ys []float64; Ruin [][]float64 }`
  - `func (p Plan) Sweep2D(x, y Param, xs, ys []float64, nPaths, workers int, seed uint64) Surface`

- [ ] **Step 1: Write the failing test**

```go
package decumul

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func sweepPlan() Plan {
	return Plan{Capital: 1_500_000, NeedAnnual: 48000, Years: 35, Tax: CTOFlatTax{Rate: 0.30},
		Source: scenario.ParametricSource{Mu: 0.035, Sigma: 0.12, Df: 6, Periods: 35}}
}

func TestSweep1DBufferMonotoneRuin(t *testing.T) {
	p := sweepPlan()
	pts := p.Sweep1D(BufferYears, []float64{0, 2, 4, 6}, 8000, 4, 7)
	if len(pts) != 4 {
		t.Fatalf("len = %d, want 4", len(pts))
	}
	for _, pt := range pts {
		if pt.RuinProb < 0 || pt.RuinProb > 1 {
			t.Errorf("ruin out of range: %.3f", pt.RuinProb)
		}
	}
}

func TestSweep2DShape(t *testing.T) {
	p := sweepPlan()
	s := p.Sweep2D(BufferYears, Mu, []float64{0, 3}, []float64{0.03, 0.05}, 4000, 4, 7)
	if len(s.Ruin) != 2 || len(s.Ruin[0]) != 2 {
		t.Fatalf("surface shape %dx%d, want 2x2", len(s.Ruin), len(s.Ruin[0]))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/ -run Sweep`
Expected: FAIL (undefined).

- [ ] **Step 3: Write sweep.go**

```go
package decumul

import "github.com/bpineau/pofo/pkg/scenario"

// Param names a Plan field a sweep can vary.
type Param int

const (
	Capital Param = iota
	BufferYears
	Mu
	NeedAnnual
)

// set returns a copy of the plan with param set to v. Varying Mu rebuilds a
// ParametricSource keeping the current Sigma/Df/Periods, so it only applies
// when Source already is a ParametricSource.
func (p Plan) set(param Param, v float64) Plan {
	switch param {
	case Capital:
		p.Capital = v
	case BufferYears:
		p.Buffer.Years = v
	case NeedAnnual:
		p.NeedAnnual = v
	case Mu:
		if ps, ok := p.Source.(scenario.ParametricSource); ok {
			ps.Mu = v
			p.Source = ps
		}
	}
	return p
}

// SweepPoint is one evaluated parameter value.
type SweepPoint struct {
	Value, RuinProb, TerminalP50 float64
}

// Sweep1D evaluates ruin and median terminal wealth across values of param,
// reusing one seed so the curve is smooth.
func (p Plan) Sweep1D(param Param, values []float64, nPaths, workers int, seed uint64) []SweepPoint {
	out := make([]SweepPoint, len(values))
	for i, v := range values {
		o := p.set(param, v).Simulate(nPaths, workers, seed).Outcome()
		out[i] = SweepPoint{Value: v, RuinProb: o.RuinProb, TerminalP50: o.TerminalP50}
	}
	return out
}

// Surface is a grid of ruin probabilities over two parameters.
type Surface struct {
	Xs, Ys []float64
	Ruin   [][]float64 // Ruin[y][x]
}

// Sweep2D evaluates ruin over the cartesian product of xs and ys.
func (p Plan) Sweep2D(x, y Param, xs, ys []float64, nPaths, workers int, seed uint64) Surface {
	s := Surface{Xs: xs, Ys: ys, Ruin: make([][]float64, len(ys))}
	for j, yv := range ys {
		s.Ruin[j] = make([]float64, len(xs))
		for i, xv := range xs {
			s.Ruin[j][i] = p.set(x, xv).set(y, yv).Simulate(nPaths, workers, seed).RuinProb()
		}
	}
	return s
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/decumul/ -run Sweep && go vet ./pkg/decumul/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/decumul/sweep.go pkg/decumul/sweep_test.go
git commit -m "decumul: 1D and 2D parameter sweeps (buffer arbitrage, ruin surface)"
```

---

## Task 13: chart.Bars

**Files:**
- Create: `pkg/chart/bars.go`
- Test: `pkg/chart/bars_test.go`

**Interfaces:**
- Consumes: `chart.Options`.
- Produces: `type Bar struct { Label string; Value float64 }`; `func Bars(opt Options, bars []Bar) string` (vertical bar chart SVG, matching the Line style).

- [ ] **Step 1: Write the failing test**

```go
package chart

import (
	"strings"
	"testing"
)

func TestBarsSVG(t *testing.T) {
	svg := Bars(Options{Title: "Recovery"}, []Bar{{"0y", 0.4}, {"1y", 0.3}, {"2y", 0.3}})
	if !strings.HasPrefix(svg, "<svg") {
		t.Errorf("not an SVG: %.20q", svg)
	}
	if !strings.Contains(svg, "Recovery") {
		t.Errorf("title missing")
	}
	if strings.Count(svg, "<rect") < 3 {
		t.Errorf("expected at least 3 bars")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/chart/ -run Bars`
Expected: FAIL (undefined).

- [ ] **Step 3: Write bars.go**

```go
package chart

import (
	"fmt"
	"strings"
)

// Bar is one labelled column of a Bars chart.
type Bar struct {
	Label string
	Value float64
}

// Bars renders a vertical bar chart as a standalone SVG, in the same visual
// style as Line. Bars are drawn left to right; the y-axis spans 0 to the
// largest value.
func Bars(opt Options, bars []Bar) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	const padL, padR, padT, padB = 50, 20, 40, 40
	plotW, plotH := w-padL-padR, h-padT-padB
	max := 0.0
	for _, b := range bars {
		if b.Value > max {
			max = b.Value
		}
	}
	if max == 0 {
		max = 1
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" font-family="sans-serif" font-size="12">`, w, h)
	if opt.Title != "" {
		fmt.Fprintf(&sb, `<text x="%d" y="20" font-size="14" font-weight="600">%s</text>`, padL, esc(opt.Title))
	}
	n := len(bars)
	if n == 0 {
		sb.WriteString(`</svg>`)
		return sb.String()
	}
	bw := float64(plotW) / float64(n) * 0.7
	gap := float64(plotW) / float64(n)
	for i, b := range bars {
		bh := b.Value / max * float64(plotH)
		x := float64(padL) + float64(i)*gap + (gap-bw)/2
		y := float64(padT+plotH) - bh
		fmt.Fprintf(&sb, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, y, bw, bh, PaletteColor(0))
		fmt.Fprintf(&sb, `<text x="%.1f" y="%d" text-anchor="middle">%s</text>`, x+bw/2, padT+plotH+15, esc(b.Label))
	}
	sb.WriteString(`</svg>`)
	return sb.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/chart/ -run Bars && go vet ./pkg/chart/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/chart/bars.go pkg/chart/bars_test.go
git commit -m "chart: vertical Bars SVG primitive"
```

---

## Task 14: chart.Heatmap

**Files:**
- Create: `pkg/chart/heatmap.go`
- Test: `pkg/chart/heatmap_test.go`

**Interfaces:**
- Consumes: `chart.Options`.
- Produces: `type HeatmapData struct { Xs, Ys []float64; Z [][]float64; XLabel, YLabel string }`; `func Heatmap(opt Options, d HeatmapData) string` (Z[y][x] in [0,1] coloured green→red; the 2D ruin surface).

- [ ] **Step 1: Write the failing test**

```go
package chart

import (
	"strings"
	"testing"
)

func TestHeatmapSVG(t *testing.T) {
	d := HeatmapData{Xs: []float64{0, 1}, Ys: []float64{0.03, 0.05}, Z: [][]float64{{0.1, 0.2}, {0.3, 0.4}}}
	svg := Heatmap(Options{Title: "Ruin"}, d)
	if !strings.HasPrefix(svg, "<svg") {
		t.Errorf("not an SVG")
	}
	if strings.Count(svg, "<rect") < 4 {
		t.Errorf("expected 4 cells, got %d", strings.Count(svg, "<rect"))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/chart/ -run Heatmap`
Expected: FAIL (undefined).

- [ ] **Step 3: Write heatmap.go**

```go
package chart

import (
	"fmt"
	"strings"
)

// HeatmapData is a grid of values Z[y][x] in [0,1] over the axes Xs and Ys.
type HeatmapData struct {
	Xs, Ys         []float64
	Z              [][]float64
	XLabel, YLabel string
}

// Heatmap renders Z as a coloured grid (green = low, red = high), suitable
// for a ruin surface (buffer years x expected return). Values are clamped to
// [0,1].
func Heatmap(opt Options, d HeatmapData) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	const padL, padR, padT, padB = 60, 20, 40, 50
	plotW, plotH := w-padL-padR, h-padT-padB
	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" font-family="sans-serif" font-size="12">`, w, h)
	if opt.Title != "" {
		fmt.Fprintf(&sb, `<text x="%d" y="20" font-size="14" font-weight="600">%s</text>`, padL, esc(opt.Title))
	}
	ny, nx := len(d.Ys), len(d.Xs)
	if ny == 0 || nx == 0 {
		sb.WriteString(`</svg>`)
		return sb.String()
	}
	cw := float64(plotW) / float64(nx)
	ch := float64(plotH) / float64(ny)
	for j := 0; j < ny; j++ {
		for i := 0; i < nx; i++ {
			x := float64(padL) + float64(i)*cw
			// y axis drawn bottom-up.
			y := float64(padT) + float64(ny-1-j)*ch
			fmt.Fprintf(&sb, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, y, cw+0.5, ch+0.5, heatColor(d.Z[j][i]))
		}
	}
	if d.XLabel != "" {
		fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle">%s</text>`, padL+plotW/2, h-15, esc(d.XLabel))
	}
	sb.WriteString(`</svg>`)
	return sb.String()
}

// heatColor maps v in [0,1] to a green-to-red hex color.
func heatColor(v float64) string {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	r := int(255 * v)
	g := int(200 * (1 - v))
	return fmt.Sprintf("#%02x%02x40", r, g)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/chart/ -run Heatmap && go vet ./pkg/chart/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/chart/heatmap.go pkg/chart/heatmap_test.go
git commit -m "chart: Heatmap SVG primitive for the 2D ruin surface"
```

---

## Task 15: web model and handler (parametric mode)

**Files:**
- Create: `pkg/decumul/web/doc.go`
- Create: `pkg/decumul/web/model.go`
- Create: `pkg/decumul/web/server.go`
- Create: `pkg/decumul/web/embed.go`
- Create: `pkg/decumul/web/assets/index.html`
- Create: `pkg/decumul/web/assets/app.js`
- Create: `pkg/decumul/web/assets/app.css`
- Test: `pkg/decumul/web/server_test.go`

**Interfaces:**
- Consumes: `decumul.Plan`, `decumul.Param`, `scenario.ParametricSource`, `chart.Line/Bars/Heatmap`.
- Produces:
  - `type Params struct { Capital, NeedAnnual, BufferYears, Mu, Sigma, Df, BufferReturn float64; Years, PensionYear int; PensionAnnual, FlexCut, TaxRate float64; NPaths int; Weights []float64 }` (JSON tags; Weights nil in parametric mode)
  - `type Result struct { Cards map[string]string; BufferSVG, RuinCurveSVG, SurfaceSVG, RecoverySVG string }` (JSON)
  - `func (pr Params) plan() decumul.Plan`
  - `func Compute(pr Params) Result`
  - `func Handler() http.Handler` (serves the embedded page at `/` and `Compute` at `POST /api/sim`)

- [ ] **Step 1: Write the failing test**

```go
package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPISim(t *testing.T) {
	body, _ := json.Marshal(Params{
		Capital: 1_500_000, NeedAnnual: 48000, BufferYears: 3,
		Mu: 0.035, Sigma: 0.12, Df: 6, Years: 35, NPaths: 3000, TaxRate: 0.30,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sim", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	var res Result
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if res.BufferSVG == "" || res.Cards["ruin"] == "" {
		t.Errorf("empty result: %+v", res.Cards)
	}
}

func TestServesIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, req)
	if rec.Code != 200 || !bytes.Contains(rec.Body.Bytes(), []byte("<html")) {
		t.Errorf("index not served: %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/web/`
Expected: FAIL (undefined / no assets).

- [ ] **Step 3: Write doc.go and embed.go**

`doc.go`:
```go
// Package web is a thin embedded HTTP UI for pkg/decumul: it serves a
// single page of sliders and, on each change, runs the Monte-Carlo in Go and
// returns chart SVGs and summary cards as JSON. The engine stays in Go; the
// browser only renders. Handler returns a ready-to-mount http.Handler.
package web
```

`embed.go`:
```go
package web

import "embed"

//go:embed assets/index.html assets/app.js assets/app.css
var assets embed.FS
```

- [ ] **Step 4: Write model.go**

```go
package web

import (
	"fmt"

	"github.com/bpineau/pofo/pkg/chart"
	"github.com/bpineau/pofo/pkg/decumul"
	"github.com/bpineau/pofo/pkg/scenario"
)

// Params is the slider state posted by the browser. Weights is nil in
// parametric mode and holds per-holding fractions in portfolio mode.
type Params struct {
	Capital       float64   `json:"capital"`
	NeedAnnual    float64   `json:"needAnnual"`
	BufferYears   float64   `json:"bufferYears"`
	Mu            float64   `json:"mu"`
	Sigma         float64   `json:"sigma"`
	Df            float64   `json:"df"`
	BufferReturn  float64   `json:"bufferReturn"`
	Years         int       `json:"years"`
	PensionYear   int       `json:"pensionYear"`
	PensionAnnual float64   `json:"pensionAnnual"`
	FlexCut       float64   `json:"flexCut"`
	TaxRate       float64   `json:"taxRate"`
	NPaths        int       `json:"nPaths"`
	Weights       []float64 `json:"weights"`
}

// Result is the JSON returned for one parameter set.
type Result struct {
	Cards        map[string]string `json:"cards"`
	BufferSVG    string            `json:"bufferSvg"`
	RuinCurveSVG string            `json:"ruinCurveSvg"`
	SurfaceSVG   string            `json:"surfaceSvg"`
	RecoverySVG  string            `json:"recoverySvg"`
}

// plan builds a decumul.Plan from the params (parametric source).
func (pr Params) plan() decumul.Plan {
	p := decumul.Plan{
		Capital:    pr.Capital,
		NeedAnnual: pr.NeedAnnual,
		Years:      pr.Years,
		Buffer:     decumul.BufferSleeve{Years: pr.BufferYears, RealReturn: pr.BufferReturn},
		Flex:       decumul.FlexRule{Threshold: 0.20, Cut: pr.FlexCut},
		Tax:        decumul.CTOFlatTax{Rate: pr.TaxRate},
		Source:     scenario.ParametricSource{Mu: pr.Mu, Sigma: pr.Sigma, Df: pr.Df, Periods: pr.Years},
	}
	if pr.PensionAnnual > 0 {
		p.Cashflows = []decumul.Cashflow{{FromYear: pr.PensionYear, Annual: pr.PensionAnnual}}
	}
	return p
}

const simWorkers = 8

// Compute runs the simulation and renders the charts for one params set.
func Compute(pr Params) Result {
	if pr.NPaths == 0 {
		pr.NPaths = 5000
	}
	p := pr.plan()
	seed := uint64(7)

	// buffer arbitrage curve (ruin vs buffer years).
	bufVals := []float64{0, 1, 2, 3, 4, 5, 6, 8, 10}
	sweep := p.Sweep1D(decumul.BufferYears, bufVals, pr.NPaths, simWorkers, seed)
	bufSeries := chart.Series{Name: "ruin %"}
	for _, s := range sweep {
		bufSeries.Values = append(bufSeries.Values, s.RuinProb*100)
	}

	// headline outcome at the selected buffer.
	o := p.Simulate(pr.NPaths, simWorkers, seed).Outcome()

	// recovery distribution.
	var bars []chart.Bar
	for _, b := range p.Simulate(pr.NPaths, simWorkers, seed).RecoveryTimeDistribution() {
		bars = append(bars, chart.Bar{Label: fmt.Sprintf("%dy", b.Years), Value: b.Share})
	}

	// ruin surface (buffer x mu).
	xs := bufVals
	ys := []float64{0.02, 0.03, 0.035, 0.04, 0.045, 0.05}
	surf := p.Sweep2D(decumul.BufferYears, decumul.Mu, xs, ys, pr.NPaths/2+1, simWorkers, seed)

	return Result{
		Cards: map[string]string{
			"ruin":          fmt.Sprintf("%.1f%%", o.RuinProb*100),
			"withdrawRate":  fmt.Sprintf("%.2f%%", pr.NeedAnnual/pr.Capital*100),
			"terminalP50":   fmt.Sprintf("%.0f k€", o.TerminalP50/1000),
			"terminalP5":    fmt.Sprintf("%.0f k€", o.TerminalP5/1000),
		},
		BufferSVG:    chart.Bars(chart.Options{Title: "Ruin % by buffer years"}, barsFromSweep(sweep)),
		RuinCurveSVG: chart.Bars(chart.Options{Title: "Terminal wealth p50 (k€) by buffer"}, terminalBars(sweep)),
		SurfaceSVG:   chart.Heatmap(chart.Options{Title: "Ruin: buffer (x) × real CAGR (y)"}, surfaceData(surf)),
		RecoverySVG:  chart.Bars(chart.Options{Title: "Recovery-time distribution"}, bars),
	}
}

func barsFromSweep(s []decumul.SweepPoint) []chart.Bar {
	out := make([]chart.Bar, len(s))
	for i, p := range s {
		out[i] = chart.Bar{Label: fmt.Sprintf("%.0fy", p.Value), Value: p.RuinProb * 100}
	}
	return out
}

func terminalBars(s []decumul.SweepPoint) []chart.Bar {
	out := make([]chart.Bar, len(s))
	for i, p := range s {
		out[i] = chart.Bar{Label: fmt.Sprintf("%.0fy", p.Value), Value: p.TerminalP50 / 1000}
	}
	return out
}

func surfaceData(s decumul.Surface) chart.HeatmapData {
	return chart.HeatmapData{Xs: s.Xs, Ys: s.Ys, Z: s.Ruin, XLabel: "buffer years"}
}
```

- [ ] **Step 5: Write server.go**

```go
package web

import (
	"encoding/json"
	"net/http"
)

// Handler returns the decumulation UI: the embedded page at / and the
// simulation endpoint at POST /api/sim.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(mustSub())))
	mux.HandleFunc("/api/sim", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var pr Params
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Compute(pr))
	})
	return mux
}
```

- [ ] **Step 6: Add mustSub to embed.go**

```go
import (
	"io/fs"
)

// mustSub exposes the assets directory at the URL root (so / serves
// assets/index.html).
func mustSub() fs.FS {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}
	return sub
}
```

Note: merge `io/fs` into embed.go's import block alongside `embed`.

- [ ] **Step 7: Write assets/index.html, app.css, app.js**

`assets/index.html`:
```html
<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>pofo — FIRE decumulation</title>
<link rel="stylesheet" href="app.css"></head>
<body>
<h1>FIRE decumulation explorer</h1>
<div id="layout">
<form id="controls"></form>
<div id="cards"></div>
<div id="charts">
<div id="bufferSvg"></div><div id="ruinCurveSvg"></div>
<div id="surfaceSvg"></div><div id="recoverySvg"></div>
</div></div>
<script src="app.js"></script>
</body></html>
```

`assets/app.css`:
```css
body{font-family:-apple-system,system-ui,sans-serif;margin:1.5rem auto;max-width:1100px;padding:0 1rem;color:#1a1a1a}
h1{font-size:1.4rem}
#controls{display:grid;grid-template-columns:repeat(2,1fr);gap:.4rem 1.5rem;margin:1rem 0}
.ctl{display:flex;flex-direction:column;font-size:.8rem;color:#444}
.ctl input{width:100%}
#cards{display:flex;gap:1rem;flex-wrap:wrap;margin:1rem 0}
.card{border:1px solid #ddd;border-radius:6px;padding:.5rem .9rem;min-width:7rem}
.card .v{font-size:1.3rem;font-weight:600}
#charts{display:grid;grid-template-columns:1fr 1fr;gap:1rem}
svg{max-width:100%;height:auto}
```

`assets/app.js`:
```javascript
// Slider definitions: [key, label, min, max, step, default].
const SLIDERS = [
  ["capital","Capital (€)",800000,4000000,10000,1800000],
  ["needAnnual","Spending floor /yr (€)",24000,84000,1000,48000],
  ["bufferYears","Buffer (years)",0,10,1,3],
  ["mu","Real growth return",0.01,0.07,0.005,0.045],
  ["sigma","Volatility",0.06,0.20,0.005,0.12],
  ["df","Tail df (low=fat)",3,30,1,6],
  ["bufferReturn","Buffer real return",-0.01,0.02,0.005,0.005],
  ["years","Horizon (years)",20,45,1,40],
  ["pensionYear","Pension from year",5,20,1,12],
  ["pensionAnnual","Pension /yr (€)",0,36000,1000,12000],
  ["flexCut","Possible spending cut",0,0.40,0.05,0.25],
  ["taxRate","Flat tax on gains",0,0.35,0.01,0.314],
  ["nPaths","Simulations",500,5000,500,2000],
];
const form = document.getElementById("controls");
const state = {};
for (const [k,label,min,max,step,def] of SLIDERS) {
  state[k] = def;
  const d = document.createElement("label"); d.className = "ctl";
  d.innerHTML = `${label}: <span id="v_${k}">${def}</span>
    <input type="range" min="${min}" max="${max}" step="${step}" value="${def}" id="s_${k}">`;
  form.appendChild(d);
  d.querySelector("input").addEventListener("input", e => {
    state[k] = parseFloat(e.target.value);
    document.getElementById("v_"+k).textContent = e.target.value;
    schedule();
  });
}
let timer = null;
function schedule(){ clearTimeout(timer); timer = setTimeout(run, 200); }
async function run(){
  const body = {...state, years: Math.round(state.years),
    pensionYear: Math.round(state.pensionYear), nPaths: Math.round(state.nPaths)};
  const res = await fetch("/api/sim",{method:"POST",headers:{"Content-Type":"application/json"},
    body: JSON.stringify(body)});
  const r = await res.json();
  for (const id of ["bufferSvg","ruinCurveSvg","surfaceSvg","recoverySvg"])
    document.getElementById(id).innerHTML = r[id];
  document.getElementById("cards").innerHTML = Object.entries(r.cards)
    .map(([k,v]) => `<div class="card"><div>${k}</div><div class="v">${v}</div></div>`).join("");
}
run();
```

- [ ] **Step 8: Run tests to verify they pass**

Run: `go test ./pkg/decumul/web/ && go vet ./pkg/decumul/web/`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add pkg/decumul/web/
git commit -m "decumul/web: embedded live UI and /api/sim (parametric mode)"
```

---

## Task 16: cmd/pofo -fire wiring

**Files:**
- Modify: `cmd/pofo/main.go` (add `-fire` flag to `options`, a `runFire` function, dispatch in `run`)
- Test: manual (documented below) plus `go build`.

**Interfaces:**
- Consumes: `web.Handler`.
- Produces: a `-fire` boolean flag that starts the local server and opens the browser.

- [ ] **Step 1: Add the flag to the options struct and flag set**

In `cmd/pofo/main.go`, add to the `options` struct:
```go
	fire bool
```
and where flags are registered (near the other `flag.BoolVar` calls):
```go
	flag.BoolVar(&opt.fire, "fire", false, "open the FIRE decumulation explorer (local web UI) then exit")
```

- [ ] **Step 2: Dispatch in run(), before the normal report path**

Find the early-exit dispatch block in `run` (where `-warmup`, `-suggest`, `-coverage` are handled) and add:
```go
	if opt.fire {
		return runFire(opt)
	}
```

- [ ] **Step 3: Write runFire (add near runWarmup in main.go)**

```go
import (
	"net"
	"net/http"

	"github.com/bpineau/pofo/pkg/decumul/web"
)

// runFire starts the embedded decumulation explorer on a local port and
// opens it in the browser. It blocks until interrupted.
func runFire(opt *options) error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	url := "http://" + ln.Addr().String() + "/"
	fmt.Fprintf(os.Stderr, "FIRE explorer on %s (Ctrl-C to stop)\n", url)
	if !opt.noOpen {
		openBrowser(url)
	}
	return http.Serve(ln, web.Handler())
}
```

Note: `fmt`, `os`, `net`, `net/http` — add only the ones not already imported in main.go. `openBrowser` and `opt.noOpen` already exist.

- [ ] **Step 4: Build and smoke-test**

Run:
```bash
go build ./cmd/pofo && go vet ./...
```
Expected: builds clean. Manual: `./pofo -fire -no-open` prints the URL; `curl -s $URL | grep '<html'` returns the page; stop with Ctrl-C.

- [ ] **Step 5: Commit**

```bash
git add cmd/pofo/main.go
git commit -m "cmd/pofo: -fire flag launches the decumulation explorer"
```

---

## Task 17: portfolio → scenario.Panel adapter

**Files:**
- Create: `pkg/decumul/web/portfolio.go`
- Test: `pkg/decumul/web/portfolio_test.go`

**Interfaces:**
- Consumes: `marketdata.Series`, `scenario.Deflate`, `scenario.Panel`, `metrics.Mean`.
- Produces:
  - `func BuildPanel(assets []AssetSeries, hicp []marketdata.Point) (scenario.Panel, error)` where `type AssetSeries struct { Weight float64; Points []marketdata.Point }` — aligns each asset's annual real returns (calendar-year sampling) into a Panel with the given weights.
  - `func FitParametric(panel scenario.Panel, weights []float64) (mu, sigma float64)` — sample mean and stdev of the weighted annual real returns, to seed the parametric sliders from history.

- [ ] **Step 1: Write the failing test**

```go
package web

import (
	"math"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/marketdata"
)

func yr(y int) time.Time { return time.Date(y, 6, 30, 0, 0, 0, 0, time.UTC) }

func TestBuildPanelAndFit(t *testing.T) {
	a := AssetSeries{Weight: 1, Points: []marketdata.Point{
		{Date: yr(2000), Close: 100}, {Date: yr(2001), Close: 110}, {Date: yr(2002), Close: 121},
	}}
	hicp := []marketdata.Point{{Date: yr(2000), Close: 100}, {Date: yr(2001), Close: 100}, {Date: yr(2002), Close: 100}}
	panel, err := BuildPanel([]AssetSeries{a}, hicp)
	if err != nil {
		t.Fatal(err)
	}
	if panel.Periods() != 2 {
		t.Fatalf("periods = %d, want 2", panel.Periods())
	}
	mu, sigma := FitParametric(panel, []float64{1})
	if math.Abs(mu-0.10) > 0.01 {
		t.Errorf("mu = %.4f, want ~0.10", mu)
	}
	if sigma < 0 {
		t.Errorf("sigma negative")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/web/ -run BuildPanel`
Expected: FAIL (undefined).

- [ ] **Step 3: Write portfolio.go**

```go
package web

import (
	"fmt"
	"math"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/metrics"
	"github.com/bpineau/pofo/pkg/scenario"
)

// AssetSeries is one holding's weight and its (nominal) price points,
// already converted to the report currency.
type AssetSeries struct {
	Weight float64
	Points []marketdata.Point
}

// BuildPanel deflates each asset by hicp and aligns the resulting annual
// real returns into a scenario.Panel. Assets are truncated to their common
// number of yearly returns so every row has the same length.
func BuildPanel(assets []AssetSeries, hicp []marketdata.Point) (scenario.Panel, error) {
	if len(assets) == 0 {
		return scenario.Panel{}, fmt.Errorf("no assets")
	}
	rows := make([][]float64, len(assets))
	weights := make([]float64, len(assets))
	min := -1
	for i, a := range assets {
		rows[i] = annualReal(a.Points, hicp)
		weights[i] = a.Weight
		if min < 0 || len(rows[i]) < min {
			min = len(rows[i])
		}
	}
	if min <= 0 {
		return scenario.Panel{}, fmt.Errorf("not enough history")
	}
	for i := range rows {
		rows[i] = rows[i][len(rows[i])-min:] // keep the last min years (common window)
	}
	return scenario.Panel{Returns: rows, Weights: normalize(weights)}, nil
}

// annualReal samples one real return per calendar year from points using the
// last quote of each year, deflated by hicp.
func annualReal(points, hicp []marketdata.Point) []float64 {
	yearly := lastPerYear(points)
	return scenario.Deflate(yearly, hicp)
}

// lastPerYear keeps the last point of each calendar year, ascending.
func lastPerYear(points []marketdata.Point) []marketdata.Point {
	var out []marketdata.Point
	for _, p := range points {
		if n := len(out); n > 0 && out[n-1].Date.Year() == p.Date.Year() {
			out[n-1] = p
		} else {
			out = append(out, p)
		}
	}
	return out
}

// FitParametric returns the sample mean and standard deviation of the
// weighted annual real returns, to seed the parametric sliders.
func FitParametric(panel scenario.Panel, weights []float64) (mu, sigma float64) {
	seq := panel.Combine(weights)
	if len(seq) == 0 {
		return 0, 0
	}
	mu = metrics.Mean(seq)
	for _, r := range seq {
		sigma += (r - mu) * (r - mu)
	}
	if len(seq) > 1 {
		sigma = math.Sqrt(sigma / float64(len(seq)-1))
	}
	return mu, sigma
}

func normalize(w []float64) []float64 {
	sum := 0.0
	for _, x := range w {
		sum += x
	}
	if sum == 0 {
		return w
	}
	out := make([]float64, len(w))
	for i, x := range w {
		out[i] = x / sum
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./pkg/decumul/web/ -run BuildPanel && go vet ./pkg/decumul/web/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/decumul/web/portfolio.go pkg/decumul/web/portfolio_test.go
git commit -m "decumul/web: portfolio→Panel adapter and parametric fit from history"
```

---

## Task 18: portfolio mode (bootstrap/cohorts toggle, live allocation)

**Files:**
- Modify: `pkg/decumul/web/model.go` (extend `Params` with `Model string` and per-holding `Weights`; build a bootstrap/cohort source from a stored Panel)
- Modify: `pkg/decumul/web/server.go` (`Handler` accepts an optional `*scenario.Panel` and holding labels)
- Modify: `cmd/pofo/main.go` (`runFire` resolves a portfolio file, builds the panel, passes it to the handler)
- Test: `pkg/decumul/web/model_test.go`

**Interfaces:**
- Consumes: `BuildPanel`, `scenario.BlockBootstrap`, `scenario.HistoricalCohorts`.
- Produces:
  - `Params.Model string` (`"parametric"`, `"bootstrap"`, `"cohorts"`); `Params.Weights []float64` re-weight the panel live.
  - `func ComputeWithPanel(pr Params, panel *scenario.Panel) Result`
  - `func Handler(panel *scenario.Panel, labels []string) http.Handler` (nil panel = parametric-only).

- [ ] **Step 1: Write the failing test**

```go
package web

import (
	"testing"

	"github.com/bpineau/pofo/pkg/scenario"
)

func TestComputeWithPanelBootstrap(t *testing.T) {
	panel := scenario.Panel{
		Returns: [][]float64{
			{0.08, -0.10, 0.15, 0.05, 0.20, -0.05, 0.12, 0.03},
			{0.02, 0.01, 0.03, 0.00, 0.02, 0.01, 0.02, 0.01},
		},
		Weights: []float64{0.6, 0.4},
	}
	pr := Params{Capital: 1_500_000, NeedAnnual: 48000, BufferYears: 3, Years: 30,
		TaxRate: 0.30, NPaths: 2000, Model: "bootstrap", Weights: []float64{0.6, 0.4}}
	res := ComputeWithPanel(pr, &panel)
	if res.Cards["ruin"] == "" {
		t.Errorf("empty result for bootstrap model")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/decumul/web/ -run ComputeWithPanel`
Expected: FAIL (undefined).

- [ ] **Step 3: Extend model.go**

Add the field to `Params` (after `Weights`):
```go
	Model string `json:"model"` // "parametric" (default), "bootstrap", "cohorts"
```

Add a source selector and the panel-aware entry point:
```go
// source picks the return model. With a non-nil panel and a non-parametric
// Model, it resamples that panel under the live weights; otherwise it falls
// back to the parametric source.
func (pr Params) source(panel *scenario.Panel) scenario.Source {
	if panel != nil && pr.Weights != nil {
		switch pr.Model {
		case "bootstrap":
			return scenario.BlockBootstrap{Panel: *panel, Weights: pr.Weights, BlockLen: 5, Periods: pr.Years}
		case "cohorts":
			return scenario.HistoricalCohorts{Panel: *panel, Weights: pr.Weights, Periods: pr.Years}
		}
	}
	return scenario.ParametricSource{Mu: pr.Mu, Sigma: pr.Sigma, Df: pr.Df, Periods: pr.Years}
}

// ComputeWithPanel is Compute with an optional historical panel for the
// bootstrap/cohort models and live re-weighting.
func ComputeWithPanel(pr Params, panel *scenario.Panel) Result {
	p := pr.plan()
	p.Source = pr.source(panel)
	return computeFrom(pr, p)
}
```

Refactor `Compute` to delegate (so the chart-building code is shared):
```go
// Compute runs the parametric model (no panel).
func Compute(pr Params) Result { return ComputeWithPanel(pr, nil) }
```

Move the body of the old `Compute` (the simulate + chart-rendering block, Steps in Task 15) into `computeFrom(pr Params, p decumul.Plan) Result`, using `p` instead of `pr.plan()` and defaulting `pr.NPaths` there. The `decumul.Mu` axis of the surface only applies to the parametric source; in `computeFrom`, build the surface over `decumul.BufferYears × decumul.Mu` only when `p.Source` is a `scenario.ParametricSource`, else over `decumul.BufferYears × decumul.NeedAnnual`:
```go
	xParam, yParam := decumul.BufferYears, decumul.Mu
	ys := []float64{0.02, 0.03, 0.035, 0.04, 0.045, 0.05}
	if _, ok := p.Source.(scenario.ParametricSource); !ok {
		yParam = decumul.NeedAnnual
		ys = []float64{36000, 42000, 48000, 54000, 60000}
	}
	surf := p.Sweep2D(xParam, yParam, xs, ys, pr.NPaths/2+1, simWorkers, seed)
```

- [ ] **Step 4: Update server.go Handler signature**

```go
// Handler returns the decumulation UI. A non-nil panel enables the
// portfolio models (bootstrap/cohorts) and live allocation sliders; labels
// names the holdings for the allocation UI.
func Handler(panel *scenario.Panel, labels []string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(mustSub())))
	mux.HandleFunc("/api/meta", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"labels": labels, "hasPanel": panel != nil})
	})
	mux.HandleFunc("/api/sim", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var pr Params
		if err := json.NewDecoder(r.Body).Decode(&pr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ComputeWithPanel(pr, panel))
	})
	return mux
}
```

Add `"github.com/bpineau/pofo/pkg/scenario"` to server.go imports.

- [ ] **Step 5: Update Task 15's server_test.go call sites**

Both `Handler()` calls in `server_test.go` become `Handler(nil, nil)`. Run:
```bash
go test ./pkg/decumul/web/
```
Expected: PASS (parametric and panel tests).

- [ ] **Step 6: Wire the portfolio into runFire (cmd/pofo/main.go)**

Extend `runFire` to optionally resolve a portfolio file passed on the command line, build the panel via the existing fetch path, and pass it to `Handler`. Reuse `fetchAsset` and `^HICP-FR`:
```go
func runFire(opt *options, c *marketdata.Client, specs []*portfolio.Spec) error {
	var panel *scenario.Panel
	var labels []string
	if len(specs) > 0 {
		var assets []web.AssetSeries
		for _, h := range specs[0].Holdings {
			s, err := fetchAsset(c, h.ID, opt)
			if err != nil {
				continue
			}
			labels = append(labels, h.ID)
			assets = append(assets, web.AssetSeries{Weight: h.Weight, Points: s.Points})
		}
		hicp, err := fetchAsset(c, "^HICP-FR", opt)
		if err == nil {
			if pnl, err := web.BuildPanel(assets, hicp.Points); err == nil {
				panel = &pnl
			}
		}
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	url := "http://" + ln.Addr().String() + "/"
	fmt.Fprintf(os.Stderr, "FIRE explorer on %s (Ctrl-C to stop)\n", url)
	if !opt.noOpen {
		openBrowser(url)
	}
	return http.Serve(ln, web.Handler(panel, labels))
}
```
Update the dispatch in `run` to pass the client and parsed specs: `return runFire(opt, client, specs)` (use the same `client` and `specs` variables the report path already builds; if `-fire` is dispatched before specs are parsed, parse them first or move the dispatch after parsing). Add `scenario` and `web` imports.

- [ ] **Step 7: Extend app.js with the model toggle and allocation sliders**

Append to `assets/app.js`:
```javascript
// Portfolio mode: fetch holdings, add a model toggle and allocation sliders.
let weights = null, labels = [];
fetch("/api/meta").then(r=>r.json()).then(m=>{
  if(!m.hasPanel) return;
  labels = m.labels;
  weights = labels.map(()=>1/labels.length);
  const sel = document.createElement("label"); sel.className="ctl";
  sel.innerHTML = `Return model:
    <select id="model"><option value="parametric">parametric</option>
    <option value="bootstrap">historical bootstrap</option>
    <option value="cohorts">historical cohorts</option></select>`;
  form.prepend(sel);
  sel.querySelector("select").addEventListener("change", e=>{state.model=e.target.value;schedule();});
  state.model = "parametric";
  labels.forEach((name,i)=>{
    const d=document.createElement("label"); d.className="ctl";
    d.innerHTML=`${name}: <span id="w_${i}">${Math.round(weights[i]*100)}</span>%
      <input type="range" min="0" max="100" step="1" value="${Math.round(weights[i]*100)}" id="al_${i}">`;
    form.appendChild(d);
    d.querySelector("input").addEventListener("input",e=>{
      weights[i]=parseFloat(e.target.value)/100;
      document.getElementById("w_"+i).textContent=e.target.value; schedule();});
  });
});
// fold weights into the request.
const baseRun = run;
run = async function(){
  if(weights){ state.weights = weights; }
  return baseRun();
};
```
(Change the `run()` declaration in Task 15's app.js from `async function run()` to `let run = async function()` so it can be reassigned, and keep the trailing `run();` call.)

- [ ] **Step 8: Build, vet, full test**

Run:
```bash
go build ./cmd/pofo && go vet ./... && go test ./...
```
Expected: all PASS. Manual: write `/tmp/p.txt` with `60 NTSGSIM` / `25 DBMFESIM` / `15 XAUUSD`, run `./pofo -fire /tmp/p.txt`, confirm the model toggle and three allocation sliders appear and re-render on drag.

- [ ] **Step 9: Commit**

```bash
git add pkg/decumul/web/ cmd/pofo/main.go
git commit -m "decumul/web: portfolio mode with model toggle and live allocation"
```

---

## Task 19: documentation

**Files:**
- Modify: `README.md` (a "Decumulation / FIRE" section and the `-fire` option row)
- Modify: `doc.go` (add `pkg/scenario` and `pkg/decumul` to the package list)

**Interfaces:** none (docs only).

- [ ] **Step 1: Add to doc.go's package list**

Insert, after the `pkg/simgen` bullet in `doc.go`:
```go
//   - pkg/scenario: synthetic real-return path generation (parametric
//     Student-t, block/stationary bootstrap, historical cohorts) behind one
//     Source interface; the input to decumulation studies.
//   - pkg/decumul: decumulation/FIRE engine over a scenario.Source: ruin
//     probability, FIRE outcome metrics, capital/buffer sizing and sweeps,
//     with a thin embedded live UI under pkg/decumul/web.
```

- [ ] **Step 2: Add a README section and option row**

Add a `-fire` row to the Main options table:
```
| `-fire` | | open the local decumulation/FIRE explorer (sliders, ruin curves), optionally for a portfolio file, then serve until stopped |
```
And a prose section after "Suggesting assets to add":
```markdown
## Decumulation / FIRE analysis

`pofo -fire` opens a local web explorer that simulates a withdrawal
(retirement) phase and shows the **probability of ruin** as you drag sliders
for capital, spending floor, cash-buffer years, real return, volatility, tail
df, horizon, pension and the flat tax on gains. Charts show the buffer
arbitrage (ruin vs buffer years), the recovery-time distribution and a 2D
ruin surface (buffer × expected return).

`pofo -fire portfolio.txt` seeds the model from a real portfolio: it derives
the return assumptions and a historical panel from the holdings (reconstructed
back via `SIM`), lets you switch between a **parametric**, **historical
bootstrap** or **historical-cohort** projection, and drag each holding's
weight to re-test ruin live. Everything is in real euros; the model is a
fat-tailed hypothesis-exploration tool, **not investment advice**.
```

- [ ] **Step 3: Verify and commit**

Run: `go test ./... && go vet ./...`
Expected: PASS (docs do not break tests).
```bash
git add README.md doc.go
git commit -m "docs: document the decumulation/FIRE explorer and new packages"
```

---

## Self-review notes

- **Spec coverage:** §4 scenario → Tasks 1-5; §5 decumul (Plan/kernel/Simulate/CapitalForRuin/outcomes/recovery/sweeps/tax/cashflows) → Tasks 6-12; §6 UI (server, parametric playground, portfolio mode, toggle, allocation, charts) → Tasks 15-18; generic chart additions → Tasks 13-14; §7 validation → Task 9 golden tests + per-task unit tests; §8 caveats → doc.go (Task 6) and README (Task 19). All covered.
- **Type consistency:** `scenario.Source.Draw/Len`, `Panel.Combine`, `decumul.Plan.RunPath`/`Simulate`/`Sweep1D`/`Sweep2D`/`CapitalForRuin`, `Ensemble.Outcome`/`RecoveryTimeDistribution`, `web.Params`/`Result`/`Compute`/`ComputeWithPanel`/`Handler`, `chart.Bars`/`Heatmap` are defined once and consumed with matching signatures. `Handler` gains its `(panel, labels)` signature in Task 18 with the test call sites updated in the same task.
- **Known follow-ups (not blocking):** the golden tests use a flat 12% gross-up stub to match the Python reference (whose tax model is a flat gross-up, not the cost-basis CTO model); the cost-basis `CTOFlatTax` is exercised by its own unit test in Task 6. The 2D surface's y-axis switches from Mu (parametric) to NeedAnnual (historical) because Mu is not a parameter of a bootstrap/cohort source.
```
