package firebook

import (
	"fmt"
	"strings"
)

// The plate of "Se protéger de l'inflation": equities during and after an
// inflation episode. The article's pedagogical core ("le point qui embrouille
// tout le monde") is that stocks lose in real terms DURING the episode and only
// catch up years later. One long real curve, the starting purchasing power as a
// rule, and the two dates that matter say it in a second.
//
// EVERYTHING IS REAL, in dollars: the S&P 500 total-return index deflated by
// the US CPI, the deflator pkg/replay uses for the reference 60/40.

// rattrapIndex is the real (CPI-deflated) S&P 500, month-end, base 100 at the
// end of December 1972, one value per month through December 1987: the fifteen
// years an investor who bought just before the 1973 episode lived through.
//
// How to reproduce: read pkg/datasets/refdata/SP500-USD.csv (month-end total
// return, so no dividend double-count and no anchor shift), keep the closes
// from December 1972 to December 1987, run scenario.Deflate against the ^CPI-US
// levels marketdata serves offline, and compound the resulting monthly real
// returns from 100. That is exactly what pkg/replay does for its reference
// series, minus the bond sleeve. figures_protection_test.go recomputes all 181
// values and both annotated dates, and fails on drift.
//
// The two annotated months are read off the series, never assumed:
//   - rattrapTrough: the minimum of the window, September 1974 at 47.74, i.e.
//     52 % of the starting purchasing power gone.
//   - rattrapBack: the first month back above 100 that is never followed by a
//     month below it. It is January 1985, 145 months (twelve years and one
//     month) after the base, and the record stays above 100 to this day, so the
//     first crossing and the durable one are the same month here.
var (
	rattrapTrough = 21  // September 1974, the minimum
	rattrapBack   = 145 // January 1985, above 100 for good
	rattrapIndex  = []float64{
		100.00, 97.82, 93.52, 92.97, 88.77, 86.72, 86.17, 88.11,
		84.87, 87.79, 87.27, 77.05, 77.88, 76.36, 75.35, 73.51,
		69.99, 67.30, 66.04, 60.37, 54.51, 47.74, 55.31, 52.20,
		51.16, 57.23, 60.65, 61.95, 64.75, 67.34, 69.77, 65.13,
		63.67, 61.33, 64.87, 66.46, 65.78, 73.54, 72.79, 74.97,
		73.86, 72.67, 75.43, 74.67, 74.25, 75.77, 74.19, 73.59,
		77.31, 72.91, 71.09, 69.76, 69.68, 67.83, 70.81, 69.70,
		68.26, 68.13, 65.13, 66.94, 66.99, 62.73, 60.98, 62.29,
		67.30, 67.10, 65.70, 69.20, 70.75, 69.91, 63.48, 64.54,
		65.26, 67.32, 64.50, 67.59, 67.12, 64.91, 67.00, 67.21,
		70.31, 70.07, 64.92, 67.28, 67.75, 70.94, 69.87, 62.35,
		64.57, 67.16, 69.22, 73.51, 73.67, 75.07, 75.95, 83.31,
		80.08, 75.97, 76.75, 79.26, 77.11, 76.65, 75.29, 74.85,
		69.83, 66.23, 69.58, 72.20, 70.12, 69.00, 65.18, 64.59,
		66.87, 63.86, 62.48, 61.24, 68.52, 69.16, 77.26, 80.71,
		82.12, 85.08, 87.03, 89.63, 96.12, 94.95, 97.97, 95.07,
		95.99, 97.01, 95.69, 97.60, 96.55, 95.52, 91.97, 93.13,
		93.70, 88.23, 89.80, 88.33, 97.62, 97.37, 97.70, 96.59,
		98.94, 106.15, 107.04, 106.64, 106.10, 111.89, 113.43, 113.04,
		111.77, 107.87, 112.53, 119.92, 125.36, 126.38, 136.42, 144.31,
		142.30, 149.07, 151.53, 142.79, 152.73, 139.91, 147.85, 151.32,
		146.54, 165.68, 171.46, 175.44, 173.23, 174.12, 182.36, 190.60,
		196.66, 191.82, 150.37, 137.97, 148.08,
	}
)

