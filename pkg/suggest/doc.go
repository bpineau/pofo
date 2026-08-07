// Package suggest recommends catalog assets to add to a portfolio so it
// covers the market regimes it is missing, and flags redundant holdings.
//
// It is the structure-first half of the optimizer: candidates are screened
// by what they ARE (asset class, strategy, bond duration, currency hedging,
// factor tilts and notional exposures → macro-regime coverage and statistical
// diversification) before any return-based ranking, and the
// return ranking is validated out-of-sample (walk-forward) so a suggestion
// reflects a consistent benefit rather than one lucky period. Conventions
// match pkg/metrics: simple daily returns, 252 trading days per year.
//
// The same metadata also powers the look-through composition views
// (composition.go): AssetClassSplit opens stacked funds into their legs,
// GeographySplit and EquitySectorSplit aggregate the published breakdowns,
// CurrencySplit derives the fiat-currency exposure (quote currencies are
// ignored in favor of what the capital actually moves with), DurationSplit
// tallies interest-rate duration, and Contributors decomposes Coverage per
// holding. All splits are label → fraction-of-capital maps; presentation is
// the caller's job.
package suggest
