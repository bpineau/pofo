package chart

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Series is one line on a chart.
type Series struct {
	Name   string
	Dates  []time.Time
	Values []float64
	Color  string // CSS color; picked from a default palette when empty
}

// Options controls the rendering of a chart. The zero Style keeps the
// default look; see Style and StyleMinimal for the available knobs.
type Options struct {
	Title  string
	Width  int // pixels, defaults to 960
	Height int // pixels, defaults to 420
	Style  Style
}

// defaultPalette is the pofo "instrument" series palette: a petrol anchor
// (the brand accent) then rust, indigo, ochre, violet, green, magenta and
// sky. The slot order is deliberate and CVD-validated: it maximises the
// color distance between adjacent series under common color vision
// deficiencies (worst adjacent pair delta-E 15.0), so a two- or three-line
// chart stays readable for everyone; do not permute it casually. Risk
// semantics (green/amber/red) use distinct steps, so a series never
// impersonates a verdict.
var defaultPalette = []string{
	"#0880A8", "#C2452B", "#4C63D2", "#B45309",
	"#6D28D9", "#35803B", "#BE185D", "#2B9BD0",
}

// PaletteColor returns the i-th default series color (hex), cycling; the
// same palette Line falls back to, exported so callers can keep multiple
// charts color-consistent.
func PaletteColor(i int) string { return defaultPalette[i%len(defaultPalette)] }

// maxPlotPoints bounds the number of points drawn per series; longer series
// are decimated (statistics are always computed on full series elsewhere).
const maxPlotPoints = 1600

