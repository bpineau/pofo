package suggest

import (
	"math"
	"reflect"
	"testing"
)

// The regime classifier is the report's whole macro story, so every branch of
// it is pinned here: the keyword refinements on the equity branch, the
// duration and hedging refinements on the sovereign one, and the per-class
// defaults. Order does not matter (Coverage reads a set), so the comparison
// is set-wise.
func TestRegimeLegBranches(t *testing.T) {
	cases := []struct {
		name string
		m    Meta
		want []Category
	}{
		{"plain equity", Meta{AssetClass: "equity", Benchmark: "MSCI World"}, []Category{Growth}},
		{"gold miners", Meta{AssetClass: "equity", Underlying: "precious metal miners"}, []Category{Inflation}},
		{"energy equity", Meta{AssetClass: "equity", Benchmark: "MSCI World Energy"}, []Category{Growth, Inflation}},
		{"dividend tilt", Meta{AssetClass: "equity", Benchmark: "FTSE All-World High Dividend Yield"}, []Category{Growth, Inflation}},
		{"real estate", Meta{AssetClass: "real-estate"}, []Category{Growth, Inflation}},
		{"corporate bond", Meta{AssetClass: "corporate-bond"}, []Category{Growth, Deflation}},
		{"ultra-short sovereign is cash", Meta{AssetClass: "government-bond", Duration: 0.3}, []Category{Deflation}},
		{"long unhedged sovereign hedges a crisis", Meta{AssetClass: "government-bond", Duration: 25}, []Category{Deflation, Crisis}},
		{"long hedged sovereign loses the currency bid", Meta{AssetClass: "government-bond", Duration: 25, CurrencyHedged: true}, []Category{Deflation}},
		{"intermediate aggregate", Meta{AssetClass: "aggregate-bond", Duration: 6}, []Category{Deflation}},
		{"linkers", Meta{AssetClass: "inflation-linked-bond"}, []Category{Inflation, Deflation}},
		{"money market", Meta{AssetClass: "money-market"}, []Category{Deflation}},
		{"broad commodity", Meta{AssetClass: "broad-commodity"}, []Category{Inflation, Crisis}},
		{"insurance-linked", Meta{AssetClass: "insurance-linked"}, []Category{Crisis}},
		{"tail risk", Meta{AssetClass: "tail-risk"}, []Category{Deflation, Crisis}},
		{"no class at all", Meta{}, []Category{Crisis}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Regimes(c.m); !sameRegimes(got, c.want) {
				t.Errorf("Regimes = %v, want %v", got, c.want)
			}
		})
	}
}

// The free-text notes are deliberately NOT part of the keyword hint: fund
// boilerplate ("distributes annual dividends") would otherwise turn every
// distributing world tracker into an inflation hedge.
func TestRegimeHintIgnoresNotes(t *testing.T) {
	m := Meta{AssetClass: "equity", Benchmark: "MSCI World", Notes: "distributes an annual dividend; energy is a large sector"}
	if got := Regimes(m); !sameRegimes(got, []Category{Growth}) {
		t.Errorf("Regimes = %v, want [growth]: the notes must not feed the classifier", got)
	}
}