// The two ratios the footnote quotes over the same window, December 1972 to
// December 1987: US consumer prices multiplied by 2.716 and the nominal index
// by 4.02. The plate does NOT draw the cumulated inflation as a second curve.
// The main series is already net of it, so a second line on the same axis would
// read as a rival portfolio that beat stocks, which is the exact misreading the
// article is trying to kill. It is stated in words instead.
var (
	rattrapCPIx     = 2.72
	rattrapNominalx = 4.02
)

// rattrapWash is figBad at .10 and rattrapGreen figGood at .10, both PRE-BLENDED
// onto the figure card background #fffdf9 (bg*(1-a) + colour*a per channel):
// crengine paints rgba fills solid black, so tints are always solid hex here.
const (
	rattrapWash  = "#F9EEE9"
	rattrapGreen = "#ECF2EB"
)

// rattrapRuns cuts a pixel curve into the runs that lie on one side of a
// horizontal rule, splitting each crossing at its exact abscissa, and closes
// every run onto the rule so it can be filled. Below the rule in pixel space
// (y greater) means below the rule in value space.
func rattrapRuns(px [][2]float64, yRule float64) (below, above [][][2]float64) {
	side := func(p [2]float64) int {
		switch {
		case p[1] > yRule:
			return -1
		case p[1] < yRule:
			return 1
		}
		return 0
	}
	cut := func(a, b [2]float64) [2]float64 {
		t := (yRule - a[1]) / (b[1] - a[1])
		return [2]float64{a[0] + t*(b[0]-a[0]), yRule}
	}
	flush := func(run [][2]float64, s int) {
		if len(run) < 2 || run[0][0] == run[len(run)-1][0] {
			return // nothing, or a vertical sliver with no area
		}
		closed := append([][2]float64{{run[0][0], yRule}}, run...)
		closed = append(closed, [2]float64{run[len(run)-1][0], yRule})
		if s < 0 {
			below = append(below, closed)
		} else {
			above = append(above, closed)
		}
	}
	run, cur := [][2]float64{px[0]}, side(px[0])
	for i := 1; i < len(px); i++ {
		s := side(px[i])
		if s != 0 && cur != 0 && s != cur {
			x := cut(px[i-1], px[i])
			flush(append(run, x), cur)
			run, cur = [][2]float64{x}, s
			run = append(run, px[i])
			continue
		}
		if cur == 0 {
			cur = s
		}
		run = append(run, px[i])
	}
	flush(run, cur)
	return below, above
}

// rattrapArea emits one filled polygon.
func rattrapArea(pts [][2]float64, fill string) string {
	var b strings.Builder
	b.WriteString(`<path d="M `)
	for i, p := range pts {
		if i > 0 {
			b.WriteString(" L ")
		}
		fmt.Fprintf(&b, "%.1f,%.1f", p[0], p[1])
	}
	fmt.Fprintf(&b, ` Z" fill="%s"/>`, fill)
	return b.String()
}

