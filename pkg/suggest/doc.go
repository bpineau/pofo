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
package suggest
