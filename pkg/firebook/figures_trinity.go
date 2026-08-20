package firebook

import (
	"fmt"
	"strings"
)

// The cliff plate of the Trinity article. The article's three lasting lessons
// are all statements about the SHAPE of one surface, success against withdrawal
// rate, and the article states them in words: the collapse between four and
// five percent, the asymmetry of the allocation (too few equities being far
// worse than too many), and the very idea of a success rate. A table of cells,
// which is what Trinity published, hides the shape; four curves on one axis
// are the shape.
//
// Nothing here is copied from Trinity. The four curves are recomputed on the
// bundled Jorda-Schularick-Taylor record for the United States, which starts
// half a century before Trinity's sample does, so the plate is the article's
// claim tested on more evidence rather than its published table redrawn. The
// legend says both things.

// The recomputation's conventions, all of them the article's: rolling
// thirty-year windows, a fixed real withdrawal taken at the start of each year
// as a share of the STARTING capital, the two legs rebalanced every year, and
// success meaning the capital paid every withdrawal in full.
const (
	trinityYears  = 30
	trinityFirst  = 1872
	trinityLast   = 2020
	trinityWindow = 120 // rolling cohorts the record yields
)

// trinityAlloc is one curve: its equity share, the name the plate prints, and
// its success rate at each withdrawal rate of the grid.
type trinityAlloc struct {
	equity  float64
	name    string
	color   string
	success []float64
}

// trinityRates is the grid, from three to eight percent by quarter points.
var trinityRates = []float64{
	3.00, 3.25, 3.50, 3.75, 4.00, 4.25, 4.50, 4.75, 5.00, 5.25, 5.50,
	5.75, 6.00, 6.25, 6.50, 6.75, 7.00, 7.25, 7.50, 7.75, 8.00,
}

// The four allocations of the article's grid, measured on the bundled record.
var trinityAllocs = []trinityAlloc{
	{1.00, "100 % actions", figDeep, []float64{
		100.0, 100.0, 100.0, 100.0, 98.3, 95.8, 92.5, 88.3, 81.7, 79.2, 75.8,
		72.5, 70.8, 67.5, 63.3, 60.0, 53.3, 50.8, 48.3, 45.0, 37.5}},
	{0.75, "75 / 25", figBlue, []float64{
		100.0, 100.0, 100.0, 100.0, 98.3, 95.8, 90.8, 84.2, 78.3, 75.0, 70.8,
		68.3, 65.0, 60.0, 54.2, 46.7, 46.7, 40.0, 37.5, 30.8, 27.5}},
	{0.50, "50 / 50", figGreen, []float64{
		100.0, 100.0, 100.0, 98.3, 94.2, 90.0, 80.8, 73.3, 70.8, 62.5, 60.8,
		57.5, 48.3, 40.0, 32.5, 30.0, 26.7, 23.3, 20.0, 14.2, 11.7}},
	{0.25, "25 / 75", figBad, []float64{
		100.0, 100.0, 98.3, 89.2, 74.2, 64.2, 56.7, 50.8, 47.5, 44.2, 38.3,
		31.7, 27.5, 25.8, 22.5, 20.0, 15.8, 12.5, 9.2, 5.8, 5.0}},
}

// trinityAt reads one curve at one withdrawal rate.
func trinityAt(a trinityAlloc, rate float64) float64 {
	for i, r := range trinityRates {
		if r == rate {
			return a.success[i]
		}
	}
	return 0
}

// trinityCliff is the article's first lesson measured: what the success rate of
// an allocation loses between four and five percent, against what it loses
// between three and four.
func trinityCliff(a trinityAlloc) (three, four float64) {
	return trinityAt(a, 3) - trinityAt(a, 4), trinityAt(a, 4) - trinityAt(a, 5)
}

// trinityCross is the first withdrawal rate at which all equity does strictly
// better than the 75/25 mix: the article's "trop peu d'actions est bien plus
// dangereux que trop", read at the top of the range where it becomes visible.
func trinityCross() float64 {
	for _, r := range trinityRates {
		if trinityAt(trinityAllocs[0], r) > trinityAt(trinityAllocs[1], r) {
			return r
		}
	}
	return 0
}

// The plate's geometry.
const (
	triX0, triX1   = 78.0, 600.0
	triTop, triBot = 112.0, 312.0
)

func trinityScales() (rate, success figScale) {
	return figScale{Min: 3, Max: 8, Px0: triX0, Px1: triX1},
		figScale{Min: 0, Max: 100, Px0: triBot, Px1: triTop}
}

