package firebook

import (
	"fmt"
	"strings"
)

// The plate of the factor article. Its question is not "does the tilt pay more"
// but "does it suffer in the same decades as the broad market", so the form is a
// single diverging area around zero: the ten-year real CAGR gap between US
// small-cap value and the S&P 500, window by window. Above the rule, the
// decades the tilt won; below it, the years of shame that are its entry price.
//
// EVERYTHING IS REAL: inflation removed, before tax and before fund costs.

// Pre-blended solid fills, never rgba (crengine paints rgba solid black): each
// is its colour composited once onto the figure card background #fffdf9,
// figScvAreaUp = figAccent at .38, figScvAreaDown = figBad at .45 (the losing
// side gets the heavier tint: it is the one the article asks the reader to look at).
const (
	figScvAreaUp   = "#E3CBB1"
	figScvAreaDown = "#E3B9B2"
)

// --- The rolling ten-year gap: small-cap value minus the broad market ---

// scvGapStart is the month key (year*12 + month - 1) of the first plotted
// window END, July 1973: the record opens in July 1963 and the first complete
// 120-month window closes ten years later.
const scvGapStart = 1973*12 + 6

// scvGapPoints is that gap in percentage points a year, one value per month-end
// window from July 1973 to May 2026 (635 windows), each dated by the month that
// CLOSES its window.
//
// Reproduce: read the bundled USSCV-USD (Kenneth French's value-weighted small
// x high book-to-market portfolio, daily since 1963-07) and SP500-USD (S&P 500
// total return, month-end) with marketdata.ReadSimdataFS(datasets.Refdata(), id)
// and keep the last quote of each calendar month; divide both by the ^CPI-US
// snapshot level of the same month (marketdata.NewClient("").Fetch, the deflator
// pkg/replay already uses); then for every month i >= 120 subtract
// (sp500[i]/sp500[i-120])^(1/10) from (scv[i]/scv[i-120])^(1/10) and read the
// result in points. figures_facteurs_test.go recomputes all 635 values from
// pkg/datasets and fails if the figure and the record disagree.
var scvGapPoints = []float64{
	5.27, 5.08, 5.67, 5.81, 5.05, 4.95, 6.28, 6.16, 6.05, 6.26, 5.73, 5.69,
	6.05, 6.13, 6.44, 5.60, 5.47, 5.03, 6.26, 5.62, 5.93, 5.58, 5.91, 6.59,
	6.71, 5.89, 5.77, 5.00, 4.57, 4.19, 4.74, 5.51, 5.41, 4.94, 5.30, 5.13,
	5.23, 5.48, 5.59, 6.06, 6.32, 6.72, 6.58, 6.48, 6.44, 6.73, 6.73, 6.34,
	5.93, 5.81, 5.76, 5.99, 6.69, 6.26, 5.94, 6.64, 7.24, 6.60, 6.37, 6.50,
	6.37, 6.58, 6.28, 4.99, 5.04, 4.77, 5.29, 5.77, 6.07, 6.45, 6.63, 7.51,
	7.95, 8.16, 8.08, 7.15, 7.78, 8.64, 8.40, 7.99, 6.98, 7.54, 8.19, 8.56,
	8.66, 8.70, 7.91, 8.37, 7.89, 7.68, 7.46, 7.43, 7.73, 8.32, 8.63, 9.02,
	9.20, 8.94, 9.06, 9.49, 10.01, 9.70, 8.88, 9.35, 9.69, 9.69, 10.34, 10.56,
	10.89, 10.70, 11.44, 11.84, 12.20, 12.48, 12.95, 13.82, 14.29, 14.55, 15.96, 16.17,
	15.99, 15.93, 15.72, 15.49, 16.69, 17.15, 15.85, 15.58, 15.42, 15.40, 15.99, 15.92,
	15.24, 15.22, 15.14, 15.85, 16.17, 16.94, 15.21, 15.47, 14.92, 15.26, 14.74, 14.59,
	14.36, 15.12, 15.04, 15.41, 15.29, 15.18, 13.97, 12.57, 12.76, 13.02, 13.17, 13.22,
	12.96, 12.87, 13.22, 12.88, 12.18, 11.65, 10.46, 10.44, 10.43, 9.85, 9.55, 9.19,
	9.12, 9.04, 9.02, 7.61, 7.56, 7.18, 6.89, 6.72, 6.95, 7.24, 6.56, 6.57,
	6.57, 6.17, 6.06, 7.25, 6.84, 6.95, 6.22, 6.52, 6.35, 5.90, 5.70, 5.54,
	4.74, 4.63, 4.77, 4.93, 4.44, 3.80, 3.37, 3.79, 4.85, 4.41, 3.51, 3.43,
	2.92, 2.30, 2.06, 1.32, 1.98, 1.78, 1.47, 1.81, 1.86, 1.40, 1.07, 0.89,
	0.91, 1.02, 1.09, 1.09, 0.85, 0.32, 1.64, 1.72, 1.61, 1.22, 1.21, 0.97,
	0.85, 1.18, 0.80, 0.94, 0.74, 1.30, 1.76, 1.22, 1.11, 1.13, 0.53, 0.76,
	0.61, 0.79, 0.88, 1.06, 0.64, 0.69, 0.56, 0.74, 0.68, 0.80, 0.51, 0.71,
	0.82, 0.95, 0.98, 0.58, 0.44, 0.52, 0.22, 0.22, -0.10, -0.24, -0.08, 0.05,
	0.07, 0.34, 0.20, -0.07, -0.08, 0.13, -0.38, -0.13, 0.08, 0.23, 0.74, 0.41,
	0.33, 0.89, 0.22, 0.47, 0.30, 0.70, 0.75, 0.67, 0.84, 0.46, 0.71, 1.10,
	0.78, 1.83, 2.04, 3.24, 2.21, 2.61, 2.14, 1.81, 1.04, 1.03, 1.07, 0.43,
	-0.40, -1.21, -1.33, -1.61, -1.56, -2.01, -2.15, -3.22, -4.09, -3.38, -2.66, -2.64,
	-1.78, -2.19, -2.15, -2.70, -2.22, -1.84, -1.08, 0.45, -0.73, -0.64, -0.41, 0.20,
	1.06, 1.40, 2.44, 3.01, 3.73, 4.63, 4.59, 5.18, 5.25, 4.82, 5.63, 6.25,
	6.38, 6.69, 5.79, 6.07, 6.47, 7.50, 6.64, 6.08, 6.63, 8.23, 7.77, 8.47,
	7.46, 7.42, 7.74, 6.73, 6.60, 6.39, 5.92, 5.68, 5.53, 5.75, 6.30, 6.43,
	6.64, 7.10, 6.94, 7.46, 8.08, 7.72, 8.15, 7.86, 8.09, 7.49, 7.58, 7.81,
	7.66, 7.53, 7.56, 7.95, 8.54, 8.58, 8.72, 8.72, 8.90, 8.35, 8.75, 9.01,
	9.09, 8.62, 8.81, 9.15, 9.25, 9.14, 10.05, 9.73, 9.98, 9.57, 9.06, 9.35,
	9.22, 8.97, 9.07, 9.43, 9.85, 9.46, 9.83, 9.86, 9.54, 9.84, 9.53, 9.21,
	8.75, 7.60, 6.91, 6.62, 6.58, 6.46, 6.89, 6.77, 6.84, 6.38, 6.66, 6.96,
	8.31, 9.59, 9.95, 9.78, 8.84, 9.56, 8.87, 8.88, 9.82, 10.17, 9.32, 9.24,
	9.67, 10.50, 10.97, 10.99, 10.58, 11.13, 10.66, 9.64, 11.03, 12.01, 11.99, 10.60,
	10.34, 9.75, 9.55, 9.69, 9.54, 9.03, 8.47, 7.77, 7.42, 7.47, 6.63, 6.03,
	5.93, 4.93, 5.28, 5.20, 5.04, 4.46, 4.12, 3.90, 3.29, 2.21, 2.40, 1.97,
	2.52, 2.80, 2.56, 3.48, 3.29, 3.45, 3.40, 3.66, 3.73, 3.34, 3.09, 2.98,
	2.81, 2.28, 2.33, 1.75, 1.51, 1.66, 1.09, 1.26, 1.12, 1.26, 1.10, 1.09,
	0.85, 1.02, 0.19, 0.51, -0.36, -0.03, -0.25, -0.37, -0.10, 0.15, -0.15, -0.35,
	-1.41, -1.07, -1.25, -1.47, -1.14, -1.66, -2.40, -2.25, -2.51, -2.11, -2.11, -2.12,
	-1.64, -1.33, -1.02, -1.28, -0.28, -0.08, -0.39, -0.88, -0.87, -0.55, -1.06, -0.78,
	-0.36, -0.42, 0.59, 0.60, 0.98, 0.89, 0.25, 0.32, 0.74, 1.31, 1.47, 1.54,
	0.62, -0.05, -0.62, -0.48, 0.32, -0.61, 0.64, 1.26, 0.30, -0.66, -0.78, -0.90,
	-1.69, -2.90, -2.90, -2.22, -2.02, -2.50, -3.54, -4.15, -6.37, -6.94, -6.82, -6.06,
	-6.38, -6.18, -6.51, -5.54, -5.13, -4.92, -3.51, -3.05, -2.92, -3.16, -2.37, -2.77,
	-3.34, -2.89, -1.91, -2.43, -2.78, -2.81, -2.45, -1.71, -2.04, -1.63, -1.21, -1.62,
	-1.34, -1.28, -1.50, -1.05, -1.39, -1.74, -1.62, -1.54, -3.07, -3.30, -4.11, -4.11,
	-3.79, -4.11, -4.47, -4.82, -4.87, -3.78, -4.51, -4.87, -4.89, -4.75, -4.60, -5.39,
	-3.79, -4.33, -4.02, -4.31, -3.40, -4.27, -4.15, -4.30, -4.65, -4.87, -4.95, -5.29,
	-4.47, -4.08, -4.16, -4.28, -4.27, -3.67, -2.94, -2.78, -2.41, -3.04, -3.25,
}

