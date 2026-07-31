package chart

import "strings"

// Chart theme: the single source of truth for the "chrome" colors and fonts
// every chart draws with (grids, axes, labels, backgrounds and the semantic
// good/warn/bad/accent hues). Series colors are separate, see defaultPalette.
//
// THIS IS THE RESKIN SURFACE. The tokens are the light "instrument" look
// (mirroring pkg/webui/theme.css), emitted as literal hex because the standalone
// report has no CSS variables. They are constants, so the light output stays
// byte-identical (a snapshot golden depends on it) and `fmt` format strings stay
// constant.
//
// A dark theme is provided not by swapping these constants (which would break
// the report and the golden) but by Darken, a final light->dark hex substitution
// the terminal FIRE UI applies to each rendered SVG. A guard test (theme_test.go)
// fails if a chart hardcodes a chrome color outside this registry.
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
// and the instrument-style charts. Theme-independent.
const (
	themeMono = "'Spline Sans Mono', ui-monospace, SF Mono, Menlo, Consolas, monospace"
	themeSans = "'Instrument Sans',system-ui,sans-serif"
)

// darkChrome maps each light chrome hex to its terminal-dark counterpart (amber
// phosphor on warm charcoal). The semantic hues are the dataviz-validated dark
// set (validate_palette.js, all six checks pass on the #17130D surface); accent
// is the deep amber chart mark. Darken applies this substitution.
var darkReplacer = strings.NewReplacer(
	themeInk, "#EFE7D6",
	themeInkSoft, "#B4A991",
	themeMuted, "#7E7458",
	themeFaint, "#5B5340",
	themeGrid, "#2B2317",
	themeWell, "#221B11",
	themeAxis, "#3C3121",
	themeSurface, "#17130D",
	themeAccent, "#C4820F",
	themeGood, "#34A46E",
	themeWarn, "#D98A2E",
	themeBad, "#D2503F",
	themeDead, "#6B6250",
	// Caller-passed lifecycle (Rich/Broke/Dead) series colors -> dark set.
	"#12B76A", "#34A46E", // funded
	"#D92D20", "#D2503F", // broke
	"#D5DBE5", "#6B6250", // gone
)

// Darken rewrites a chart SVG from the light theme to the terminal-dark theme by
// substituting each chrome color. The FIRE terminal UI wraps every rendered SVG
// with it; the standalone report leaves the light output untouched. Series
// colors a caller passes in explicitly (e.g. stacked-area layers) are not chrome
// and pass through unchanged, so the caller supplies dark-ready series colors.
func Darken(svg string) string { return darkReplacer.Replace(svg) }

// darkMode makes finish apply Darken to every chart this process renders.
var darkMode bool

// SetDark selects the terminal dark theme for every chart rendered afterwards.
// The FIRE UI calls SetDark(true) at startup; the standalone report leaves it
// off. Process-global, set once before rendering.
func SetDark(on bool) { darkMode = on }

// finish applies the active theme to a freshly rendered light-theme SVG.
func finish(svg string) string {
	if darkMode {
		return Darken(svg)
	}
	return svg
}
