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

// Options controls the rendering of a chart.
type Options struct {
	Title  string
	Width  int // pixels, defaults to 960
	Height int // pixels, defaults to 420
}

// defaultPalette is the pofo "risk desk" series palette: a deep petrol anchor,
// a burnt-amber counterpoint, then supporting instrument hues, all legible on
// the cool-paper background and distinct from the matplotlib default.
var defaultPalette = []string{
	"#0F766E", "#B4531F", "#3A5A8C", "#2E7D5B",
	"#6D5A9C", "#9C6B3F", "#B0476B", "#227C9D",
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
	multi := len(series) > 1
	left, right, bottom := 64.0, 14.0, 32.0
	top := 40.0
	if multi {
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
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" font-family="-apple-system, Segoe UI, Helvetica, Arial, sans-serif">`+"\n", w, h, w, h)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#ffffff"/>`+"\n", w, h)
	if opt.Title != "" {
		fmt.Fprintf(&b, `<text x="%g" y="24" font-size="16" font-weight="600" fill="#14232B">%s</text>`+"\n", left, esc(opt.Title))
	}

	// Horizontal grid and y-axis labels.
	step := niceStep(vmax-vmin, 6)
	for v := math.Ceil(vmin/step) * step; v <= vmax+step/1e6; v += step {
		y := yAt(v)
		fmt.Fprintf(&b, `<line x1="%g" y1="%.1f" x2="%g" y2="%.1f" stroke="#E8ECEA"/>`+"\n", x0, y, x1, y)
		fmt.Fprintf(&b, `<text x="%g" y="%.1f" dy="0.35em" font-size="12" fill="#55666E" text-anchor="end">%s</text>`+"\n", x0-8, y, fmtTick(v, step))
	}
	// Vertical grid and x-axis labels.
	// Use the location of the first series point so intraday charts show
	// exchange-local clock times; daily series use UTC-midnight dates, so
	// this is a no-op for them.
	loc := time.UTC
	if len(series) > 0 && len(series[0].Dates) > 0 {
		loc = series[0].Dates[0].Location()
	}
	for _, tk := range timeTicks(time.Unix(tmin, 0).In(loc), time.Unix(tmax, 0).In(loc)) {
		x := xAt(tk.t)
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%g" x2="%.1f" y2="%g" stroke="#E8ECEA"/>`+"\n", x, y0, x, y1)
		fmt.Fprintf(&b, `<text x="%.1f" y="%g" font-size="12" fill="#55666E" text-anchor="middle">%s</text>`+"\n", x, y1+18, esc(tk.label))
	}
	// Axes.
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="#AEB9B8"/>`+"\n", x0, y1, x1, y1)
	fmt.Fprintf(&b, `<line x1="%g" y1="%g" x2="%g" y2="%g" stroke="#AEB9B8"/>`+"\n", x0, y0, x0, y1)

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
		fmt.Fprintf(&b, `<path d="%s" fill="none" stroke="%s" stroke-width="1.8" stroke-linejoin="round" stroke-linecap="round"/>`+"\n", p.String(), s.Color)
	}

	// Legend (only useful with several series).
	if multi {
		x := left
		for _, s := range plot {
			fmt.Fprintf(&b, `<rect x="%.1f" y="36" width="12" height="12" rx="2" fill="%s"/>`, x, s.Color)
			fmt.Fprintf(&b, `<text x="%.1f" y="46" font-size="12" fill="#14232B">%s</text>`+"\n", x+17, esc(s.Name))
			x += 17 + 7.2*float64(len([]rune(s.Name))) + 18
		}
	}
	b.WriteString("</svg>")
	return b.String()
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
