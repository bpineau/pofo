package firebook

import (
	"math"
	"strings"
	"testing"
)

// The enveloppes-francaises plate draws nothing but the arithmetic of the rates
// in force on 1 January 2026. This pins those rates and that arithmetic on every
// friction the article quotes in prose, so plate and article can never drift
// apart: change one and this test fails.

// The rates themselves, as the tax reference set gives them (LFSS 2026 art. 12
// for the social levies, art. 125-0 A CGI for the assurance-vie). A rate is
// never quoted from memory, and the plate's footnote states the same six.
func TestEnvelopeRates2026(t *testing.T) {
	for _, tc := range []struct {
		name      string
		got, want float64
	}{
		{"prélèvements sociaux PEA et CTO", socialLevies2026, 0.186},
		{"prélèvements sociaux assurance-vie", socialLeviesAV2026, 0.172},
		{"IR du PFU", pfuIncomeTax, 0.128},
		{"IR assurance-vie après 8 ans", avIncomeTax8Y, 0.075},
		{"abattement couple", avAllowanceCouple, 9200},
		{"abattement célibataire", avAllowanceSingle, 4600},
		// The three slopes the article names: 18,6 %, 31,4 % and 24,7 %.
		{"pente PEA", peaSlope, 0.186},
		{"pente CTO", ctoSlope, 0.314},
		{"pente assurance-vie", avSlope, 0.247},
	} {
		if math.Abs(tc.got-tc.want) > 1e-9 {
			t.Errorf("%s: %.4f, attendu %.4f", tc.name, tc.got, tc.want)
		}
	}
	// The assurance-vie was excluded from the LFSS 2026 rise, which is the whole
	// reason its slope is not built on 18,6 %.
	if socialLeviesAV2026 >= socialLevies2026 {
		t.Error("l'assurance-vie doit rester sous le taux relevé par la LFSS 2026")
	}
}

// Every friction the article states, recomputed from the rates.
func TestEnvelopeFrictionsMatchTheArticle(t *testing.T) {
	// "Le taux effectif d'un retrait vaut 18,6 % de la fraction de plus-value.
	// Pour un PEA qui a doublé, cela fait environ 9,3 % de friction sur le flux."
	if got := 0.5 * peaSlope; math.Abs(got-0.093) > 1e-4 {
		t.Errorf("PEA doublé: %.3f, l'article dit 9,3 %%", got)
	}
	// "La friction démarre alors à 3-8 % du flux" for a CTO whose PMP is still
	// close to the price, that is 10 to 25 % of gain inside each sale.
	if lo, hi := 0.10*ctoSlope, 0.25*ctoSlope; lo < 0.03 || hi > 0.08 {
		t.Errorf("CTO jeune: %.3f à %.3f, l'article dit 3 à 8 %%", lo, hi)
	}
	// The two readings the plate spells out under the traces.
	if got := 0.20 * ctoSlope; math.Abs(got-0.0628) > 1e-4 {
		t.Errorf("CTO à 20 %% de gain: %.4f, la planche dit 6,3 %%", got)
	}
	if got := 0.80 * avSlope; math.Abs(got-0.1976) > 1e-4 {
		t.Errorf("assurance-vie à 80 %% de gain: %.4f, la planche dit 19,8 %%", got)
	}

	// The couple of the assurance-vie section: 30 000 € redeemed from a contract
	// with 40 % of gain. The allowance covers the income tax and nothing else,
	// so 210 € of IR and 2 064 € of social levies are due, a 7,6 % friction.
	const redeem, gainShare = 30000.0, 0.40
	gain := gainShare * redeem
	if math.Abs(gain-12000) > 1 {
		t.Errorf("quote-part de gains %.0f €, l'article dit 12 000 €", gain)
	}
	ir := avIncomeTax8Y * (gain - avAllowanceCouple)
	ps := socialLeviesAV2026 * gain
	if math.Abs(ir-210) > 1 || math.Abs(ps-2064) > 1 {
		t.Errorf("IR %.0f € et PS %.0f €, l'article dit 210 € et 2 064 €", ir, ps)
	}
	if got := avFriction(gainShare, redeem, avAllowanceCouple); math.Abs(got-0.0758) > 5e-4 {
		t.Errorf("friction du couple: %.4f, l'article dit 7,6 %%", got)
	}
	// "contre 9,9 % sans abattement", and the allowance is worth 690 € a year.
	if got := gain * avSlope / redeem; math.Abs(got-0.0988) > 5e-4 {
		t.Errorf("friction sans abattement: %.4f, l'article dit 9,9 %%", got)
	}
	if worth := avIncomeTax8Y * avAllowanceCouple; math.Abs(worth-690) > 1 {
		t.Errorf("l'abattement vaut %.0f €, l'article dit 690 €", worth)
	}
}

