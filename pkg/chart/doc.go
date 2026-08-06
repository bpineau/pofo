// Package chart trace des graphes de séries financières sans aucune
// dépendance :
//
//   - Line produit un document SVG autonome (axes, grille, légende,
//     décimation des longues séries), embarquable tel quel dans une page
//     HTML ;
//   - Term produit un graphe pour le terminal (couleurs ANSI sur un TTY,
//     marqueurs distincts par série sinon) ;
//   - les deux partagent le modèle Series et la palette par défaut,
//     accessible via PaletteColor pour garder plusieurs graphes cohérents.
package chart
