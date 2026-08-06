// Package metrics calcule les statistiques de risque et de rendement d'une
// série de valeurs datées : CAGR, volatilité, Sharpe, Sortino, Ulcer Index,
// Max Drawdown, TTR (durée de récupération) et Beta contre un benchmark.
//
// # Conventions
//
// Connaître les conventions est indispensable pour comparer les résultats à
// d'autres outils :
//
//   - les séries sont des clôtures quotidiennes ; volatilité et ratios sont
//     annualisés sur 252 jours de bourse ;
//   - Sharpe et Sortino utilisent un taux sans risque nul (comme Curvo) ;
//     PortfolioVisualizer et LazyPortfolioETF utilisent les T-bills et des
//     données mensuelles — leurs Sharpe ressortent ≈ 0,10–0,15 plus bas ;
//   - Max Drawdown, Ulcer et TTR sont mesurés sur clôtures quotidiennes,
//     plus sévères que les outils à pas mensuel (COVID 2020 : −33,7 % en
//     quotidien contre ≈ −20 % en clôtures mensuelles) ;
//   - le CAGR utilise des années de 365,25 jours entre la première et la
//     dernière date.
//
// Point d'entrée principal : Compute. Beta apparie les rendements
// quotidiens par date avec ceux du benchmark. Returns et Mean sont exposés
// comme briques de base.
//
// Ces calculs sont verrouillés par les tests étalon du paquet golden, qui
// les confrontent à des références externes (rendements annuels officiels
// S&P 500 TR, drawdowns canoniques).
package metrics
