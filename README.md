# portfodor

Outil en Go pour visualiser et comparer des portefeuilles boursiers dans le
temps — et bibliothèques réutilisables pour récupérer des historiques de
cours, calculer des métriques de risque/rendement et produire des graphes SVG.

Le CLI lit des fichiers d'allocation, télécharge les historiques (Yahoo
Finance, Financial Times, Morningstar, Stooq), reconstruit le passé manquant
(proxys et données simulées), simule chaque portefeuille avec rebalancement
périodique et génère un rapport HTML autonome ouvert dans le navigateur
(sections par portefeuille repliées, comparaison et statistiques en évidence).

## Utilisation

```sh
go build ./cmd/portfodor
./portfodor mon-portefeuille.txt autre.txt     # rapport HTML dans /tmp + open
./portfodor -assets WPEA,NTSG,CSPX             # compare des actifs isolés (100 % chacun)
./portfodor -cli -assets VOO,IWDA              # quick check dans le terminal
./portfodor -warmup                            # précharge le cache du catalogue
```

L'option `-assets` traite chaque identifiant comme un portefeuille investi à
100 % dessus — pratique pour comparer des ETF entre eux sans écrire de
fichier. Elle se cumule avec des fichiers de portefeuille.

## Format des fichiers de portefeuille

Une ligne par actif : `<poids en %> <identifiant> [texte libre]`. Tout ce qui
suit un `#` est un commentaire ; les lignes vides sont ignorées. Le nom du
portefeuille est le nom du fichier sans extension.

```
# Description, liens, notes…
#meta rebalance:30   # directive: ce portefeuille rebalance tous les 30 jours
#meta extra-fees:0.5 # frais d'enveloppe/mandat, appliqués en plus à tout le portefeuille
60   VTI            actions US
25,5 IE00B4L5Y983   # ISIN résolu automatiquement (virgule décimale acceptée)
14.5 GOLD           # alias intégré → or XAU/USD
```

Les lignes `#meta clé:valeur` portent des directives par portefeuille :
`rebalance:N` (jours entre rebalancements, `0` = jamais) et `extra-fees:X`
(synonyme `envelope-fees:X`) — des frais **additionnels en %/an appliqués à
l'ensemble du portefeuille**, en plus des TER individuels des actifs :
enveloppe assurance-vie/PER, mandat de gestion, frais courtier… N'étant pas
inclus dans les cours (contrairement aux TER), ils sont **déduits** de la
performance simulée.
Une troisième colonne numérique optionnelle déclare le TER d'un actif
(ex. `60 VOO 0.03`), sinon il est récupéré automatiquement (FT, justETF)
et mis en cache 6 mois ; `-no-fees` désactive cette récupération. Le
rapport affiche les frais par actif et une ligne « Frais courants
pondérés » dans le tableau de statistiques.

L'identifiant peut être un ticker US (`VTI`), un ticker européen de la liste
embarquée (`IWDA`, `CSPX`, `CW8`…), un ISIN, ou un alias intégré (`GOLD`,
`WTI`, `BHMG`, `AMUNDI-VOLATILITY`, `WINTON-TREND-EQUITY`…). Si la somme des
poids ne fait pas 100, ils sont normalisés avec avertissement.

**Convention SIM** : un identifiant nu (`DBMF`, `NTSG`, `VOO`) utilise les
seules cotations réelles de l'actif — l'historique commence à sa date de
création. Le suffixe `SIM` (`DBMFSIM`, `NTSGSIM`, `VOOSIM`…) autorise en plus
l'extension de la période non couverte, via les fichiers `simdata/` puis les
proxys connus ; les cotations réelles gardent toujours la priorité là où
elles existent. `-no-simulate` ignore les suffixes SIM globalement.

## Options principales

| Option | Défaut | Description |
|---|---|---|
| `-out` | `/tmp/portfodor-<horodatage>.html` | fichier HTML généré |
| `-data` | `data` | cache des cotations (JSON) |
| `-simdata` | `simdata` | historiques simulés permanents |
| `-rebalance` | `90` | rebalancement tous les N jours calendaires (0 = jamais) |
| `-start` | `2006-01-01` | date de début souhaitée |
| `-benchmark` | `^GSPC` | référence pour le Beta |
| `-cache-age` | `720h` (1 mois) | fraîcheur du cache avant retéléchargement |
| `-assets` | | liste `A,B,C` : chaque actif comparé comme un portefeuille à 100 % |
| `-cli` | | courbes et tableau récapitulatif dans le terminal, sans HTML |
| `-width` | `$COLUMNS` ou 100 | largeur du graphe `-cli` (plus large = plus de granularité) |
| `-warmup` | | précharge le catalogue d'actifs intégré puis s'arrête |
| `-no-open`, `-no-simulate` | | ne pas ouvrir le navigateur / ignorer les suffixes SIM |

## Données

- **Résolution** : aliases → liste embarquée ticker→ISIN (ETF/OPCVM européens)
  → catalogue intégré de résolutions épinglées → recherche multi-sources
  (Yahoo, FT, Morningstar via Boursorama), la série la plus profonde gagnant.
