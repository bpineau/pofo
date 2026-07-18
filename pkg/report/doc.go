// Package report formats portfolio comparison results from a single model
// (Page):
//
//   - Render writes a self-contained HTML document: embedded SVG charts,
//     comparison and statistics table up top, detailed per-portfolio
//     sections folded (<details>), each with its performance curve,
//     realized-contribution timeline (two windows behind a toggle),
//     composition pies, segmented coverage bars and per-regime matrix;
//   - RenderText writes the summary for the terminal (aligned table,
//     best cells in green or starred).
//
// The document embeds one small self-contained script (no dependency, no
// network): instant tooltips for data-tip marks and a crosshair reading the
// charts' hover metadata (see pkg/chart/hover.go). Everything else is
// static; without JavaScript the report stays fully readable.
//
// Cells marked Best are highlighted; labels and notes are supplied by the
// caller, the package does no computation.
package report
