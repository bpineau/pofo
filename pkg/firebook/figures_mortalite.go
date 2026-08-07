package firebook

import (
	"fmt"
	"strings"
)

// The "alive, broke or gone" plate crosses a decumulation ensemble with a
// couple survival curve. Every number below was produced by this repository's
// own engines and is re-derived by TestVivantRuinePartiMatchesTheEngine.
//
// The plan is the article's couple (Léa and Sam, retiring in their forties on
// a horizon that runs to Léa's 100th birthday), spending the rate the same
// article's horizon table gives for a 50-year plan:
//
//	plan: Capital 1e6, NeedAnnual 33000 (3.3 %), Years 53,
//	      Cashflows {FromYear: 20, Annual: 14000},
//	      Flex {Threshold: 0.20, Cut: 0.10}, no tax, no buffer,
//	source: scenario.PooledBootstrap of the bundled Jorda-Schularick-Taylor
//	      per-country real returns blended 60/40 within each country,
//	      mean block 10 years,
//	draw: 200 000 paths, seed 1, 8 workers (DrawPaths splits the sampling per
//	      worker, so the worker count is part of the reproduction recipe).
//
// Mortality is decumul.FrenchMortality (Gompertz, modal age 88, dispersion
// 10), read through CoupleSurvival at age 47. That law is UNISEX and
// CoupleSurvival assumes two independent lives of the SAME age, so the plate
// says so: no sex split is invented here, and Sam's two extra years are not
// modelled.
//
// Sampling noise at 200 000 paths is well inside a tenth of a point: seeds 1,
// 2 and 3 give a gross ruin of 17.68 / 17.67 / 17.58 % and a lived ruin of
// 14.12 / 14.07 / 14.04 %.
const (
	vrpGross = 17.68 // ruin probability (%), mortality ignored
	vrpLived = 14.12 // probability (%) of ever being alive AND broke
	vrpPeak  = 9.72  // largest alive-and-broke share (%), reached at year 35
	vrpPeakY = 35
	// vrpAliveY40 is the couple survival 40 years in (ages 87), the number that
	// explains why the weighting relieves so little.
	vrpAliveY40 = 65.35
	// vrpLived65 reads the SAME ensemble, hence the same failure timing,
	// against a couple aged 65 instead of 47. Only the mortality reading
	// changes, and the relief goes from a fifth to more than a half. The
	// article quotes it to show that the discount is a property of the reader's
	// age, not of the plan. It is not drawn on the plate.
	vrpLived65 = 7.67
)

// vrpBroke is the share of households (%) that are alive AND broke at each
// year-end 0..53: Ensemble.LifeCurve(couple survival).Broke.
var vrpBroke = []float64{
	0.00, 0.00, 0.00, 0.03, 0.07, 0.20,
	0.31, 0.47, 0.70, 0.94, 1.28, 1.69,
	2.08, 2.43, 2.85, 3.34, 3.82, 4.30,
	4.85, 5.48, 5.83, 6.19, 6.54, 6.89,
	7.24, 7.61, 7.95, 8.23, 8.51, 8.79,
	9.04, 9.25, 9.45, 9.60, 9.69, 9.72,
	9.70, 9.62, 9.45, 9.22, 8.92, 8.52,
	8.05, 7.51, 6.91, 6.25, 5.56, 4.86,
	4.15, 3.48, 2.84, 2.25, 1.74, 1.28,
}

// vrpAlive is the couple survival curve (%) at age 47, year-ends 0..53:
// decumul.FrenchMortality.CoupleSurvival(47, t). It is the ceiling that
// crushes the red band from above.
var vrpAlive = []float64{
	100.00, 100.00, 100.00, 100.00, 99.99, 99.99,
	99.98, 99.97, 99.96, 99.94, 99.92, 99.89,
	99.86, 99.81, 99.76, 99.69, 99.60, 99.49,
	99.36, 99.19, 98.99, 98.75, 98.45, 98.09,
	97.66, 97.14, 96.52, 95.78, 94.90, 93.86,
	92.65, 91.22, 89.57, 87.66, 85.47, 82.97,
	80.14, 76.97, 73.45, 69.57, 65.35, 60.82,
	56.00, 50.96, 45.77, 40.51, 35.28, 30.17,
	25.30, 20.76, 16.62, 12.97, 9.83, 7.22,
}

