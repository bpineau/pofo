package chart

import (
	"fmt"
	"math"
	"strings"
)

// MatrixColumn is one column of a BarMatrix: a titled group holding one value
// per row (e.g. one macro regime with each asset's contribution in it).
type MatrixColumn struct {
	Title    string
	Subtitle string // small line under the title (e.g. "56 months · 22%")
	Color    string // header swatch; empty = no swatch
	Values   []float64
}

// BarMatrixOptions styles a BarMatrix. RowColors defaults to the palette in
// row order. An optional Summary row totals each column; it is drawn clamped
// to the shared row scale (its label always carries the true value).
type BarMatrixOptions struct {
	Title     string
	Width     int // default 1200
	RowLabels []string
	RowColors []string
	Unit      string // value unit for tooltips and labels (e.g. "pts/yr")

	Summary      []float64 // optional bottom row, one value per column
	SummaryLabel string    // its row label (default "total")
}

// BarMatrix renders a small-multiples matrix of horizontal diverging bars:
// each column is a category, each row an entity, every cell a signed value
// drawn from a per-column zero axis on one shared scale, so magnitudes
// compare across the whole grid. Cells carry instant-tooltip data-tip
// attributes and values at or above 5 % of the scale are labeled directly.
// Returns "" without at least one column and row.
func BarMatrix(opt BarMatrixOptions, cols []MatrixColumn) string {
	nR := len(opt.RowLabels)
	if len(cols) == 0 || nR == 0 {
		return ""
	}
	w := opt.Width
	if w == 0 {
		w = 1200
	}
	const padL, padT = 150, 92
	const rowH, cellGap, padR = 24, 26, 30

	lo, hi := 0.0, 0.0
	for _, c := range cols {
		for s := 0; s < nR && s < len(c.Values); s++ {
			lo, hi = math.Min(lo, c.Values[s]), math.Max(hi, c.Values[s])
		}
	}
	if lo == hi {
		lo, hi = -1, 1
	}
	step := niceStep(hi-lo, 8)
	lo, hi = math.Floor(lo/step)*step, math.Ceil(hi/step)*step

	nRows := nR
	if len(opt.Summary) > 0 {
		nRows++
	}
	h := padT + nRows*rowH + 40
	colW := (w - padL - padR - (len(cols)-1)*cellGap) / len(cols)
	xin := func(q int, v float64) float64 {
		return float64(padL+q*(colW+cellGap)) + float64(colW)*(v-lo)/(hi-lo)
	}
	rowColor := func(s int) string {
		if s < len(opt.RowColors) && opt.RowColors[s] != "" {
			return opt.RowColors[s]
		}
		return PaletteColor(s)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeMono+`">`, w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="`+themeSurface+`"/>`, w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="26" y="20" font-size="13" font-weight="600" fill="`+themeInk+`">%s</text>`, esc(opt.Title))
	}

	// Column headers, zero axes and scale labels.
	for qi, c := range cols {
		cx := padL + qi*(colW+cellGap)
		if c.Color != "" {
			fmt.Fprintf(&b, `<rect x="%d" y="46" width="8" height="8" rx="2" fill="%s" fill-opacity="0.75"/>`, cx, c.Color)
			fmt.Fprintf(&b, `<text x="%d" y="54" font-size="12" font-weight="600" fill="`+themeInk+`">%s</text>`, cx+13, esc(c.Title))
		} else {
			fmt.Fprintf(&b, `<text x="%d" y="54" font-size="12" font-weight="600" fill="`+themeInk+`">%s</text>`, cx, esc(c.Title))
		}
		if c.Subtitle != "" {
			fmt.Fprintf(&b, `<text x="%d" y="70" font-size="10" fill="`+themeMuted+`">%s</text>`, cx, esc(c.Subtitle))
		}
		zx := xin(qi, 0)
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%d" x2="%.1f" y2="%d" stroke="`+themeAxis+`" stroke-width="1.2"/>`, zx, padT-6, zx, padT+nRows*rowH+2)
		for _, v := range []float64{lo, hi} {
			fmt.Fprintf(&b, `<text x="%.1f" y="%d" text-anchor="middle" font-size="9" fill="`+themeMuted+`">%+.4g</text>`, xin(qi, v), padT+nRows*rowH+16, v)
		}
	}

	// Rows: label + one diverging bar per column, zebra background.
	tip := func(row, col string, v float64) string {
		return esc(fmt.Sprintf("%s · %s: %+.1f %s", row, col, v, opt.Unit))
	}
	labelThreshold := 0.05 * (hi - lo)
	for s := range nR {
		yy := padT + s*rowH
		fmt.Fprintf(&b, `<rect x="26" y="%d" width="10" height="10" rx="2" fill="%s"/>`, yy+2, rowColor(s))
		fmt.Fprintf(&b, `<text x="42" y="%d" font-size="11" fill="`+themeInkSoft+`">%s</text>`, yy+11, esc(opt.RowLabels[s]))
		if s%2 == 0 {
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="`+themeGrid+`" fill-opacity="0.5"/>`, padL-8, yy-4, w-padL-20, rowH)
		}
		for qi, c := range cols {
			v := 0.0
			if s < len(c.Values) {
				v = c.Values[s]
			}
			x0, x1 := xin(qi, math.Min(0, v)), xin(qi, math.Max(0, v))
			if x1-x0 < 0.8 {
				x0, x1 = xin(qi, 0)-0.4, xin(qi, 0)+0.4
			}
			fmt.Fprintf(&b, `<rect x="%.1f" y="%d" width="%.1f" height="10" rx="2" fill="%s" data-tip="%s"/>`,
				x0, yy+2, x1-x0, rowColor(s), tip(opt.RowLabels[s], c.Title, v))
			if math.Abs(v) >= labelThreshold {
				tx, anch := x1+4, "start"
				if v < 0 {
					tx, anch = x0-4, "end"
				}
				fmt.Fprintf(&b, `<text x="%.1f" y="%d" text-anchor="%s" font-size="9" fill="`+themeMuted+`">%+.1f</text>`, tx, yy+11, anch, v)
			}
		}
	}

	// Summary row: clamped into the shared scale, the label says the truth.
	if len(opt.Summary) > 0 {
		label := opt.SummaryLabel
		if label == "" {
			label = "total"
		}
		yy := padT + nR*rowH + 6
		fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="`+themeAxis+`"/>`, padL-8, yy-5, w-padR, yy-5)
		fmt.Fprintf(&b, `<text x="42" y="%d" font-size="11" font-weight="600" fill="`+themeInk+`">%s</text>`, yy+11, esc(label))
		for qi, c := range cols {
			if qi >= len(opt.Summary) {
				break
			}
			t := opt.Summary[qi]
			tc := math.Max(lo, math.Min(hi, t))
			x0, x1 := xin(qi, math.Min(0, tc)), xin(qi, math.Max(0, tc))
			fmt.Fprintf(&b, `<rect x="%.1f" y="%d" width="%.1f" height="10" rx="2" fill="`+themeInkSoft+`" fill-opacity="0.55" data-tip="%s"/>`,
				x0, yy+2, math.Max(x1-x0, 0.8), tip(label, c.Title, t))
			tx, anch := x1+4, "start"
			if t < 0 {
				tx, anch = x0-4, "end"
			}
			fmt.Fprintf(&b, `<text x="%.1f" y="%d" text-anchor="%s" font-size="9" font-weight="600" fill="`+themeInk+`">%+.1f</text>`, tx, yy+11, anch, t)
		}
	}
	b.WriteString(`</svg>`)
	return finish(b.String())
}
