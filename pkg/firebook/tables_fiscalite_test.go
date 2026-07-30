package firebook

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
)

// Three reference tables of the French tax chapters carry rates and thresholds
// that go stale on their own. These guards recompute every cell from dated
// constants and from the engines the book already owns, so a table can never
// drift away from the law, from the article around it, or from the plate that
// shares its subject.
//
// Every constant below was verified on 2026-07-30 against a primary or
// near-primary source, never quoted from memory:
//
//   - LFSS 2026 (loi n° 2025-1403 du 30 décembre 2025, art. 12) splits the
//     social levies in two, 18,6 % and 17,2 %. The rates themselves are the
//     dated constants of figures_enveloppes.go, reused here rather than
//     restated.
//   - Art. 669 CGI, unchanged since the loi de finances 2004: the usufruct is
//     worth 90 % below 21 and loses ten points per decade of the donor's age.
//   - Micro-BNC: 34 % abattement; contributions at 25,6 % of turnover for a
//     liberal activity outside the CIPAV from 1 January 2026 (décret
//     n° 2025-943 du 8 septembre 2025, which lowered the planned 26,1 %).
//   - SMIC horaire 12,02 € from 1 January 2026, and 150 hours of it buy one
//     retirement quarter, so 1 803 € of retained income per quarter.
const (
	microAbattementBNC = 0.34
	microCotisBNC2026  = 0.256
	smicHoraire2026    = 12.02
	trimestreHeures    = 150.0
	// The realized capital income the CSM column of retour-au-travail is drawn
	// for: the headline example of taxe-puma, worth ~2 340 € of cotisation with
	// no activity at all.
	csmCapitalRef = 60000.0
)

// bookTable returns the rows of the first pipe table of an article whose header
// line starts with the given first cell. Cells come back trimmed, separator and
// header excluded.
func fiscTable(t *testing.T, slug, firstHeader string) [][]string {
	t.Helper()
	raw, err := assets.ReadFile("assets/book/fr/" + slug + ".md")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	head := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "| "+firstHeader+" |") {
			head = i
			break
		}
	}
	if head < 0 {
		t.Fatalf("%s: no table whose first header is %q", slug, firstHeader)
	}
	var rows [][]string
	for i := head + 2; i < len(lines) && strings.HasPrefix(lines[i], "|"); i++ {
		cells := strings.Split(strings.Trim(lines[i], "|"), "|")
		for j := range cells {
			cells[j] = strings.TrimSpace(cells[j])
		}
		rows = append(rows, cells)
	}
	if len(rows) == 0 {
		t.Fatalf("%s: the %q table has no row", slug, firstHeader)
	}
	return rows
}

// The map of the two social-levy rates. Every rate cell must be one of the two
// dated constants the plates already use, and the calendar column must follow
// the split the LFSS 2026 actually draws: the revenus du patrimoine of art.
// L. 136-6 CSS are hit from the 2025 income year, the produits de placement of
// art. L. 136-7 CSS only from the 2026 cash-ins.
func TestTableauPrelevementsSociaux(t *testing.T) {
	const (
		patrimoine = "revenus 2025"
		placement  = "encaissements 2026"
		exclus     = "hausse écartée"
	)
	haut, bas := pctFR(socialLevies2026), pctFR(socialLeviesAV2026)
	if haut != "18,6 %" || bas != "17,2 %" {
		t.Fatalf("les taux datés valent %s et %s, la LFSS 2026 dit 18,6 %% et 17,2 %%", haut, bas)
	}

	want := map[string]struct{ taux, quand string }{
		"Plus-values mobilières (CTO)":    {haut, patrimoine},
		"Crypto":                          {haut, patrimoine},
		"Location meublée":                {haut, patrimoine},
		"Dividendes et intérêts":          {haut, placement},
		"Gains de PEA à la sortie":        {haut, placement},
		"Assurance-vie et capitalisation": {bas, exclus},
		"Foncier nu et SCPI":              {bas, exclus},
		"Plus-values immobilières":        {bas, exclus},
		"PEL, CEL et PEP anciens":         {bas, exclus},
	}
	rows := fiscTable(t, "flat-tax-et-imposition", "Support")
	if len(rows) != len(want) {
		t.Fatalf("%d lignes dans la carte des PS, %d attendues", len(rows), len(want))
	}
	for _, r := range rows {
		if len(r) != 3 {
			t.Fatalf("ligne %v: %d colonnes, la carte en a 3", r, len(r))
		}
		w, ok := want[r[0]]
		if !ok {
			t.Errorf("support %q hors de la carte vérifiée", r[0])
			continue
		}
		if r[1] != w.taux {
			t.Errorf("%s: PS %s, la LFSS 2026 dit %s", r[0], r[1], w.taux)
		}
		if r[2] != w.quand {
			t.Errorf("%s: application %q, attendu %q", r[0], r[2], w.quand)
		}
		// A rate that stayed at 17,2 % is a rate the rise never reached, and a
		// rate that rose has a date. The two columns cannot contradict.
		if (r[1] == bas) != (r[2] == exclus) {
			t.Errorf("%s: taux %s et calendrier %q ne peuvent pas aller ensemble", r[0], r[1], r[2])
		}
	}

	article := bookArticle(t, "flat-tax-et-imposition")
	// The exempt livrets belong to the note under the table, never to a row.
	if !strings.Contains(article, "Les livrets défiscalisés (A, LDDS, LEP) restent hors de tout cela.") {
		t.Error("la note des livrets exonérés a disparu sous le tableau")
	}
	// The PFU the article quotes is the sum of its two dated legs.
	if got := pctFR(pfuIncomeTax + socialLevies2026); !strings.Contains(article, got) {
		t.Errorf("l'article ne dit plus %s, la somme des deux jambes du PFU", got)
	}
	// The prose must no longer re-enumerate what the table carries.
	for _, gone := range []string{
		"Le taux de **17,2 %** survit pour",
		"La hausse frappe les plus-values mobilières",
	} {
		if strings.Contains(article, gone) {
			t.Errorf("la prose répète le tableau: %q", gone)
		}
	}
}

