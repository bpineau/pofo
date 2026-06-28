// Package report formats portfolio comparison results from a single model
// (Page):
//
//   - Render writes a self-contained HTML document (embedded SVG charts,
//     comparison and statistics table up top, detailed per-portfolio
//     sections folded (<details>), no JavaScript;
//   - RenderText writes the summary for the terminal (aligned table,
//     best cells in green or starred).
//
// Cells marked Best are highlighted; labels and notes are supplied by the
// caller, the package does no computation.
package report
