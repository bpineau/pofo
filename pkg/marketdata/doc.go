// Package marketdata fetches, caches and post-processes historical asset
// prices (daily closes) from public sources, addressed by ticker, ISIN or
// alias.
//
// # Résolution
//
// Un identifiant passe par les étapes suivantes (voir ResolveCanonical):
//
//  1. les aliases intégrés (GOLD → XAUUSD, BHMG → GG00BQBFY362, …);
//  2. la liste embarquée ticker → ISIN des ETF/OPCVM européens (FundISIN);
//  3. le catalogue intégré de résolutions épinglées (Catalog), qui rend les
//     actifs courants déterministes et indépendants des moteurs de recherche;
//  4. sinon, une résolution multi-sources: chaque candidat de la recherche
//     Yahoo (entrées « fonds » d'abord), puis le Financial Times, puis
//     l'identifiant Morningstar découvert via Boursorama — la série à
//     l'historique le plus profond gagne, et la résolution est mise en cache.
//
// # Sources
//
// Yahoo Finance (clôtures ajustées), Stooq (secours tickers), Financial
// Times et Morningstar (valeurs liquidatives des fonds européens). Les
// téléchargements sont mis en cache sur disque (JSON, un fichier par
// instrument); un rafraîchissement en échec sert la donnée périmée avec un
// avertissement plutôt que d'échouer.
//
// # Données simulées
//
// ReadSimdata/WriteSimdata lisent et écrivent les historiques simulés
// permanents (répertoire simdata/) produits par le package simgen;
// ExtendBack recolle ces séries — ou un proxy (ProxySymbol) — devant les
// cotations réelles. La convention « suffixe SIM » (DBMFSIM = DBMF avec
// extension simulée) est décodée par SplitSim.
//
// # Boîte à outils
//
//   - Align fusionne les calendriers de cotation de plusieurs séries
//     (union des dates, cours forward-fillés) ;
//   - Fees renvoie le TER publié d'un actif (catalogue épinglé, cache
//     disque, sinon tearsheets FT et justETF) ;
//   - UCITSFlag/GuessUCITS et LooksDistributing qualifient les fonds ;
//   - CanonicalID normalise tout identifiant accepté (alias, ISIN, ticker
//     de la liste embarquée) vers sa forme canonique ;
//   - IsISIN valide un ISIN, clé de contrôle comprise.
package marketdata