// The usufruct scale of art. 669 CGI: ten points of usufruct per decade, the
// bare ownership as the complement, and the full-property value that one
// 100 000 € allowance lets through.
func TestTableauBaremeUsufruit(t *testing.T) {
	const abattement = 100000.0
	rows := fiscTable(t, "succession-et-transmission", "Âge du donateur")
	if len(rows) != 5 {
		t.Fatalf("%d lignes de barème, le tableau en pose 5", len(rows))
	}
	for _, r := range rows {
		if len(r) != 4 {
			t.Fatalf("ligne %v: %d colonnes, le barème en a 4", r, len(r))
		}
		var lo, hi int
		if _, err := fmt.Sscanf(r[0], "%d à %d ans", &lo, &hi); err != nil {
			t.Errorf("tranche %q illisible", r[0])
			continue
		}
		if hi != lo+9 {
			t.Errorf("tranche %q: l'art. 669 CGI procède par décennies", r[0])
		}
		// Art. 669 CGI in closed form: "moins de (hi+1) ans révolus" is worth
		// 110 − (hi+1) + 1 = 110 − hi points of usufruct.
		usufruit := 110 - hi
		nue := 100 - usufruit
		if got, want := r[1], fmt.Sprintf("%d %%", usufruit); got != want {
			t.Errorf("%s: usufruit %s, l'art. 669 CGI dit %s", r[0], got, want)
		}
		if got, want := r[2], fmt.Sprintf("%d %%", nue); got != want {
			t.Errorf("%s: nue-propriété %s, attendu %s", r[0], got, want)
		}
		passe := math.Round(abattement/(float64(nue)/100)/100) * 100
		if got, want := r[3], frSpace(passe)+" €"; got != want {
			t.Errorf("%s: %s passe sous l'abattement, le calcul donne %s", r[0], got, want)
		}
	}
	// The scale must be a strictly decreasing staircase, oldest last.
	for i := 1; i < len(rows); i++ {
		if rows[i][1] >= rows[i-1][1] {
			t.Errorf("le barème remonte entre %q et %q", rows[i-1][0], rows[i][0])
		}
	}

	article := bookArticle(t, "succession-et-transmission")
	// The couple's four allowances, read through the 61-70 bracket. The article
	// states 400 000 € of allowances elsewhere, and this is that number
	// demembered, so the two cannot drift apart.
	quatre := succParents * succEnfants * succAbattementDirect
	if quatre != 400000 {
		t.Fatalf("les quatre abattements valent %.0f €, l'article dit 400 000 €", quatre)
	}
	demembre := math.Round(quatre/0.6/100) * 100
	if want := frSpace(demembre) + " €"; !strings.Contains(article, want) {
		t.Errorf("l'article ne dit pas %s de pleine propriété au barème de 61 à 70 ans", want)
	}
	// The prose must no longer sample the scale at a single age.
	for _, gone := range []string{
		"Selon le barème fiscal de l'usufruit, entre 61 et 70 ans l'usufruit vaut 40 %",
		"On transmet donc 100 en n'étant taxé que sur 60.",
	} {
		if strings.Contains(article, gone) {
			t.Errorf("la prose répète le barème: %q", gone)
		}
	}
}

