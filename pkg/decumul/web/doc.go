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
// Beyond the model strip and the sweeps, the analysis endpoints serve the
// sequence-risk decomposition (/api/decade), the deterministic replay of
// infamous historical vintages through the user's plan (/api/vintages), the
// median funding-mix layers (/api/income), the lived-spending fan, the
// mortality lifecycle and the planning curves.
package web
