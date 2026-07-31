package firebook

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// The all-weather article's second plate: what is inside each recipe, and
// which season each sleeve is there to cover. See figTousTempsSaisons.

// The four seasons of the growth x inflation grid, in the order every row of
// the plate stacks them. The order never changes from one row to the next:
// that is what lets the eye read the family as cousins, and read the 60/40's
// missing seasons as holes.
const (
	saisonProsperite = iota
	saisonInflation
	saisonDeflation
	saisonLiquidite
)

// saisonHatchWash is a PRE-BLENDED solid hex (figGreen #2f9068 @ .18
// composited onto the figure card background #fffdf9), not rgba: crengine
// (KOReader's EPUB SVG renderer) paints rgba fills solid black. It is the
// ground of the hatch, never a fill in its own right.
const saisonHatchWash = "#DAE9DF"

// saisonDef is one season: the colour it owns on the plate, its French name
// and the legend line that names the sleeves it takes in.
type saisonDef struct {
	fill  string
	name  string
	label string
}

var saisonDefs = []saisonDef{
	{figAccent, "prospérité", "prospérité : actions, small value"},
	{figBad, "inflation", "inflation : or, matières premières, trend"},
	{figBlue, "déflation", "déflation : obligations longues, intermédiaires"},
	{figGreen, "récession et liquidité", "récession et liquidité : cash, courtes, volatilité longue"},
}

// saisonSeg is one sleeve of one recipe: the weight the article gives it, in
// percent, and the season it is held for. hatch marks a sleeve that cannot be
// bought (the Dragon's long-volatility pocket).
type saisonSeg struct {
	weight float64
	season int
	hatch  bool
}

// saisonRow is one recipe. segs are listed in the ARTICLE's own order, not in
// season order, so the guard test can compare them one by one with the
// percentages the article's paragraph enumerates; the plate sorts them by
// season at draw time.
type saisonRow struct {
	name string
	sub  string
	segs []saisonSeg
}

// Every weight below is copied verbatim from the article's own paragraphs
// (pkg/firebook/assets/book/fr/portefeuilles-tous-temps.md), and
// TestSaisonsWeightsAreTheArticles re-reads the embedded markdown, extracts
// the percentages of each recipe's sentence and fails when a weight or a
// recipe has left the text.
//
//   - 60/40: the reference rung, 60 % equities and 40 % bonds, the reading of
//     its own name; the sibling plate's ladder (tousTempsLadder) measures the
//     same rung against intermediate Treasuries.
//   - Browne, paragraph "Le Permanent Portfolio de Browne : l'archétype":
//     "25 % actions, 25 % obligations d'État longues, 25 % or, 25 % cash".
//   - All-Weather, paragraph "La lignée": "30 % actions, 40 % obligations
//     longues, 15 % intermédiaires, 7,5 % or et 7,5 % matières premières".
//   - Golden Butterfly, same paragraph: "20 % actions larges, 20 % small-cap
//     value, 20 % obligations longues, 20 % courtes, 20 % or".
//   - Dragon, same paragraph: "24 % actions, 18 % obligations longues, 19 %
//     or, 18 % matières premières et trend, et 21 % de volatilité longue".
//
// The season of each sleeve is an editorial reading, not a measurement. The
// rule, stated on the plate, is Browne's own correspondence: "Les actions
// couvrent la prospérité, les obligations longues la déflation, l'or
// l'inflation et les crises de confiance, le cash la récession et le rachat
// du reste." A sleeve counts for ONE season, the one where it is the first
// winner, so gold never appears twice and cash never appears twice.
var saisonRows = []saisonRow{
	{"60/40", "la référence", []saisonSeg{
		{60, saisonProsperite, false}, // actions
		{40, saisonDeflation, false},  // obligations
	}},
	{"Browne 4 × 25", "1981", []saisonSeg{
		{25, saisonProsperite, false}, // actions
		{25, saisonDeflation, false},  // obligations d'Etat longues
		{25, saisonInflation, false},  // or
		{25, saisonLiquidite, false},  // cash
	}},
	{"All-Weather", "Dalio, version Robbins", []saisonSeg{
		{30, saisonProsperite, false}, // actions
		{40, saisonDeflation, false},  // obligations longues
		{15, saisonDeflation, false},  // intermediates
		{7.5, saisonInflation, false}, // or
		{7.5, saisonInflation, false}, // matieres premieres
	}},
	{"Golden Butterfly", "Portfolio Charts", []saisonSeg{
		{20, saisonProsperite, false}, // actions larges
		{20, saisonProsperite, false}, // small-cap value
		{20, saisonDeflation, false},  // obligations longues
		{20, saisonLiquidite, false},  // obligations courtes
		{20, saisonInflation, false},  // or
	}},
	{"Dragon", "Artemis, 2020", []saisonSeg{
		{24, saisonProsperite, false}, // actions
		{18, saisonDeflation, false},  // obligations longues
		{19, saisonInflation, false},  // or
		{18, saisonInflation, false},  // matieres premieres et trend
		{21, saisonLiquidite, true},   // volatilite longue, non investissable
	}},
}

// saisonsCovered counts the seasons a recipe holds at least one sleeve for.
func (r saisonRow) saisonsCovered() int {
	seen := map[int]bool{}
	for _, s := range r.segs {
		seen[s.season] = true
	}
	return len(seen)
}

// saisonTotal sums a recipe's weights, in percent.
func (r saisonRow) saisonTotal() float64 {
	t := 0.0
	for _, s := range r.segs {
		t += s.weight
	}
	return t
}