- **Cache** : 1 mois par défaut ; un rafraîchissement raté **sert la donnée
  périmée** avec un avertissement stderr (les graphes peuvent s'arrêter avant
  aujourd'hui), et n'efface jamais rien.
- **Extension d'historique** (identifiants `…SIM` uniquement) : d'abord les
  fichiers `simdata/` (ci-dessous), sinon un proxy connu (VOO→^GSPC,
  BND→VBMFX, …), recalé sur la première cotation réelle. Le rapport signale
  chaque portion simulée.

## Données simulées (simdata/)

Les actifs complexes (fonds 90/60, managed futures…) sont reconstruits par
`cmd/simgen` à partir de briques à long historique, validés contre leurs
cotations réelles, puis stockés en CSV auto-documentés (méthode, validation,
date) dans `simdata/` :

```sh
go build ./cmd/simgen && ./simgen          # régénère tout
./simgen -dry NTSX                         # valide sans écrire
```

Recettes livrées et qualité mesurée (corrélation quotidienne/hebdomadaire des
rendements vs réel ; le réel est toujours greffé par-dessus la simulation là
où il existe) :

| Actif | Méthode | Validation |
|---|---|---|
| NTSX (UCITS) | 0.90×VFINX + 0.60×(VFITX−cash) + 0.10×cash (1991→) | corr 0.96 / hebdo 0.99 vs NTSX US |
| NTSG (UCITS) | variante monde 60/40 US/intl | hebdo 0.86 (cotation LSE peu liquide) |
| URTH, IWDA | 0.60×VFINX + 0.40×VTMGX (1999→) | corr 0.90 / hebdo 0.97 |
| ZROZ, IEF, TLT | réf. importées dérivées des courbes de taux US (1962→) | corr 1.00 sur 16–24 ans d'overlap |
| XAUUSD (GOLD) | or spot importé (1968→), réel GC=F greffé | corr 1.00 |
| KMLM | indice MLM officiel (1987→) + frais ETF 0.90 % | corr 1.00 |
| DBMF | SG CTA Index officiel (2000→) | corr 0.68 / hebdo 0.75, beta 0.96 |
| CTA | SG Trend Index officiel (2000→) | corr 0.54 — stratégie propriétaire, écart accepté |
| Winton Trend-Equity | actions monde + 0.5×fonds Winton Trend (réel 2019→, sim avant) | hebdo 0.92 |
| Amundi Volatility, BH Macro | backcast par régression **refusé** (R² 0.20 / 0.00) | historique réel seul (2007→) |

Les stratégies discrétionnaires ne se répliquent pas honnêtement par
facteurs : plutôt que d'inventer des données, le générateur les refuse sous
un R² plancher.

## Données de référence (refdata/)

`refdata/` contient des séries de référence importées une fois pour toutes
(provenance et méthode en tête de chaque fichier) : indices officiels SG
Trend/SG CTA, historique de l'indice MLM, treasuries 7-10/20+/25+ dérivés des
courbes de taux US depuis 1962, or spot depuis 1968, fonds Winton Trend.
`cmd/simgen` les consomme en priorité (`-refdata`), avant les sources réseau.

## Utilisation comme bibliothèque

Les packages sont indépendants du CLI et documentés (`go doc portfodor/...`) :

```go
// Récupérer un historique (résolution + cache transparents).
client := marketdata.NewClient("data")
series, err := client.Fetch("IWDA", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

// Calculer CAGR, Sharpe, Sortino, Ulcer, MaxDD, TTR, Beta…
stats, err := metrics.Compute(dates, values)

// Tracer un SVG autonome.
svg := chart.Line(chart.Options{Title: "Comparaison"}, []chart.Series{{Name: "P1", Dates: dates, Values: values}})

// Parser et simuler un portefeuille (rebalancement N jours).
spec, _ := portfolio.ParseFile("p.txt")
sim, _ := portfolio.Simulate(p, 90)
```

- `marketdata` — résolution (aliases, ISIN, catalogue), téléchargement
  multi-sources, cache, simdata, proxys.
- `metrics` — statistiques de séries de valeurs (rendements, drawdowns, Beta).
- `chart` — graphes SVG en ligne pure stdlib.
- `portfolio` — parsing des fichiers d'allocation et simulation rebalancée.
- `report` — rendu du rapport HTML.
- `simgen` — moteur de reconstruction (composites linéaires, références
  importées, backcasts par régression) et recettes validées.

## Limites connues

- Pas de conversion de change : mélanger des devises déclenche un
  avertissement, les rendements restent dans la devise de chaque actif.
- Les proxys d'indices de prix (^GSPC, ^NDX…) omettent les dividendes sur la
  portion simulée ; les réplications managed futures (corr ≈ 0.3–0.5)
  reflètent le régime de ces stratégies, pas leurs positions quotidiennes.

## Tests étalon (golden)

`golden/` rejoue la simulation sur des données réelles gelées (SPY 2006-2025,
URTH 2012-2025) et compare CAGR, volatilité, Sharpe, Sortino, Ulcer, Max
Drawdown et TTR à des références externes validées (rendements annuels
officiels S&P 500 TR, drawdowns canoniques GFC/COVID, LazyPortfolioETF).
Toute dérive de calcul au-delà des tolérances fait échouer `go test ./golden`.

## Développement

```sh
go test ./...   # tests unitaires + exemples, sans réseau
go vet ./...
```

Aucune dépendance externe : uniquement la bibliothèque standard.
