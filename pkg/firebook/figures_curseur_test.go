package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/bpineau/pofo/pkg/datasets"
	"github.com/bpineau/pofo/pkg/marketdata"
)

// The slider plate reads the same legs, the same window and the same engine as
// the family plate of the same article (ttMonthEnds / ttRecompute, in
// figures_tous_temps_test.go): one convention for real dollar portfolios,
// written once. That sharing is also what makes the control table below a
// meaningful check on this plate rather than a separate ritual.

// curseurLegs assembles the monthly real-return legs the sweep and the control
// table both draw on.
func curseurLegs(t *testing.T) (map[string]func(prev, cur string) float64, map[string]float64) {
	t.Helper()
	sim, ref := datasets.Simdata(), datasets.Refdata()
	price := map[string]map[string]float64{
		"equities":     ttMonthEnds(t, sim, "SP500"),
		"long":         ttMonthEnds(t, sim, "TLT"),
		"intermediate": ttMonthEnds(t, sim, "IEF"),
		"short":        ttMonthEnds(t, sim, "SHY"),
		"gold":         ttMonthEnds(t, sim, "XAUUSD"),
		"smallvalue":   ttMonthEnds(t, ref, "USSCV-USD"),
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
	legs["cash"] = func(prev, cur string) float64 { return bills[prev] / 1200 }
	legs["smallvalue"] = ttSmallValueNet(price["smallvalue"])
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

// curseurWeights is the portfolio at one dose of the slider: the growth core
// scaled down by the dose, the regime pocket scaled up by it.
func curseurWeights(dose float64) map[string]float64 {
	d := dose / 100
	return map[string]float64{
		"equities":     (1 - d) * curseurCoreEquities,
		"intermediate": (1 - d) * curseurCoreBonds,
		"gold":         d * curseurPocketGold,
		"long":         d * curseurPocketLong,
	}
}

// The implementation guardrail, run BEFORE the sweep is trusted: four
// portfolios of this family whose real return and real volatility are
// published. If this repository's series and conventions reproduce them, the
// sweep computed on the same engine means what it says; if they drift apart,
// the sweep is measuring something else and no amount of tuning would fix it.
//
// WHICH volatility, is the whole point of the row. The published figures below
// are portfoliocharts.com's, the reference the article names, and that site
// states what it measures: the "year-to-year variation in returns" of annual,
// inflation-adjusted, year-end data (Permanent Portfolio 7.6, Golden Butterfly
// 8.3, Classic 60-40 11.4, US large caps 17.2, each over its own 1970-2025
// window, allocations identical to the four below down to the intermediate
// Treasury leg of the 60/40). An annualized standard deviation of MONTHLY
// returns is a different statistic on the same portfolio, and a lower one here
// by one to two points, monthly noise in real returns being partly washed out
// over a calendar year. The control used to compute the monthly one and compare
// it to these annual publications: not a like-for-like reading, and it only
// looked green because the bond leg was smoothed. Rebuilding the Treasury
// references month-end (a first-of-month stamp on month-average yields had put
// the bond leg half a month out of step with month-end equities, so its
// correlation to equities read too low) raised the monthly measure of the 60/40
// from 9.8 to 10.2 and left the gap with the 9.4 that stood here impossible to
// read. Measured the way the publisher measures it, the same rebuild moves the
// 60/40 by 0.01 point, and all four portfolios land within half a point of
// their published volatility. So the anchors are the published numbers, put
// back untouched, and ttStats.vol is the publisher's statistic.
//
// The return column is the book's own table in this article (4.4 / 5.8 / 5.6
// real), except the all-equity line, which the article does not print. The
// site's headline return is NOT comparable: it publishes a 15-year baseline,
// not the window's compound return. Its drawdowns are not comparable either,
// being read on year-end data where the plate reads the monthly index, which is
// why the deeper numbers below are the article's own.
func TestCurseurControlTableStillHolds(t *testing.T) {
	legs, cpi := curseurLegs(t)
	for _, c := range []struct {
		name       string
		weights    map[string]float64
		cagr, vol  float64
		cagrTol    float64
		volTol     float64
		drawdown   float64
		drawdnTol  float64
		hasDrawdwn bool
	}{
		{name: "Browne 4 × 25", cagr: 4.4, vol: 7.6, cagrTol: 0.25, volTol: 0.5,
			drawdown: -22, drawdnTol: 1, hasDrawdwn: true,
			weights: map[string]float64{"equities": .25, "long": .25, "gold": .25, "cash": .25}},
		{name: "Golden Butterfly", cagr: 5.8, vol: 8.3, cagrTol: 0.25, volTol: 0.5,
			drawdown: -22, drawdnTol: 1, hasDrawdwn: true,
			weights: map[string]float64{"equities": .20, "smallvalue": .20, "long": .20, "short": .20, "gold": .20}},
		{name: "60/40", cagr: 5.6, vol: 11.4, cagrTol: 0.25, volTol: 0.5,
			weights: map[string]float64{"equities": .60, "intermediate": .40}},
		{name: "100 % actions", cagr: 6.8, vol: 17.2, cagrTol: 0.25, volTol: 0.5,
			weights: map[string]float64{"equities": 1}},
	} {
		got := ttRecompute(t, legs, c.weights, cpi)
		if math.Abs(got.cagr-c.cagr) > c.cagrTol {
			t.Errorf("%s: real return %.2f %%, the control table says %.1f", c.name, got.cagr, c.cagr)
		}
		if math.Abs(got.vol-c.vol) > c.volTol {
			t.Errorf("%s: real volatility %.2f %%, the control table says %.1f", c.name, got.vol, c.vol)
		}
		if c.hasDrawdwn && math.Abs(got.drawdown-c.drawdown) > c.drawdnTol {
			t.Errorf("%s: worst real drawdown %.1f %%, the article says %.0f", c.name, got.drawdown, c.drawdown)
		}
	}
}

// The sweep itself, rebuilt: this is what "make figure-drift" reports when a
// data refresh moves the curves.
func TestCurseurSweepMatchesTheData(t *testing.T) {
	legs, cpi := curseurLegs(t)
	for _, s := range curseurSweep {
		got := ttRecompute(t, legs, curseurWeights(s.dose), cpi)
		if math.Abs(got.cagr-s.cagr) > 0.05 {
			t.Errorf("dose %.0f %%: real return %.2f %%, the plate draws %.2f", s.dose, got.cagr, s.cagr)
		}
		if math.Abs(got.drawdown-s.drawdown) > 0.3 {
			t.Errorf("dose %.0f %%: worst drawdown %.1f %%, the plate draws %.1f",
				s.dose, got.drawdown, s.drawdown)
		}
		if math.Abs(got.worst-s.worst) > 0.3 {
			t.Errorf("dose %.0f %%: worst year %.1f %%, the plate says %.1f", s.dose, got.worst, s.worst)
		}
		if got.worstYear != s.worstYear {
			t.Errorf("dose %.0f %%: the worst year is %d, the plate says %d",
				s.dose, got.worstYear, s.worstYear)
		}
	}
	// The naked core is the article's own 70/30, which the family plate of the
	// same article already measures: the two plates must agree on it.
	ladder := tousTempsLadder[7]
	if ladder.name != "70/30" {
		t.Fatalf("the ladder's eighth step is %q, not the 70/30 the slider starts from", ladder.name)
	}
	if math.Abs(curseurSweep[0].cagr-ladder.cagr) > 0.02 ||
		math.Abs(curseurSweep[0].drawdown-ladder.drawdn) > 0.1 {
		t.Errorf("the slider starts at %.2f / %.1f, the family plate's 70/30 is at %.2f / %.1f",
			curseurSweep[0].cagr, curseurSweep[0].drawdown, ladder.cagr, ladder.drawdn)
	}
}

// The note's sentence about bad years, rebuilt: the pocket empties the 1974
// kind of year, where equities and inflation hit together, and does nothing at
// all for 2022, where gold and long duration fell with equities. Which is why
// the worst year stops improving as soon as 2022 inherits the title, and why
// the plate says so rather than reading the slide as a general retreat.
func TestCurseurWorstYearsAreTwoDifferentAnimals(t *testing.T) {
	legs, cpi := curseurLegs(t)
	year := func(dose float64, y int) float64 {
		real := ttRealIndex(legs, curseurWeights(dose), cpi)
		return ttCalendarYears(real)[y-ttFirst-1] * 100
	}
	if got := curseurAt(0).worstYear; got != 1974 {
		t.Errorf("the naked core's worst year is %d, the note names 1974", got)
	}
	for _, s := range curseurSweep[1:] {
		if s.worstYear != 2022 {
			t.Errorf("dose %.0f %%: the worst year is %d, the note names 2022", s.dose, s.worstYear)
		}
	}
	if bought := year(40, 1974) - year(0, 1974); bought < 15 {
		t.Errorf("the pocket only lifts 1974 by %.1f points, the note says it empties that year", bought)
	}
	if moved := math.Abs(year(40, 2022) - year(0, 2022)); moved > 1 {
		t.Errorf("the pocket moves 2022 by %.1f point, the note says it changes nothing there", moved)
	}
}

// The shape the plate exists to show, checked on the frozen numbers alone: the
// return falls slowly and monotonically, the worst path improves fast and then
// stops improving, and the flattening happens inside the band the article
// recommends. If the plateau were somewhere else, the marking would be an
// opinion rather than a reading.
func TestCurseurPlateauIsWhereTheArticleSaysItIs(t *testing.T) {
	for i := 1; i < len(curseurSweep); i++ {
		if curseurSweep[i].cagr >= curseurSweep[i-1].cagr {
			t.Errorf("dose %.0f %%: the return no longer falls with the dose", curseurSweep[i].dose)
		}
		if curseurSweep[i].drawdown < curseurSweep[i-1].drawdown {
			t.Errorf("dose %.0f %%: the worst path no longer improves with the dose", curseurSweep[i].dose)
		}
	}
	// Almost all of the tail is bought before the plateau, and the last ten
	// points of dose buy almost nothing.
	before := curseurBought(curseurPlateauLo)
	last := curseurAt(curseurPlateauHi).drawdown - curseurAt(curseurPlateauLo).drawdown
	if before < 10 {
		t.Errorf("the first thirty points of dose only buy %.1f points of worst path", before)
	}
	if last > before/10 {
		t.Errorf("the plateau still buys %.2f points of worst path, against %.1f before it: it is not a plateau",
			last, before)
	}
	// And the whole slider costs less than the article's upper bound. The
	// measured cost lands at or below the LOW edge of its 0.3 to 0.6 band,
	// which the legend states rather than rounding away.
	if c := curseurCost(curseurPlateauHi); c <= 0 || c > 0.6 {
		t.Errorf("forty points of dose cost %.2f point, the article says 0.3 to 0.6", c)
	}
	if curseurCost(curseurPlateauLo) >= curseurCost(curseurPlateauHi) {
		t.Error("the dose stopped costing anything between 30 and 40 %")
	}
}

// The plate against the article that carries it.
func TestCurseurAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "portefeuilles-tous-temps")
	for _, want := range []string{
		"::: figure tous-temps-curseur",
		"un cœur de 60-70 % de croissance mondiale, plus 30-40 % de poche de régimes (or, linkers, duration longue et éventuellement trend)",
		"pour un coût d'espérance de 0,3-0,6 point",
		"La conception moderne traite le tous-temps comme un **curseur**",
		"Variante A : un 70/30 classique en obligations intermédiaires.",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The plate's core and pocket are the article's own: its variant A for the
	// core, and its worked example's pocket (long duration and gold, its
	// linkers leg renormalized away) for the pocket.
	if curseurCoreEquities != 0.70 || curseurCoreBonds != 0.30 {
		t.Errorf("the core is %.0f/%.0f, the article's variant A is 70/30",
			curseurCoreEquities*100, curseurCoreBonds*100)
	}
	if curseurPocketGold+curseurPocketLong != 1 {
		t.Errorf("the pocket's legs sum to %.2f", curseurPocketGold+curseurPocketLong)
	}
	// The recommended band is the article's 30 to 40 %.
	if curseurPlateauLo != 30 || curseurPlateauHi != 40 {
		t.Errorf("the plate marks %.0f-%.0f %%, the article recommends 30-40",
			curseurPlateauLo, curseurPlateauHi)
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestCurseurPlateRenders(t *testing.T) {
	svg := FigureSVG("tous-temps-curseur")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"\u2014", "\u2013", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, s := range curseurSweep {
		if want := ">" + frNum(s.cagr, 2) + "<"; !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw the return %q", want)
		}
		if want := ">" + frMinus(s.drawdown, 0) + "<"; !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw the worst path %q", want)
		}
		if want := fmt.Sprintf(">%.0f %%<", s.dose); !strings.Contains(svg, want) {
			t.Errorf("the plate does not label the dose %q", want)
		}
	}
	for _, want := range []string{
		">plateau recommandé<",
		">rendement réel annualisé<",
		">pire recul réel<",
		"Quarante points de poche retirent 16 points de recul et coûtent 0,3 point de rendement.",
		"aucune série de linkers",
		"Pire année civile : −26,3 % en 1974 sans poche, mais dès la première dose c'est 2022 qui prend la place",
		"La poche efface le choc de 1974 et ne change rien à 2022",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// Two curves, one dot per point on each, and the plateau band.
	if n := strings.Count(svg, "<polyline"); n != 2 {
		t.Errorf("%d curves drawn, the plate has exactly two", n)
	}
	if n := strings.Count(svg, "<circle"); n != 2*len(curseurSweep) {
		t.Errorf("%d dots drawn, expected two per dose", n)
	}
	if n := strings.Count(svg, "<rect"); n != 1+2 {
		t.Errorf("%d rectangles drawn, expected the plateau band and the two legend chips", n)
	}
}
