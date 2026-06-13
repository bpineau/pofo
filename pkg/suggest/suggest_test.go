package suggest

import (
	"math"
	"strings"
	"testing"
)

func TestRegimes(t *testing.T) {
	cases := []struct {
		m    Meta
		want []Regime
	}{
		{Meta{AssetClass: "equity", Benchmark: "MSCI World"}, []Regime{Growth}},
		{Meta{AssetClass: "equity", Underlying: "gold mining equities"}, []Regime{Inflation}},
		{Meta{AssetClass: "equity", Benchmark: "MSCI World Enhanced Value"}, []Regime{Growth, Inflation}},
		{Meta{AssetClass: "gold"}, []Regime{Inflation, Crisis}},
		{Meta{AssetClass: "managed-futures"}, []Regime{Crisis, Inflation}},
		{Meta{AssetClass: "government-bond", Underlying: "20+ year treasuries"}, []Regime{Deflation}},
		{Meta{AssetClass: "long-volatility"}, []Regime{Deflation, Crisis}},
		{Meta{AssetClass: "multi-asset"}, []Regime{Growth, Deflation}},
		{Meta{AssetClass: "other"}, []Regime{Crisis}},
	}
	for _, c := range cases {
		got := Regimes(c.m)
		if !sameRegimes(got, c.want) {
			t.Errorf("Regimes(%q/%q) = %v, want %v", c.m.AssetClass, c.m.Benchmark, got, c.want)
		}
	}
}

func sameRegimes(a, b []Regime) bool {
	if len(a) != len(b) {
		return false
	}
	set := map[Regime]bool{}
	for _, r := range a {
		set[r] = true
	}
	for _, r := range b {
		if !set[r] {
			return false
		}
	}
	return true
}

func TestCoverageAndGaps(t *testing.T) {
	holdings := []Holding{
		{ID: "EQ", Weight: 0.6, HasMeta: true, Meta: Meta{AssetClass: "equity"}},
		{ID: "TLT", Weight: 0.4, HasMeta: true, Meta: Meta{AssetClass: "government-bond"}},
	}
	cov, uncl := Coverage(holdings)
	if uncl != 0 || math.Abs(cov[Growth]-0.6) > 1e-9 || math.Abs(cov[Deflation]-0.4) > 1e-9 ||
		cov[Inflation] != 0 || cov[Crisis] != 0 {
		t.Fatalf("coverage = %v (uncl %v)", cov, uncl)
	}
	gaps := Gaps(cov, 0.10)
	if len(gaps) != 2 || gaps[0] != Inflation || gaps[1] != Crisis {
		t.Fatalf("gaps = %v, want [inflation crisis]", gaps)
	}
}

func TestCorrelationAndDiversification(t *testing.T) {
	a := []float64{0.01, -0.01, 0.01, -0.01, 0.01, -0.01}
	if c := Correlation(a, a); math.Abs(c-1) > 1e-9 {
		t.Fatalf("corr(a,a) = %v", c)
	}
	neg := []float64{-0.01, 0.01, -0.01, 0.01, -0.01, 0.01}
	if c := Correlation(a, neg); math.Abs(c+1) > 1e-9 {
		t.Fatalf("corr(a,-a) = %v", c)
	}
	// Two uncorrelated equal-vol assets at 50/50 → diversification ratio √2.
	b := []float64{0.01, 0.01, -0.01, -0.01, 0.01, -0.01, -0.01, 0.01}
	c := []float64{0.01, -0.01, 0.01, -0.01, -0.01, 0.01, -0.01, 0.01}
	if cc := Correlation(b, c); math.Abs(cc) > 1e-9 {
		t.Fatalf("b,c not orthogonal: corr %v", cc)
	}
	dr := DiversificationRatio([]float64{0.5, 0.5}, [][]float64{b, c})
	if math.Abs(dr-math.Sqrt2) > 1e-9 {
		t.Fatalf("DR = %v, want √2", dr)
	}
}

