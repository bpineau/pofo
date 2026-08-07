package firebook

import (
	"math"
	"strings"
	"testing"
)

// The flat-tax-et-imposition plate draws one closed formula, rate × (1 − PMP /
// cours), against time. This recomputes every number it shows and every year it
// names, so plate, formula and article can never drift apart.

// The two readings the figure backlog and the plate quote: about 12 % of the
// flow after ten years, about 24 % after thirty, for a CTO at the PFU on a line
// whose price compounds at 5 % a year in real terms.
func TestPmpDriftMatchesTheQuotedReadings(t *testing.T) {
	for _, tc := range []struct {
		years int
		want  float64
	}{{10, 0.1212}, {30, 0.2413}} {
		got := ctoSlope * gainFractionLump(float64(tc.years), pmpRealGrowth)
		if math.Abs(got-tc.want) > 5e-4 {
			t.Errorf("à %d ans: %.4f, la planche dit %.4f", tc.years, got, tc.want)
		}
		if lbl := pctFR(got); !strings.Contains(figFrictionDerivePmp(), lbl) {
			t.Errorf("la planche ne porte pas %q", lbl)
		}
	}
	// The gain fraction is the whole mechanism: a line bought yesterday carries
	// no gain, so it is extracted almost free.
	if got := gainFractionLump(0, pmpRealGrowth); got != 0 {
		t.Errorf("part de gain à l'achat: %.4f, attendu 0", got)
	}
	if got := ctoSlope * gainFractionLump(3, pmpRealGrowth); got > 0.05 {
		t.Errorf("friction à trois ans: %.4f, la planche la dit sous 5 %%", got)
	}
}

// Every curve climbs, never falls, and converges to the rate it is drawn for:
// the plateau is the nominal rate, reached only when the PMP has become
// negligible against the price.
func TestPmpDriftPlateaux(t *testing.T) {
	for _, rate := range []float64{ctoSlope, peaSlope} {
		if got := rate * gainFractionLump(500, pmpRealGrowth); math.Abs(got-rate) > 1e-6 {
			t.Errorf("plateau de l'achat unique: %.6f, attendu %.6f", got, rate)
		}
		if got := rate * gainFractionRegular(500, pmpRealGrowth); math.Abs(got-rate) > 1e-3 {
			t.Errorf("plateau des versements: %.6f, attendu %.6f", got, rate)
		}
		prev := -1.0
		for y := 0; y <= pmpMaxYears; y++ {
			for _, f := range []float64{rate * gainFractionLump(float64(y), pmpRealGrowth), rate * gainFractionRegular(y, pmpRealGrowth)} {
				if f > rate+1e-12 {
					t.Errorf("à %d ans, la friction %.4f dépasse son plafond %.4f", y, f, rate)
				}
			}
			if cur := rate * gainFractionLump(float64(y), pmpRealGrowth); cur < prev {
				t.Errorf("la courbe redescend à %d ans", y)
			} else {
				prev = cur
			}
		}
	}
	// The plate names both plateaux in the rates of 1 January 2026.
	svg := figFrictionDerivePmp()
	for _, want := range []string{"plafond du PFU : 31,4 %", "plafond du PEA : 18,6 %"} {
		if !strings.Contains(svg, want) {
			t.Errorf("la planche ne porte pas %q", want)
		}
	}
}

// The dashed variant: one constant contribution a year keeps pulling the PMP
// back towards the price, so the drift is slower at every anniversary but the
// first, where the two histories are still the same single purchase.
func TestPmpRegularContributionsSlowTheDrift(t *testing.T) {
	if lump, reg := gainFractionLump(1, pmpRealGrowth), gainFractionRegular(1, pmpRealGrowth); math.Abs(lump-reg) > 1e-12 {
		t.Errorf("à un an: %.6f contre %.6f, les deux histoires sont identiques", lump, reg)
	}
	for y := 2; y <= 120; y++ {
		lump := gainFractionLump(float64(y), pmpRealGrowth)
		reg := gainFractionRegular(y, pmpRealGrowth)
		if reg >= lump {
			t.Errorf("à %d ans: versements %.6f, achat unique %.6f, l'apport doit freiner la dérive", y, reg, lump)
		}
		if reg <= 0 || reg >= 1 {
			t.Errorf("à %d ans, part de gain hors bornes: %.6f", y, reg)
		}
	}
	// The two end labels the plate carries, at thirty-five years.
	if got := peaSlope * gainFractionLump(pmpMaxYears, pmpRealGrowth); math.Abs(got-0.1523) > 5e-4 {
		t.Errorf("PEA à %d ans: %.4f, la planche dit 15,2 %%", pmpMaxYears, got)
	}
	if got := peaSlope * gainFractionRegular(pmpMaxYears, pmpRealGrowth); math.Abs(got-0.1174) > 5e-4 {
		t.Errorf("PEA alimenté à %d ans: %.4f, la planche dit 11,7 %%", pmpMaxYears, got)
	}
}

// The reading written under the traces, year by year: a CTO at the PFU only
// crosses 15 % of the flow after thirteen years and 25 % after thirty-two.
func TestPmpDriftCrossesTheCalibrationBand(t *testing.T) {
	f := func(y int) float64 { return ctoSlope * gainFractionLump(float64(y), pmpRealGrowth) }
	if f(13) >= mixedRateLow || f(14) <= mixedRateLow {
		t.Errorf("le passage des 15 %% tombe hors de la treizième année: %.4f puis %.4f", f(13), f(14))
	}
	if f(32) >= mixedRateHigh || f(33) <= mixedRateHigh {
		t.Errorf("le passage des 25 %% tombe hors de la trente-deuxième année: %.4f puis %.4f", f(32), f(33))
	}
	// The band is the article's own: "le taux mixte d'un plan bien organisé
	// ressort typiquement à 15-25 %", a rate on gains, so it must bracket the
	// mature-PEA rate and sit under the PFU.
	if mixedRateLow > peaSlope || peaSlope > mixedRateHigh {
		t.Errorf("la fourchette %.2f-%.2f n'encadre plus le taux du PEA (%.3f)", mixedRateLow, mixedRateHigh, peaSlope)
	}
	if ctoSlope <= mixedRateHigh {
		t.Errorf("le PFU (%.3f) devrait rester au-dessus de la fourchette", ctoSlope)
	}
}

// The plate itself: registered, honest about its hypotheses, and free of what
// crengine cannot render.
func TestFrictionDerivePmpPlate(t *testing.T) {
	svg := figFrictionDerivePmp()
	for _, want := range []string{
		"25,7 %", "15,2 %", "11,7 %", // the three end labels
		"impôt payé pour 100 € de flux extrait", // the y axis, in flow terms
		"années depuis l'achat de la ligne",     // the x axis, in years
		"1er janvier 2026",                      // the reference year of the rates
		"Cours en hausse de 5 % par an en réel", // the growth hypothesis
		"un versement constant chaque année",    // the contribution schedule
		"est un taux sur les gains",             // what the band is, and is not
		"la fourchette de calibrage du curseur", // the band itself
		"un CTO au PFU ne dépasse 15 % du flux", // the reading
		"qu'après treize ans, et 25 % après trente-deux",
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
	if figures["friction-derive-pmp"] == nil {
		t.Error("la planche n'est pas enregistrée sous friction-derive-pmp")
	}
}
