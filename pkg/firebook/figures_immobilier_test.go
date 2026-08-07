package firebook

import (
	"math"
	"strings"
	"testing"
)

// round1 rounds a percentage the way the plate prints it.
func immoRound1(v float64) float64 { return math.Round(v*10) / 10 }

// The cascade must add up on its own terms: the marches drawn are exactly what
// separates the gross yield from the pre-tax rent, and the tax step exactly
// what separates the pre-tax rent from the bare-letting net-net.
func TestImmobilierCascadeAddsUp(t *testing.T) {
	cuts := immoCoproPNO + immoTaxeFonciere + immoEntretien + immoVacancy() + immoGestion()
	if got := immoGross() - cuts; math.Abs(got-immoPreTax()) > 1e-9 {
		t.Errorf("the five marches leave %.2f €, immoPreTax says %.2f €", got, immoPreTax())
	}
	if got := immoPreTax() * immoTaxRateNu(); math.Abs(immoPreTax()-got-immoNetNu()) > 1e-9 {
		t.Errorf("the tax step of %.2f € does not land on the net-net %.2f €", got, immoNetNu())
	}
	// The bite on bare rental income: the marginal bracket plus the social
	// levies the LFSS 2026 left at 17,2 % for that income, less what the
	// deductible CSG hands back at the barème.
	if want := 0.4516; math.Abs(immoTaxRateNu()-want) > 1e-12 {
		t.Errorf("effective tax rate %.4f, expected %.4f", immoTaxRateNu(), want)
	}
}

// The plate must land where the article says the reality is: an advert yield
// inside 5-7 %, and both net-net outcomes inside the 2-4 % band the article
// announces as typical, the fork between regimes covering that band.
func TestImmobilierCascadeLandsInTheArticleRange(t *testing.T) {
	gross := immoRound1(immoYield(immoGross()))
	if gross < 5 || gross > 7 {
		t.Errorf("gross yield %.1f %% is outside the article's 5-7 %% advert range", gross)
	}
	nu := immoRound1(immoYield(immoNetNu()))
	lmnp := immoRound1(immoYield(immoPreTax()))
	for _, tc := range []struct {
		name string
		v    float64
	}{{"foncier nu", nu}, {"LMNP au réel", lmnp}} {
		if tc.v < 2 || tc.v > 4 {
			t.Errorf("net-net %s: %.1f %%, outside the article's 2-4 %% band", tc.name, tc.v)
		}
	}
	if nu >= lmnp {
		t.Errorf("the two regimes do not fork: nu %.1f %%, LMNP %.1f %%", nu, lmnp)
	}
}

// Figure and prose must agree. Every marche the plate draws is a hypothesis the
// article states in words, so the reader can audit it and substitute their own;
// this fails the moment one of those numbers leaves the text.
func TestImmobilierPlateAgreesWithTheArticle(t *testing.T) {
	raw, err := assets.ReadFile("assets/book/fr/immobilier-en-retrait.md")
	if err != nil {
		t.Fatal(err)
	}
	article := string(raw)
	for _, want := range []string{
		"::: figure immobilier-net-net",
		"**Le chiffre honnête.**",
		euroFR(immoPrice),                  // the flat
		euroFR(immoRentMonthly),            // its monthly rent
		euroFR(immoGross()),                // the yearly rent
		euroFR(immoCoproPNO),               // marche 1
		euroFR(immoTaxeFonciere),           // marche 2
		euroFR(immoEntretien),              // marche 3
		euroFR(immoVacancy()),              // marche 4
		euroFR(immoGestion()),              // marche 5
		euroFR(immoPreTax()),               // the fork
		euroFR(immoPreTax() - immoNetNu()), // the tax step of the bare regime
		frNum(immoYield(immoGross()), 1) + " %",
		frNum(immoYield(immoPreTax()), 1) + " %",
		frNum(immoYield(immoNetNu()), 1) + " %",
		frNum(immoTMI*100, 0) + " %",         // the marginal bracket
		frNum(immoPSFoncierNu*100, 1) + " %", // social levies on bare rental income
		"18,6 %",                             // and the higher rate furnished letting took
	} {
		if !strings.Contains(article, want) {
			t.Errorf("the article no longer carries %q, which the plate draws", want)
		}
	}
}

func TestImmobilierPlateRenders(t *testing.T) {
	svg := FigureSVG("immobilier-net-net")
	if svg == "" {
		t.Fatal("the plate is not registered")
	}
	for _, want := range []string{
		"Du brut d'annonce au net-net : la cascade",
		"loyer net avant impôt",
		"récupérables", "foncière", "et impayés", "LMNP au réel", "aucun impôt",
		euroFR(immoGestion()), euroFR(immoPreTax() - immoNetNu()),
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("the plate does not draw %q", want)
		}
	}
	if strings.Contains(svg, "—") { // the em-dash, banned book-wide
		t.Error("no em-dash in a figure")
	}
	if strings.Contains(svg, "rgba(") || strings.Contains(svg, "opacity") {
		t.Error("no rgba and no opacity: crengine paints them solid black")
	}
}
