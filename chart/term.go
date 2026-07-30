package chart

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// TermOptions controls the terminal rendering of Term.
type TermOptions struct {
	Title  string
	Width  int  // total width in columns, gutter included; default 100
	Height int  // plot height in rows; default 18
	Color  bool // ANSI colors; without them each series gets its own marker
}

// ansiPalette mirrors defaultPalette with ANSI-256 codes.
var ansiPalette = []int{33, 208, 40, 160, 134, 130, 170, 44}

// plainMarkers distinguishes series when colors are disabled.
var plainMarkers = []rune{'•', '+', '×', 'o', '#', '@', '*', '%'}

// Term renders series as a line chart for the terminal: one column per time
// step, colored (or distinctly marked) per series, with value and year axes.
func Term(opt TermOptions, series []Series) string {
	width := opt.Width
	if width <= 0 {
		width = 100
	}
	height := opt.Height
	if height <= 0 {
		height = 18
	}
	const gutter = 10 // "  123456 ┤"
	plotW := max(width-gutter, 20)

	// Bounds.
	var tmin, tmax int64 = math.MaxInt64, math.MinInt64
	vmin, vmax := math.Inf(1), math.Inf(-1)
	for _, s := range series {
		for i, d := range s.Dates {
			if !isFinite(s.Values[i]) {
				continue
			}
			u := d.Unix()
			tmin, tmax = min(tmin, u), max(tmax, u)
			vmin, vmax = math.Min(vmin, s.Values[i]), math.Max(vmax, s.Values[i])
		}
	}
	if tmin >= tmax || vmin >= vmax {
		return "(pas assez de données à tracer)\n"
	}

	// Plot grid: -1 empty, otherwise the series index (later series win).
	grid := make([][]int8, height)
	for r := range grid {
		grid[r] = make([]int8, plotW)
		for c := range grid[r] {
			grid[r][c] = -1
		}
	}
	rowFor := func(v float64) int {
		f := (v - vmin) / (vmax - vmin)
		return height - 1 - int(math.Round(f*float64(height-1)))
	}
	for si, s := range series {
		prev := -1
		for x := range plotW {
			t := tmin + int64(float64(tmax-tmin)*float64(x)/float64(plotW-1))
			v, ok := valueAt(s, t)
			if !ok {
				prev = -1
				continue
			}
			r := rowFor(v)
			grid[r][x] = int8(si)
			// Connect steep moves vertically for visual continuity.
			if prev >= 0 {
				lo, hi := min(prev, r), max(prev, r)
				for rr := lo + 1; rr < hi; rr++ {
					grid[rr][x] = int8(si)
				}
			}
			prev = r
		}
	}

	mark := func(si int8) string {
		if opt.Color {
			return fmt.Sprintf("\x1b[38;5;%dm•\x1b[0m", ansiPalette[int(si)%len(ansiPalette)])
		}
		return string(plainMarkers[int(si)%len(plainMarkers)])
	}

	var b strings.Builder
	if opt.Title != "" {
		fmt.Fprintf(&b, "%s\n", opt.Title)
	}
	if len(series) > 1 {
		for si, s := range series {
			if si > 0 {
				b.WriteString("   ")
			}
			fmt.Fprintf(&b, "%s %s", mark(int8(si)), s.Name)
		}
		b.WriteString("\n")
	}
	labelEvery := max(height/4, 1)
	for r := range height {
		if r%labelEvery == 0 || r == height-1 {
			v := vmax - (vmax-vmin)*float64(r)/float64(height-1)
			fmt.Fprintf(&b, "%8s ┤", fmtTick(v, (vmax-vmin)/6))
		} else {
			fmt.Fprintf(&b, "%8s │", "")
		}
		for c := range plotW {
			if si := grid[r][c]; si >= 0 {
				b.WriteString(mark(si))
			} else {
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "%8s └%s\n", "", strings.Repeat("─", plotW))

	// Time labels: five marks across the axis.
	labels := make([]string, 0, 5)
	span := time.Unix(tmax, 0).Sub(time.Unix(tmin, 0))
	for i := range 5 {
		t := time.Unix(tmin+int64(span.Seconds()*float64(i)/4), 0).UTC()
		if span > 3*365*24*time.Hour {
			labels = append(labels, t.Format("2006"))
		} else {
			labels = append(labels, t.Format("2006-01"))
		}
	}
	axis := make([]byte, plotW+gutter)
	for i := range axis {
		axis[i] = ' '
	}
	for i, l := range labels {
		pos := gutter + (plotW-1)*i/4 - len(l)/2
		pos = max(min(pos, plotW+gutter-len(l)), 0)
		copy(axis[pos:], l)
	}
	b.WriteString(strings.TrimRight(string(axis), " "))
	b.WriteByte('\n')
	return b.String()
}

// valueAt returns the series value at the last date ≤ t (unix seconds).
func valueAt(s Series, t int64) (float64, bool) {
	i := sort.Search(len(s.Dates), func(k int) bool { return s.Dates[k].Unix() > t })
	if i == 0 {
		return 0, false
	}
	v := s.Values[i-1]
	return v, isFinite(v)
}
