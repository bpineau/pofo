// Package chart plots financial series charts without any dependency:
//
//   - Line produces a self-contained SVG document (axes, grid, legend,
//     decimation of long series), embeddable as-is in an HTML page;
//   - Pie produces a self-contained SVG donut with a title and a legend,
//     for composition breakdowns (geography, sector, asset type);
//   - Term produces a chart for the terminal (ANSI colors on a TTY,
//     distinct markers per series otherwise);
//   - Line and Term share the Series model and the default palette,
//     accessible via PaletteColor to keep multiple charts consistent.
//
// Line labels sub-day spans with clock times (HH:MM), so the same
// renderer draws both daily and intraday series without any extra
// configuration.
package chart
