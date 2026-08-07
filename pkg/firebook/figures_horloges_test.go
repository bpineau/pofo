package firebook

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

// The horloges-enveloppes plate draws no market data: it draws the structure of
// the article itself, the two legal clocks and the two households it describes.
// So this test pins that structure on both sides. Every state the plate states
// is recomputed here, and the article is read back to check it still quotes the
// same maturities, the same ages and the same rates. Change one and this fails.

// envArticle returns the source of the article the plate serves.
func envArticle(t *testing.T) string {
	t.Helper()
	raw, err := assets.ReadFile("assets/book/fr/enveloppes-francaises.md")
	if err != nil {
		t.Fatalf("lecture de l'article: %v", err)
	}
	return string(raw)
}

// The maturities and the ages of the plate are the article's own.
func TestHorlogesGroundedInTheArticle(t *testing.T) {
	art := envArticle(t)
	for _, want := range []string{
		// the PEA clock: 5 years, counted from the opening date
		fmt.Sprintf("Après %d ans, les retraits sont exonérés", peaClockYears),
		fmt.Sprintf("c'est la date d'ouverture qui démarre l'horloge des %d ans, pas les versements", peaClockYears),
		// the assurance-vie clock: 8 years, same rule
		fmt.Sprintf("Après %d ans, un abattement annuel s'applique", avClockYears),
		fmt.Sprintf("pour lancer l'horloge des %d ans", avClockYears),
		fmt.Sprintf("les horloges de %d et %d ans partent de la date d'ouverture", peaClockYears, avClockYears),
		// the ages the two tracks are drawn at
		fmt.Sprintf("Prenons un couple de %d ans", envDepartureAge),
		"à deux ans du départ",
		"le rentier de 50 ans hérite des clics du trentenaire",
		// the example the plate says it illustrates
		"gainé à 45 % environ",
		"friction de 14 % environ",
		// and the block itself
		"::: figure horloges-enveloppes",
	} {
		if !strings.Contains(art, want) {
			t.Errorf("l'article ne dit plus %q", want)
		}
	}
	if envLateLead != 2 {
		t.Errorf("la voie du bas ouvre %d ans avant le départ, l'article dit deux ans", envLateLead)
	}
	if envLateOpenAge != 56 || envEarlyOpenAge != 30 {
		t.Errorf("ouvertures à %d et %d ans, attendu 30 et 56", envEarlyOpenAge, envLateOpenAge)
	}
	// The axis must contain every date the plate marks, including the two
	// maturities that fall after the departure.
	for _, l := range envLanes {
		if m := envMatureAge(envLateOpenAge, l.years); m > envAxisLastAge {
			t.Errorf("la maturité à %d ans tombe hors de l'axe (%d ans)", m, envAxisLastAge)
		}
	}
}

// The state of each clock on the departure day, the whole point of the plate.
func TestHorlogesClockStates(t *testing.T) {
	for _, tc := range []struct {
		name       string
		openAge    int
		clockYears int
		matureAge  int
		mature     bool
		years      int
	}{
		{"PEA ouvert à 30 ans", envEarlyOpenAge, peaClockYears, 35, true, 23},
		{"assurances-vie ouvertes à 30 ans", envEarlyOpenAge, avClockYears, 38, true, 20},
		{"PEA ouvert à 56 ans", envLateOpenAge, peaClockYears, 61, false, 3},
		{"assurances-vie ouvertes à 56 ans", envLateOpenAge, avClockYears, 64, false, 6},
	} {
		matureAge, mature, years := envClockState(tc.openAge, tc.clockYears)
		if matureAge != tc.matureAge || mature != tc.mature || years != tc.years {
			t.Errorf("%s: mûr à %d ans (%v), %d ans d'écart ; attendu %d ans (%v), %d",
				tc.name, matureAge, mature, years, tc.matureAge, tc.mature, tc.years)
		}
	}
	// Three taps against one: the reading the plate exists for. The CTO has no
	// clock, so it always counts.
	if got := envTapsFlowing(envEarlyOpenAge); got != len(envLanes)+1 || got != 3 {
		t.Errorf("voie du haut: %d robinets au départ, attendu 3", got)
	}
	if got := envTapsFlowing(envLateOpenAge); got != 1 {
		t.Errorf("voie du bas: %d robinets au départ, attendu 1 (le CTO seul)", got)
	}
	// The lanes carry the legal maturities, not a copy of them.
	if envLanes[0].years != peaClockYears || envLanes[1].years != avClockYears {
		t.Error("les voies ne portent plus les durées légales")
	}
}

// The rates and the allowance the verdict lines quote, each recomputed from the
// dated constants of the frictions plate (never from memory).
func TestHorlogesRatesAndAllowance(t *testing.T) {
	if got := pctFR(peaSlope); got != "18,6 %" {
		t.Errorf("PEA: %q, attendu 18,6 %%", got)
	}
	if got := pctFR(ctoSlope); got != "31,4 %" {
		t.Errorf("CTO: %q, attendu 31,4 %%", got)
	}
	// "Chaque euro retiré porte 45 % de gain taxé à 31,4 %, soit une friction de
	// 14 % environ" (organisation A of the article's example).
	friction := envGainShare * ctoSlope
	if math.Abs(friction-0.14) > 5e-3 {
		t.Errorf("friction de l'organisation A: %.4f, l'article dit 14 %% environ", friction)
	}
	if got := pctFR(friction); got != "14,1 %" {
		t.Errorf("friction affichée: %q, attendu 14,1 %%", got)
	}
	if got := euroFR(avAllowanceCouple); got != "9 200 €" {
		t.Errorf("abattement: %q, attendu 9 200 €", got)
	}
	for _, tc := range []struct {
		v    float64
		want string
	}{{0, "0 €"}, {150, "150 €"}, {4600, "4 600 €"}, {150000, "150 000 €"}} {
		if got := euroFR(tc.v); got != tc.want {
			t.Errorf("euroFR(%.0f) = %q, attendu %q", tc.v, got, tc.want)
		}
	}
}

// The plate itself: registered, saying what it must say, and safe for crengine.
func TestHorlogesEnveloppesPlate(t *testing.T) {
	svg := figHorlogesEnveloppes()
	for _, want := range []string{
		"départ à 58 ans",
		"PEA (5 ans)", "assurances-vie (8 ans)", "CTO (aucune horloge)",
		"mûr à 35 ans, et pour toujours",   // the prepared PEA
		"mûres à 38 ans, et pour toujours", // the prepared contracts
		"mûr à 61 ans : 3 ans trop tard",   // the late PEA
		"mûres à 64 ans : 6 ans trop tard", // the late contracts
		"3 robinets sur 3 coulent",         // the verdict of the top track
		"1 robinet sur 3 coule",            // the verdict of the bottom one
		"il manque 3 ans au PEA et 6 ans aux assurances-vie",
		"18,6 %", "31,4 %", "14,1 %", "9 200 €",
		"Durées légales : 5 ans pour le PEA, 8 ans pour l'assurance-vie",
		"depuis la date d'ouverture et non depuis les versements",
		"Frise illustrative de l'exemple de l'article",
		"1er janvier 2026",
		"âge du titulaire",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("la planche ne contient pas %q", want)
		}
	}
	if strings.Contains(svg, "rgba(") || strings.Contains(svg, "opacity") {
		t.Error("la planche utilise rgba ou opacity, que crengine peint en noir")
	}
	if strings.ContainsRune(svg, '—') {
		t.Error("tiret cadratin dans la planche")
	}
	if figures["horloges-enveloppes"] == nil {
		t.Error("la planche n'est pas enregistrée sous horloges-enveloppes")
	}
}
