package firebook

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// The plate of the withdrawal-psychology article: the terminal-wealth
// distribution the article prescribes to chronic under-spenders, drawn for the
// two rates the article names, the one the plan authorises (4 %) and the one
// the under-spender actually takes (3 %).
//
// It is deliberately a DISTRIBUTION, not a history: the vintages are ranked
// from the worst outcome to the best, so the reader looks up where a world
// stands among the others. The article's neighbours draw the same record as a
// history (bengen-millesimes, a scatter over the start year, forty years) or as
// paths (millesimes-1966-1982), which is the opposite reading.
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed, before tax.

// The estate the book's reference household left after thirty years, for every
// complete thirty-year window of the bundled record, in multiples of the
// starting stake. One million euros of capital, Bengen's fixed real withdrawal
// (indexed to inflation, never adjusted), on the bundled real US 60/40
// (S&P 500 total return + 5-year Treasuries, rebalanced every January, deflated
// by CPI-U). Zero means the money ran out.
//
// mourirRicheSpent is the plan's full permission, 40 000 EUR a year (4 %);
// mourirRicheThrift is the chronic under-spender of the article, 30 000 EUR a
// year (3 %). Vintage order, from mourirRicheFirst.
//
// Reproduce with pkg/replay, for spend = 40000 then 30000, and start running
// from 1954 to 1996 (the last complete thirty-year window of the record):
//
//	res, _ := replay.Run(replay.Setup{Start: start, Capital: 1e6, Spend: spend,
//	    Years: 30, Mu: 0.045, Sigma: 0.10, Df: 5, TargetRuin: 0.05, RaiseCap: 1.5})
//	multiple := res.Rules[0].Final / 1e6 // Rules[0] is the fixed rule
//
// figures_psycho_test.go recomputes every value, both medians and all four
// counts, and fails on any drift.
var (
	mourirRicheFirst = 1954
	mourirRicheSpent = []float64{
		1.538, 0.841, 0.616, 0.772, 1.097, 0.574, 0.587, 0.570, 0.266, 0.522,
		0.254, 0.037, 0.000, 0.364, 0.091, 0.086, 0.981, 0.930, 0.657, 0.351,
		1.433, 3.469, 2.593, 2.172, 3.178, 3.104, 3.792, 3.888, 4.603, 3.909,
		4.010, 4.199, 3.080, 2.547, 3.074, 2.739, 2.436, 3.182, 2.469, 1.968,
		2.055, 2.685, 1.919,
	}
	mourirRicheThrift = []float64{
		1.989, 1.293, 1.149, 1.371, 1.670, 1.165, 1.265, 1.199, 1.001, 1.264,
		1.022, 0.754, 0.817, 1.320, 1.210, 1.383, 2.346, 2.180, 1.797, 1.323,
		2.521, 4.552, 3.600, 3.206, 4.185, 3.886, 4.611, 4.712, 5.373, 4.672,
		4.828, 5.021, 3.837, 3.285, 3.844, 3.429, 3.200, 3.990, 3.282, 2.585,
		2.715, 3.388, 2.647,
	}
)

// mourirRicheMedian is the middle vintage of a series (the sample is odd-sized,
// so the median is a value the record actually produced, not an average of two).
func mourirRicheMedian(vs []float64) float64 {
	s := append([]float64(nil), vs...)
	sort.Float64s(s)
	return s[len(s)/2]
}

// mourirRicheAboveStake counts the vintages that left more than the capital
// they started with: the frontier the plate draws in bold.
func mourirRicheAboveStake(vs []float64) int {
	n := 0
	for _, v := range vs {
		if v > 1 {
			n++
		}
	}
	return n
}

// mourirRicheAtLeast counts the vintages that reached at least the given
// multiple of the starting stake.
func mourirRicheAtLeast(vs []float64, m float64) int {
	n := 0
	for _, v := range vs {
		if v >= m {
			n++
		}
	}
	return n
}

