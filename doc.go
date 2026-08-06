// Package pofo is the root of a dependency-free Go toolkit for tracking
// and designing stock-market portfolios. The command in cmd/pofo is one
// application built on it; every capability it exposes is reusable from the
// library packages under pkg/.
//
// # Packages
//
// The toolkit is a layered set of focused packages:
//
//   - pkg/datasets: the bundled, versioned data, embedded into the binary:
//     the curated asset catalog (assetmeta/assets.json, typed as
//     datasets.Asset) and the permanent simulated histories (simdata/).
//     This is the single source of truth other packages read from.
//   - pkg/marketdata: fetches, caches and post-processes daily and intraday
//     prices from public sources, addressed by ticker, ISIN or alias; resolves
//     identifiers against the embedded catalog and aligns trading calendars.
//   - pkg/metrics: risk/return statistics (CAGR, volatility, Sharpe,
//     Sortino, Ulcer, max drawdown, time-to-recovery, Beta, IRR).
//   - pkg/optimize: long-only weights for an objective (max-sharpe,
//     min-volatility, risk-parity) from the assets' historical returns.
//   - pkg/portfolio: the allocation file format and the rebalanced,
//     fee-aware simulation that replays a portfolio over time.
//   - pkg/suggest: structure-first analysis: regime coverage, redundancy
//     and out-of-sample-validated gap-filling suggestions from the catalog.
//   - pkg/simgen: reconstruction of the missing past of complex assets
//     (capital-efficient funds, managed futures) into simdata files.
//   - pkg/scenario: synthetic real-return path generation (parametric
//     Student-t, block/stationary bootstrap, historical cohorts) behind one
//     Source interface; the input to decumulation studies.
//   - pkg/decumul: decumulation/FIRE engine over a scenario.Source: ruin
//     probability, FIRE outcome metrics, capital/buffer sizing and sweeps,
//     with a thin embedded live UI under pkg/decumul/web.
//   - pkg/chart: dependency-free SVG and terminal charts (line, pie, bars,
//     heatmap).
//   - pkg/report: HTML and text rendering of a portfolio-comparison model.
//
// # Typical pipeline
//
// A tracking or design application usually wires the packages in this order:
// parse an allocation with pkg/portfolio, resolve and fetch each holding's
// series with pkg/marketdata (splicing simulated history from pkg/datasets
// when a fund is young), simulate the portfolio, score it with pkg/metrics,
// and either render it with pkg/chart and pkg/report or improve it with
// pkg/optimize and pkg/suggest.
//
// # Conventions
//
// Series are daily closes; volatility and ratios annualize over 252 trading
// days with a zero risk-free rate. Watch the unit conventions, documented per
// package: pkg/portfolio and marketdata.Fees express fees in PERCENT per year
// while pkg/simgen uses FRACTIONS, and weights are fractions in the simulation
// but percent in portfolio files.
//
// This package itself holds no code: start from the package docs above
// (for example, go doc github.com/bpineau/pofo/pkg/marketdata).
package pofo
