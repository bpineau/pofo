package marketdata

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
)

// walk builds a daily series of n points from a fixed seed of returns scaled to
// the wanted daily sigma, so a test can hand the doctor a series of a chosen
// shape without pulling in a random source.
func walk(n int, start, dailySigma, dailyDrift float64) []Point {
	steps := []float64{1, -0.6, 0.3, -1.2, 0.9, -0.4, 1.4, -1.1, 0.2, -0.7}
	out := make([]Point, n)
	day := time.Date(2010, 1, 4, 0, 0, 0, 0, time.UTC)
	v := start
	for i := range out {
		out[i] = Point{Date: day.AddDate(0, 0, i), Close: v}
		v *= 1 + dailyDrift + dailySigma*steps[i%len(steps)]
	}
	return out
}

// messages flattens issues to their text, for substring assertions.
func messages(issues []Issue) string {
	var b strings.Builder
	for _, is := range issues {
		b.WriteString(is.String())
		b.WriteByte('\n')
	}
	return b.String()
}

// judge is what VerifyAsset does with the record's band, in one call: the
// exemptions, the leverage scaling, then the whole-history verdicts.
func judge(a datasets.Asset, s *Series) []Issue {
	band, ok := assetBand(a, s)
	if !ok {
		return nil
	}
	return plausibilityIssues(a, s, band)
}

func TestPlausibilityIssues(t *testing.T) {
	// An aggregate-bond record served through a foreign-currency line: the
	// symptom the campaign kept meeting is a bond fund at equity volatility.
	bond := datasets.Asset{AssetClass: "aggregate-bond", Leverage: 1}
	loud := &Series{Symbol: "X", Points: walk(1200, 100, 0.012, 0)}
	got := messages(judge(bond, loud))
	if !strings.Contains(got, "volatility") || !strings.Contains(got, "aggregate-bond") {
		t.Fatalf("a 19 %%/yr aggregate-bond series must be flagged, got:\n%s", got)
	}

	// The same shape is ordinary for equity.
	equity := datasets.Asset{AssetClass: "equity", Leverage: 1}
	if got := judge(equity, loud); len(got) != 0 {
		t.Fatalf("equity at 19 %%/yr must be clean, got %v", got)
	}

	// Leverage widens every bound: what is too loud for a plain aggregate-bond
	// fund is ordinary for a 90/60 stacked one carrying the same class label.
	middling := &Series{Symbol: "X", Points: walk(1200, 100, 0.0085, 0)}
	if got := judge(bond, middling); len(got) == 0 {
		t.Fatal("a 13 %/yr unlevered aggregate-bond series must be flagged")
	}
	if got := judge(datasets.Asset{AssetClass: "aggregate-bond", Leverage: 1.5}, middling); len(got) != 0 {
		t.Fatalf("the same series at leverage 1.5 must be clean, got %v", got)
	}

	// A young share class is not asked about its CAGR: two years of a hot
	// start would leave every class's band.
	young := &Series{Symbol: "X", Points: walk(400, 100, 0.004, 0.002)}
	if got := messages(judge(equity, young)); strings.Contains(got, "CAGR") {
		t.Fatalf("under three years, no CAGR verdict; got:\n%s", got)
	}
	old := &Series{Symbol: "X", Points: walk(1600, 100, 0.004, 0.002)}
	if got := messages(judge(equity, old)); !strings.Contains(got, "CAGR") {
		t.Fatalf("a +65 %%/yr equity CAGR over four years must be flagged; got:\n%s", got)
	}
}

// TestPlausibilitySingleName covers the second property that stretches a class
// band: a record naming ONE issuer. Every row of the table describes a
// diversified holding, and no single stock's volatility, earnings gap or
// drawdown fits inside the equity row.
func TestPlausibilitySingleName(t *testing.T) {
	// 55 %/yr, the shape the Datadog employee-savings FCPE measures.
	single := &Series{Symbol: "X", Points: walk(1200, 100, 0.033, 0.0009)}
	fund := datasets.Asset{AssetClass: "equity", Leverage: 1}
	if got := judge(fund, single); len(got) == 0 {
		t.Fatal("a diversified equity fund at 55 %/yr must be flagged")
	}
	stock := datasets.Asset{AssetClass: "equity", Leverage: 1, Strategy: datasets.StrategySingleStock}
	if got := judge(stock, single); len(got) != 0 {
		t.Fatalf("the same shape on a single-name record must be clean, got %v", got)
	}
	// The stretch is a widening, not an exemption: a series no share makes
	// honestly still leaves the band, and the verdict says what widened it.
	wild := &Series{Symbol: "X", Points: walk(1200, 100, 0.07, 0)}
	if got := messages(judge(stock, wild)); !strings.Contains(got, "single name") {
		t.Fatalf("a 115 %%/yr single-name series must be flagged, and say why; got:\n%s", got)
	}
	// One earnings print is allowed to gap: DDOG did +31 % in one session.
	if b, ok := assetBand(stock, single); !ok || b.Move < 0.32 {
		t.Errorf("a single name must be allowed an earnings gap, Move = %.2f", b.Move)
	}
}