// The four windows the plate names, and the two full-period real CAGRs its
// footnote quotes (1963-07 to 2026-05, in % a year). All of them are argmax /
// argmin readings of scvGapPoints, not recollections; the guard test checks that
// each one is still the extreme of its era.
const (
	scvPeakMonth   = 1983*12 + 11 // widest window, 1973-12 to 1983-12: +17.15 pt
	scvPeak2Month  = 2010*12 + 3  // second peak, 2000-04 to 2010-04: +12.01 pt
	scvDipMonth    = 1999*12 + 2  // the 1990s trough, 1989-03 to 1999-03: -4.09 pt
	scvTroughMonth = 2020*12 + 3  // deepest window, 2010-04 to 2020-04: -6.94 pt

	scvFullCAGR   = 10.87 // small-cap value, real, % a year
	spFullCAGR    = 6.68  // S&P 500, real, % a year
	scvShareAbove = 76.1  // share of the 635 windows above zero, in %

	// What the broad market itself did over the two winning windows, in % a
	// year real: this is what makes the plate's claim checkable rather than
	// decorative, so the footnote quotes both.
	scvPeakIndexCAGR  = 2.17
	scvPeak2IndexCAGR = -2.57
)

// scvSeg is one run of the gap curve with a constant sign, in pixel space; a run
// bounded by a crossing starts or ends exactly on the zero rule.
type scvSeg struct {
	pts [][2]float64
	pos bool
}

