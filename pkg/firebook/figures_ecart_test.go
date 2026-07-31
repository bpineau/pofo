package firebook

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// ecartRecompute rebuilds the calendar-year real returns of one portfolio the
// way figures_ecart.go's comment says, on the very same construction as the
// sibling plate (ttMonths, monthly total returns, weights reset every
// December, US CPI deflator), so the two plates of the article cannot disagree.
func ecartRecompute(t *testing.T, legs map[string]func(prev, cur string) float64,
	weights map[string]float64, cpi map[string]float64) []float64 {
	t.Helper()
	mons := ttMonths()
	hold := make(map[string]float64, len(weights))
	for id, w := range weights {
		hold[id] = w
	}
	real := make([]float64, 1, len(mons))
	real[0] = 1
	base := cpi[mons[0]]
	for i := 1; i < len(mons); i++ {
		prev, cur := mons[i-1], mons[i]
		total := 0.0
		for id := range weights {
			hold[id] *= 1 + legs[id](prev, cur)
			total += hold[id]
		}
		real = append(real, total/(cpi[cur]/base))
		if strings.HasSuffix(cur, "-12") {
			for id, w := range weights {
				hold[id] = total * w
			}
		}
	}
	out := make([]float64, 0, ttYears)
	for y := 1; y <= ttYears; y++ {
		out = append(out, (real[y*12]/real[(y-1)*12]-1)*100)
	}
	return out
}