// Line renders an SVG line chart of the given series.
func Line(opt Options, series []Series) string {
	w, h := opt.Width, opt.Height
	if w == 0 {
		w = 960
	}
	if h == 0 {
		h = 420
	}
	st := opt.Style
	font := st.Font
	if font == "" {
		font = "'Spline Sans Mono', ui-monospace, SF Mono, Menlo, Consolas, monospace"
	}
	fontSize := st.FontSize
	if fontSize == 0 {
		fontSize = 12
	}
	strokeW := st.StrokeWidth
	if strokeW == 0 {
		strokeW = 1.8
	}
	yTicks := st.YTicks
	if yTicks == 0 {
		yTicks = 6
	}
	tickFmt := func(v, step float64) string {
		if st.TickFormat != nil {
			return st.TickFormat(v)
		}
		return fmtTick(v, step)
	}
	legend := len(series) > 1 && !st.HideLegend
	left, right, bottom := 64.0, 14.0, 32.0
	top := 40.0
	if legend {
		top = 64.0 // room for the legend row
	}
	x0, x1 := left, float64(w)-right
	y0, y1 := top, float64(h)-bottom

	// Decimate long series and compute global bounds.
	plot := make([]Series, len(series))
	var tmin, tmax int64 = math.MaxInt64, math.MinInt64
	vmin, vmax := math.Inf(1), math.Inf(-1)
	for i, s := range series {
		d, v := decimate(s.Dates, s.Values)
		color := s.Color
		if color == "" {
			color = defaultPalette[i%len(defaultPalette)]
		}
		plot[i] = Series{Name: s.Name, Dates: d, Values: v, Color: color}
		for j := range d {
			if !isFinite(v[j]) {
				continue
			}
			if u := d[j].Unix(); u < tmin {
				tmin = u
			}
			if u := d[j].Unix(); u > tmax {
				tmax = u
			}
			vmin = math.Min(vmin, v[j])
			vmax = math.Max(vmax, v[j])
		}
	}
	if tmin > tmax { // no drawable point at all
		tmin, tmax = 0, 1
		vmin, vmax = 0, 1
	}
	if tmin == tmax {
		tmax = tmin + 1
	}
	rawMin := vmin
	if vmin == vmax {
		vmin, vmax = vmin-1, vmax+1
	}
	pad := (vmax - vmin) * 0.05
	vmin, vmax = vmin-pad, vmax+pad
	if vmin < 0 && rawMin >= 0 {
		vmin = 0
	}

	xAt := func(t time.Time) float64 {
		return x0 + float64(t.Unix()-tmin)/float64(tmax-tmin)*(x1-x0)
	}
	yAt := func(v float64) float64 {
		return y1 - (v-vmin)/(vmax-vmin)*(y1-y0)
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="%s">`+"\n", w, h, w, h, font)
	switch st.Background {
	case "none":
	case "":
		fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="`+themeSurface+`"/>`+"\n", w, h)
	default:
		fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="%s"/>`+"\n", w, h, st.Background)
	}
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="24" font-size="%d" font-weight="600" fill="`+themeInk+`">%s</text>`+"\n", left, fontSize+4, esc(opt.Title))
	}

	// Horizontal grid and y-axis labels.
	step := niceStep(vmax-vmin, yTicks)
	for v := math.Ceil(vmin/step) * step; v <= vmax+step/1e6; v += step {
		y := yAt(v)
		if !st.HideGrid {
			fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="`+themeGrid+`"/>`+"\n", x0, y, x1, y)
		}
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="%d" fill="`+themeMuted+`" text-anchor="end">%s</text>`+"\n", x0-8, y, fontSize, tickFmt(v, step))
	}
	// X-axis labels (and vertical grid). Use the location of the first series
	// point so intraday charts show exchange-local clock times; daily series
	// use UTC-midnight dates, so this is a no-op for them.
	loc := time.UTC
	if len(series) > 0 && len(series[0].Dates) > 0 {
		loc = series[0].Dates[0].Location()
	}
	tfrom, tto := time.Unix(tmin, 0).In(loc), time.Unix(tmax, 0).In(loc)
	if st.CornerDates {
		layout := "2006-01-02"
		if d := tto.Sub(tfrom); d > 0 && d <= 36*time.Hour {
			layout = "15:04"
		}
		fmt.Fprintf(&b, `<text x="%g" y="%g" font-size="%d" fill="`+themeMuted+`">%s</text>`+"\n", x0, y1+18, fontSize, tfrom.Format(layout))
		fmt.Fprintf(&b, `<text x="%g" y="%g" font-size="%d" fill="`+themeMuted+`" text-anchor="end">%s</text>`+"\n", x1, y1+18, fontSize, tto.Format(layout))
	} else {
		for _, tk := range timeTicks(tfrom, tto) {
			x := xAt(tk.t)
			if !st.HideGrid {
				fmt.Fprintf(&b, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%g" stroke="`+themeGrid+`"/>`+"\n", x, y0, x, y1)
			}
			fmt.Fprintf(&b, `<text x="%.1f" y="%g" font-size="%d" fill="`+themeMuted+`" text-anchor="middle">%s</text>`+"\n", x, y1+18, fontSize, esc(tk.label))
		}
	}
	// Axes.
	if !st.HideAxes {
		fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="`+themeAxis+`"/>`+"\n", x0, y1, x1, y1)
		fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="`+themeAxis+`"/>`+"\n", x0, y0, x0, y1)
	}

	// Area fill under the first series (finite stretches only).
	if st.Fill && len(plot) > 0 {
		for _, s := range plot[:1] {
			var pts strings.Builder
			first, last := -1, -1
			for j := range s.Dates {
				if !isFinite(s.Values[j]) {
					continue
				}
				if first < 0 {
					first = j
				}
				last = j
				fmt.Fprintf(&pts, "%.1f,%.1f ", xAt(s.Dates[j]), yAt(s.Values[j]))
			}
			if first < 0 {
				continue
			}
			fmt.Fprintf(&b, `<polygon points="%s%.1f,%g %.1f,%g" fill="%s" fill-opacity="0.07"/>`+"\n",
				pts.String(), xAt(s.Dates[last]), y1, xAt(s.Dates[first]), y1, s.Color)
		}
	}

	// Series lines.
	for _, s := range plot {
		var p strings.Builder
		pen := false
		for j := range s.Dates {
			if !isFinite(s.Values[j]) {
				pen = false
				continue
			}
			cmd := "L"
			if !pen {
				cmd, pen = "M", true
			}
			fmt.Fprintf(&p, "%s%.1f %.1f", cmd, xAt(s.Dates[j]), yAt(s.Values[j]))
		}
		if p.Len() == 0 {
			continue
		}
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="%s" stroke-width="%g" stroke-linejoin="round" stroke-linecap="round"/>`+"\n", p.String(), s.Color, strokeW)
	}

	// Legend (only useful with several series).
	if legend {
		x := left
		for _, s := range plot {
			fmt.Fprintf(&b, `<rect x="%.1f" y="36" width="12" height="12" rx="2" fill="%s"/>`, x, s.Color)
			fmt.Fprintf(&b, `<text x="%.1f" y="46" font-size="%d" fill="`+themeInk+`">%s</text>`+"\n", x+17, fontSize, esc(s.Name))
			x += 17 + 7.2*float64(len([]rune(s.Name))) + 18
		}
	}
	b.WriteString("</svg>")
	return finish(b.String())
}

type tick struct {
	t     time.Time
	label string
}

// timeTicks picks round year boundaries producing at most ~10 labels, falling
// back to evenly spaced month labels for short spans.
func timeTicks(from, to time.Time) []tick {
	// Sub-day spans (intraday) get clock-time labels.
	if d := to.Sub(from); d > 0 && d <= 36*time.Hour {
		const n = 5
		out := make([]tick, 0, n+1)
		for i := 0; i <= n; i++ {
			t := from.Add(time.Duration(i) * d / n)
			out = append(out, tick{t, t.Format("15:04")})
		}
		return out
	}
	years := to.Year() - from.Year()
	for _, step := range []int{1, 2, 5, 10, 20} {
		if years/step > 10 {
			continue
		}
		var out []tick
		first := from.Year()
		if !(from.Month() == time.January && from.Day() == 1) {
			first++
		}
		for first%step != 0 {
			first++
		}
		for y := first; ; y += step {
			t := time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
			if t.After(to) {
				break
			}
			out = append(out, tick{t, fmt.Sprintf("%d", y)})
		}
		if len(out) >= 2 {
			return out
		}
	}
	const n = 5
	out := make([]tick, 0, n+1)
	for i := 0; i <= n; i++ {
		t := from.Add(time.Duration(i) * to.Sub(from) / n)
		out = append(out, tick{t, t.Format("2006-01")})
	}
	return out
}

// niceStep returns a 1/2/5 × 10^k step so that span/step stays under maxTicks.
func niceStep(span float64, maxTicks int) float64 {
	if span <= 0 {
		return 1
	}
	raw := span / float64(maxTicks)
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	for _, m := range []float64{1, 2, 5, 10} {
		if mag*m >= raw {
			return mag * m
		}
	}
	return mag * 10
}

func fmtTick(v, step float64) string {
	if math.Abs(v) < step/1e6 {
		v = 0
	}
	// Compact large magnitudes (e.g. wealth axes in raw euros) so ticks read
	// "15M" / "500k" instead of "15000000". The 100k gate keeps already-scaled
	// axes (k€, percentages, small counts) in their plain form.
	if a := math.Abs(v); a >= 1e6 {
		return fmt.Sprintf("%gM", trim(v/1e6))
	} else if a >= 1e5 {
		return fmt.Sprintf("%gk", trim(v/1e3))
	}
	switch {
	case step >= 1:
		return fmt.Sprintf("%.0f", v)
	case step >= 0.1:
		return fmt.Sprintf("%.1f", v)
	default:
		return fmt.Sprintf("%.2f", v)
	}
}

// trim rounds to two significant decimals so compacted ticks stay short.
func trim(v float64) float64 { return math.Round(v*100) / 100 }

// decimate stride-samples long series down to maxPlotPoints, keeping the
// first and last points.
func decimate(dates []time.Time, values []float64) ([]time.Time, []float64) {
	n := len(dates)
	if n <= maxPlotPoints {
		return dates, values
	}
	stride := (n + maxPlotPoints - 1) / maxPlotPoints
	dd := make([]time.Time, 0, maxPlotPoints+1)
	vv := make([]float64, 0, maxPlotPoints+1)
	for i := 0; i < n; i += stride {
		dd = append(dd, dates[i])
		vv = append(vv, values[i])
	}
	if !dd[len(dd)-1].Equal(dates[n-1]) {
		dd = append(dd, dates[n-1])
		vv = append(vv, values[n-1])
	}
	return dd, vv
}

func isFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func esc(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
