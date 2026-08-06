// Package report met en forme les résultats de comparaison de portefeuilles
// à partir d'un modèle unique (Page) :
//
//   - Render écrit un document HTML autonome — graphes SVG embarqués,
//     comparaison et tableau de statistiques en tête, sections détaillées
//     par portefeuille repliées (<details>), sans JavaScript ;
//   - RenderText écrit le résumé pour le terminal (tableau aligné,
//     meilleures cellules en vert ou étoilées).
//
// Les cellules marquées Best sont surlignées ; les libellés et notes sont
// fournis par l'appelant, le paquet ne fait aucun calcul.
package report
