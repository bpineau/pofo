package chart

import (
	"fmt"
	"math"
	"strings"
)

// DivergingStackSeries is one signed layer of a DivergingStack chart: at each
// x position its value stacks above zero when positive and below when
// negative (e.g. an asset's contribution to a portfolio's return).
type DivergingStackSeries struct {
	Name   string
	Color  string    // optional; the default palette applies in series order
	Values []float64 // one signed value per x position
}

// StripBand is one contiguous run of the categorical annotation strip drawn
// above a DivergingStack plot (e.g. a macro-regime timeline). From and To are
// x positions, inclusive.
type StripBand struct {
	From, To int
	Label    string
	Color    string
}

// DivergingStackOptions styles a DivergingStack. The zero value renders a
// bare 1200x470 chart; XLabels, Total and Strip are optional layers.
type DivergingStackOptions struct {
	Title  string
	Width  int // default 1200
	Height int // default 470

	XLabels []string // per-position tick label; "" = no tick at that position
	XTips   []string // per-position hover header (e.g. "2014-06"); optional
	XLabel  string   // x axis name for the hover layer (e.g. "month")
	YLabel  string   // y unit (e.g. "pts"), shown in the hover header

	Total     []float64 // optional net line overlaid on the stack
	TotalName string    // its legend/hover label (default "total")

	Strip       []StripBand // optional categorical strip above the plot
	StripName   string      // small heading over the strip legend
	StripLegend []Slice     // strip legend entries (Label and Color; Value ignored)
}