// scvSegments splits the gap into sign runs, interpolating the month-to-month
// zero crossing so the two fills meet on the rule instead of overlapping it.
func scvSegments(vals []float64, x, y func(float64) float64) []scvSeg {
	if len(vals) == 0 {
		return nil
	}
	var out []scvSeg
	cur := scvSeg{pos: vals[0] >= 0, pts: [][2]float64{{x(0), y(vals[0])}}}
	for i := 1; i < len(vals); i++ {
		prev, v := vals[i-1], vals[i]
		if (v >= 0) != (prev >= 0) {
			cross := [2]float64{x(float64(i-1) + prev/(prev-v)), y(0)}
			cur.pts = append(cur.pts, cross)
			out = append(out, cur)
			cur = scvSeg{pos: v >= 0, pts: [][2]float64{cross}}
		}
		cur.pts = append(cur.pts, [2]float64{x(float64(i)), y(v)})
	}
	return append(out, cur)
}

// scvArea closes a run onto the zero rule and fills it.
func scvArea(pts [][2]float64, yZero float64, fill string) string {
	var d strings.Builder
	fmt.Fprintf(&d, "M %.1f,%.1f", pts[0][0], yZero)
	for _, p := range pts {
		fmt.Fprintf(&d, " L %.1f,%.1f", p[0], p[1])
	}
	fmt.Fprintf(&d, " L %.1f,%.1f Z", pts[len(pts)-1][0], yZero)
	return fmt.Sprintf(`<path d="%s" fill="%s"/>`, d.String(), fill)
}

// scvGapLabel prints a gap in the book's French convention, minus sign included.
func scvGapLabel(v float64) string {
	if v < 0 {
		return "−" + frNum(-v, 1) + " pt"
	}
	return "+" + frNum(v, 1) + " pt"
}

