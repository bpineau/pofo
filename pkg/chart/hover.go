package chart

import (
	"encoding/json"
	"fmt"
)

// Hover metadata: every chart embeds a machine-readable copy of its data as
// an SVG <metadata> element, so a thin front-end layer can add a
// crosshair-plus-tooltip without re-deriving anything from the drawn
// geometry. The payload is invisible (metadata does not render), additive
// (consumers that ignore it see the same chart) and self-contained: it
// carries the plot box, and either the x domain (continuous kinds) or the
// pixel anchor of every mark (discrete kinds), which is what lets a pointer
// ANYWHERE in the plot box be answered rather than only over a drawn mark.
//
// The discrete kinds also carry a native <title> per mark, for consumers with
// no hover layer of their own (the standalone report).
type hoverMeta struct {
	// Kind: "line" (explicit xs), "fan"/"stack" (x = the 0..n-1 index), or
	// "bars"/"cat"/"scatter" (discrete marks, located by Marks).
	Kind   string        `json:"kind"`
	X0     float64       `json:"x0,omitempty"` // plot box, viewBox pixels
	X1     float64       `json:"x1,omitempty"`
	Y0     float64       `json:"y0,omitempty"`
	Y1     float64       `json:"y1,omitempty"`
	Xmin   float64       `json:"xmin"` // x domain mapped onto [X0, X1]; 0 is common (indexed kinds), so never omitted
	Xmax   float64       `json:"xmax,omitempty"`
	XLabel string        `json:"xlabel,omitempty"`
	YLabel string        `json:"ylabel,omitempty"`
	Rows   []string      `json:"rows,omitempty"` // row labels of discrete kinds; per-x hover headers for "stack"
	Series []hoverSeries `json:"series"`
	// Axis says how a discrete chart's marks are laid out, so a pointer
	// anywhere in the plot box maps to one: "x" (columns, nearest in x), "y"
	// (rows, nearest in y) or "xy" (free points, nearest in both). Empty on
	// the continuous kinds, which snap along x through Xmin/Xmax.
	Axis string `json:"axis,omitempty"`
	// Marks is the pixel anchor of each discrete mark, index for index with
	// Rows and with every series' Ys.
	Marks []hoverMark `json:"marks,omitempty"`
}

// hoverMark is one discrete mark's anchor in viewBox pixels: the centre of a
// bar's band, or a scatter point. The front-end finds the mark nearest the
// pointer along Axis, so it needs no knowledge of the chart's geometry.
//
// Text is the mark's own value label when the chart draws one ("7.8%",
// "-13.8 pp"): the caller has already formatted it WITH its unit, which a
// tooltip reading a bare number cannot recover.
type hoverMark struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Text string  `json:"t,omitempty"`
}

// hoverSeries is one tooltip row source: a named series with its values, and
// its x positions when they are not the plain 0..n-1 index.
type hoverSeries struct {
	Name  string    `json:"name"`
	Color string    `json:"color,omitempty"`
	Xs    []float64 `json:"xs,omitempty"`
	Ys    []float64 `json:"ys"`
}

// bandFillColor is the shade of every fan percentile band (the accent hue at
// low opacity); the hover payload reuses it as the bands' tooltip key.
const bandFillColor = themeAccent

// hoverBlock renders the payload as an SVG metadata element. json.Marshal
// HTML-escapes <, > and & by default, so the JSON is XML-safe as-is.
func hoverBlock(m hoverMeta) string {
	j, err := json.Marshal(m)
	if err != nil {
		return "" // a chart without hover data is still a valid chart
	}
	return fmt.Sprintf(`<metadata class="hover">%s</metadata>`, j) + "\n"
}

// fanBandNames labels percentile bands for the tooltip: the friendly names
// for the canonical 5- and 3-band fans, positional names otherwise.
func fanBandNames(n int) []string {
	switch n {
	case 5:
		return []string{"p5", "p25", "median", "p75", "p95"}
	case 3:
		return []string{"p5", "median", "p95"}
	}
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("band %d", i+1)
	}
	return out
}
