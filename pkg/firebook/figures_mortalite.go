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
//	draw: 200 000 paths, seed 1 (frozen with 8 workers when Draw still split
//	      the sampling per worker; since 2026-10 each path has its own stream
//	      and the worker count no longer reaches the result).
//
// Mortality is decumul.FrenchMortality read through CoupleSurvival at age 47:
// the Gompertz law refitted on INSEE's 2023-2025 table in 2026-10 (modal age
// 89.8, dispersion 9.3; the plate was first frozen on 88 / 10, which ran a
// year and a half short of INSEE and lifted the relief to a fifth). That law
// is UNISEX and CoupleSurvival assumes two independent lives of the SAME age,
// so the plate says so: no sex split is invented here, and Sam's two extra
// years are not modelled.
//
// Sampling noise at 200 000 paths stays within a quarter of a point: seeds
// 1, 2 and 3 give a gross ruin of 17.56 / 17.66 / 17.79 % and a lived
// ruin of 14.57 / 14.67 / 14.77 %, a relief of 17.0 % on all three.
const (
	vrpGross = 17.56  // ruin probability (%), mortality ignored
	vrpLived = 14.57  // probability (%) of ever being alive AND broke
	vrpPeak  = 10.352 // largest alive-and-broke share (%), reached at year 37; three decimals so the plate prints 10,4
	vrpPeakY = 37
	// vrpAliveY40 is the couple survival 40 years in (ages 87), the number that
	// explains why the weighting relieves so little.
	vrpAliveY40 = 73.16
	// vrpLived65 reads the SAME ensemble, hence the same failure timing,
	// against a couple aged 65 instead of 47. Only the mortality reading
	// changes, and the relief goes from a sixth to more than a half. The
	// article quotes it to show that the discount is a property of the reader's
	// age, not of the plan. It is not drawn on the plate.
	vrpLived65 = 8.15
)

// vrpBroke is the share of households (%) that are alive AND broke at each
// year-end 0..53: Ensemble.LifeCurve(couple survival).Broke.
var vrpBroke = []float64{
	0.00, 0.00, 0.00, 0.03, 0.07, 0.18,
	0.30, 0.46, 0.69, 0.92, 1.26, 1.66,
	2.04, 2.40, 2.80, 3.31, 3.79, 4.33,
	4.89, 5.52, 5.87, 6.21, 6.56, 6.89,
	7.24, 7.61, 7.97, 8.28, 8.60, 8.94,
	9.23, 9.50, 9.76, 9.97, 10.13, 10.26,
	10.33, 10.35, 10.29, 10.14, 9.90, 9.58,
	9.18, 8.70, 8.12, 7.47, 6.76, 6.00,
	5.22, 4.45, 3.69, 2.97, 2.33, 1.73,
}

// vrpAlive is the couple survival curve (%) at age 47, year-ends 0..53:
// decumul.FrenchMortality.CoupleSurvival(47, t). It is the ceiling that
// crushes the red band from above.
var vrpAlive = []float64{
	100.00, 100.00, 100.00, 100.00, 100.00, 99.99,
	99.99, 99.99, 99.98, 99.97, 99.96, 99.95,
	99.93, 99.91, 99.88, 99.84, 99.80, 99.74,
	99.67, 99.58, 99.46, 99.32, 99.15, 98.93,
	98.67, 98.35, 97.96, 97.48, 96.90, 96.20,
	95.36, 94.36, 93.16, 91.75, 90.08, 88.14,
	85.88, 83.27, 80.30, 76.93, 73.16, 68.98,
	64.42, 59.50, 54.27, 48.81, 43.21, 37.58,
	32.05, 26.73, 21.76, 17.24, 13.25, 9.86,
}

// figVivantRuineParti stacks the three exclusive states of a retirement plan,
// year by year: alive and funded, alive and broke, gone. The red ribbon is the
// only state that costs anything, and it is squeezed from below by the plans
// that hold and from above by the couples that are no longer there. The plate
// exists to replace an assertion by a measurement: the mortality weighting
// does NOT divide a FIRE couple's ruin by two or three, it takes about a sixth
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
	b.WriteString(mTxt(24, 76, 15, figMuted, "start", "600", frNum(vrpGross, 1)+" %"))
	b.WriteString(sTxt(82, 76, 11, figMuted, "start", "400", "de ruine brute"))
	b.WriteString(sTxt(168, 76, 12, figMuted, "start", "400", "→"))
	b.WriteString(mTxt(190, 76, 15, figBad, "start", "600", frNum(vrpLived, 1)+" %"))
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

	// the ribbon, called out with a leader into the pale zone below it; the
	// leader lands in the middle of the ribbon at its peak, the year the
	// callout names
	px, py := x(vrpPeakY), y(vrpAlive[vrpPeakY]-vrpBroke[vrpPeakY]/2)
	b.WriteString(line(292, 216, px, py, figBad, 1))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="2.8" fill="%s"/>`, px, py, figBad)
	b.WriteString(sTxt(180, 222, 12, figBad, "start", "600", "vivant et ruiné"))
	b.WriteString(sTxt(180, 238, 10.5, figMuted, "start", "400", "le seul état qui coûte quelque chose"))
	b.WriteString(mTxt(180, 256, 10.5, figBad, "start", "600", frNum(vrpPeak, 1)+" %"))
	b.WriteString(sTxt(226, 256, 10.5, figMuted, "start", "400", fmt.Sprintf("au pic, à l'année %d", vrpPeakY)))

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
		"À l'année 40, près de trois couples sur quatre ont encore un survivant, d'où une remise d'un sixième seulement.",
	} {
		b.WriteString(sTxt(24, 412+float64(i)*14, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, 452, b.String())
}