func figTrinityFalaise() string {
	rate, success := trinityScales()
	var b strings.Builder
	b.WriteString(plateHead("la falaise de trinity",
		"Entre quatre et cinq pour cent, le sol se dérobe"))
	b.WriteString(plateDeck(
		"Part des retraites de trente ans qui sont allées au bout, taux de retrait par taux de retrait"))

	axisTicks(&b, success, []float64{0, 25, 50, 75, 100}, 0, " %", triX0, triX1, false)
	for r := 3.0; r <= 8.001; r++ {
		b.WriteString(mTxt(rate.Map(r), triBot+18, 10, figMuted, "middle", "400",
			fmt.Sprintf("%.0f %%", r)))
	}
	b.WriteString(sTxt(triX0, triBot+37, 9.5, figMuted, "start", "400",
		"taux de retrait initial, indexé sur l'inflation ensuite  →"))

	// The cliff itself, marked under the curves before they are drawn.
	cx0, cx1 := rate.Map(4), rate.Map(5)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		cx0, triTop, cx1-cx0, triBot-triTop, mixHex("#fffdf9", figBad, 0.07))
	b.WriteString(sTxt((cx0+cx1)/2, triBot-30, 10, figBad, "middle", "600", "la falaise"))

	for _, a := range trinityAllocs {
		var pts [][2]float64
		for i, r := range trinityRates {
			pts = append(pts, [2]float64{rate.Map(r), success.Map(a.success[i])})
		}
		b.WriteString(poly(pts, a.color, 2.4, ""))
	}
	// Each curve named at its right end, where the four are furthest apart.
	for _, a := range trinityAllocs {
		last := a.success[len(a.success)-1]
		b.WriteString(sTxt(triX1+8, success.Map(last)+3.5, 10, a.color, "start", "600", a.name))
	}

	// The crossing the article's second lesson rests on: past this rate, all
	// equity is the safer of the two, not the bolder.
	cr := trinityCross()
	cy := success.Map(trinityAt(trinityAllocs[0], cr))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.5" fill="%s"/>`, rate.Map(cr), cy, figDeep)
	lx := rate.Map(5.6)
	b.WriteString(line(lx-4, 126, rate.Map(cr)+7, cy, figMuted, 1))
	b.WriteString(sTxt(lx, 122, 10, figSoft, "start", "600",
		"à partir de "+frNum(cr, 2)+" %, le 100 % actions"))
	b.WriteString(sTxt(lx, 134, 10, figSoft, "start", "600",
		"repasse devant le 75/25"))

	b.WriteString(plateConclusion(376, fmt.Sprintf(
		"Un point de retrait en plus, de 4 à 5 %%, coûte %s points de succès au 75/25 ; "+
			"le point précédent n'en coûtait que %s.",
		frNum(trinityCliffFour(), 0), frNum(trinityCliffThree(), 0))))
	b.WriteString(plateFoot(398, []string{
		"Recalcul maison sur le panel Jorda-Schularick-Taylor des États-Unis (rendements réels annuels, " +
			"actions et obligations d'État,",
		fmt.Sprintf("%d-%d), soit %d fenêtres glissantes de trente ans. Aucun chiffre publié par Trinity "+
			"n'est repris ici :", trinityFirst, trinityLast, trinityWindow),
		"l'étude d'origine porte sur 1926-1995 et en nominal pour partie, quand ce panel remonte un demi-siècle " +
			"plus haut et reste réel.",
		"Retrait fixe pris en début d'année, part du capital de départ, tenu en pouvoir d'achat, " +
			"rééquilibrage annuel des deux jambes.",
		"L'échantillon plus profond durcit le verdict : à 4 %, le 75/25 échoue deux fois sur cent ici, " +
			"jamais chez Trinity, et le 25/75",
		"échoue une fois sur quatre là où le texte dit une sur cinq. L'asymétrie, elle, est intacte : " +
			"trop peu d'actions reste bien plus dangereux que trop.",
	}))
	return svg(700, 490, b.String())
}

// trinityCliffThree and trinityCliffFour are the two steps the conclusion
// compares, on the article's own 75/25 line.
func trinityCliffThree() float64 {
	three, _ := trinityCliff(trinityAllocs[1])
	return three
}

func trinityCliffFour() float64 {
	_, four := trinityCliff(trinityAllocs[1])
	return four
}