// The micro-entreprise threshold grid. It exists to kill one belief, that the
// PUMa switch and the four quarters are bought with the invoiced turnover: both
// thresholds are read AFTER the 34 % abattement. Every cell is recomputed here,
// the CSM one through the very function the taxe-puma plate draws.
func TestTableauSeuilsMicroEntreprise(t *testing.T) {
	trimestre := trimestreHeures * smicHoraire2026
	if trimestre != 1803 {
		t.Fatalf("un trimestre coûte %.2f € de revenu retenu, 150 h de SMIC en valent 1 803", trimestre)
	}

	rows := fiscTable(t, "retour-au-travail", "CA facturé")
	if len(rows) != 4 {
		t.Fatalf("%d lignes de seuils, le tableau en pose 4", len(rows))
	}
	for _, r := range rows {
		if len(r) != 5 {
			t.Fatalf("ligne %v: %d colonnes, la grille en a 5", r, len(r))
		}
		ca, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSuffix(r[0], " €"), " ", ""), 64)
		if err != nil {
			t.Errorf("chiffre d'affaires %q illisible", r[0])
			continue
		}
		retenu := ca * (1 - microAbattementBNC)
		trimestres := math.Min(4, math.Floor(retenu/trimestre))
		net := ca * (1 - microCotisBNC2026)

		if want := frSpace(retenu) + " €"; r[1] != want {
			t.Errorf("%s: revenu retenu %s, l'abattement de 34 %% donne %s", r[0], r[1], want)
		}
		if want := fmt.Sprintf("%.0f", trimestres); r[2] != want {
			t.Errorf("%s: %s trimestres, %.0f € de revenu retenu en valident %s", r[0], r[2], retenu, want)
		}
		if want := frSpace(csm(csmCapitalRef, retenu)) + " €"; r[3] != want {
			t.Errorf("%s: CSM %s, la formule de l'art. D. 380-1 donne %s", r[0], r[3], want)
		}
		if want := frSpace(net) + " €"; r[4] != want {
			t.Errorf("%s: net %s, 25,6 %% de cotisations laissent %s", r[0], r[4], want)
		}
	}

	// The two turnover thresholds the paragraph under the table quotes, both of
	// them the after-abattement threshold read back through the abattement.
	article := bookArticle(t, "retour-au-travail")
	for _, tc := range []struct {
		retenu float64
		what   string
	}{
		{4 * trimestre, "les quatre trimestres"},
		{csmSwitch * csmPASS, "l'extinction de la CSM"},
	} {
		seuil := math.Round(tc.retenu/(1-microAbattementBNC)/100) * 100
		if want := "environ " + frSpace(seuil) + " €"; !strings.Contains(article, want) {
			t.Errorf("%s: l'article ne dit pas %q", tc.what, want)
		}
	}

	// The lead-in states the four dated parameters; each must be the constant.
	for _, want := range []string{
		"abattement est de 34 %",
		"cotisations de 25,6 %",
		frSpace(trimestre) + " € de revenu retenu",
		"SMIC horaire de 12,02 €",
		frSpace(csmSwitch*csmPASS) + " € de revenu retenu",
	} {
		if !strings.Contains(article, want) {
			t.Errorf("le chapeau du tableau ne dit pas %q", want)
		}
	}

	// The article's own two round figures, which the table now explains: four
	// quarters from ~7,2 k€ and the PUMa switch at ~9,6 k€, both of retained
	// income and never of turnover.
	if got := 4 * trimestre; math.Abs(got-7200) > 50 {
		t.Errorf("quatre trimestres coûtent %.0f €, l'article dit ~7 200 €", got)
	}
	if got := csmSwitch * csmPASS; math.Abs(got-9600) > 50 {
		t.Errorf("le seuil PUMa vaut %.0f €, l'article dit ~9 600 €", got)
	}
	// The prose must no longer carry the single example the table replaced.
	if strings.Contains(article, "environ 15 k€ facturés en services pour 10 k€ de revenu retenu") {
		t.Error("la prose répète le tableau des seuils")
	}
}