// The factor classifier is coarser but has the same duty: pin the tilts it
// reads from the explicit factor list, from keywords when the list is empty,
// and the class defaults.
func TestFactorLegBranches(t *testing.T) {
	cases := []struct {
		name string
		m    Meta
		want []Category
	}{
		{"explicit tilts win", Meta{AssetClass: "equity", Factors: []string{"size", "value", "momentum", "quality", "low-vol"}},
			[]Category{Market, Size, Value, Momentum, Quality}},
		{"multi-factor keyword", Meta{AssetClass: "equity", Benchmark: "MSCI World Diversified Multi-Factor"},
			[]Category{Market, Value, Momentum, Quality}},
		{"gold miners are alternative", Meta{AssetClass: "equity", Underlying: "gold mining equities"}, []Category{Alternative}},
		{"small-cap value keywords", Meta{AssetClass: "equity", Benchmark: "MSCI Europe Small Cap Value"},
			[]Category{Market, Size, Value}},
		{"energy adds alternative", Meta{AssetClass: "equity", Benchmark: "MSCI World Oil & Gas"},
			[]Category{Market, Alternative}},
		{"plain equity is market only", Meta{AssetClass: "equity", Benchmark: "MSCI World"}, []Category{Market}},
		{"stacked fund", Meta{AssetClass: "multi-asset"}, []Category{Market, Term}},
		{"real estate", Meta{AssetClass: "real-estate"}, []Category{Market, Alternative}},
		{"short sovereign is cash", Meta{AssetClass: "government-bond", Duration: 0.3}, []Category{Cash}},
		{"long sovereign is term", Meta{AssetClass: "government-bond", Duration: 25}, []Category{Term}},
		{"short aggregate is cash", Meta{AssetClass: "aggregate-bond", Duration: 1}, []Category{Cash}},
		{"aggregate is term plus credit", Meta{AssetClass: "aggregate-bond", Duration: 6}, []Category{Term, Credit}},
		{"corporate bond", Meta{AssetClass: "corporate-bond"}, []Category{Credit, Term}},
		{"money market", Meta{AssetClass: "money-market"}, []Category{Cash}},
		{"managed futures", Meta{AssetClass: "managed-futures"}, []Category{Alternative}},
	}
	fw := FactorFramework()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := fw.Classify(c.m); !sameRegimes(got, c.want) {
				t.Errorf("Classify = %v, want %v", got, c.want)
			}
		})
	}
}

// A stacked fund is classified leg by leg, and the union must be stable run to
// run: the exposures arrive in a map, so the legs are visited in sorted order.
// A 90/60 carries the equity legs AND the sovereign ones, under both
// frameworks.
func TestClassifyExposuresUnionIsDeterministic(t *testing.T) {
	m := Meta{
		AssetClass: "multi-asset",
		Exposures:  map[string]float64{"equity": 0.9, "government-bond": 0.6, "gold": 0.1},
	}
	first := Regimes(m)
	for range 30 {
		if got := Regimes(m); !reflect.DeepEqual(got, first) {
			t.Fatalf("classification is not deterministic: %v vs %v", got, first)
		}
	}
	// Sorted legs: equity (growth), gold (inflation, crisis), government-bond
	// (deflation). The union keeps first-seen order.
	want := []Category{Growth, Inflation, Crisis, Deflation}
	if !reflect.DeepEqual(first, want) {
		t.Errorf("Regimes = %v, want %v", first, want)
	}
	// The legs' own asset_class drives the mapping, not the wrapper's:
	// "multi-asset" alone would read [growth deflation] and miss the gold leg.
	if got := FactorFramework().Classify(m); !sameRegimes(got, []Category{Market, Term, Alternative}) {
		t.Errorf("factor classification = %v, want [market term alternative]", got)
	}
}

// PortfolioReturns is the aggregate the suggestion engine compares a candidate
// against. It tolerates a short series (it stops at the shortest) and returns
// nothing at all when there is no asset.
func TestPortfolioReturns(t *testing.T) {
	got := PortfolioReturns([]float64{0.6, 0.4}, [][]float64{
		{0.01, 0.02, -0.01},
		{0.00, -0.01, 0.03},
	})
	want := []float64{0.006, 0.008, 0.006}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-12 {
			t.Errorf("PortfolioReturns[%d] = %v, want %v", i, got[i], want[i])
		}
	}
	if got := PortfolioReturns(nil, nil); got != nil {
		t.Errorf("PortfolioReturns(nil) = %v, want nil", got)
	}
	// A ragged input must not panic: the shorter leg simply stops
	// contributing.
	short := PortfolioReturns([]float64{0.5, 0.5}, [][]float64{{0.02, 0.02}, {0.04}})
	if len(short) != 2 || math.Abs(short[0]-0.03) > 1e-12 || math.Abs(short[1]-0.01) > 1e-12 {
		t.Errorf("ragged PortfolioReturns = %v, want [0.03 0.01]", short)
	}
}