func TestRedundancies(t *testing.T) {
	// Three near-identical S&P trackers + one uncorrelated bond.
	sp := []float64{0.01, -0.012, 0.008, -0.004, 0.011, -0.009, 0.006, -0.007}
	jitter := func(s []float64, d float64) []float64 {
		out := make([]float64, len(s))
		for i, x := range s {
			out[i] = x + d*float64((i%3)-1)*0.0001
		}
		return out
	}
	bond := []float64{0.002, 0.001, -0.001, 0.0015, -0.0005, 0.0008, 0.0009, -0.0011}
	holdings := []Holding{
		{ID: "CSPX", Weight: 0.25, Meta: Meta{AssetClass: "equity"}},
		{ID: "VUAA", Weight: 0.25, Meta: Meta{AssetClass: "equity"}},
		{ID: "SPYL", Weight: 0.25, Meta: Meta{AssetClass: "equity"}},
		{ID: "AGGH", Weight: 0.25, Meta: Meta{AssetClass: "aggregate-bond"}},
	}
	returns := [][]float64{sp, jitter(sp, 1), jitter(sp, 2), bond}
	groups := Redundancies(holdings, returns, 0.95)
	if len(groups) != 1 {
		t.Fatalf("want 1 redundancy group, got %d: %+v", len(groups), groups)
	}
	if len(groups[0].IDs) != 3 || math.Abs(groups[0].Weight-0.75) > 1e-9 {
		t.Fatalf("group = %+v, want the 3 S&P trackers at 0.75", groups[0])
	}
}

func TestAnalyzeSuggestsGapFiller(t *testing.T) {
	const n = 480
	portR := make([]float64, n)
	cand := make([]float64, n) // a calm, uncorrelated, positive-drift asset
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			portR[i] = 0.02
		} else {
			portR[i] = -0.018
		}
		cand[i] = 0.0012 + 0.0001*math.Sin(float64(i))
	}
	holdings := []Holding{{ID: "EQ", Weight: 1, HasMeta: true, Meta: Meta{AssetClass: "equity"}}}
	candidates := []Candidate{
		// Gold fills the inflation/crisis gap.
		{Meta: Meta{ID: "XAUUSD", AssetClass: "gold"}, PortReturns: portR, Returns: cand, Years: 5},
		// Another equity helps no gap (growth already covered) → excluded.
		{Meta: Meta{ID: "VOO", AssetClass: "equity"}, PortReturns: portR, Returns: cand, Years: 5},
	}
	res := Analyze(holdings, [][]float64{portR}, candidates, DefaultOptions())

	if len(res.Suggestions) != 1 {
		t.Fatalf("want 1 suggestion, got %d: %+v", len(res.Suggestions), res.Suggestions)
	}
	s := res.Suggestions[0]
	if s.Meta.ID != "XAUUSD" || s.Fills != Inflation {
		t.Fatalf("suggestion = %s filling %s, want XAUUSD/inflation", s.Meta.ID, s.Fills)
	}
	if s.Weight <= 0 || s.Windows == 0 || s.SharpeWins*2 < s.Windows {
		t.Fatalf("weak suggestion: %+v", s)
	}
	if s.VolAfter >= s.VolBefore {
		t.Fatalf("adding a calm uncorrelated asset should lower vol: %.5f -> %.5f", s.VolBefore, s.VolAfter)
	}
}

func TestLoadMeta(t *testing.T) {
	js := `[{"id":"XAUUSD","isin":null,"asset_class":"gold","benchmark_index":null,"leverage":1.0},
	         {"id":"IE00B4L5Y983","isin":"IE00B4L5Y983","asset_class":"equity","leverage":1.0}]`
	m, err := LoadMeta(strings.NewReader(js))
	if err != nil {
		t.Fatal(err)
	}
	if m["XAUUSD"].AssetClass != "gold" || m["IE00B4L5Y983"].AssetClass != "equity" {
		t.Fatalf("parsed %+v", m)
	}
	// ISIN also resolves.
	if _, ok := m["IE00B4L5Y983"]; !ok {
		t.Fatal("ISIN key missing")
	}
}
