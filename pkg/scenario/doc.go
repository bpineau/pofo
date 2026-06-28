// Package scenario generates synthetic real-return paths from either
// parameters or a panel of historical returns, behind a single Source
// interface. It is decumulation-agnostic and reusable for any
// path-dependent study (accumulation, glidepaths, stress tests).
//
// All returns are periodic and real (inflation already removed): use
// Deflate to obtain them from nominal prices and an HICP series. A Source
// yields one Sequence per Draw; callers run many Draws for a Monte-Carlo,
// or iterate HistoricalCohorts for a deterministic every-start-date backtest.
package scenario
