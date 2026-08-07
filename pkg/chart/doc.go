// Package chart plots financial series charts without any dependency:
//
//   - Line produces a self-contained SVG document (axes, grid, legend,
//     decimation of long series), embeddable as-is in an HTML page;
//   - Sparkline produces a bare inline SVG curve (no axes, no labels)
//     for table cells and summaries;
//   - Pie produces a self-contained SVG donut with a title and a legend,
//     for composition breakdowns (geography, sector, asset type);
//   - DivergingStack stacks signed series around a zero axis (positives up,
//     negatives down), with an optional net line and categorical strip: the
//     shape of a return-contribution timeline;
//   - BarMatrix lays out a small-multiples grid of horizontal diverging
//     bars (rows x categories on one shared scale), e.g. per-regime
//     realized contributions;
//   - Term produces a chart for the terminal (ANSI colors on a TTY,
//     distinct markers per series otherwise; Braille mode packs 2x4 dots
//     per cell for a smoother curve);
//   - Line and Term share the Series model and the default palette,
//     accessible via PaletteColor to keep multiple charts consistent.
//     The palette is the "instrument" identity set (petrol first); its
//     slot order is validated for adjacent-pair distinctness under common
//     color vision deficiencies and must not be permuted casually.
//
// Line labels sub-day spans with clock times (HH:MM), so the same
// renderer draws both daily and intraday series without any extra
// configuration.
//
// # Styling
//
// Options.Style adjusts everything beyond dimensions: background, font,
// grid/axes/legend visibility, an area fill under the first series, the
// tick formatter (see Compact), stroke width and date labelling. The zero
// Style is the default pofo look; StyleMinimal is a bare dialect for
// dense pages embedding many small charts.
//
// The "chrome" colors and fonts every chart draws with (grids, axes, labels,
// backgrounds, the semantic good/warn/bad/accent hues) live in one place,
// theme.go: that file is the reskin surface, mirroring pkg/webui/theme.css.
// Change a token there to restyle every chart at once, then re-run the chart
// snapshot golden (TestChartSnapshots with UPDATE_SNAPSHOTS=1). A guard test
// keeps new charts from scattering hardcoded hex again.
package chart
