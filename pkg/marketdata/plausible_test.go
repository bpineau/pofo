package marketdata

import (
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