func TestPlausibilityExemptions(t *testing.T) {
	loud := &Series{Symbol: "X", Points: walk(1200, 100, 0.012, 0)}
	for _, tc := range []struct {
		name  string
		asset datasets.Asset
		s     *Series
	}{
		{"index reconstruction", datasets.Asset{AssetClass: "aggregate-bond", Source: "index"}, loud},
		{"rate level", datasets.Asset{AssetClass: "government-bond", Symbol: "^TNX"}, loud},
		{"policy rate", datasets.Asset{AssetClass: "money-market", Symbol: "^ESTR"}, loud},
		{"continuous future", datasets.Asset{AssetClass: "gold", Symbol: "GC=F"}, loud},
		{"unknown class", datasets.Asset{AssetClass: "crypto"}, loud},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := judge(tc.asset, tc.s); len(got) != 0 {
				t.Fatalf("must be exempt from plausibility, got %v", got)
			}
		})
	}
}

func TestIdentityIssues(t *testing.T) {
	day := func(y, m, d int) time.Time { return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC) }
	base := &Series{Symbol: "X", Currency: "EUR", Name: "Acme Bond UCITS ETF EUR Acc",
		Points: []Point{{Date: day(2010, 1, 4), Close: 100}, {Date: day(2010, 1, 5), Close: 101}, {Date: day(2010, 1, 6), Close: 102}}}

	t.Run("currency", func(t *testing.T) {
		// The FOLOW shape: a euro fund re-resolved onto a Swiss listing.
		a := datasets.Asset{Currency: "EUR"}
		s := *base
		s.Currency = "CHF"
		if got := messages(identityIssues(a, &s)); !strings.Contains(got, "serves CHF") {
			t.Fatalf("a served CHF line under a EUR record must be flagged, got:\n%s", got)
		}
		// GBp and GBP are the same line spelled two ways.
		a.Currency = "GBp"
		s.Currency = "GBP"
		if got := identityIssues(a, &s); len(got) != 0 {
			t.Fatalf("GBp vs GBP must not be a finding, got %v", got)
		}
	})

	t.Run("share class", func(t *testing.T) {
		// The IBGS.L shape: the record names the accumulating class, the
		// provider serves its distributing sibling.
		s := *base
		s.Name = "iShares € Govt Bond 1-3yr UCITS ETF EUR (Dist)"
		if got := messages(identityIssues(datasets.Asset{Distribution: "accumulating"}, &s)); !strings.Contains(got, "reads distributing") {
			t.Fatalf("a (Dist) name under an accumulating record must be flagged, got:\n%s", got)
		}
		if got := identityIssues(datasets.Asset{Distribution: "distributing"}, &s); len(got) != 0 {
			t.Fatalf("a (Dist) name under a distributing record is agreement, got %v", got)
		}
		// A name carrying both markers says nothing either way.
		s.Name = "Acme Fund Acc (Dist share class family)"
		if got := identityIssues(datasets.Asset{Distribution: "accumulating"}, &s); len(got) != 0 {
			t.Fatalf("an ambiguous name must not be a finding, got %v", got)
		}
	})

	t.Run("since", func(t *testing.T) {
		// Predecessor history served under a later class's name (the PFOCX
		// shape), and provider depth that never reaches the inception.
		if got := messages(identityIssues(datasets.Asset{Since: "2012-06-01"}, base)); !strings.Contains(got, "predecessor history") {
			t.Fatalf("quotes starting 2.4 years before inception must be flagged, got:\n%s", got)
		}
		if got := messages(identityIssues(datasets.Asset{Since: "2007-01-01"}, base)); !strings.Contains(got, "missing provider depth") {
			t.Fatalf("quotes starting 3 years after inception must be flagged, got:\n%s", got)
		}
		// Within a year, either way, is ordinary.
		if got := identityIssues(datasets.Asset{Since: "2009-06-01"}, base); len(got) != 0 {
			t.Fatalf("a seven-month drift must be silent, got %v", got)
		}
		// A SIM-extended series starts before every quote by construction.
		s := *base
		s.SimulatedBefore = day(2010, 1, 4)
		if got := identityIssues(datasets.Asset{Since: "2012-06-01"}, &s); len(got) != 0 {
			t.Fatalf("a simulated series must not be asked, got %v", got)
		}
	})
}

