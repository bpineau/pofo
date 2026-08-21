package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The crisis-report plate of the defensive-assets hub. The article reviews the
// candidates one by one and says of each one which regime it serves; the
// argument it builds from that review is a negative one, that NO asset defends
// against everything. A list of verdicts cannot show it, because the reader
// meets the candidates one at a time and never sees them side by side in the
// same crisis.
//
// Drawn as a signed grid, candidates by episodes, the argument is the picture:
// read down a column and one cell stands out, read across the row of that
// winner and it collapses somewhere else. Three defenders share the four
// framed cells and not one of them holds all four, which is the whole thesis.
//
// Every number is a REAL total return in dollars over the episode's own dates,
// measured on the bundled series and deflated by the US CPI, frozen here and
// recomputed by the guard test under frozenAgainstData.

// The four episodes, DATED BEFORE any of them was measured. Each one is a
// different enemy of a retirement plan, which is what makes the row of winners
// change:
//
//   - 1973-1974: the inflationary bear market, the reference vintage of the
//     book's whole sequence-risk argument. Measured over the two calendar
//     years, from the last quote of 1972 to the last of 1974, because the
//     episode is a regime and not a crash.
//   - 2008: the growth crash and deflation scare of the financial crisis.
//     Calendar 2008, which is also the window the article's own figures for
//     long bonds (+25 to +30 %) and trend (+21 %) quote.
//   - March 2020: the covid crash, measured from the S&P 500's record close of
//     19 February 2020 to its trough close of 23 March 2020. A calendar year
//     would hide it entirely: the index ended 2020 up. This is the V-shaped
//     reversal the article names as the trend follower's killing regime.
//   - 2022: the inflation and rate shock, calendar year, the window the
//     article quotes for long bonds (-30 to -40 %) and trend (+27 %).
var defensesEpisodes = []defensesEpisode{
	{"1973-1974", "31/12/1972 au 31/12/1974", "1972-12-31", "1974-12-31",
		[2]string{"inflation, régime", "baissier long"}},
	{"2008", "31/12/2007 au 31/12/2008", "2007-12-31", "2008-12-31",
		[2]string{"krach de croissance,", "peur de déflation"}},
	{"Mars 2020", "19/02/2020 au 23/03/2020", "2020-02-19", "2020-03-23",
		[2]string{"krach éclair,", "retournement en V"}},
	{"2022", "31/12/2021 au 31/12/2022", "2021-12-31", "2022-12-31",
		[2]string{"choc d'inflation", "et de taux"}},
}

// defensesEpisode is one column: its title, its dates in the legend's form,
// the bounds the measurement uses, and the enemy it stands for, set on the two
// lines the column is wide enough for.
type defensesEpisode struct {
	label, dates string
	from, to     string
	enemy        [2]string
}

// defensesRow is one candidate: the name the plate prints, the bundled series
// it is measured on, and the first date that series carries. A candidate whose
// series starts after an episode gets a hatched cell there, never an
// extrapolated number.
type defensesRow struct {
	name  string
	id    string
	ref   bool // in refdata rather than simdata
	since string
}

// The candidates, in the order of the article's review, with equities on top as
// the thing being defended rather than a defender. Credit and listed real
// estate are deliberately absent: no bundled series carries them, and the
// article files their report card with the false defensives.
var defensesRows = []defensesRow{
	{"Actions US (S&amp;P 500)", "SP500", false, "1871-01-31"},
	{"Cash (bons du Trésor 3 mois)", "TBILL-3M", true, "1934-01-01"},
	{"Obligations d'État 7-10 ans", "IEF", false, "1962-01-01"},
	{"Obligations d'État 20 ans et +", "TLT", false, "1962-01-01"},
	{"Or", "XAUUSD", false, "1968-04-01"},
	{"Suivi de tendance (BTOP50)", "BTOP50", false, "1986-12-31"},
}

// defensesHas says whether a candidate's series reaches an episode. It is the
// hatching rule, derived from the data rather than declared: the trend index
// starts at the end of 1986 and therefore has nothing to say about 1973-1974.
func defensesHas(row, ep int) bool {
	return defensesRows[row].since <= defensesEpisodes[ep].from
}

