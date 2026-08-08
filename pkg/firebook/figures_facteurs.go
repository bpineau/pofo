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
// The small-value leg is NET of the cost the factor does not pay: an academic
// portfolio charges no fee, no commission and no spread, and in small caps that
// is worth 1.0 %/yr, measured over 399 months against a real small-value fund
// (simgen.USSCVGrossCost, the same haircut the investable reconstruction takes).
// It stands here against an index a tracker exists for, and the plate reads its
// footnote in achievable points, so the raw factor would flatter the tilt by
// about a point a year throughout.
//
// Reproduce: read the bundled USSCV-USD (Kenneth French's value-weighted small
// x high book-to-market portfolio, daily since 1963-07) and SP500-USD (S&P 500
// total return, month-end) with marketdata.ReadSimdataFS(datasets.Refdata(), id)
// and keep the last quote of each calendar month; charge the small-value leg
// (1 - simgen.USSCVGrossCost)^(1/12) a month; divide both by the ^CPI-US
// snapshot level of the same month (marketdata.NewClient("").Fetch, the deflator
// pkg/replay already uses); then for every month i >= 120 subtract
// (sp500[i]/sp500[i-120])^(1/10) from (scv[i]/scv[i-120])^(1/10) and read the
// result in points. figures_facteurs_test.go recomputes all 635 values from
// pkg/datasets and fails if the figure and the record disagree.
var scvGapPoints = []float64{
	4.18, 4.00, 4.58, 4.73, 3.99, 3.89, 5.20, 5.09, 4.99, 5.20, 4.68, 4.64,
	5.01, 5.10, 5.42, 4.58, 4.45, 4.02, 5.23, 4.59, 4.90, 4.55, 4.87, 5.54,
	5.65, 4.85, 4.74, 3.97, 3.55, 3.17, 3.71, 4.47, 4.37, 3.90, 4.25, 4.08,
	4.17, 4.42, 4.52, 5.00, 5.26, 5.65, 5.52, 5.43, 5.39, 5.68, 5.68, 5.29,
	4.90, 4.77, 4.73, 4.96, 5.64, 5.22, 4.91, 5.60, 6.19, 5.56, 5.33, 5.46,
	5.33, 5.54, 5.25, 3.97, 4.03, 3.76, 4.27, 4.74, 5.04, 5.42, 5.60, 6.46,
	6.89, 7.09, 7.01, 6.10, 6.72, 7.57, 7.32, 6.92, 5.93, 6.47, 7.11, 7.46,
	7.57, 7.61, 6.82, 7.28, 6.80, 6.60, 6.39, 6.37, 6.66, 7.25, 7.55, 7.94,
	8.12, 7.87, 7.99, 8.41, 8.92, 8.63, 7.82, 8.29, 8.63, 8.63, 9.28, 9.49,
	9.82, 9.62, 10.36, 10.74, 11.10, 11.38, 11.83, 12.69, 13.15, 13.40, 14.79, 14.99,
	14.82, 14.76, 14.55, 14.33, 15.50, 15.96, 14.67, 14.41, 14.24, 14.21, 14.80, 14.73,
	14.05, 14.00, 13.91, 14.63, 14.94, 15.70, 13.99, 14.26, 13.71, 14.06, 13.54, 13.40,
	13.16, 13.91, 13.83, 14.20, 14.08, 13.96, 12.77, 11.38, 11.57, 11.82, 11.96, 12.01,
	11.77, 11.67, 12.03, 11.68, 10.98, 10.47, 9.27, 9.24, 9.23, 8.65, 8.36, 8.00,
	7.92, 7.84, 7.82, 6.45, 6.41, 6.02, 5.73, 5.55, 5.79, 6.08, 5.41, 5.40,
	5.42, 5.03, 4.91, 6.08, 5.68, 5.79, 5.05, 5.35, 5.18, 4.74, 4.53, 4.38,
	3.58, 3.47, 3.60, 3.76, 3.28, 2.65, 2.23, 2.64, 3.68, 3.25, 2.35, 2.28,
	1.78, 1.18, 0.95, 0.22, 0.88, 0.68, 0.36, 0.69, 0.74, 0.28, -0.05, -0.23,
	-0.21, -0.11, -0.05, -0.05, -0.27, -0.82, 0.49, 0.56, 0.46, 0.07, 0.05, -0.19,
	-0.31, 0.04, -0.34, -0.19, -0.39, 0.17, 0.62, 0.09, -0.01, 0.02, -0.58, -0.35,
	-0.50, -0.33, -0.24, -0.06, -0.47, -0.43, -0.56, -0.38, -0.43, -0.32, -0.61, -0.41,
	-0.30, -0.16, -0.14, -0.54, -0.67, -0.59, -0.88, -0.88, -1.20, -1.34, -1.19, -1.06,
	-1.04, -0.77, -0.92, -1.18, -1.20, -0.98, -1.48, -1.23, -1.02, -0.87, -0.37, -0.69,
	-0.77, -0.21, -0.89, -0.64, -0.82, -0.42, -0.36, -0.43, -0.26, -0.65, -0.41, -0.02,
	-0.34, 0.71, 0.91, 2.07, 1.04, 1.45, 0.98, 0.65, -0.13, -0.13, -0.09, -0.72,
	-1.54, -2.33, -2.46, -2.74, -2.70, -3.15, -3.28, -4.34, -5.20, -4.50, -3.78, -3.76,
	-2.90, -3.31, -3.27, -3.81, -3.34, -2.97, -2.22, -0.71, -1.88, -1.79, -1.54, -0.95,
	-0.10, 0.22, 1.25, 1.82, 2.54, 3.44, 3.40, 4.00, 4.08, 3.65, 4.45, 5.07,
	5.20, 5.52, 4.64, 4.91, 5.30, 6.33, 5.47, 4.92, 5.46, 7.06, 6.60, 7.30,
	6.31, 6.26, 6.60, 5.59, 5.46, 5.26, 4.79, 4.56, 4.42, 4.62, 5.16, 5.29,
	5.50, 5.95, 5.80, 6.31, 6.92, 6.56, 6.98, 6.69, 6.92, 6.33, 6.41, 6.64,
	6.50, 6.38, 6.40, 6.79, 7.37, 7.40, 7.54, 7.55, 7.73, 7.19, 7.58, 7.85,
	7.93, 7.46, 7.66, 7.99, 8.09, 7.98, 8.88, 8.57, 8.82, 8.41, 7.91, 8.20,
	8.06, 7.82, 7.92, 8.27, 8.70, 8.30, 8.68, 8.71, 8.39, 8.69, 8.39, 8.07,
	7.63, 6.49, 5.80, 5.51, 5.48, 5.36, 5.80, 5.69, 5.77, 5.31, 5.59, 5.89,
	7.22, 8.48, 8.85, 8.70, 7.79, 8.50, 7.83, 7.85, 8.77, 9.12, 8.27, 8.19,
	8.61, 9.42, 9.88, 9.92, 9.51, 10.05, 9.59, 8.57, 9.95, 10.91, 10.90, 9.54,
	9.27, 8.69, 8.48, 8.62, 8.46, 7.95, 7.39, 6.69, 6.34, 6.39, 5.57, 4.97,
	4.87, 3.88, 4.22, 4.13, 3.98, 3.41, 3.07, 2.85, 2.25, 1.17, 1.36, 0.92,
	1.46, 1.74, 1.48, 2.40, 2.22, 2.37, 2.31, 2.56, 2.63, 2.26, 2.01, 1.90,
	1.73, 1.21, 1.26, 0.68, 0.44, 0.60, 0.03, 0.21, 0.06, 0.20, 0.03, 0.02,
	-0.22, -0.05, -0.87, -0.55, -1.42, -1.09, -1.30, -1.42, -1.16, -0.92, -1.21, -1.41,
	-2.45, -2.11, -2.29, -2.51, -2.18, -2.70, -3.42, -3.28, -3.54, -3.14, -3.14, -3.15,
	-2.68, -2.37, -2.06, -2.31, -1.32, -1.13, -1.44, -1.92, -1.92, -1.60, -2.10, -1.82,
	-1.41, -1.47, -0.47, -0.47, -0.09, -0.18, -0.84, -0.76, -0.34, 0.22, 0.38, 0.43,
	-0.48, -1.14, -1.72, -1.58, -0.81, -1.72, -0.50, 0.10, -0.84, -1.79, -1.90, -2.02,
	-2.80, -3.99, -3.99, -3.31, -3.12, -3.60, -4.63, -5.22, -7.40, -7.97, -7.87, -7.12,
	-7.44, -7.25, -7.57, -6.59, -6.20, -5.99, -4.59, -4.14, -4.01, -4.25, -3.47, -3.87,
	-4.43, -4.00, -3.04, -3.55, -3.89, -3.92, -3.55, -2.81, -3.14, -2.72, -2.31, -2.70,
	-2.43, -2.37, -2.57, -2.14, -2.48, -2.82, -2.70, -2.62, -4.13, -4.36, -5.16, -5.17,
	-4.85, -5.17, -5.52, -5.85, -5.91, -4.84, -5.56, -5.92, -5.94, -5.80, -5.65, -6.44,
	-4.85, -5.39, -5.08, -5.36, -4.47, -5.33, -5.21, -5.35, -5.69, -5.91, -5.99, -6.34,
	-5.53, -5.15, -5.23, -5.35, -5.34, -4.75, -4.03, -3.87, -3.49, -4.13, -4.33,
}

