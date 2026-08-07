package firebook

import (
	"fmt"
	"strings"
)

// The long-volatility article's plate: permanent put protection read against
// the equity/bond ladder, on the ladder's own axes. See figPutsDomines.

// putsRung is one portfolio measured over the PPUT window: annualized total
// return and deepest month-end drawdown, both nominal, both in percent.
type putsRung struct {
	name   string
	cagr   float64
	drawdn float64
}

// The ladder is built exactly like the all-weather plate's (figTousTempsEchange):
// same two legs, S&P 500 total return (pkg/datasets/simdata/SP500) against
// intermediate Treasuries (simdata/IEF), month-end values, weights reset every
// December. Two things change, and both are imposed by the object it is
// compared with. The window is the published one of the protection index,
// 30 June 1986 to 31 December 2018 (390 months, the Cboe/Wilshire study), and
// the numbers are nominal total returns rather than real ones, because that is
// how the index is quoted. Measuring the ladder in real terms here, or over
// 1972-2024, would mean transferring a published figure across a window and
// across the nominal/real boundary, which is exactly what the plate must not do.
//
// The anchor that makes the two objects comparable: the 100/0 rung comes out at
// 9.80 % and −50.95 %, the very numbers Cboe/Wilshire publish for the S&P 500
// over the same window (9.8 % and −50.95 %). TestPutsLadderMatchesTheData
// recomputes every rung from the bundled series and fails on any disagreement.
var putsLadder = []putsRung{
	{"50/50", 8.79, -20.22},
	{"55/45", 8.94, -23.58},
	{"60/40", 9.08, -26.87},
	{"65/35", 9.21, -30.10},
	{"70/30", 9.33, -33.26},
	{"75/25", 9.44, -36.37},
	{"80/20", 9.53, -39.41},
	{"85/15", 9.62, -42.38},
	{"90/10", 9.69, -45.30},
	{"95/5", 9.75, -48.15},
	{"100/0", 9.80, -50.95},
}

// putsProtected is the Cboe S&P 500 5 % Put Protection Index (PPUT), the S&P 500
// carrying a permanent 5 % out-of-the-money put, rolled monthly. These two
// numbers are PUBLISHED, not measured here: no option series is bundled in this
// repository, and none could be reconstructed honestly. Source: Wilshire
// Associates for Cboe, "Options-Based Benchmark Indexes", 30 June 1986 to
// 31 December 2018, annualized return 6.6 % and maximum drawdown −38.92 %,
// against 9.8 % and −50.95 % for the S&P 500 over the same 390 months. The plate
// draws the point as a different kind of mark and says so in words.
var putsProtected = putsRung{"PPUT", 6.6, -38.92}

// putsRungAt returns the rung with the given name, or panics: the plate reads
// two rungs by name and must never silently point at the wrong one.
func putsRungAt(name string) putsRung {
	for _, r := range putsLadder {
		if r.name == name {
			return r
		}
	}
	panic("firebook: no ladder rung named " + name)
}

