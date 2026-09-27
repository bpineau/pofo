// Package refgen holds what several reference-data generators under cmd/ share:
// reading a public source (Get, FRED), the checks every refreshed series must
// pass before it may replace the bundled file (SameHistory, MonthlyCadence,
// FlatRun), and writing the file in the bundle's simdata format with its
// "# source:" header and, for a series that stops by design, its "# ends:"
// declaration (Write).
//
// The checks exist because a download that completed proves nothing. A source
// that stops being updated keeps answering HTTP 200 with the same rows, a
// renamed column parses into the wrong numbers, and a publisher's revision of
// old data silently moves every plate and golden that read it. Each check
// names one of those failures, and a generator refuses to write while any of
// them fails.
//
// The package lives under cmd/internal because it is a generator's tool: the
// pofo binary embeds the CSVs these generators write and never fetches a
// reference series itself.
package refgen
