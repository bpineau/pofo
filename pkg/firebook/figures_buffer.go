package firebook

import (
	"fmt"
	"strings"
)

// The cash-buffer plate: how long the real crossings actually last, against how
// far a buffer of 18 or 36 months reaches. It is the missing proof of the
// article's "arithmétique des durées", and the reason its ruin curve comes out
// flat: the buffer covers the first half of the underwater months, and the
// episodes it fails to cover are exactly the long ones.
//
// EVERYTHING IS REAL: constant purchasing power, inflation removed.

// bufferDepthThreshold is the drawdown, in percent, an episode must exceed to
// reach the plate. It is the article's own consumption trigger ("les retraits
// basculent sur le matelas quand le portefeuille est en drawdown réel de plus
// de 10-20 %"), so a shallower dip never calls on the buffer and has no place
// in a plate about the buffer's reach. The full record holds 103 episodes, 21
// of them deeper than 5 %, most of those three-month dips that would clutter
// the plate without changing its reading.
const bufferDepthThreshold = -10.0

// bufferEpisodes freezes the underwater episodes of the bundled real US 60/40
// (pkg/replay.Reference: S&P 500 + 5-year Treasuries rebalanced every January,
// deflated by the US CPI-U, month-end from April 1953 to May 2026). An episode
// runs from a real all-time high to the month the index first regains that
// high; months is that span and depth the worst real drawdown inside it. That
// span is the honest upper bound of what a buffer would have to carry: a
// written trigger only diverts the withdrawals while the drawdown holds past
// its threshold, so a few months at each end of a bar would still be paid from
// the portfolio.
//
// To reproduce: take replay.Reference().Index, walk it forward keeping the
// running maximum, open an episode on the first month below the peak and close
// it on the first month at or above it, then keep the episodes deeper than
// bufferDepthThreshold and sort them longest first. No episode is open at the
// end of the record (the index sets a new real high in its last month), so
// every bar here is a completed crossing. figures_buffer_test.go recomputes the
// list from pkg/replay and fails on any drift, including a future refresh that
// would leave the last episode unresolved.
var bufferEpisodes = []struct {
	from, to string  // month the peak was set, month it was regained (YYYY-MM)
	months   int     // from the peak to the recovery
	depth    float64 // worst real drawdown of the episode, in percent
	name     string  // the crisis, when it has a name worth printing
}{
	{"1972-12", "1983-04", 124, -38.0, "stagflation des années 1970"},
	{"2000-08", "2006-10", 74, -24.0, "bulle internet"},
	{"1968-11", "1972-05", 42, -24.8, ""},
	{"2007-10", "2011-02", 40, -28.6, "crise financière"},
	{"2021-12", "2024-08", 32, -22.6, "choc d'inflation"},
	{"1956-07", "1958-07", 24, -11.3, ""},
	{"1987-08", "1989-07", 23, -19.7, ""},
	{"1965-10", "1967-04", 18, -12.3, ""},
	{"1961-12", "1963-04", 16, -13.0, ""},
}

