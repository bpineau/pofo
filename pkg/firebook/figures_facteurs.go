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
// window from July 1973 to August 2026 (638 windows), each dated by the month that
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
// result in points. figures_facteurs_test.go recomputes all 638 values from
// pkg/datasets and fails if the figure and the record disagree.
var scvGapPoints = []float64{
	4.16, 3.98, 4.57, 4.71, 3.96, 3.86, 5.19, 5.09, 4.98, 5.20, 4.67, 4.64,
	5.01, 5.10, 5.42, 4.57, 4.45, 4.01, 5.23, 4.59, 4.89, 4.54, 4.86, 5.53,
	5.65, 4.85, 4.73, 3.97, 3.54, 3.17, 3.70, 4.46, 4.36, 3.90, 4.25, 4.08,
	4.17, 4.42, 4.51, 4.99, 5.25, 5.65, 5.52, 5.43, 5.39, 5.68, 5.68, 5.29,
	4.90, 4.77, 4.73, 4.95, 5.64, 5.22, 4.91, 5.60, 6.19, 5.56, 5.33, 5.46,
	5.33, 5.54, 5.25, 3.97, 4.03, 3.76, 4.27, 4.74, 5.04, 5.41, 5.60, 6.46,
	6.89, 7.09, 7.01, 6.10, 6.72, 7.56, 7.32, 6.92, 5.93, 6.47, 7.11, 7.46,
	7.57, 7.61, 6.82, 7.28, 6.80, 6.60, 6.39, 6.37, 6.66, 7.25, 7.55, 7.94,
	8.12, 7.87, 7.99, 8.41, 8.92, 8.63, 7.82, 8.29, 8.63, 8.63, 9.28, 9.49,
	9.82, 9.62, 10.36, 10.74, 11.10, 11.38, 11.83, 12.69, 13.15, 13.40, 14.79, 14.99,
	14.82, 14.76, 14.55, 14.33, 15.50, 15.96, 14.67, 14.40, 14.24, 14.21, 14.80, 14.73,
	14.03, 13.98, 13.87, 14.59, 14.90, 15.65, 13.96, 14.21, 13.66, 14.01, 13.49, 13.33,
	13.09, 13.84, 13.76, 14.13, 14.01, 13.90, 12.70, 11.32, 11.50, 11.76, 11.89, 11.95,
	11.70, 11.60, 11.96, 11.61, 10.92, 10.40, 9.21, 9.17, 9.16, 8.59, 8.29, 7.94,
	7.86, 7.77, 7.76, 6.38, 6.35, 5.96, 5.67, 5.49, 5.73, 6.01, 5.34, 5.34,
	5.35, 4.97, 4.84, 6.01, 5.59, 5.70, 4.96, 5.27, 5.10, 4.66, 4.46, 4.31,
	3.51, 3.40, 3.53, 3.69, 3.21, 2.58, 2.16, 2.57, 3.61, 3.18, 2.28, 2.21,
	1.71, 1.11, 0.89, 0.15, 0.81, 0.61, 0.29, 0.62, 0.67, 0.21, -0.12, -0.29,
	-0.28, -0.18, -0.12, -0.12, -0.34, -0.89, 0.42, 0.49, 0.39, -0.00, -0.02, -0.26,
	-0.38, -0.03, -0.41, -0.25, -0.45, 0.10, 0.56, 0.02, -0.08, -0.05, -0.65, -0.41,
	-0.57, -0.40, -0.30, -0.13, -0.54, -0.49, -0.63, -0.45, -0.50, -0.38, -0.68, -0.48,
	-0.35, -0.21, -0.17, -0.56, -0.70, -0.61, -0.92, -0.91, -1.22, -1.36, -1.21, -1.06,
	-1.05, -0.78, -0.93, -1.19, -1.20, -0.98, -1.49, -1.24, -1.03, -0.88, -0.38, -0.70,
	-0.78, -0.22, -0.90, -0.65, -0.83, -0.43, -0.37, -0.44, -0.27, -0.66, -0.42, -0.03,
	-0.35, 0.70, 0.90, 2.06, 1.03, 1.43, 0.97, 0.64, -0.14, -0.14, -0.10, -0.74,
	-1.55, -2.34, -2.45, -2.73, -2.68, -3.13, -3.25, -4.32, -5.19, -4.50, -3.77, -3.76,
	-2.90, -3.29, -3.26, -3.80, -3.32, -2.96, -2.20, -0.68, -1.85, -1.76, -1.52, -0.92,
	-0.07, 0.25, 1.28, 1.84, 2.57, 3.47, 3.43, 4.03, 4.11, 3.67, 4.48, 5.09,
	5.23, 5.55, 4.66, 4.94, 5.32, 6.35, 5.50, 4.95, 5.49, 7.08, 6.63, 7.32,
	6.34, 6.29, 6.62, 5.62, 5.48, 5.28, 4.82, 4.59, 4.44, 4.65, 5.19, 5.32,
	5.53, 5.98, 5.82, 6.33, 6.94, 6.59, 7.01, 6.72, 6.94, 6.35, 6.44, 6.67,
	6.52, 6.40, 6.42, 6.81, 7.39, 7.42, 7.56, 7.57, 7.75, 7.21, 7.61, 7.88,
	7.96, 7.50, 7.69, 8.02, 8.12, 8.01, 8.92, 8.60, 8.85, 8.45, 7.94, 8.23,
	8.10, 7.85, 7.95, 8.30, 8.73, 8.34, 8.71, 8.74, 8.42, 8.73, 8.42, 8.11,
	7.66, 6.50, 5.81, 5.52, 5.48, 5.37, 5.83, 5.72, 5.81, 5.34, 5.60, 5.91,
	7.24, 8.49, 8.86, 8.71, 7.80, 8.51, 7.84, 7.87, 8.79, 9.14, 8.28, 8.20,
	8.63, 9.43, 9.89, 9.91, 9.51, 10.06, 9.59, 8.55, 9.94, 10.90, 10.90, 9.53,
	9.26, 8.68, 8.47, 8.61, 8.45, 7.94, 7.38, 6.68, 6.33, 6.38, 5.56, 4.96,
	4.86, 3.87, 4.21, 4.13, 3.97, 3.40, 3.06, 2.84, 2.24, 1.16, 1.35, 0.91,
	1.45, 1.73, 1.48, 2.40, 2.22, 2.37, 2.31, 2.56, 2.64, 2.27, 2.02, 1.91,
	1.74, 1.22, 1.26, 0.69, 0.45, 0.61, 0.04, 0.21, 0.07, 0.21, 0.04, 0.03,
	-0.20, -0.04, -0.86, -0.55, -1.41, -1.08, -1.30, -1.41, -1.15, -0.90, -1.20, -1.39,
	-2.43, -2.10, -2.27, -2.50, -2.17, -2.68, -3.41, -3.26, -3.52, -3.12, -3.13, -3.14,
	-2.67, -2.35, -2.05, -2.30, -1.31, -1.12, -1.42, -1.91, -1.91, -1.59, -2.09, -1.81,
	-1.39, -1.44, -0.43, -0.42, -0.05, -0.14, -0.81, -0.74, -0.34, 0.24, 0.41, 0.46,
	-0.46, -1.12, -1.70, -1.56, -0.78, -1.69, -0.47, 0.11, -0.83, -1.77, -1.87, -2.00,
	-2.78, -3.97, -3.97, -3.29, -3.10, -3.58, -4.61, -5.20, -7.38, -7.96, -7.85, -7.11,
	-7.42, -7.23, -7.55, -6.58, -6.18, -5.98, -4.57, -4.12, -4.00, -4.23, -3.45, -3.85,
	-4.42, -3.98, -3.02, -3.53, -3.87, -3.90, -3.53, -2.79, -3.12, -2.70, -2.29, -2.69,
	-2.42, -2.36, -2.56, -2.13, -2.47, -2.81, -2.69, -2.61, -4.12, -4.36, -5.17, -5.17,
	-4.85, -5.17, -5.51, -5.85, -5.91, -4.84, -5.56, -5.92, -5.94, -5.79, -5.65, -6.44,
	-4.86, -5.39, -5.09, -5.36, -4.47, -5.32, -5.20, -5.36, -5.70, -5.91, -6.00, -6.35,
	-5.54, -5.18, -5.26, -5.38, -5.37, -4.77, -4.05, -3.88, -3.52, -4.15, -4.37, -3.89,
	-4.12, -4.75,
}

// The four windows the plate names, and the two full-period real CAGRs its
// footnote quotes (1963-07 to 2026-08, in % a year). All of them are argmax /
// argmin readings of scvGapPoints, not recollections; the guard test checks that
// each one is still the extreme of its era.
const (
	scvPeakMonth   = 1983*12 + 11 // widest window, 1973-12 to 1983-12: +15.96 pt
	scvPeak2Month  = 2010*12 + 3  // second peak, 2000-04 to 2010-04: +10.90 pt
	scvDipMonth    = 1999*12 + 2  // the 1990s trough, 1989-03 to 1999-03: -5.19 pt
	scvTroughMonth = 2020*12 + 3  // deepest window, 2010-04 to 2020-04: -7.96 pt

	scvFullCAGR   = 9.75 // small-cap value net of its cost, real, % a year
	spFullCAGR    = 6.68 // S&P 500, real, % a year
	scvShareAbove = 62.4 // share of the 638 windows above zero, in %

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
	b.WriteString(mTxt(x1-2, 228, 10, figBad, "end", "600", "août 2026 : "+scvGapLabel(last)))

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