// DivergingStack renders signed series as stacked areas around a zero axis:
// positive values stack upward, negative downward, so at every x the band
// heights above and below zero decompose the total. Series order is the
// stacking order (first series sits against the axis). An optional Total
// line overlays the net sum and an optional categorical strip annotates the
// timeline. The chart embeds hover metadata (kind "stack") for the
// crosshair-tooltip front-end layer. Returns "" when no series has a value.
func DivergingStack(opt DivergingStackOptions, series []DivergingStackSeries) string {
	n := 0
	for _, s := range series {
		if len(s.Values) > n {
			n = len(s.Values)
		}
	}
	if n < 2 {
		return ""
	}
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 1200
	}
	if h == 0 {
		h = 470
	}
	const padL, padR, padB = 56, 24, 36
	padT := 64
	if len(opt.Strip) > 0 {
		padT = 88
	}
	plotW, plotH := float64(w-padL-padR), float64(h-padT-padB)

	// Envelopes give the y range; a lone flat stack still needs a span.
	up, dn := 0.0, 0.0
	for m := range n {
		pu, pd := 0.0, 0.0
		for _, s := range series {
			if m < len(s.Values) {
				if v := s.Values[m]; v > 0 {
					pu += v
				} else {
					pd += v
				}
			}
		}
		up, dn = math.Max(up, pu), math.Min(dn, pd)
	}
	for _, t := range opt.Total {
		up, dn = math.Max(up, t), math.Min(dn, t)
	}
	if up == dn {
		up, dn = 1, -1
	}
	step := niceStep(up-dn, 7)
	up, dn = math.Ceil(up/step)*step, math.Floor(dn/step)*step

	x := func(m int) float64 { return float64(padL) + plotW*float64(m)/float64(n-1) }
	y := func(v float64) float64 { return float64(padT) + plotH*(up-v)/(up-dn) }

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="`+themeMono+`">`, w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="`+themeSurface+`"/>`, w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%d" y="20" font-size="13" font-weight="600" fill="`+themeInk+`">%s</text>`, padL, esc(opt.Title))
	}

	// Series legend row under the title.
	lx := float64(padL)
	for i, s := range series {
		color := s.Color
		if color == "" {
			color = PaletteColor(i)
		}
		fmt.Fprintf(&b, `<rect x="%.1f" y="34" width="10" height="10" rx="2" fill="%s"/><text x="%.1f" y="43" font-size="11" fill="`+themeInkSoft+`">%s</text>`,
			lx, color, lx+14, esc(s.Name))
		lx += 14 + float64(len(s.Name))*7 + 18
	}
	if len(opt.Total) > 0 {
		name := opt.TotalName
		if name == "" {
			name = "total"
		}
		fmt.Fprintf(&b, `<line x1="%.1f" y1="39" x2="%.1f" y2="39" stroke="`+themeInk+`" stroke-width="1.6"/><text x="%.1f" y="43" font-size="11" fill="`+themeInkSoft+`">%s</text>`,
			lx, lx+16, lx+20, esc(name))
	}

	// Categorical strip hugging the plot top, with its own legend row.
	if len(opt.Strip) > 0 {
		stripY := float64(padT) - 18
		colW := plotW / float64(n-1)
		for _, band := range opt.Strip {
			if band.To < band.From {
				continue
			}
			x0 := x(band.From)
			x1 := x(band.To) + colW
			if band.To >= n-1 {
				x1 = x(n - 1)
			}
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="10" fill="%s" fill-opacity="0.75" data-tip="%s"/>`,
				x0, stripY, math.Max(x1-x0, 1), band.Color, esc(band.Label))
		}
		if len(opt.StripLegend) > 0 {
			lx := float64(padL)
			if opt.StripName != "" {
				fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" font-size="10" fill="`+themeMuted+`">%s</text>`, lx, stripY-4, esc(opt.StripName))
				lx += float64(len(opt.StripName))*6 + 12
			}
			for _, e := range opt.StripLegend {
				fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="9" height="9" rx="2" fill="%s" fill-opacity="0.75"/><text x="%.1f" y="%.1f" font-size="10" fill="`+themeInkSoft+`">%s</text>`,
					lx, stripY-13, e.Color, lx+13, stripY-5, esc(e.Label))
				lx += 13 + float64(len(e.Label))*6 + 14
			}
		}
	}

	// Horizontal gridlines and y labels; the zero axis is emphasized.
	for v := dn; v <= up+step/2; v += step {
		col, sw := themeGrid, 1.0
		if math.Abs(v) < step/2 {
			col, sw = themeAxis, 1.2
		}
		fmt.Fprintf(&b, `<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="%s" stroke-width="%.1f"/>`, padL, y(v), w-padR, y(v), col, sw)
		fmt.Fprintf(&b, `<text x="%d" y="%.1f" text-anchor="end" font-size="10" fill="`+themeMuted+`">%+.4g</text>`, padL-6, y(v)+3, v)
	}
	// X ticks from the sparse labels.
	for m := 0; m < n && m < len(opt.XLabels); m++ {
		if opt.XLabels[m] == "" {
			continue
		}
		fmt.Fprintf(&b, `<text x="%.1f" y="%d" text-anchor="middle" font-size="10" fill="`+themeMuted+`">%s</text>`, x(m), h-padB+16, esc(opt.XLabels[m]))
	}

	// Stacked areas: one positive and one negative polygon per series, so a
	// sign flip never produces a self-crossing shape. Hairline separators in
	// the surface color keep adjacent fills distinct.
	upEnv := make([]float64, n)
	dnEnv := make([]float64, n)
	val := func(s DivergingStackSeries, m int) float64 {
		if m < len(s.Values) {
			return s.Values[m]
		}
		return 0
	}
	for i, s := range series {
		color := s.Color
		if color == "" {
			color = PaletteColor(i)
		}
		var posT, posB, negT, negB []string
		for m := range n {
			v := val(s, m)
			pv, nv := math.Max(v, 0), math.Min(v, 0)
			posT = append(posT, fmt.Sprintf("%.1f,%.1f", x(m), y(upEnv[m]+pv)))
			posB = append(posB, fmt.Sprintf("%.1f,%.1f", x(m), y(upEnv[m])))
			negT = append(negT, fmt.Sprintf("%.1f,%.1f", x(m), y(dnEnv[m])))
			negB = append(negB, fmt.Sprintf("%.1f,%.1f", x(m), y(dnEnv[m]+nv)))
			upEnv[m] += pv
			dnEnv[m] += nv
		}
		reverse(posB)
		reverse(negB)
		for _, poly := range [][]string{append(posT, posB...), append(negT, negB...)} {
			fmt.Fprintf(&b, `<polygon points="%s" fill="%s" fill-opacity="0.92" stroke="`+themeSurface+`" stroke-width="0.6"/>`,
				strings.Join(poly, " "), color)
		}
	}
	if len(opt.Total) > 0 {
		var pl []string
		for m := 0; m < n && m < len(opt.Total); m++ {
			pl = append(pl, fmt.Sprintf("%.1f,%.1f", x(m), y(opt.Total[m])))
		}
		fmt.Fprintf(&b, `<polyline points="%s" fill="none" stroke="`+themeInk+`" stroke-width="1.6"/>`, strings.Join(pl, " "))
	}

	// Hover payload: indexed x, one row per series plus the total.
	hm := hoverMeta{
		Kind: "stack", X0: float64(padL), X1: float64(w - padR),
		Y0: float64(padT), Y1: float64(h - padB),
		Xmin: 0, Xmax: float64(n - 1), XLabel: opt.XLabel, YLabel: opt.YLabel,
		Rows: opt.XTips,
	}
	for i, s := range series {
		color := s.Color
		if color == "" {
			color = PaletteColor(i)
		}
		hm.Series = append(hm.Series, hoverSeries{Name: s.Name, Color: color, Ys: s.Values})
	}
	if len(opt.Total) > 0 {
		name := opt.TotalName
		if name == "" {
			name = "total"
		}
		hm.Series = append(hm.Series, hoverSeries{Name: name, Color: themeInk, Ys: opt.Total})
	}
	b.WriteString(hoverBlock(hm))
	b.WriteString(`</svg>`)
	return finish(b.String())
}

func reverse(p []string) {
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
}
