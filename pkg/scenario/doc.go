// Package scenario generates synthetic real-return paths from either
// parameters or a panel of historical returns, behind a single Source
// interface. It is decumulation-agnostic and reusable for any
// path-dependent study (accumulation, glidepaths, stress tests).
//
// All returns are periodic and real (inflation already removed): use
// Deflate to obtain them from nominal prices and an HICP series. A Source
// yields one Sequence per Draw; callers run many Draws for a Monte-Carlo,
// or iterate HistoricalCohorts for a deterministic every-start-date backtest.
//
// Resampling sources come in two shapes. BlockBootstrap and
// StationaryBootstrap resample one combined history (a Panel collapsed by its
// weights), preserving cross-asset correlation and regimes. PooledBootstrap
// resamples across a POOL of separate histories, keeping each series' internal
// ordering but mixing series between blocks: fed per-country records it models a
// random developed-market retiree, whose run can land inside a single market's
// disaster that a pre-diversified world index would have averaged away.
//
// A driver that draws thousands of paths should call Prepare on its Source
// once, before the loop: the resampling sources collapse their Panel into one
// weighted history at every Draw, which Prepare hoists out. The prepared
// source draws byte-identical paths, and a source with nothing to hoist comes
// back unchanged.
package scenario
