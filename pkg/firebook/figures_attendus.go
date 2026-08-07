package firebook

import (
	"fmt"
	"math"
	"strings"
)

// Plates of "rendements-attendus": how a forward-looking expectation is built
// brick by brick, and what the recommended withdrawal rate does when the
// conditions that feed it move.

// --- The building blocks of an expected real return ---

// briqueSeg is one segment of a row: a brick of the main bar, or a piece of the
// valuation term drawn on the sub-row under it.
type briqueSeg struct {
	from, to float64 // percent of expected real return per year
	fill     string
	val      string // mono label, French decimal comma
}

// briqueRow is one asset class. The valuation term is held as a range in
// points, not as pixels: the sub-row segments and the row's total are both
// derived from it, so the drawing cannot disagree with the arithmetic.
type briqueRow struct {
	name, detail   string
	bricks         []briqueSeg // observable bricks, laid left to right from 0
	socle          float64     // where the observable bricks end, 0 = none
	hasTerm        bool
	termLo, termHi float64    // the valuation term, in points of real return
	negVal, posVal string     // mono labels for its negative and positive pieces
	hollow         [2]float64 // dashed "assumption" box on the main bar, zero = none
	note           string     // free note drawn past the right end of the row
	total          string     // what the totals column prints
}

// extent is the rightmost value the row reaches on the value axis.
func (r briqueRow) extent() float64 {
	switch {
	case r.hasTerm:
		return r.socle + math.Max(r.termHi, 0)
	case r.hollow[1] > r.hollow[0]:
		return r.hollow[1]
	default:
		return r.socle
	}
}

// termSegs cuts the valuation term into the pieces the sub-row draws: what it
// can take away, left of the socle, and what it can add, right of it.
func (r briqueRow) termSegs() []briqueSeg {
	if !r.hasTerm {
		return nil
	}
	var out []briqueSeg
	if r.termLo < 0 {
		out = append(out, briqueSeg{r.socle + r.termLo, r.socle, figBad, r.negVal})
	}
	if r.termHi > 0 {
		out = append(out, briqueSeg{r.socle, r.socle + r.termHi, figGreen, r.posVal})
	}
	return out
}

// briqueRows is the article's own arithmetic, class by class
// (rendements-attendus, section "Comment se fabrique une prévision : les
// briques"). Every number is quoted from that prose, and
// figures_attendus_test.go redoes each addition:
//
//	Actions US        2,25 + 1,75 = 4,00 ; valorisation 0 à −1,50     -> 2,50 à 4,00
//	Actions hors US   3,00 + 1,75 = 4,75 ; valorisation −0,75 à +1,25 -> 4,00 à 6,00
//	Obligations euro  YTM 3,20 − inflation anticipée 2,00             -> 1,20
//	Monétaire euro    taux directeur réel, ~0,50                      -> 0 à 1
//	Or                aucune brique                                   -> 0 à 1
//
// The article states the non-US TOTAL (4 à 6 % réel) rather than its valuation
// term, so that term is the residual of the two bricks it does state
// (distribution ~3 % for Europe, real earnings growth 1,75 %). Putting the whole
// width of every range in that one term is the article's own claim: it names the
// valuation term as the only disputed one. The plate says so in its footer.
func briqueRows() []briqueRow {
	return []briqueRow{
		{
			name: "Actions US", detail: "S&amp;P 500, conditions 2024-2026",
			bricks: []briqueSeg{
				{0, 2.25, figAccent, "2,25"},
				{2.25, 4.00, figBlue, "1,75"},
			},
			socle: 4.00, hasTerm: true, termLo: -1.50, termHi: 0, negVal: "0 à −1,5",
			note:  "terme de valorisation",
			total: "2,5 à 4,0 %",
		},
		{
			name: "Actions hors US", detail: "Europe, Japon, émergents",
			bricks: []briqueSeg{
				{0, 3.00, figAccent, "3,00"},
				{3.00, 4.75, figBlue, "1,75"},
			},
			socle: 4.75, hasTerm: true, termLo: -0.75, termHi: 1.25,
			negVal: "−0,75", posVal: "+1,25",
			total: "4,0 à 6,0 %",
		},
		{
			name: "Obligations euro", detail: "OAT ou Bund 10 ans, nominal",
			bricks: []briqueSeg{{0, 3.20, figAccent, "3,20"}},
			socle:  3.20, hasTerm: true, termLo: -2.00, termHi: -2.00, negVal: "−2,0",
			note:  "inflation anticipée",
			total: "1,2 %",
		},
		{
			name: "Monétaire euro", detail: "taux directeur réel",
			bricks: []briqueSeg{{0, 0.50, figAccent, "0,5"}},
			hollow: [2]float64{0, 1.00},
			note:   "une seule brique, et elle suit le cycle",
			total:  "0 à 1 %",
		},
		{
			name: "Or", detail: "ni coupon ni bénéfices",
			hollow: [2]float64{0, 1.00},
			note:   "rien à empiler : le chiffre est une hypothèse, pas une lecture",
			total:  "0 à 1 %",
		},
	}
}