func figScvEcart10Ans() string {
	const (
		x0, x1     = 62.0, 616.0
		yTop, yBot = 96.0, 330.0
		vMax, vMin = 20.0, -8.0
	)
	n := float64(len(scvGapPoints) - 1)
	x := func(i float64) float64 { return x0 + i/n*(x1-x0) }
	y := func(v float64) float64 { return yTop + (vMax-v)/(vMax-vMin)*(yBot-yTop) }
	// at reads the window that a month closes, and where the plate draws it.
	at := func(month int) (float64, float64, float64) {
		v := scvGapPoints[month-scvGapStart]
		return x(float64(month - scvGapStart)), y(v), v
	}
	dot := func(px, py float64, col string) string {
		return fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="3.4" fill="%s"/>`, px, py, col)
	}
	yz := y(0)

	var b strings.Builder
	b.WriteString(plateHead("les facteurs en retrait", "Pas les mêmes décennies perdues"))
	b.WriteString(sTxt(24, 62, 10.2, figMuted, "start", "400",
		"écart de rendement réel annualisé sur dix ans glissants, small-cap value américain moins S&amp;P 500, en points"))

	// Grid, the zero rule carrying the reading of the whole plate.
	for _, g := range []float64{20, 15, 10, 5, 0, -5} {
		gy := y(g)
		col := figGrid
		lab := "+" + frNum(g, 0)
		switch {
		case g == 0:
			col, lab = figRule, "0"
		case g < 0:
			lab = "−" + frNum(-g, 0)
		}
		b.WriteString(line(x0, gy, x1, gy, col, 1))
		b.WriteString(mTxt(x0-8, gy+3.5, 10, figMuted, "end", "400", lab))
	}

	// The area, one solid tint per side, then the curve drawn over it.
	segs := scvSegments(scvGapPoints, x, y)
	for _, s := range segs {
		fill := figScvAreaDown
		if s.pos {
			fill = figScvAreaUp
		}
		b.WriteString(scvArea(s.pts, yz, fill))
	}
	b.WriteString(line(x0, yz, x1, yz, figRule, 1.2))
	for _, s := range segs {
		col := figBad
		if s.pos {
			col = figAccent
		}
		b.WriteString(poly(s.pts, col, 1.4, ""))
	}

	// The two decades the tilt won, both of them lost decades of the index.
	px, py, pv := at(scvPeakMonth)
	b.WriteString(dot(px, py, figDeep))
	b.WriteString(mTxt(px, 106, 10.5, figDeep, "middle", "600", scvGapLabel(pv)))
	b.WriteString(sTxt(px+34, 106, 10.5, figSoft, "start", "600",
		"les sommets du tilt : les décennies perdues du marché large"))
	p2x, p2y, p2v := at(scvPeak2Month)
	b.WriteString(dot(p2x, p2y, figDeep))
	b.WriteString(mTxt(p2x+9, p2y-4, 10.5, figDeep, "start", "600", scvGapLabel(p2v)))

	// The two decades it lost, the entry price the article calls by its name.
	dx, dy, dv := at(scvDipMonth)
	b.WriteString(dot(dx, dy, figBad))
	b.WriteString(sTxt(dx-19, dy-6, 10.5, figSoft, "end", "600", "creux de 1999"))
	b.WriteString(mTxt(dx-19, dy+8, 10.5, figBad, "end", "600", scvGapLabel(dv)))
	tx, ty, tv := at(scvTroughMonth)
	b.WriteString(dot(tx, ty, figBad))
	b.WriteString(sTxt(tx-18, ty-22, 10.5, figSoft, "end", "600", "purgatoire de la value"))
	b.WriteString(mTxt(tx-18, ty-8, 10.5, figBad, "end", "600", scvGapLabel(tv)))

	// The x axis says which of the window's two dates is plotted.
	for yr := 1975; yr <= 2025; yr += 5 {
		b.WriteString(mTxt(x(float64(yr*12-scvGapStart)), 348, 10, figMuted, "middle", "400", frNum(float64(yr), 0)))
	}
	b.WriteString(sTxt(x1, 366, 10.5, figMuted, "end", "400", "année de fin de la fenêtre de dix ans"))

	// Where the reader stands today, dotted back to the last window plotted.
	last := scvGapPoints[len(scvGapPoints)-1]
	b.WriteString(dot(x1, y(last), figBad))
	b.WriteString(dashLine(x1, y(last)-7, x1, 234, figBad, 1, "2 2"))
	b.WriteString(mTxt(x1-2, 228, 10, figBad, "end", "600", "mai 2026 : "+scvGapLabel(last)))

	for i, s := range []string{
		fmt.Sprintf("Fenêtres de 120 mois, chacune datée du mois qui la ferme. Aux deux sommets, le S&amp;P 500 réel ne fait que +%s puis −%s %%/an.",
			frNum(scvPeakIndexCAGR, 1), frNum(-scvPeak2IndexCAGR, 1)),
		"Petites capitalisations décotées de Kenneth French, brut de frais, contre S&amp;P 500 total return, déflatés par le CPI américain.",
		fmt.Sprintf("Sur 1963-2026, %s %%/an réel contre %s %%/an, et %s %% des fenêtres au-dessus de zéro : un portefeuille académique, pas un fonds.",
			frNum(scvFullCAGR, 1), frNum(spFullCAGR, 1), frNum(scvShareAbove, 0)),
	} {
		b.WriteString(sTxt(24, 388+float64(i)*13, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, 428, b.String())
}