func figActionsRattrapent() string {
	const (
		x0, x1   = 62.0, 600.0
		yTop     = 92.0  // 205
		yBot     = 306.0 // 40
		vLo, vHi = 40.0, 205.0
	)
	n := len(rattrapIndex) - 1 // 180 monthly steps, December 1972 to December 1987
	px := func(i float64) float64 { return x0 + i/float64(n)*(x1-x0) }
	py := func(v float64) float64 { return yBot - (v-vLo)/(vHi-vLo)*(yBot-yTop) }
	curve := make([][2]float64, len(rattrapIndex))
	for i, v := range rattrapIndex {
		curve[i] = [2]float64{px(float64(i)), py(v)}
	}
	yRule := py(100)

	var b strings.Builder
	b.WriteString(plateHead("les actions pendant l'inflation",
		"La moitié du pouvoir d'achat perdue en 1974, tout repris en 1985"))
	b.WriteString(plateDeck(
		"S&amp;P 500 dividendes réinvestis, en dollars, déflaté par le CPI américain. Base 100 fin décembre 1972."))

	// The two seas: the deficit to climb out of, then what is above water.
	below, above := rattrapRuns(curve, yRule)
	for _, a := range below {
		b.WriteString(rattrapArea(a, rattrapWash))
	}
	for _, a := range above {
		b.WriteString(rattrapArea(a, rattrapGreen))
	}

	// Horizontal grid, then the year hairlines.
	for _, g := range []float64{50, 100, 150, 200} {
		gy := py(g)
		if g != 100 {
			b.WriteString(line(x0, gy, x1, gy, figGrid, 1))
		}
		b.WriteString(mTxt(x0-9, gy+3.5, 10, figMuted, "end", "400", fmt.Sprintf("%.0f", g)))
	}
	for y := 1973; y <= 1987; y += 2 {
		gx := px(float64(12*(y-1973) + 1)) // January of that year
		b.WriteString(line(gx, yTop, gx, yBot, figGrid, 1))
		b.WriteString(mTxt(gx, 322, 10, figMuted, "middle", "400", fmt.Sprintf("%d", y)))
	}
	b.WriteString(line(x0, yBot, x1, yBot, figRule, 1))

	// The frontier the whole article is about.
	b.WriteString(line(x0, yRule, x1, yRule, figSoft, 2))
	b.WriteString(sTxt(x0+120, yRule-8, 10.5, figSoft, "start", "600",
		"pouvoir d'achat de départ"))

	b.WriteString(poly(curve, figAccent, 2.2, ""))

	// The trough, read off the series.
	tx, ty := curve[rattrapTrough][0], curve[rattrapTrough][1]
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, tx, ty, figBad)
	b.WriteString(dashLine(tx, ty-8, tx+14, 202, figBad, 1, "2 3"))
	b.WriteString(sTxt(tx+18, 175, 11, figBad, "start", "600", "creux de septembre 1974"))
	b.WriteString(mTxt(tx+18, 190, 10.5, figBad, "start", "400", "−52 % de pouvoir d'achat"))

	// The durable recovery, also read off the series.
	rx, ry := curve[rattrapBack][0], curve[rattrapBack][1]
	fmt.Fprintf(&b, `<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, rx, ry, figGood)
	b.WriteString(dashLine(rx, ry-8, rx-10, 148, figGood, 1, "2 3"))
	b.WriteString(sTxt(rx-14, 122, 11, figGood, "end", "600", "janvier 1985 : au-dessus, pour de bon"))
	b.WriteString(mTxt(rx-14, 137, 10.5, figGood, "end", "400", "12 ans et 1 mois plus tard"))

	// Direct end label instead of a legend.
	last := curve[len(curve)-1]
	b.WriteString(mTxt(x1+6, last[1]+4, 10.5, figDeep, "start", "600", "148"))

	// The time spent under water, measured under the axis.
	by := 340.0
	b.WriteString(line(x0, by, rx, by, figBad, 1.4))
	b.WriteString(line(x0, by-4, x0, by+4, figBad, 1.4))
	b.WriteString(line(rx, by-4, rx, by+4, figBad, 1.4))
	b.WriteString(sTxt((x0+rx)/2, by-8, 10.5, figBad, "middle", "600",
		"douze ans et un mois sous le pouvoir d'achat de départ"))

	b.WriteString(sTxt(24, 368, 10, figMuted, "start", "400",
		fmt.Sprintf("Sur ces quinze ans, les prix américains sont multipliés par %s et l'indice nominal par %s : la courbe est déjà nette de cette hausse.",
			frNum(rattrapCPIx, 1), frNum(rattrapNominalx, 1))))
	b.WriteString(sTxt(24, 382, 10, figMuted, "start", "400",
		"La chute finale est le krach d'octobre 1987, qui laisse encore l'indice 48 % au-dessus du départ."))
	b.WriteString(sTxt(24, 396, 10, figMuted, "start", "400",
		"« Durablement » : premier mois repassé au-dessus de 100 sans jamais y redescendre ensuite, vérifié jusqu'à aujourd'hui."))
	return svg(640, 408, b.String())
}
