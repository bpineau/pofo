package firebook

import (
	"math"
	"strings"
	"testing"
)

// voyantReplay recomputes one vintage's gauge from the bundled record: a
// constant real withdrawal taken at the start of each year, the rest compounded
// at the year's real 60/40 return, and the reading is that withdrawal over the
// capital it is taken from. A zero means the capital could no longer pay it.
func voyantReplay(s jstSeries, start int) [voyantYears]float64 {
	var out [voyantYears]float64
	off := start - s.first
	w := 1.0
	spend := voyantRule / 100
	for k := 0; k < voyantYears; k++ {
		if w <= 0 {
			continue
		}
		out[k] = spend / w * 100
		w -= spend
		if w <= 0 {
			w = 0
			continue
		}
		w *= 1 + voyantEquity*s.equity[off+k] + (1-voyantEquity)*s.bond[off+k]
	}
	return out
}

// Both curves, rebuilt from the bundled panel: this is what "make figure-drift"
// reports when the record is regenerated.
func TestVoyantsCurvesMatchTheData(t *testing.T) {
	us := jstUSA(t)
	for _, v := range voyantVintages {
		got := voyantReplay(us, v.start)
		for k := range got {
			if math.Abs(got[k]-v.rate[k]) > 0.05 {
				t.Errorf("vintage %d, year %d: the gauge reads %.2f %%, the plate draws %.2f",
					v.start, v.start+k, got[k], v.rate[k])
			}
		}
	}
	// The readings the plate poses by name.
	crash, crashYear := voyantVintages[0].crashPeak()
	if crashYear != 1932 || math.Abs(crash-6.30) > 0.05 {
		t.Errorf("the 1929 crash peaks at %.2f %% in %d, the plate says 6,3 in 1932", crash, crashYear)
	}
	top, topYear := voyantVintages[0].peak()
	if topYear != 1949 || math.Abs(top-6.53) > 0.05 {
		t.Errorf("the 1929 vintage's maximum is %.2f %% in %d, the plate says 6,5 in 1949", top, topYear)
	}
	if got := voyantVintages[1].crossing(); got != 1975 {
		t.Errorf("the 1966 vintage enters the band in %d, the plate says 1975", got)
	}
	if got := voyantVintages[1].ruinYear(); got != 1991 {
		t.Errorf("the 1966 vintage empties in %d, the plate says 1991", got)
	}
}

// The contrast the plate exists to show, and the article's own sentence: the
// crash gives back, the episode never does. If the 1929 gauge did not come
// home, the plate would be arbitrating its own thesis and would have no right
// to be drawn.
func TestVoyantsContrastHolds(t *testing.T) {
	crashV, episodeV := voyantVintages[0], voyantVintages[1]
	// The crash makes the gauge ring, then hands it back.
	crash, _ := crashV.crashPeak()
	if crash <= crashV.rate[0] {
		t.Errorf("the 1929 gauge never rose above its opening %.2f %%", crashV.rate[0])
	}
	if !crashV.returns() {
		t.Error("the 1929 gauge does not come back under the band: the plate's thesis fails")
	}
	if crashV.crossing() != 0 {
		t.Errorf("the 1929 gauge entered the irrecoverable band in %d", crashV.crossing())
	}
	if last := crashV.rate[voyantYears-1]; last > crashV.rate[0]+1 {
		t.Errorf("the 1929 gauge ends at %.2f %% against %.2f at the start: it did not give back",
			last, crashV.rate[0])
	}
	// The episode makes it ring and keeps it.
	if episodeV.crossing() == 0 {
		t.Error("the 1966 gauge never enters the band")
	}
	if episodeV.returns() {
		t.Error("the 1966 gauge comes home: the contrast the plate draws is gone")
	}
	if episodeV.ruinYear() == 0 {
		t.Error("the 1966 vintage survives its horizon")
	}
	// And the episode is the worse of the two, which is the article's ranking.
	ep, _ := episodeV.peak()
	cr, _ := crashV.peak()
	if ep <= cr {
		t.Errorf("the 1966 gauge peaks at %.2f %% against %.2f for 1929: 1966 is no longer the worse",
			ep, cr)
	}
}

// The plate against the article that carries it.
func TestVoyantsAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "inflation-et-taux-de-retrait")
	for _, want := range []string{
		"::: figure voyants-1929-1966",
		"Le krach de 1929 est brutal, puis il rend",
		"la déflation fait **baisser** les retraits en nominal avec les prix",
		"L'épisode 1966-1981, lui, ne rend jamais",
		"C'est la situation où le portefeuille 60/40 n'a aucune poche qui gagne.",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
	// The plate runs the allocation the article names for the episode.
	if voyantEquity != 0.60 {
		t.Errorf("the plate runs a %.0f/%.0f, the article names the 60/40",
			voyantEquity*100, (1-voyantEquity)*100)
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw.
func TestVoyantsPlateRenders(t *testing.T) {
	svg := FigureSVG("voyants-1929-1966")
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
	for _, v := range voyantVintages {
		if !strings.Contains(svg, ">"+v.name+"<") {
			t.Errorf("the plate does not name %q", v.name)
		}
	}
	for _, want := range []string{
		">le krach : 6,3 % en 1932<",
		">son vrai maximum : 6,5 % en 1949<",
		">franchit 8 % en 1975<",
		">sort de l'échelle en 1988, compte vide en 1991<",
		">revenu à 4,8 % trente ans plus tard<",
		">la zone dont on ne revient pas<",
		"Le krach de 1929 fait sonner le voyant trois ans puis le rend",
		"la déflation des années trente n'apparaît donc pas au numérateur",
		"ne vient pas du krach mais de l'inflation d'après-guerre",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	// Two curves, four posed readings, one band.
	if n := strings.Count(svg, "<polyline"); n != 2 {
		t.Errorf("%d curves drawn, the plate has exactly two", n)
	}
	if n := strings.Count(svg, "<circle"); n != 3 {
		t.Errorf("%d dots drawn, expected the three posed readings", n)
	}
	if n := strings.Count(svg, "<rect"); n != 1+2 {
		t.Errorf("%d rectangles drawn, expected the band and the two legend chips", n)
	}
}