// The walk-forward helpers: a flat window has no Sharpe (zero volatility, no
// division), and the median of an even-length sample is the midpoint of the
// two central values.
func TestWindowStatsAndMedian(t *testing.T) {
	if got := windowSharpe([]float64{0.01, 0.01, 0.01}); got != 0 {
		t.Errorf("windowSharpe(flat) = %v, want 0 (zero volatility)", got)
	}
	if got := windowSharpe([]float64{0.01, -0.005, 0.02, 0.0}); got <= 0 {
		t.Errorf("windowSharpe(positive drift) = %v, want > 0", got)
	}
	if got := windowMaxDD([]float64{0.10, -0.20, 0.05}); math.Abs(got-(-0.20)) > 1e-12 {
		t.Errorf("windowMaxDD = %v, want -0.20", got)
	}
	if got := windowMaxDD([]float64{0.01, 0.02}); got != 0 {
		t.Errorf("windowMaxDD(never underwater) = %v, want 0", got)
	}
	if got := median(nil); got != 0 {
		t.Errorf("median(nil) = %v, want 0", got)
	}
	if got := median([]float64{3, 1}); got != 2 {
		t.Errorf("median(even) = %v, want 2", got)
	}
	if got := median([]float64{3, 1, 2}); got != 2 {
		t.Errorf("median(odd) = %v, want 2", got)
	}
}

// A candidate that helps no gap, or whose benefit is not consistent enough
// across the walk-forward windows, is dropped; and the survivors are capped at
// one per asset class, because a second gold adds nothing the first did not.
func TestRankCandidatesFiltersAndDiversifies(t *testing.T) {
	const n = 400
	port := make([]float64, n)
	goldA := make([]float64, n)
	goldB := make([]float64, n)
	equity := make([]float64, n)
	for i := range n {
		port[i] = 0.004 * math.Sin(float64(i)/5)
		goldA[i] = 0.004*math.Cos(float64(i)/5) + 0.0004
		goldB[i] = 0.004*math.Cos(float64(i)/5) + 0.0002
		equity[i] = port[i] // a clone of what is already held
	}
	cov := map[Category]float64{Growth: 1, Deflation: 0, Inflation: 0, Crisis: 0}
	gaps := []Category{Inflation, Crisis}
	cands := []Candidate{
		{Meta: Meta{ID: "GOLDA", AssetClass: "gold"}, PortReturns: port, Returns: goldA, Years: 20},
		{Meta: Meta{ID: "GOLDB", AssetClass: "gold"}, PortReturns: port, Returns: goldB, Years: 20},
		{Meta: Meta{ID: "EQ", AssetClass: "equity"}, PortReturns: port, Returns: equity, Years: 20},
	}
	got := RankCandidates(gaps, cov, cands, DefaultOptions(), RegimeFramework())
	if len(got) != 1 {
		t.Fatalf("suggestions = %d (%+v), want 1: one per asset class, and equity fills no gap", len(got), got)
	}
	s := got[0]
	if s.Meta.ID != "GOLDA" {
		t.Errorf("kept %q, want the better of the two golds (GOLDA)", s.Meta.ID)
	}
	if s.Fills != Inflation && s.Fills != Crisis {
		t.Errorf("fills %q, want one of the gap categories", s.Fills)
	}
	if s.VolAfter >= s.VolBefore {
		t.Errorf("vol %v -> %v, want the anticorrelated sleeve to cut it", s.VolBefore, s.VolAfter)
	}
	if s.Windows == 0 || s.SharpeWins == 0 {
		t.Errorf("walk-forward not run: %+v", s)
	}
}
