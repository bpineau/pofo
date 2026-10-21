package chart

import (
	"fmt"
	"math"
	"strings"
)

// AreaSeries is one layer of a stacked-area chart: values over the x index,
// stacked bottom-up in slice order.
type AreaSeries struct {
	Name   string    // legend label
	Values []float64 // one value per x position, in the chart's unit
	Color  string    // CSS color; picked from the default palette when empty
	Weight float64   // fill opacity, 0..1; 0 means the default 0.5
	// Outline draws the layer's two boundaries as crisp lines of its own
	// colour, on top of everything, under StackOptions.Gaps (where the other
	// layers are only parted by the surface). It is for the one or two
	// frontiers the chart is about. Without Gaps every layer is outlined.
	Outline bool
}

// StackOptions are the optional layers of StackedAreaWith; the zero value is
// exactly StackedArea.
type StackOptions struct {
	XLabel, YLabel string // axis names; the hover layer reads them too

	// XTicks is a tick label per x position, "" for no tick there: an age
	// axis over a year index, say. nil keeps the index at automatic steps.
	XTicks []string
	// XTips is the hover header per x position ("year 12 · age 64"); nil
	// lets the front end print the x label and the index.
	XTips []string
	// Max fixes the top of the y axis (100 for shares that sum to 100);
	// 0 takes the stack's own top.
	Max float64
	// Unit is appended to the y tick labels and to the end labels ("%").
	Unit string
	// Gaps parts the layers with a hairline of the chart surface (the
	// stacked-segment spacer) instead of a line of each layer's colour, so
	// only the layers marked Outline draw a frontier.
	Gaps bool
	// EndLabels prints each layer's name and last value at the right edge,
	// beside its own band, wherever the band is tall enough to carry it
	// without colliding with its neighbour. The legend still names every
	// layer, so a label that does not fit is only dropped, never lost.
	EndLabels bool
	// Markers are vertical reference lines at x positions (Axis 'x'; other
	// axes are ignored), labelled at the top of the plot.
	Markers []Marker
}

// StackedArea renders layers stacked bottom-up over the x index (year 0..N):
// layer i fills the space between the running sum below it and itself. It is
// meant for part-to-whole shares over time, e.g. the alive-broke-dead
// decomposition of a retirement (pass shares already scaled to percent).
//
// The information in a stack is in its BOUNDARIES, so each layer is a soft
// fill under a crisp line of its own colour: the eye follows the frontiers
// rather than the mass. A layer that is a ground rather than a subject (the
// share where nothing has happened yet) should be given a low Weight, so the
// ink goes to the layers that carry the story.
func StackedArea(opt Options, xLabel, yLabel string, series []AreaSeries) string {
	return StackedAreaWith(opt, StackOptions{XLabel: xLabel, YLabel: yLabel}, series)
}

