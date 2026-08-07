package chart

// Chart theme: the single source of truth for the "chrome" colors and fonts
// every chart in this package draws with (grids, axes, labels, backgrounds and
// the semantic good/warn/bad/accent hues). Series colors are separate, see
// defaultPalette.
//
// THIS IS THE RESKIN SURFACE. To restyle every chart at once, change the values
// here and re-run `make golden` (the frozen SVG fixtures will move; that is
// expected for a deliberate reskin). The values mirror pkg/webui/theme.css so
// the standalone report SVGs and the embedded web UI share one look; keep them
// in sync. Charts emit these as literal hex (the standalone report has no CSS
// variables), so a token must be a compile-time constant, not a CSS var().
//
// A guard test (theme_test.go) fails if a chart source hardcodes a chrome color
// outside this registry, so new charts are pushed to reuse the tokens rather
// than reintroduce scattered hex.
const (
	themeInk     = "#16181D" // primary ink: median/needle strokes, value text
	themeInkSoft = "#4A5160" // secondary text
	themeMuted   = "#7A8294" // axis labels, captions
	themeFaint   = "#A8AEBC" // context / sample lines
	themeGrid    = "#EDF0F3" // hairline gridlines
	themeWell    = "#EEF0F3" // inset/well fill (gauge track)
	themeAxis    = "#CDD2DA" // axis strokes, frontier connector
	themeSurface = "#FFFFFF" // chart background rect

	// Semantic risk hues (kept distinct from the series palette accent).
	themeAccent = "#0B7285" // petrol brand accent
	themeGood   = "#0C8A47" // safe
	themeWarn   = "#C77E17" // caution
	themeBad    = "#D2402F" // danger / ruin
	themeDead   = "#9AA2B1" // neutral grey ("gone" in the lifecycle)
)

// Fonts. The mono face carries every numeric/axis label; the sans face titles
// and the newer instrument-style charts. They mirror the webui theme's stacks.
const (
	themeMono = "'Spline Sans Mono', ui-monospace, SF Mono, Menlo, Consolas, monospace"
	themeSans = "'Instrument Sans',system-ui,sans-serif"
)
