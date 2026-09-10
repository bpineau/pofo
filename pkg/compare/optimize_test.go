package compare

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/marketdata"
	"github.com/bpineau/pofo/pkg/optimize"
	"github.com/bpineau/pofo/pkg/portfolio"
)

// optSeries builds a deterministic daily series over the given dates from a
// per-day return function, so a test can state what an asset does (drift and
// wiggle) instead of hard-coding prices.
func optSeries(symbol string, n int, ret func(i int) float64) *marketdata.Series {
	dates := dailyDates(n)
	pts := make([]marketdata.Point, n)
	px := 100.0
	for i, d := range dates {
		if i > 0 {
			px *= 1 + ret(i)
		}
		pts[i] = marketdata.Point{Date: d, Close: px}
	}
	return &marketdata.Series{Symbol: symbol, Name: symbol, Currency: "EUR", Source: "index", Points: pts}
}

// optFixture returns a two-line portfolio whose assets are deliberately
// unequal: STEADY compounds smoothly, WILD swings around the same drift. Any
// risk-aware objective must therefore prefer STEADY.
func optFixture(n int) *portfolio.Portfolio {
	steady := optSeries("STEADY", n, func(i int) float64 { return 0.0003 + 0.001*math.Sin(float64(i)/9) })
	wild := optSeries("WILD", n, func(i int) float64 { return 0.0003 + 0.012*math.Sin(float64(i)/4) })
	return &portfolio.Portfolio{Name: "P", Assets: []portfolio.Asset{
		{ID: "STEADY", Symbol: "STEADY", Weight: 0.5, Fees: -1, Series: steady},
		{ID: "WILD", Symbol: "WILD", Weight: 0.5, Fees: -1, Series: wild},
	}}
}

func optSpec(t *testing.T, text string) *portfolio.Spec {
	t.Helper()
	os, err := optimize.ParseSpec(text)
	if err != nil {
		t.Fatal(err)
	}
	return &portfolio.Spec{Name: "P", Optimize: &os}
}

// The written weights are the baseline the optimizer is shown against, so the
// computed portfolio must be a COPY: the original keeps the file's weights
// whatever the solver decides. The note is the report's whole account of that
// decision, so it must name the objective, the window it fitted on and the
// weights it chose.
func TestOptimizedPortfolioLeavesTheWrittenWeightsAlone(t *testing.T) {
	base := optFixture(1500)
	spec := optSpec(t, "min-volatility")

	got, note, err := optimizedPortfolio(base, spec, nil)
	if err != nil {
		t.Fatal(err)
	}
	if base.Assets[0].Weight != 0.5 || base.Assets[1].Weight != 0.5 {
		t.Errorf("the written weights moved: %v", base.Assets)
	}
	sum := got.Assets[0].Weight + got.Assets[1].Weight
	if math.Abs(sum-1) > 1e-6 {
		t.Errorf("optimized weights sum to %v, want 1", sum)
	}
	if got.Assets[0].Weight <= got.Assets[1].Weight {
		t.Errorf("min-volatility put %v in the wild line vs %v in the steady one",
			got.Assets[1].Weight, got.Assets[0].Weight)
	}
	if got.Name != "P (min-volatility)" {
		t.Errorf("name = %q, want the objective appended", got.Name)
	}
	for _, want := range []string{"min-volatility", "2000-01", "STEADY", "WILD", "in-sample CAGR", "volatility", "Sharpe"} {
		if !strings.Contains(note, want) {
			t.Errorf("note misses %q: %s", want, note)
		}
	}
}