// figPutsDomines draws Israelov's conclusion instead of asserting it. The
// equity/bond ladder is a line of eleven measured points; permanent put
// protection is one published point, far to the left of the line. The shaded
// quadrant hanging off that point is everything that beats it on both axes at
// once, and six rungs of the ladder sit inside it.
func figPutsDomines() string {
	m := mapper(6.2, 10.0, -54, -18, 96, 556, 320, 96)
	var b strings.Builder
	b.WriteString(plateHead("long volatility", "Détenir moins d'actions domine la protection par puts"))
	legendChips(&b, 56, [][2]string{
		{figBlue, "échelle actions / obligations (recalculée)"},
		{figBad, "protection par puts (publiée)"},
	})

	// the dominance quadrant, behind everything: more return AND less drawdown
	// than the protected point
	pp := m(putsProtected.cagr, putsProtected.drawdn)
	fmt.Fprintf(&b, `<rect x="%.1f" y="96" width="%.1f" height="%.1f" fill="%s"/>`, pp[0], 556-pp[0], pp[1]-96, figWash)
	b.WriteString(dashLine(pp[0], 96, pp[0], pp[1]-9, figMuted, 1, "2 4"))

	// horizontal grid and the two axis titles
	for _, g := range []float64{-20, -30, -40, -50} {
		gy := m(6.2, g)[1]
		b.WriteString(line(96, gy, 556, gy, figGrid, 1))
		b.WriteString(mTxt(88, gy+3.5, 10, figMuted, "end", "400", ttNum(g, 0)+" %"))
	}
	b.WriteString(sTxt(96, 86, 10.5, figMuted, "start", "400", "pire drawdown (%)"))
	b.WriteString(line(96, 320, 556, 320, figRule, 1))
	for _, t := range []float64{7, 8, 9, 10} {
		b.WriteString(mTxt(m(t, 0)[0], 336, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", t)))
	}
	b.WriteString(sTxt(326, 356, 11, figMuted, "middle", "400", "rendement annualisé, dividendes réinvestis, en nominal (%)"))

	// the ladder itself, from 50/50 up to 100 % equities
	lad := make([][2]float64, len(putsLadder))
	for i, r := range putsLadder {
		lad[i] = m(r.cagr, r.drawdn)
	}
	b.WriteString(poly(lad, figBlue, 2.2, ""))
	for i, p := range lad {
		fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.2" fill="%s" stroke="#fffdf9" stroke-width="1.4"/>`, p[0], p[1], figBlue)
		b.WriteString(sTxt(p[0]+9, p[1]+3.5, 10, figBlue, "start", "400", putsLadder[i].name))
	}

	// what the quadrant means, in the empty band left of the ladder
	b.WriteString(sTxt(158, 150, 11, figSoft, "start", "600", "dans cette zone : plus de rendement"))
	b.WriteString(sTxt(158, 165, 11, figSoft, "start", "600", "et moins de drawdown"))
	b.WriteString(sTxt(158, 182, 10.5, figMuted, "start", "400", "on y trouve toute l'échelle, du 50/50 au 75/25"))

	// the read at equal drawdown, terminating on a real rung
	eq := putsRungAt("80/20")
	ep := m(eq.cagr, eq.drawdn)
	b.WriteString(dashLine(pp[0]+10, pp[1], ep[0]-7, ep[1], figDeep, 1.2, "4 4"))
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="5" fill="none" stroke="%s" stroke-width="1.6"/>`, ep[0], ep[1], figDeep)
	b.WriteString(sTxt(452, 205, 11, figDeep, "end", "600", "à pire drawdown égal, l'échelle rend "+ttNum(eq.cagr, 1)+" %"))
	b.WriteString(sTxt(452, 220, 10.5, figMuted, "end", "400", "soit "+ttNum(eq.cagr-putsProtected.cagr, 1)+" points de plus par an"))

	// the protected point, a different kind of object
	fmt.Fprintf(&b, `<path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f L %.1f,%.1f Z" fill="%s" stroke="#fffdf9" stroke-width="1.6"/>`,
		pp[0], pp[1]-7.5, pp[0]+7.5, pp[1], pp[0], pp[1]+7.5, pp[0]-7.5, pp[1], figBad)
	b.WriteString(sTxt(96, 250, 11, figBad, "start", "600", "S&amp;P 500 + put 5 % permanent"))
	b.WriteString(mTxt(96, 265, 10.5, figBad, "start", "600", ttNum(putsProtected.cagr, 1)+" % · "+ttNum(putsProtected.drawdn, 1)+" %"))
	b.WriteString(sTxt(96, 281, 10, figMuted, "start", "400", "indice Cboe PPUT : chiffres publiés,"))
	b.WriteString(sTxt(96, 294, 10, figMuted, "start", "400", "pas une mesure de ce livre"))

	b.WriteString(sTxt(24, 382, 9.5, figMuted, "start", "400",
		"Même échelle actions / obligations que la planche des portefeuilles tous-temps : S&amp;P 500 et Treasuries intermédiaires, fin de mois,"))
	b.WriteString(sTxt(24, 396, 9.5, figMuted, "start", "400",
		"rééquilibrage annuel, mesurée ici sur la fenêtre publiée du PPUT (juin 1986 - décembre 2018, 390 mois) et en nominal."))
	b.WriteString(sTxt(24, 410, 9.5, figMuted, "start", "400",
		"Son barreau 100 % actions retrouve le S&amp;P 500 publié (9,8 % et −50,95 %) : les deux objets sont bien sur le même pied."))
	return svg(640, 424, b.String())
}
