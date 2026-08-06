// Package simgen reconstruit le passé manquant des actifs complexes (fonds
// 90/60, managed futures, …) et valide chaque reconstruction contre les
// cotations réelles. Les résultats sont stockés en fichiers « simdata »
// permanents que portfodor recolle devant les historiques réels.
//
// # Boîte à outils
//
//   - BuildFrame aligne les rendements quotidiens de plusieurs composants
//     (les séries de taux comme ^IRX sont converties en accrual) ;
//   - Composite compose un indice base 100 à poids constants, jambes
//     « excess » (futures) et frais annuels compris ;
//   - TSMOM est un moteur time-series momentum paramétrable (marchés,
//     lookback, vol cible, levier) pour répliquer des stratégies trend ;
//   - FitBackcast régresse un actif sur des facteurs et rejoue le modèle
//     sur tout l'historique (refusé sous un R² plancher : ErrUnfaithful) ;
//   - WithRefData sert des séries de référence locales (datasets/refdata/)
//     avant
//     toute source réseau ; Validate mesure corrélation quotidienne et
//     hebdomadaire, beta, tracking error et CAGR contre le réel ;
//   - les recettes livrées (All, Find) assemblent ces briques pour NTSX,
//     NTSG, URTH, IWDA, ZROZ, IEF, TLT, XAUUSD, DBMF, KMLM, CTA et le
//     fonds Winton Trend-Equity.
package simgen