// figMourirRiche ranks every vintage's estate on a log scale, twice: once for
// the household that spends everything the plan allows, once for the one that
// quietly takes a quarter less. The bold line at one times the stake is the
// frontier between leaving something and leaving nothing, and the whole point
// of the plate is that both clouds sit mostly above it.
func figMourirRiche() string {
	const (
		x0, x1     = 76.0, 596.0
		yTop, yBot = 118.0, 392.0
		lo, hi     = 0.03, 6.0 // the log axis, in multiples of the stake
	)
	n := len(mourirRicheSpent)
	x := func(i int) float64 { return x0 + float64(i)/float64(n-1)*(x1-x0) }
	y := func(v float64) float64 {
		f := math.Log10(v/lo) / math.Log10(hi/lo)
		return yBot - f*(yBot-yTop)
	}

	var b strings.Builder
	b.WriteString(plateHead("la distribution du patrimoine terminal",
		"Ce qu'il restait au bout de trente ans, millésime par millésime"))
	b.WriteString(plateDeck(
		"1 M€ de capital, retrait fixe indexé sur l'inflation et jamais ajusté : ce qui restait à l'arrivée, en multiple de la mise"))
	b.WriteString(sTxt(24, 78, 10.5, figMuted, "start", "400",
		fmt.Sprintf("%d départs, de %d à %d, sur le 60/40 américain réel du livre, classés du pire au meilleur", n, mourirRicheFirst, mourirRicheFirst+n-1)))
	legendChips(&b, 92, [][2]string{
		{figAccent, "40 000 € par an (4 %) : tout ce que le plan autorise"},
		{figBlue, "30 000 € par an (3 %) : le sous-dépensier"},
	})

	// Everything under the starting stake, tinted once.
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		x0, y(1), x1-x0, yBot-y(1), figWash)
	for _, g := range []struct {
		v     float64
		label string
	}{{0.05, "0,05x"}, {0.1, "0,1x"}, {0.5, "0,5x"}, {2, "2x"}, {3, "3x"}, {5, "5x"}} {
		gy := y(g.v)
		b.WriteString(line(x0, gy, x1, gy, figGrid, 1))
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", g.label))
	}

	b.WriteString(line(x0, yBot, x1, yBot, figRule, 1))

	// The frontier: leaving the stake intact, or not.
	b.WriteString(line(x0, y(1), x1, y(1), figDeep, 1.8))
	b.WriteString(mTxt(x0-8, y(1)+3.5, 10.5, figDeep, "end", "600", "1x"))
	b.WriteString(sTxt(x1-4, y(1)+17, 10, figDeep, "end", "400", "sous la mise de départ"))

	// The two ranked clouds. Each series is ranked on its own, so a rank is a
	// place in a distribution, not a vintage shared by both.
	spent := append([]float64(nil), mourirRicheSpent...)
	thrift := append([]float64(nil), mourirRicheThrift...)
	sort.Float64s(spent)
	sort.Float64s(thrift)
	for i, v := range thrift {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.2" fill="%s"/>`, x(i), y(v), figBlue)
	}
	for i, v := range spent {
		px := x(i)
		if v <= 0 { // the ruined vintage has no place on a log axis
			b.WriteString(line(px-4, yBot-4, px+4, yBot+4, figBad, 1.8))
			b.WriteString(line(px-4, yBot+4, px+4, yBot-4, figBad, 1.8))
			continue
		}
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.2" fill="%s"/>`, px, y(v), figAccent)
	}
	b.WriteString(sTxt(x0+16, yBot+18, 10, figBad, "start", "400",
		"un seul millésime a fini à zéro, celui de 1966"))

	// The median rank, and the two medians it reads off, annotated down in the
	// empty quarter of the panel rather than next to the crowded marks.
	med := n / 2
	mSpent, mThrift := spent[med], thrift[med]
	b.WriteString(dashLine(x(med), y(mSpent)+10, x(med), yBot-54, figMuted, 1, "2 3"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5.4" fill="none" stroke="%s" stroke-width="1.6"/>`,
		x(med), y(mThrift), figBlue)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5.4" fill="none" stroke="%s" stroke-width="1.6"/>`,
		x(med), y(mSpent), figAccent)
	b.WriteString(sTxt(x(med)+8, yBot-46, 10, figMuted, "start", "400",
		fmt.Sprintf("la médiane, le %dᵉ millésime sur %d", med+1, n)))
	b.WriteString(sTxt(x(med)+8, yBot-28, 10.5, figBlue, "start", "600",
		"sous-dépensier : "+frNum(mThrift, 1)+" fois la mise"))
	b.WriteString(sTxt(x(med)+8, yBot-10, 10.5, figAccent, "start", "600",
		"plan dépensé en entier : "+frNum(mSpent, 1)+" fois la mise"))

	// The counts, which are the sentence the article needs.
	b.WriteString(line(24, 434, 616, 434, figGrid, 1))
	b.WriteString(sTxt(400, 450, 9.5, figMuted, "middle", "400", "finissent au-dessus de la mise"))
	b.WriteString(sTxt(552, 450, 9.5, figMuted, "middle", "400", "à 3 fois la mise ou plus"))
	for _, r := range []struct {
		y         float64
		col, name string
		vs        []float64
	}{
		{472, figAccent, "40 000 € par an", mourirRicheSpent},
		{492, figBlue, "30 000 € par an", mourirRicheThrift},
	} {
		above := mourirRicheAboveStake(r.vs)
		top := mourirRicheAtLeast(r.vs, 3)
		b.WriteString(line(24, r.y-4, 40, r.y-4, r.col, 2.6))
		b.WriteString(sTxt(48, r.y, 10.5, r.col, "start", "600", r.name))
		b.WriteString(mTxt(400, r.y, 10.5, figSoft, "middle", "600", fmt.Sprintf("%d sur %d", above, n)))
		b.WriteString(mTxt(552, r.y, 10.5, figSoft, "middle", "600", fmt.Sprintf("%d sur %d", top, n)))
	}

	// The closing line's arithmetic: exactly one vintage ran out at 40 000 EUR
	// and none did at 30 000, so the quarter given up bought an escape from ruin
	// once in forty-three worlds, and an estate in all the others.
	b.WriteString(plateConclusion(522,
		fmt.Sprintf("Se rationner d'un quart n'a évité la ruine que dans un millésime sur %d. Dans les %d autres, cela n'a fait que grossir le legs.", n, n-1)))
	b.WriteString(sTxt(24, 540, 10, figMuted, "start", "400",
		"60/40 américain réel (S&amp;P 500, Treasuries 5 ans, déflatés CPI-U), reconstruction du livre ; sans fiscalité."))
	b.WriteString(sTxt(24, 554, 10, figMuted, "start", "400",
		fmt.Sprintf("%d fenêtres de trente ans, largement chevauchantes ; chaque série est classée pour elle-même, rang par rang.", n)))
	return svg(640, 568, b.String())
}