// figTraverseesMatelas draws one bar per crossing, sorted longest first, cut at
// the two sizes the article recommends: what 18 months of buffer covers, what
// the next 18 months add, and what stays uncovered beyond 36.
func figTraverseesMatelas() string {
	const (
		x0, x1  = 160.0, 568.0 // the bar area
		mMax    = 132.0        // months, the axis span
		yTop    = 112.0        // first row
		pitch   = 27.0
		bh      = 14.0
		lockLo  = 18.0
		lockHi  = 36.0
		labelX  = 112.0 // right edge of the episode column
		depthX  = 152.0 // right edge of the drawdown column
		summary = 316.0 // left edge of the closing note
	)
	x := func(m float64) float64 { return x0 + m/mMax*(x1-x0) }
	yBot := yTop + pitch*float64(len(bufferEpisodes)-1) + bh

	var b strings.Builder
	b.WriteString(plateHead("le matelas de liquidités",
		"Neuf traversées contre la portée du matelas"))
	b.WriteString(sTxt(24, 62, 10.5, figMuted, "start", "400",
		"mois passés sous le dernier sommet réel, 60/40 américain réel, 1953-2026"))
	legendChips(&b, 72, [][2]string{
		{figAccent, "les 18 premiers mois"},
		{figBlue, "de 18 à 36 mois"},
		{figBad, "au-delà : à découvert"},
	})

	// Vertical grid every two years, plus the two locks the article's sizing
	// rule names, and the bracket that reads them as one recommended band.
	for _, g := range []float64{0, 24, 48, 72, 96, 120} {
		gx := x(g)
		col := figGrid
		if g == 0 {
			col = figRule
		}
		b.WriteString(line(gx, yTop-8, gx, yBot+8, col, 1))
		b.WriteString(mTxt(gx, yBot+26, 10, figMuted, "middle", "400", fmt.Sprintf("%.0f", g)))
	}
	for _, l := range []float64{lockLo, lockHi} {
		b.WriteString(dashLine(x(l), 104, x(l), yBot+8, figMuted, 1, "2 4"))
	}
	b.WriteString(line(x(lockLo), 100, x(lockHi), 100, figDeep, 1.4))
	b.WriteString(line(x(lockLo), 97, x(lockLo), 103, figDeep, 1.4))
	b.WriteString(line(x(lockHi), 97, x(lockHi), 103, figDeep, 1.4))
	b.WriteString(sTxt(x(lockHi)+10, 104, 10.5, figDeep, "start", "600",
		"la taille recommandée : 18 à 36 mois"))
	b.WriteString(sTxt(labelX, 100, 9.5, figMuted, "end", "600", "épisode"))
	b.WriteString(sTxt(depthX, 100, 9.5, figMuted, "end", "600", "creux"))

	for i, e := range bufferEpisodes {
		y := yTop + pitch*float64(i)
		b.WriteString(sTxt(labelX, y+bh-3, 11, figSoft, "end", "600", e.from[:4]+"-"+e.to[:4]))
		b.WriteString(mTxt(depthX, y+bh-3.5, 10, figMuted, "end", "400",
			fmt.Sprintf("−%.0f %%", -e.depth)))

		// The bar, cut at the two locks: each segment but the last is a plain
		// rect with a hairline gap, the last one carries the rounded data end.
		cuts := []float64{0, lockLo, lockHi, mMax}
		fills := []string{figAccent, figBlue, figBad}
		m := float64(e.months)
		for s := 0; s < 3; s++ {
			lo, hi := cuts[s], cuts[s+1]
			if m <= lo {
				break
			}
			if hi >= m {
				b.WriteString(barH(x(lo), x(m), y, bh, fills[s]))
				break
			}
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
				x(lo), y, x(hi)-x(lo)-1.5, bh, fills[s])
		}
		// The month count sits at the bar end, nudged clear of a lock line when
		// the bar happens to stop just short of one.
		lx := x(m) + 8
		for _, l := range []float64{lockLo, lockHi} {
			if lx < x(l)+3 && lx+15 > x(l)-3 {
				lx = x(l) + 6
			}
		}
		b.WriteString(mTxt(lx, y+bh-3, 10.5, figInk, "start", "600", fmt.Sprintf("%d", e.months)))

		if e.name == "" {
			continue
		}
		// The named crossings: inside the uncovered tail when it is wide enough
		// to hold the words, otherwise just after the month count.
		if x(m)-x(lockHi) > 120 {
			b.WriteString(sTxt(x(lockHi)+9, y+bh-3.5, 10, "#fffdf9", "start", "600", e.name))
		} else {
			b.WriteString(sTxt(x(m)+34, y+bh-3, 10, figMuted, "start", "400", e.name))
		}
	}

	// The reading, in numbers the plate itself carries.
	inside, uncovered, worst := 0, 0, 0
	for _, e := range bufferEpisodes {
		if float64(e.months) <= lockHi {
			inside++
			continue
		}
		over := e.months - int(lockHi)
		uncovered += over
		if over > worst {
			worst = over
		}
	}
	b.WriteString(sTxt(summary, 250, 10.5, figSoft, "start", "600",
		fmt.Sprintf("%d épisodes sur %d tiennent dans 36 mois.", inside, len(bufferEpisodes))))
	b.WriteString(sTxt(summary, 266, 10.5, figMuted, "start", "400",
		fmt.Sprintf("Les %d autres laissent %d mois à découvert,", len(bufferEpisodes)-inside, uncovered)))
	b.WriteString(sTxt(summary, 282, 10.5, figMuted, "start", "400",
		fmt.Sprintf("dont %d pour la seule traversée des", worst)))
	b.WriteString(sTxt(summary, 298, 10.5, figMuted, "start", "400", "années 1970."))

	b.WriteString(line(x0, yBot+8, x1, yBot+8, figRule, 1))
	b.WriteString(sTxt((x0+x1)/2, yBot+46, 11, figMuted, "middle", "400",
		"durée de l'épisode, du sommet réel au retour à ce sommet (mois)"))
	return svg(640, int(yBot+62), b.String())
}
