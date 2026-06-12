// Package chart plots financial series charts without any dependency:
//
//   - Line produces a self-contained SVG document (axes, grid, legend,
//     decimation of long series), embeddable as-is in an HTML page;
//   - Term produces a chart for the terminal (ANSI colors on a TTY,
//     distinct markers per series otherwise);
//   - both share the Series model and the default palette, accessible
//     via PaletteColor to keep multiple charts consistent.
package chart
