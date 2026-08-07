package firebook

import (
	"fmt"
	"math"
	"strings"
)

// The Weimar plate of hyperinflation-et-extremes: what an equity holder and a
// government-bond holder were left with, in real terms, through the German
// hyperinflation.

// weimarPath is one leg of the plate: the real capital left for 100 invested at
// the end of 1913, year by year.
type weimarPath struct {
	Label string
	Color string
	Vals  []float64 // one value per year, weimarFrom to weimarTo inclusive
}

const (
	weimarFrom = 1913 // the base year, index 100
	weimarTo   = 1926
	// The measurement of German asset prices collapses with the currency: the
	// panel's 1922-1924 marks read +93 %, +150 % then −87 % real for equities,
	// the classic artefact of pricing anything in a currency losing orders of
	// magnitude a month. The plate greys that window and refuses to draw a
	// confident trace through it.
	weimarMurkFrom = 1922
	weimarMurkTo   = 1924
	// weimarFloor is the bottom of the log axis. A leg that reaches exactly
	// zero cannot be plotted on a log scale, so the bond is drawn down to the
	// floor with a dashed segment and an explicit "atteint zéro" marker rather
	// than a silent floor that would read as "almost nothing left".
	weimarFloor = 0.7
	weimarCeil  = 260.0
)

// weimarPaths freezes the two cumulative real paths the plate draws.
//
// Source: datasets.BroadSample(), the bundled Jorda-Schularick-Taylor
// Macrohistory R6 panel, rows iso = DEU, columns equity and bond. Those are
// annual REAL total returns, each year already deflated by the German CPI, so
// each path is the running product of (1 + r) from 1914 on, based at 100 at the
// end of 1913. The bond's 1923 return is coded −1 exactly, that is a total
// wipe-out, so the leg is zero from 1923 on and stays zero: nothing compounds
// out of nothing.
//
// figures_weimar_test.go recomputes both paths from pkg/datasets and fails the
// moment the plate and the panel disagree.
var weimarPaths = []weimarPath{
	{"Actions allemandes", figAccent, []float64{
		100, 88.97, 77.85, 68.26, 56.15, 33.50, 40.10, 27.53, 47.01, 90.76, 226.58, 28.46, 16.97, 40.06,
	}},
	{"Obligations d'État", figBad, []float64{
		100, 101.74, 86.37, 68.49, 47.10, 39.76, 31.95, 14.58, 15.93, 1.48, 0, 0, 0, 0,
	}},
}

// weimarYear returns the index of a year in a path's values.
func weimarYear(y int) int { return y - weimarFrom }

// figWeimarReel plots both legs on a LOG axis, the only scale on which a leg
// divided by seventy and a leg divided by infinity can share a plate, and the
// only one that makes "amoché" and "anéanti" visibly different states rather
// than two lines crushed against the same floor.
func figWeimarReel() string {
	const (
		plotL, plotR = 92.0, 498.0
		plotT, plotB = 92.0, 330.0
	)
	x := func(year float64) float64 {
		return plotL + (year-weimarFrom)/float64(weimarTo-weimarFrom)*(plotR-plotL)
	}
	y := func(v float64) float64 {
		if v < weimarFloor {
			v = weimarFloor
		}
		lo, hi := math.Log10(weimarFloor), math.Log10(weimarCeil)
		return plotB - (math.Log10(v)-lo)/(hi-lo)*(plotB-plotT)
	}

	var b strings.Builder
	b.WriteString(plateHead("hyperinflation",
		"Weimar en termes réels : la créance à zéro, l'action seulement amochée"))
	b.WriteString(sTxt(24, 64, 10.5, figMuted, "start", "400",
		"capital réel restant pour 100 placés fin 1913, échelle log"))

	// The unreliable window, drawn first so everything else sits on top.
	mx0, mx1 := x(float64(weimarMurkFrom)-0.5), x(float64(weimarMurkTo)+0.5)
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
		mx0, plotT, mx1-mx0, plotB-plotT, figWash)
	b.WriteString(sTxt((mx0+mx1)/2, plotT-22, 10, figMuted, "middle", "600", "1922-1924"))
	b.WriteString(sTxt((mx0+mx1)/2, plotT-9, 9.5, figMuted, "middle", "400", "mesure non fiable"))

	// Decade gridlines, with the 100 line solid: it is the frontier between a
	// fortune dented and a fortune destroyed.
	for _, g := range []float64{1, 10, 100} {
		gy := y(g)
		col := figGrid
		if g == 100 {
			col = figRule
		}
		b.WriteString(line(plotL, gy, plotR, gy, col, 1))
		b.WriteString(mTxt(plotL-8, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	b.WriteString(sTxt(plotL-8, y(100)-9, 9.5, figMuted, "end", "400", "capital intact"))

	// Year ticks, every second year so the labels never touch.
	for yr := weimarFrom; yr <= weimarTo; yr++ {
		if (yr-weimarFrom)%2 != 0 {
			continue
		}
		b.WriteString(mTxt(x(float64(yr)), plotB+18, 10, figMuted, "middle", "400", fmt.Sprintf("%d", yr)))
	}

	for _, p := range weimarPaths {
		// Solid before the murk, solid after, dashed across it.
		var pre, murk, post [][2]float64
		for i, v := range p.Vals {
			yr := weimarFrom + i
			pt := [2]float64{x(float64(yr)), y(v)}
			switch {
			case yr < weimarMurkFrom:
				pre = append(pre, pt)
			case yr > weimarMurkTo:
				post = append(post, pt)
			}
			if yr >= weimarMurkFrom-1 && yr <= weimarMurkTo+1 {
				murk = append(murk, pt)
			}
		}
		b.WriteString(poly(pre, p.Color, 2.4, ""))
		b.WriteString(poly(murk, p.Color, 1.6, "3 3"))
		b.WriteString(poly(post, p.Color, 2.4, ""))
	}

	// The bond dies inside the murk: mark the year it reaches zero.
	zx, zy := x(1923), y(0)
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, zx, zy, figBad)
	b.WriteString(sTxt(zx+10, zy-16, 11, figBad, "start", "600", "l'obligation d'État atteint zéro"))
	b.WriteString(sTxt(zx+10, zy-3, 10, figMuted, "start", "400", "en 1923, et n'en revient pas"))

	// The two arrival points carry the message; the murk carries none.
	eq := weimarPaths[0].Vals
	lowIdx := weimarYear(1925)
	lx, ly := x(1925), y(eq[lowIdx])
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, lx, ly, figAccent)
	b.WriteString(mTxt(lx-8, ly+4, 10.5, figDeep, "end", "600", "17"))
	b.WriteString(sTxt(lx-30, ly+4, 10.5, figSoft, "end", "400", "1925"))

	endIdx := weimarYear(weimarTo)
	ex, ey := x(float64(weimarTo)), y(eq[endIdx])
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="4.2" fill="%s"/>`, ex, ey, figAccent)
	b.WriteString(sTxt(ex+10, ey-5, 11, figDeep, "start", "600", "Actions allemandes"))
	b.WriteString(mTxt(ex+10, ey+11, 10.5, figSoft, "start", "400", "40 en 1926"))

	b.WriteString(sTxt(24, 372, 11, figSoft, "start", "600",
		"Le rentier obligataire est ruiné, l'actionnaire a perdu plus de 80 % et lui survit."))
	b.WriteString(sTxt(24, 390, 9.5, figMuted, "start", "400",
		"Rendements réels annuels allemands du panel Jorda-Schularick-Taylor (R6), capitalisés depuis fin 1913."))
	return svg(640, 408, b.String())
}
