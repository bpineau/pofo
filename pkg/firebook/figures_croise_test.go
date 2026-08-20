package firebook

import (
	"math"
	"strings"
	"testing"

	"github.com/bpineau/pofo/pkg/decumul"
)

// croisePlan is the plan both curves judge, on the broad-sample side.
func croisePlan(t *testing.T, rate float64) decumul.Plan {
	t.Helper()
	return decumul.Plan{
		Capital: 1e6, NeedAnnual: 1e6 * rate / 100, Years: croiseYears,
		Source: broadPool(t, croiseYears),
	}
}

// The two curves and the two safe rates, rebuilt from the bundled panel and the
// book's own broad-sample machinery.
func TestCroiseCurvesMatchTheData(t *testing.T) {
	us := jstUSA(t)
	for _, s := range croiseGrid {
		ok, _ := jstSuccess(us, croiseEquity, s.rate/100, croiseYears)
		if math.Abs((1-ok)*100-s.us) > 0.1 {
			t.Errorf("US at %.2f %%: ruin %.2f %%, the plate draws %.2f", s.rate, (1-ok)*100, s.us)
		}
		ruin := croisePlan(t, s.rate).Simulate(croisePaths, croiseWorkers, croiseSeed).RuinProb()
		if math.Abs(ruin*100-s.broad) > 0.3 {
			t.Errorf("broad at %.2f %%: ruin %.2f %%, the plate draws %.2f", s.rate, ruin*100, s.broad)
		}
	}
	safe := croisePlan(t, 4).Solve(croiseTarget/100,
		decumul.WithdrawalAxis(0, 150000), croisePaths, croiseWorkers, croiseSeed)
	if math.Abs(safe/1e4-croiseSafeBroad) > 0.05 {
		t.Errorf("the broad sample's safe rate is %.3f %%, the plate marks %.2f", safe/1e4, croiseSafeBroad)
	}
}

// The anchor the ARTICLE states for this very panel: on it, the United States
// hold 3.75 % and no more. It is the tightest check the plate has, because it
// pins the extraction, the withdrawal convention and the window count at once.
func TestCroiseReproducesTheArticlesAmericanAnchor(t *testing.T) {
	us := jstUSA(t)
	if ok, _ := jstSuccess(us, croiseEquity, 0.0375, croiseYears); ok != 1 {
		t.Errorf("at 3.75 %% the American record already fails (%.2f %% succeed), the article says it holds",
			ok*100)
	}
	if ok, _ := jstSuccess(us, croiseEquity, 0.0380, croiseYears); ok == 1 {
		t.Error("at 3.80 % the American record still never fails: the article's 3,75 % would not be its limit")
	}
}

// The shape the plate exists to show: one plan, two verdicts, and a gap that
// opens exactly in the band where plans are sized.
func TestCroiseGapOpensInTheSizingBand(t *testing.T) {
	for i := 1; i < len(croiseGrid); i++ {
		if croiseGrid[i].us < croiseGrid[i-1].us || croiseGrid[i].broad < croiseGrid[i-1].broad {
			t.Errorf("at %.2f %% a curve stopped rising with the withdrawal", croiseGrid[i].rate)
		}
	}
	// At the rate everyone quotes, the developed century is several times
	// harsher; that ratio is the article's headline.
	if r := croiseRatio(); r < 4 {
		t.Errorf("at 4 %% the developed century is only %.1f times harsher", r)
	}
	// The gap widens across the sizing band.
	lo := croiseAt(croiseBandLo).broad - croiseAt(croiseBandLo).us
	hi := croiseAt(croiseBandHi).broad - croiseAt(croiseBandHi).us
	if hi <= lo {
		t.Errorf("the gap goes from %.1f to %.1f points across the band: it no longer opens", lo, hi)
	}
	// And the two safe rates sit either side of it, which is why the band is
	// where the argument lives.
	if croiseSafeBroad >= croiseBandLo || croiseSafeUS <= croiseBandLo || croiseSafeUS >= 5 {
		t.Errorf("the safe rates (%.2f, %.2f) no longer straddle the sizing band",
			croiseSafeBroad, croiseSafeUS)
	}
}

// The plate against the article that carries it.
func TestCroiseAgreesWithItsArticle(t *testing.T) {
	art := bookArticle(t, "anarkulova-cederburg")
	for _, want := range []string{
		"::: figure echantillon-croise",
		"les États-Unis ne tiennent que 3,75 %",
		"~17 % d'échec à 4 % ; 2,26 % à 5 % d'échec",
		"ces trois nombres ne mesurent pas la même chose",
		"Le modèle broad-sample | 16 pays (panel JST, 1870-2020), blocs d'un même pays",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("the article no longer says %q", want)
		}
	}
}

// The rendered plate obeys the book's drawing rules and carries every reading
// it claims to draw, including the paper's own published numbers, which it
// prints rather than pretends to reproduce.
func TestCroisePlateRenders(t *testing.T) {
	svg := FigureSVG("echantillon-croise")
	if !strings.HasPrefix(svg, "<svg viewBox=") {
		t.Fatal("the plate must render an SVG")
	}
	// U+2014 and U+2013 are the em- and en-dashes, banned from the book down to
	// its figure labels; they stay escaped here so a repository-wide grep never
	// trips on them.
	for _, banned := range []string{"—", "–", "rgba(", "opacity", "rotate("} {
		if strings.Contains(svg, banned) {
			t.Errorf("the plate uses %q, which the book's figures never do", banned)
		}
	}
	for _, want := range []string{
		">le siècle développé<",
		">le siècle américain<",
		">là où tout le monde se dimensionne<",
		">5 % d'échec<",
		">sûr ici : 1,72 %<",
		">sûr ici : 4,07 %<",
		"À 4 %, le même plan échoue 3 fois sur cent sur le siècle américain, et 23 fois sur cent",
		"Le panel donne aux États-Unis 3,75 % sans aucun échec",
		"Le papier de 2023 publie 17 % d'échec à 4 % et 2,26 % à 5 % d'échec",
		"ces nombres ne mesurent pas la même chose",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not carry %q", want)
		}
	}
	if n := strings.Count(svg, "<polyline"); n != 3 {
		t.Errorf("%d lines drawn, expected the two curves and the failure rule", n)
	}
	if n := strings.Count(svg, "<circle"); n != 2 {
		t.Errorf("%d dots drawn, expected one per safe rate", n)
	}
}
