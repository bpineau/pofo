package portfolio

import (
	"math"
	"strings"
	"testing"
)

func TestParseBasic(t *testing.T) {
	in := `
# commentaire
60   VOO    Vanguard S&P 500
40	BND  obligations US
`
	spec, err := Parse("test", strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Holdings) != 2 {
		t.Fatalf("attendu 2 lignes, trouvé %d", len(spec.Holdings))
	}
	h := spec.Holdings[0]
	if h.ID != "VOO" || math.Abs(h.Weight-0.60) > 1e-12 || h.Note != "Vanguard S&P 500" {
		t.Errorf("ligne 1 mal lue: %+v", h)
	}
	h = spec.Holdings[1]
	if h.ID != "BND" || math.Abs(h.Weight-0.40) > 1e-12 || h.Note != "obligations US" {
		t.Errorf("ligne 2 mal lue: %+v", h)
	}
	if len(spec.Warnings) != 0 {
		t.Errorf("pas de warning attendu, trouvé %v", spec.Warnings)
	}
}

func TestParseInlineComments(t *testing.T) {
	in := `
# Portefeuille de test
# https://exemple.invalid/doc

60 VOO  note utile # le S&P 500
40 BND# collé au ticker
`
	spec, err := Parse("t", strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Holdings) != 2 {
		t.Fatalf("2 lignes attendues, trouvé %d", len(spec.Holdings))
	}
	if h := spec.Holdings[0]; h.ID != "VOO" || h.Note != "note utile" {
		t.Errorf("commentaire mal retiré: %+v", h)
	}
	if h := spec.Holdings[1]; h.ID != "BND" || h.Note != "" {
		t.Errorf("commentaire collé mal retiré: %+v", h)
	}
}

func TestParseDecimalCommaAndPercent(t *testing.T) {
	spec, err := Parse("t", strings.NewReader("33,5% IWDA.AS\n66.5 IE00B4L5Y983 monde"))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(spec.Holdings[0].RawWeight-33.5) > 1e-12 {
		t.Errorf("virgule décimale: %v", spec.Holdings[0].RawWeight)
	}
	if math.Abs(spec.Holdings[1].RawWeight-66.5) > 1e-12 {
		t.Errorf("point décimal: %v", spec.Holdings[1].RawWeight)
	}
}

func TestParseNormalizesWeights(t *testing.T) {
	spec, err := Parse("t", strings.NewReader("50 A\n100 B"))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Warnings) != 1 {
		t.Fatalf("warning de normalisation attendu, trouvé %v", spec.Warnings)
	}
	if math.Abs(spec.Holdings[0].Weight-1.0/3) > 1e-12 || math.Abs(spec.Holdings[1].Weight-2.0/3) > 1e-12 {
		t.Errorf("poids non normalisés: %+v", spec.Holdings)
	}
}

