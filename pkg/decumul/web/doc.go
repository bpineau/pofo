// Package web is a thin embedded HTTP UI for pkg/decumul: it serves a
// single page of sliders and, on each change, runs the Monte-Carlo in Go and
// returns chart SVGs and summary cards as JSON. The engine stays in Go; the
// browser only renders. Handler returns a ready-to-mount http.Handler.
//
// With a nil panel it serves the parametric playground (returns from
// mu/sigma/df sliders). With a historical panel it also offers the
// bootstrap and historical-cohort models and live per-holding re-weighting;
// a panel shorter than two years is treated as absent (Fit.Valid,
// minPanelMonths) so a degenerate fit can never seed a doom model.
//
// Three options tune the chrome, all surfaced to the front end through
// /api/meta: WithNav (cross-navigation to sibling surfaces),
// WithSourceLabel (which market this mount runs on, shown as a provenance
// pill in the top bar) and WithPicker (the in-drawer portfolio loader, whose
// example list, catalog URL and comparison-report URL the caller supplies,
// since only the embedding server knows those mounts and endpoints). The
// loader fills the drawer where no portfolio is bound, and folds away under
// the allocation bar where one is, so a loaded mount is not a dead end.
// Without a picker the page states in one line how to bind a portfolio from
// the command line.
//
// Spending policies are exclusive in the rail. The kernel resolves clashes
// by a fixed precedence (see the package comment of pkg/decumul), which is
// invisible to a reader ticking two boxes, so claiming one policy clears the
// controls of the others; a shared URL is left as it arrived, only dimmed,
// so old links keep reproducing the run their sender saw.
//
// Beyond the model strip and the sweeps, the analysis endpoints serve the
// sequence-risk decomposition (/api/decade), the deterministic replay of
// infamous historical vintages through the user's plan (/api/vintages), the
// median funding-mix layers (/api/income), the lived-spending fan, the
// mortality lifecycle and the planning curves.
package web