// briqueRect draws one brick. Segments are inset on the right so a stack reads
// as separate blocks laid end to end rather than as one long bar.
func briqueRect(x0, x1, y, h float64, fill string) string {
	w := x1 - x0 - 1.5
	if w < 1 {
		w = 1
	}
	return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="2.5" fill="%s"/>`, x0, y, w, h, fill)
}

// briqueVal prints a brick's value inside it when it fits, next to it otherwise.
func briqueVal(x0, x1, baseline, size float64, val string) string {
	if x1-x0 >= 36 {
		return mTxt((x0+x1)/2, baseline, size, "#fffdf9", "middle", "600", val)
	}
	return mTxt(x1+5, baseline, size, figSoft, "start", "600", val)
}

func figBriquesEsperance() string {
	const (
		px0, px1 = 176.0, 508.0 // pixels of value 0 and of vMax
		vMax     = 6.5
		colTot   = 616.0 // right edge of the totals column
		rowTop   = 100.0
		pitch    = 58.0
		barH     = 18.0
		subDy    = 21.0
		subH     = 13.0
	)
	xv := func(v float64) float64 { return px0 + v/vMax*(px1-px0) }

	var b strings.Builder
	b.WriteString(plateHead("rendements attendus",
		"L'espérance d'une obligation est affichée, celle d'une action se construit"))

	for g := 0.0; g <= 6.0; g++ {
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(xv(g), 88, xv(g), 372, col, 1))
		if int(g)%2 == 0 {
			b.WriteString(mTxt(xv(g), 80, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f %%", g)))
		}
	}
	b.WriteString(sTxt(24, 80, 9.5, figMuted, "start", "600", "classe d'actifs"))
	b.WriteString(sTxt(colTot, 80, 9.5, figMuted, "end", "600", "espérance réelle"))

	for i, r := range briqueRows() {
		top := rowTop + pitch*float64(i)
		b.WriteString(sTxt(24, top+12, 11.5, figSoft, "start", "600", r.name))
		b.WriteString(sTxt(24, top+28, 9.5, figMuted, "start", "400", r.detail))

		if r.hollow[1] > r.hollow[0] {
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="3" fill="none" stroke="%s" stroke-width="1.2" stroke-dasharray="4 3"/>`,
				xv(r.hollow[0]), top, xv(r.hollow[1])-xv(r.hollow[0]), barH, figMuted)
		}
		for _, s := range r.bricks {
			b.WriteString(briqueRect(xv(s.from), xv(s.to), top, barH, s.fill))
			b.WriteString(briqueVal(xv(s.from), xv(s.to), top+12.5, 9.5, s.val))
		}
		for _, s := range r.termSegs() {
			b.WriteString(briqueRect(xv(s.from), xv(s.to), top+subDy, subH, s.fill))
			b.WriteString(briqueVal(xv(s.from), xv(s.to), top+subDy+9.5, 9, s.val))
		}
		if r.note != "" {
			ny := top + 13
			if r.hasTerm {
				ny = top + subDy + 10
			}
			b.WriteString(sTxt(xv(r.extent())+10, ny, 10, figSoft, "start", "400", r.note))
		}
		// carets under each end of the expectation range: the total the row
		// argues for, marked where it is actually read off the axis.
		lo, hi := r.socle+r.termLo, r.socle+r.termHi
		caretY := top + subDy + subH + 4
		if !r.hasTerm {
			lo, hi = r.hollow[0], r.hollow[1]
			caretY = top + barH + 5
		}
		for _, v := range []float64{lo, hi} {
			fmt.Fprintf(&b, `<path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f Z" fill="%s"/>`,
				xv(v), caretY, xv(v)-3.4, caretY+5, xv(v)+3.4, caretY+5, figInk)
			if hi == lo {
				break
			}
		}
		b.WriteString(dashLine(xv(r.extent())+8, top+9, 538, top+9, figGrid, 1, "2 4"))
		b.WriteString(mTxt(colTot, top+13, 11.5, figInk, "end", "600", r.total))
	}

	// the first row teaches the grammar, brick by brick
	b.WriteString(sTxt((xv(0)+xv(2.25))/2, 94, 9.5, figMuted, "middle", "400", "distribution"))
	b.WriteString(sTxt((xv(2.25)+xv(4.0))/2, 94, 9.5, figMuted, "middle", "400", "croissance des bénéfices"))

	b.WriteString(sTxt(24, 392, 10, figMuted, "start", "400",
		"Briques pleines : ce que les prix affichent déjà. Sous la barre, le terme de valorisation, le seul disputé, qui porte toute la fourchette."))
	return svg(640, 406, b.String())
}

