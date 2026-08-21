package firebook

import (
	"fmt"
	"strings"
)

// The two-vintage gauge of the inflation article. Its pivot sentence is that
// the crash gives back and the episode never does: "Le krach de 1929 est
// brutal, puis il rend... L'épisode 1966-1981, lui, ne rend jamais." That is a
// claim about the SHAPE of two paths, and the book states it without ever
// drawing either.
//
// Drawn as the plan's own warning light, the current withdrawal rate read year
// after year, the two shapes are the argument: one rises, turns and comes home;
// the other crosses the irrecoverable band and leaves the plate.
//
// The subtlety the plate has to respect: the withdrawal is fixed in REAL terms,
// so on a real-return panel the gauge is a constant real spend divided by the
// real wealth left. The deflation of the thirties therefore does not appear as
// a falling numerator; it appears where the article says it belongs, in the
// wealth path itself, which recovers because deflation paid the bondholder and
// the indexed withdrawal shrank in nominal terms with the prices.

// The two vintages, run on the same panel, the same portfolio and the same
// rule. The allocation is the 60/40 the article names when it describes the
// episode ("C'est la situation où le portefeuille 60/40 n'a aucune poche qui
// gagne"), not the four-allocation grid of the Trinity plate.
const (
	voyantYears  = 30
	voyantRule   = 4.0 // percent of the starting capital, held in real terms
	voyantEquity = 0.60
	voyantBandLo = 8.0
	voyantBandHi = 10.0
)

// voyantVintage is one replay: its start year, the colour it is drawn in, and
// the current withdrawal rate read at the start of every year, in percent. A
// zero means the capital was already gone, and the curve stops there.
type voyantVintage struct {
	start int
	name  string
	color string
	rate  [voyantYears]float64
}

// The two replays, measured on the bundled Jorda-Schularick-Taylor record for
// the United States. Frozen, and recomputed by the guard test.
var voyantVintages = []voyantVintage{
	{1929, "millésime 1929", figBlue, [voyantYears]float64{
		4.00, 4.19, 4.86, 6.30, 6.12, 4.71, 4.96, 4.12, 3.50, 4.60,
		4.13, 4.10, 4.44, 5.17, 5.50, 5.27, 5.02, 4.26, 5.09, 6.02,
		6.53, 6.17, 5.68, 5.70, 5.58, 5.84, 4.76, 4.15, 4.33, 4.80}},
	{1966, "millésime 1966", figBad, [voyantYears]float64{
		4.00, 4.43, 4.32, 4.32, 5.23, 5.46, 5.22, 4.94, 6.23, 8.60,
		8.20, 7.84, 9.34, 10.71, 11.92, 12.92, 16.48, 16.68, 18.26, 21.38,
		21.60, 22.48, 30.28, 39.39, 54.12, 122.74, 0, 0, 0, 0}},
}

// voyantLive is how many years of the replay carry a reading: the whole horizon
// unless the capital ran out first.
func (v voyantVintage) live() int {
	for i, r := range v.rate {
		if r == 0 {
			return i
		}
	}
	return voyantYears
}

// voyantRuinYear is the calendar year the account emptied, or zero if it never
// did: the first year whose withdrawal the remaining capital could not pay in
// full, which is the year the gauge passes a hundred percent and not the blank
// year after it.
func (v voyantVintage) ruinYear() int {
	for i, r := range v.rate {
		if r >= 100 || r == 0 {
			return v.start + i
		}
	}
	return 0
}

// voyantPeak is the highest reading of a replay and the year it falls on.
func (v voyantVintage) peak() (float64, int) {
	best, year := 0.0, 0
	for i := 0; i < v.live(); i++ {
		if v.rate[i] > best {
			best, year = v.rate[i], v.start+i
		}
	}
	return best, year
}

// voyantCrossing is the first year a replay enters the irrecoverable band, or
// zero if it never does.
func (v voyantVintage) crossing() int {
	for i := 0; i < v.live(); i++ {
		if v.rate[i] >= voyantBandLo {
			return v.start + i
		}
	}
	return 0
}

// voyantReturns reports whether a replay comes back under the band and stays
// there: the article's "il rend".
func (v voyantVintage) returns() bool {
	n := v.live()
	if n < voyantYears {
		return false
	}
	for i := n / 2; i < n; i++ {
		if v.rate[i] >= voyantBandLo {
			return false
		}
	}
	return true
}

// voyantCrashPeak is the highest reading of the first decade, which for the
// 1929 replay is the crash itself rather than the post-war inflation that
// eventually tops it.
func (v voyantVintage) crashPeak() (float64, int) {
	best, year := 0.0, 0
	for i := 0; i < 10 && i < v.live(); i++ {
		if v.rate[i] > best {
			best, year = v.rate[i], v.start+i
		}
	}
	return best, year
}

// The plate's geometry.
const (
	voyX0, voyX1   = 80.0, 596.0
	voyTop, voyBot = 112.0, 320.0
	voyRateHi      = 24.0
)

func voyantScales() (year, rate figScale) {
	return figScale{Min: 0, Max: voyantYears - 1, Px0: voyX0, Px1: voyX1},
		figScale{Min: 0, Max: voyRateHi, Px0: voyBot, Px1: voyTop}
}