// The measured grid, in percent of real dollar total return over each episode.
// Cells the series does not reach hold zero and are never drawn.
var defensesGrid = [6][4]float64{
	{-48.76, -37.06, -33.64, -23.08}, // actions
	{-5.03, 1.29, 0.28, -4.15},       // cash
	{-11.39, 17.81, 6.61, -20.30},    // 7-10 ans
	{-15.53, 33.83, 14.48, -35.40},   // 20 ans et plus
	{135.32, 5.74, -2.31, -6.46},     // or
	{0, 13.47, -4.56, 7.92},          // suivi de tendance
}

// defensesGoldEUR is what the dollar's own 2022 rise gave back to a European
// holder of an asset priced in dollars, in points: the euro fell from 1.1325 to
// 1.0661 over the year. The plate's grid is measured in real dollars, and the
// article's verdict on gold in 2022 ("honorable en euros") is only readable
// with this correction, so the note carries it and the guard test recomputes
// it from the bundled exchange rate.
const defensesGoldEUR = 6.23

// defensesBest is the index of the best candidate of a column, equities
// excluded: the plate frames a DEFENDER, and the row it frames changes every
// time, which is the plate's whole point.
func defensesBest(ep int) int {
	best := -1
	for i := 1; i < len(defensesRows); i++ {
		if !defensesHas(i, ep) {
			continue
		}
		if best < 0 || defensesGrid[i][ep] > defensesGrid[best][ep] {
			best = i
		}
	}
	return best
}

// defensesCell sets one reading: a signed whole percent, the book's
// typographic minus, and no sign at all on a zero, because "+0 %" would claim a
// precision the yearly resolution does not have.
func defensesCell(v float64) string {
	n := math.Round(v)
	switch {
	case n > 0:
		return fmt.Sprintf("+%.0f %%", n)
	case n < 0:
		return fmt.Sprintf("%s%.0f %%", figMinus, -n)
	}
	return "0 %"
}

// defensesFill is the cell's ground: the figure card tinted toward green or
// red, by an amount that saturates at forty points so the seventies' gold does
// not flatten every other reading into a pale wash. The ink stays the book's,
// which is why the tint is capped well short of the pure colour.
func defensesFill(v float64) string {
	col := figGreen
	if v < 0 {
		col = figBad
	}
	t := math.Abs(v) / 40
	if t > 1 {
		t = 1
	}
	return mixHex("#fffdf9", col, 0.10+0.48*t)
}

// The plate's geometry: one label column and four episode columns, one band
// per candidate, with the equities band held apart above the defenders by a
// hairline because it is the thing being defended, not a defender.
const (
	defGridX = 220.0 // the first cell's left edge
	defColW  = 100.0 // one column, cell + gutter
	defCellW = 96.0
	defRowY  = 130.0 // the equities band's top
	defRowH  = 30.0
	defRowG  = 5.0
	defSplit = 12.0 // the extra air under the equities band
)

// defensesRowY is the top edge of a candidate's band.
func defensesRowY(i int) float64 {
	y := defRowY + float64(i)*(defRowH+defRowG)
	if i > 0 {
		y += defSplit
	}
	return y
}

