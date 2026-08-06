// Package portfolio lit des descriptions de portefeuilles et les simule
// dans le temps.
//
// # Format des fichiers
//
// Une ligne par actif :
//
//	<poids en %> <identifiant> [frais en %/an] [texte libre]
//
// Tout ce qui suit un # est un commentaire ; lignes vides ignorées ; le
// poids et les frais acceptent la virgule décimale et un suffixe %. Les
// poids ne sommant pas à 100 sont normalisés (avertissement dans
// Spec.Warnings). Les lignes « #meta clé:valeur » portent des directives :
//
//	#meta rebalance:N     rebalancement tous les N jours (0 = jamais)
//	#meta extra-fees:X    frais annuels appliqués à tout le portefeuille
//	                      (enveloppe, mandat), déduits par Simulate
//
// L'interprétation des identifiants (tickers, ISIN, aliases, suffixe SIM)
// appartient à l'appelant — voir marketdata.Fetch et marketdata.SplitSim.
//
// # Simulation
//
// Simulate rejoue le portefeuille en base 100 sur l'union des calendriers
// de cotation (cours forward-fillés via marketdata.Align), du premier jour
// où tous les actifs cotent au dernier jour où tous cotent encore, avec
// rebalancement vers les poids cibles tous les N jours calendaires et
// déduction quotidienne des frais d'enveloppe. Les TER des actifs ne sont
// jamais déduits : ils sont déjà reflétés dans les cours.
package portfolio