// StackedAreaWith is StackedArea with the optional layers of StackOptions: a
// labelled x axis (ages over a year index), a fixed y top, surface gaps with a
// chosen frontier or two, direct end labels and reference markers. Every
// layer's values ride in the hover payload (kind "stack", top layer first,
// with the per-position headers of XTips), so the tooltip reads the whole
// stack anywhere over the plot.
func StackedAreaWith(opt Options, so StackOptions, series []AreaSeries) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	marginL, marginR, top, bottom := 64.0, 16.0, 64.0, 48.0
	if so.EndLabels {
		marginR = 132
	}
	x0, x1 := marginL, float64(w)-marginR
	y0, y1 := top, float64(h)-bottom

	steps := 0
	for i := range series {
		if series[i].Color == "" {
			series[i].Color = PaletteColor(i)
		}
		steps = max(steps, len(series[i].Values))
	}

	// Cumulative sums: cum[0] is the zero baseline, cum[i] tops layer i-1.
	cum := make([][]float64, len(series)+1)
	cum[0] = make([]float64, steps)
	for i, s := range series {
		row := make([]float64, steps)
		for j := range row {
			v := 0.0
			if j < len(s.Values) {
				v = s.Values[j]
			}
			row[j] = cum[i][j] + v
		}
		cum[i+1] = row
	}
	vmax := 1.0
	for _, v := range cum[len(series)] {
		vmax = math.Max(vmax, v)
	}
	if so.Max > 0 {
		vmax = so.Max
	}

	xmax := float64(max(steps-1, 1))
	xAt := func(i int) float64 { return x0 + float64(i)/xmax*(x1-x0) }
	yAt := func(v float64) float64 { return y1 - v/vmax*(y1-y0) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeMono+`">`+"\n", w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="`+themeSurface+`"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="24" font-size="16" font-weight="600" fill="`+themeInk+`">%s</text>`+"\n", x0, esc(opt.Title))
	}

	// Layers first, so the grid stays visible on top of the soft fills.
	for i := range series {
		op := series[i].Weight
		if op <= 0 {
			op = 0.5
		}
		if so.Gaps {
			fmt.Fprintf(&b, `<polygon points="%s" fill="%s" fill-opacity="%.2f" stroke="`+themeSurface+`" stroke-width="1.5" stroke-linejoin="round"/>`+"\n",
				bandPolygon(cum[i], cum[i+1], xAt, yAt), series[i].Color, op)
			continue
		}
		fmt.Fprintf(&b, `<polygon points="%s" fill="%s" fill-opacity="%.2f" stroke="%s" stroke-width="1.25" stroke-opacity="0.9"/>`+"\n",
			bandPolygon(cum[i], cum[i+1], xAt, yAt), series[i].Color, op, series[i].Color)
	}

	// Horizontal grid and y labels, drawn over the fills as hairlines. Under
	// Gaps the layers are parted by surface lines already, and a grid running
	// across them would read as more frontiers: the axis carries short ticks
	// instead, and the hover reads the exact shares.
	ystep := niceStep(vmax, 5)
	for v, k := 0.0, 0; v <= vmax+ystep/1e6 && k < maxGridLines; v, k = v+ystep, k+1 {
		y := yAt(v)
		if so.Gaps {
			fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="`+themeAxis+`"/>`+"\n", x0-4, y, x0, y)
		} else {
			fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="`+themeGrid+`" stroke-opacity="0.6"/>`+"\n", x0, y, x1, y)
		}
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="12" fill="`+themeMuted+`" text-anchor="end">%s</text>`+"\n", x0-8, y, fmtTick(v, ystep)+so.Unit)
	}
	// The chosen frontiers, over the grid so a hairline never cuts them.
	if so.Gaps {
		for i := range series {
			if !series[i].Outline {
				continue
			}
			for _, edge := range [][]float64{cum[i], cum[i+1]} {
				fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="%s" stroke-width="1.75" stroke-linejoin="round"/>`+"\n",
					linePath(edge, xAt, yAt), series[i].Color)
			}
		}
	}
	// X ticks and labels.
	if so.XTicks != nil {
		for i, l := range so.XTicks {
			if l == "" || i >= steps {
				continue
			}
			fmt.Fprintf(&b, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%g" stroke="`+themeAxis+`"/>`, xAt(i), y1, xAt(i), y1+4)
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle">%s</text>`+"\n", xAt(i), y1+17, esc(l))
		}
	} else {
		for _, i := range axisTicks(steps) {
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle">%d</text>`+"\n", xAt(i), y1+16, i)
		}
	}
	if so.XLabel != "" {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle">%s</text>`+"\n", (x0+x1)/2, y1+34, esc(so.XLabel))
	}
	if so.YLabel != "" {
		fmt.Fprintf(&b, `<text x="14" y="%.1f" font-size="12" fill="`+themeMuted+`" text-anchor="middle" transform="rotate(-90 14 %.1f)">%s</text>`+"\n", (y0+y1)/2, (y0+y1)/2, esc(so.YLabel))
	}
	// Axes.
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="`+themeAxis+`"/>`+"\n", x0, y1, x1, y1)
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="`+themeAxis+`"/>`+"\n", x0, y0, x0, y1)

	// Reference markers, in the dialect of MultiLine's.
	for _, m := range so.Markers {
		if m.Axis != 'x' || m.Value < 0 || m.Value > xmax {
			continue
		}
		x := x0 + m.Value/xmax*(x1-x0)
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%g" stroke="`+themeInkSoft+`" stroke-dasharray="4 3"/>`+"\n", x, y0, x, y1)
		fmt.Fprintf(&b, `<text x="%.1f" y="%g" font-size="11" fill="`+themeInkSoft+`" text-anchor="middle">%s</text>`+"\n", x, y0-4, esc(m.Label))
	}

	if so.EndLabels && steps > 0 {
		b.WriteString(endLabels(series, cum, steps-1, x1, yAt, so.Unit))
	}

	// Hover payload: the raw (non-cumulative) layer values, top layer first
	// so the tooltip reads like the stack.
	hm := hoverMeta{Kind: "stack", X0: x0, X1: x1, Y0: y0, Y1: y1,
		Xmin: 0, Xmax: xmax, XLabel: so.XLabel, YLabel: so.YLabel, Rows: so.XTips}
	for i := len(series) - 1; i >= 0; i-- {
		hm.Series = append(hm.Series, hoverSeries{Name: series[i].Name, Color: series[i].Color, Ys: series[i].Values})
	}
	b.WriteString(hoverBlock(hm))

	// Legend row.
	lx := x0
	for _, s := range series {
		if s.Name == "" {
			continue
		}
		// Under Gaps the swatch wears the layer's own weight: layers that
		// share a hue and differ by step (an ordinal ramp) stay told apart.
		swatch := 0.75
		if so.Gaps {
			swatch = s.Weight
			if swatch <= 0 {
				swatch = 0.5
			}
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="36" width="12" height="12" rx="2" fill="%s" fill-opacity="%.2f"/>`, lx, s.Color, swatch)
		fmt.Fprintf(&b, `<text x="%.1f" y="46" font-size="12" fill="`+themeInk+`">%s</text>`+"\n", lx+17, esc(s.Name))
		lx += 17 + 7.2*float64(len([]rune(s.Name))) + 18
	}
	b.WriteString("</svg>")
	return finish(b.String())
}

// endLabels writes each layer's name and value at x position last, just right
// of the plot, centred on the layer's own band. Text wears the ink tokens; a
// key square in the layer's colour carries the identity. A band thinner than
// a line of text gets no label, and neither does one that would overlap the
// label below it, walking bottom-up.
func endLabels(series []AreaSeries, cum [][]float64, last int, x1 float64, yAt func(float64) float64, unit string) string {
	const lineH = 13.0
	var b strings.Builder
	prev := math.Inf(1) // y of the label below, in pixels (y grows downward)
	for i, s := range series {
		lo, hi := yAt(cum[i][last]), yAt(cum[i+1][last])
		if lo-hi < lineH-2 {
			continue
		}
		y := (lo + hi) / 2
		if prev-y < lineH {
			continue
		}
		prev = y
		v := 0.0
		if last < len(s.Values) {
			v = s.Values[last]
		}
		op := s.Weight
		if op <= 0 {
			op = 0.5
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="8" height="8" rx="1.5" fill="%s" fill-opacity="%.2f"/>`, x1+8, y-4, s.Color, op)
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" dy="0.35em" font-size="11.5" fill="`+themeInk+`">%s`, x1+21, y, fmtTick(math.Round(v), 1)+unit)
		fmt.Fprintf(&b, ` <tspan fill="`+themeMuted+`">%s</tspan></text>`+"\n", esc(s.Name))
	}
	return b.String()
}