func figDefensesBulletin() string {
	var b strings.Builder
	b.WriteString(plateHead("le bulletin de crise", "Aucun actif ne défend contre tout"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"Ce que chaque candidat a rendu, en dollars réels, pendant quatre crises de nature différente"))

	// The column heads: the episode, then the enemy it stands for.
	for j, ep := range defensesEpisodes {
		cx := defGridX + float64(j)*defColW + defCellW/2
		b.WriteString(sTxt(cx, 92, 11.5, figInk, "middle", "600", ep.label))
		b.WriteString(sTxt(cx, 106, 9, figMuted, "middle", "400", ep.enemy[0]))
		b.WriteString(sTxt(cx, 117, 9, figMuted, "middle", "400", ep.enemy[1]))
	}

	// The bands, one per candidate.
	for i, row := range defensesRows {
		y := defensesRowY(i)
		mid := y + defRowH/2
		name, weight := figSoft, "600"
		if i == 0 {
			name, weight = figMuted, "400"
			b.WriteString(sTxt(defGridX-14, mid-1, 10.5, name, "end", weight, row.name))
			b.WriteString(sTxt(defGridX-14, mid+11, 8.5, figMuted, "end", "400",
				"ce que la défense doit protéger"))
		} else {
			b.WriteString(sTxt(defGridX-14, mid+3.5, 10.5, name, "end", weight, row.name))
		}
		for j := range defensesEpisodes {
			x := defGridX + float64(j)*defColW
			if !defensesHas(i, j) {
				b.WriteString(defensesHatch(x, y))
				continue
			}
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="4" fill="%s"/>`,
				x, y, defCellW, defRowH, defensesFill(defensesGrid[i][j]))
			b.WriteString(mTxt(x+defCellW/2, mid+4.5, 12.5, figInk, "middle", "600",
				defensesCell(defensesGrid[i][j])))
		}
	}
	// The hairline that holds the engine apart from its defenders.
	sep := defensesRowY(0) + defRowH + defSplit/2
	b.WriteString(line(24, sep, 616, sep, figRule, 1))

	// The framed winner of every column, drawn last so no cell ground covers it.
	for j := range defensesEpisodes {
		best := defensesBest(j)
		fmt.Fprintf(&b, `<rect class="win" x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="4" `+
			`fill="none" stroke="%s" stroke-width="2"/>`,
			defGridX+float64(j)*defColW-1, defensesRowY(best)-1, defCellW+2, defRowH+2, figDeep)
	}

	bottom := defensesRowY(len(defensesRows)-1) + defRowH
	b.WriteString(sTxt(24, bottom+26, 10.5, figSoft, "start", "600",
		"Trois défenseurs se partagent quatre crises, et pas un ne les tient toutes : "+
			"chacun s'effondre dans une autre colonne."))
	b.WriteString(plateFoot(bottom+48, []string{
		"Rendement total réel en dollars sur la fenêtre, déflaté par l'IPC américain. " +
			"Le cadre marque le meilleur défenseur de la colonne.",
		"Fenêtres datées avant toute mesure : 1973-1974 = " + defensesEpisodes[0].dates +
			" ; 2008 = " + defensesEpisodes[1].dates + " ;",
		"mars 2020 = " + defensesEpisodes[2].dates + " (record puis creux du S&amp;P 500) ; 2022 = " +
			defensesEpisodes[3].dates + ".",
		"Séries : S&amp;P 500 total return, bons du Trésor 3 mois, Treasuries 7-10 ans et 20 ans et plus, " +
			"or au comptant, indice BTOP50 net.",
		"Le BTOP50 commence fin 1986 (case vide, jamais extrapolée) ; indice diversifié, " +
			"moins vif que le trend pur du texte.",
		"L'or perd " + frNum(-defensesGrid[4][3], 0) + " % en dollars réels en 2022, mais la hausse du dollar rend " +
			frNum(defensesGoldEUR, 1) + " points à un porteur en euros.",
	}))
	return svg(640, int(bottom+48+15*6+8), b.String())
}

// defensesHatch draws the cell of a candidate whose series does not reach the
// episode: hairline diagonals and the words, never a number.
func defensesHatch(x, y float64) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="4" fill="#fffdf9"/>`,
		x, y, defCellW, defRowH)
	for d := -defRowH; d < defCellW; d += 9 {
		x0, y0 := x+d, y+defRowH
		x1, y1 := x+d+defRowH, y
		if x0 < x {
			y0, x0 = y+defRowH-(x-x0), x
		}
		if x1 > x+defCellW {
			y1, x1 = y+(x1-x-defCellW), x+defCellW
		}
		b.WriteString(line(x0, y0, x1, y1, figGrid, 1))
	}
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="62" height="13" rx="3" fill="#fffdf9"/>`,
		x+defCellW/2-31, y+defRowH/2-7)
	b.WriteString(sTxt(x+defCellW/2, y+defRowH/2+3, 8.5, figMuted, "middle", "400", "pas de série"))
	return b.String()
}