func figVoyants19291966() string {
	year, rate := voyantScales()
	var b strings.Builder
	b.WriteString(plateHead("le krach rend, l'épisode jamais",
		"Deux millésimes, deux voyants"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Le taux de retrait courant, année après année, pour le même plan parti en 1929 puis en 1966"))
	legendChips(&b, 74, [][2]string{
		{figBlue, "millésime 1929"},
		{figBad, "millésime 1966"},
	})

	// The band a plan does not come back from, marked before the curves.
	by0, by1 := rate.Map(voyantBandHi), rate.Map(voyantBandLo)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		voyX0, by0, voyX1-voyX0, by1-by0, mixHex("#fffdf9", figBad, 0.10))
	b.WriteString(sTxt(voyX1-6, by0-6, 9.5, figBad, "end", "600",
		"la zone dont on ne revient pas"))

	axisTicks(&b, rate, []float64{0, 4, 8, 12, 16, 20, 24}, 0, " %", voyX0, voyX1, false)
	for y := 0; y < voyantYears; y += 5 {
		b.WriteString(mTxt(year.Map(float64(y)), voyBot+18, 10, figMuted, "middle", "400",
			fmt.Sprint(y)))
	}
	b.WriteString(sTxt(voyX0, voyBot+37, 9.5, figMuted, "start", "400",
		"années depuis le départ en retraite  →"))

	for _, v := range voyantVintages {
		var pts [][2]float64
		for i := 0; i < v.live(); i++ {
			r := v.rate[i]
			if r > voyRateHi {
				// The reading leaves the panel: stop it on the edge, and say
				// so in words rather than compressing everyone else's scale.
				prev := v.rate[i-1]
				t := (voyRateHi - prev) / (r - prev)
				pts = append(pts, [2]float64{
					year.Map(float64(i-1) + t), rate.Map(voyRateHi)})
				break
			}
			pts = append(pts, [2]float64{year.Map(float64(i)), rate.Map(r)})
		}
		b.WriteString(poly(pts, v.color, 2.6, ""))
	}

	// What each replay did, posed on its own curve.
	crash, crashYear := voyantVintages[0].crashPeak()
	b.WriteString(voyantDot(year.Map(3), rate.Map(crash), figBlue))
	// Hung above the band, where neither curve runs, with a leader down to the
	// reading it names: the space just over the 1929 peak is crossed by the
	// 1966 curve a few years later.
	b.WriteString(line(year.Map(3)+13, rate.Map(11.5)+4, year.Map(3)+5, rate.Map(crash)-6, figMuted, 1))
	b.WriteString(sTxt(year.Map(3)+9, rate.Map(11.5), 10, figSoft, "start", "600",
		fmt.Sprintf("le krach : %s %% en %d", frNum(crash, 1), crashYear)))

	top, topYear := voyantVintages[0].peak()
	b.WriteString(voyantDot(year.Map(20), rate.Map(top), figBlue))
	b.WriteString(sTxt(year.Map(20), rate.Map(top)-12, 10, figBlue, "middle", "600",
		fmt.Sprintf("son vrai maximum : %s %% en %d", frNum(top, 1), topYear)))

	cross := voyantVintages[1].crossing()
	ci := cross - voyantVintages[1].start
	b.WriteString(voyantDot(year.Map(float64(ci)), rate.Map(voyantVintages[1].rate[ci]), figBad))
	// Set in the corridor between the two curves, which stays clear from the
	// crossing year to the end of the 1929 replay.
	b.WriteString(sTxt(year.Map(float64(ci))+9, rate.Map(6.2), 10,
		figBad, "start", "600", fmt.Sprintf("franchit 8 %% en %d", cross)))
	b.WriteString(sTxt(year.Map(21), rate.Map(voyRateHi)-8, 10, figBad, "end", "600",
		fmt.Sprintf("sort de l'échelle en 1988, compte vide en %d", voyantVintages[1].ruinYear())))
	b.WriteString(sTxt(voyX1, rate.Map(voyantVintages[0].rate[voyantYears-1])+16, 10,
		figBlue, "end", "600",
		fmt.Sprintf("revenu à %s %% trente ans plus tard",
			frNum(voyantVintages[0].rate[voyantYears-1], 1))))

	b.WriteString(sTxt(24, 380, 10.5, figSoft, "start", "600",
		"Le krach de 1929 fait sonner le voyant trois ans puis le rend ; "+
			"l'épisode de 1966 le fait sonner quinze ans et le garde."))
	b.WriteString(plateFoot(402, []string{
		"Panel Jorda-Schularick-Taylor des États-Unis, rendements réels annuels, 60/40 actions et " +
			"obligations d'État rééquilibré chaque année.",
		"Retrait de 4 % du capital de départ, tenu en pouvoir d'achat, pris en début d'année ; " +
			"voyant = retrait de l'année / capital restant.",
		"Tout est en réel : la déflation des années trente n'apparaît donc pas au numérateur mais dans " +
			"le capital lui-même, qui se refait,",
		"pendant que le retrait indexé baissait en nominal avec les prix. Le millésime 1966, lui, " +
			"épuise son capital en " + fmt.Sprint(voyantVintages[1].ruinYear()) + ".",
		"Le maximum du millésime 1929 ne vient pas du krach mais de l'inflation d'après-guerre : " +
			"c'est encore un épisode de prix, pas un krach.",
	}))
	return svg(640, 480, b.String())
}

// voyantDot marks one reading on a curve.
func voyantDot(x, y float64, color string) string {
	return fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, x, y, color)
}
