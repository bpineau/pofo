package chart

import (
	"math"
	"strconv"
	"strings"
)

// Style adjusts the rendering of a chart beyond its dimensions. The zero
// value reproduces the default pofo look (white background, grid, axes,
// legend on multi-series charts), so existing callers are unaffected;
// each field opts one aspect out or overrides one default. StyleMinimal
// bundles the knobs into a bare, embeddable dialect.
type Style struct {
	Background  string               // CSS color of the background rect; "none" draws no rect
	Font        string               // font-family of every label
	FontSize    int                  // label font size in px (titles render 4px larger)
	HideGrid    bool                 // no horizontal/vertical grid lines (labels remain)
	HideAxes    bool                 // no axis strokes
	HideLegend  bool                 // no legend row on multi-series charts
	Fill        bool                 // translucent area fill under the first series
	StrokeWidth float64              // series stroke width in px
	YTicks      int                  // number of y-label steps (approximate; default 6)
	CornerDates bool                 // label only the first and last date, at the corners
	TickFormat  func(float64) string // y-axis label formatter; see Compact
}

// StyleMinimal returns the bare chart dialect: transparent background,
// monospace labels, no grid, no axes, an area fill under the first series,
// four compact y labels and the date range at the bottom corners. It suits
// dense pages that embed many small charts (the finador web UI look).
func StyleMinimal() Style {
	return Style{
		Background:  "none",
		Font:        "ui-monospace,monospace",
		FontSize:    11,
		HideGrid:    true,
		HideAxes:    true,
		Fill:        true,
		YTicks:      4,
		CornerDates: true,
		TickFormat:  Compact,
	}
}

// Compact shortens a number for a tight axis label: 1.23M, 473.9k, 12.35.
func Compact(v float64) string {
	a := math.Abs(v)
	switch {
	case a >= 1e6:
		return trimZeros(strconv.FormatFloat(v/1e6, 'f', 2, 64)) + "M"
	case a >= 1e3:
		return trimZeros(strconv.FormatFloat(v/1e3, 'f', 1, 64)) + "k"
	default:
		return trimZeros(strconv.FormatFloat(v, 'f', 2, 64))
	}
}

func trimZeros(s string) string {
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}
