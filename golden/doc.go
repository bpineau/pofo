// Package golden contient les tests « étalon » de portfodor : ils rejouent
// la simulation et les métriques sur des données réelles gelées (testdata/)
// et comparent les résultats à des références externes validées à la main
// (rendements annuels officiels S&P 500 TR, drawdowns canoniques,
// statistiques publiées par LazyPortfolioETF). Toute dérive des calculs —
// CAGR, volatilité, Sharpe, Sortino, Ulcer, Max Drawdown, TTR — au-delà des
// tolérances fait échouer la suite.
package golden
