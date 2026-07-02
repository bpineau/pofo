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
	Title   string
	Width   int  // total width in columns, gutter included; default 100
	Height  int  // plot height in rows; default 18
	Color   bool // ANSI colors; without them each series gets its own marker
	Braille bool // pack 2x4 braille dots per cell for a smoother curve
}

// ansiPalette mirrors defaultPalette with ANSI-256 codes.
var ansiPalette = []int{33, 208, 40, 160, 134, 130, 170, 44}

// plainMarkers distinguishes series when colors are disabled.
var plainMarkers = []rune{'•', '+', '×', 'o', '#', '@', '*', '%'}

// Term renders series as a line chart for the terminal: one column per time
// step, colored (or distinctly marked) per series, with value and year axes.
// With Braille, each cell packs 2x4 dots (U+2800 block) for a smoother
// curve; overlapping series are told apart by color when Color is set.
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
		return "(not enough data to plot)\n"
	}

	// Dot resolution: one dot per cell, or 2x4 braille dots per cell.
	dotsX, dotsY := 1, 1
	if opt.Braille {
		dotsX, dotsY = 2, 4
	}
	gridW, gridH := plotW*dotsX, height*dotsY

	// Plot grid: -1 empty, otherwise the series index (later series win).
	grid := make([][]int8, gridH)
	for r := range grid {
		grid[r] = make([]int8, gridW)
		for c := range grid[r] {
			grid[r][c] = -1
		}
	}
	rowFor := func(v float64) int {
		f := (v - vmin) / (vmax - vmin)
		return gridH - 1 - int(math.Round(f*float64(gridH-1)))
	}
	for si, s := range series {
		prev := -1
		for x := range gridW {
			t := tmin + int64(float64(tmax-tmin)*float64(x)/float64(gridW-1))
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

	paint := func(si int8, glyph rune) string {
		if opt.Color {
			return fmt.Sprintf("\x1b[38;5;%dm%c\x1b[0m", ansiPalette[int(si)%len(ansiPalette)], glyph)
		}
		return string(glyph)
	}
	mark := func(si int8) string {
		if opt.Braille {
			return paint(si, '⣿') // legend swatch in braille mode
		}
		if !opt.Color {
			return string(plainMarkers[int(si)%len(plainMarkers)])
		}
		return paint(si, '•')
	}
	// cell renders the character at plot position (row, col): the marker of
	// the owning series, or in braille mode the composition of its 2x4 dots
	// colored by the series owning most of them.
	cell := func(row, col int) (string, bool) {
		if !opt.Braille {
			if si := grid[row][col]; si >= 0 {
				return mark(si), true
			}
			return "", false
		}
		var bits rune
		counts := [8]int{}
		best, bestN := int8(-1), 0
		for dy := range dotsY {
			for dx := range dotsX {
				si := grid[row*dotsY+dy][col*dotsX+dx]
				if si < 0 {
					continue
				}
				bits |= brailleBit(dx, dy)
				counts[int(si)%len(counts)]++
				if n := counts[int(si)%len(counts)]; n > bestN {
					best, bestN = si, n
				}
			}
		}
		if bits == 0 {
			return "", false
		}
		return paint(best, 0x2800+bits), true
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
			if s, ok := cell(r, c); ok {
				b.WriteString(s)
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

// brailleBit maps a (dx, dy) dot to its bit in the U+2800 braille block.
func brailleBit(dx, dy int) rune {
	bits := [4][2]rune{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}
	return bits[dy][dx]
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
