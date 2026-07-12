package chart

import (
	"encoding/json"
	"fmt"
)

// Hover metadata: line/fan/stacked charts embed a machine-readable copy of
// their data as an SVG <metadata> element, so a thin front-end layer can add
// a crosshair-plus-tooltip without re-deriving anything from the drawn
// geometry. The payload is invisible (metadata does not render), additive
// (consumers that ignore it see the same chart) and self-contained: it
// carries the plot box and x domain needed to map pointer positions back to
// data. Discrete-mark charts (Bars, Scatter, CategoryBars) do not use it;
// they carry a native <title> per mark instead.
type hoverMeta struct {
	// Kind: "line" (explicit xs), "fan"/"stack" (x = the 0..n-1 index), or
	// "bars"/"cat"/"scatter" (discrete marks: no crosshair, native titles
	// carry the hover; the payload only feeds the table view).
	Kind   string        `json:"kind"`
	X0     float64       `json:"x0,omitempty"` // plot box, viewBox pixels (continuous kinds)
	X1     float64       `json:"x1,omitempty"`
	Y0     float64       `json:"y0,omitempty"`
	Y1     float64       `json:"y1,omitempty"`
	Xmin   float64       `json:"xmin,omitempty"` // x domain mapped onto [X0, X1]
	Xmax   float64       `json:"xmax,omitempty"`
	XLabel string        `json:"xlabel,omitempty"`
	YLabel string        `json:"ylabel,omitempty"`
	Rows   []string      `json:"rows,omitempty"` // row labels of discrete kinds
	Series []hoverSeries `json:"series"`
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