// "train:" is what makes the report's numbers honest: the weights are fitted
// on a slice and the note reports how they did on the years the solver never
// saw. A holdout shorter than a year is not worth reporting and is omitted.
func TestOptimizedPortfolioReportsItsHoldout(t *testing.T) {
	base := optFixture(3000) // 2000-01-03 → 2008-03
	got, note, err := optimizedPortfolio(base, optSpec(t, "max-sharpe,train:2000..2003"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note, "which it did not see") {
		t.Errorf("note has no out-of-sample leg: %s", note)
	}
	if !strings.Contains(note, "over 2000-01-04→2003-12-31") {
		t.Errorf("note does not state the fitting window: %s", note)
	}
	if len(got.Assets) != 2 {
		t.Fatalf("assets = %d", len(got.Assets))
	}

	// The whole span fitted (no train:) leaves nothing outside, so no holdout
	// sentence.
	_, full, err := optimizedPortfolio(base, optSpec(t, "max-sharpe"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(full, "which it did not see") {
		t.Errorf("a full-span fit claims an out-of-sample leg: %s", full)
	}
}

// Every objective carries its own achieved figure into the note, and the two
// self-defeating configurations warn in it: an unconstrained max-return
// (degenerate by construction) and weight bounds handed to risk-parity, whose
// solver cannot honor them.
func TestOptimizedPortfolioObjectiveNotes(t *testing.T) {
	base := optFixture(1500)
	cases := []struct {
		spec string
		want string
	}{
		{"max-sortino", "achieved Sortino"},
		{"return-to-drawdown", "achieved return/max-drawdown"},
		{"min-ulcer", "achieved Ulcer Index"},
		{"max-worst-5y", "achieved worst rolling 5y return"},
		{"max-return", "unconstrained, max-return is degenerate"},
		{"max-return,max-vol:8", "max-return under vol"},
		{"risk-parity,max-weight:60", "weight bounds do not apply to risk-parity"},
	}
	for _, c := range cases {
		t.Run(c.spec, func(t *testing.T) {
			_, note, err := optimizedPortfolio(base, optSpec(t, c.spec), nil)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(note, c.want) {
				t.Errorf("note misses %q: %s", c.want, note)
			}
		})
	}
}

// A limit no allocation of these lines can meet must be said out loud: the
// returned weights are the least-violating point, not an answer, and a report
// that presented them as one would be lying by omission.
func TestOptimizedPortfolioFlagsInfeasibleLimits(t *testing.T) {
	base := optFixture(1500)
	_, note, err := optimizedPortfolio(base, optSpec(t, "max-return,max-vol:0.1"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note, "WARNING: no allocation of these holdings meets the limits") {
		t.Errorf("an infeasible solve went unflagged: %s", note)
	}
}

// Per-line bounds arrive keyed by the identifier the file writes, and are
// resolved here, where the holdings are known. A bound naming no holding is a
// typo, and a typo must fail loudly rather than read as "no bound".
func TestOptimizedPortfolioResolvesBounds(t *testing.T) {
	base := optFixture(1500)
	got, _, err := optimizedPortfolio(base, optSpec(t, "min-volatility,bounds:WILD:30-70"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if w := got.Assets[1].Weight; w < 0.30-1e-6 || w > 0.70+1e-6 {
		t.Errorf("WILD at %v, outside its 30-70 %% bound", w)
	}
	if _, _, err := optimizedPortfolio(base, optSpec(t, "min-volatility,bounds:TYPO:30-70"), nil); err == nil {
		t.Error("a bound on an unknown identifier was accepted")
	}
}

// Black-Litterman's prior is the file's own weights, so with no view at all
// the objective must return those weights EXACTLY: the identity the whole
// model hangs on, checked here through the report's entry point rather than
// only inside the solver.
func TestOptimizedPortfolioBlackLittermanIdentity(t *testing.T) {
	base := optFixture(1500)
	base.Assets[0].Weight, base.Assets[1].Weight = 0.7, 0.3

	got, note, err := optimizedPortfolio(base, optSpec(t, "black-litterman"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Assets[0].Weight-0.7) > 1e-6 || math.Abs(got.Assets[1].Weight-0.3) > 1e-6 {
		t.Errorf("no view moved the weights off the prior: %v", got.Assets)
	}
	if !strings.Contains(note, "No view moved an expected return") {
		t.Errorf("note does not say the posterior is the prior: %s", note)
	}
}

// The two ways the setup itself is impossible: assets that never quote on the
// same day (nothing to optimize over) and CWARP without the benchmark it
// scores against. Both must be errors, not a silent half-answer.
func TestOptimizedPortfolioRefusesImpossibleSetups(t *testing.T) {
	base := optFixture(1500)
	if _, _, err := optimizedPortfolio(base, optSpec(t, "cwarp"), nil); err == nil {
		t.Error("cwarp without a benchmark was accepted")
	}

	// Two series that never overlap: the second starts after the first ends.
	early := optSeries("EARLY", 300, func(int) float64 { return 0.0002 })
	late := optSeries("LATE", 300, func(int) float64 { return 0.0002 })
	shift := early.Points[len(early.Points)-1].Date.AddDate(0, 0, 10).Sub(late.Points[0].Date)
	for i := range late.Points {
		late.Points[i].Date = late.Points[i].Date.Add(shift)
	}
	disjoint := &portfolio.Portfolio{Name: "D", Assets: []portfolio.Asset{
		{ID: "EARLY", Symbol: "EARLY", Weight: 0.5, Fees: -1, Series: early},
		{ID: "LATE", Symbol: "LATE", Weight: 0.5, Fees: -1, Series: late},
	}}
	if _, _, err := optimizedPortfolio(disjoint, optSpec(t, "max-sharpe"), nil); err == nil {
		t.Error("assets with no common period were optimized anyway")
	}

	// A fitting window outside the data is refused by trainSpan, through here.
	if _, _, err := optimizedPortfolio(base, optSpec(t, "max-sharpe,train:1980..1985"), nil); err == nil {
		t.Error("a fitting window outside the data was accepted")
	}
}

// CWARP scores a blend against the benchmark, so the benchmark's returns must
// be aligned on the very same dates as the assets' and then split back off. A
// mis-split would score the blend against one of its own legs.
func TestOptimizedPortfolioCWARPUsesTheBenchmark(t *testing.T) {
	base := optFixture(1500)
	bench := optSeries("^BENCH", 1500, func(i int) float64 { return 0.0004 + 0.008*math.Sin(float64(i)/6) })

	got, note, err := optimizedPortfolio(base, optSpec(t, "cwarp"), bench)
	if err != nil {
		t.Fatal(err)
	}
	if sum := got.Assets[0].Weight + got.Assets[1].Weight; math.Abs(sum-1) > 1e-6 {
		t.Errorf("weights sum to %v, want 1", sum)
	}
	if !strings.Contains(note, "achieved CWARP") || !strings.Contains(note, "^BENCH") {
		t.Errorf("note does not report the CWARP against the benchmark: %s", note)
	}
	if !strings.Contains(note, "not a standalone portfolio") {
		t.Errorf("note omits the diversifier caveat: %s", note)
	}
}