// ecartLegs builds the return functions of every sleeve the two all-weather
// plates use, out of the bundled series alone.
func ecartLegs(t *testing.T) (map[string]func(prev, cur string) float64, map[string]float64) {
	t.Helper()
	sim, ref := datasets.Simdata(), datasets.Refdata()
	price := map[string]map[string]float64{
		"equities":   ttMonthEnds(t, sim, "SP500"),
		"long":       ttMonthEnds(t, sim, "TLT"),
		"short":      ttMonthEnds(t, sim, "SHY"),
		"gold":       ttMonthEnds(t, sim, "XAUUSD"),
		"smallvalue": ttMonthEnds(t, ref, "USSCV-USD"),
	}
	bills := ttMonthEnds(t, ref, "TBILL-3M")

	cpiSeries, err := marketdata.NewClient("").Fetch(t.Context(), "^CPI-US", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	cpi := map[string]float64{}
	for _, p := range cpiSeries.Points {
		if p.Date.Day() == 1 {
			cpi[p.Date.Format("2006-01")] = p.Close
		}
	}

	legs := map[string]func(prev, cur string) float64{}
	for name, s := range price {
		s := s
		legs[name] = func(prev, cur string) float64 { return s[cur]/s[prev] - 1 }
	}
	// SHY only starts in 1991-10: T-bills stand in for the short sleeve before
	shy := price["short"]
	legs["short"] = func(prev, cur string) float64 {
		if p, ok := shy[prev]; ok && p != 0 {
			if c, ok := shy[cur]; ok {
				return c/p - 1
			}
		}
		return bills[prev] / 1200
	}
	return legs, cpi
}

// The plate's 53 pairs are frozen literals, so the book's figures stay pure
// functions with no data dependency at render time. This rebuilds both legs
// from pkg/datasets and fails the moment the plate and the data disagree,
// which is also what happens when those series are regenerated.
func TestEcartFigureMatchesTheData(t *testing.T) {
	legs, cpi := ecartLegs(t)
	butterfly := ecartRecompute(t, legs, map[string]float64{
		"equities": .20, "smallvalue": .20, "long": .20, "short": .20, "gold": .20}, cpi)
	equities := ecartRecompute(t, legs, map[string]float64{"equities": 1}, cpi)

	if len(ecartYears) != ttYears {
		t.Fatalf("the plate holds %d years, the window has %d", len(ecartYears), ttYears)
	}
	for i, e := range ecartYears {
		if want := ttFirst + 1 + i; e.year != want {
			t.Fatalf("year %d out of order, expected %d", e.year, want)
		}
		if math.Abs(butterfly[i]-e.allWeather) > 0.06 {
			t.Errorf("%d: Golden Butterfly %.2f %%, plate says %.2f %%", e.year, butterfly[i], e.allWeather)
		}
		if math.Abs(equities[i]-e.equities) > 0.06 {
			t.Errorf("%d: 100 %% equities %.2f %%, plate says %.2f %%", e.year, equities[i], e.equities)
		}
	}
}

// The bands are the whole point of the plate, and the brief forbids drawing
// them from memory: they must be the actual longest runs of the actual series,
// with their actual dates.
func TestEcartBandsAreTheRealStreaks(t *testing.T) {
	want := []ecartStreak{{1983, 1986, 4}, {1988, 1991, 4}, {1994, 1999, 6}, {2012, 2017, 6}}
	got := ecartStreaks(ecartBandMin)
	if len(got) != len(want) {
		t.Fatalf("the plate bands %d runs, the series has %d of at least %d years: %v",
			len(want), len(got), ecartBandMin, got)
	}
	for i, s := range got {
		if s != want[i] {
			t.Errorf("run %d is %v, the plate draws %v", i, s, want[i])
		}
		if s.to-s.from+1 != s.length {
			t.Errorf("run %v does not span its own length", s)
		}
	}
	// no run longer than the two the plate calls the longest, and none of the
	// annotated lengths is off by a year
	longest := 0
	for _, s := range ecartStreaks(1) {
		if s.length > longest {
			longest = s.length
		}
	}
	if longest != 6 {
		t.Errorf("the longest run is %d years, the plate annotates 6", longest)
	}
	// every banded run must really be behind the index, year after year
	for _, s := range got {
		for _, e := range ecartYears {
			if e.year >= s.from && e.year <= s.to && e.gap() >= 0 {
				t.Errorf("%d sits inside the band %v with a gap of %+.2f", e.year, s, e.gap())
			}
		}
	}
}

// Every number the plate writes in words must come out of the same series.
func TestEcartPlateClaimsHold(t *testing.T) {
	behind, gbGrowth, eqGrowth := 0, 1.0, 1.0
	worst, best := ecartYears[0], ecartYears[0]
	for _, e := range ecartYears {
		if e.gap() < 0 {
			behind++
		}
		gbGrowth *= 1 + e.allWeather/100
		eqGrowth *= 1 + e.equities/100
		if e.gap() < worst.gap() {
			worst = e
		}
		if e.gap() > best.gap() {
			best = e
		}
	}
	if behind != 33 {
		t.Errorf("the plate says 33 lagging years, the series has %d", behind)
	}
	n := float64(len(ecartYears))
	drag := (math.Pow(eqGrowth, 1/n) - math.Pow(gbGrowth, 1/n)) * 100
	if math.Abs(drag-0.8) > 0.05 {
		t.Errorf("the plate says 0,8 point a year, the series says %.2f", drag)
	}
	// the two years the plate names, with the numbers it writes next to them
	if worst.year != 2013 || math.Abs(worst.allWeather-5) > 0.5 || math.Abs(worst.equities-30) > 0.5 {
		t.Errorf("the worst year is %d (%+.2f vs %+.2f), the plate says 2013, +5 %% contre +30 %%",
			worst.year, worst.allWeather, worst.equities)
	}
	if best.year != 2008 || math.Abs(best.gap()-32) > 0.5 {
		t.Errorf("the best year is %d at %+.2f points, the plate says 2008, +32", best.year, best.gap())
	}
	// the notch between the two four-year runs, the only year of that stretch
	// that is not a lagging one
	for _, e := range ecartYears {
		if e.year == 1987 && math.Abs(e.gap()-0.04) > 0.005 {
			t.Errorf("1987 is %+.4f point, the plate writes +0,04", e.gap())
		}
	}
}

// The plate draws its bands and its named years; a silent rename must not
// leave the reader with an unlabelled wash.
func TestEcartFigureRenders(t *testing.T) {
	svg := FigureSVG("tous-temps-ecart")
	for _, want := range []string{"6 ans", "4 ans", "2008 : +32", "2013 : +5 % contre +30 %", "1987 : +0,04"} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate no longer writes %q", want)
		}
	}
	if strings.Contains(svg, "rgba") || strings.Contains(svg, "opacity") {
		t.Error("no rgba and no opacity: crengine paints them solid black")
	}
	if strings.Contains(svg, "\u2014") { // the em-dash, banned book-wide
		t.Error("no em-dash in a figure")
	}
}

// Figure and prose must agree: the plate answers the article's most serious
// criticism, and the article must still carry it.
func TestEcartFigureAgreesWithTheArticle(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/portefeuilles-tous-temps.md")
	if err != nil {
		t.Fatal(err)
	}
	article := string(raw)
	if !strings.Contains(article, "::: figure tous-temps-ecart") {
		t.Error("the article must carry the plate")
	}
	if !strings.Contains(article, "L'écart à l'indice, la critique la plus sérieuse") {
		t.Error("the plate belongs to the passage the article calls its most serious criticism")
	}
}