func TestVerifyAssetAddsCatalogChecks(t *testing.T) {
	now := time.Date(2026, 8, 9, 0, 0, 0, 0, time.UTC)
	s := &Series{Symbol: "?", Currency: "EUR", Points: walk(60, 100, 0.001, 0)}
	for i := range s.Points {
		s.Points[i].Date = now.AddDate(0, 0, i-60)
	}
	// An identifier the catalog does not know gets series hygiene only.
	if got := VerifyAsset("NOT-A-CATALOG-ID", s, now); len(got) != len(Verify(s, now)) {
		t.Fatalf("an uncatalogued id must not gain catalog findings, got %v", got)
	}
	// A catalogued one is judged against its record; VOO is an equity ETF, and
	// a 1.6 %/yr series is far below what equity can be.
	got := messages(VerifyAsset("VOO", s, now))
	if !strings.Contains(got, "volatility") {
		t.Fatalf("a flat series pinned to an equity record must be flagged, got:\n%s", got)
	}
}

func TestBandForReachesEveryIdentifierShape(t *testing.T) {
	// Lookup answers for a canonical id, an ISIN and an alias; bandBySymbol
	// covers the one caller that only holds the provider symbol.
	for _, tc := range []struct{ id, why string }{
		{"VOO", "canonical id"},
		{"IE00B5M4WH52", "ISIN"},
		{"IEML.L", "provider symbol"},
	} {
		if got := bandFor(tc.id); got == widestBand {
			t.Errorf("%s (%s) fell back to the widest band", tc.id, tc.why)
		}
	}
	if got := bandFor("NOT-A-CATALOG-ID"); got != widestBand {
		t.Errorf("an unknown identifier must get the widest band, got %+v", got)
	}
}

// TestClassBandsCoverTheVocabulary keeps the table and the catalog in step: an
// asset_class with no band is an asset nobody checks.
func TestClassBandsCoverTheVocabulary(t *testing.T) {
	for _, a := range datasets.Catalog() {
		if _, ok := ClassBand(a.AssetClass); !ok {
			t.Errorf("%s: asset_class %q has no plausibility band", a.ID, a.AssetClass)
		}
	}
	for class, b := range classBands {
		switch {
		case b.VolLo >= b.VolHi:
			t.Errorf("%s: empty volatility band [%g, %g]", class, b.VolLo, b.VolHi)
		case b.CAGRLo >= b.CAGRHi:
			t.Errorf("%s: empty CAGR band [%g, %g]", class, b.CAGRLo, b.CAGRHi)
		case b.Move <= 0 || b.Drawdown <= 0 || b.Drawdown > 1:
			t.Errorf("%s: implausible Move/Drawdown %g/%g", class, b.Move, b.Drawdown)
		}
	}
}