// The ::: exemple block of the article, line by line: the same 55 000 € of
// spending, drawn from one envelope or from three.
func TestEnvelopeHouseholdExample(t *testing.T) {
	const need, gainShare = 55000.0, 0.45

	// Organisation A, everything in a CTO: "une friction de 14 % environ. Il
	// faut vendre 64 000 € environ, pour 9 000 € d'impôts par an."
	friction := gainShare * ctoSlope
	if math.Abs(friction-0.1413) > 5e-4 {
		t.Errorf("friction CTO: %.4f, l'article dit 14 %%", friction)
	}
	sold := need / (1 - friction)
	taxA := sold - need
	if math.Abs(sold-64000) > 500 || math.Abs(taxA-9000) > 500 {
		t.Errorf("organisation A: %.0f € vendus pour %.0f € d'impôt, l'article dit 64 000 € et 9 000 €", sold, taxA)
	}

	// Organisation B: 20 000 € of assurance-vie (zero income tax, but 1 548 € of
	// social levies), 28 000 € of PEA, 11 600 € of CTO at a 20 % gain share. The
	// CTO leg is what closes the household's need: B must NET the same 55 000 €
	// that A nets, so its gross sale follows from its own friction.
	avTax := avFriction(gainShare, 20000, avAllowanceCouple) * 20000
	if math.Abs(avTax-1550) > 5 {
		t.Errorf("rachat d'assurance-vie: %.0f € d'impôt, l'article dit 1 550 €", avTax)
	}
	if ir := math.Max(0, gainShare*20000-avAllowanceCouple); ir != 0 {
		t.Errorf("la quote-part imposable à l'IR vaut %.0f €, l'article la dit nulle", ir)
	}
	peaTax := 28000 * gainShare * peaSlope
	ctoTax := 11600 * 0.20 * ctoSlope
	if math.Abs(peaTax-2300) > 100 || math.Abs(ctoTax-730) > 50 {
		t.Errorf("PEA %.0f € et CTO %.0f € d'impôt, l'article dit 2 300 € et 730 €", peaTax, ctoTax)
	}
	taxB := avTax + peaTax + ctoTax
	if math.Abs(taxB-4600) > 100 {
		t.Errorf("organisation B: %.0f € d'impôt, l'article dit 4 600 €", taxB)
	}
	if got := taxB / 59600; math.Abs(got-0.078) > 2e-3 {
		t.Errorf("friction totale de B: %.4f, l'article dit 7,8 %% des 59 600 € vendus", got)
	}
	if gap := taxA - taxB; math.Abs(gap-4400) > 200 {
		t.Errorf("l'écart annuel vaut %.0f €, l'article dit 4 400 €", gap)
	}
	// "Capitalisé sur 30 ans de retraite, il équivaut à 180 000 € de capital
	// environ", the future value of that yearly gap at 2 % real.
	fv := (taxA - taxB) * (math.Pow(1.02, 30) - 1) / 0.02
	if math.Abs(fv-180000) > 10000 {
		t.Errorf("capitalisé sur 30 ans: %.0f €, l'article dit 180 000 €", fv)
	}
}

// The fourth curve is the only bent one, and it bends exactly where the
// allowance runs out: flat social-levy slope before, full assurance-vie slope
// after, the two halves meeting without a jump.
func TestEnvelopeAllowanceKink(t *testing.T) {
	gStar := avAllowanceExhausted(avRedemption, avAllowanceCouple)
	if math.Abs(gStar-0.368) > 1e-9 {
		t.Errorf("le coude tombe à %.4f de part de gain, attendu 0,368", gStar)
	}
	f := func(g float64) float64 { return avFriction(g, avRedemption, avAllowanceCouple) }

	// Below the kink, the friction is the bare social levies: the allowance
	// never spares them, whatever the article used to claim.
	for _, g := range []float64{0, 0.1, 0.2, gStar} {
		if got := f(g); math.Abs(got-socialLeviesAV2026*g) > 1e-12 {
			t.Errorf("à %.3f de gain: %.5f, attendu %.5f (prélèvements sociaux seuls)", g, got, socialLeviesAV2026*g)
		}
	}
	// Above it, the curve runs parallel to the plain 24,7 % slope, offset by
	// what the allowance saved, and it joins that slope continuously at gStar.
	saved := avIncomeTax8Y * avAllowanceCouple / avRedemption
	for _, g := range []float64{gStar, 0.5, 0.75, 1} {
		if got := f(g); math.Abs(got-(avSlope*g-saved)) > 1e-12 {
			t.Errorf("à %.3f de gain: %.5f, attendu %.5f (pente pleine décalée)", g, got, avSlope*g-saved)
		}
	}
	if left, right := socialLeviesAV2026*gStar, avSlope*gStar-saved; math.Abs(left-right) > 1e-12 {
		t.Errorf("le raccord au coude saute: %.6f contre %.6f", left, right)
	}
	// The slope really does change there, from 17,2 % to 24,7 %.
	const h = 1e-6
	if before, after := (f(gStar)-f(gStar-h))/h, (f(gStar+h)-f(gStar))/h; math.Abs(before-socialLeviesAV2026) > 1e-6 || math.Abs(after-avSlope) > 1e-6 {
		t.Errorf("pentes autour du coude: %.4f puis %.4f, attendu 0,172 puis 0,247", before, after)
	}
	// The couple's curve ends at 21,9 %, the value the plate labels, and stays
	// under the plain assurance-vie everywhere.
	if got := f(1); math.Abs(got-0.2194) > 1e-4 {
		t.Errorf("bout de courbe: %.4f, la planche dit 21,9 %%", got)
	}
	for g := 0.0; g <= 1.0001; g += 0.05 {
		if f(g) > avSlope*g+1e-12 {
			t.Errorf("à %.2f de gain, l'abattement coûte plus cher que son absence", g)
		}
	}
}

// The plate itself: the id is registered, and it says what it must say.
func TestFrictionsEnveloppesPlate(t *testing.T) {
	svg := figFrictionsEnveloppes()
	for _, want := range []string{
		"31,4 %", "24,7 %", "21,9 %", "18,6 %", // the four end labels
		"1er janvier 2026",                // the reference year
		"impôt payé, en % du flux retiré", // the unambiguous y axis
		"jamais les prélèvements sociaux", // what the allowance does not buy
		"part de gain dans le retrait",    // the x axis
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
	if figures["frictions-enveloppes"] == nil {
		t.Error("la planche n'est pas enregistrée sous frictions-enveloppes")
	}
}