// The four windows the plate names, and the two full-period real CAGRs its
// footnote quotes (1963-07 to 2026-05, in % a year). All of them are argmax /
// argmin readings of scvGapPoints, not recollections; the guard test checks that
// each one is still the extreme of its era.
const (
	scvPeakMonth   = 1983*12 + 11 // widest window, 1973-12 to 1983-12: +15.96 pt
	scvPeak2Month  = 2010*12 + 3  // second peak, 2000-04 to 2010-04: +10.91 pt
	scvDipMonth    = 1999*12 + 2  // the 1990s trough, 1989-03 to 1999-03: -5.20 pt
	scvTroughMonth = 2020*12 + 3  // deepest window, 2010-04 to 2020-04: -7.97 pt

	scvFullCAGR   = 9.76 // small-cap value net of its cost, real, % a year
	spFullCAGR    = 6.68 // S&P 500, real, % a year
	scvShareAbove = 63.3 // share of the 635 windows above zero, in %

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
		vMax, vMin = 20.0, -10.0
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
		"Petites capitalisations décotées de Kenneth French moins 1,0 point par an de coûts mesurés, contre S&amp;P 500 total return, en réel.",
		fmt.Sprintf("Sur 1963-2026, %s %%/an réel contre %s %%/an, et %s %% des fenêtres au-dessus de zéro : une prime réelle, qu'une fenêtre sur trois perd.",
			frNum(scvFullCAGR, 1), frNum(spFullCAGR, 1), frNum(scvShareAbove, 0)),
	} {
		b.WriteString(sTxt(24, 388+float64(i)*13, 9.5, figMuted, "start", "400", s))
	}
	return svg(640, 428, b.String())
}