// --- Morningstar's staircase, and the conditions under it ---

// morningstarSWR is the base-case starting safe withdrawal rate of "The State
// of Retirement Income", vintage by vintage, in percent: a 30-year horizon, a
// 90 % success rate and a balanced portfolio, rebuilt every year from the
// firm's own forward-looking capital market assumptions.
//
// Source, verified 2026-07-30: "The State of Retirement Income: 2025"
// (Morningstar, published 2 December 2025, data as of 30 September 2025).
// Its key-takeaways page states 3,9 % for the 2025 vintage, "up slightly from
// the starting safe withdrawal percentage of 3.7% we estimated in last year's
// report", and recalls in the same breath that the base cases "were 3.3% in
// 2021, 3.8% in 2022, and 4.0% in 2023". The vintage labels are Morningstar's
// own (the 2025 edition is the one the press reports as the rate "for 2026").
var morningstarSWR = []float64{3.3, 3.8, 4.0, 3.7, 3.9}

// morningstarCAPE is the Shiller CAPE (PE10) of September of each vintage,
// which is the data cutoff every edition uses ("Data as of Sept. 30").
// Frozen from datasets.CAPE(), rows 2021-09-01 to 2025-09-01;
// figures_attendus_test.go reads them back from the bundled series.
var morningstarCAPE = []float64{37.62, 28.23, 30.81, 35.70, 38.59}

// morningstarWhy annotates each step. The first four reasons are Morningstar's
// own "temperature check" wording (2021: "bond yields were ultralow, and equity
// valuations were high"; 2023: "thanks in large part to higher fixed-income
// yields/return prospects and moderating inflation"; 2024 and 2025: "equity
// valuations aren't inexpensive, but bond yields are a plus"). The last is the
// 2025 edition's own explanation of its rise: not the markets but a change of
// method, capital market assumptions now blending top-down inputs with the
// bottom-up views of Morningstar's equity analysts.
var morningstarWhy = [][2]string{
	{"taux nuls,", "actions chères"},
	{"valorisations", "dégonflées"},
	{"taux obligataires", "restaurés"},
	{"actions", "redevenues chères"},
	{"méthode revue,", "pas les marchés"},
}

// morningstarYears labels the columns.
var morningstarYears = []string{"2021", "2022", "2023", "2024", "2025"}

// capeTint composites figDeep onto the figure card background #fffdf9 at an
// opacity that grows with the CAPE, and returns a SOLID hex: crengine
// (KOReader's EPUB renderer) paints rgba fills black, so no band cell may carry
// one. 28 maps to a barely tinted cell, 39 to a firm one.
func capeTint(cape float64) string {
	a := 0.10 + 0.34*(cape-28)/(39-28)
	if a < 0.06 {
		a = 0.06
	}
	if a > 0.46 {
		a = 0.46
	}
	bg := [3]float64{255, 253, 249}
	fg := [3]float64{138, 85, 38} // figDeep #8a5526
	var out [3]int
	for i := range bg {
		out[i] = int(bg[i] + (fg[i]-bg[i])*a + 0.5)
	}
	return fmt.Sprintf("#%02X%02X%02X", out[0], out[1], out[2])
}

