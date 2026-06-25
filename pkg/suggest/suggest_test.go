package suggest

import (
	"math"
	"strings"
	"testing"
)

func TestRegimes(t *testing.T) {
	cases := []struct {
		m    Meta
		want []Category
	}{
		{Meta{AssetClass: "equity", Benchmark: "MSCI World"}, []Category{Growth}},
		{Meta{AssetClass: "equity", Underlying: "gold mining equities"}, []Category{Inflation}},
		{Meta{AssetClass: "equity", Benchmark: "MSCI World Enhanced Value"}, []Category{Growth, Inflation}},
		{Meta{AssetClass: "gold"}, []Category{Inflation, Crisis}},
		{Meta{AssetClass: "managed-futures"}, []Category{Crisis, Inflation}},
		{Meta{AssetClass: "government-bond", Underlying: "20+ year treasuries"}, []Category{Deflation}},
		{Meta{AssetClass: "long-volatility"}, []Category{Deflation, Crisis}},
		{Meta{AssetClass: "multi-asset"}, []Category{Growth, Deflation}},
		{Meta{AssetClass: "other"}, []Category{Crisis}},
	}
	for _, c := range cases {
		got := Regimes(c.m)
		if !sameRegimes(got, c.want) {
			t.Errorf("Regimes(%q/%q) = %v, want %v", c.m.AssetClass, c.m.Benchmark, got, c.want)
		}
	}
}

func sameRegimes(a, b []Category) bool {
	if len(a) != len(b) {
		return false
	}
	set := map[Category]bool{}
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
	cov, uncl := Coverage(holdings, RegimeFramework())
	if uncl != 0 || math.Abs(cov[Growth]-0.6) > 1e-9 || math.Abs(cov[Deflation]-0.4) > 1e-9 ||
		cov[Inflation] != 0 || cov[Crisis] != 0 {
		t.Fatalf("coverage = %v (uncl %v)", cov, uncl)
	}
	gaps := Gaps(cov, RegimeFramework(), 0.10)
	if len(gaps) != 2 || gaps[0] != Inflation || gaps[1] != Crisis {
		t.Fatalf("gaps = %v, want [inflation crisis]", gaps)
	}
}

func TestCoverageExposureWeighted(t *testing.T) {
	// A 50% NTSX-style 90/60 stacked fund + 50% plain equity. The stacked
	// fund contributes its leg notionals, not a flat weight: equity legs sum
	// to growth, the bond leg to deflation.
	holdings := []Holding{
		{ID: "EQ", Weight: 0.5, HasMeta: true, Meta: Meta{AssetClass: "equity"}},
		{ID: "NTSX", Weight: 0.5, HasMeta: true, Meta: Meta{
			AssetClass: "multi-asset",
			Exposures:  map[string]float64{"equity": 0.9, "government-bond": 0.6},
		}},
	}
	cov, _ := Coverage(holdings, RegimeFramework())
	// growth: 0.5 (plain equity) + 0.5*0.9 (NTSX equity leg) = 0.95
	// deflation: 0.5*0.6 (NTSX bond leg) = 0.30
	if math.Abs(cov[Growth]-0.95) > 1e-9 || math.Abs(cov[Deflation]-0.30) > 1e-9 {
		t.Fatalf("exposure-weighted coverage = %v, want growth 0.95 deflation 0.30", cov)
	}
	// Factor framework: the bond leg is intermediate duration → term.
	fcov, _ := Coverage(holdings, FactorFramework())
	if math.Abs(fcov[Market]-0.95) > 1e-9 || math.Abs(fcov[Term]-0.30) > 1e-9 {
		t.Fatalf("factor coverage = %v, want market 0.95 term 0.30", fcov)
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
	res := Analyze(holdings, [][]float64{portR}, candidates, DefaultOptions(), RegimeFramework())

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

func TestFactorFramework(t *testing.T) {
	fw := FactorFramework()
	cases := []struct {
		m    Meta
		want []Category
	}{
		{Meta{AssetClass: "equity", Benchmark: "MSCI World"}, []Category{Market}},
		{Meta{AssetClass: "equity", Benchmark: "MSCI World Value"}, []Category{Market, Value}},
		{Meta{AssetClass: "equity", Benchmark: "MSCI World Momentum"}, []Category{Market, Momentum}},
		{Meta{AssetClass: "equity", Underlying: "small cap stocks"}, []Category{Market, Size}},
		{Meta{AssetClass: "gold"}, []Category{Alternative}},
		{Meta{AssetClass: "managed-futures"}, []Category{Alternative}},
		{Meta{AssetClass: "government-bond"}, []Category{Term}},
		{Meta{AssetClass: "aggregate-bond"}, []Category{Term, Credit}},
		{Meta{AssetClass: "money-market"}, []Category{Cash}},
	}
	for _, c := range cases {
		if got := fw.Classify(c.m); !sameRegimes(got, c.want) {
			t.Errorf("Factors(%q/%q) = %v, want %v", c.m.AssetClass, c.m.Benchmark, got, c.want)
		}
	}
}