// figVivantRuineParti stacks the three exclusive states of a retirement plan,
// year by year: alive and funded, alive and broke, gone. The red ribbon is the
// only state that costs anything, and it is squeezed from below by the plans
// that hold and from above by the couples that are no longer there. The plate
// exists to replace an assertion by a measurement: the mortality weighting
// does NOT divide a FIRE couple's ruin by two or three, it takes about a fifth
// off, because the failures land while the couple is still very much alive.
func figVivantRuineParti() string {
	m := mapper(0, 53, 0, 100, 60, 600, 352, 100)
	x := func(t float64) float64 { return m(t, 0)[0] }
	y := func(v float64) float64 { return m(0, v)[1] }

	funded := make([]float64, len(vrpAlive))
	ceiling := make([]float64, len(vrpAlive))
	floor := make([]float64, len(vrpAlive))
	for t := range vrpAlive {
		funded[t] = vrpAlive[t] - vrpBroke[t]
		ceiling[t] = 100
	}

	// area fills the ribbon between an upper and a lower series.
	area := func(hi, lo []float64, fill string) string {
		var d strings.Builder
		for t, v := range hi {
			verb := " L %.1f,%.1f"
			if t == 0 {
				verb = "M %.1f,%.1f"
			}
			fmt.Fprintf(&d, verb, x(float64(t)), y(v))
		}
		for t := len(lo) - 1; t >= 0; t-- {
			fmt.Fprintf(&d, " L %.1f,%.1f", x(float64(t)), y(lo[t]))
		}
		return fmt.Sprintf(`<path d="%s Z" fill="%s"/>`, d.String(), fill)
	}

	var b strings.Builder
	b.WriteString(plateHead("ruine et mortalité", "Vivant, ruiné ou parti : les trois états, année par année"))

	// the headline pair, read left to right
	b.WriteString(mTxt(24, 76, 15, figMuted, "start", "600", "17,7 %"))
	b.WriteString(sTxt(82, 76, 11, figMuted, "start", "400", "de ruine brute"))
	b.WriteString(sTxt(168, 76, 12, figMuted, "start", "400", "→"))
	b.WriteString(mTxt(190, 76, 15, figBad, "start", "600", "14,1 %"))
	b.WriteString(sTxt(248, 76, 11, figSoft, "start", "600", "de ruine vécue, un jour vivant et ruiné"))

	// the three states, stacked
	b.WriteString(area(ceiling, vrpAlive, figRule))
	b.WriteString(area(vrpAlive, funded, figBad))
	b.WriteString(area(funded, floor, figWash))

	// y scale, outside the plot so nothing is hidden behind the areas
	for _, g := range []float64{0, 25, 50, 75, 100} {
		b.WriteString(line(54, y(g), 60, y(g), figRule, 1))
		b.WriteString(mTxt(48, y(g)+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(600, 92, 10.5, figMuted, "end", "400", "part des scénarios (%)"))

	// direct labels, one per state
	b.WriteString(sTxt(596, 130, 12, figSoft, "end", "600", "décédé"))
	b.WriteString(sTxt(596, 146, 10.5, figMuted, "end", "400", "le couple n'est plus là"))
	b.WriteString(sTxt(76, 300, 12, figSoft, "start", "600", "vivant et solvable"))
	b.WriteString(sTxt(76, 316, 10.5, figMuted, "start", "400", "le plan tient"))

	// the ribbon, called out with a leader into the pale zone below it
	b.WriteString(line(292, 216, 414, 156, figBad, 1))
	fmt.Fprintf(&b, `<circle cx="414.0" cy="156.0" r="2.8" fill="%s"/>`, figBad)
	b.WriteString(sTxt(180, 222, 12, figBad, "start", "600", "vivant et ruiné"))
	b.WriteString(sTxt(180, 238, 10.5, figMuted, "start", "400", "le seul état qui coûte quelque chose"))
	b.WriteString(mTxt(180, 256, 10.5, figBad, "start", "600", "9,7 %"))
	b.WriteString(sTxt(219, 256, 10.5, figMuted, "start", "400", "au pic, à l'année 35"))

	// x scale
	b.WriteString(line(60, 352, 600, 352, figRule, 1))
	for _, t := range []float64{0, 10, 20, 30, 40, 50} {
		b.WriteString(mTxt(x(t), 368, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", t)))
	}
	b.WriteString(sTxt(330, 390, 11, figMuted, "middle", "400",
		"année du plan (départ à 47 ans, donc 87 ans à l'année 40)"))

	for i, s := range []string{
		"Plan : 1 M€, 33 k€/an réels (3,3 %), 53 ans, pension de 14 k€/an à l'année 20, coupe tenue de 10 %.",
		"Modèle : 16 pays, 1871-2020, 60/40, blocs de 10 ans, 200 000 tirages. Gompertz unisexe, couple de même âge.",
		"À l'année 40, deux couples sur trois ont encore un survivant, d'où une remise d'un cinquième seulement.",
	} {
		b.WriteString(sTxt(24, 412+float64(i)*14, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, 452, b.String())
}