// saisonWeight formats a weight the French way, without a needless decimal.
func saisonWeight(v float64) string {
	if v == math.Trunc(v) {
		return frNum(v, 0)
	}
	return frNum(v, 1)
}

// hatchRect fills a rectangle with a pale wash, rules it with 45-degree
// hairlines and outlines it: the plate's mark for a sleeve nobody can buy.
// The hairlines are clipped by arithmetic rather than by a clipPath, which
// crengine (KOReader's EPUB SVG renderer) does not honour.
func hatchRect(x, y, w, h float64, wash, ink string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`, x, y, w, h, wash)
	for d := -h; d < w; d += 5 {
		lo, hi := math.Max(0, -d), math.Min(h, w-d)
		if lo >= hi {
			continue
		}
		b.WriteString(line(x+d+lo, y+h-lo, x+d+hi, y+h-hi, ink, 1))
	}
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="%s" stroke-width="1"/>`, x, y, w, h, ink)
	return b.String()
}

// figTousTempsSaisons draws the family's shared grid: one stacked bar per
// recipe, the segments coloured by the SEASON each sleeve covers rather than
// by asset class, the seasons always in the same order. The four all-weather
// recipes then read as dosages of one idea, and the 60/40 reference line
// shows the two seasons it simply does not hold anything for.
func figTousTempsSaisons() string {
	const (
		x0, x1 = 172.0, 544.0 // the 0 % and 100 % ends of every bar
		y0     = 110.0        // top of the first bar
		pitch  = 52.0
		bh     = 25.0
		gap    = 2.0 // surface gap between two sleeves
	)
	x := func(pct float64) float64 { return x0 + pct*(x1-x0)/100 }
	var b strings.Builder
	b.WriteString(plateHead("portefeuilles tous-temps", "Tous cousins sur la même grille, et le trou du 60/40"))

	// legend: two rows of two chips, each naming a season and its sleeves
	for i, d := range saisonDefs {
		lx, ly := 24.0, 56.0
		if i%2 == 1 {
			lx = 296
		}
		if i >= 2 {
			ly = 74
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="10" height="10" rx="2.5" fill="%s"/>`, lx, ly, d.fill)
		b.WriteString(sTxt(lx+15, ly+9, 10, figSoft, "start", "400", d.label))
	}

	// the percent scale, read as a column header above the bars
	bot := y0 + pitch*float64(len(saisonRows)-1) + bh
	for _, g := range []float64{25, 50, 75} {
		b.WriteString(line(x(g), y0-6, x(g), bot+6, figGrid, 1))
	}
	for _, g := range []float64{0, 25, 50, 75} {
		b.WriteString(mTxt(x(g), 102, 9.5, figMuted, "middle", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(mTxt(x(100), 102, 9.5, figMuted, "middle", "400", "100 %"))
	b.WriteString(sTxt(616, 90, 9.5, figMuted, "end", "400", "saisons"))
	b.WriteString(sTxt(616, 102, 9.5, figMuted, "end", "400", "couvertes"))

	for i, r := range saisonRows {
		y := y0 + pitch*float64(i)
		b.WriteString(sTxt(x0-14, y+11, 11.5, figSoft, "end", "600", r.name))
		b.WriteString(sTxt(x0-14, y+24, 10, figMuted, "end", "400", r.sub))

		// sleeves, sorted into the plate's fixed season order
		segs := append([]saisonSeg(nil), r.segs...)
		sort.SliceStable(segs, func(a, c int) bool { return segs[a].season < segs[c].season })
		acc := 0.0
		for _, s := range segs {
			sx, sw := x(acc), x(acc+s.weight)-x(acc)-gap
			lbl := saisonWeight(s.weight)
			if s.hatch {
				b.WriteString(hatchRect(sx, y, sw, bh, saisonHatchWash, saisonDefs[s.season].fill))
				// the hatch would swallow the number: clear a slot for it
				fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="22" height="14" fill="%s"/>`,
					sx+sw/2-11, y+bh/2-7, saisonHatchWash)
				b.WriteString(mTxt(sx+sw/2, y+bh/2+3.5, 9.5, figInk, "middle", "600", lbl))
			} else {
				b.WriteString(barH(sx, sx+sw, y, bh, saisonDefs[s.season].fill))
				b.WriteString(mTxt(sx+sw/2, y+bh/2+3.5, 9.5, "#fffdf9", "middle", "600", lbl))
			}
			acc += s.weight
		}

		// the tally: how many of the four seasons this recipe holds
		n := r.saisonsCovered()
		col := figDeep
		if n < 4 {
			col = figBad
		}
		b.WriteString(mTxt(616, y+bh/2+4, 12, col, "end", "600", fmt.Sprintf("%d / 4", n)))
	}

	// the reference line's two holes, said where they are missing
	b.WriteString(sTxt(x0, y0+bh+14, 10.5, figBad, "start", "600",
		"rien contre l'inflation, rien pour la liquidité : deux saisons sur quatre"))
	// and the Dragon's sleeve that cannot be bought
	b.WriteString(sTxt(x0, bot+14, 10.5, figSoft, "start", "600",
		"hachures : la volatilité longue n'existe pas en format UCITS accessible"))

	for i, ln := range []string{
		"Poids : ceux des paragraphes de l'article. Saisons : lecture éditoriale de la correspondance de Browne, un actif ne comptant",
		"que pour une seule saison, celle où il est le premier gagnant. L'or couvre aussi les crises de confiance et le cash sert aussi",
		"de munition au rééquilibrage : ni l'un ni l'autre n'est compté deux fois.",
	} {
		b.WriteString(sTxt(24, bot+40+14*float64(i), 9.5, figMuted, "start", "400", ln))
	}
	return svg(640, 420, b.String())
}