func figEscalierMorningstar() string {
	const (
		x0, x1     = 76.0, 612.0
		yTop, yBot = 82.0, 238.0
		rLo, rHi   = 2.90, 4.25
		bandTop    = 304.0
		bandH      = 34.0
	)
	n := float64(len(morningstarSWR))
	colW := (x1 - x0) / n
	yr := func(v float64) float64 { return yBot - (v-rLo)/(rHi-rLo)*(yBot-yTop) }

	var b strings.Builder
	b.WriteString(plateHead("rendements attendus",
		"Le taux « sûr » de Morningstar bouge avec les conditions d'entrée"))
	b.WriteString(sTxt(24, 70, 10, figMuted, "start", "400",
		"taux de retrait initial recommandé, en % (30 ans, 90 % de succès)"))

	// the book's own working range, as a wash behind everything
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, yr(3.5), x1-x0, yr(3.0)-yr(3.5), figWash)
	b.WriteString(sTxt(x1-6, yr(3.12), 10, figMuted, "end", "400", "la fourchette de travail du livre : 3 à 3,5 %"))

	for _, g := range []float64{3.0, 3.5, 4.0} {
		b.WriteString(line(x0, yr(g), x1, yr(g), figGrid, 1))
		b.WriteString(mTxt(x0-8, yr(g)+3.5, 10, figMuted, "end", "400",
			strings.Replace(fmt.Sprintf("%.1f", g), ".", ",", 1)))
	}
	b.WriteString(dashLine(x0, yr(4.0), x1, yr(4.0), figMuted, 1.2, "5 4"))
	b.WriteString(sTxt(x0+6, yr(4.0)-7, 10, figMuted, "start", "400", "la règle des 4 %"))

	// the treads, joined by thin risers
	for i, v := range morningstarSWR {
		cx0 := x0 + colW*float64(i)
		ty := yr(v)
		b.WriteString(poly([][2]float64{{cx0 + 9, ty}, {cx0 + colW - 9, ty}}, figAccent, 4.5, ""))
		// every value hangs UNDER its tread: 3,9 sits a tenth of a point below
		// the 4 % reference, and a label above it would cross that line.
		b.WriteString(mTxt(cx0+colW/2, ty+16, 12, figInk, "middle", "600",
			strings.Replace(fmt.Sprintf("%.1f %%", v), ".", ",", 1)))
		if i > 0 {
			b.WriteString(dashLine(cx0, yr(morningstarSWR[i-1]), cx0, ty, figMuted, 1, "3 3"))
		}
		b.WriteString(sTxt(cx0+colW/2, 262, 10, figSoft, "middle", "600", morningstarWhy[i][0]))
		b.WriteString(sTxt(cx0+colW/2, 275, 10, figMuted, "middle", "400", morningstarWhy[i][1]))
	}

	// the 2025 counterfactual: same markets, the method of the year before
	last := x0 + colW*(n-1)
	b.WriteString(dashLine(last+9, yr(3.6), last+colW-9, yr(3.6), figMuted, 1.6, "4 3"))
	b.WriteString(mTxt(last+colW-9, yr(3.6)+14, 9.5, figMuted, "end", "600", "3,6"))

	// the conditioning band: the CAPE each edition worked from
	b.WriteString(sTxt(x0, 297, 9.5, figMuted, "start", "400",
		"conditions d'entrée : le CAPE de Shiller au 30 septembre, date d'arrêté de chaque édition"))
	for i, c := range morningstarCAPE {
		cx0 := x0 + colW*float64(i)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
			cx0+1.5, bandTop, colW-3, bandH, capeTint(c))
		b.WriteString(mTxt(cx0+colW/2, bandTop+22, 12, figInk, "middle", "600",
			strings.Replace(fmt.Sprintf("%.1f", c), ".", ",", 1)))
		b.WriteString(mTxt(cx0+colW/2, 358, 11, figSoft, "middle", "600", morningstarYears[i]))
	}

	b.WriteString(sTxt(24, 380, 10, figMuted, "start", "400",
		"Le CAPE n'est qu'une des deux conditions : 2023 monte grâce aux taux obligataires restaurés, pas aux actions."))
	b.WriteString(sTxt(24, 394, 10, figMuted, "start", "400",
		"2025 monte pour une autre raison encore : à méthode 2024 inchangée, le 50/50 sortait à 3,6 % (tireté)."))
	return svg(640, 408, b.String())
}