func TestParseMetaRebalance(t *testing.T) {
	in := `
#meta rebalance:30   # commentaire toléré
60 VOO
40 BND
`
	spec, err := Parse("t", strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if spec.RebalanceDays != 30 {
		t.Errorf("RebalanceDays = %d, attendu 30", spec.RebalanceDays)
	}
	if spec.Meta["rebalance"] != "30" {
		t.Errorf("Meta brut: %+v", spec.Meta)
	}

	// Sans directive: -1 (le défaut de l'appelant s'applique).
	spec, err = Parse("t", strings.NewReader("100 VOO"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.RebalanceDays != -1 {
		t.Errorf("RebalanceDays sans directive = %d, attendu -1", spec.RebalanceDays)
	}

	// rebalance:0 = jamais rebalancer (distinct de non spécifié).
	spec, err = Parse("t", strings.NewReader("#meta rebalance:0"+"\n"+"100 VOO"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.RebalanceDays != 0 {
		t.Errorf("RebalanceDays = %d, attendu 0", spec.RebalanceDays)
	}

	// Clé inconnue: avertissement, pas d'erreur.
	spec, err = Parse("t", strings.NewReader("#meta fancy:yes"+"\n"+"100 VOO"))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec.Warnings) != 1 {
		t.Errorf("warning attendu pour clé inconnue: %v", spec.Warnings)
	}

	// "#metadata" n'est pas une directive, juste un commentaire.
	if _, err := Parse("t", strings.NewReader("#metadata blabla"+"\n"+"100 VOO")); err != nil {
		t.Errorf("#metadata doit rester un commentaire: %v", err)
	}

	// Valeur invalide: erreur explicite.
	if _, err := Parse("t", strings.NewReader("#meta rebalance:souvent"+"\n"+"100 VOO")); err == nil {
		t.Error("erreur attendue pour une valeur non numérique")
	}
}

func TestParseFeesColumnAndEnvelope(t *testing.T) {
	in := `
#meta extra-fees:0,60  # enveloppe assurance-vie
60 VOO 0.03  S&P 500     # 3e colonne numérique = TER
40 BND       sans frais déclarés
`
	spec, err := Parse("t", strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if spec.EnvelopeFees != 0.60 {
		t.Errorf("EnvelopeFees = %v, attendu 0.60", spec.EnvelopeFees)
	}
	if h := spec.Holdings[0]; h.Fees != 0.03 || h.Note != "S&P 500" {
		t.Errorf("frais colonne: %+v", h)
	}
	if h := spec.Holdings[1]; h.Fees != -1 || h.Note != "sans frais déclarés" {
		t.Errorf("frais absents: %+v", h)
	}
	// Frais hors limites: erreur.
	if _, err := Parse("t", strings.NewReader("60 VOO 25 note")); err == nil {
		t.Error("erreur attendue pour des frais de 25 %/an")
	}
	// Point décimal (convention par défaut) avec suffixe %.
	sp2bis, err := Parse("t", strings.NewReader("100 VOO 0.25% note"))
	if err != nil || sp2bis.Holdings[0].Fees != 0.25 {
		t.Errorf("frais 0.25%%: %+v, %v", sp2bis.Holdings, err)
	}
	// Virgule décimale et suffixe %% acceptés en 3e colonne.
	sp2, err := Parse("t", strings.NewReader("100 VOO 0,25% note"))
	if err != nil || sp2.Holdings[0].Fees != 0.25 || sp2.Holdings[0].Note != "note" {
		t.Errorf("frais 0,25%%: %+v, %v", sp2.Holdings, err)
	}
	// Une 3e colonne commençant par un chiffre mais non numérique = texte.
	sp2, err = Parse("t", strings.NewReader("100 VOO 3a-objectif long terme"))
	if err != nil || sp2.Holdings[0].Fees != -1 || !strings.HasPrefix(sp2.Holdings[0].Note, "3a-objectif") {
		t.Errorf("3e colonne textuelle: %+v, %v", sp2.Holdings, err)
	}
	// Synonyme accepté.
	sp, err := Parse("t", strings.NewReader("#meta envelope-fees:1"+"\n"+"100 VOO"))
	if err != nil || sp.EnvelopeFees != 1 {
		t.Errorf("envelope-fees: %+v, %v", sp, err)
	}
	// L'ancienne clé française n'existe plus: clé inconnue = simple warning.
	sp, err = Parse("t", strings.NewReader("#meta frais:1"+"\n"+"100 VOO"))
	if err != nil || sp.EnvelopeFees != -1 || len(sp.Warnings) != 1 {
		t.Errorf("frais doit être une clé inconnue: %+v, %v", sp, err)
	}
}

func TestSimulateEnvelopeFees(t *testing.T) {
	n := 253 // ~1 an de bourse
	p := &Portfolio{
		Name:         "t",
		EnvelopeFees: 2.52, // 0.01 %/jour de bourse
		Assets: []Asset{
			{Symbol: "A", Weight: 1, Series: constSeries("A", 0, n, 100)},
		},
	}
	sim, err := Simulate(p, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := 100 * math.Pow(1-0.0001, float64(n-1))
	got := sim.Values[len(sim.Values)-1]
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("valeur finale avec frais d'enveloppe: %v, attendu %v", got, want)
	}
}

func TestSingle(t *testing.T) {
	spec := Single(" NTSG ")
	if spec.Name != "NTSG" || len(spec.Holdings) != 1 {
		t.Fatalf("spec: %+v", spec)
	}
	h := spec.Holdings[0]
	if h.ID != "NTSG" || h.Weight != 1 || h.RawWeight != 100 {
		t.Errorf("holding: %+v", h)
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{
		"", // vide
		"# que des commentaires",
		"VOO",     // pas de poids
		"abc VOO", // poids non numérique
		"0 VOO",   // poids nul
		"150 VOO", // poids > 100
		"60",      // pas d'identifiant
	} {
		if _, err := Parse("t", strings.NewReader(in)); err == nil {
			t.Errorf("erreur attendue pour %q", in)
		}
	}
}
