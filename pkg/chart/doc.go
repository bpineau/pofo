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
//     For a chart that draws n series at once, prefer PaletteFor(n): it
//     picks WHICH n hues to use so the least distinguishable pair of the
//     set is as far apart as possible (measured in OKLab, for normal
//     vision and for the three common deficiencies). Taking the first n
//     slots instead pairs rust with ochre from four series on, the
//     palette's closest pair. PaletteColor stays the right call for fixed
//     vocabularies (asset classes, regimes), where an entity's color must
//     not depend on how many others are on screen.
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
// # Hover metadata
//
// Every chart embeds a machine-readable copy of its data as an SVG
// <metadata class="hover"> element (hover.go), so a front-end can read a
// value out under the pointer without re-deriving the drawn geometry. The
// payload always carries the plot box, plus either the x domain (the
// continuous kinds: line, fan, stack) or one pixel anchor per mark (the
// discrete kinds: bars, cat, scatter, with Axis saying whether they are laid
// out along x, along y, or freely). The anchors are what let a hover layer
// answer the pointer ANYWHERE over the plot rather than only where ink was
// painted, which is the difference between a chart that reads out at its
// centre and one that stays mute there. Discrete marks also carry a native
// <title> as a fallback for consumers with no hover layer of their own.
//
// The "chrome" colors and fonts every chart draws with (grids, axes, labels,
// backgrounds, the semantic good/warn/bad/accent hues) live in one place,
// theme.go: that file is the reskin surface, mirroring pkg/webui/theme.css.
// Change a token there to restyle every chart at once, then re-run the chart
// snapshot golden (TestChartSnapshots with UPDATE_SNAPSHOTS=1). A guard test
// keeps new charts from scattering hardcoded hex again.
package chart