// TestMeasureRefusals: the shape of a series is what every plausibility
// verdict is computed from, so it must decline rather than produce a number
// on data that cannot carry one.
func TestMeasureRefusals(t *testing.T) {
	day := func(i int) time.Time { return time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	cases := []struct {
		name string
		pts  []Point
	}{
		{"too short", []Point{{Date: day(0), Close: 100}, {Date: day(1), Close: 101}}},
		{"no span at all", []Point{
			{Date: day(0), Close: 100}, {Date: day(0), Close: 101}, {Date: day(0), Close: 102}}},
		{"a non-positive first close", []Point{
			{Date: day(0), Close: 0}, {Date: day(1), Close: 101}, {Date: day(2), Close: 102}}},
		{"a non-positive last close", []Point{
			{Date: day(0), Close: 100}, {Date: day(1), Close: 101}, {Date: day(2), Close: 0}}},
		{"a non-positive close inside", []Point{
			{Date: day(0), Close: 100}, {Date: day(1), Close: 0}, {Date: day(2), Close: 102}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, ok := measure(tc.pts); ok {
				t.Error("measure should have declined")
			}
			// And the verdict built on it stays silent rather than guessing.
			if got := plausibilityIssues(datasets.Asset{AssetClass: "equity_developed"},
				&Series{Symbol: "X", Points: tc.pts}, Band{VolHi: 0.42}); len(got) != 0 {
				t.Errorf("issues = %v, want none", got)
			}
		})
	}
}

// TestBandScaleGuards: a band widens linearly with leverage, except the
// drawdown, which cannot reach a total loss, and a missing leverage reads as 1
// rather than collapsing every bound to zero.
func TestBandScaleGuards(t *testing.T) {
	b := Band{VolLo: 0.05, VolHi: 0.20, CAGRLo: -0.02, CAGRHi: 0.12, Move: 0.10, Drawdown: 0.60}
	for _, lev := range []float64{0, -1} {
		if got := b.Scale(lev); got != b {
			t.Errorf("Scale(%v) = %+v, want the band unchanged", lev, got)
		}
	}
	got := b.Scale(3)
	near := func(a, want float64) bool { return math.Abs(a-want) < 1e-12 }
	if !near(got.VolHi, 0.60) || !near(got.Move, 0.30) || !near(got.CAGRLo, -0.06) {
		t.Errorf("Scale(3) = %+v", got)
	}
	if got.Drawdown != 0.99 {
		t.Errorf("Scale(3).Drawdown = %v, want the 0.99 cap", got.Drawdown)
	}
}

// TestRecordStretch: the two properties that widen a class band, and the fact
// that they compound.
func TestRecordStretch(t *testing.T) {
	cases := []struct {
		name string
		a    datasets.Asset
		want float64
	}{
		{"a plain fund", datasets.Asset{}, 1},
		{"a 3x fund", datasets.Asset{Leverage: 3}, 3},
		{"a single name", datasets.Asset{Strategy: datasets.StrategySingleStock}, singleNameStretch},
		{"both compound", datasets.Asset{Leverage: 2, Strategy: datasets.StrategySingleStock}, 2 * singleNameStretch},
	}
	for _, tc := range cases {
		if got := recordStretch(tc.a); got != tc.want {
			t.Errorf("%s: recordStretch = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestVerifyAssetOffCatalog: everything the catalog cannot judge falls back to
// the plain series checks, and never panics on a nil or a two-point series.
func TestVerifyAssetOffCatalog(t *testing.T) {
	now := time.Date(2020, 1, 10, 0, 0, 0, 0, time.UTC)
	if got := VerifyAsset("VOO", nil, now); len(got) != 1 || got[0].Severity != "error" {
		t.Errorf("a nil series = %v, want the single 'no quotes' error", got)
	}
	short := &Series{Symbol: "VOO", Points: []Point{
		{Date: now.AddDate(0, 0, -2), Close: 100}, {Date: now.AddDate(0, 0, -1), Close: 101}}}
	if got := VerifyAsset("VOO", short, now); len(got) != 0 {
		t.Errorf("a clean two-point series = %v, want no issue", got)
	}
	// An identifier the catalog does not know has no class to be judged
	// against, so only the blanket checks run: a 90 % session is beyond any
	// unlevered asset and must still be caught.
	unknown := &Series{Symbol: "NOSUCH", Points: []Point{
		{Date: now.AddDate(0, 0, -3), Close: 100},
		{Date: now.AddDate(0, 0, -2), Close: 190},
		{Date: now.AddDate(0, 0, -1), Close: 191}}}
	if got := countMoves(VerifyAsset("NOSUCHIDENTIFIER", unknown, now)); got != 1 {
		t.Errorf("got %d move finding(s), want 1", got)
	}
}

// TestIdentityDistributingRecordServedAccumulating is the mirror of the (Dist)
// case: the record names the distributing class and the provider serves the
// accumulating sibling, whose NAV is a total return and would flatter it.
func TestIdentityDistributingRecordServedAccumulating(t *testing.T) {
	day := func(i int) time.Time { return time.Date(2010, 1, 4, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i) }
	s := &Series{Symbol: "X", Name: "iShares Core MSCI World UCITS ETF USD (Acc)", Points: []Point{
		{Date: day(0), Close: 100}, {Date: day(1), Close: 101}, {Date: day(2), Close: 102}}}
	if got := messages(identityIssues(datasets.Asset{Distribution: "distributing"}, s)); !strings.Contains(got, "reads accumulating") {
		t.Fatalf("an (Acc) name under a distributing record must be flagged, got:\n%s", got)
	}
}
